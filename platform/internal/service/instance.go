// Package service contains application business logic implementations.
// Services orchestrate domain operations and coordinate dependencies
// such as repositories, Kubernetes clients, or external providers.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/infra/loki"
	"github.com/foyez/dbaas-platform/platform/internal/ports"
)

// instanceService implements domain.InstanceService.
type instanceService struct {
	client  ports.InstanceClient
	querier ports.Querier
	loki    *loki.Client
	logger  *slog.Logger
}

func NewInstanceService(
	client ports.InstanceClient,
	querier ports.Querier,
	loki *loki.Client,
	logger *slog.Logger,
) *instanceService {
	return &instanceService{
		client:  client,
		querier: querier,
		loki:    loki,
		logger:  logger,
	}
}

// CreateInstance creates a new instance or returns an existing instance
// when the same idempotency key was already processed.
func (s *instanceService) CreateInstance(
	ctx context.Context,
	input domain.CreateInstanceInput,
) (*domain.CreateInstanceResult, error) {
	if err := validateCreateInstance(input); err != nil {
		return nil, err
	}

	instance, err := s.client.CreateInstance(ctx, input)
	if err != nil {
		return nil, err
	}

	return &domain.CreateInstanceResult{
		Instance: instance,
		Replayed: false,
	}, nil
}

func validateCreateInstance(
	input domain.CreateInstanceInput,
) error {
	if input.Name == "" {
		return ErrInvalidInput
	}
	if input.Version < 14 || input.Version > 16 {
		return ErrInvalidInput
	}
	if input.Storage == "" {
		return ErrInvalidInput
	}

	return nil
}

// GetInstance retrives a single instances by its ID.
func (s *instanceService) GetInstance(ctx context.Context, id, userID string) (*domain.Instance, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	return s.client.GetInstance(ctx, id, userID)
}

// GetInstanceLogs retrives logs of a instance by its ID.
func (s *instanceService) GetInstanceLogs(ctx context.Context, id, userID string) ([]loki.LogLine, error) {
	if _, err := s.client.GetInstance(ctx, id, userID); err != nil {
		return nil, err
	}

	logql := fmt.Sprintf(`{instance_id=%q}`, id)
	return s.loki.Query(ctx, logql, 200, 24*time.Hour, time.Now())
}

// GetCredentials retrieves the credentials for an instance.
func (s *instanceService) GetCredentials(
	ctx context.Context,
	id, userID string,
) (*domain.InstanceCredentials, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	return s.client.GetInstanceCredentials(ctx, id, userID)
}

// ExecuteQuery runs a single ad-hoc SQL statement against id's database.
// Ownership is enforced the same way as every other instance operation,
// via GetInstanceInternalCredentials.
func (s *instanceService) ExecuteQuery(
	ctx context.Context,
	id, userID, sql string,
) (*domain.QueryResult, error) {
	if id == "" || strings.TrimSpace(sql) == "" {
		return nil, ErrInvalidInput
	}

	creds, err := s.client.GetInstanceInternalCredentials(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	result, err := s.querier.ExecuteQuery(ctx, id, creds, sql)
	if err != nil {
		return nil, err
	}

	// Deliberately not logging the SQL text itself - it may contain
	// values a user typed into WHERE/INSERT clauses. Logging that this
	// happened, by whom, and its shape is enough of an audit trail
	// without risking sensitive data ending up in the API's own logs.
	s.logger.Info("query executed",
		"instanceId", id,
		"userId", userID,
		"rowCount", result.RowCount,
		"durationMs", result.DurationMs,
	)

	return result, nil
}

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// ListInstances retrieves a page of instances belonging to userID.
// Pagination is cursor-based (see ports.InstanceClient.ListInstances) since
// that's the only pagination primitive the underlying Kubernetes List API
// supports - there's no "skip N" offset available to build page numbers on
// top of.
func (s *instanceService) ListInstances(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) (*domain.ListInstancesResult, error) {
	switch {
	case limit <= 0:
		limit = defaultListLimit
	case limit > maxListLimit:
		limit = maxListLimit
	}

	instances, nextCursor, err := s.client.ListInstances(ctx, userID, limit, cursor)
	if err != nil {
		return nil, err
	}

	return &domain.ListInstancesResult{
		Instances:  instances,
		NextCursor: nextCursor,
	}, nil
}

// UpdateInstance updates an instance.
func (s *instanceService) UpdateInstance(
	ctx context.Context,
	input domain.UpdateInstanceInput,
) (*domain.Instance, error) {
	if err := validateUpdateInstance(input); err != nil {
		return nil, err
	}

	return s.client.UpdateInstance(ctx, input)
}

func validateUpdateInstance(
	input domain.UpdateInstanceInput,
) error {
	if strings.TrimSpace(input.ID) == "" {
		return ErrInvalidInput
	}
	if input.Version == nil && input.Storage == nil {
		return ErrNoUpdateFields
	}
	if input.Version != nil && (*input.Version < 14 || *input.Version > 16) {
		return ErrInvalidInput
	}
	if input.Storage != nil && *input.Storage == "" {
		return ErrInvalidInput
	}

	return nil
}

func (s *instanceService) DeleteInstance(ctx context.Context, id, userID string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}

	return s.client.DeleteInstance(ctx, id, userID)
}

const (
	defaultAuditLogLimit = 50
	maxAuditLogLimit     = 200
	auditLogLookback     = 168 * time.Hour // 7 days, matches your Loki retention_period
)

// GetAuditLogs retrieves a page of audit log entries, newest first.
// Pagination is cursor-based, matching ListInstances: Loki has no
// offset/skip concept either, only a time-range window. The cursor is the
// oldest entry's timestamp from the previous page (Loki's own
// nanosecond-epoch string format), used as the next page's end boundary -
// Loki's end is exclusive, so the boundary entry itself is never
// re-returned.
func (s *instanceService) GetAuditLogs(
	ctx context.Context,
	resourceID string,
	limit int,
	cursor string,
) ([]loki.LogLine, string, error) {
	switch {
	case limit <= 0:
		limit = defaultAuditLogLimit
	case limit > maxAuditLogLimit:
		limit = maxAuditLogLimit
	}

	end := time.Now()
	if cursor != "" {
		ns, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, "", ErrInvalidInput
		}
		end = time.Unix(0, ns)
	}

	logql := `{namespace="dbaas-api"} | json | component="audit"`
	if resourceID != "" {
		logql = fmt.Sprintf(`{namespace="dbaas-api"} | json | component="audit" | resource_id=%q`, resourceID)
	}

	logs, err := s.loki.Query(ctx, logql, limit, auditLogLookback, end)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(logs) == limit {
		// A full page - there may be more before the oldest entry we
		// got back. If this page came back short, we've reached
		// either the start of the lookback window or true history.
		nextCursor = logs[len(logs)-1].Timestamp
	}

	return logs, nextCursor, nil
}
