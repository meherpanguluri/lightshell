//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern void AppSetBadgeCount(int count);
*/
import "C"
import (
	"encoding/json"
)

func handleAppSetBadgeCount(params json.RawMessage) (any, error) {
	var p struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	C.AppSetBadgeCount(C.int(p.Count))
	return nil, nil
}
