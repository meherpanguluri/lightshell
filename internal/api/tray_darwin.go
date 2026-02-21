//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern void TraySet(const char* tooltip);
extern void TrayRemove();
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func handleTraySet(params json.RawMessage) (any, error) {
	var p struct {
		Tooltip string `json:"tooltip"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	cTooltip := C.CString(p.Tooltip)
	defer C.free(unsafe.Pointer(cTooltip))
	C.TraySet(cTooltip)
	return nil, nil
}

func handleTrayRemove(params json.RawMessage) (any, error) {
	C.TrayRemove()
	return nil, nil
}
