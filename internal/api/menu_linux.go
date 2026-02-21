//go:build linux

package api

import (
	"encoding/json"
	"fmt"
)

func handleMenuSet(params json.RawMessage) (any, error) {
	return nil, fmt.Errorf("menu.set not yet implemented on linux")
}
