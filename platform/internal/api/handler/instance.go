// Package handler contains HTTP handlers responsible for handling incoming
// API requests, validating DTOs, calling application services, and
// returning HTTP responses.
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/foyez/dbaas-platform/platform/internal/api"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/httpx"
	"github.com/foyez/dbaas-platform/platform/internal/infra/loki"
	"github.com/foyez/dbaas-platform/platform/internal/observability"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/gin-gonic/gin"
)

type InstanceHandler struct {
	svc     domain.InstanceService
	logger  *slog.Logger
	authmw  *auth.Middleware
	metrics *observability.Metrics
}

func NewInstanceHandler(
	svc domain.InstanceService,
	logger *slog.Logger,
	authmw *auth.Middleware,
	metrics *observability.Metrics,
) *InstanceHandler {
	return &InstanceHandler{
		svc:     svc,
		logger:  logger,
		authmw:  authmw,
		metrics: metrics,
	}
}

// CreateInstance handles POST /v1/instances requests.
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var req api.CreateInstanceRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		h.logger.WarnContext(
			c.Request.Context(),
			"invalid create instance request",
			"error", err,
		)

		httpx.Error(
			c,
			http.StatusBadRequest,
			httpx.CodeInvalidRequest,
			err.Error(),
		)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())

	input := domain.CreateInstanceInput{
		Name:      req.Name,
		Instances: req.Instances,
		Version:   req.Version,
		Storage:   req.Storage,
		Username:  req.Username,
		UserID:    userID,
	}

	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		httpx.RespondError(c, service.ErrMissingIdempotencyKey)
		return
	}

	input.IdempotencyKey = key

	result, err := h.svc.CreateInstance(
		c.Request.Context(),
		input,
	)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to create instance",
			"error", err,
			"name", req.Name,
		)

		httpx.RespondError(c, err)
		return
	}

	instance := result.Instance

	resp := api.InstanceResponse{
		ID:             instance.ID,
		Name:           instance.Name,
		Database:       instance.Database,
		Version:        instance.Version,
		Storage:        instance.Storage,
		Instances:      instance.Instances,
		ReadyInstances: instance.ReadyInstances,
		Status:         instance.Status,
		CreatedAt:      instance.CreatedAt,
		ExternalPort:   instance.ExternalPort,
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"instance created",
		"id", instance.ID,
		"name", instance.Name,
		"instances", instance.Instances,
		"replayed", result.Replayed,
	)

	httpx.JSON(c, http.StatusAccepted, resp)
}

func (h *InstanceHandler) ListInstances(c *gin.Context) {
	userID := h.authmw.UserID(c.Request.Context())

	limit, _ := strconv.Atoi(c.Query("limit"))
	cursor := c.Query("cursor")

	result, err := h.svc.ListInstances(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to list instances",
			"error", err,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.ListInstancesResponse{
		Items:      make([]api.InstanceResponse, 0, len(result.Instances)),
		NextCursor: result.NextCursor,
	}

	for _, instance := range result.Instances {
		resp.Items = append(resp.Items, api.InstanceResponse{
			ID:             instance.ID,
			Name:           instance.Name,
			Database:       instance.Database,
			Version:        instance.Version,
			Storage:        instance.Storage,
			Instances:      instance.Instances,
			ReadyInstances: instance.ReadyInstances,
			Status:         instance.Status,
			CreatedAt:      instance.CreatedAt,
			ExternalPort:   instance.ExternalPort,
		})
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// GetInstance handles GET /v1/instances/:id requests.
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())
	instance, err := h.svc.GetInstance(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to get instance",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.InstanceResponse{
		ID:             instance.ID,
		Name:           instance.Name,
		Database:       instance.Database,
		Version:        instance.Version,
		Storage:        instance.Storage,
		Instances:      instance.Instances,
		ReadyInstances: instance.ReadyInstances,
		Status:         instance.Status,
		CreatedAt:      instance.CreatedAt,
		ExternalPort:   instance.ExternalPort,
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// GetInstanceLogs handles GET /v1/instances/:id/logs requests.
func (h *InstanceHandler) GetInstanceLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())
	logs, err := h.svc.GetInstanceLogs(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to get instance logs",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	if logs == nil {
		logs = []loki.LogLine{}
	}

	resp := api.InstanceLogsResponse{
		InstanceID: id,
		Logs:       logs,
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// GetCredentials handles GET /v1/instances/:id/credentials requests.
func (h *InstanceHandler) GetCredentials(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())

	credentials, err := h.svc.GetCredentials(
		c.Request.Context(),
		id,
		userID,
	)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to get instance credentials",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.InstanceCredentialsResponse{
		Host:     credentials.Host,
		Port:     credentials.Port,
		Database: credentials.Database,
		Username: credentials.Username,
		Password: credentials.Password,
		URI:      credentials.URI,
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// UpdateInstance handles PATCH /v1/instances/:id requests.
func (h *InstanceHandler) UpdateInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())

	var req api.UpdateInstanceRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		h.logger.WarnContext(
			c.Request.Context(),
			"invalid update instance request",
			"error", err,
		)

		httpx.Error(
			c,
			http.StatusBadRequest,
			httpx.CodeInvalidRequest,
			err.Error(),
		)
		return
	}

	input := domain.UpdateInstanceInput{
		ID:      id,
		Version: req.Version,
		Storage: req.Storage,
		UserID:  userID,
	}

	instance, err := h.svc.UpdateInstance(c.Request.Context(), input)
	if err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to update instance",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.InstanceResponse{
		ID:             instance.ID,
		Name:           instance.Name,
		Database:       instance.Database,
		Version:        instance.Version,
		Storage:        instance.Storage,
		Instances:      instance.Instances,
		ReadyInstances: instance.ReadyInstances,
		Status:         instance.Status,
		CreatedAt:      instance.CreatedAt,
		ExternalPort:   instance.ExternalPort,
	}

	h.logger.InfoContext(
		c.Request.Context(),
		"instance updated",
		"id", instance.ID,
	)

	httpx.JSON(c, http.StatusAccepted, resp)
}

// DeleteInstance handles DELETE /v1/instances/:id requests.
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())

	if err := h.svc.DeleteInstance(c, id, userID); err != nil {
		h.logger.ErrorContext(
			c.Request.Context(),
			"failed to delete instance",
			"error", err,
		)

		httpx.RespondError(c, err)
		return
	}

	httpx.JSON(c, http.StatusAccepted, api.DeleteInstanceResponse{
		Message: "Deletion initiated",
		Status:  "Deleting",
	})
}

// GetAuditLogs handles GET /v1/audit-logs requests (admin only).
func (h *InstanceHandler) GetAuditLogs(c *gin.Context) {
	resourceID := c.Query("resourceId")
	limit, _ := strconv.Atoi(c.Query("limit"))
	cursor := c.Query("cursor")

	logs, nextCursor, err := h.svc.GetAuditLogs(c.Request.Context(), resourceID, limit, cursor)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "failed to get audit logs", "error", err)

		httpx.RespondError(c, err)
		return
	}

	httpx.JSON(c, http.StatusOK, api.AuditLogsResponse{Logs: logs, NextCursor: nextCursor})
}

// ExecuteQuery handles POST /v1/instances/:id/query requests.
func (h *InstanceHandler) ExecuteQuery(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.WarnContext(c.Request.Context(), "invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	var req api.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())

	result, err := h.svc.ExecuteQuery(c.Request.Context(), id, userID, req.SQL)
	if err != nil {
		// A rejected statement is the user's mistake, not an API failure,
		// and its message can contain row values: log only the SQLSTATE.
		ctx := c.Request.Context()
		var queryErr *domain.QueryError
		switch {
		case errors.As(err, &queryErr):
			h.metrics.SQLQueriesTotal.WithLabelValues("query_failed").Inc()
			h.logger.InfoContext(ctx, "query rejected by database", "id", id, "sqlState", queryErr.SQLState)
		case errors.Is(err, service.ErrQueryTimeout):
			h.metrics.SQLQueriesTotal.WithLabelValues("timeout").Inc()
			h.logger.InfoContext(ctx, "query timed out", "id", id)
		case errors.Is(err, service.ErrInstanceNotReady):
			h.metrics.SQLQueriesTotal.WithLabelValues("not_ready").Inc()
		default:
			h.metrics.SQLQueriesTotal.WithLabelValues("error").Inc()
			h.logger.ErrorContext(ctx, "failed to execute query", "error", err, "id", id)
		}

		httpx.RespondError(c, err)
		return
	}

	h.metrics.SQLQueriesTotal.WithLabelValues("ok").Inc()
	httpx.JSON(c, http.StatusOK, api.QueryResponse{
		Columns:    result.Columns,
		Rows:       result.Rows,
		RowCount:   result.RowCount,
		Truncated:  result.Truncated,
		DurationMs: result.DurationMs,
	})
}
