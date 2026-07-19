package api

import (
	"net/http"
	"strconv"

	"gomock/internal/entity"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// listEnvelope is the PokeAPI-style paginated response wrapper.
type listEnvelope struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []entity.Entity `json:"results"`
}

// paginate parses limit/offset query parameters (applying defaults and an
// upper bound on limit), slices items accordingly, and builds next/previous
// links for the response envelope.
func paginate(r *http.Request, items []entity.Entity) listEnvelope {
	limit := parseIntParam(r, "limit", defaultLimit)
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := parseIntParam(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	total := len(items)

	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	env := listEnvelope{
		Count:   total,
		Results: items[start:end],
	}

	if end < total {
		next := buildPageURL(r, limit, end)
		env.Next = &next
	}
	if start > 0 {
		prevOffset := start - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		prev := buildPageURL(r, limit, prevOffset)
		env.Previous = &prev
	}

	return env
}

// parseIntParam extracts an integer query parameter, returning fallback
// when it is missing or invalid.
func parseIntParam(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

// buildPageURL rebuilds the request URL with updated limit/offset params,
// preserving any other query parameters (e.g. filters).
func buildPageURL(r *http.Request, limit, offset int) string {
	q := r.URL.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))

	u := *r.URL
	u.RawQuery = q.Encode()
	u.Host = r.Host
	u.Scheme = "http"
	if r.TLS != nil {
		u.Scheme = "https"
	}

	return u.String()
}
