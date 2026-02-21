package cli

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	lsruntime "github.com/meherpanguluri/lightshell/internal/runtime"
)

//go:embed buildfiles/webview_darwin.m
var webviewDarwinM string

// Build compiles the app for the current platform.
func Build() error {
	start := time.Now()

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := lsruntime.LoadConfig(dir)
	if err != nil {
		return err
	}

	// Create staging directory
	staging, err := os.MkdirTemp("", "lightshell-build-*")
	if err != nil {
		return fmt.Errorf("failed to create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	// Copy user source into staging
	srcDir := filepath.Join(dir, filepath.Dir(cfg.Entry))
	stagingSrc := filepath.Join(staging, "src")
	if err := copyDir(srcDir, stagingSrc); err != nil {
		return fmt.Errorf("failed to stage source files: %w", err)
	}

	// Copy scripts (polyfills + lightshell client)
	stageScripts := filepath.Join(staging, "scripts")
	os.MkdirAll(stageScripts, 0o755)
	os.WriteFile(filepath.Join(stageScripts, "polyfills.js"), []byte(polyfillsJS), 0o644)
	os.WriteFile(filepath.Join(stageScripts, "lightshell.js"), []byte(clientJS), 0o644)

	// Generate the embed-based main.go for the built app
	buildMain := filepath.Join(staging, "main.go")
	if err := generateBuildMain(buildMain, cfg); err != nil {
		return fmt.Errorf("failed to generate build source: %w", err)
	}

	// Copy the Objective-C webview bridge
	if runtime.GOOS == "darwin" {
		os.WriteFile(filepath.Join(staging, "webview_darwin.m"), []byte(webviewDarwinM), 0o644)
	}

	// Generate go.mod for the staging dir
	buildGoMod := filepath.Join(staging, "go.mod")
	os.WriteFile(buildGoMod, []byte("module lightshell-app\n\ngo 1.23\n"), 0o644)

	// Compile the Go binary
	distDir := filepath.Join(dir, "dist")
	os.MkdirAll(distDir, 0o755)

	binaryName := cfg.Name
	if binaryName == "" {
		binaryName = "app"
	}
	binaryPath := filepath.Join(staging, binaryName)

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", binaryPath, ".")
	cmd.Dir = staging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Package for platform
	var outputPath string
	switch runtime.GOOS {
	case "darwin":
		outputPath, err = packageDarwin(binaryPath, distDir, cfg)
	default:
		return fmt.Errorf("build not yet supported on %s", runtime.GOOS)
	}
	if err != nil {
		return fmt.Errorf("packaging failed: %w", err)
	}

	// Print result
	sizeMB := float64(dirSize(outputPath)) / 1024 / 1024
	sizeStr := fmt.Sprintf("%.1fMB", sizeMB)

	elapsed := time.Since(start).Seconds()
	fmt.Printf("Built %s in %.1fs -> %s\n", cfg.Name, elapsed, sizeStr)
	fmt.Printf("Output: %s\n", outputPath)

	return nil
}

func generateBuildMain(path string, cfg lsruntime.Config) error {
	tmpl := `package main

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

#include <stdlib.h>

extern void WebviewCreate(const char* title, int width, int height, int minWidth, int minHeight,
	int resizable, int frameless, int alwaysOnTop, int transparent, int devTools);
extern void WebviewLoadURL(const char* url);
extern void WebviewLoadHTML(const char* html);
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
*/
import "C"

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

//go:embed src
var srcFS embed.FS

//go:embed scripts/polyfills.js
var polyfillsJS string

//go:embed scripts/lightshell.js
var clientJS string

var msgHandler func(string)
var ipcHandlers = map[string]func(json.RawMessage)(any, error){}

// Security: declared permissions and allowed paths
var permissions = map[string]bool{
{{- range .Permissions}}
	"{{.}}": true,
{{- end}}
}

func checkPerm(perm string) error {
	if !permissions[perm] {
		return fmt.Errorf("permission denied: %%q not declared in lightshell.json permissions", perm)
	}
	return nil
}

var allowedPaths []string

func initSecurity() {
	// Allow temp dir
	allowedPaths = append(allowedPaths, os.TempDir())
	// Allow app data dir
	if home, err := os.UserHomeDir(); err == nil {
		allowedPaths = append(allowedPaths, home+"/Library/Application Support/{{.Name}}")
		allowedPaths = append(allowedPaths, home+"/.local/share/{{.Name}}")
	}
}

func checkPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %%w", err)
	}
	for _, dir := range allowedPaths {
		if len(abs) >= len(dir) && abs[:len(dir)] == dir {
			return nil
		}
	}
	return fmt.Errorf("access denied: path %%q is outside allowed directories", path)
}

//export goMessageHandler
func goMessageHandler(msg *C.char) {
	if msgHandler != nil {
		msgHandler(C.GoString(msg))
	}
}

func evalJS(js string) {
	cJS := C.CString(js)
	C.WebviewEval(cJS)
	C.free(unsafe.Pointer(cJS))
}

func registerHandler(method string, fn func(json.RawMessage)(any, error)) {
	ipcHandlers[method] = fn
}

func handleMessage(rawMsg string) string {
	var req struct {
		ID     string          {{.BTick}}json:"id"{{.BTick}}
		Method string          {{.BTick}}json:"method"{{.BTick}}
		Params json.RawMessage {{.BTick}}json:"params"{{.BTick}}
	}
	if err := json.Unmarshal([]byte(rawMsg), &req); err != nil {
		resp, _ := json.Marshal(map[string]any{"id": "", "error": err.Error()})
		return string(resp)
	}
	handler, ok := ipcHandlers[req.Method]
	if !ok {
		resp, _ := json.Marshal(map[string]any{"id": req.ID, "error": "unknown method: " + req.Method})
		return string(resp)
	}
	result, err := handler(req.Params)
	if err != nil {
		resp, _ := json.Marshal(map[string]any{"id": req.ID, "error": err.Error()})
		return string(resp)
	}
	resp, _ := json.Marshal(map[string]any{"id": req.ID, "result": result})
	return string(resp)
}

func registerAPIs() {
	registerHandler("system.platform", func(p json.RawMessage) (any, error) {
		return runtime.GOOS, nil
	})
	registerHandler("system.arch", func(p json.RawMessage) (any, error) {
		return runtime.GOARCH, nil
	})
	registerHandler("system.homeDir", func(p json.RawMessage) (any, error) {
		return os.UserHomeDir()
	})
	registerHandler("system.tempDir", func(p json.RawMessage) (any, error) {
		return os.TempDir(), nil
	})
	registerHandler("system.hostname", func(p json.RawMessage) (any, error) {
		return os.Hostname()
	})
	registerHandler("app.version", func(p json.RawMessage) (any, error) {
		return "{{.Version}}", nil
	})
	registerHandler("app.quit", func(p json.RawMessage) (any, error) {
		os.Exit(0)
		return nil, nil
	})
	registerHandler("app.dataDir", func(p json.RawMessage) (any, error) {
		home, err := os.UserHomeDir()
		if err != nil { return nil, err }
		dir := home + "/Library/Application Support/{{.Name}}"
		os.MkdirAll(dir, 0755)
		return dir, nil
	})

	registerHandler("fs.readFile", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		data, err := os.ReadFile(params.Path)
		return string(data), err
	})
	registerHandler("fs.writeFile", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct {
			Path string {{.BTick}}json:"path"{{.BTick}}
			Data string {{.BTick}}json:"data"{{.BTick}}
		}
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		os.MkdirAll(filepath.Dir(params.Path), 0755)
		return nil, os.WriteFile(params.Path, []byte(params.Data), 0644)
	})
	registerHandler("fs.exists", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		_, err := os.Stat(params.Path)
		return err == nil, nil
	})
	registerHandler("fs.readDir", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		entries, err := os.ReadDir(params.Path)
		if err != nil { return nil, err }
		result := []map[string]any{}
		for _, e := range entries {
			info, _ := e.Info()
			size := int64(0)
			if info != nil { size = info.Size() }
			result = append(result, map[string]any{"name": e.Name(), "isDir": e.IsDir(), "size": size})
		}
		return result, nil
	})
	registerHandler("fs.stat", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		info, err := os.Stat(params.Path)
		if err != nil { return nil, err }
		return map[string]any{
			"name": info.Name(), "size": info.Size(), "isDir": info.IsDir(),
			"modTime": info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			"mode": info.Mode().String(),
		}, nil
	})
	registerHandler("fs.mkdir", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		return nil, os.MkdirAll(params.Path, 0755)
	})
	registerHandler("fs.remove", func(p json.RawMessage) (any, error) {
		if err := checkPerm("fs"); err != nil { return nil, err }
		var params struct { Path string {{.BTick}}json:"path"{{.BTick}} }
		json.Unmarshal(p, &params)
		if err := checkPath(params.Path); err != nil { return nil, err }
		return nil, os.RemoveAll(params.Path)
	})

	registerHandler("window.setTitle", func(p json.RawMessage) (any, error) {
		var params struct { Title string {{.BTick}}json:"title"{{.BTick}} }
		json.Unmarshal(p, &params)
		cTitle := C.CString(params.Title)
		defer C.free(unsafe.Pointer(cTitle))
		C.WebviewSetTitle(cTitle)
		return nil, nil
	})
	registerHandler("window.setSize", func(p json.RawMessage) (any, error) {
		var params struct { Width int {{.BTick}}json:"width"{{.BTick}}; Height int {{.BTick}}json:"height"{{.BTick}} }
		json.Unmarshal(p, &params)
		C.WebviewSetSize(C.int(params.Width), C.int(params.Height))
		return nil, nil
	})
	registerHandler("window.minimize", func(p json.RawMessage) (any, error) {
		C.WebviewMinimize()
		return nil, nil
	})
	registerHandler("window.maximize", func(p json.RawMessage) (any, error) {
		C.WebviewMaximize()
		return nil, nil
	})
	registerHandler("window.fullscreen", func(p json.RawMessage) (any, error) {
		C.WebviewFullscreen()
		return nil, nil
	})
	registerHandler("window.restore", func(p json.RawMessage) (any, error) {
		C.WebviewRestore()
		return nil, nil
	})
	registerHandler("window.close", func(p json.RawMessage) (any, error) {
		C.WebviewClose()
		return nil, nil
	})
}

func init() {
	runtime.LockOSThread()
}

func main() {
	subFS, err := fs.Sub(srcFS, "src")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	go http.Serve(listener, mux)

	initSecurity()
	registerAPIs()

	cTitle := C.CString("{{.Title}}")
	defer C.free(unsafe.Pointer(cTitle))
	C.WebviewCreate(cTitle, {{.Width}}, {{.Height}}, {{.MinWidth}}, {{.MinHeight}}, {{.ResizableInt}}, 0, 0, 0, 0)

	msgHandler = func(msg string) {
		response := handleMessage(msg)
		js := fmt.Sprintf("__lightshell_receive(%s)", response)
		evalJS(js)
	}

	evalJS(polyfillsJS)
	evalJS(clientJS)

	url := fmt.Sprintf("http://127.0.0.1:%d/{{.EntryFile}}", port)
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	C.WebviewLoadURL(cURL)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		C.WebviewDestroy()
		os.Exit(0)
	}()

	C.WebviewRun()
}
`
	t, err := template.New("main").Parse(tmpl)
	if err != nil {
		return err
	}

	resizable := 1
	if cfg.Window.Resizable != nil && !*cfg.Window.Resizable {
		resizable = 0
	}

	// Default to all permissions if none declared
	perms := cfg.Permissions
	if len(perms) == 0 {
		perms = []string{"fs", "dialog", "clipboard", "shell", "notification", "tray", "menu"}
	}

	data := map[string]any{
		"Title":        cfg.Window.Title,
		"Width":        cfg.Window.Width,
		"Height":       cfg.Window.Height,
		"MinWidth":     cfg.Window.MinWidth,
		"MinHeight":    cfg.Window.MinHeight,
		"ResizableInt": resizable,
		"Version":      cfg.Version,
		"Name":         cfg.Name,
		"EntryFile":    filepath.Base(cfg.Entry),
		"BTick":        "`",
		"Permissions":  perms,
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}

func packageDarwin(binaryPath, distDir string, cfg lsruntime.Config) (string, error) {
	title := cfg.Window.Title
	if title == "" {
		title = cfg.Name
	}
	appName := strings.ReplaceAll(title, " ", "") + ".app"
	appPath := filepath.Join(distDir, appName)

	// Clean previous build
	os.RemoveAll(appPath)

	// Create .app bundle structure
	macosDir := filepath.Join(appPath, "Contents", "MacOS")
	resDir := filepath.Join(appPath, "Contents", "Resources")
	os.MkdirAll(macosDir, 0o755)
	os.MkdirAll(resDir, 0o755)

	// Copy binary
	binaryDest := filepath.Join(macosDir, cfg.Name)
	input, err := os.ReadFile(binaryPath)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(binaryDest, input, 0o755); err != nil {
		return "", err
	}

	// Generate Info.plist
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	plist := generatePlist(cfg)
	if err := os.WriteFile(plistPath, []byte(plist), 0o644); err != nil {
		return "", err
	}

	return appPath, nil
}

func generatePlist(cfg lsruntime.Config) string {
	appID := cfg.Build.AppID
	if appID == "" {
		appID = "com.lightshell." + cfg.Name
	}
	title := cfg.Window.Title
	if title == "" {
		title = cfg.Name
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleExecutable</key>
	<string>%s</string>
	<key>CFBundleIdentifier</key>
	<string>%s</string>
	<key>CFBundleName</key>
	<string>%s</string>
	<key>CFBundleVersion</key>
	<string>%s</string>
	<key>CFBundleShortVersionString</key>
	<string>%s</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>`, cfg.Name, appID, title, cfg.Version, cfg.Version)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
