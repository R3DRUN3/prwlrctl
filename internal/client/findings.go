package client

import (
	"context"
	"net/url"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
)

// ListFindings lists findings, typically filtered by scan, provider,
// severity, status, or date range (Prowler requires at least one date
// filter, e.g. filters["updated_at.gte"]).
func (c *Client) ListFindings(ctx context.Context, filters map[string]string, page, size int) (jsonapi.Document, error) {
	q := buildListQuery(filters, page, size)
	q.Set("include", "resources")

	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/findings", q, nil, &doc)
	return doc, err
}

// GetFinding retrieves a single finding by ID.
//
// The resources relationship is included so the response contains the
// related resource in the top-level JSON:API "included" array.
func (c *Client) GetFinding(ctx context.Context, id string) (jsonapi.Document, error) {
	q := url.Values{}
	q.Set("include", "resources")

	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/findings/"+url.PathEscape(id), q, nil, &doc)
	return doc, err
}
