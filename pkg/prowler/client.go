// package prowler is a small, dependency-free HTTP client for the Prowler
// JSON:API backend: auth headers, retries on 429/5xx, and JSON:API error
// unwrapping into Go errors.
package prowler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/r3drun3/prwlrctl/pkg/prowler/jsonapi"
)

type Client struct {
	BaseURL     string
	APIKey      string
	AccessToken string
	HTTPClient  *http.Client
	MaxRetries  int
}

func New(baseURL, apiKey, accessToken string, timeout time.Duration) *Client {
	return &Client{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		APIKey:      apiKey,
		AccessToken: accessToken,
		HTTPClient:  &http.Client{Timeout: timeout},
		MaxRetries:  3,
	}
}

// APIError is returned when the server responds with a JSON:API error body.
type APIError struct {
	StatusCode int
	Errors     []jsonapi.APIError
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("request failed with status %d", e.StatusCode)
	}
	parts := make([]string, 0, len(e.Errors))
	for _, er := range e.Errors {
		if er.Detail != "" {
			parts = append(parts, er.Detail)
		} else if er.Title != "" {
			parts = append(parts, er.Title)
		}
	}
	return fmt.Sprintf("request failed with status %d: %s", e.StatusCode, strings.Join(parts, "; "))
}

// Do issues an HTTP request against path (e.g. "/scans", or an absolute
// URL such as a JSON:API pagination "next" link), with optional query
// params and a JSON body, decoding the JSON:API response into out.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, out *jsonapi.Document) error {
	full := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		full = c.BaseURL + path
	}
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff(attempt)):
			}
			if reader != nil {
				// rewind body for retry
				b, _ := json.Marshal(body)
				reader = bytes.NewReader(b)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, full, reader)
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		c.applyHeaders(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue // network error: retry
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = &APIError{StatusCode: resp.StatusCode}
			continue // retryable
		}

		if resp.StatusCode >= 400 {
			var doc jsonapi.Document
			_ = json.Unmarshal(data, &doc)
			return &APIError{StatusCode: resp.StatusCode, Errors: doc.Errors}
		}

		if out != nil && len(data) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("giving up after %d attempts: %w", c.MaxRetries+1, lastErr)
}

// ListAll follows JSON:API pagination ("next" links) until exhausted and
// returns every resource across all pages. Use this in scripts/cron jobs
// that need the complete result set (e.g. "every finding since X") without
// hand-rolling page loops. path is the initial endpoint (e.g. "/findings");
// query carries the initial filters (page params are ignored/overridden by
// whatever the server returns in "next").
func (c *Client) ListAll(ctx context.Context, path string, query url.Values) ([]jsonapi.Resource, error) {
	var all []jsonapi.Resource
	next := path
	q := query

	for next != "" {
		var doc jsonapi.Document
		if err := c.Do(ctx, "GET", next, q, nil, &doc); err != nil {
			return nil, err
		}
		items, err := doc.Many()
		if err != nil {
			return nil, err
		}
		all = append(all, items...)

		nextLink, _ := doc.Links["next"].(string)
		if nextLink == "" {
			break
		}
		next = nextLink
		q = nil // the "next" link already carries its own query string
	}
	return all, nil
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", jsonapi.ContentType)
	req.Header.Set("Accept", jsonapi.ContentType)
	// JWT takes precedence over API key, mirroring server-side behavior.
	switch {
	case c.AccessToken != "":
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	case c.APIKey != "":
		req.Header.Set("Authorization", "Api-Key "+c.APIKey)
	}
}

func backoff(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt-1)) * 300 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
