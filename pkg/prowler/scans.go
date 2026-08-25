package prowler

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/r3drun3/prwlrctl/pkg/prowler/jsonapi"
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

// GetComplianceOverviews returns the compliance overview for all frameworks
// associated with a scan.
func (c *Client) GetComplianceOverviews(ctx context.Context, scanID string) (jsonapi.Document, error) {
	q := url.Values{}
	q.Set("filter[scan_id]", scanID)

	var doc jsonapi.Document
	err := c.Do(ctx, "GET", "/compliance-overviews", q, nil, &doc)
	return doc, err
}

// CreateScan launches a new scan against a provider. Prowler responds with
// the asynchronous task resource; the actual scan resource is created shortly
// afterwards.
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

// CreateScanAndGetID launches a scan and waits until Prowler creates the
// corresponding scan resource. The POST response identifies the asynchronous
// task, not the scan itself, so the task relationship is used to correlate
// the newly-created scan.
func (c *Client) CreateScanAndGetID(ctx context.Context, providerID, name string) (string, error) {
	doc, err := c.CreateScan(ctx, providerID, name)
	if err != nil {
		return "", err
	}

	task, err := doc.One()
	if err != nil {
		return "", fmt.Errorf("decoding scan launch response: %w", err)
	}

	if task.ID == "" {
		return "", fmt.Errorf("scan launch response does not contain a task ID")
	}

	taskID := task.ID

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		scanID, err := c.findScanByTask(ctx, taskID)
		if err != nil {
			return "", err
		}
		if scanID != "" {
			return scanID, nil
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf(
				"timed out waiting for scan for task %s: %w",
				taskID,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

// findScanByTask returns the scan ID whose task relationship matches taskID.
// An empty ID means the task exists but the scan resource has not appeared yet.
func (c *Client) findScanByTask(ctx context.Context, taskID string) (string, error) {
	doc, err := c.ListScans(ctx, nil, 0, 0)
	if err != nil {
		return "", err
	}

	scans, err := doc.Many()
	if err != nil {
		return "", fmt.Errorf("decoding scans: %w", err)
	}

	for _, scan := range scans {
		for _, related := range scan.RelatedIDs("task") {
			if related.ID == taskID {
				return scan.ID, nil
			}
		}
	}

	// If the current page does not contain it, follow all pages. This matters
	// when the new scan appears on the first page only after another request.
	if pagination, ok := doc.Pagination(); ok && pagination.Pages > 1 {
		for page := 2; page <= pagination.Pages; page++ {
			doc, err := c.ListScans(ctx, nil, page, 0)
			if err != nil {
				return "", err
			}

			scans, err := doc.Many()
			if err != nil {
				return "", fmt.Errorf("decoding scans page %d: %w", page, err)
			}

			for _, scan := range scans {
				for _, related := range scan.RelatedIDs("task") {
					if related.ID == taskID {
						return scan.ID, nil
					}
				}
			}
		}
	}

	return "", nil
}
