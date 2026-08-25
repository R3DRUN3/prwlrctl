package prowler

import (
	"context"
	"net/url"

	"github.com/r3drun3/prwlrctl/pkg/prowler/jsonapi"
)

// ListProviders returns cloud provider accounts/connections registered in
// Prowler (AWS/Azure/GCP/Kubernetes/...). Supports raw JSON:API filters,
// e.g. filter["provider"] = "aws".
func (c *Client) ListProviders(ctx context.Context, filters map[string]string, page, size int) (jsonapi.Document, error) {
	q := buildListQuery(filters, page, size)
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/providers", q, nil, &doc)
	return doc, err
}

func (c *Client) GetProvider(ctx context.Context, id string) (jsonapi.Document, error) {
	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/providers/"+url.PathEscape(id), nil, nil, &doc)
	return doc, err
}

// BuildQuery turns filters into JSON:API filter[...] query params, with no
// page/size params set. Pair with prowler.ListAll, which follows the
// server's "next" links instead of manual pagination.
func BuildQuery(filters map[string]string) url.Values {
	return buildListQuery(filters, 0, 0)
}

func buildListQuery(filters map[string]string, page, size int) url.Values {
	q := url.Values{}
	for k, v := range filters {
		if v != "" {
			q.Set("filter["+k+"]", v)
		}
	}
	if page > 0 {
		q.Set("page[number]", itoa(page))
	}
	if size > 0 {
		q.Set("page[size]", itoa(size))
	}
	return q
}

func itoa(i int) string {
	// Avoid importing strconv twice across files; kept local & trivial.
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
