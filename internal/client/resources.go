package client

import (
	"context"
	"net/url"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
)

// GetResource retrieves a single resource by ID.
func (c *Client) GetResource(ctx context.Context, id string) (jsonapi.Document, error) {
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/resources/"+url.PathEscape(id), nil, nil, &doc)
	return doc, err
}
