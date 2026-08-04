// Package api contains HTTP handlers responsible for handling incoming
// API requests, validating DTOs, calling application services, and
// returning HTTP responses.
package api

import (
	"log/slog"
	"net/http"

	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/httpx"
	"github.com/gin-gonic/gin"
)

type InstanceHandler struct {
	svc    domain.InstanceService
	logger *slog.Logger
}

func NewInstanceHandler(
	svc domain.InstanceService,
	logger *slog.Logger,
) *InstanceHandler {
	return &InstanceHandler{
		svc:    svc,
		logger: logger,
	}
}

// CreateInstance handles POST /v1/instances requests.
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var req CreateInstanceRequest

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

	input := domain.CreateInstanceInput{
		Name:           req.Name,
		Version:        req.Version,
		Storage:        req.Storage,
		Username:       req.Username,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	}

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

	resp := InstanceResponse{
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
