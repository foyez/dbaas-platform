package audit

import (
	"context"

	"github.com/gin-gonic/gin"
)

// routeActions maps "METHOD route" to a human-readable audit action name.
// Only routes listed here are audited
var routeActions = map[string]string{
	"POST /api/v1/instances":                "instance.create",
	"PATCH /api/v1/instances/:id":           "instance.update",
	"DELETE /api/v1/instances/:id":          "instance.delete",
	"GET /api/v1/instances/:id/credentials": "instance.credentials.access",
}

const ResourceType = "instance"

// Middleware records an audit event for every request matching routeActions.
func Middleware(logger *AuditLogger, userID func(ctx context.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		route := c.FullPath()
		if c.Request.Method == "" {
			return // unmatched route
		}

		action, ok := routeActions[c.Request.Method+" "+route]
		if !ok {
			return
		}

		logger.Record(c.Request.Context(),
			Event{
				ActorID:      userID(c.Request.Context()),
				Action:       action,
				ResourceType: ResourceType,
				ResourceID:   c.Param("id"),
				Status:       c.Writer.Status(),
			})
	}
}
