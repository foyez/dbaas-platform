package dbaas

import (
	"context"
	"fmt"
	"net/http"
)

// AuditLogsOptions selects one page of the audit trail.
type AuditLogsOptions struct {
	// ResourceID limits the trail to one instance. Empty means all.
	ResourceID string
	// Limit is the page size. Zero means the server's default (50); the
	// server caps it at 200.
	Limit int
	// Cursor is the NextCursor of the previous page; empty for the first.
	Cursor string
}

// AuditLogList is one page of the audit trail, newest first.
type AuditLogList struct {
	// Logs are JSON-encoded audit events.
	Logs []LogLine `json:"logs"`
	// NextCursor fetches the next page. Empty on the last page.
	NextCursor string `json:"nextCursor,omitempty"`
}

// AuditLogs returns one page of the audit trail: who changed what, and
// when. It needs the admin role, and fails with ErrForbidden without it.
func (c *Client) AuditLogs(ctx context.Context, opts AuditLogsOptions) (*AuditLogList, error) {
	var out AuditLogList

	q := pageQuery(opts.Limit, opts.Cursor)
	if opts.ResourceID != "" {
		q.Set("resourceId", opts.ResourceID)
	}

	if err := c.do(ctx, request{method: http.MethodGet, path: "/api/v1/audit-logs", query: q, out: &out}); err != nil {
		return nil, fmt.Errorf("get audit logs: %w", err)
	}

	return &out, nil
}
