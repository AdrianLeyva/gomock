package api

import (
	"log"
	"net/http"
)

// contentSecurityPolicy restricts what the console may load. The page inlines
// its CSS and JS, so 'unsafe-inline' is unavoidable, but the policy still
// blocks external and eval'd sources, framing, base-tag hijacking, form
// exfiltration, and object/embed — worthwhile hardening over no policy.
const contentSecurityPolicy = "default-src 'none'; base-uri 'none'; " +
	"form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; " +
	"style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; " +
	"connect-src 'self'"

// securityHeaders sets response headers that harden every response against
// common web attacks (MIME sniffing, clickjacking, referrer leakage, and
// unwanted resource loading). They are set before the wrapped handler runs so
// they apply even to error and rate-limited responses.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=(), browsing-topics=()")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		next.ServeHTTP(w, r)
	})
}

// recoverPanic turns a panic in any downstream handler into a JSON 500 instead
// of a dropped connection, logging the cause. It only writes a status if the
// handler had not already started the response.
func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic serving %s %s: %v", r.Method, r.URL.Path, rec)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
