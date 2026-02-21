package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lightshell-dev/lightshell/internal/compat"
)

// createTestProject creates a temporary project directory with a src/ subdirectory
// and writes the given files into it. Returns the project root path.
func createTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(srcDir, name)
		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create parent dir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// findIssueByRuleID returns the first issue matching the given rule ID, or nil.
func findIssueByRuleID(issues []compat.Issue, ruleID string) *compat.Issue {
	for i, issue := range issues {
		if issue.Rule.ID == ruleID {
			return &issues[i]
		}
	}
	return nil
}

func TestScannerFindsBackdropFilter(t *testing.T) {
	projectDir := createTestProject(t, map[string]string{
		"style.css": `.overlay {
  backdrop-filter: blur(10px);
  background: rgba(0, 0, 0, 0.5);
}`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	issue := findIssueByRuleID(issues, "CSS-001")
	if issue == nil {
		t.Fatal("expected to find CSS-001 (backdrop-filter) warning, but found none")
	}
	if issue.Severity != "warning" {
		t.Errorf("expected severity %q, got %q", "warning", issue.Severity)
	}
	if issue.Line != 2 {
		t.Errorf("expected issue on line 2, got line %d", issue.Line)
	}
	if !issue.AutoFix {
		t.Error("expected CSS-001 to have AutoFix=true")
	}
}

func TestScannerFindsStructuredClone(t *testing.T) {
	projectDir := createTestProject(t, map[string]string{
		"app.js": `function copyData(data) {
  return structuredClone(data);
}`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	issue := findIssueByRuleID(issues, "JS-001")
	if issue == nil {
		t.Fatal("expected to find JS-001 (structuredClone) warning, but found none")
	}
	if issue.Severity != "warning" {
		t.Errorf("expected severity %q, got %q", "warning", issue.Severity)
	}
	if issue.Line != 2 {
		t.Errorf("expected issue on line 2, got line %d", issue.Line)
	}
	if !issue.AutoFix {
		t.Error("expected JS-001 to have AutoFix=true")
	}
}

func TestScannerFindsNavigationAPI(t *testing.T) {
	projectDir := createTestProject(t, map[string]string{
		"app.js": `// Set up routing
navigation.navigate('/home');
navigation.addEventListener('navigate', (e) => {
  console.log(e);
});`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	issue := findIssueByRuleID(issues, "JS-003")
	if issue == nil {
		t.Fatal("expected to find JS-003 (Navigation API) error, but found none")
	}
	if issue.Severity != "error" {
		t.Errorf("expected severity %q, got %q", "error", issue.Severity)
	}
	if issue.AutoFix {
		t.Error("expected JS-003 to have AutoFix=false (Navigation API is not auto-fixable)")
	}
}

func TestScannerCleanFile(t *testing.T) {
	projectDir := createTestProject(t, map[string]string{
		"app.js": `// A clean LightShell app
async function main() {
  const content = await lightshell.fs.readFile('./data.json');
  console.log('Loaded:', content);
  await lightshell.window.setTitle('My App');
}
main();`,
		"style.css": `body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  margin: 0;
  padding: 20px;
  background: #fff;
  color: #333;
}
h1 { font-size: 24px; }`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for clean files, got %d:", len(issues))
		for _, issue := range issues {
			t.Logf("  - %s (line %d in %s): %s", issue.Rule.ID, issue.Line, issue.File, issue.Title)
		}
	}
}

func TestAllRulesHavePatterns(t *testing.T) {
	for _, rule := range compat.Rules {
		t.Run(rule.ID, func(t *testing.T) {
			if len(rule.Patterns) == 0 {
				t.Errorf("rule %s (%s) has no patterns", rule.ID, rule.Title)
			}
			if rule.ID == "" {
				t.Error("rule has empty ID")
			}
			if rule.Severity == "" {
				t.Error("rule has empty Severity")
			}
			if rule.Severity != "error" && rule.Severity != "warning" && rule.Severity != "info" {
				t.Errorf("rule has invalid severity %q (expected error, warning, or info)", rule.Severity)
			}
			if rule.Title == "" {
				t.Error("rule has empty Title")
			}
			if len(rule.FileTypes) == 0 {
				t.Error("rule has no FileTypes")
			}
			if rule.Fix == "" {
				t.Error("rule has empty Fix")
			}
		})
	}
}

func TestScannerMultipleIssuesSameFile(t *testing.T) {
	projectDir := createTestProject(t, map[string]string{
		"app.js": `const clone = structuredClone(data);
const picker = showOpenFilePicker();
require('fs');`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	// Should find at least 3 different rules triggered
	ruleIDs := make(map[string]bool)
	for _, issue := range issues {
		ruleIDs[issue.Rule.ID] = true
	}

	expectedRules := []string{"JS-001", "JS-004", "JS-008"}
	for _, id := range expectedRules {
		if !ruleIDs[id] {
			t.Errorf("expected rule %s to be triggered, but it was not. Found rules: %v", id, ruleIDs)
		}
	}
}

func TestScannerHTMLWithInlineJS(t *testing.T) {
	// HTML files should be scanned for JS issues too (FileTypes includes "html" for JS rules)
	projectDir := createTestProject(t, map[string]string{
		"index.html": `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<script>
  navigation.navigate('/page');
</script>
</body>
</html>`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	issue := findIssueByRuleID(issues, "JS-003")
	if issue == nil {
		t.Fatal("expected to find JS-003 (Navigation API) in HTML file, but found none")
	}
}

func TestScannerCSSIssuesInHTMLStyle(t *testing.T) {
	// HTML files should also be scanned for CSS issues (CSS rules include "html" FileType)
	projectDir := createTestProject(t, map[string]string{
		"index.html": `<!DOCTYPE html>
<html>
<head>
<style>
  .blur { backdrop-filter: blur(5px); }
</style>
</head>
<body></body>
</html>`,
	})

	issues, err := compat.ScanProject(projectDir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	issue := findIssueByRuleID(issues, "CSS-001")
	if issue == nil {
		t.Fatal("expected to find CSS-001 (backdrop-filter) in HTML file, but found none")
	}
}

func TestScannerNodeAPIs(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		ruleID   string
		severity string
	}{
		{
			name:     "require statement",
			code:     `const fs = require('fs');`,
			ruleID:   "JS-008",
			severity: "error",
		},
		{
			name:     "process.env",
			code:     `const key = process.env.API_KEY;`,
			ruleID:   "JS-009",
			severity: "error",
		},
		{
			name:     "process.cwd",
			code:     `const dir = process.cwd();`,
			ruleID:   "JS-009",
			severity: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := createTestProject(t, map[string]string{
				"app.js": tt.code,
			})

			issues, err := compat.ScanProject(projectDir)
			if err != nil {
				t.Fatalf("ScanProject failed: %v", err)
			}

			issue := findIssueByRuleID(issues, tt.ruleID)
			if issue == nil {
				t.Fatalf("expected to find rule %s, but found none", tt.ruleID)
			}
			if issue.Severity != tt.severity {
				t.Errorf("expected severity %q, got %q", tt.severity, issue.Severity)
			}
		})
	}
}

func TestScannerWebAPIs(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		ruleID string
	}{
		{
			name:   "Web USB",
			code:   `navigator.usb.requestDevice({});`,
			ruleID: "JS-005",
		},
		{
			name:   "Web Bluetooth",
			code:   `navigator.bluetooth.requestDevice({});`,
			ruleID: "JS-006",
		},
		{
			name:   "Web Serial",
			code:   `navigator.serial.requestPort();`,
			ruleID: "JS-007",
		},
		{
			name:   "showOpenFilePicker",
			code:   `const handle = await showOpenFilePicker();`,
			ruleID: "JS-004",
		},
		{
			name:   "showSaveFilePicker",
			code:   `const handle = await showSaveFilePicker();`,
			ruleID: "JS-004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := createTestProject(t, map[string]string{
				"app.js": tt.code,
			})

			issues, err := compat.ScanProject(projectDir)
			if err != nil {
				t.Fatalf("ScanProject failed: %v", err)
			}

			issue := findIssueByRuleID(issues, tt.ruleID)
			if issue == nil {
				t.Fatalf("expected to find rule %s for %s, but found none", tt.ruleID, tt.name)
			}
			if issue.Severity != "error" {
				t.Errorf("expected severity %q, got %q", "error", issue.Severity)
			}
		})
	}
}

func TestScannerEmptyProject(t *testing.T) {
	// A project with an empty src/ directory should return no issues
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	issues, err := compat.ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues for empty project, got %d", len(issues))
	}
}

func TestScannerRulesCount(t *testing.T) {
	// The spec requires at least 15 rules
	if len(compat.Rules) < 15 {
		t.Errorf("expected at least 15 compat rules, got %d", len(compat.Rules))
	}
}

func TestScannerRuleIDsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, rule := range compat.Rules {
		if seen[rule.ID] {
			t.Errorf("duplicate rule ID: %s", rule.ID)
		}
		seen[rule.ID] = true
	}
}
