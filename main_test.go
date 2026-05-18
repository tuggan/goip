package main

import (
	"bytes"
	"io"
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
