package api

import (
	"fmt"
	"net/http"
	"strings"

	"gomock/internal/entity"
)

// reservedQueryParams are query parameters interpreted by pagination
// rather than treated as entity field filters.
var reservedQueryParams = map[string]bool{
	"limit":  true,
	"offset": true,
}

// applyFilters returns the subset of items matching every non-reserved
// query parameter as a case-insensitive equality check against either a
// core field (name, slug) or an Attributes entry. Multiple filters are
// combined with AND semantics.
func applyFilters(r *http.Request, items []entity.Entity) []entity.Entity {
	filters := make(map[string]string)
	for key, values := range r.URL.Query() {
		if reservedQueryParams[key] || len(values) == 0 {
			continue
		}
		filters[key] = values[0]
	}

	if len(filters) == 0 {
		return items
	}

	filtered := make([]entity.Entity, 0, len(items))
	for _, item := range items {
		if matchesFilters(item, filters) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// matchesFilters reports whether item satisfies every key/value filter.
func matchesFilters(item entity.Entity, filters map[string]string) bool {
	for key, want := range filters {
		got, ok := fieldValue(item, key)
		if !ok || !strings.EqualFold(got, want) {
			return false
		}
	}
	return true
}

// fieldValue resolves a filter key against an entity's core fields first,
// then falls back to its Attributes map, stringifying whatever value is
// found so filters work uniformly across arbitrary custom schemas.
func fieldValue(item entity.Entity, key string) (string, bool) {
	switch strings.ToLower(key) {
	case "name":
		return item.Name, true
	case "slug":
		return item.Slug, true
	}

	if v, ok := item.Attributes[key]; ok {
		return fmt.Sprintf("%v", v), true
	}
	return "", false
}
