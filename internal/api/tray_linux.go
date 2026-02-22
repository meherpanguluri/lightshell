//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleTraySet(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("tray.set not yet implemented on linux")
}

func handleTrayRemove(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("tray.remove not yet implemented on linux")
}

func SetupDevTray(evalFunc func(string)) {}

