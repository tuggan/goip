package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ----------------
// RateLimiter.Allow
// ----------------

func TestAllow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(1000, 1000, time.Minute)
	for i := 0; i < 50; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Errorf("request %d should be allowed (under limit)", i+1)
		}
	}
}

func TestAllow_OverLimit(t *testing.T) {
	rl := NewRateLimiter(10, 10, time.Minute)
	var allowed, blocked int
	for i := 0; i < 30; i++ {
		if rl.Allow("1.2.3.4") {
			allowed++
		} else {
			blocked++
		}
	}
	if allowed != 10 {
		t.Errorf("expected 10 allowed, got %d", allowed)
	}
	if blocked != 20 {
		t.Errorf("expected 20 blocked, got %d", blocked)
	}
}

func TestAllow_ZeroRate(t *testing.T) {
	rl := NewRateLimiter(0, 0, time.Minute)
	for i := 0; i < 100; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Errorf("request %d should be allowed when rate is 0", i+1)
		}
	}
}

func TestAllow_Burst(t *testing.T) {
	rl := NewRateLimiter(1, 5, time.Minute) // 1 token/sec, burst 5
	// First 5 requests should all pass (burst capacity).
	for i := 0; i < 5; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Errorf("request %d should be allowed (burst)", i+1)
		}
	}
	// 6th request must be blocked (no time elapsed to refill).
	if rl.Allow("1.2.3.4") {
		t.Errorf("6th request should be blocked (burst exhausted)")
	}
}

func TestAllow_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute)
	// Each IP gets its own bucket.
	if !rl.Allow("10.0.0.1") {
		t.Errorf("first request from 10.0.0.1 should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Errorf("second request from 10.0.0.1 should be blocked")
	}
	if !rl.Allow("10.0.0.2") {
		t.Errorf("first request from 10.0.0.2 should be allowed")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(10, 10, 100*time.Millisecond)
	// Should not deadlock or panic.
	rl.Stop()
	// Allow should still work after Stop.
	if !rl.Allow("1.2.3.4") {
		t.Errorf("Allow should work after Stop")
	}
}

func TestRateLimiter_StopMultiple(t *testing.T) {
	rl := NewRateLimiter(10, 10, 100*time.Millisecond)
	rl.Stop()
	// Calling Stop again must not panic (guarded by sync.Once).
	rl.Stop()
}

// ----------------
// RateLimiter.Middleware
// ----------------

func TestRateLimiter_Middleware_BlocksExcess(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute) // 1 req/sec, burst 1
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First request should pass.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", w1.Code)
	}

	// Second request (same IP, no time elapsed) should be rate-limited.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", w2.Code)
	}
	if w2.Header().Get("Retry-After") != "1" {
		t.Errorf("expected Retry-After: 1, got %q", w2.Header().Get("Retry-After"))
	}
	if w2.Body.String() != "429 Too Many Requests" {
		t.Errorf("expected body %q, got %q", "429 Too Many Requests", w2.Body.String())
	}
}

func TestRateLimiter_Middleware_InvalidRemoteAddr(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute)
	var called bool
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "" // will cause SplitHostPort to fail
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called when RemoteAddr is invalid")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiter_Middleware_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute) // 1 req/sec, burst 1 per IP
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP — first request passes.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("10.0.0.1 request 1: expected 200, got %d", w1.Code)
	}

	// First IP — second request blocked.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.1:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("10.0.0.1 request 2: expected 429, got %d", w2.Code)
	}

	// Second IP — first request passes (separate bucket).
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "10.0.0.2:54321"
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("10.0.0.2 request 1: expected 200, got %d", w3.Code)
	}
}
