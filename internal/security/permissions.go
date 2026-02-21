package security

import (
	"fmt"
	"os"
	"path/filepath"
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
	// window, system, and app are always allowed — they're core APIs
)

// AllPermissions is the set of all declarable permissions.
var AllPermissions = []Permission{
	PermFS, PermDialog, PermClipboard, PermShell,
	PermNotification, PermTray, PermMenu,
}

// Policy holds the security policy for a running app.
type Policy struct {
	mu          sync.RWMutex
	permissions map[Permission]bool
	allowedDirs []string // directories the app can access via fs APIs
	devMode     bool     // dev mode disables restrictions
}

// NewPolicy creates a security policy from the declared permissions.
func NewPolicy(perms []string, projectDir string, appName string, devMode bool) *Policy {
	p := &Policy{
		permissions: make(map[Permission]bool),
		devMode:     devMode,
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

// Check returns an error if the given permission is not granted.
func (p *Policy) Check(perm Permission) error {
	if p.devMode {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.permissions[perm] {
		return fmt.Errorf("permission denied: %q not declared in lightshell.json permissions. Add %q to the \"permissions\" array to enable this API", string(perm), string(perm))
	}
	return nil
}

// CheckPath verifies that a file path is within the allowed directories.
func (p *Policy) CheckPath(path string) error {
	if p.devMode {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
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

	for _, dir := range p.allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(checkPath, absDir+string(os.PathSeparator)) || checkPath == absDir {
			return nil
		}
	}

	return fmt.Errorf("access denied: path %q is outside allowed directories. Apps can only access project files, app data directory, and temp directory", path)
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
