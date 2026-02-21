package webview

// Webview is the interface for platform-specific webview implementations.
type Webview interface {
	Create(config WindowConfig) error
	LoadHTML(html string) error
	LoadURL(url string) error
	Eval(js string) error
	SetTitle(title string) error
	SetSize(w, h int) error
	SetMinSize(w, h int) error
	SetMaxSize(w, h int) error
	SetPosition(x, y int) error
	Fullscreen() error
	Minimize() error
	Maximize() error
	Restore() error
	Close() error
	OnMessage(handler func(msg string))
	Run() error
	Destroy()
}

// WindowConfig holds the configuration for creating a webview window.
type WindowConfig struct {
	Title       string
	Width       int
	Height      int
	MinWidth    int
	MinHeight   int
	Resizable   bool
	Frameless   bool
	AlwaysOnTop bool
	Transparent bool
	DevTools    bool
}
