package api

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"strconv"
)

// consoleHTML is the interactive API console, compiled into the binary so
// the service stays a single self-contained artifact with no runtime asset
// directory to deploy or mount.
//
// go:embed patterns are package-relative, which is why the asset lives at
// internal/api/web/ rather than a repository-root web/ directory: the
// Dockerfile's build stage copies only go.mod, cmd/ and internal/, so an
// asset anywhere else would compile locally and fail in the image build.
//
//go:embed web/index.html
var consoleHTML []byte

// consoleETag is a strong validator over the embedded page. It is fixed for
// the lifetime of a binary and changes whenever the asset does, so browsers
// revalidate with a cheap 304 and always pick up a redeploy.
var consoleETag = strongETag(consoleHTML)

// strongETag renders a quoted hex digest suitable for an ETag header.
func strongETag(b []byte) string {
	sum := sha256.Sum256(b)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

// Console handles GET /console, serving the interactive HTML console
// unconditionally. It exists alongside the content-negotiated root so the
// page has a stable, shareable URL that does not depend on the client's
// Accept header.
func (h *Handlers) Console(w http.ResponseWriter, r *http.Request) {
	h.serveConsole(w, r)
}

// serveConsole writes the embedded console page.
//
// No Content-Security-Policy is set: the page inlines its CSS and script
// (deliberately — it has no build step and loads nothing from the network),
// so any workable policy would need 'unsafe-inline', which would advertise
// a protection it does not provide. The page's own defence against hostile
// stored data is that it renders values as text, never as markup.
func (h *Handlers) serveConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Revalidate every time rather than caching for a fixed window: a
	// console left stale after a redeploy is a worse failure than one
	// conditional request per visit.
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
