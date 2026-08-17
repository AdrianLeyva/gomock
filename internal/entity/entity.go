// Package entity defines the generic record type served by the API.
package entity

// Entity is the generic envelope every entity type is normalized into. Core
// fields are fixed; everything else from the source JSON lives under
// Attributes, allowing arbitrary per-type schemas with no code changes.
type Entity struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Slug       string         `json:"slug"`
	Attributes map[string]any `json:"attributes"`
}
