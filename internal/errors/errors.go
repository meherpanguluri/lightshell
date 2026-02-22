package errors

import "fmt"

// Error codes for LightShell errors.
const (
	// Filesystem errors
	FSNotFound         = "FS_NOT_FOUND"
	FSPermissionDenied = "FS_PERMISSION_DENIED"
	FSReadError        = "FS_READ_ERROR"
	FSWriteError       = "FS_WRITE_ERROR"
	FSPathTraversal    = "FS_PATH_TRAVERSAL"

	// Permission errors
	PermissionDenied = "PERMISSION_DENIED"

	// HTTP errors
	HTTPTimeout       = "HTTP_TIMEOUT"
	HTTPRequestFailed = "HTTP_REQUEST_FAILED"
	HTTPDomainDenied  = "HTTP_DOMAIN_DENIED"

	// Process errors
	ProcessDenied   = "PROCESS_DENIED"
	ProcessTimeout  = "PROCESS_TIMEOUT"
	ProcessFailed   = "PROCESS_FAILED"
	ProcessNotFound = "PROCESS_NOT_FOUND"

	// Platform errors
	PlatformUnsupported = "PLATFORM_UNSUPPORTED"

	// Store errors
	StoreReadError  = "STORE_READ_ERROR"
	StoreWriteError = "STORE_WRITE_ERROR"

	// Updater errors
	UpdaterCheckFailed   = "UPDATER_CHECK_FAILED"
	UpdaterHashMismatch  = "UPDATER_HASH_MISMATCH"
	UpdaterInstallFailed = "UPDATER_INSTALL_FAILED"

	// Config errors
	ConfigInvalid  = "CONFIG_INVALID"
	ConfigNotFound = "CONFIG_NOT_FOUND"

	// Signing errors
	SigningKeyNotFound = "SIGNING_KEY_NOT_FOUND"
	SigningFailed      = "SIGNING_FAILED"

	// Release errors
	ReleaseFailed = "RELEASE_FAILED"
	UploadFailed  = "UPLOAD_FAILED"
)

// LightShellError is a structured error with code, context, and remediation info.
type LightShellError struct {
	Namespace string // e.g. "fs", "http", "process"
	Method    string // e.g. "readFile", "fetch", "exec"
	Code      string // e.g. FS_NOT_FOUND
	Message   string // human-readable description
	Cause     error  // underlying error, if any
	Fix       string // how to fix it
	DocsURL   string // link to relevant docs
}

// Error implements the error interface with an AI-debuggable format.
func (e *LightShellError) Error() string {
	header := fmt.Sprintf("LightShell Error [%s.%s]: %s", e.Namespace, e.Method, e.Message)
	if e.Fix != "" {
		header += fmt.Sprintf("\n  -> %s", e.Fix)
	}
	if e.DocsURL != "" {
		header += fmt.Sprintf("\n  -> Docs: %s", e.DocsURL)
	}
	if e.Cause != nil {
		header += fmt.Sprintf("\n  -> Cause: %s", e.Cause.Error())
	}
	return header
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *LightShellError) Unwrap() error {
	return e.Cause
}

// New creates a new LightShellError with the given fields.
func New(namespace, method, code, message string) *LightShellError {
	return &LightShellError{
		Namespace: namespace,
		Method:    method,
		Code:      code,
		Message:   message,
	}
}

// WithCause sets the underlying cause.
func (e *LightShellError) WithCause(cause error) *LightShellError {
	e.Cause = cause
	return e
}

// WithFix sets the remediation message.
func (e *LightShellError) WithFix(fix string) *LightShellError {
	e.Fix = fix
	return e
}

// WithDocs sets the documentation URL.
func (e *LightShellError) WithDocs(url string) *LightShellError {
	e.DocsURL = url
	return e
}

// --- Helper constructors for common error types ---

// PermissionError creates a permission denied error with AI-friendly context.
func PermissionError(namespace, method, attempted string, allowed []string, configKey string) *LightShellError {
	msg := fmt.Sprintf("Permission denied")
	fix := fmt.Sprintf("Attempted: %s\n  -> Allowed: %v\n  -> To allow this, update %s in lightshell.json", attempted, allowed, configKey)
	return &LightShellError{
		Namespace: namespace,
		Method:    method,
		Code:      PermissionDenied,
		Message:   msg,
		Fix:       fix,
		DocsURL:   "https://lightshell.dev/guides/security-and-permissions",
	}
}

// FSError creates a filesystem error.
func FSError(method, code, message string, cause error) *LightShellError {
	return &LightShellError{
		Namespace: "fs",
		Method:    method,
		Code:      code,
		Message:   message,
		Cause:     cause,
		DocsURL:   "https://lightshell.dev/guides/security-and-permissions#fs",
	}
}

// PathTraversalError creates a path traversal error with allowed paths listed.
func PathTraversalError(method, path string, allowedPatterns []string) *LightShellError {
	return &LightShellError{
		Namespace: "fs",
		Method:    method,
		Code:      FSPathTraversal,
		Message:   "Permission denied",
		Fix:       fmt.Sprintf("Attempted to access: %s\n  -> Allowed paths: %v\n  -> To allow this path, update permissions.fs.read in lightshell.json", path, allowedPatterns),
		DocsURL:   "https://lightshell.dev/guides/security-and-permissions#fs",
	}
}

// HTTPError creates an HTTP request error.
func HTTPError(method, code, message string, cause error) *LightShellError {
	return &LightShellError{
		Namespace: "http",
		Method:    method,
		Code:      code,
		Message:   message,
		Cause:     cause,
		DocsURL:   "https://lightshell.dev/guides/security-and-permissions#http",
	}
}

// ProcessError creates a process execution error.
func ProcessError(method, code, message string, cause error) *LightShellError {
	return &LightShellError{
		Namespace: "process",
		Method:    method,
		Code:      code,
		Message:   message,
		Cause:     cause,
		DocsURL:   "https://lightshell.dev/guides/security-and-permissions#process",
	}
}

// PlatformError creates a platform unsupported error.
func PlatformError(namespace, method, platform string) *LightShellError {
	return &LightShellError{
		Namespace: namespace,
		Method:    method,
		Code:      PlatformUnsupported,
		Message:   fmt.Sprintf("Not supported on %s", platform),
		DocsURL:   "https://lightshell.dev/guides/platform-support",
	}
}
