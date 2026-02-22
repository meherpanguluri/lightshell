package security

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Permission represents an API permission that an app can request.
type Permission string

const (
	PermFS           Permission = "fs"
	PermDialog       Permission = "dialog"
	PermClipboard    Permission = "clipboard"
	PermShell        Permission = "shell"
	PermNotification Permission = "notification"
	PermTray         Permission = "tray"
	PermMenu         Permission = "menu"
	PermHTTP         Permission = "http"
	PermProcess      Permission = "process"
	PermStore        Permission = "store"
	PermShortcuts    Permission = "shortcuts"
	PermUpdater      Permission = "updater"
	// window, system, and app are always allowed -- they're core APIs
)

// AllPermissions is the set of all declarable permissions.
var AllPermissions = []Permission{
	PermFS, PermDialog, PermClipboard, PermShell,
	PermNotification, PermTray, PermMenu,
	PermHTTP, PermProcess, PermStore, PermShortcuts, PermUpdater,
}

// FSScope holds scoped filesystem permission patterns.
type FSScope struct {
	Read  []string `json:"read"`
	Write []string `json:"write"`
}

// HTTPScope holds scoped HTTP permission patterns.
type HTTPScope struct {
	Allow []string `json:"allow"` // allowed domain patterns (e.g., "*.github.com", "api.example.com")
	Deny  []string `json:"deny"`  // denied domain patterns (checked first)
}

// ProcessScope holds scoped process execution permissions.
type ProcessScope struct {
	Exec []ProcessRule `json:"exec"`
}

// ProcessRule defines an allowed command and its permitted arguments.
type ProcessRule struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"` // if empty or contains "*", any args allowed
}

// Policy holds the security policy for a running app.
type Policy struct {
	mu          sync.RWMutex
	permissions map[Permission]bool
	allowedDirs []string // directories the app can access via fs APIs
	devMode     bool     // dev mode disables restrictions
	appName     string

	// Scoped permissions (used when permissions key has detailed config)
	fsScope      *FSScope
	httpScope    *HTTPScope
	processScope *ProcessScope
}

// NewPolicy creates a security policy from the declared permissions.
func NewPolicy(perms []string, projectDir string, appName string, devMode bool) *Policy {
	p := &Policy{
		permissions: make(map[Permission]bool),
		devMode:     devMode,
		appName:     appName,
	}

	for _, perm := range perms {
		p.permissions[Permission(perm)] = true
	}

	// Build allowed directories for FS access
	p.allowedDirs = []string{projectDir}

	// Always allow app data dir
	if home, err := os.UserHomeDir(); err == nil {
		dataDir := filepath.Join(home, "Library", "Application Support", appName)
		p.allowedDirs = append(p.allowedDirs, dataDir)
		// Also allow Linux data dir
		linuxDataDir := filepath.Join(home, ".local", "share", appName)
		p.allowedDirs = append(p.allowedDirs, linuxDataDir)
	}

	// Always allow temp dir
	p.allowedDirs = append(p.allowedDirs, os.TempDir())

	return p
}

// DevPolicy creates a permissive policy for development mode.
func DevPolicy() *Policy {
	p := &Policy{
		permissions: make(map[Permission]bool),
		devMode:     true,
	}
	// In dev mode, all permissions are granted
	for _, perm := range AllPermissions {
		p.permissions[perm] = true
	}
	return p
}

// SetFSScope configures scoped filesystem permissions with glob patterns.
func (p *Policy) SetFSScope(scope FSScope) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fsScope = &scope
}

// SetHTTPScope configures scoped HTTP permissions with domain patterns.
func (p *Policy) SetHTTPScope(scope HTTPScope) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.httpScope = &scope
}

// SetProcessScope configures scoped process execution permissions.
func (p *Policy) SetProcessScope(scope ProcessScope) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.processScope = &scope
}

// Check returns an error if the given permission is not granted.
func (p *Policy) Check(perm Permission) error {
	if p.devMode {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.permissions[perm] {
		return &PermissionError{
			Namespace: string(perm),
			Method:    "",
			Attempted: fmt.Sprintf("use %s API", string(perm)),
			Allowed:   p.declaredPermissions(),
			ConfigKey: "permissions",
		}
	}
	return nil
}

// CheckPath verifies that a file path is within the allowed directories.
// It resolves symlinks to prevent traversal attacks.
func (p *Policy) CheckPath(path string) error {
	if p.devMode {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return &PermissionError{
			Namespace: "fs",
			Method:    "",
			Attempted: fmt.Sprintf("access path: %s", path),
			Allowed:   []string{"valid absolute paths only"},
			ConfigKey: "permissions.fs",
		}
	}

	// Resolve symlinks to prevent symlink attacks
	resolved, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		// Directory might not exist yet (e.g., for writeFile with mkdir)
		// Fall back to checking the raw absolute path
		resolved = filepath.Dir(absPath)
	}
	checkPath := filepath.Join(resolved, filepath.Base(absPath))

	p.mu.RLock()
	defer p.mu.RUnlock()

	// If scoped FS permissions are set, check against glob patterns
	if p.fsScope != nil {
		allPatterns := make([]string, 0, len(p.fsScope.Read)+len(p.fsScope.Write))
		allPatterns = append(allPatterns, p.fsScope.Read...)
		allPatterns = append(allPatterns, p.fsScope.Write...)
		for _, pattern := range allPatterns {
			resolvedPattern := resolvePathVariable(pattern, p.appName)
			if matchGlob(checkPath, resolvedPattern) {
				return nil
			}
		}
		return &PermissionError{
			Namespace: "fs",
			Method:    "",
			Attempted: fmt.Sprintf("access: %s", path),
			Allowed:   allPatterns,
			ConfigKey: "permissions.fs.read or permissions.fs.write",
		}
	}

	// Fall back to allowed directories check
	for _, dir := range p.allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(checkPath, absDir+string(os.PathSeparator)) || checkPath == absDir {
			return nil
		}
	}

	return &PermissionError{
		Namespace: "fs",
		Method:    "",
		Attempted: fmt.Sprintf("access: %s", path),
		Allowed:   p.allowedDirs,
		ConfigKey: "permissions.fs.read or permissions.fs.write",
	}
}

// CheckFSRead verifies that the path is allowed for reading.
func (p *Policy) CheckFSRead(path string) error {
	if p.devMode {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return &PermissionError{
			Namespace: "fs",
			Method:    "readFile",
			Attempted: fmt.Sprintf("read: %s", path),
			Allowed:   []string{"valid absolute paths only"},
			ConfigKey: "permissions.fs.read",
		}
	}

	resolved := resolveRealPath(absPath)

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.fsScope != nil && len(p.fsScope.Read) > 0 {
		for _, pattern := range p.fsScope.Read {
			resolvedPattern := resolvePathVariable(pattern, p.appName)
			if matchGlob(resolved, resolvedPattern) {
				return nil
			}
		}
		return &PermissionError{
			Namespace: "fs",
			Method:    "readFile",
			Attempted: fmt.Sprintf("read: %s", path),
			Allowed:   p.fsScope.Read,
			ConfigKey: "permissions.fs.read",
		}
	}

	// Fall back to directory check
	return p.checkPathAgainstDirs(resolved, "fs", "readFile", path)
}

// CheckFSWrite verifies that the path is allowed for writing.
func (p *Policy) CheckFSWrite(path string) error {
	if p.devMode {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return &PermissionError{
			Namespace: "fs",
			Method:    "writeFile",
			Attempted: fmt.Sprintf("write: %s", path),
			Allowed:   []string{"valid absolute paths only"},
			ConfigKey: "permissions.fs.write",
		}
	}

	resolved := resolveRealPath(absPath)

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.fsScope != nil && len(p.fsScope.Write) > 0 {
		for _, pattern := range p.fsScope.Write {
			resolvedPattern := resolvePathVariable(pattern, p.appName)
			if matchGlob(resolved, resolvedPattern) {
				return nil
			}
		}
		return &PermissionError{
			Namespace: "fs",
			Method:    "writeFile",
			Attempted: fmt.Sprintf("write: %s", path),
			Allowed:   p.fsScope.Write,
			ConfigKey: "permissions.fs.write",
		}
	}

	// Fall back to directory check
	return p.checkPathAgainstDirs(resolved, "fs", "writeFile", path)
}

// CheckHTTP verifies that an HTTP request to the given URL is allowed.
func (p *Policy) CheckHTTP(rawURL string) error {
	if p.devMode {
		return nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return &PermissionError{
			Namespace: "http",
			Method:    "fetch",
			Attempted: fmt.Sprintf("request to: %s", rawURL),
			Allowed:   []string{"valid URLs only"},
			ConfigKey: "permissions.http",
		}
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// If no HTTP scope is set, all URLs are allowed (default permissive)
	if p.httpScope == nil {
		return nil
	}

	host := parsed.Hostname()

	// Check deny list first
	for _, pattern := range p.httpScope.Deny {
		if matchDomain(host, pattern) {
			return &PermissionError{
				Namespace: "http",
				Method:    "fetch",
				Attempted: fmt.Sprintf("request to: %s", rawURL),
				Allowed:   p.httpScope.Allow,
				ConfigKey: "permissions.http.deny",
			}
		}
	}

	// If allow list is empty, everything not denied is allowed
	if len(p.httpScope.Allow) == 0 {
		return nil
	}

	// Check allow list
	for _, pattern := range p.httpScope.Allow {
		if matchDomain(host, pattern) {
			return nil
		}
	}

	return &PermissionError{
		Namespace: "http",
		Method:    "fetch",
		Attempted: fmt.Sprintf("request to: %s", rawURL),
		Allowed:   p.httpScope.Allow,
		ConfigKey: "permissions.http.allow",
	}
}

// CheckProcess verifies that a command execution is allowed.
func (p *Policy) CheckProcess(cmd string, args []string) error {
	if p.devMode {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// If no process scope is set, deny all (secure by default for process execution)
	if p.processScope == nil {
		return &PermissionError{
			Namespace: "process",
			Method:    "exec",
			Attempted: fmt.Sprintf("execute: %s %s", cmd, strings.Join(args, " ")),
			Allowed:   []string{"no commands allowed"},
			ConfigKey: "permissions.process.exec",
		}
	}

	for _, rule := range p.processScope.Exec {
		if rule.Cmd != cmd {
			continue
		}

		// If no args restriction or wildcard, allow any args
		if len(rule.Args) == 0 || (len(rule.Args) == 1 && rule.Args[0] == "*") {
			return nil
		}

		// Check that ALL user-provided args are in the allowed list
		allowedSet := make(map[string]bool, len(rule.Args))
		for _, a := range rule.Args {
			allowedSet[a] = true
		}
		allAllowed := true
		for _, arg := range args {
			if !allowedSet[arg] {
				allAllowed = false
				break
			}
		}
		if allAllowed && len(args) > 0 {
			return nil
		}

		// Args not allowed
		allowedStr := make([]string, len(rule.Args))
		for i, a := range rule.Args {
			allowedStr[i] = cmd + " " + a
		}
		return &PermissionError{
			Namespace: "process",
			Method:    "exec",
			Attempted: fmt.Sprintf("execute: %s %s", cmd, strings.Join(args, " ")),
			Allowed:   allowedStr,
			ConfigKey: "permissions.process.exec",
		}
	}

	// Command not in allowed list at all
	allowedCmds := make([]string, len(p.processScope.Exec))
	for i, rule := range p.processScope.Exec {
		allowedCmds[i] = rule.Cmd
	}
	return &PermissionError{
		Namespace: "process",
		Method:    "exec",
		Attempted: fmt.Sprintf("execute: %s", cmd),
		Allowed:   allowedCmds,
		ConfigKey: "permissions.process.exec",
	}
}

// AllowDir adds an additional allowed directory (e.g., user-selected via dialog).
func (p *Policy) AllowDir(dir string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	p.allowedDirs = append(p.allowedDirs, absDir)
}

// HasPermission checks if a permission is granted without returning an error.
func (p *Policy) HasPermission(perm Permission) bool {
	if p.devMode {
		return true
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.permissions[perm]
}

// --- PermissionError type ---

// PermissionError is a structured permission denial error with AI-debuggable output.
type PermissionError struct {
	Namespace string
	Method    string
	Attempted string
	Allowed   []string
	ConfigKey string
}

// Error returns the AI-friendly formatted error message.
func (e *PermissionError) Error() string {
	methodStr := e.Namespace
	if e.Method != "" {
		methodStr = e.Namespace + "." + e.Method
	}

	lines := []string{
		fmt.Sprintf("LightShell Error [%s]: Permission denied", methodStr),
	}
	if e.Attempted != "" {
		lines = append(lines, fmt.Sprintf("  -> %s", e.Attempted))
	}
	if len(e.Allowed) > 0 {
		lines = append(lines, fmt.Sprintf("  -> Allowed: %s", strings.Join(e.Allowed, ", ")))
	}
	if e.ConfigKey != "" {
		lines = append(lines, fmt.Sprintf("  -> To allow this, update %s in lightshell.json", e.ConfigKey))
	}
	lines = append(lines, "  -> Docs: https://lightshell.dev/guides/security-and-permissions")

	return strings.Join(lines, "\n")
}

// --- Internal helpers ---

// checkPathAgainstDirs checks a resolved path against the allowed directories list.
// Must be called with p.mu held.
func (p *Policy) checkPathAgainstDirs(resolved, namespace, method, originalPath string) error {
	for _, dir := range p.allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(resolved, absDir+string(os.PathSeparator)) || resolved == absDir {
			return nil
		}
	}

	return &PermissionError{
		Namespace: namespace,
		Method:    method,
		Attempted: fmt.Sprintf("access: %s", originalPath),
		Allowed:   p.allowedDirs,
		ConfigKey: "permissions.fs",
	}
}

// declaredPermissions returns the list of currently declared permission names.
func (p *Policy) declaredPermissions() []string {
	var perms []string
	for perm, granted := range p.permissions {
		if granted {
			perms = append(perms, string(perm))
		}
	}
	return perms
}

// resolveRealPath resolves a path through symlinks, falling back to the
// original absolute path if symlink resolution fails (e.g., file doesn't exist yet).
func resolveRealPath(absPath string) string {
	resolved, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		return absPath
	}
	return filepath.Join(resolved, filepath.Base(absPath))
}

// resolvePathVariable expands path variables like $APP_DATA, $HOME, $TEMP, $DOWNLOADS, $DESKTOP.
func resolvePathVariable(pattern string, appName string) string {
	home, _ := os.UserHomeDir()

	replacements := map[string]string{
		"$HOME":      home,
		"$TEMP":      os.TempDir(),
		"$DOWNLOADS": filepath.Join(home, "Downloads"),
		"$DESKTOP":   filepath.Join(home, "Desktop"),
	}

	// Platform-specific app data directory
	switch runtime.GOOS {
	case "darwin":
		replacements["$APP_DATA"] = filepath.Join(home, "Library", "Application Support", appName)
	default:
		replacements["$APP_DATA"] = filepath.Join(home, ".local", "share", appName)
	}

	result := pattern
	for variable, value := range replacements {
		result = strings.ReplaceAll(result, variable, value)
	}

	return result
}

// matchGlob matches a real path against a glob pattern.
// Supports * (matches within directory) and ** (matches recursively).
func matchGlob(realPath, pattern string) bool {
	// Handle ** patterns by splitting
	if strings.Contains(pattern, "**") {
		return matchDoubleStarGlob(realPath, pattern)
	}

	// Simple glob match using filepath.Match
	matched, err := filepath.Match(pattern, realPath)
	if err != nil {
		return false
	}
	return matched
}

// matchDoubleStarGlob handles ** patterns which match any number of path segments.
func matchDoubleStarGlob(realPath, pattern string) bool {
	// Split pattern on **
	parts := strings.SplitN(pattern, "**", 2)
	if len(parts) != 2 {
		return false
	}

	prefix := filepath.Clean(parts[0])
	suffix := parts[1]

	// The path must start with the prefix
	if !strings.HasPrefix(realPath, prefix) {
		return false
	}

	// If suffix is empty or just "/", any path under prefix matches
	if suffix == "" || suffix == "/" || suffix == string(os.PathSeparator) {
		return true
	}

	// Check if the remaining path matches the suffix pattern
	remaining := realPath[len(prefix):]
	if strings.HasPrefix(remaining, string(os.PathSeparator)) {
		remaining = remaining[1:]
	}

	// Try matching the suffix against the tail of the remaining path
	if strings.HasPrefix(suffix, "/") || strings.HasPrefix(suffix, string(os.PathSeparator)) {
		suffix = suffix[1:]
	}

	// If the suffix contains wildcards, match each segment
	if strings.ContainsAny(suffix, "*?[") {
		matched, _ := filepath.Match(suffix, filepath.Base(realPath))
		return matched
	}

	// Literal suffix match
	return strings.HasSuffix(realPath, suffix)
}

// matchDomain matches a hostname against a domain pattern.
// Supports wildcards: "*.example.com" matches "api.example.com" and "sub.api.example.com".
// Exact match: "example.com" matches only "example.com".
func matchDomain(host, pattern string) bool {
	host = strings.ToLower(host)
	pattern = strings.ToLower(pattern)

	if host == pattern {
		return true
	}

	// Wildcard prefix: *.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".example.com"
		return strings.HasSuffix(host, suffix)
	}

	return false
}
