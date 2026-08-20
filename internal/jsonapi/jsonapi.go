// Package jsonapi provides minimal, generic building blocks for talking to
// JSON:API (application/vnd.api+json) backends such as the Prowler API.
package jsonapi

import "encoding/json"

const ContentType = "application/vnd.api+json"

// ResourceIdentifier references another resource by type+id, used in
// relationships (e.g. pointing a scan at a provider).
type ResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Relationship wraps one or many ResourceIdentifiers.
type Relationship struct {
	Data json.RawMessage `json:"data"`
	Meta map[string]any  `json:"meta,omitempty"`
}

func ToOne(id, typ string) Relationship {
	b, _ := json.Marshal(ResourceIdentifier{ID: id, Type: typ})
	return Relationship{Data: b}
}

// Resource is a generic JSON:API resource object. Attributes is kept
// flexible (map[string]any) because Prowler's schema evolves independently
// of this CLI; callers pull out the fields they need defensively.
type Resource struct {
	ID            string                  `json:"id,omitempty"`
	Type          string                  `json:"type"`
	Attributes    map[string]any          `json:"attributes,omitempty"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
}

// Document is the top-level JSON:API envelope. Data may unmarshal into a
// single Resource or a slice of Resource depending on the endpoint, so
// callers decode Data lazily via json.RawMessage.
type Document struct {
	Data     json.RawMessage `json:"data,omitempty"`
	Included json.RawMessage `json:"included,omitempty"`
	Meta     map[string]any  `json:"meta,omitempty"`
	Links    map[string]any  `json:"links,omitempty"`
	Errors   []APIError      `json:"errors,omitempty"`
}

type APIError struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Source struct {
		Pointer string `json:"pointer"`
	} `json:"source"`
}

// PageInfo mirrors Prowler's meta.pagination block on list responses.
type PageInfo struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Count int `json:"count"`
}

// Pagination reads meta.pagination, if the server included it. ok is false
// for endpoints/responses that don't carry pagination metadata.
func (d Document) Pagination() (PageInfo, bool) {
	raw, ok := d.Meta["pagination"]
	if !ok {
		return PageInfo{}, false
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return PageInfo{}, false
	}
	var p PageInfo
	if err := json.Unmarshal(b, &p); err != nil {
		return PageInfo{}, false
	}
	return p, true
}

// Request builds a JSON:API create/update body: {"data": {...}}.
func Request(res Resource) map[string]Resource {
	return map[string]Resource{"data": res}
}

// One decodes Data as a single resource.
func (d Document) One() (Resource, error) {
	var r Resource
	err := json.Unmarshal(d.Data, &r)
	return r, err
}

// Many decodes Data as a list of resources.
func (d Document) Many() ([]Resource, error) {
	var rs []Resource
	err := json.Unmarshal(d.Data, &rs)
	return rs, err
}

// IncludedResources decodes the top-level "included" array (populated when
// a request uses ?include=...).
func (d Document) IncludedResources() ([]Resource, error) {
	if len(d.Included) == 0 {
		return nil, nil
	}
	var rs []Resource
	err := json.Unmarshal(d.Included, &rs)
	return rs, err
}

// IncludedIndex builds a "type/id" -> Resource lookup from the included
// array, so callers can resolve relationships without extra requests.
func (d Document) IncludedIndex() (map[string]Resource, error) {
	items, err := d.IncludedResources()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]Resource, len(items))
	for _, r := range items {
		idx[r.Type+"/"+r.ID] = r
	}
	return idx, nil
}

// RelatedIDs returns the type+id pairs referenced by a to-many or to-one
// relationship, regardless of which shape the server used.
func (r Resource) RelatedIDs(relName string) []ResourceIdentifier {
	rel, ok := r.Relationships[relName]
	if !ok || len(rel.Data) == 0 {
		return nil
	}
	// Try to-many first.
	var many []ResourceIdentifier
	if err := json.Unmarshal(rel.Data, &many); err == nil {
		return many
	}
	// Fall back to to-one.
	var one ResourceIdentifier
	if err := json.Unmarshal(rel.Data, &one); err == nil && one.ID != "" {
		return []ResourceIdentifier{one}
	}
	return nil
}

// Str safely reads a string attribute, returning "" if absent/wrong type.
func (r Resource) Str(key string) string {
	if v, ok := r.Attributes[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// StrAny tries each key in order and returns the first non-empty string
// attribute found. Useful when a field name may differ across API versions.
func (r Resource) StrAny(keys ...string) string {
	for _, k := range keys {
		if v := r.Str(k); v != "" {
			return v
		}
	}
	return ""
}

// Any returns a raw attribute value, or nil.
func (r Resource) Any(key string) any {
	return r.Attributes[key]
}

// NestedBool reads attrs[parentKey][childKey] as a bool, e.g. Prowler's
// providers expose connectivity as connection: {"connected": true, ...}.
// ok is false when the path is missing or not a bool.
func (r Resource) NestedBool(parentKey, childKey string) (value bool, ok bool) {
	parent, ok := r.Attributes[parentKey].(map[string]any)
	if !ok {
		return false, false
	}
	v, ok := parent[childKey].(bool)
	return v, ok
}
