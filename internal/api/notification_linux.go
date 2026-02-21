//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleNotifySend(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("notify.send not yet implemented on linux")
}
