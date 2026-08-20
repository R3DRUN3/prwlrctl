package client

import (
	"context"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
)

// ListFindings lists findings, typically filtered by scan, provider,
// severity, status, or date range (Prowler requires at least one date
// filter, e.g. filters["updated_at.gte"]). Always requests the related
// "resources" so callers can resolve region/name/uid without extra calls.
func (c *Client) ListFindings(ctx context.Context, filters map[string]string, page, size int) (jsonapi.Document, error) {
	q := buildListQuery(filters, page, size)
	q.Set("include", "resources")
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/findings", q, nil, &doc)
	return doc, err
}
