//go:build darwin

package webview

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

#include <stdlib.h>

extern void WebviewCreate(const char* title, int width, int height, int minWidth, int minHeight,
	int resizable, int frameless, int alwaysOnTop, int transparent, int devTools);
extern void WebviewLoadHTML(const char* html);
extern void WebviewLoadURL(const char* url);
extern void WebviewEval(const char* js);
extern void WebviewSetTitle(const char* title);
extern void WebviewSetSize(int width, int height);
extern void WebviewSetMinSize(int width, int height);
extern void WebviewSetMaxSize(int width, int height);
extern void WebviewSetPosition(int x, int y);
extern void WebviewFullscreen();
extern void WebviewMinimize();
extern void WebviewMaximize();
extern void WebviewRestore();
extern void WebviewClose();
extern void WebviewRun();
extern void WebviewDestroy();
extern void WebviewSetContentProtection(int enabled);
extern void WebviewSetVibrancy(const char* style);
extern void WebviewSetColorScheme(const char* scheme);
extern void WebviewEnableFileDrop(void);
*/
import "C"
import (
	"unsafe"
)

var messageHandler func(string)

//export goMessageHandler
func goMessageHandler(msg *C.char) {
	if messageHandler != nil {
		messageHandler(C.GoString(msg))
	}
}

// DarwinWebview implements the Webview interface for macOS using WKWebView.
type DarwinWebview struct{}

// New creates a new macOS webview.
func New() Webview {
	return &DarwinWebview{}
}

func (w *DarwinWebview) Create(config WindowConfig) error {
	cTitle := C.CString(config.Title)
	defer C.free(unsafe.Pointer(cTitle))

	resizable := 0
	if config.Resizable {
		resizable = 1
	}
	frameless := 0
	if config.Frameless {
		frameless = 1
	}
	alwaysOnTop := 0
	if config.AlwaysOnTop {
		alwaysOnTop = 1
	}
	transparent := 0
	if config.Transparent {
		transparent = 1
	}
	devTools := 0
	if config.DevTools {
		devTools = 1
	}

	C.WebviewCreate(cTitle, C.int(config.Width), C.int(config.Height),
		C.int(config.MinWidth), C.int(config.MinHeight),
		C.int(resizable), C.int(frameless), C.int(alwaysOnTop),
		C.int(transparent), C.int(devTools))
	return nil
}

func (w *DarwinWebview) LoadHTML(html string) error {
	cHTML := C.CString(html)
	defer C.free(unsafe.Pointer(cHTML))
	C.WebviewLoadHTML(cHTML)
	return nil
}

func (w *DarwinWebview) LoadURL(url string) error {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	C.WebviewLoadURL(cURL)
	return nil
}

func (w *DarwinWebview) Eval(js string) error {
	cJS := C.CString(js)
	defer C.free(unsafe.Pointer(cJS))
	C.WebviewEval(cJS)
	return nil
}

func (w *DarwinWebview) SetTitle(title string) error {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.WebviewSetTitle(cTitle)
	return nil
}

func (w *DarwinWebview) SetSize(width, height int) error {
	C.WebviewSetSize(C.int(width), C.int(height))
	return nil
}

func (w *DarwinWebview) SetMinSize(width, height int) error {
	C.WebviewSetMinSize(C.int(width), C.int(height))
	return nil
}

func (w *DarwinWebview) SetMaxSize(width, height int) error {
	C.WebviewSetMaxSize(C.int(width), C.int(height))
	return nil
}

func (w *DarwinWebview) SetPosition(x, y int) error {
	C.WebviewSetPosition(C.int(x), C.int(y))
	return nil
}

func (w *DarwinWebview) Fullscreen() error {
	C.WebviewFullscreen()
	return nil
}

func (w *DarwinWebview) Minimize() error {
	C.WebviewMinimize()
	return nil
}

func (w *DarwinWebview) Maximize() error {
	C.WebviewMaximize()
	return nil
}

func (w *DarwinWebview) Restore() error {
	C.WebviewRestore()
	return nil
}

func (w *DarwinWebview) Close() error {
	C.WebviewClose()
	return nil
}

func (w *DarwinWebview) SetContentProtection(enabled bool) error {
	e := 0
	if enabled {
		e = 1
	}
	C.WebviewSetContentProtection(C.int(e))
	return nil
}

func (w *DarwinWebview) SetVibrancy(style string) error {
	cStyle := C.CString(style)
	defer C.free(unsafe.Pointer(cStyle))
	C.WebviewSetVibrancy(cStyle)
	return nil
}

func (w *DarwinWebview) SetColorScheme(scheme string) error {
	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))
	C.WebviewSetColorScheme(cScheme)
	return nil
}

func (w *DarwinWebview) EnableFileDrop() error {
	C.WebviewEnableFileDrop()
	return nil
}

func (w *DarwinWebview) OnMessage(handler func(msg string)) {
	messageHandler = handler
}

func (w *DarwinWebview) Run() error {
	C.WebviewRun()
	return nil
}

func (w *DarwinWebview) Destroy() {
	C.WebviewDestroy()
}
