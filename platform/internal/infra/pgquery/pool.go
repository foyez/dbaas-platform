// Package pgquery implements ports.Querier using pgx connection pools,
// one per instance, created lazily and evicted after a period of no use.
package pgquery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
)

const (
	// maxRows caps how many rows are ever returned to a client in one
	// response, so a `SELECT *` on a large table can't exhaust the API
	// pod's memory or the browser rendering the result table.
	maxRows = 500

	// queryTimeout bounds how long a single statement may run. pgx
	// cancels the in-flight query server-side when ctx is done, so a
	// runaway query can't hold a pooled connection (or the underlying
	// Postgres backend) forever.
	queryTimeout = 10 * time.Second

	// idleEvictAfter closes a per-instance pool that hasn't been queried
	// in this long, so the API doesn't accumulate open connections to
	// every instance anyone has ever queried.
	idleEvictAfter = 10 * time.Minute
	evictInterval  = time.Minute

	maxConnsPerInstance = 3
)

type pooledConn struct {
	pool     *pgxpool.Pool
	lastUsed time.Time
}

type Pool struct {
	mu    sync.Mutex
	pools map[string]*pooledConn

	stopEviction chan struct{}
}

func NewPool() *Pool {
	p := &Pool{
		pools:        map[string]*pooledConn{},
		stopEviction: make(chan struct{}),
	}

	go p.evictLoop()

	return p
}

func (p *Pool) ExecuteQuery(
	ctx context.Context,
	instanceID string,
	creds *domain.InstanceCredentials,
	sql string,
) (result *domain.QueryResult, err error) {
	// A span per statement, deliberately without the SQL text: it can
	// contain values the user typed, same reason the service doesn't log it.
	ctx, span := tracer.Start(ctx, "sql.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system.name", "postgresql"),
			attribute.String("dbaas.instance_id", instanceID),
		),
	)
	defer func() { endQuerySpan(span, result, err) }()

	pool, err := p.getOrCreatePool(ctx, instanceID, creds)
	if err != nil {
		return nil, fmt.Errorf("connecting to instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	start := time.Now()

	// SimpleProtocol forces text format for results (same as psql),
	// instead of pgx's default binary format - which decodes types like
	// uuid into raw [16]byte arrays that JSON-marshal as arrays of
	// numbers rather than readable strings. Safe here since this is a
	// single ad-hoc statement with no bind parameters.
	rows, err := pool.Query(ctx, sql, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", queryError(ctx, err))
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = fd.Name
	}

	resultRows := [][]any{}

	truncated := false

	for rows.Next() {
		if len(resultRows) >= maxRows {
			truncated = true
			break
		}

		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("reading row: %w", err)
		}

		for i, v := range values {
			values[i] = normalizeValue(v)
		}

		resultRows = append(resultRows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading results: %w", queryError(ctx, err))
	}

	rowCount := len(resultRows)

	// Statements with no result set (INSERT/UPDATE/DELETE/DDL) report
	// their effect via the command tag instead of returned rows.
	if len(columns) == 0 {
		rowCount = int(rows.CommandTag().RowsAffected())
	}

	return &domain.QueryResult{
		Columns:    columns,
		Rows:       resultRows,
		RowCount:   rowCount,
		Truncated:  truncated,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

func (p *Pool) getOrCreatePool(
	ctx context.Context,
	instanceID string,
	creds *domain.InstanceCredentials,
) (*pgxpool.Pool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if existing, ok := p.pools[instanceID]; ok {
		existing.lastUsed = time.Now()
		return existing.pool, nil
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		creds.Username, creds.Password, creds.Host, creds.Port,
		creds.Database,
	)

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parsing connection config: %w", err)
	}

	cfg.MaxConns = maxConnsPerInstance
	cfg.MaxConnIdleTime = idleEvictAfter
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	p.pools[instanceID] = &pooledConn{pool: pool, lastUsed: time.Now()}

	return pool, nil
}

func (p *Pool) evictLoop() {
	ticker := time.NewTicker(evictInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopEviction:
			return
		case <-ticker.C:
			p.evictIdle()
		}
	}
}

func (p *Pool) evictIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, pc := range p.pools {
		if time.Since(pc.lastUsed) > idleEvictAfter {
			pc.pool.Close()
			delete(p.pools, id)
		}
	}
}

// Close releases every pooled connection. Call once on server shutdown.
func (p *Pool) Close() {
	close(p.stopEviction)

	p.mu.Lock()
	defer p.mu.Unlock()

	for id, pc := range p.pools {
		pc.pool.Close()
		delete(p.pools, id)
	}
}

// normalizeValue converts pgx-decoded Go values that don't have a sane
// JSON representation into ones that do. [16]byte (a fixed-size array, used
// for uuid columns) is the concrete case that surfaced this: encoding/json
// has a special case that marshals a *slice* of bytes as a base64 string,
// but that special case only applies to []byte - a fixed-size *array* like
// [16]byte falls through to the generic array path and marshals as a JSON
// array of 16 numbers instead of a UUID string. Deliberately a switch, not
// a generic reflection-based fallback, so any future case added here is a
// conscious decision rather than silently-different behavior.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case [16]byte:
		return fmt.Sprintf(
			"%x-%x-%x-%x-%x",
			val[0:4], val[4:6], val[6:8], val[8:10], val[10:16],
		)
	default:
		return v
	}
}

// queryError turns a failure of the user's statement into a domain error the
// API can show them. ctx is the statement's own timeout context. Anything
// else (connection problems, decoding errors) is returned unchanged.
func queryError(ctx context.Context, err error) error {
	// Check the deadline first: when pgx cancels a statement that ran too
	// long, PostgreSQL reports it as a generic 57014 query_canceled error.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return domain.ErrQueryTimeout
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return &domain.QueryError{
			SQLState: pgErr.Code,
			Message:  pgErr.Message,
			Detail:   pgErr.Detail,
			Hint:     pgErr.Hint,
			Position: int(pgErr.Position),
		}
	}

	return err
}

var tracer = otel.Tracer("github.com/foyez/dbaas-platform/platform/internal/infra/pgquery")

// endQuerySpan records a statement's outcome on its span. For a rejected
// statement it records only the SQLSTATE: PostgreSQL's message can contain
// row values.
func endQuerySpan(span trace.Span, result *domain.QueryResult, err error) {
	defer span.End()

	var queryErr *domain.QueryError
	switch {
	case err == nil:
		span.SetAttributes(
			attribute.Int("db.response.returned_rows", result.RowCount),
			attribute.Bool("dbaas.truncated", result.Truncated),
		)
	case errors.As(err, &queryErr):
		span.SetAttributes(attribute.String("db.response.status_code", queryErr.SQLState))
		span.SetStatus(codes.Error, "SQLSTATE "+queryErr.SQLState)
	case errors.Is(err, domain.ErrQueryTimeout):
		span.SetStatus(codes.Error, "query timed out")
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, "query failed")
	}
}
