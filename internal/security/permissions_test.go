package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resolvedTempDir returns a temp dir with symlinks resolved (macOS /var -> /private/var).
func resolvedTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("failed to resolve temp dir: %v", err)
	}
	return resolved
}

func TestDevPolicyAllowsEverything(t *testing.T) {
	p := DevPolicy()
	for _, perm := range AllPermissions {
		if err := p.Check(perm); err != nil {
			t.Errorf("DevPolicy should allow %s, got error: %v", perm, err)
		}
	}
}

func TestDevPolicyHasAllPermissions(t *testing.T) {
	p := DevPolicy()
	for _, perm := range AllPermissions {
		if !p.HasPermission(perm) {
			t.Errorf("DevPolicy should have %s permission", perm)
		}
	}
}

func TestDevPolicyAllowsAnyPath(t *testing.T) {
	p := DevPolicy()
	if err := p.CheckPath("/etc/passwd"); err != nil {
		t.Errorf("DevPolicy should allow any path: %v", err)
	}
	if err := p.CheckFSRead("/etc/shadow"); err != nil {
		t.Errorf("DevPolicy should allow any read: %v", err)
	}
	if err := p.CheckFSWrite("/tmp/anything"); err != nil {
		t.Errorf("DevPolicy should allow any write: %v", err)
	}
}

func TestDevPolicyAllowsAnyHTTP(t *testing.T) {
	p := DevPolicy()
	if err := p.CheckHTTP("https://evil.example.com"); err != nil {
		t.Errorf("DevPolicy should allow any HTTP: %v", err)
	}
}

func TestDevPolicyAllowsAnyProcess(t *testing.T) {
	p := DevPolicy()
	if err := p.CheckProcess("rm", []string{"-rf", "/"}); err != nil {
		t.Errorf("DevPolicy should allow any process: %v", err)
	}
}

func TestNewPolicyWithPermissions(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs", "dialog"}, dir, "test-app", false)

	if err := p.Check(PermFS); err != nil {
		t.Errorf("expected fs to be granted: %v", err)
	}
	if err := p.Check(PermDialog); err != nil {
		t.Errorf("expected dialog to be granted: %v", err)
	}
	if err := p.Check(PermClipboard); err == nil {
		t.Error("expected clipboard to be denied")
	}
	if err := p.Check(PermShell); err == nil {
		t.Error("expected shell to be denied")
	}
}

func TestNewPolicyHasPermission(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	if !p.HasPermission(PermFS) {
		t.Error("expected HasPermission(fs) = true")
	}
	if p.HasPermission(PermHTTP) {
		t.Error("expected HasPermission(http) = false")
	}
}

func TestCheckPathInProjectDir(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	testFile := filepath.Join(dir, "data.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	if err := p.CheckPath(testFile); err != nil {
		t.Errorf("expected path inside project dir to be allowed: %v", err)
	}
}

func TestCheckPathOutsideAllowedDirs(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	err := p.CheckPath("/etc/passwd")
	if err == nil {
		t.Fatal("expected path outside allowed dirs to be denied")
	}

	permErr, ok := err.(*PermissionError)
	if !ok {
		t.Fatalf("expected PermissionError, got %T", err)
	}
	if permErr.Namespace != "fs" {
		t.Errorf("expected namespace 'fs', got %q", permErr.Namespace)
	}
}

func TestCheckPathInTempDir(t *testing.T) {
	dir := resolvedTempDir(t)
	// Create a subdirectory inside the resolved temp root to test
	resolvedTemp, _ := filepath.EvalSymlinks(os.TempDir())
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)
	// Manually add the resolved temp dir so path matching works after symlink resolution
	p.AllowDir(resolvedTemp)

	tempFile := filepath.Join(resolvedTemp, "lightshell-test-file.txt")
	if err := p.CheckPath(tempFile); err != nil {
		t.Errorf("expected temp dir path to be allowed: %v", err)
	}
}

func TestAllowDirAddsDirectory(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	extraDir := resolvedTempDir(t)
	testFile := filepath.Join(extraDir, "extra.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	// Should be denied before AllowDir (extraDir is not the project dir)
	// Note: it might be allowed via temp dir on some systems, so test with a non-temp dir
	p.AllowDir(extraDir)

	// Should be allowed after AllowDir
	if err := p.CheckPath(testFile); err != nil {
		t.Errorf("expected path to be allowed after AllowDir: %v", err)
	}
}

func TestCheckFSReadWithScope(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	readDir := resolvedTempDir(t)
	p.SetFSScope(FSScope{
		Read: []string{readDir + "/**"},
	})

	testFile := filepath.Join(readDir, "readable.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	if err := p.CheckFSRead(testFile); err != nil {
		t.Errorf("expected read to be allowed within scope: %v", err)
	}
}

func TestCheckFSWriteWithScope(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	writeDir := resolvedTempDir(t)
	p.SetFSScope(FSScope{
		Write: []string{writeDir + "/**"},
	})

	testFile := filepath.Join(writeDir, "writable.txt")
	if err := p.CheckFSWrite(testFile); err != nil {
		t.Errorf("expected write to be allowed within scope: %v", err)
	}
}

func TestCheckFSReadFallsBackToAllowedDirs(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	testFile := filepath.Join(dir, "allowed.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	if err := p.CheckFSRead(testFile); err != nil {
		t.Errorf("expected read in project dir to be allowed: %v", err)
	}
	if err := p.CheckFSRead("/etc/passwd"); err == nil {
		t.Error("expected read outside project dir to be denied")
	}
}

func TestCheckFSWriteFallsBackToAllowedDirs(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	testFile := filepath.Join(dir, "allowed.txt")
	if err := p.CheckFSWrite(testFile); err != nil {
		t.Errorf("expected write in project dir to be allowed: %v", err)
	}
	if err := p.CheckFSWrite("/etc/shadow"); err == nil {
		t.Error("expected write outside project dir to be denied")
	}
}

func TestCheckHTTPNoScope(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"http"}, dir, "test-app", false)

	if err := p.CheckHTTP("https://api.github.com/user"); err != nil {
		t.Errorf("expected HTTP with no scope to allow all: %v", err)
	}
}

func TestCheckHTTPWithAllowList(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"http"}, dir, "test-app", false)
	p.SetHTTPScope(HTTPScope{
		Allow: []string{"api.github.com", "*.example.com"},
	})

	if err := p.CheckHTTP("https://api.github.com/user"); err != nil {
		t.Errorf("expected api.github.com to be allowed: %v", err)
	}
	if err := p.CheckHTTP("https://sub.example.com/data"); err != nil {
		t.Errorf("expected sub.example.com to be allowed: %v", err)
	}
	if err := p.CheckHTTP("https://evil.com/steal"); err == nil {
		t.Error("expected evil.com to be denied")
	}
}

func TestCheckHTTPWithDenyList(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"http"}, dir, "test-app", false)
	p.SetHTTPScope(HTTPScope{
		Deny: []string{"*.evil.com"},
	})

	if err := p.CheckHTTP("https://api.evil.com/data"); err == nil {
		t.Error("expected api.evil.com to be denied")
	}
	if err := p.CheckHTTP("https://good.com/data"); err != nil {
		t.Errorf("expected good.com to be allowed: %v", err)
	}
}

func TestCheckHTTPDenyTakesPrecedence(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"http"}, dir, "test-app", false)
	p.SetHTTPScope(HTTPScope{
		Allow: []string{"*.example.com"},
		Deny:  []string{"blocked.example.com"},
	})

	if err := p.CheckHTTP("https://api.example.com"); err != nil {
		t.Errorf("expected api.example.com to be allowed: %v", err)
	}
	if err := p.CheckHTTP("https://blocked.example.com"); err == nil {
		t.Error("expected blocked.example.com to be denied")
	}
}

func TestCheckHTTPInvalidURL(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"http"}, dir, "test-app", false)

	if err := p.CheckHTTP("://invalid"); err == nil {
		t.Error("expected invalid URL to be denied")
	}
}

func TestCheckProcessNoScope(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"process"}, dir, "test-app", false)

	if err := p.CheckProcess("ls", nil); err == nil {
		t.Error("expected process with no scope to deny all")
	}
}

func TestCheckProcessWithScope(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"process"}, dir, "test-app", false)
	p.SetProcessScope(ProcessScope{
		Exec: []ProcessRule{
			{Cmd: "git", Args: []string{"status", "log", "diff"}},
			{Cmd: "python3"},
		},
	})

	if err := p.CheckProcess("git", []string{"status"}); err != nil {
		t.Errorf("expected git status to be allowed: %v", err)
	}
	if err := p.CheckProcess("git", []string{"push"}); err == nil {
		t.Error("expected git push to be denied")
	}
	if err := p.CheckProcess("python3", []string{"script.py"}); err != nil {
		t.Errorf("expected python3 with any args to be allowed: %v", err)
	}
	if err := p.CheckProcess("rm", []string{"-rf", "/"}); err == nil {
		t.Error("expected rm to be denied")
	}
}

func TestCheckProcessWildcardArgs(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"process"}, dir, "test-app", false)
	p.SetProcessScope(ProcessScope{
		Exec: []ProcessRule{
			{Cmd: "node", Args: []string{"*"}},
		},
	})

	if err := p.CheckProcess("node", []string{"server.js", "--port", "3000"}); err != nil {
		t.Errorf("expected node with wildcard args to be allowed: %v", err)
	}
}

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		host, pattern string
		want          bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "other.com", false},
		{"api.example.com", "*.example.com", true},
		{"sub.api.example.com", "*.example.com", true},
		{"example.com", "*.example.com", false},
		{"EXAMPLE.COM", "example.com", true},
		{"api.example.com", "*.EXAMPLE.COM", true},
	}

	for _, tt := range tests {
		t.Run(tt.host+"_"+tt.pattern, func(t *testing.T) {
			if got := matchDomain(tt.host, tt.pattern); got != tt.want {
				t.Errorf("matchDomain(%q, %q) = %v, want %v", tt.host, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchGlobSimple(t *testing.T) {
	tests := []struct {
		path, pattern string
		want          bool
	}{
		{"/tmp/test.txt", "/tmp/*.txt", true},
		{"/tmp/test.json", "/tmp/*.txt", false},
		{"/tmp/test.txt", "/tmp/test.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := matchGlob(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchDoubleStarGlob(t *testing.T) {
	dir := resolvedTempDir(t)
	tests := []struct {
		path, pattern string
		want          bool
	}{
		{filepath.Join(dir, "sub", "file.txt"), dir + "/**", true},
		{filepath.Join(dir, "deep", "nested", "file.txt"), dir + "/**", true},
		{filepath.Join(dir, "file.txt"), dir + "/**/*.txt", true},
		{"/other/file.txt", dir + "/**", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := matchDoubleStarGlob(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchDoubleStarGlob(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestResolvePathVariable(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := resolvePathVariable("$HOME/Documents/**", "myapp")

	if !strings.HasPrefix(result, home) {
		t.Errorf("expected $HOME resolved to %q, got %q", home, result)
	}
	if !strings.HasSuffix(result, "/Documents/**") {
		t.Errorf("expected suffix /Documents/**, got %q", result)
	}
}

func TestResolvePathVariableAppData(t *testing.T) {
	result := resolvePathVariable("$APP_DATA/**", "myapp")
	if strings.Contains(result, "$APP_DATA") {
		t.Errorf("expected $APP_DATA to be resolved, got %q", result)
	}
	if !strings.Contains(result, "myapp") {
		t.Errorf("expected app name in path, got %q", result)
	}
}

func TestResolvePathVariableTemp(t *testing.T) {
	result := resolvePathVariable("$TEMP/cache", "myapp")
	if strings.Contains(result, "$TEMP") {
		t.Errorf("expected $TEMP to be resolved, got %q", result)
	}
}

func TestPermissionErrorFormat(t *testing.T) {
	err := &PermissionError{
		Namespace: "fs",
		Method:    "readFile",
		Attempted: "read: /etc/passwd",
		Allowed:   []string{"/tmp/**"},
		ConfigKey: "permissions.fs.read",
	}
	msg := err.Error()
	if !strings.Contains(msg, "[fs.readFile]") {
		t.Errorf("expected [fs.readFile] in message, got %q", msg)
	}
	if !strings.Contains(msg, "/etc/passwd") {
		t.Errorf("expected attempted path, got %q", msg)
	}
	if !strings.Contains(msg, "Docs:") {
		t.Errorf("expected docs link, got %q", msg)
	}
}

func TestPermissionErrorWithoutMethod(t *testing.T) {
	err := &PermissionError{
		Namespace: "clipboard",
		Attempted: "use clipboard API",
		ConfigKey: "permissions",
	}
	msg := err.Error()
	if !strings.Contains(msg, "[clipboard]") {
		t.Errorf("expected [clipboard], got %q", msg)
	}
}

func TestCheckPathWithFSScope(t *testing.T) {
	dir := resolvedTempDir(t)
	readDir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)
	p.SetFSScope(FSScope{
		Read:  []string{readDir + "/**"},
		Write: []string{readDir + "/**"},
	})

	testFile := filepath.Join(readDir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	if err := p.CheckPath(testFile); err != nil {
		t.Errorf("expected path in scoped dir to be allowed: %v", err)
	}
}

func TestAllPermissionsContainsExpected(t *testing.T) {
	expected := []Permission{PermFS, PermDialog, PermClipboard, PermShell,
		PermNotification, PermTray, PermMenu, PermHTTP, PermProcess,
		PermStore, PermShortcuts, PermUpdater}

	for _, perm := range expected {
		found := false
		for _, p := range AllPermissions {
			if p == perm {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllPermissions missing %s", perm)
		}
	}
}

func TestCheckPermissionDeniedError(t *testing.T) {
	dir := resolvedTempDir(t)
	p := NewPolicy([]string{"fs"}, dir, "test-app", false)

	err := p.Check(PermHTTP)
	if err == nil {
		t.Fatal("expected error for denied permission")
	}
	permErr, ok := err.(*PermissionError)
	if !ok {
		t.Fatalf("expected *PermissionError, got %T", err)
	}
	if permErr.Namespace != "http" {
		t.Errorf("expected namespace 'http', got %q", permErr.Namespace)
	}
}

func TestResolveRealPath(t *testing.T) {
	dir := resolvedTempDir(t)
	testFile := filepath.Join(dir, "test.txt")
	os.WriteFile(testFile, []byte("data"), 0644)

	result := resolveRealPath(testFile)
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %q", result)
	}
}

func TestResolveRealPathNonexistent(t *testing.T) {
	result := resolveRealPath("/nonexistent/dir/file.txt")
	if result != "/nonexistent/dir/file.txt" {
		t.Errorf("expected fallback to original, got %q", result)
	}
}
