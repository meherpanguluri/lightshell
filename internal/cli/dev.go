package cli

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
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

	// If a dev command is configured, delegate to bundler-aware dev mode
	if cfg.DevCommand != "" {
		return devWithBundler(dir, cfg)
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
		router.RunShutdownHooks()
		if mcpSrv != nil {
			mcpSrv.close()
		}
		server.Close()
		wv.Destroy()
		os.Exit(0)
	}()

	// Run the event loop (blocking)
	err = wv.Run()
	router.RunShutdownHooks()
	return err
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

// devWithBundler runs in dev mode using an external dev server (e.g. Vite).
func devWithBundler(dir string, cfg runtime.Config) error {
	// Check node_modules exists
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); os.IsNotExist(err) {
		return fmt.Errorf("node_modules not found. Run 'npm install' first")
	}

	// Parse port from devCommand
	port := parsePort(cfg.DevCommand)
	if port == 0 {
		port = 5173 // Vite default
	}

	// Start the dev command as a child process
	parts := strings.Fields(cfg.DevCommand)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start dev command %q: %w", cfg.DevCommand, err)
	}

	// Wait for dev server to be ready
	devURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := waitForServer(devURL, 30*time.Second); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("dev server did not start within 30s: %w", err)
	}

	fmt.Printf("Framework dev server running at %s\n", devURL)

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
		cmd.Process.Kill()
		return fmt.Errorf("failed to create window: %w", err)
	}

	// Wire IPC
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

	// Register all APIs
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

	// Set up dev tray
	api.SetupDevTray(func(js string) { wv.Eval(js) })

	// Inject polyfills + client library + debug console
	injectScripts(wv)

	// Load the Vite dev URL
	if err := wv.LoadURL(devURL); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to load dev URL: %w", err)
	}

	// No file watcher needed — Vite handles HMR natively

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		router.RunShutdownHooks()
		cmd.Process.Kill()
		wv.Destroy()
		os.Exit(0)
	}()

	// Run the event loop (blocking)
	err := wv.Run()
	router.RunShutdownHooks()
	cmd.Process.Kill()
	return err
}

// parsePort extracts the port number from a command string (looks for --port NNNN).
func parsePort(cmd string) int {
	parts := strings.Fields(cmd)
	for i, p := range parts {
		if p == "--port" && i+1 < len(parts) {
			port, err := strconv.Atoi(parts[i+1])
			if err == nil {
				return port
			}
		}
	}
	return 0
}

// waitForServer polls a URL until it responds or the timeout expires.
func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
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
