package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/domain/mocks"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/testify/v2/require"
	"github.com/stretchr/testify/mock"
)

func TestCreateInstance_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.MockInstanceService)

	log := logger.New("debug", "test")
	handler := handler.NewInstanceHandler(svc, log)

	router := gin.New()
	router.POST("/api/v1/instances", handler.CreateInstance)

	svc.On(
		"CreateInstance",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("internal"))

	body := `{
        "name":"demo",
        "version":16,
        "storage":"10Gi",
        "username":"postgres"
    }`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/instances",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "abc")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
