package api

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strconv"
)

// consoleHTML is the interactive API console, embedded so the service stays
// a single self-contained binary.
//
//go:embed web/index.html
var consoleHTML []byte

// consoleETag is a strong validator over the embedded page, letting browsers
// revalidate with a cheap 304 and pick up a redeploy.
var consoleETag = strongETag(consoleHTML)

// strongETag renders a quoted hex digest suitable for an ETag header.
func strongETag(b []byte) string {
	sum := sha256.Sum256(b)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

// Console handles GET /console, serving the interactive HTML console
// unconditionally so it has a stable URL independent of the Accept header.
func (h *Handlers) Console(w http.ResponseWriter, r *http.Request) {
	h.serveConsole(w, r)
}

// serveConsole writes the embedded console page.
func (h *Handlers) serveConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Revalidate every visit so a redeploy is picked up immediately.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", consoleETag)

	if r.Header.Get("If-None-Match") == consoleETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(consoleHTML)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(consoleHTML)
}
