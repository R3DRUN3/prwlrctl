package client

import (
	"context"
	"net/url"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
)

// ListScans lists scans, optionally filtered (e.g. filters["provider"] = id,
// filters["state"] = "completed").
func (c *Client) ListScans(ctx context.Context, filters map[string]string, page, size int) (jsonapi.Document, error) {
	q := buildListQuery(filters, page, size)
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/scans", q, nil, &doc)
	return doc, err
}

func (c *Client) GetScan(ctx context.Context, id string) (jsonapi.Document, error) {
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/scans/"+url.PathEscape(id), nil, nil, &doc)
	return doc, err
}

// CreateScan launches a new scan against a provider. Prowler treats this as
// asynchronous: the returned resource carries the initial state (typically
// "scheduled"/"executing"); poll GetScan or WaitForScan to follow progress.
func (c *Client) CreateScan(ctx context.Context, providerID, name string) (jsonapi.Document, error) {
	attrs := map[string]any{}
	if name != "" {
		attrs["name"] = name
	}
	res := jsonapi.Resource{
		Type:       "scans",
		Attributes: attrs,
		Relationships: map[string]jsonapi.Relationship{
			"provider": jsonapi.ToOne(providerID, "providers"),
		},
	}
	var doc jsonapi.Document
	err := c.Do(ctx, "POST", "/scans", nil, jsonapi.Request(res), &doc)
	return doc, err
}
