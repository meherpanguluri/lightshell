package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/meherpanguluri/lightshell/internal/api"
	"github.com/meherpanguluri/lightshell/internal/ipc"
	"github.com/meherpanguluri/lightshell/internal/runtime"
	"github.com/meherpanguluri/lightshell/internal/security"
	"github.com/meherpanguluri/lightshell/internal/webview"
)

// Dev runs the app in development mode with hot reload.
func Dev() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := runtime.LoadConfig(dir)
	if err != nil {
		return err
	}

	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("could not find free port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Determine the source directory from the entry path
	srcDir := filepath.Join(dir, filepath.Dir(cfg.Entry))

	// Start HTTP server for serving source files
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(srcDir)))

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Dev server error: %v\n", err)
		}
	}()

	entryFile := filepath.Base(cfg.Entry)
	devURL := fmt.Sprintf("http://127.0.0.1:%d/%s", port, entryFile)

	fmt.Printf("Dev server running at http://127.0.0.1:%d\n", port)

	// Set up IPC router and register APIs
	router := ipc.NewRouter()

	// Create the webview
	wv := webview.New()
	wcfg := webview.WindowConfig{
		Title:     cfg.Window.Title,
		Width:     cfg.Window.Width,
		Height:    cfg.Window.Height,
		MinWidth:  cfg.Window.MinWidth,
		MinHeight: cfg.Window.MinHeight,
		Resizable: true,
		Frameless: cfg.Window.Frameless,
		DevTools:  true,
	}
	if cfg.Window.Resizable != nil {
		wcfg.Resizable = *cfg.Window.Resizable
	}

	if err := wv.Create(wcfg); err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}

	// Wire IPC: webview messages go to router, router can eval JS back
	router.SetEvalFunc(func(js string) {
		wv.Eval(js)
	})
	wv.OnMessage(func(msg string) {
		response := router.HandleMessage(msg)
		js := fmt.Sprintf("__lightshell_receive(%s)", response)
		wv.Eval(js)
	})

	// Dev mode: all permissions granted
	policy := security.DevPolicy()

	// Register all APIs with security policy
	api.RegisterWindow(router, wv)
	api.RegisterFS(router, policy)
	api.RegisterDialog(router, policy)
	api.RegisterClipboard(router, policy)
	api.RegisterShell(router, policy)
	api.RegisterSystem(router, cfg.Version, cfg.Name)
	api.RegisterNotification(router, policy)
	api.RegisterTray(router, policy)
	api.RegisterMenu(router, policy)

	// Inject polyfills + client library
	injectScripts(wv)

	// Load the dev URL
	if err := wv.LoadURL(devURL); err != nil {
		return fmt.Errorf("failed to load dev URL: %w", err)
	}

	// Start file watcher for hot reload
	go watchFiles(srcDir, func() {
		fmt.Println("File changed, reloading...")
		wv.Eval("location.reload()")
	})

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		server.Close()
		wv.Destroy()
		os.Exit(0)
	}()

	// Run the event loop (blocking)
	return wv.Run()
}

func injectScripts(wv webview.Webview) {
	wv.Eval(polyfillsJS)
	wv.Eval(clientJS)
}

func watchFiles(dir string, onchange func()) {
	lastMod := map[string]int64{}
	var mu sync.Mutex

	scan := func() bool {
		changed := false
		mu.Lock()
		defer mu.Unlock()
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			mod := info.ModTime().UnixNano()
			if prev, ok := lastMod[path]; ok && prev != mod {
				changed = true
			}
			lastMod[path] = mod
			return nil
		})
		return changed
	}

	// Initial scan to populate timestamps
	scan()

	for {
		time.Sleep(500 * time.Millisecond)
		if scan() {
			onchange()
		}
	}
}
