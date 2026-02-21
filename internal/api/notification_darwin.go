//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa -framework UserNotifications

#include <stdlib.h>

extern void NotifySend(const char* title, const char* body);
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func handleNotifySend(params json.RawMessage) (any, error) {
	var p struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cBody := C.CString(p.Body)
	defer C.free(unsafe.Pointer(cBody))
	C.NotifySend(cTitle, cBody)
	return nil, nil
}
