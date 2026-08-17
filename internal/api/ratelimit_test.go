package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsBurstThenBlocks(t *testing.T) {
	rl := NewRateLimiter(1, 3)
	now := time.Now()

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4", now) {
			t.Fatalf("request %d within burst should be allowed", i+1)
		}
	}
	if rl.allow("1.2.3.4", now) {
		t.Error("request beyond burst should be blocked")
	}

	// A different client has its own bucket.
	if !rl.allow("5.6.7.8", now) {
		t.Error("a different client should not be throttled")
	}
}

func TestRateLimiterRefillsOverTime(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	now := time.Now()

	if !rl.allow("1.2.3.4", now) {
		t.Fatal("first request should be allowed")
	}
	if rl.allow("1.2.3.4", now) {
		t.Fatal("second immediate request should be blocked")
	}
	if !rl.allow("1.2.3.4", now.Add(2*time.Second)) {
		t.Error("request after refill window should be allowed")
	}
}

func TestRateLimiterDisabledPassesThrough(t *testing.T) {
	rl := NewRateLimiter(0, 0)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/types", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200 (limiter disabled)", i+1, rec.Code)
		}
	}
}

func TestRateLimiterMiddlewareReturns429(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	call := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/types", nil)
		req.RemoteAddr = "8.8.8.8:5555"
		handler.ServeHTTP(rec, req)
		return rec
	}

	if rec := call(); rec.Code != http.StatusOK {
		t.Fatalf("first call status = %d, want 200", rec.Code)
	}
	rec := call()
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second call status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("429 response should set Retry-After")
	}
}

func TestClientIPPrefersForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 70.41.3.18")

	if got := clientIP(req); got != "203.0.113.7" {
		t.Errorf("clientIP = %q, want %q", got, "203.0.113.7")
	}

	req.Header.Del("X-Forwarded-For")
	if got := clientIP(req); got != "10.0.0.1" {
		t.Errorf("clientIP fallback = %q, want %q", got, "10.0.0.1")
	}
}
