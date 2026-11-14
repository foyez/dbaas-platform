package dbaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// maxErrorBody caps how much of an error response is read into an APIError.
const maxErrorBody = 64 << 10

// request is one API call.
type request struct {
	method string
	path   string
	query  url.Values
	// headers are extra request headers, e.g. Idempotency-Key.
	headers map[string]string
	// body is encoded as JSON when non-nil.
	body any
	// out is decoded from the JSON response when non-nil.
	out any
}

func (c *Client) do(ctx context.Context, r request) error {
	var reqBody io.Reader

	if r.body != nil {
		b, err := json.Marshal(r.body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	u := c.baseURL + r.path
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, r.method, u, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if r.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	if c.tokenSource != nil {
		token, err := c.tokenSource.Token()
		if err != nil {
			return fmt.Errorf("get access token: %w", err)
		}
		token.SetAuthHeader(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return newAPIError(resp.StatusCode, body)
	}

	if r.out != nil {
		if err := json.NewDecoder(resp.Body).Decode(r.out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// instancePath builds /api/v1/instances/{id}[/suffix], escaping the id so
// it can't change the path.
func instancePath(id, suffix string) string {
	return "/api/v1/instances/" + url.PathEscape(id) + suffix
}
