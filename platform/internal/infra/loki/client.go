package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, httpClient: &http.Client{}}
}

type LogLine struct {
	Timestamp string `json:"timestamp"`
	Line      string `json:"line"`
}

// Query runs a LogQL query against Loki's range endpoint.
func (c *Client) Query(ctx context.Context, logql string, limit int, lookback time.Duration) ([]LogLine, error) {
	end := time.Now()
	start := end.Add(-lookback)

	params := url.Values{}
	params.Set("query", logql)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("direction", "backward")
	params.Set("start", fmt.Sprintf("%d", start.UnixNano()))
	params.Set("end", fmt.Sprintf("%d", end.UnixNano()))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/loki/api/v1/query_range?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query loki: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Result []struct {
				Values [][2]string `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode loki response: %w", err)
	}

	var lines []LogLine
	for _, stream := range result.Data.Result {
		for _, v := range stream.Values {
			lines = append(lines, LogLine{Timestamp: v[0], Line: v[1]})
		}
	}

	// direction=backward only sorts entries within each stream. Loki
	// groups results by distinct label set (stream), so a query spanning
	// more than one stream - e.g. audit events for different resource
	// IDs - comes back as several separately-sorted runs concatenated
	// together, not one globally-sorted list. Loki's timestamps are
	// nanoseconds-since-epoch as a string, so they sort correctly as
	// integers.
	sort.Slice(lines, func(i, j int) bool {
		ti, _ := strconv.ParseInt(lines[i].Timestamp, 10, 64)
		tj, _ := strconv.ParseInt(lines[j].Timestamp, 10, 64)
		return ti > tj // descending: now -> old
	})

	if len(lines) > limit {
		lines = lines[:limit]
	}

	return lines, nil
}
