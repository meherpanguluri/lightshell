package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNewCreatesError(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "file not found")

	if err.Namespace != "fs" {
		t.Errorf("expected namespace 'fs', got %q", err.Namespace)
	}
	if err.Method != "readFile" {
		t.Errorf("expected method 'readFile', got %q", err.Method)
	}
	if err.Code != FSNotFound {
		t.Errorf("expected code %q, got %q", FSNotFound, err.Code)
	}
	if err.Message != "file not found" {
		t.Errorf("expected message 'file not found', got %q", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("expected nil cause, got %v", err.Cause)
	}
	if err.Fix != "" {
		t.Errorf("expected empty fix, got %q", err.Fix)
	}
	if err.DocsURL != "" {
		t.Errorf("expected empty docs URL, got %q", err.DocsURL)
	}
}

func TestErrorFormat(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "file not found")
	msg := err.Error()

	if !strings.Contains(msg, "LightShell Error [fs.readFile]: file not found") {
		t.Errorf("expected formatted error, got %q", msg)
	}
}

func TestErrorFormatWithFix(t *testing.T) {
	err := New("fs", "readFile", FSPermissionDenied, "permission denied").
		WithFix("Add 'fs' to permissions in lightshell.json")
	msg := err.Error()

	if !strings.Contains(msg, "permission denied") {
		t.Errorf("expected message in error, got %q", msg)
	}
	if !strings.Contains(msg, "Add 'fs' to permissions") {
		t.Errorf("expected fix in error, got %q", msg)
	}
}

func TestErrorFormatWithDocs(t *testing.T) {
	err := New("http", "fetch", HTTPTimeout, "request timed out").
		WithDocs("https://lightshell.dev/guides/http")
	msg := err.Error()

	if !strings.Contains(msg, "Docs: https://lightshell.dev/guides/http") {
		t.Errorf("expected docs URL in error, got %q", msg)
	}
}

func TestErrorFormatWithCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := New("http", "fetch", HTTPRequestFailed, "request failed").
		WithCause(cause)
	msg := err.Error()

	if !strings.Contains(msg, "Cause: connection refused") {
		t.Errorf("expected cause in error, got %q", msg)
	}
}

func TestErrorFormatComplete(t *testing.T) {
	cause := fmt.Errorf("ENOENT")
	err := New("fs", "readFile", FSNotFound, "file not found").
		WithCause(cause).
		WithFix("Check that the file exists").
		WithDocs("https://lightshell.dev/guides/fs")
	msg := err.Error()

	expected := []string{
		"LightShell Error [fs.readFile]: file not found",
		"Check that the file exists",
		"Docs: https://lightshell.dev/guides/fs",
		"Cause: ENOENT",
	}
	for _, s := range expected {
		if !strings.Contains(msg, s) {
			t.Errorf("expected %q in error message, got %q", s, msg)
		}
	}
}

func TestWithCauseChaining(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "not found").
		WithCause(fmt.Errorf("underlying"))

	if err.Cause == nil {
		t.Fatal("expected cause to be set")
	}
	if err.Cause.Error() != "underlying" {
		t.Errorf("expected cause 'underlying', got %q", err.Cause.Error())
	}
}

func TestWithFixChaining(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "not found").
		WithFix("try again")

	if err.Fix != "try again" {
		t.Errorf("expected fix 'try again', got %q", err.Fix)
	}
}

func TestWithDocsChaining(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "not found").
		WithDocs("https://example.com")

	if err.DocsURL != "https://example.com" {
		t.Errorf("expected docs URL, got %q", err.DocsURL)
	}
}

func TestFluentChaining(t *testing.T) {
	err := New("http", "fetch", HTTPTimeout, "timeout").
		WithCause(fmt.Errorf("deadline exceeded")).
		WithFix("Increase timeout").
		WithDocs("https://docs.example.com")

	if err.Cause == nil {
		t.Fatal("expected cause")
	}
	if err.Fix != "Increase timeout" {
		t.Errorf("expected fix, got %q", err.Fix)
	}
	if err.DocsURL != "https://docs.example.com" {
		t.Errorf("expected docs, got %q", err.DocsURL)
	}
}

func TestUnwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := New("fs", "readFile", FSNotFound, "not found").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("expected unwrapped to equal cause")
	}
}

func TestUnwrapNil(t *testing.T) {
	err := New("fs", "readFile", FSNotFound, "not found")
	if err.Unwrap() != nil {
		t.Errorf("expected nil unwrap, got %v", err.Unwrap())
	}
}

func TestErrorsIs(t *testing.T) {
	cause := fmt.Errorf("root")
	err := New("fs", "readFile", FSNotFound, "not found").WithCause(cause)

	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to find the cause")
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	var err error = New("test", "method", "CODE", "message")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error string")
	}
}

// --- Test helper constructors ---

func TestPermissionErrorHelper(t *testing.T) {
	err := PermissionError("fs", "readFile", "/etc/passwd", []string{"/tmp/**"}, "permissions.fs.read")

	if err.Code != PermissionDenied {
		t.Errorf("expected code %q, got %q", PermissionDenied, err.Code)
	}
	if err.Namespace != "fs" {
		t.Errorf("expected namespace 'fs', got %q", err.Namespace)
	}
	if !strings.Contains(err.Fix, "/etc/passwd") {
		t.Errorf("expected attempted path in fix, got %q", err.Fix)
	}
	if err.DocsURL == "" {
		t.Error("expected docs URL to be set")
	}
}

func TestFSErrorHelper(t *testing.T) {
	cause := fmt.Errorf("ENOENT")
	err := FSError("readFile", FSNotFound, "file not found", cause)

	if err.Namespace != "fs" {
		t.Errorf("expected namespace 'fs', got %q", err.Namespace)
	}
	if err.Method != "readFile" {
		t.Errorf("expected method 'readFile', got %q", err.Method)
	}
	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
	if err.DocsURL == "" {
		t.Error("expected docs URL to be set")
	}
}

func TestPathTraversalErrorHelper(t *testing.T) {
	err := PathTraversalError("readFile", "/etc/passwd", []string{"/tmp/**"})

	if err.Code != FSPathTraversal {
		t.Errorf("expected code %q, got %q", FSPathTraversal, err.Code)
	}
	if !strings.Contains(err.Fix, "/etc/passwd") {
		t.Errorf("expected path in fix, got %q", err.Fix)
	}
}

func TestHTTPErrorHelper(t *testing.T) {
	err := HTTPError("fetch", HTTPTimeout, "timeout", nil)

	if err.Namespace != "http" {
		t.Errorf("expected namespace 'http', got %q", err.Namespace)
	}
	if err.Code != HTTPTimeout {
		t.Errorf("expected code %q, got %q", HTTPTimeout, err.Code)
	}
}

func TestProcessErrorHelper(t *testing.T) {
	cause := fmt.Errorf("exit code 1")
	err := ProcessError("exec", ProcessFailed, "command failed", cause)

	if err.Namespace != "process" {
		t.Errorf("expected namespace 'process', got %q", err.Namespace)
	}
	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
}

func TestPlatformErrorHelper(t *testing.T) {
	err := PlatformError("tray", "set", "linux")

	if err.Code != PlatformUnsupported {
		t.Errorf("expected code %q, got %q", PlatformUnsupported, err.Code)
	}
	if !strings.Contains(err.Message, "linux") {
		t.Errorf("expected platform in message, got %q", err.Message)
	}
}

func TestFSErrorHelperWithNilCause(t *testing.T) {
	err := FSError("writeFile", FSWriteError, "write failed", nil)

	if err.Cause != nil {
		t.Errorf("expected nil cause, got %v", err.Cause)
	}
	msg := err.Error()
	if strings.Contains(msg, "Cause:") {
		t.Errorf("expected no cause line in output, got %q", msg)
	}
}
