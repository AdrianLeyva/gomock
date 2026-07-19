// Package store provides a thread-safe in-memory collection of entities
// grouped by user-defined type, plus the loader/adapter logic that
// normalizes arbitrary user JSON into the generic entity.Entity schema.
package store

import (
	"fmt"
	"regexp"
	"strings"

	"gomock/internal/entity"
)

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

// AdaptRecord converts a raw, user-supplied flat JSON object into the
// generic Entity envelope. "id" and "name" are required; "slug" is
// derived from "name" when absent. Every other field is copied verbatim
// into Attributes, which is what lets callers plug in arbitrary custom
// schemas without any code changes.
func AdaptRecord(raw map[string]any) (entity.Entity, error) {
	rawID, ok := raw["id"]
	if !ok {
		return entity.Entity{}, fmt.Errorf("missing required field %q", "id")
	}
	id, err := toInt(rawID)
	if err != nil {
		return entity.Entity{}, fmt.Errorf("field %q: %w", "id", err)
	}

	rawName, ok := raw["name"]
	if !ok {
		return entity.Entity{}, fmt.Errorf("missing required field %q", "name")
	}
	name, ok := rawName.(string)
	if !ok || name == "" {
		return entity.Entity{}, fmt.Errorf("field %q must be a non-empty string", "name")
	}

	slug := slugify(name)
	if rawSlug, ok := raw["slug"]; ok {
		s, ok := rawSlug.(string)
		if !ok || s == "" {
			return entity.Entity{}, fmt.Errorf("field %q must be a non-empty string", "slug")
		}
		slug = s
	}

	attributes := make(map[string]any, len(raw))
	for k, v := range raw {
		if k == "id" || k == "name" || k == "slug" {
			continue
		}
		attributes[k] = v
	}

	return entity.Entity{
		ID:         id,
		Name:       name,
		Slug:       slug,
		Attributes: attributes,
	}, nil
}

// toInt converts a decoded JSON numeric value (float64) or a Go int into
// an int, returning an error for any other type.
func toInt(v any) (int, error) {
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	default:
		return 0, fmt.Errorf("must be a number, got %T", v)
	}
}

// slugify derives a URL-friendly slug from an entity name (lowercase,
// non-alphanumeric runs collapsed to a single hyphen, trimmed).
func slugify(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	slug := slugInvalidChars.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}
