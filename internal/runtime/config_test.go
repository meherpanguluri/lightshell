package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	config := `{"name": "test-app", "version": "1.0.0"}`
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte(config), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-app" {
		t.Errorf("expected name 'test-app', got %q", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", cfg.Version)
	}
	if cfg.Window.Width != 1024 {
		t.Errorf("expected default width 1024, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 768 {
		t.Errorf("expected default height 768, got %d", cfg.Window.Height)
	}
	if cfg.Window.Title != "test-app" {
		t.Errorf("expected title to default to name, got %q", cfg.Window.Title)
	}
	if cfg.Window.Resizable == nil || !*cfg.Window.Resizable {
		t.Error("expected resizable to default to true")
	}
	if cfg.Entry != "src/index.html" {
		t.Errorf("expected default entry 'src/index.html', got %q", cfg.Entry)
	}
}

func TestLoadConfigAllFields(t *testing.T) {
	dir := t.TempDir()
	config := `{
		"name": "myapp",
		"version": "2.0.0",
		"entry": "dist/index.html",
		"window": {
			"title": "My Application",
			"width": 1280,
			"height": 720,
			"minWidth": 400,
			"minHeight": 300,
			"resizable": false,
			"frameless": true
		},
		"tray": true,
		"build": {
			"icon": "icon.png",
			"appId": "com.example.myapp"
		},
		"permissions": ["fs", "dialog", "clipboard"],
		"devCommand": "npm run dev",
		"buildCommand": "npm run build"
	}`
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte(config), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", cfg.Name)
	}
	if cfg.Entry != "dist/index.html" {
		t.Errorf("expected entry 'dist/index.html', got %q", cfg.Entry)
	}
	if cfg.Window.Title != "My Application" {
		t.Errorf("expected title 'My Application', got %q", cfg.Window.Title)
	}
	if cfg.Window.Width != 1280 {
		t.Errorf("expected width 1280, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 720 {
		t.Errorf("expected height 720, got %d", cfg.Window.Height)
	}
	if cfg.Window.MinWidth != 400 {
		t.Errorf("expected minWidth 400, got %d", cfg.Window.MinWidth)
	}
	if cfg.Window.MinHeight != 300 {
		t.Errorf("expected minHeight 300, got %d", cfg.Window.MinHeight)
	}
	if cfg.Window.Resizable == nil || *cfg.Window.Resizable {
		t.Error("expected resizable to be false")
	}
	if !cfg.Window.Frameless {
		t.Error("expected frameless to be true")
	}
	if !cfg.Tray {
		t.Error("expected tray to be true")
	}
	if cfg.Build.Icon != "icon.png" {
		t.Errorf("expected icon 'icon.png', got %q", cfg.Build.Icon)
	}
	if cfg.Build.AppID != "com.example.myapp" {
		t.Errorf("expected appId 'com.example.myapp', got %q", cfg.Build.AppID)
	}
	if len(cfg.Permissions) != 3 {
		t.Errorf("expected 3 permissions, got %d", len(cfg.Permissions))
	}
	if cfg.DevCommand != "npm run dev" {
		t.Errorf("expected devCommand 'npm run dev', got %q", cfg.DevCommand)
	}
	if cfg.BuildCommand != "npm run build" {
		t.Errorf("expected buildCommand 'npm run build', got %q", cfg.BuildCommand)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing lightshell.json")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte("{invalid"), 0644)

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadConfigEmptyJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte("{}"), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have all defaults applied
	if cfg.Window.Width != 1024 {
		t.Errorf("expected default width, got %d", cfg.Window.Width)
	}
	if cfg.Entry != "src/index.html" {
		t.Errorf("expected default entry, got %q", cfg.Entry)
	}
}

func TestLoadConfigWindowTitleDefaultsToName(t *testing.T) {
	dir := t.TempDir()
	config := `{"name": "cool-app"}`
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte(config), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Window.Title != "cool-app" {
		t.Errorf("expected title to default to name 'cool-app', got %q", cfg.Window.Title)
	}
}

func TestLoadConfigCustomTitle(t *testing.T) {
	dir := t.TempDir()
	config := `{"name": "myapp", "window": {"title": "Custom Title"}}`
	os.WriteFile(filepath.Join(dir, "lightshell.json"), []byte(config), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Window.Title != "Custom Title" {
		t.Errorf("expected title 'Custom Title', got %q", cfg.Window.Title)
	}
}
