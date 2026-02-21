package api

import (
	"encoding/json"

	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/webview"
)

// RegisterWindow registers window management API handlers.
func RegisterWindow(router *ipc.Router, wv webview.Webview) {
	router.Handle("window.setTitle", func(params json.RawMessage) (any, error) {
		var p struct {
			Title string `json:"title"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, wv.SetTitle(p.Title)
	})

	router.Handle("window.setSize", func(params json.RawMessage) (any, error) {
		var p struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, wv.SetSize(p.Width, p.Height)
	})

	router.Handle("window.getSize", func(params json.RawMessage) (any, error) {
		// TODO: implement actual size retrieval from native window
		return map[string]int{"width": 0, "height": 0}, nil
	})

	router.Handle("window.setPosition", func(params json.RawMessage) (any, error) {
		var p struct {
			X int `json:"x"`
			Y int `json:"y"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return nil, wv.SetPosition(p.X, p.Y)
	})

	router.Handle("window.getPosition", func(params json.RawMessage) (any, error) {
		// TODO: implement actual position retrieval from native window
		return map[string]int{"x": 0, "y": 0}, nil
	})

	router.Handle("window.minimize", func(params json.RawMessage) (any, error) {
		return nil, wv.Minimize()
	})

	router.Handle("window.maximize", func(params json.RawMessage) (any, error) {
		return nil, wv.Maximize()
	})

	router.Handle("window.fullscreen", func(params json.RawMessage) (any, error) {
		return nil, wv.Fullscreen()
	})

	router.Handle("window.restore", func(params json.RawMessage) (any, error) {
		return nil, wv.Restore()
	})

	router.Handle("window.close", func(params json.RawMessage) (any, error) {
		return nil, wv.Close()
	})
}
