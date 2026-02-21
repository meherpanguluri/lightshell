package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lightshell-dev/lightshell/internal/webview"
)

// Config represents the lightshell.json configuration.
type Config struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Entry       string       `json:"entry"`
	Window      WindowConfig `json:"window"`
	Tray        bool         `json:"tray"`
	Build       BuildConfig  `json:"build"`
	Permissions []string     `json:"permissions"`
}

type WindowConfig struct {
	Title     string `json:"title"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	MinWidth  int    `json:"minWidth"`
	MinHeight int    `json:"minHeight"`
	Resizable *bool  `json:"resizable"`
	Frameless bool   `json:"frameless"`
}

type BuildConfig struct {
	Icon  string `json:"icon"`
	AppID string `json:"appId"`
}

// App is the main LightShell application.
type App struct {
	Config     Config
	Webview    webview.Webview
	ProjectDir string
	DevMode    bool
	DevURL     string // set by dev server to load URL instead of file
}

// LoadConfig reads and parses lightshell.json from the given directory.
func LoadConfig(dir string) (Config, error) {
	path := filepath.Join(dir, "lightshell.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read lightshell.json: %w\n\nMake sure you're in a LightShell project directory, or run 'lightshell init' to create one.", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid lightshell.json: %w", err)
	}

	// Defaults
	if cfg.Window.Width == 0 {
		cfg.Window.Width = 1024
	}
	if cfg.Window.Height == 0 {
		cfg.Window.Height = 768
	}
	if cfg.Window.Title == "" {
		cfg.Window.Title = cfg.Name
	}
	if cfg.Window.Resizable == nil {
		t := true
		cfg.Window.Resizable = &t
	}
	if cfg.Entry == "" {
		cfg.Entry = "src/index.html"
	}

	return cfg, nil
}

// Run starts the LightShell application.
func (a *App) Run() error {
	wv := webview.New()
	a.Webview = wv

	wcfg := webview.WindowConfig{
		Title:     a.Config.Window.Title,
		Width:     a.Config.Window.Width,
		Height:    a.Config.Window.Height,
		MinWidth:  a.Config.Window.MinWidth,
		MinHeight: a.Config.Window.MinHeight,
		Resizable: *a.Config.Window.Resizable,
		Frameless: a.Config.Window.Frameless,
		DevTools:  a.DevMode,
	}

	if err := wv.Create(wcfg); err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	// Inject scripts will be done by the caller (CLI layer) after IPC is set up

	// Load content
	if a.DevURL != "" {
		if err := wv.LoadURL(a.DevURL); err != nil {
			return fmt.Errorf("failed to load dev URL: %w", err)
		}
	} else {
		entryPath := filepath.Join(a.ProjectDir, a.Config.Entry)
		absPath, err := filepath.Abs(entryPath)
		if err != nil {
			return fmt.Errorf("failed to resolve entry path: %w", err)
		}
		fileURL := "file://" + absPath
		if err := wv.LoadURL(fileURL); err != nil {
			return fmt.Errorf("failed to load entry: %w", err)
		}
	}

	return wv.Run()
}
