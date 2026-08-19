package audit

import (
	"context"
	"log/slog"
)

type AuditLogger struct {
	log *slog.Logger
}

func NewLogger(base *slog.Logger) *AuditLogger {
	// dedicated child logger, always tagged so it's filterable in Loki
	return &AuditLogger{log: base.With("component", "audit")}
}

type Event struct {
	ActorID      string
	Action       string // e.g. "instance.create", "instance.delete"
	ResourceType string // e.g. "instance"
	ResourceID   string
	Status       int
}

func (a *AuditLogger) Record(
	ctx context.Context,
	e Event,
) {
	a.log.Info("audit_event",
		"actor_id", e.ActorID,
		"action", e.Action,
		"resource_type", e.ResourceType,
		"resource_id", e.ResourceID,
		"status", e.Status,
	)
}
