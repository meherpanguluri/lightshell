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

	"github.com/lightshell-dev/lightshell/internal/api"
	"github.com/lightshell-dev/lightshell/internal/ipc"
	"github.com/lightshell-dev/lightshell/internal/runtime"
	"github.com/lightshell-dev/lightshell/internal/security"
	"github.com/lightshell-dev/lightshell/internal/webview"
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

	// Check for --mcp-socket flag (used when launched by the MCP server)
	var mcpSocketPath string
	for i, arg := range os.Args {
		if arg == "--mcp-socket" && i+1 < len(os.Args) {
			mcpSocketPath = os.Args[i+1]
		}
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

	// Set up the MCP socket server if --mcp-socket was specified.
	// This must be done before wiring OnMessage so we can intercept MCP messages.
	var mcpSrv *mcpSocketServer
	if mcpSocketPath != "" {
		mcpSrv = newMCPSocketServer(mcpSocketPath, wv)
	}

	// Wire IPC: webview messages go to router, router can eval JS back.
	// When running in MCP mode, messages are first checked for MCP-specific
	// prefixes (console forwarding, eval results) before being routed to IPC.
	router.SetEvalFunc(func(js string) {
		wv.Eval(js)
	})
	wv.OnMessage(func(msg string) {
		// If MCP socket server is active, check for MCP-specific messages first
		if mcpSrv != nil && mcpSrv.handleMCPMessage(msg) {
			return // was an MCP message, don't route to IPC
		}

		response := router.HandleMessage(msg)
		js := fmt.Sprintf("__lightshell_receive(%s)", response)
		wv.Eval(js)
	})

	// Dev mode: all permissions granted
	policy := security.DevPolicy()

	// Register all APIs with security policy
	api.RegisterWindow(router, wv)
	api.RegisterWindowExtended(router, wv)
	api.RegisterFS(router, policy)
	api.RegisterDialog(router, policy)
	api.RegisterClipboard(router, policy)
	api.RegisterShell(router, policy)
	api.RegisterSystem(router, cfg.Version, cfg.Name, wv)
	api.RegisterNotification(router, policy)
	api.RegisterTray(router, policy)
	api.RegisterMenu(router, policy)
	api.RegisterAppExtended(router, cfg.Name)

	// Set up dev tray with Debug Console menu
	api.SetupDevTray(func(js string) { wv.Eval(js) })

	// Inject polyfills + client library + debug console
	injectScripts(wv)

	// If MCP mode, inject the console forwarding script that wraps
	// console.log/warn/error to forward entries to Go via postMessage
	if mcpSrv != nil {
		wv.AddUserScript(mcpConsoleForwardScript)
	}

	// Load the dev URL
	if err := wv.LoadURL(devURL); err != nil {
		return fmt.Errorf("failed to load dev URL: %w", err)
	}

	// Start the MCP socket server if configured
	if mcpSrv != nil {
		go func() {
			if err := mcpSrv.serve(); err != nil {
				fmt.Fprintf(os.Stderr, "MCP socket server error: %v\n", err)
			}
		}()
		defer mcpSrv.close()
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
		if mcpSrv != nil {
			mcpSrv.close()
		}
		server.Close()
		wv.Destroy()
		os.Exit(0)
	}()

	// Run the event loop (blocking)
	return wv.Run()
}

func injectScripts(wv webview.Webview) {
	// Use AddUserScript so scripts persist across page navigations (including initial LoadURL)
	wv.AddUserScript(polyfillsJS)
	wv.AddUserScript(clientJS)
	wv.AddUserScript(debugConsoleJS)
	// Inject defaults CSS as a <style> tag
	cssInjection := fmt.Sprintf(`(function(){var s=document.createElement('style');s.id='lightshell-defaults';s.textContent=%q;document.head.insertBefore(s,document.head.firstChild)})()`, defaultsCSS)
	wv.AddUserScript(cssInjection)
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
