package web

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tuggan/goip/logger"
)

func TestMain(m *testing.M) {
	logger.Init(io.Discard, os.Stdout, os.Stdout, os.Stderr)
	os.Exit(m.Run())
}

// helpers

func testHandler() handler {
	return NewHandler(false, "../html", "test-version", "test-branch",
		"2024-01-01", "Test Author", "test@example.com", nil)
}

func testHandlerWithGzip() handler {
	return NewHandler(true, "../html", "test-version", "test-branch",
		"2024-01-01", "Test Author", "test@example.com", nil)
}

func mustReadBody(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(b)
}

// ----------------
// MainHandler — route tests
// ----------------

func TestMainHandler_Root(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html; charset=utf-8, got %q", ct)
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "192.168.1.1") {
		t.Errorf("expected body to contain client IP, got:\n%s", body)
	}
	if !strings.Contains(body, "GoIP test-version") {
		t.Errorf("expected body to contain version info, got:\n%s", body)
	}
}

func TestMainHandler_IP(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.5:9999"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "10.0.0.5" {
		t.Errorf("expected IP '10.0.0.5', got %q", body)
	}
}

func TestMainHandler_UserAgent(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/user-agent", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "TestAgent/1.0" {
		t.Errorf("expected 'TestAgent/1.0', got %q", body)
	}
}

func TestMainHandler_Host(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/host", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Host = "example.com"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "example.com" {
		t.Errorf("expected 'example.com', got %q", body)
	}
}

func TestMainHandler_Proto(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/proto", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Proto = "HTTP/2.0"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "HTTP/2.0" {
		t.Errorf("expected 'HTTP/2.0', got %q", body)
	}
}

func TestMainHandler_Accept(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/accept", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "application/json" {
		t.Errorf("expected 'application/json', got %q", body)
	}
}

func TestMainHandler_AcceptEncoding(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/accept-encoding", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "gzip, deflate" {
		t.Errorf("expected 'gzip, deflate', got %q", body)
	}
}

func TestMainHandler_AcceptLanguage(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/accept-language", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "en-US,en;q=0.9" {
		t.Errorf("expected 'en-US,en;q=0.9', got %q", body)
	}
}

func TestMainHandler_Method(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodPost, "/method", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != http.MethodPost {
		t.Errorf("expected 'POST', got %q", body)
	}
}

func TestMainHandler_ContentType(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/content-type", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "application/json" {
		t.Errorf("expected 'application/json', got %q", body)
	}
}

func TestMainHandler_Origin(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/origin", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "https://example.com" {
		t.Errorf("expected 'https://example.com', got %q", body)
	}
}

func TestMainHandler_Referer(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/referer", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Referer", "https://example.com/previous")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "https://example.com/previous" {
		t.Errorf("expected 'https://example.com/previous', got %q", body)
	}
}

func TestMainHandler_XForwardedFor_Raw(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/x-forwarded-for", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "203.0.113.5, 10.0.0.1" {
		t.Errorf("expected raw X-Forwarded-For value, got %q", body)
	}
}

func TestMainHandler_NotFound(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "not found") {
		t.Errorf("expected body to contain 'not found', got:\n%s", body)
	}
}

// ----------------
// X-Forwarded-For
// ----------------

func TestMainHandler_XForwardedFor_Untrusted(t *testing.T) {
	// No trusted proxies configured — X-Forwarded-For must be ignored.
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.1:4444"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "10.0.0.1" {
		t.Errorf("expected RemoteAddr IP (X-Forwarded-For ignored), got %q", body)
	}
}

func TestMainHandler_XForwardedFor_TrustedExactIP(t *testing.T) {
	// Trusted proxy with exact IP match.
	h := NewHandler(false, "../html", "v", "b", "d", "a", "e", []string{"10.0.0.1"})
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.1:4444"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "203.0.113.5" {
		t.Errorf("expected X-Forwarded-For IP '203.0.113.5', got %q", body)
	}
}

func TestMainHandler_XForwardedFor_TrustedCIDR(t *testing.T) {
	// Trusted proxy from a CIDR range.
	h := NewHandler(false, "../html", "v", "b", "d", "a", "e", []string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.1:4444"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "203.0.113.5" {
		t.Errorf("expected X-Forwarded-For IP '203.0.113.5', got %q", body)
	}
}

func TestMainHandler_XForwardedFor_TrustedNotMatching(t *testing.T) {
	// Trusted proxy configured, but request is from a different IP.
	h := NewHandler(false, "../html", "v", "b", "d", "a", "e", []string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "192.168.1.1:4444"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "192.168.1.1" {
		t.Errorf("expected RemoteAddr IP (not in trusted range), got %q", body)
	}
}

func TestMainHandler_NoXForwardedFor(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.1:4444"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "10.0.0.1" {
		t.Errorf("expected RemoteAddr IP '10.0.0.1', got %q", body)
	}
}

// ----------------
// Gzip
// ----------------

func TestMainHandler_GzipEnabledAndAccepted(t *testing.T) {
	h := testHandlerWithGzip()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got %q", enc)
	}

	// Verify body is valid gzip
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()
	body := mustReadBody(t, gr)
	if !strings.Contains(body, "1.2.3.4") {
		t.Errorf("expected decompressed body to contain IP, got:\n%s", body)
	}
}

func TestMainHandler_GzipEnabledNotAccepted(t *testing.T) {
	h := testHandlerWithGzip()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	// No Accept-Encoding header
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc == "gzip" {
		t.Errorf("expected no gzip encoding when client doesn't accept it")
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "1.2.3.4") {
		t.Errorf("expected body to contain IP, got:\n%s", body)
	}
}

func TestMainHandler_GzipDisabled(t *testing.T) {
	h := testHandler() // gzip disabled
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if enc := resp.Header.Get("Content-Encoding"); enc == "gzip" {
		t.Errorf("expected no gzip when disabled")
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "1.2.3.4") {
		t.Errorf("expected body to contain IP, got:\n%s", body)
	}
}

// ----------------
// Server header
// ----------------

func TestServerHeader(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if server := resp.Header.Get("Server"); server != "GoIP test-version" {
		t.Errorf("expected Server 'GoIP test-version', got %q", server)
	}
}

// ----------------
// GETHandler
// ----------------

func TestGETHandler_GET(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/GET?foo=bar&baz=1", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.GETHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "foo=bar&baz=1" {
		t.Errorf("expected 'foo=bar&baz=1', got %q", body)
	}
}

func TestGETHandler_POST(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodPost, "/GET", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.GETHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "method not GET") {
		t.Errorf("expected body to mention 'method not GET', got:\n%s", body)
	}
}

func TestGETHandler_EmptyQuery(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/GET", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.GETHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := strings.TrimSpace(mustReadBody(t, resp.Body))
	if body != "" {
		t.Errorf("expected empty body, got %q", body)
	}
}

// ----------------
// FaviconHandler
// ----------------

func TestFaviconHandler_Success(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.FaviconHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := mustReadBody(t, resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty favicon body")
	}
}

func TestFaviconHandler_Missing(t *testing.T) {
	h := NewHandler(false, t.TempDir(), "v", "b", "d", "a", "e", nil)
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.FaviconHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for missing favicon, got %d", resp.StatusCode)
	}
}

// ----------------
// RobotsHandler
// ----------------

func TestRobotsHandler_Success(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.RobotsHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body := mustReadBody(t, resp.Body)
	if !strings.Contains(body, "User-agent:") {
		t.Errorf("expected robots.txt content, got:\n%s", body)
	}
}

func TestRobotsHandler_Missing(t *testing.T) {
	h := NewHandler(false, t.TempDir(), "v", "b", "d", "a", "e", nil)
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	h.RobotsHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for missing robots.txt, got %d", resp.StatusCode)
	}
}

// ----------------
// RemoteAddr parsing failure (edge case)
// ----------------

func TestMainHandler_MissingRemoteAddr(t *testing.T) {
	h := testHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Set empty RemoteAddr so SplitHostPort fails
	req.RemoteAddr = ""
	w := httptest.NewRecorder()

	h.MainHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for missing/invalid RemoteAddr, got %d", resp.StatusCode)
	}
}

// ----------------
// NewHandler defaults
// ----------------

func TestNewHandler_SetsServerField(t *testing.T) {
	h := NewHandler(false, "/templates", "1.0.0", "main", "2024-06-15", "Alice", "alice@example.com", nil)
	expected := fmt.Sprintf("GoIP %s", "1.0.0")
	if h.server != expected {
		t.Errorf("expected server %q, got %q", expected, h.server)
	}
}
