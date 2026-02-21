//go:build darwin

package api

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <stdlib.h>

extern const char* DialogOpen(const char* title, const char* defaultPath, int directory, int multiple);
extern const char* DialogSave(const char* title, const char* defaultPath);
extern void DialogMessage(const char* title, const char* message);
extern int DialogConfirm(const char* title, const char* message);
extern const char* DialogPrompt(const char* title, const char* defaultValue);
*/
import "C"
import (
	"encoding/json"
	"strings"
	"unsafe"
)

func handleDialogOpen(params json.RawMessage) (any, error) {
	var p struct {
		Title       string `json:"title"`
		DefaultPath string `json:"defaultPath"`
		Directory   bool   `json:"directory"`
		Multiple    bool   `json:"multiple"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cDefault := C.CString(p.DefaultPath)
	defer C.free(unsafe.Pointer(cDefault))

	dir := 0
	if p.Directory {
		dir = 1
	}
	multi := 0
	if p.Multiple {
		multi = 1
	}

	result := C.DialogOpen(cTitle, cDefault, C.int(dir), C.int(multi))
	if result == nil {
		return nil, nil
	}
	goResult := C.GoString(result)
	if goResult == "" {
		return nil, nil
	}

	if p.Multiple {
		paths := strings.Split(goResult, "\n")
		return paths, nil
	}
	return goResult, nil
}

func handleDialogSave(params json.RawMessage) (any, error) {
	var p struct {
		Title       string `json:"title"`
		DefaultPath string `json:"defaultPath"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cDefault := C.CString(p.DefaultPath)
	defer C.free(unsafe.Pointer(cDefault))

	result := C.DialogSave(cTitle, cDefault)
	if result == nil {
		return nil, nil
	}
	goResult := C.GoString(result)
	if goResult == "" {
		return nil, nil
	}
	return goResult, nil
}

func handleDialogMessage(params json.RawMessage) (any, error) {
	var p struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cMsg := C.CString(p.Message)
	defer C.free(unsafe.Pointer(cMsg))

	C.DialogMessage(cTitle, cMsg)
	return nil, nil
}

func handleDialogConfirm(params json.RawMessage) (any, error) {
	var p struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cMsg := C.CString(p.Message)
	defer C.free(unsafe.Pointer(cMsg))

	result := C.DialogConfirm(cTitle, cMsg)
	return result == 1, nil
}

func handleDialogPrompt(params json.RawMessage) (any, error) {
	var p struct {
		Title   string `json:"title"`
		Default string `json:"default"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	cTitle := C.CString(p.Title)
	defer C.free(unsafe.Pointer(cTitle))
	cDefault := C.CString(p.Default)
	defer C.free(unsafe.Pointer(cDefault))

	result := C.DialogPrompt(cTitle, cDefault)
	if result == nil {
		return nil, nil
	}
	return C.GoString(result), nil
}
