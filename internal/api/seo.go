package api

import (
	"net/http"
	"strconv"
)

// Robots handles GET /robots.txt, allowing all crawlers and pointing them at
// the sitemap so the console can be discovered and indexed.
func (h *Handlers) Robots(w http.ResponseWriter, r *http.Request) {
	body := "User-agent: *\nAllow: /\nSitemap: " + baseURL(r) + "/sitemap.xml\n"
	writeText(w, "text/plain; charset=utf-8", body)
}

// Sitemap handles GET /sitemap.xml, listing the console's canonical URL.
func (h *Handlers) Sitemap(w http.ResponseWriter, r *http.Request) {
	body := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n" +
		"  <url><loc>" + baseURL(r) + "/</loc></url>\n" +
		"</urlset>\n"
	writeText(w, "application/xml; charset=utf-8", body)
}

// writeText writes a plain string body with the given content type.
func writeText(w http.ResponseWriter, contentType, body string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}
