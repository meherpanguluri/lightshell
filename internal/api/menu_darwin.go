//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern void MenuSet(const char* jsonTemplate);
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func handleMenuSet(params json.RawMessage) (any, error) {
	var p struct {
		Template json.RawMessage `json:"template"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	cJSON := C.CString(string(p.Template))
	defer C.free(unsafe.Pointer(cJSON))
	C.MenuSet(cJSON)
	return nil, nil
}
