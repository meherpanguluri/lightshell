---
title: Architecture
description: How LightShell works under the hood — Go backend, system webviews, IPC, and script injection.
---

LightShell is a desktop app framework where you write JavaScript, HTML, and CSS, and get a native binary. This page explains how the pieces fit together.

## The Stack

```
┌─────────────────────────────────────────────┐
│  Your JS/HTML/CSS                           │
├─────────────────────────────────────────────┤
│  lightshell.js (injected client library)    │
│  → exposes window.lightshell.* APIs         │
│  → communicates with Go over IPC            │
├─────────────────────────────────────────────┤
│  polyfills.js (injected platform fixes)     │
│  → normalizes WebKitGTK quirks              │
│  → form element resets                      │
│  → scrollbar + font stack normalization     │
├─────────────────────────────────────────────┤
│  IPC Layer (Unix Domain Socket + JSON)      │
├─────────────────────────────────────────────┤
│  Go Runtime (you never see this)            │
│  → Webview management                       │
│  → Native API handlers                      │
│  → Window, FS, dialogs, clipboard, etc.     │
├─────────────────────────────────────────────┤
│  System Webview                             │
│  → WKWebView (macOS) / WebKitGTK (Linux)   │
├─────────────────────────────────────────────┤
│  OS (macOS / Linux)                         │
└─────────────────────────────────────────────┘
```

## Why Go?

LightShell uses Go for its backend runtime. You never write Go, see Go, or configure Go. It is an implementation detail — like how esbuild uses Go but users only interact with the npm CLI.

Go was chosen because:
- **Single binary**: Go compiles to a single executable with no runtime dependencies
- **Cross-compilation**: one command builds for macOS arm64, macOS x64, Linux x64, Linux arm64
- **Small binaries**: a typical Go binary is ~2MB, compared to 50MB+ for bundled runtimes
- **AI fluency**: AI models generate correct Go code reliably, making the project maintainable by AI agents
- **cgo support**: Go can call C and Objective-C code directly, which is needed for webview integration

## System Webviews

LightShell does not bundle a browser. It uses the webview already installed on the user's operating system:

| Platform | Webview | Technology |
|----------|---------|-----------|
| macOS | WKWebView | Cocoa + WebKit (via Objective-C bridge) |
| Linux | WebKitGTK 2.40+ | GTK3 + WebKit (via C bridge) |

This is the main reason LightShell binaries are small. Electron bundles Chromium (~120MB). Tauri bundles a smaller webview layer. LightShell uses what is already there — zero binary overhead for the browser engine.

### macOS: WKWebView

On macOS, LightShell creates an `NSWindow` with a `WKWebView` programmatically (no XIB or storyboard). The Go code calls Objective-C through cgo:

- Window management uses `NSWindow` and its delegate
- JavaScript execution uses `evaluateJavaScript:`
- JS-to-Go messages use `WKScriptMessageHandler`
- DevTools are enabled via `WKPreferences._developerExtrasEnabled` in dev mode

### Linux: WebKitGTK

On Linux, LightShell creates a `GtkWindow` with a `WebKitWebView`. The Go code calls C through cgo:

- Window management uses GTK3 window APIs
- JavaScript execution uses `webkit_web_view_evaluate_javascript`
- JS-to-Go messages use `webkit_user_content_manager_register_script_message_handler`
- Web Inspector is enabled in dev mode

Platform-specific code uses Go build tags (`//go:build darwin` and `//go:build linux`) so only the relevant code compiles on each platform.

## Script Injection Order

When the webview loads, LightShell injects scripts in this exact order:

1. **polyfills.js** — platform normalization (fixes WebKitGTK quirks, adds platform CSS classes, polyfills missing APIs like `structuredClone`)
2. **lightshell.js** — the API client library (creates `window.lightshell` with all the native API bindings)
3. **Your HTML/JS** — your application code

This order is critical. Polyfills must patch the environment before the API client runs. The API client must exist before your code calls `lightshell.*`.

Both `polyfills.js` and `lightshell.js` are embedded in the Go binary at compile time using `embed.FS`. They add less than 8KB to the binary.

## IPC: How JS Talks to Go

When you call `lightshell.fs.readFile('/tmp/test.txt')`, here is what happens:

1. **Your JS** calls the client library function
2. **lightshell.js** creates a JSON message with a unique ID, method name, and parameters
3. The message is sent to Go via `window.webkit.messageHandlers.lightshell.postMessage()`
4. **Go receives the message** through the webview's message handler callback
5. **The IPC router** dispatches to the correct handler (e.g., the `fs.readFile` handler)
6. **The handler** executes the native operation (reads the file using `os.ReadFile`)
7. **Go sends the response** back by calling `webview.Eval("__lightshell_receive(...)")` which executes JavaScript in the webview
8. **lightshell.js** receives the response, matches it to the pending Promise by ID, and resolves it

The full round-trip takes less than 5ms for local operations.

See the [IPC Protocol](/concepts/ipc/) page for the message format details.

## Asset Embedding

When you run `lightshell build`, your `src/` files are embedded into the Go binary using Go's `embed.FS` directive. At runtime, the binary serves these files from memory — there are no external files to ship alongside the executable.

The build process:
1. Reads `lightshell.json` for configuration
2. Copies your `src/` directory into a staging area
3. Compiles a Go binary that embeds the staged assets
4. Wraps the binary in a platform-specific package (`.app` bundle on macOS, AppImage on Linux)

## The Binary

A LightShell binary contains:

| Component | Size |
|-----------|------|
| Go runtime and standard library | ~1.5MB |
| Webview bindings (cgo bridge) | ~200KB |
| LightShell runtime (IPC, API handlers) | ~300KB |
| polyfills.js + lightshell.js + normalize.css | ~8KB |
| Your HTML/CSS/JS | varies |

Total for a typical app: **~2.8MB**.

The binary has no external dependencies. On macOS, it links against system frameworks (Cocoa, WebKit) which are always present. On Linux, it requires WebKitGTK 2.40+ which is available in most modern distributions.

## Development Mode vs Production

| Aspect | `lightshell dev` | `lightshell build` |
|--------|-------------------|--------------------|
| Asset loading | HTTP server (localhost) | Embedded in binary |
| Hot reload | Yes (watches `src/`) | No |
| DevTools | Enabled | Disabled |
| IPC transport | Same (Unix domain socket) | Same |
| Window behavior | Same | Same |

In dev mode, a local HTTP server serves your files and the webview loads from `http://localhost:{port}`. File changes trigger a reload signal through IPC. In production, assets are served from the embedded filesystem.
