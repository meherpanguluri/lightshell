//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern const char* ClipboardRead();
extern void ClipboardWrite(const char* text);
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func handleClipboardRead(params json.RawMessage) (any, error) {
	result := C.ClipboardRead()
	if result == nil {
		return "", nil
	}
	return C.GoString(result), nil
}

func handleClipboardWrite(params json.RawMessage) (any, error) {
	var p struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	cText := C.CString(p.Text)
	defer C.free(unsafe.Pointer(cText))
	C.ClipboardWrite(cText)
	return nil, nil
}
