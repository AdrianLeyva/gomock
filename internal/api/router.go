package api

import "net/http"

// NewRouter builds the HTTP handler wiring all API routes to their
// handlers, using Go 1.22+ method- and pattern-based ServeMux routing.
func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/types", h.ListTypes)
	mux.HandleFunc("GET /api/v1/{type}", h.ListEntities)
	mux.HandleFunc("GET /api/v1/{type}/{idOrSlug}", h.GetEntity)
	mux.HandleFunc("PUT /api/v1/admin/{type}", h.ReplaceType)

	return mux
}
