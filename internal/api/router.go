package api

import "net/http"

// NewRouter builds the HTTP handler wiring all API routes to their handlers
// (Go 1.22+ pattern routing) and wrapping them in the middleware chain. The
// limiter may be nil or disabled, in which case throttling is skipped.
func NewRouter(h *Handlers, limiter *RateLimiter) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.Index)
	mux.HandleFunc("GET /console", h.Console)
	mux.HandleFunc("GET /robots.txt", h.Robots)
	mux.HandleFunc("GET /sitemap.xml", h.Sitemap)
	mux.HandleFunc("GET /api/v1/types", h.ListTypes)
	mux.HandleFunc("GET /api/v1/{type}", h.ListEntities)
	mux.HandleFunc("GET /api/v1/{type}/{idOrSlug}", h.GetEntity)

	// Outermost recovers panics; security headers apply to every response
	// (including 429s from the limiter, which sits closest to the routes).
	return recoverPanic(securityHeaders(limiter.Middleware(mux)))
}
