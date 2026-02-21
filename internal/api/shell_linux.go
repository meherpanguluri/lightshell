//go:build linux

package api

import (
	"encoding/json"
	"os/exec"
)

func handleShellOpen(params json.RawMessage) (any, error) {
	var p struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return nil, exec.Command("xdg-open", p.URL).Start()
}
