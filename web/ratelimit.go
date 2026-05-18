package web

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimiter implements a per-IP token bucket rate limiter using only the
// standard library. It is safe for concurrent use.
type RateLimiter struct {
	mu              sync.Mutex
	visitors        map[string]*visitor
	rate            float64
	burst           int
	cleanupInterval time.Duration
	done            chan struct{}
	stopOnce        sync.Once
}

// NewRateLimiter creates a RateLimiter with the given rate (tokens/sec),
// burst size, and cleanup interval for stale visitor entries. If rate <= 0
// the limiter is disabled (all requests pass).
func NewRateLimiter(rate float64, burst int, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:        make(map[string]*visitor),
		rate:            rate,
		burst:           burst,
		cleanupInterval: cleanupInterval,
		done:            make(chan struct{}),
	}
	if cleanupInterval > 0 {
		go rl.cleanup()
	}
	return rl
}

// cleanup runs in a background goroutine and periodically removes stale
// visitors that have not been seen for more than 2× the cleanup interval.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-rl.done:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.Sub(v.lastCheck) > 2*rl.cleanupInterval {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Stop terminates the background cleanup goroutine. Safe to call multiple
// times; subsequent calls are no-ops.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.done)
	})
}

// Allow reports whether a request from the given IP should be permitted.
// If the rate limiter is disabled (rate <= 0) every call returns true.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.rate <= 0 {
		return true
	}

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		v = &visitor{tokens: float64(rl.burst), lastCheck: now}
		rl.visitors[ip] = v
	}

	elapsed := now.Sub(v.lastCheck)
	v.tokens += elapsed.Seconds() * rl.rate
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.lastCheck = now

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

// Middleware returns an http.Handler that rate-limits incoming requests by
// client IP. When a request is denied it responds with 429 Too Many Requests
// and a Retry-After header.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If we cannot parse the remote address, let the request
			// through rather than rejecting it.
			next.ServeHTTP(w, r)
			return
		}
		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 Too Many Requests"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
