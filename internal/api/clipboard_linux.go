//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleClipboardRead(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("clipboard.read not yet implemented on linux")
}

func handleClipboardWrite(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("clipboard.write not yet implemented on linux")
}
