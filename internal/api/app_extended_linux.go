//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleAppSetBadgeCount(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("app.setBadgeCount not yet implemented on linux")
}
