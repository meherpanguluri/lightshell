//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleDialogOpen(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("dialog.open not yet implemented on linux")
}

func handleDialogSave(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("dialog.save not yet implemented on linux")
}

func handleDialogMessage(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("dialog.message not yet implemented on linux")
}

func handleDialogConfirm(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("dialog.confirm not yet implemented on linux")
}

func handleDialogPrompt(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("dialog.prompt not yet implemented on linux")
}
