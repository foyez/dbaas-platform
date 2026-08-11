// Package handler contains HTTP handlers responsible for handling incoming
// API requests, validating DTOs, calling application services, and
// returning HTTP responses.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/foyez/dbaas-platform/platform/internal/api"
	"github.com/foyez/dbaas-platform/platform/internal/auth"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/httpx"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/gin-gonic/gin"
)

type InstanceHandler struct {
	svc    domain.InstanceService
	logger *slog.Logger
	authmw *auth.Middleware
}

func NewInstanceHandler(
	svc domain.InstanceService,
	logger *slog.Logger,
	authmw *auth.Middleware,
) *InstanceHandler {
	return &InstanceHandler{
		svc:    svc,
		logger: logger,
		authmw: authmw,
	}
}

// CreateInstance handles POST /v1/instances requests.
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var req api.CreateInstanceRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		h.logger.Warn(
			"invalid create isntance request",
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
		Name:     req.Name,
		Version:  req.Version,
		Storage:  req.Storage,
		Username: req.Username,
		UserID:   userID,
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
		h.logger.Error(
			"failed to create instance",
			"error", err,
			"name", req.Name,
		)

		httpx.RespondError(c, err)
		return
	}

	instance := result.Instance

	// if result.Replayed {
	// 	h.logger.Info(
	// 		"idemptent replayed detected",
	// 		"id", instance.ID,
	// 		"name", instance.Name,
	// 	)
	//
	// 	httpx.JSON(c, http.StatusOK, result.Instance)
	// 	return
	// }

	resp := api.InstanceResponse{
		ID:        instance.ID,
		Name:      instance.Name,
		Version:   input.Version,
		Storage:   input.Storage,
		Status:    instance.Status,
		CreatedAt: instance.CreatedAt,
	}

	h.logger.Info(
		"instance created",
		"id", instance.ID,
		"name", instance.Name,
		"replayed", result.Replayed,
	)

	httpx.JSON(c, http.StatusAccepted, resp)
}

func (h *InstanceHandler) ListInstances(c *gin.Context) {
	userID := h.authmw.UserID(c.Request.Context())
	result, err := h.svc.ListInstances(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error(
			"failed to list instances",
			"error", err,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.ListInstancesResponse{
		Items: make([]api.InstanceResponse, 0, len(result.Instances)),
	}

	for _, inst := range result.Instances {
		resp.Items = append(resp.Items, api.InstanceResponse{
			ID:        inst.ID,
			Name:      inst.Name,
			Version:   inst.Version,
			Storage:   inst.Storage,
			Status:    inst.Status,
			CreatedAt: inst.CreatedAt,
		})
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// GetInstance handles GET /v1/instances/:id requests.
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn("invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	userID := h.authmw.UserID(c.Request.Context())
	instance, err := h.svc.GetInstance(c.Request.Context(), id, userID)
	if err != nil {
		h.logger.Error(
			"failed to get instance",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.InstanceResponse{
		ID:        instance.ID,
		Name:      instance.Name,
		Version:   instance.Version,
		Storage:   instance.Storage,
		Status:    instance.Status,
		CreatedAt: instance.CreatedAt,
	}

	httpx.JSON(c, http.StatusOK, resp)
}

// UpdateInstance handles PATCH /v1/instances/:id requests.
func (h *InstanceHandler) UpdateInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn("invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	var req api.UpdateInstanceRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		h.logger.Warn(
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
	}

	instance, err := h.svc.UpdateInstance(c.Request.Context(), input)
	if err != nil {
		h.logger.Error(
			"failed to update instance",
			"error", err,
			"id", id,
		)

		httpx.RespondError(c, err)
		return
	}

	resp := api.InstanceResponse{
		ID:        instance.ID,
		Name:      instance.Name,
		Version:   instance.Version,
		Storage:   instance.Storage,
		Status:    instance.Status,
		CreatedAt: instance.CreatedAt,
	}

	h.logger.Info(
		"instance updated",
		"id", instance.ID,
	)

	httpx.JSON(c, http.StatusAccepted, resp)
}

// DeleteInstance handles DELETE /v1/instances/:id requests.
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Warn("invalid instance id")

		httpx.RespondError(c, service.ErrInvalidInput)
		return
	}

	if err := h.svc.DeleteInstance(c, id); err != nil {
		h.logger.Error(
			"failed to delete instance",
			"error", err,
		)

		httpx.RespondError(c, err)
		return
	}

	httpx.JSON(c, http.StatusAccepted, api.DeleteInstanceResponse{
		Message: "Deletetion initiated",
		Status:  "Deleting",
	})
}
