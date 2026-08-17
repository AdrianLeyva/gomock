package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRobotsTxt(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandlers().Robots(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "User-agent: *") {
		t.Errorf("robots.txt missing user-agent line: %q", body)
	}
	if !strings.Contains(body, "Sitemap: http://example.com/sitemap.xml") {
		t.Errorf("robots.txt missing absolute sitemap URL: %q", body)
	}
}

func TestSitemapXML(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandlers().Sitemap(rec, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/xml") {
		t.Errorf("Content-Type = %q, want application/xml", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<loc>http://example.com/</loc>") {
		t.Errorf("sitemap missing canonical loc: %q", body)
	}
}

func TestConsoleSetsCanonicalLinkHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandlers().Console(rec, httptest.NewRequest(http.MethodGet, "/console", nil))

	link := rec.Header().Get("Link")
	if !strings.Contains(link, `rel="canonical"`) || !strings.Contains(link, "http://example.com/") {
		t.Errorf("Link header = %q, want a canonical link to the root", link)
	}
}
