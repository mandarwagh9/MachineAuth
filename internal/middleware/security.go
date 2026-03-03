package middleware

import (
	"net/http"
	"sync"
	"time"
)

// ── Rate Limiter ─────────────────────────────────────────────────────

// RateLimiterConfig holds per-bucket configuration.
type RateLimiterConfig struct {
	// Requests allowed per window.
	Limit int
	// Window duration.
	Window time.Duration
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter is an in-memory IP-based rate limiter using a fixed window.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
	cfg     RateLimiterConfig
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*rateBucket),
		cfg:     cfg,
	}
	// Periodically clean up expired buckets.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()
	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for key, b := range rl.buckets {
		if now.After(b.resetAt) {
			delete(rl.buckets, key)
		}
	}
}

// Allow checks if the given key is within the rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]
	if !exists || now.After(b.resetAt) {
		rl.buckets[key] = &rateBucket{
			count:   1,
			resetAt: now.Add(rl.cfg.Window),
		}
		return true
	}

	if b.count >= rl.cfg.Limit {
		return false
	}
	b.count++
	return true
}

// RateLimit wraps an http.Handler and rejects requests that exceed the rate.
// The key is derived from the client IP (RemoteAddr).
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			if !rl.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate_limit_exceeded","error_description":"too many requests, please try again later"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (behind proxy/LB).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Use the first IP in the chain.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr.
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

// ── Security Headers ─────────────────────────────────────────────────

// SecurityHeaders adds recommended HTTP security headers to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		// HSTS — only effective over TLS but harmless otherwise.
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// ── Request Body Size Limit ──────────────────────────────────────────

const defaultMaxBodySize = 1 << 20 // 1 MB

// BodyLimit limits the size of incoming request bodies.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBodySize
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
