package compat

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)
	for name, content := range files {
		path := filepath.Join(srcDir, name)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func findIssueByRuleID(issues []Issue, ruleID string) *Issue {
	for i, issue := range issues {
		if issue.Rule.ID == ruleID {
			return &issues[i]
		}
	}
	return nil
}

func TestScannerFindsBackdropFilter(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"style.css": ".overlay { backdrop-filter: blur(10px); }",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	if findIssueByRuleID(issues, "CSS-001") == nil {
		t.Fatal("expected CSS-001 warning")
	}
}

func TestScannerFindsStructuredClone(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"app.js": "const x = structuredClone(data);",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	issue := findIssueByRuleID(issues, "JS-001")
	if issue == nil {
		t.Fatal("expected JS-001 warning")
	}
	if !issue.AutoFix {
		t.Error("expected JS-001 to have AutoFix=true")
	}
}

func TestScannerCleanFile(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"app.js":    "console.log('hello');",
		"style.css": "body { margin: 0; }",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestScannerEmptyProject(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestScannerNavigationAPI(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"app.js": "navigation.navigate('/home');",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	issue := findIssueByRuleID(issues, "JS-003")
	if issue == nil {
		t.Fatal("expected JS-003")
	}
	if issue.Severity != "error" {
		t.Errorf("expected severity 'error', got %q", issue.Severity)
	}
}

func TestScannerMultipleIssues(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"app.js": "structuredClone(x);\nshowOpenFilePicker();\nrequire('fs');",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	ruleIDs := map[string]bool{}
	for _, issue := range issues {
		ruleIDs[issue.Rule.ID] = true
	}
	for _, id := range []string{"JS-001", "JS-004", "JS-008"} {
		if !ruleIDs[id] {
			t.Errorf("expected rule %s to fire", id)
		}
	}
}

func TestAllRulesHavePatterns(t *testing.T) {
	for _, rule := range Rules {
		if len(rule.Patterns) == 0 {
			t.Errorf("rule %s has no patterns", rule.ID)
		}
		if rule.ID == "" {
			t.Error("rule has empty ID")
		}
		if rule.Severity == "" {
			t.Error("rule has empty Severity")
		}
		if rule.Title == "" {
			t.Error("rule has empty Title")
		}
		if len(rule.FileTypes) == 0 {
			t.Error("rule has no FileTypes")
		}
	}
}

func TestScannerRuleIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, rule := range Rules {
		if seen[rule.ID] {
			t.Errorf("duplicate rule ID: %s", rule.ID)
		}
		seen[rule.ID] = true
	}
}

func TestScannerRulesCount(t *testing.T) {
	if len(Rules) < 15 {
		t.Errorf("expected at least 15 rules, got %d", len(Rules))
	}
}

func TestMatchesFileType(t *testing.T) {
	tests := []struct {
		fileTypes []string
		ext       string
		want      bool
	}{
		{[]string{"js"}, "js", true},
		{[]string{"css"}, "css", true},
		{[]string{"html"}, "html", true},
		{[]string{"css"}, "js", false},
		{[]string{"js"}, "ts", false},
		{[]string{"js", "html"}, "js", true},
	}

	for _, tt := range tests {
		got := matchesFileType(tt.fileTypes, tt.ext)
		if got != tt.want {
			t.Errorf("matchesFileType(%v, %q) = %v, want %v", tt.fileTypes, tt.ext, got, tt.want)
		}
	}
}

func TestScannerHTMLWithJS(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"index.html": "<script>navigation.navigate('/page');</script>",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	if findIssueByRuleID(issues, "JS-003") == nil {
		t.Fatal("expected JS-003 in HTML")
	}
}

func TestScannerHTMLWithCSS(t *testing.T) {
	dir := createTestProject(t, map[string]string{
		"index.html": "<style>.blur { backdrop-filter: blur(5px); }</style>",
	})
	issues, err := ScanProject(dir)
	if err != nil {
		t.Fatalf("ScanProject failed: %v", err)
	}
	if findIssueByRuleID(issues, "CSS-001") == nil {
		t.Fatal("expected CSS-001 in HTML")
	}
}

func TestScannerNodeAPIs(t *testing.T) {
	tests := []struct {
		code   string
		ruleID string
	}{
		{"require('fs');", "JS-008"},
		{"process.env.API_KEY;", "JS-009"},
	}
	for _, tt := range tests {
		dir := createTestProject(t, map[string]string{"app.js": tt.code})
		issues, err := ScanProject(dir)
		if err != nil {
			t.Fatalf("ScanProject failed: %v", err)
		}
		if findIssueByRuleID(issues, tt.ruleID) == nil {
			t.Errorf("expected %s for %q", tt.ruleID, tt.code)
		}
	}
}

func TestScannerWebAPIs(t *testing.T) {
	tests := []struct {
		code   string
		ruleID string
	}{
		{"navigator.usb.requestDevice({});", "JS-005"},
		{"navigator.bluetooth.requestDevice({});", "JS-006"},
		{"navigator.serial.requestPort();", "JS-007"},
		{"showOpenFilePicker();", "JS-004"},
	}
	for _, tt := range tests {
		dir := createTestProject(t, map[string]string{"app.js": tt.code})
		issues, err := ScanProject(dir)
		if err != nil {
			t.Fatalf("ScanProject failed: %v", err)
		}
		if findIssueByRuleID(issues, tt.ruleID) == nil {
			t.Errorf("expected %s for %q", tt.ruleID, tt.code)
		}
	}
}
