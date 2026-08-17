package api

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter is a small in-memory token-bucket limiter keyed by client IP.
// It is per-process (not shared across replicas) and meant as basic abuse
// protection for a read-only mock service, not a hard quota. A nil or
// disabled limiter lets every request through.
type RateLimiter struct {
	rps   float64 // tokens refilled per second
	burst float64 // maximum tokens a bucket can hold

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

// idleTTL is how long a client's bucket is kept after its last request before
// the janitor evicts it, bounding memory under churn.
const idleTTL = 10 * time.Minute

// NewRateLimiter builds a limiter allowing bursts of burst requests and rps
// sustained requests per second per client. If rps or burst is non-positive
// the limiter is disabled and Middleware becomes a pass-through.
func NewRateLimiter(rps, burst int) *RateLimiter {
	if rps <= 0 || burst <= 0 {
		return &RateLimiter{}
	}
	rl := &RateLimiter{
		rps:     float64(rps),
		burst:   float64(burst),
		buckets: make(map[string]*bucket),
	}
	go rl.janitor()
	return rl
}

func (rl *RateLimiter) enabled() bool { return rl != nil && rl.buckets != nil }

// allow reports whether a request from key may proceed, consuming one token
// when it can.
func (rl *RateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b := rl.buckets[key]
	if b == nil {
		// A fresh client starts with a full bucket, minus the current request.
		rl.buckets[key] = &bucket{tokens: rl.burst - 1, last: now}
		return true
	}

	b.tokens += now.Sub(b.last).Seconds() * rl.rps
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// janitor periodically drops buckets that have been idle long enough to have
// refilled completely, so the map does not grow without bound.
func (rl *RateLimiter) janitor() {
	ticker := time.NewTicker(idleTTL)
	defer ticker.Stop()
	for now := range ticker.C {
		rl.mu.Lock()
		for key, b := range rl.buckets {
			if now.Sub(b.last) > idleTTL {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware throttles requests per client IP, returning 429 with a
// Retry-After header when a client exceeds its rate. It is a pass-through when
// the limiter is disabled.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	if !rl.enabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r), time.Now()) {
			retry := int(math.Ceil(1 / rl.rps))
			if retry < 1 {
				retry = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP resolves the caller's IP, honouring the leftmost X-Forwarded-For
// entry set by a proxy (as on Render) and falling back to RemoteAddr. Like
// the scheme header, this is trusted: on a direct hit it can be spoofed, but
// the only effect is which bucket a request counts against.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			xff = xff[:i]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
