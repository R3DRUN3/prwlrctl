package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type HealthResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	ReleaseID   string `json:"releaseId"`
	ServiceID   string `json:"serviceId"`
	Description string `json:"description"`
}

func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	url := strings.TrimSuffix(c.BaseURL, "/api/v1") + "/health/live"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building health check request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prowler server unhealthy: HTTP %s", resp.Status)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("invalid health response: %w", err)
	}

	if health.Status != "pass" {
		return &health, fmt.Errorf("prowler server unhealthy: status=%q", health.Status)
	}

	return &health, nil
}
