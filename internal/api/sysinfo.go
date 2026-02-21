package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/lightshell-dev/lightshell/internal/ipc"
)

// RegisterSystem registers system info API handlers.
func RegisterSystem(router *ipc.Router, appVersion string, appName string) {
	router.Handle("system.platform", func(params json.RawMessage) (any, error) {
		return goruntime.GOOS, nil
	})

	router.Handle("system.arch", func(params json.RawMessage) (any, error) {
		return goruntime.GOARCH, nil
	})

	router.Handle("system.homeDir", func(params json.RawMessage) (any, error) {
		home, err := os.UserHomeDir()
		return home, err
	})

	router.Handle("system.tempDir", func(params json.RawMessage) (any, error) {
		return os.TempDir(), nil
	})

	router.Handle("system.hostname", func(params json.RawMessage) (any, error) {
		return os.Hostname()
	})

	router.Handle("app.quit", func(params json.RawMessage) (any, error) {
		os.Exit(0)
		return nil, nil
	})

	router.Handle("app.version", func(params json.RawMessage) (any, error) {
		return appVersion, nil
	})

	router.Handle("app.dataDir", func(params json.RawMessage) (any, error) {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		var dir string
		switch goruntime.GOOS {
		case "darwin":
			dir = filepath.Join(home, "Library", "Application Support", appName)
		default:
			dir = filepath.Join(home, ".local", "share", appName)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		return dir, nil
	})
}
