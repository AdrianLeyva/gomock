// Package entity defines the generic record type served by the API.
package entity

// Entity is the generic envelope every entity type (e.g. "pokemon",
// "items") is normalized into. Core fields are fixed; everything else
// supplied by the user's custom JSON lives under Attributes, allowing
// arbitrary, per-type schemas without changing any Go code.
type Entity struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	Attributes map[string]any `json:"attributes"`
}
