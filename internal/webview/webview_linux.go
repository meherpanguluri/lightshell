//go:build linux

package webview

import "fmt"

type LinuxWebview struct{}

func New() Webview {
	return &LinuxWebview{}
}

func (w *LinuxWebview) Create(config WindowConfig) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) LoadHTML(html string) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) LoadURL(url string) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Eval(js string) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) SetTitle(title string) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) SetSize(w2, h int) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) SetMinSize(w2, h int) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) SetMaxSize(w2, h int) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) SetPosition(x, y int) error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Fullscreen() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Minimize() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Maximize() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Restore() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Close() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) OnMessage(handler func(msg string)) {}

func (w *LinuxWebview) Run() error {
	return fmt.Errorf("linux webview not yet implemented")
}

func (w *LinuxWebview) Destroy() {}
