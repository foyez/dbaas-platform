package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/foyez/dbaas-platform/platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   ErrorCode
	}{
		{"not found", service.ErrNotFound, http.StatusNotFound, CodeNotFound},
		{"invalid input", service.ErrInvalidInput, http.StatusBadRequest, CodeInvalidRequest},
		{"wrapped invalid input", fmt.Errorf("parsing cursor: %w", service.ErrInvalidInput), http.StatusBadRequest, CodeInvalidRequest},
		{"missing idempotency key", service.ErrMissingIdempotencyKey, http.StatusBadRequest, CodeMissingIdempotencyKey},
		{"no update fields", service.ErrNoUpdateFields, http.StatusBadRequest, CodeNoUpdateFields},
		{"not ready", service.ErrInstanceNotReady, http.StatusConflict, CodeInstanceNotReady},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondError(c, tt.err)

			assert.Equal(t, tt.wantStatus, w.Code)
			var body ErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, tt.wantCode, body.Error.Code)
		})
	}
}
