package api

import (
	"encoding/json"
	"fmt"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/webview"
)

// RegisterWindowExtended registers extended window API handlers.
// These are window APIs beyond the basic set (setTitle, setSize, etc.)
// that are registered separately to keep the core window.go minimal.
func RegisterWindowExtended(router *ipc.Router, wv webview.Webview) {
	router.Handle("window.setContentProtection", func(params json.RawMessage) (any, error) {
		var p struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, wv.SetContentProtection(p.Enabled)
	})

	router.Handle("window.setVibrancy", func(params json.RawMessage) (any, error) {
		var p struct {
			Style string `json:"style"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		validStyles := map[string]bool{"sidebar": true, "header": true, "content": true, "sheet": true}
		if !validStyles[p.Style] {
			return nil, fmt.Errorf("invalid vibrancy style %q: must be one of sidebar, header, content, sheet", p.Style)
		}
		return nil, wv.SetVibrancy(p.Style)
	})

	router.Handle("window.setColorScheme", func(params json.RawMessage) (any, error) {
		var p struct {
			Scheme string `json:"scheme"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		validSchemes := map[string]bool{"light": true, "dark": true, "system": true}
		if !validSchemes[p.Scheme] {
			return nil, fmt.Errorf("invalid color scheme %q: must be one of light, dark, system", p.Scheme)
		}
		return nil, wv.SetColorScheme(p.Scheme)
	})

	router.Handle("window.enableFileDrop", func(params json.RawMessage) (any, error) {
		return nil, wv.EnableFileDrop()
	})
}
