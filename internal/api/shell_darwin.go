//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern void ShellOpen(const char* url);
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func handleShellOpen(params json.RawMessage) (any, error) {
	var p struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	cURL := C.CString(p.URL)
	defer C.free(unsafe.Pointer(cURL))
	C.ShellOpen(cURL)
	return nil, nil
}
