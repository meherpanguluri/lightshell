//go:build linux

package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
)

func handleShellOpen(params json.RawMessage) (any, error) {
	var p struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if err := validateShellOpenURL(p.URL); err != nil {
		return nil, err
	}
	return nil, exec.Command("xdg-open", p.URL).Start()
}

func validateShellOpenURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("shell.open: invalid URL: %w", err)
	}
	switch parsed.Scheme {
	case "http", "https", "mailto":
		return nil
	case "":
		return fmt.Errorf("shell.open: URL must have a scheme (http, https, or mailto)")
	default:
		return fmt.Errorf("shell.open: scheme %q not allowed (only http, https, mailto)", parsed.Scheme)
	}
}
