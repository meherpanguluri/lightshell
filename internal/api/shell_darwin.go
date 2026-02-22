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
	"fmt"
	"net/url"
	"unsafe"
)

func handleShellOpen(params json.RawMessage) (any, error) {
	var p struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if err := validateShellOpenURL(p.URL); err != nil {
		return nil, err
	}
	cURL := C.CString(p.URL)
	defer C.free(unsafe.Pointer(cURL))
	C.ShellOpen(cURL)
	return nil, nil
}

func validateShellOpenURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("shell.open: invalid URL: %w", err)
	}
	switch parsed.Scheme {
	case "http", "https", "mailto":
		return nil
	case "":
		return fmt.Errorf("shell.open: URL must have a scheme (http, https, or mailto)")
	default:
		return fmt.Errorf("shell.open: scheme %q not allowed (only http, https, mailto)", parsed.Scheme)
	}
}
