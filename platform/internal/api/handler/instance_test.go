// Package handler_test exercises InstanceHandler through the real router.New wiring
package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/foyez/dbaas-platform/platform/internal/api"
	"github.com/foyez/dbaas-platform/platform/internal/api/handler"
	"github.com/foyez/dbaas-platform/platform/internal/api/router"
	"github.com/foyez/dbaas-platform/platform/internal/domain"
	"github.com/foyez/dbaas-platform/platform/internal/domain/mocks"
	"github.com/foyez/dbaas-platform/platform/internal/logger"
	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/testify/v2/require"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestServer wires a mocked domain.InstanceService into the actual
// application router
func newTestServer(svc domain.InstanceService) *gin.Engine {
	log := logger.New("debug", "test")
	return router.New(handler.NewInstanceHandler(svc, log))
}

func doRequest(
	t *testing.T,
	engine *gin.Engine,
	method, path, body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	return w
}

// POST /api/v1/instances
func TestCreateInstance(t *testing.T) {
	validBody := `{"name":"demo","version":16,"storage":"10Gi","username":"postgres"}`
	idempotencyHeaders := map[string]string{"Idempotency-Key": "key-1"}

	tests := []struct {
		name       string
		body       string
		headers    map[string]string
		mockSetup  func(svc *mocks.MockInstanceService)
		wantStatus int
		wantCalled bool
	}{
		{
			name:    "success",
			body:    validBody,
			headers: idempotencyHeaders,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("CreateInstance", mock.Anything, mock.Anything).
					Return(&domain.CreateInstanceResult{
						Instance: &domain.Instance{ID: "123", Name: "demo"},
					}, nil)
			},
			wantStatus: http.StatusAccepted,
			wantCalled: true,
		},
		{
			name:       "invalid json",
			body:       "{",
			headers:    idempotencyHeaders,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing required field",
			body:       `{"version":16,"storage":"10Gi"}`,
			headers:    idempotencyHeaders,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "version outside allowed set",
			body:       `{"name":"demo","version":38,"storage":"10Gi","username":"postgres"}`,
			headers:    idempotencyHeaders,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing idempotency key",
			body:       validBody,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "name already exists",
			body:    validBody,
			headers: idempotencyHeaders,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("CreateInstance", mock.Anything, mock.Anything).
					Return(nil, service.ErrAlreadyExists)
			},
			wantStatus: http.StatusConflict,
			wantCalled: true,
		},
		{
			name:    "unexpected service error",
			body:    validBody,
			headers: idempotencyHeaders,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("CreateInstance", mock.Anything, mock.Anything).
					Return(nil, errors.New("unknown"))
			},
			wantStatus: http.StatusInternalServerError,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockInstanceService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			w := doRequest(t, newTestServer(svc), http.MethodPost, "/api/v1/instances", tt.body, tt.headers)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCalled {
				svc.AssertExpectations(t)
			} else {
				svc.AssertNotCalled(t, "CreateInstance", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestCreateInstance_ResponseBody(t *testing.T) {
	svc := new(mocks.MockInstanceService)
	svc.On("CreateInstance", mock.Anything, mock.Anything).
		Return(&domain.CreateInstanceResult{
			Instance: &domain.Instance{ID: "123", Name: "demo", Status: "Pending"},
		}, nil)

	body := `{"name":"demo","version":16,"storage":"10Gi","username":"postgres"}`
	w := doRequest(t, newTestServer(svc), http.MethodPost, "/api/v1/instances", body,
		map[string]string{"Idempotency-Key": "key-1"},
	)

	require.Equal(t, http.StatusAccepted, w.Code)

	var resp api.InstanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "123", resp.ID)
	require.Equal(t, "demo", resp.Name)
	require.Equal(t, domain.InstanceStatusPending, resp.Status)
}

// GET /api/v1/instances/:id
func TestGetInstance(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockSetup  func(svc *mocks.MockInstanceService)
		wantStatus int
	}{
		{
			name: "success",
			id:   "123",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("GetInstance", mock.Anything, "123").
					Return(&domain.Instance{ID: "123", Name: "demo"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "missing",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("GetInstance", mock.Anything, "missing").
					Return(nil, service.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "unexpected service error",
			id:   "123",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("GetInstance", mock.Anything, "123").
					Return(nil, errors.New("error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockInstanceService)
			tt.mockSetup(svc)

			w := doRequest(t, newTestServer(svc), http.MethodGet, "/api/v1/instances/"+tt.id, "", nil)

			require.Equal(t, tt.wantStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestGetInstance_ResponseBody(t *testing.T) {
	svc := new(mocks.MockInstanceService)
	svc.On("GetInstance", mock.Anything, "123").
		Return(&domain.Instance{
			ID:      "123",
			Name:    "demo",
			Version: 15,
			Storage: "5Gi",
			Status:  domain.InstanceStatusRunning,
		}, nil)

	w := doRequest(t, newTestServer(svc), http.MethodGet, "/api/v1/instances/123", "", nil)
	require.Equal(t, http.StatusOK, w.Code)

	var resp api.InstanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "123", resp.ID)
	require.Equal(t, 15, resp.Version)
	require.Equal(t, "5Gi", resp.Storage)
	require.Equal(t, domain.InstanceStatusRunning, resp.Status)
}

// GET /api/v1/instances
func TestListInstances(t *testing.T) {
	tests := []struct {
		name       string
		mockSetup  func(svc *mocks.MockInstanceService)
		wantStatus int
		wantItems  int
	}{
		{
			name: "success with items",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("ListInstances", mock.Anything).Return(&domain.ListInstancesResult{
					Instances: []*domain.Instance{
						{ID: "1", Name: "demo-1"},
						{ID: "2", Name: "demo-2"},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantItems:  2,
		},
		{
			name: "success with no items",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("ListInstances", mock.Anything).
					Return(&domain.ListInstancesResult{Instances: nil}, nil)
			},
			wantStatus: http.StatusOK,
			wantItems:  0,
		},
		{
			name: "unexpected service error",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("ListInstances", mock.Anything).
					Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockInstanceService)
			tt.mockSetup(svc)

			w := doRequest(t, newTestServer(svc), http.MethodGet, "/api/v1/instances", "", nil)

			require.Equal(t, tt.wantStatus, w.Code)
			svc.AssertExpectations(t)

			if tt.wantStatus == http.StatusOK {
				var resp api.ListInstancesResponse
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				require.Len(t, resp.Items, tt.wantItems)
			}
		})
	}
}

// PATCH /api/v1/instances/:id
func TestUpdateInstance(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		mockSetup  func(svc *mocks.MockInstanceService)
		wantStatus int
		wantCalled bool
	}{
		{
			name: "success",
			id:   "123",
			body: `{"storage":"7Gi"}`,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(&domain.Instance{ID: "123", Storage: "7Gi"}, nil)
			},
			wantStatus: http.StatusAccepted,
			wantCalled: true,
		},
		{
			name:       "invalid version",
			id:         "123",
			body:       `{"version":99}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "no fields provided",
			id:   "123",
			body: `{}`,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(nil, service.ErrNoUpdateFields)
			},
			wantStatus: http.StatusBadRequest,
			wantCalled: true,
		},
		{
			name: "not found",
			id:   "missing",
			body: `{"storage":"20Gi"}`,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(nil, service.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantCalled: true,
		},
		{
			name: "unexpected service error",
			id:   "123",
			body: `{"storage":"20Gi"}`,
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("UpdateInstance", mock.Anything, mock.Anything).
					Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockInstanceService)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			w := doRequest(t, newTestServer(svc), http.MethodPatch, "/api/v1/instances/"+tt.id, tt.body, nil)

			require.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCalled {
				svc.AssertExpectations(t)
			} else {
				svc.AssertNotCalled(t, "UpdateInstance", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestUpdateInstance_ResponseBody(t *testing.T) {
	svc := new(mocks.MockInstanceService)
	svc.On("UpdateInstance", mock.Anything, mock.Anything).Return(&domain.Instance{
		ID:      "123",
		Storage: "20Gi",
	}, nil)

	w := doRequest(
		t, newTestServer(svc), http.MethodPatch,
		"/api/v1/instances/123", `{"storage":"20Gi"}`, nil,
	)

	require.Equal(t, http.StatusAccepted, w.Code)

	var resp api.InstanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "20Gi", resp.Storage)
}

// DELETE /api/v1/instances/:id
func TestDeleteInstance(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mockSetup  func(svc *mocks.MockInstanceService)
		wantStatus int
	}{
		{
			name: "success",
			id:   "123",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("DeleteInstance", mock.Anything, "123").Return(nil)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "not found",
			id:   "missing",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("DeleteInstance", mock.Anything, "missing").
					Return(service.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "unexpected service error",
			id:   "123",
			mockSetup: func(svc *mocks.MockInstanceService) {
				svc.On("DeleteInstance", mock.Anything, "123").
					Return(errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mocks.MockInstanceService)
			tt.mockSetup(svc)

			w := doRequest(t, newTestServer(svc), http.MethodDelete, "/api/v1/instances/"+tt.id, "", nil)

			require.Equal(t, tt.wantStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestDeleteInstance_ResponseBody(t *testing.T) {
	svc := new(mocks.MockInstanceService)
	svc.On("DeleteInstance", mock.Anything, "123").Return(nil)

	w := doRequest(t, newTestServer(svc), http.MethodDelete, "/api/v1/instances/123", "", nil)
	require.Equal(t, http.StatusAccepted, w.Code)

	var resp api.DeleteInstanceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "Deleting", resp.Status)
}
