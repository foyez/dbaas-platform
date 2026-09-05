// Package pgquery implements ports.Querier using pgx connection pools,
// one per instance, created lazily and evicted after a period of no use.
package pgquery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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
) (*domain.QueryResult, error) {
	pool, err := p.getOrCreatePool(ctx, instanceID, creds)
	if err != nil {
		return nil, fmt.Errorf("connecting to instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	start := time.Now()

	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
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

		resultRows = append(resultRows, values)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading results: %w", err)
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
