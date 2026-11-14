package dbaas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors for the API's error codes. An *APIError matches one with
// errors.Is, so callers don't have to compare codes or status codes:
//
//	if errors.Is(err, dbaas.ErrNotFound) { ... }
var (
	ErrInvalidRequest        = errors.New("dbaas: invalid request")
	ErrUnauthorized          = errors.New("dbaas: unauthorized")
	ErrForbidden             = errors.New("dbaas: forbidden")
	ErrNotFound              = errors.New("dbaas: not found")
	ErrAlreadyExists         = errors.New("dbaas: instance already exists")
	ErrInstanceNotReady      = errors.New("dbaas: instance not ready")
	ErrMissingIdempotencyKey = errors.New("dbaas: missing idempotency key")
	ErrNoUpdateFields        = errors.New("dbaas: no update fields")
	ErrQueryFailed           = errors.New("dbaas: query failed")
	ErrQueryTimeout          = errors.New("dbaas: query timed out")

	// ErrInstanceFailed is returned by WaitUntilRunning when the instance
	// ends up Failed instead of Running. It doesn't come from the API.
	ErrInstanceFailed = errors.New("dbaas: instance failed")
)

// codeErrors maps the API's stable error codes (platform/internal/httpx) to
// the sentinels above.
var codeErrors = map[string]error{
	"INVALID_REQUEST":         ErrInvalidRequest,
	"UNAUTHORIZED":            ErrUnauthorized,
	"FORBIDDEN":               ErrForbidden,
	"NOT_FOUND":               ErrNotFound,
	"INSTANCE_NOT_FOUND":      ErrNotFound,
	"INSTANCE_ALREADY_EXISTS": ErrAlreadyExists,
	"INSTANCE_NOT_READY":      ErrInstanceNotReady,
	"MISSING_IDEMPOTENCY_KEY": ErrMissingIdempotencyKey,
	"NO_UPDATE_FIELDS":        ErrNoUpdateFields,
	"QUERY_FAILED":            ErrQueryFailed,
	"QUERY_TIMEOUT":           ErrQueryTimeout,
}

// statusErrors is the fallback for responses without an error code, such as
// the 401 and 403 the auth middleware writes before the handlers run.
var statusErrors = map[int]error{
	http.StatusUnauthorized: ErrUnauthorized,
	http.StatusForbidden:    ErrForbidden,
	http.StatusNotFound:     ErrNotFound,
}

// APIError is a non-2xx response from the API.
type APIError struct {
	// StatusCode is the HTTP status code, e.g. 404.
	StatusCode int
	// Code is the API's error code, e.g. "NOT_FOUND". Empty when the
	// response wasn't in the API's error format.
	Code string
	// Message is the human-readable description from the API, or the raw
	// response body when it wasn't in the API's error format.
	Message string
	// Details is extra, code-specific data. For QUERY_FAILED, decode it
	// with QueryDetails.
	Details json.RawMessage
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("dbaas: %d %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("dbaas: %d: %s", e.StatusCode, e.Message)
}

// Is lets errors.Is match an APIError against the sentinel for its code,
// falling back to its status code.
func (e *APIError) Is(target error) bool {
	if sentinel, ok := codeErrors[e.Code]; ok {
		return sentinel == target
	}
	return statusErrors[e.StatusCode] == target
}

// QueryErrorDetails is PostgreSQL's own description of why it rejected a
// statement sent with InstancesService.Query.
type QueryErrorDetails struct {
	// SQLState is the PostgreSQL error code, e.g. "42P01" (undefined table).
	SQLState string `json:"sqlState"`
	Detail   string `json:"detail,omitempty"`
	Hint     string `json:"hint,omitempty"`
	// Position is the 1-based character offset of the error in the
	// statement, or 0 when PostgreSQL doesn't report one.
	Position int `json:"position,omitempty"`
}

// QueryDetails returns the details of a QUERY_FAILED error, and false for
// any other error.
func (e *APIError) QueryDetails() (*QueryErrorDetails, bool) {
	if e.Code != "QUERY_FAILED" || len(e.Details) == 0 {
		return nil, false
	}

	var d QueryErrorDetails
	if err := json.Unmarshal(e.Details, &d); err != nil {
		return nil, false
	}

	return &d, true
}

// errorResponse is the API's error envelope: {"error": {...}}.
type errorResponse struct {
	Error struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Details json.RawMessage `json:"details"`
	} `json:"error"`
}

// newAPIError builds an APIError from a non-2xx response body, whether or
// not the body is in the API's error format.
func newAPIError(status int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: status}

	var envelope errorResponse
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Code != "" {
		apiErr.Code = envelope.Error.Code
		apiErr.Message = envelope.Error.Message
		apiErr.Details = envelope.Error.Details
		return apiErr
	}

	apiErr.Message = strings.TrimSpace(string(body))
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(status)
	}

	return apiErr
}
