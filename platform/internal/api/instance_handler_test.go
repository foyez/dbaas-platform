package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/foyez/dbaas-platform/platform/internal/api/mocks"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/testify/v2/require"
	"github.com/stretchr/testify/mock"
)

func TestCreateInstance_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.MockInstanceService)

	log := logger.New("debug", "test")
	handler := NewInstanceHandler(svc, log)

	router := gin.New()
	router.POST("/api/v1/instances", handler.CreateInstance)

	svc.On("CreateInstance", mock.Anything, mock.Anything).Return(&domain.CreateInstanceResult{
		Instance: &domain.Instance{ID: "123", Name: "demo"},
		Replayed: false,
	}, nil)

	body := `{
		"name": "demo",
		"version": 15,
		"storage": "5Gi",
		"username": "postgres"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/instances",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "abc-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
	svc.AssertExpectations(t)
}

func TestCreateInstance_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.MockInstanceService)

	log := logger.New("debug", "test")
	handler := NewInstanceHandler(svc, log)

	router := gin.New()
	router.POST("/api/v1/instances", handler.CreateInstance)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/instances",
		strings.NewReader("{"),
	)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateInstance")
}

func TestCreateInstance_MissingIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.MockInstanceService)

	log := logger.New("debug", "test")
	handler := NewInstanceHandler(svc, log)

	router := gin.New()
	router.POST("/api/v1/instances", handler.CreateInstance)

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

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertNotCalled(t, "CreateInstance")
}
