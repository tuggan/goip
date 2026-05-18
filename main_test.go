package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestPrintVersion(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	printVersion()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "GoIP") {
		t.Errorf("expected output to contain 'GoIP', got: %q", output)
	}
}

func TestPrintHelp(t *testing.T) {
	// Flags need to be registered before printHelp will show them
	// (they're normally registered in main()). We register a minimal set here.
	pflag.StringP("endpoint", "e", "127.0.0.1:3000", "Endpoint to listen on")
	pflag.String("tlsEndpoint", "127.0.0.1:3000", "TLS endpoint to listen on")
	pflag.String("tlsCert", "", "Path to TLS certificate file")
	pflag.String("tlsKey", "", "Path to TLS key file")
	pflag.StringP("config", "c", ".", "Path to config file")

	// Capture both stdout and stderr (pflag.PrintDefaults writes to stderr)
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	printHelp()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "Usage") {
		t.Errorf("expected output to contain 'Usage', got: %q", output)
	}
	if !strings.Contains(output, "endpoint") {
		t.Errorf("expected output to contain 'endpoint', got: %q", output)
	}
	if !strings.Contains(output, "config") {
		t.Errorf("expected output to contain 'config', got: %q", output)
	}
}

func TestSecurityHeadersMiddleware_SetsHeaders(t *testing.T) {
	var gotHeaders http.Header
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = w.Header().Clone()
		w.WriteHeader(http.StatusOK)
	})

	middleware := securityHeadersMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if gotHeaders.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", gotHeaders.Get("X-Content-Type-Options"))
	}
	if gotHeaders.Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %q", gotHeaders.Get("X-Frame-Options"))
	}
	if gotHeaders.Get("Content-Security-Policy") == "" {
		t.Error("expected Content-Security-Policy to be set")
	}
}

func TestSecurityHeadersMiddleware_HSTSWithoutTLS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := securityHeadersMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No TLS on request
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if h := resp.Header.Get("Strict-Transport-Security"); h != "" {
		t.Errorf("expected no HSTS header without TLS, got %q", h)
	}
}

func TestSecurityHeadersMiddleware_HSTSWithTLS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := securityHeadersMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Simulate TLS by setting req.TLS
	req.TLS = &tls.ConnectionState{}
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if h := resp.Header.Get("Strict-Transport-Security"); h == "" {
		t.Error("expected HSTS header with TLS, got empty")
	} else if !strings.Contains(h, "max-age=63072000") {
		t.Errorf("expected HSTS max-age=63072000, got %q", h)
	}
}

func TestPrintVersion_EmptyVersion(t *testing.T) {
	// Save and restore Version
	origVersion := Version
	Version = ""
	defer func() { Version = origVersion }()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printVersion()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "GoIP") {
		t.Errorf("expected output to contain 'GoIP' even with empty version, got: %q", output)
	}
}
