// Package output renders JSON:API resources as either a human-friendly
// table or raw JSON, so the same command works for interactive operators
// and for scripts/cron piping into jq.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/r3drun3/prwlrctl/internal/jsonapi"
)

type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
)

// RenderTable writes a simple aligned table. columns define header + how to
// read each cell from a resource.
func RenderTable(w io.Writer, resources []jsonapi.Resource, columns []Column) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	for _, r := range resources {
		row := make([]string, len(columns))
		for i, c := range columns {
			row[i] = c.Value(r)
		}
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

type Column struct {
	Header string
	Value  func(jsonapi.Resource) string
}

// Attr returns a Column that reads a plain string attribute, defaulting to
// "-" when absent, so partial/evolving schemas never crash rendering.
func Attr(header, key string) Column {
	return Column{Header: header, Value: func(r jsonapi.Resource) string {
		if v := r.Str(key); v != "" {
			return v
		}
		return "-"
	}}
}

// NestedBoolStatus reads a nested boolean attribute (e.g. connection.connected)
// and renders it as trueLabel/falseLabel, or "-" if the field is absent.
func NestedBoolStatus(header, parentKey, childKey, trueLabel, falseLabel string) Column {
	return Column{Header: header, Value: func(r jsonapi.Resource) string {
		v, ok := r.NestedBool(parentKey, childKey)
		if !ok {
			return "-"
		}
		if v {
			return trueLabel
		}
		return falseLabel
	}}
}

func IDColumn() Column {
	return Column{Header: "ID", Value: func(r jsonapi.Resource) string { return r.ID }}
}

// JSONPretty prints any JSON-marshalable value as indented JSON. Works for
// both a full jsonapi.Document and a plain []jsonapi.Resource (used by
// --all, which has no single top-level envelope after page-merging).
func JSONPretty(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
