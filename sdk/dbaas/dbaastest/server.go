// Package dbaastest provides an in-memory fake of the DBaaS Platform API, for
// testing code that uses package dbaas without a cluster.
//
//	srv := dbaastest.NewServer()
//	defer srv.Close()
//
//	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
//
// The fake follows the real API's contract (paths, status codes, error
// codes, idempotent create, cursor pagination), with simplified behaviour:
//
//   - A new or updated instance is Pending, and becomes Running the next
//     time it is fetched with GET /instances/{id}.
//   - The SQL console understands only "SELECT 1"; anything else fails with
//     a syntax error, like PostgreSQL would.
//   - All instances belong to one user.
package dbaastest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
)

// Credentials the fake accepts.
const (
	// Token is a bearer token with the user role.
	Token = "test-token"
	// AdminToken is a bearer token with the admin role, needed for
	// AuditLogs.
	AdminToken = "test-admin-token"

	// ClientID and ClientSecret are the client credentials the token
	// endpoint (Server.TokenURL) exchanges for Token.
	ClientID     = "test-client"
	ClientSecret = "test-secret"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Server is a running fake API. Its URL is the API's base URL.
type Server struct {
	*httptest.Server

	// TokenURL is the fake OAuth2 token endpoint, for
	// dbaas.WithClientCredentials.
	TokenURL string

	mu        sync.Mutex
	instances map[string]*dbaas.Instance
	// order is the instance IDs in creation order, for List.
	order []string
	audit []dbaas.LogLine
}

// NewServer starts a fake API. Call Close when done.
func NewServer() *Server {
	s := &Server{instances: map[string]*dbaas.Instance{}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /oauth/v2/token", s.token)

	api := http.NewServeMux()
	api.HandleFunc("POST /api/v1/instances", s.create)
	api.HandleFunc("GET /api/v1/instances", s.list)
	api.HandleFunc("GET /api/v1/instances/{id}", s.get)
	api.HandleFunc("PATCH /api/v1/instances/{id}", s.update)
	api.HandleFunc("DELETE /api/v1/instances/{id}", s.delete)
	api.HandleFunc("GET /api/v1/instances/{id}/credentials", s.credentials)
	api.HandleFunc("GET /api/v1/instances/{id}/logs", s.logs)
	api.HandleFunc("POST /api/v1/instances/{id}/query", s.query)
	api.HandleFunc("GET /api/v1/audit-logs", s.auditLogs)
	mux.Handle("/api/v1/", requireToken(api))

	s.Server = httptest.NewServer(mux)
	s.TokenURL = s.URL + "/oauth/v2/token"

	return s
}

// SetStatus forces an instance's status, e.g. to StatusFailed to test
// error handling. It panics if the instance doesn't exist.
func (s *Server) SetStatus(id string, status dbaas.InstanceStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.instances[id].Status = status
}

// requireToken rejects requests without a known bearer token, like the
// real API's auth middleware: plain-text 401, no error envelope.
func requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer " + Token, "Bearer " + AdminToken:
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) token(w http.ResponseWriter, r *http.Request) {
	id, secret, ok := r.BasicAuth()
	if !ok {
		id, secret = r.PostFormValue("client_id"), r.PostFormValue("client_secret")
	}
	if id != ClientID || secret != ClientSecret {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": Token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		Version   int    `json:"version"`
		Storage   string `json:"storage"`
		Instances int    `json:"instances"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if len(req.Name) < 3 || len(req.Name) > 63 || req.Version < 14 || req.Version > 16 || req.Storage == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "name, version or storage is invalid")
		return
	}

	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "missing idempotency key")
		return
	}

	// Like the real API, the ID is derived from the key, so a retry with
	// the same key finds the instance it created.
	sum := sha256.Sum256([]byte(key))
	id := "instance-" + hex.EncodeToString(sum[:])[:12]

	s.mu.Lock()
	defer s.mu.Unlock()

	if inst, ok := s.instances[id]; ok {
		writeJSON(w, http.StatusAccepted, inst)
		return
	}

	if req.Instances == 0 {
		req.Instances = 1
	}
	inst := &dbaas.Instance{
		ID:        id,
		Name:      id,
		Database:  req.Name,
		Version:   req.Version,
		Storage:   req.Storage,
		Instances: req.Instances,
		Status:    dbaas.StatusPending,
		CreatedAt: time.Now().UTC(),
	}
	s.instances[id] = inst
	s.order = append(s.order, id)
	s.recordAudit("create", id)

	writeJSON(w, http.StatusAccepted, inst)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	switch {
	case limit <= 0:
		limit = defaultListLimit
	case limit > maxListLimit:
		limit = maxListLimit
	}
	// The cursor is an offset here; clients must treat it as opaque.
	offset, _ := strconv.Atoi(r.URL.Query().Get("cursor"))

	s.mu.Lock()
	defer s.mu.Unlock()

	end := min(offset+limit, len(s.order))
	page := s.order[min(offset, end):end]
	items := make([]*dbaas.Instance, 0, len(page))
	for _, id := range page {
		items = append(items, s.instances[id])
	}

	resp := dbaas.InstanceList{Items: items}
	if end < len(s.order) {
		resp.NextCursor = strconv.Itoa(end)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.lookup(w, r)
	if !ok {
		return
	}

	// Reply with the status as it was, then let the "operator" finish.
	resp := *inst
	if inst.Status == dbaas.StatusPending {
		inst.Status = dbaas.StatusRunning
		inst.ReadyInstances = inst.Instances
		inst.ExternalPort = 30001
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	var req dbaas.UpdateInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.Version == nil && req.Storage == nil {
		writeError(w, http.StatusBadRequest, "NO_UPDATE_FIELDS", "no fields to update")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.lookup(w, r)
	if !ok {
		return
	}

	if req.Version != nil {
		inst.Version = *req.Version
	}
	if req.Storage != nil {
		inst.Storage = *req.Storage
	}
	inst.Status = dbaas.StatusPending
	s.recordAudit("update", inst.ID)

	writeJSON(w, http.StatusAccepted, inst)
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.lookup(w, r)
	if !ok {
		return
	}

	delete(s.instances, inst.ID)
	for i, id := range s.order {
		if id == inst.ID {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	s.recordAudit("delete", inst.ID)

	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Deletion initiated", "status": "Deleting"})
}

func (s *Server) credentials(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.lookupRunning(w, r)
	if !ok {
		return
	}

	const host, user, password = "db.dbaas.localhost", "app", "secret"
	writeJSON(w, http.StatusOK, dbaas.InstanceCredentials{
		Host:     host,
		Port:     inst.ExternalPort,
		Database: inst.Database,
		Username: user,
		Password: password,
		URI:      fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, password, host, inst.ExternalPort, inst.Database),
	})
}

func (s *Server) logs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.lookup(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"instanceId": inst.ID,
		"logs": []dbaas.LogLine{
			{Timestamp: nanos(inst.CreatedAt), Line: "database system is ready to accept connections"},
		},
	})
}

func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SQL == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "sql is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.lookupRunning(w, r); !ok {
		return
	}

	sql := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(req.SQL), ";"))
	if !strings.EqualFold(sql, "select 1") {
		word, _, _ := strings.Cut(sql, " ")
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error": map[string]any{
				"code":    "QUERY_FAILED",
				"message": fmt.Sprintf("syntax error at or near %q", word),
				"details": dbaas.QueryErrorDetails{SQLState: "42601", Position: 1},
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, dbaas.QueryResult{
		Columns:  []string{"?column?"},
		Rows:     [][]any{{1}},
		RowCount: 1,
	})
}

func (s *Server) auditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer "+AdminToken {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	resourceID := r.URL.Query().Get("resourceId")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Newest first, like the real API.
	logs := []dbaas.LogLine{}
	for _, l := range slices.Backward(s.audit) {
		if resourceID == "" || strings.Contains(l.Line, `"resourceId":"`+resourceID+`"`) {
			logs = append(logs, l)
		}
	}

	writeJSON(w, http.StatusOK, dbaas.AuditLogList{Logs: logs})
}

// lookup finds the instance in the request path, or writes a 404.
// The caller must hold s.mu.
func (s *Server) lookup(w http.ResponseWriter, r *http.Request) (*dbaas.Instance, bool) {
	inst, ok := s.instances[r.PathValue("id")]
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	}
	return inst, ok
}

// lookupRunning is lookup for endpoints that need a Running instance.
func (s *Server) lookupRunning(w http.ResponseWriter, r *http.Request) (*dbaas.Instance, bool) {
	inst, ok := s.lookup(w, r)
	if ok && inst.Status != dbaas.StatusRunning {
		writeError(w, http.StatusConflict, "INSTANCE_NOT_READY", "instance is not ready")
		return nil, false
	}
	return inst, ok
}

// recordAudit appends an audit event. The caller must hold s.mu.
func (s *Server) recordAudit(action, id string) {
	line, _ := json.Marshal(map[string]string{"action": action, "resourceId": id})
	s.audit = append(s.audit, dbaas.LogLine{Timestamp: nanos(time.Now()), Line: string(line)})
}

func nanos(t time.Time) string {
	return strconv.FormatInt(t.UnixNano(), 10)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}
