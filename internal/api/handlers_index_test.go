package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gomock/internal/entity"
	"gomock/internal/store"
)

const browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"

func testHandlers() *Handlers {
	return NewHandlers(store.New(map[string][]entity.Entity{
		"characters": {
			{ID: 1, Name: "Iron Man", Slug: "iron-man", Attributes: map[string]any{"affiliation": "Avengers"}},
		},
	}))
}

func doIndex(t *testing.T, target, accept string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	rec := httptest.NewRecorder()
	testHandlers().Index(rec, req)
	return rec
}

// TestIndexServesJSONToAPIClients is the regression guard for the whole
// feature: adding the console must not change what a non-browser sees at /.
func TestIndexServesJSONToAPIClients(t *testing.T) {
	for _, accept := range []string{"", "*/*", "application/json"} {
		rec := doIndex(t, "/", accept)

		if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Errorf("Accept %q: Content-Type = %q, want JSON", accept, got)
		}

		var body indexResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("Accept %q: body is not valid JSON: %v", accept, err)
		}
		if body.Name != "gomock" {
			t.Errorf("Accept %q: name = %q, want %q", accept, body.Name, "gomock")
		}
		if _, ok := body.Types["characters"]; !ok {
			t.Errorf("Accept %q: types = %v, want a 'characters' entry", accept, body.Types)
		}
	}
}

func TestIndexServesConsoleToBrowsers(t *testing.T) {
	rec := doIndex(t, "/", browserAccept)

	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want HTML", got)
	}
	if !strings.Contains(strings.ToLower(rec.Body.String()), "<!doctype html") {
		t.Error("body does not look like the console page")
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("ETag header is missing")
	}
}

func TestIndexAlwaysSetsVaryAccept(t *testing.T) {
	for _, accept := range []string{"", browserAccept} {
		if got := doIndex(t, "/", accept).Header().Get("Vary"); got != "Accept" {
			t.Errorf("Accept %q: Vary = %q, want %q", accept, got, "Accept")
		}
	}
}

// TestIndexFormatJSONOverridesAccept covers reading the discovery document
// from a browser, where the Accept header is not under the user's control.
func TestIndexFormatJSONOverridesAccept(t *testing.T) {
	rec := doIndex(t, "/?format=json", browserAccept)

	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", got)
	}
}

func TestConsoleIgnoresAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/console", nil)
	req.Header.Set("Accept", "application/json")

	rec := httptest.NewRecorder()
	testHandlers().Console(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want HTML", got)
	}
}

func TestConsoleHonoursIfNoneMatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/console", nil)
	req.Header.Set("If-None-Match", consoleETag)

	rec := httptest.NewRecorder()
	testHandlers().Console(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body length = %d, want 0 for a 304", rec.Body.Len())
	}
}

// TestRequestSchemeHonoursForwardedProto covers the link-building fix: on a
// hosted deployment the proxy terminates TLS, so the connection this process
// sees is plain HTTP and generated links would otherwise downgrade to http://.
func TestRequestSchemeHonoursForwardedProto(t *testing.T) {
	tests := []struct {
		forwarded string
		want      string
	}{
		{"", "http"},
		{"https", "https"},
		{"HTTPS", "https"},
		{" https ", "https"},
		{"https, http", "https"},
		{"http", "http"},
		{"gopher", "http"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tt.forwarded != "" {
			req.Header.Set("X-Forwarded-Proto", tt.forwarded)
		}
		if got := requestScheme(req); got != tt.want {
			t.Errorf("X-Forwarded-Proto %q: scheme = %q, want %q", tt.forwarded, got, tt.want)
		}
	}
}
