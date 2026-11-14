package dbaas

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// InstancesService groups all instance-related API calls, accessed via
// client.Instances.<Method>(...).
type InstancesService struct {
	client *Client
}

// CreateInstanceInput is what to create.
type CreateInstanceInput struct {
	// Name is the database name, 3-63 characters. The instance's ID is
	// generated, not this.
	Name string `json:"name"`
	// Version is the PostgreSQL major version: 14, 15 or 16.
	Version int `json:"version"`
	// Storage is a Kubernetes quantity, e.g. "10Gi".
	Storage string `json:"storage"`
	// Instances is the number of PostgreSQL pods, 1-5. Zero means the
	// server's default.
	Instances int `json:"instances,omitempty"`
	// Username is the database owner. Empty means the server's default.
	Username string `json:"username,omitempty"`

	// IdempotencyKey is required. Sending the same key again returns the
	// instance it created instead of creating another one, so a create
	// can be retried safely. Use one key per logical instance, and reuse
	// it on every retry. It is sent as the Idempotency-Key header.
	IdempotencyKey string `json:"-"`
}

// Create starts creating an instance and returns it with status Pending.
// The operator provisions it in the background; use WaitUntilRunning to
// wait for it.
func (s *InstancesService) Create(ctx context.Context, input CreateInstanceInput) (*Instance, error) {
	var out Instance

	err := s.client.do(ctx, request{
		method:  http.MethodPost,
		path:    "/api/v1/instances",
		headers: map[string]string{"Idempotency-Key": input.IdempotencyKey},
		body:    input,
		out:     &out,
	})
	if err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}

	return &out, nil
}

// Get returns one instance. It fails with ErrNotFound for instances that
// don't exist or belong to someone else.
func (s *InstancesService) Get(ctx context.Context, id string) (*Instance, error) {
	var out Instance

	if err := s.client.do(ctx, request{method: http.MethodGet, path: instancePath(id, ""), out: &out}); err != nil {
		return nil, fmt.Errorf("get instance: %w", err)
	}

	return &out, nil
}

// ListOptions selects one page of instances.
type ListOptions struct {
	// Limit is the page size. Zero means the server's default (20); the
	// server caps it at 100.
	Limit int
	// Cursor is the NextCursor of the previous page; empty for the first.
	Cursor string
}

// InstanceList is one page of instances.
type InstanceList struct {
	Items []*Instance `json:"items"`
	// NextCursor fetches the next page. Empty on the last page.
	NextCursor string `json:"nextCursor,omitempty"`
}

// List returns one page of the caller's instances. To go through all of
// them, use All.
func (s *InstancesService) List(ctx context.Context, opts ListOptions) (*InstanceList, error) {
	var out InstanceList

	err := s.client.do(ctx, request{
		method: http.MethodGet,
		path:   "/api/v1/instances",
		query:  pageQuery(opts.Limit, opts.Cursor),
		out:    &out,
	})
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}

	return &out, nil
}

// All iterates over all the caller's instances, fetching pages of
// pageSize (zero for the server's default) as the loop needs them. It
// stops at the first error, which it yields with a nil instance.
func (s *InstancesService) All(ctx context.Context, pageSize int) iter.Seq2[*Instance, error] {
	return func(yield func(*Instance, error) bool) {
		opts := ListOptions{Limit: pageSize}

		for {
			page, err := s.List(ctx, opts)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, inst := range page.Items {
				if !yield(inst, nil) {
					return
				}
			}

			if page.NextCursor == "" {
				return
			}
			opts.Cursor = page.NextCursor
		}
	}
}

// UpdateInstanceInput is what to change; nil fields are left as they are.
// At least one must be set.
type UpdateInstanceInput struct {
	Version *int    `json:"version,omitempty"`
	Storage *string `json:"storage,omitempty"`
}

// Update changes an instance and returns it with status Pending until the
// operator has applied the change.
func (s *InstancesService) Update(ctx context.Context, id string, input UpdateInstanceInput) (*Instance, error) {
	var out Instance

	err := s.client.do(ctx, request{method: http.MethodPatch, path: instancePath(id, ""), body: input, out: &out})
	if err != nil {
		return nil, fmt.Errorf("update instance: %w", err)
	}

	return &out, nil
}

// Delete starts deleting an instance. It returns once deletion has
// started; the instance is gone shortly after.
func (s *InstancesService) Delete(ctx context.Context, id string) error {
	if err := s.client.do(ctx, request{method: http.MethodDelete, path: instancePath(id, "")}); err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}

	return nil
}

// Credentials returns how to connect to an instance's database. It fails
// with ErrInstanceNotReady until the instance is reachable.
func (s *InstancesService) Credentials(ctx context.Context, id string) (*InstanceCredentials, error) {
	var out InstanceCredentials

	if err := s.client.do(ctx, request{method: http.MethodGet, path: instancePath(id, "/credentials"), out: &out}); err != nil {
		return nil, fmt.Errorf("get instance credentials: %w", err)
	}

	return &out, nil
}

// Logs returns an instance's recent PostgreSQL logs.
func (s *InstancesService) Logs(ctx context.Context, id string) ([]LogLine, error) {
	var out struct {
		InstanceID string    `json:"instanceId"`
		Logs       []LogLine `json:"logs"`
	}

	if err := s.client.do(ctx, request{method: http.MethodGet, path: instancePath(id, "/logs"), out: &out}); err != nil {
		return nil, fmt.Errorf("get instance logs: %w", err)
	}

	return out.Logs, nil
}

// Query runs one SQL statement against an instance's database through the
// API (the SQL console). A statement PostgreSQL rejects fails with
// ErrQueryFailed; its *APIError's QueryDetails says why.
func (s *InstancesService) Query(ctx context.Context, id, sql string) (*QueryResult, error) {
	var out QueryResult

	err := s.client.do(ctx, request{
		method: http.MethodPost,
		path:   instancePath(id, "/query"),
		body:   map[string]string{"sql": sql},
		out:    &out,
	})
	if err != nil {
		return nil, fmt.Errorf("query instance: %w", err)
	}

	return &out, nil
}

// WaitUntilRunning polls an instance until it is Running and returns it.
// It fails with ErrInstanceFailed if the instance fails, and with the
// context's error when ctx is done, so bound the wait with a timeout:
//
//	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
//	defer cancel()
//	inst, err := c.Instances.WaitUntilRunning(ctx, id)
func (s *InstancesService) WaitUntilRunning(ctx context.Context, id string) (*Instance, error) {
	ticker := time.NewTicker(s.client.pollInterval)
	defer ticker.Stop()

	for {
		inst, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}

		switch inst.Status {
		case StatusRunning:
			return inst, nil
		case StatusFailed:
			return inst, fmt.Errorf("wait for instance %s: %w", id, ErrInstanceFailed)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for instance %s: %w", id, ctx.Err())
		case <-ticker.C:
		}
	}
}

// pageQuery builds the limit and cursor query parameters, leaving out the
// ones that are unset.
func pageQuery(limit int, cursor string) url.Values {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	return q
}
