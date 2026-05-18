package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInitAndInfo(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	Info("hello %s", "world")
	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected Info output to contain 'hello world', got: %q", out)
	}
}

func TestInitAndWarning(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	Warning("careful %s", "now")
	out := buf.String()
	if !strings.Contains(out, "careful now") {
		t.Errorf("expected Warning output to contain 'careful now', got: %q", out)
	}
}

func TestInitAndError(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	Error("something %s", "broke")
	out := buf.String()
	if !strings.Contains(out, "something broke") {
		t.Errorf("expected Error output to contain 'something broke', got: %q", out)
	}
}

func TestAccess(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	Access(req, http.StatusOK)

	out := buf.String()
	if !strings.Contains(out, "200") {
		t.Errorf("expected Access output to contain status '200', got: %q", out)
	}
	if !strings.Contains(out, "/test-path") {
		t.Errorf("expected Access output to contain path '/test-path', got: %q", out)
	}
	if !strings.Contains(out, "10.0.0.1:8080") {
		t.Errorf("expected Access output to contain remote addr, got: %q", out)
	}
}

func TestAccess_DifferentStatus(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	Access(req, http.StatusNotFound)

	out := buf.String()
	if !strings.Contains(out, "404") {
		t.Errorf("expected Access output to contain status '404', got: %q", out)
	}
}

func TestOutputSeparation(t *testing.T) {
	var infoBuf bytes.Buffer
	var errBuf bytes.Buffer
	Init(&bytes.Buffer{}, &infoBuf, &bytes.Buffer{}, &errBuf)

	Info("info message")
	Error("error message")

	if !strings.Contains(infoBuf.String(), "info message") {
		t.Error("expected info buffer to contain info message")
	}
	if strings.Contains(infoBuf.String(), "error message") {
		t.Error("expected info buffer NOT to contain error message")
	}
	if !strings.Contains(errBuf.String(), "error message") {
		t.Error("expected error buffer to contain error message")
	}
	if strings.Contains(errBuf.String(), "info message") {
		t.Error("expected error buffer NOT to contain info message")
	}
}

func TestMultipleLogs(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	Info("first")
	Warning("second")
	Info("third")

	out := buf.String()
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") || !strings.Contains(out, "third") {
		t.Errorf("expected all log entries, got: %q", out)
	}
}

func TestEmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	Init(&buf, &buf, &buf, &buf)

	Info("")
	out := buf.String()
	if out == "" {
		t.Error("expected some output even for empty message (timestamp, etc.)")
	}
}
