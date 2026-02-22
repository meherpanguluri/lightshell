package mcp

// registerResources registers all MCP resources exposed by the server.
func (s *Server) registerResources() {
	s.registerResource(Resource{
		URI:         "lightshell://api-reference",
		Name:        "LightShell API Reference",
		Description: "Complete API reference for the LightShell desktop framework, including all namespaces, methods, and examples.",
		MimeType:    "text/plain",
		Handler: func() (string, error) {
			if s.apiDocs != "" {
				return s.apiDocs, nil
			}
			return defaultAPIDocs, nil
		},
	})

	s.registerResource(Resource{
		URI:         "lightshell://errors",
		Name:        "LightShell Error Catalog",
		Description: "Catalog of LightShell error codes and their meanings, with troubleshooting guidance.",
		MimeType:    "text/plain",
		Handler: func() (string, error) {
			return errorCatalog, nil
		},
	})
}

const defaultAPIDocs = `# LightShell API Reference

LightShell is a desktop application framework that lets you build native apps with HTML, CSS, and JavaScript. The Go backend provides system APIs accessible via window.lightshell in JavaScript.

## Namespaces

### lightshell.window
Control the application window.
- setTitle(title: string) — set window title
- setSize(width: number, height: number) — set window dimensions
- getSize() — returns {width, height}
- setPosition(x: number, y: number) — set window position
- getPosition() — returns {x, y}
- minimize() — minimize window
- maximize() — maximize/restore window
- fullscreen() — enter fullscreen
- restore() — restore from minimize/maximize/fullscreen
- close() — close the window

### lightshell.fs
File system operations. Paths support $APP_DATA, $HOME, $TEMP, $DOWNLOADS, $DESKTOP variables.
- readFile(path: string, options?: {encoding?: string}) — read file contents
- writeFile(path: string, data: string, options?: {encoding?: string}) — write file
- readDir(path: string) — list directory entries
- exists(path: string) — check if path exists
- stat(path: string) — get file metadata
- mkdir(path: string, options?: {recursive?: boolean}) — create directory
- remove(path: string, options?: {recursive?: boolean}) — delete file/directory
- watch(path: string, callback: function) — watch for file changes

### lightshell.dialog
Native system dialogs.
- open(options?: {title?, filters?, multiple?, directory?}) — file open dialog
- save(options?: {title?, defaultPath?, filters?}) — file save dialog
- message(text: string, options?: {title?, type?}) — message box
- confirm(text: string, options?: {title?}) — confirmation dialog
- prompt(text: string, options?: {title?, defaultValue?}) — input dialog

### lightshell.clipboard
System clipboard.
- read() — read clipboard text
- write(text: string) — write text to clipboard

### lightshell.shell
Shell integration.
- open(url: string) — open URL in default browser / file in default app

### lightshell.notify
Desktop notifications.
- send(options: {title: string, body?: string}) — show a notification

### lightshell.tray
System tray icon.
- set(options: {icon: string, tooltip?: string, menu?: MenuItem[]}) — set tray
- remove() — remove tray icon
- onClick(callback: function) — handle tray click

### lightshell.menu
Application menu bar.
- set(template: MenuItem[]) — set the application menu

### lightshell.system
System information.
- platform() — returns "darwin" or "linux"
- arch() — returns CPU architecture
- homeDir() — returns home directory path
- tempDir() — returns temp directory path
- hostname() — returns system hostname

### lightshell.app
Application lifecycle.
- quit() — quit the application
- version() — returns app version from lightshell.json
- dataDir() — returns app data directory path
- onOpenUrl(callback: function) — handle deep link URLs

### lightshell.store
Persistent key-value storage.
- get(key: string) — get a value
- set(key: string, value: any) — set a value
- delete(key: string) — delete a key
- has(key: string) — check if key exists
- keys(prefix?: string) — list keys by prefix
- clear() — delete all keys

### lightshell.http
CORS-free HTTP client (requests go through Go backend).
- fetch(url: string, options?: {method?, headers?, body?, timeout?}) — make HTTP request
- download(url: string, options?: {saveTo, headers?, onProgress?}) — download file to disk

### lightshell.process
System command execution (scoped by permissions).
- exec(cmd: string, args?: string[], options?: {cwd?, env?, timeout?}) — run a command

### lightshell.shortcuts
Global keyboard shortcuts.
- register(accelerator: string, callback: function) — register global hotkey
- unregister(accelerator: string) — unregister hotkey
- unregisterAll() — unregister all hotkeys
- isRegistered(accelerator: string) — check if registered

### lightshell.updater
Auto-update support.
- check() — check for updates
- install() — download and install update
- checkAndInstall() — check and install in one call
- onProgress(callback: function) — listen to download progress

## Configuration

Apps are configured via lightshell.json in the project root:
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "My App",
    "width": 800,
    "height": 600,
    "minWidth": 400,
    "minHeight": 300,
    "resizable": true,
    "frameless": false
  },
  "permissions": { ... },
  "security": { "csp": "..." },
  "updater": { "enabled": true, "endpoint": "...", "interval": "24h" }
}

## CLI Commands
- lightshell init [name] — create a new project
- lightshell dev — run in dev mode with hot reload
- lightshell build — build for production
- lightshell doctor — check for cross-platform issues
- lightshell mcp — run MCP server for AI integration
`

const errorCatalog = `# LightShell Error Catalog

## Permission Errors

### fs.readFile: Permission denied
Cause: Attempted to read a file outside the allowed read paths.
Fix: Add the path pattern to permissions.fs.read in lightshell.json.
Example:
  "permissions": { "fs": { "read": ["$APP_DATA/**", "$HOME/Documents/**"] } }

### fs.writeFile: Permission denied
Cause: Attempted to write a file outside the allowed write paths.
Fix: Add the path pattern to permissions.fs.write in lightshell.json.

### process.exec: Permission denied
Cause: Attempted to run a command not listed in permissions.process.exec.
Fix: Add the command to the allowed list:
  "permissions": { "process": { "exec": [{"cmd": "git", "args": ["*"]}] } }

### http.fetch: Permission denied
Cause: Attempted to make a request to a URL not in the allowed list.
Fix: Add the URL pattern to permissions.http.allow:
  "permissions": { "http": { "allow": ["https://api.example.com/**"] } }

## Path Errors

### Path traversal blocked
Cause: A path containing .. was resolved to a location outside allowed directories.
Fix: Use absolute paths or paths within allowed directories. Symlinks pointing outside allowed directories are also blocked.

### Invalid path variable
Cause: An unrecognized path variable like $UNKNOWN was used.
Fix: Use one of: $APP_DATA, $HOME, $TEMP, $RESOURCE, $DOWNLOADS, $DESKTOP

## Runtime Errors

### IPC timeout
Cause: An API call took longer than the timeout period (default 30s).
Fix: Increase timeout in options or check for blocking operations.

### Window not ready
Cause: An API call was made before the window finished loading.
Fix: Wait for the DOMContentLoaded event before calling lightshell APIs.

### Store error: database locked
Cause: Multiple concurrent write operations to the store.
Fix: Serialize write operations or use await to ensure sequential execution.

## Build Errors

### Entry file not found
Cause: The entry path in lightshell.json points to a file that doesn't exist.
Fix: Check the "entry" field in lightshell.json and ensure the file exists.

### Build target not supported
Cause: Attempted to build for an unsupported target on the current OS.
Fix: DMG and .app targets require macOS. DEB/RPM/AppImage targets require Linux.

### Code signing failed
Cause: The signing identity is invalid or not found in the keychain.
Fix: Check build.mac.identity in lightshell.json. Run "security find-identity -v" to list available identities.

## Update Errors

### Update check failed
Cause: Could not reach the update endpoint or the response was invalid.
Fix: Verify the updater.endpoint URL in lightshell.json is correct and reachable.

### SHA256 mismatch
Cause: The downloaded update file hash doesn't match the manifest.
Fix: This is a security error — the update file may be corrupted or tampered with. Re-upload the release with the correct hash.

### HTTP endpoint not allowed in production
Cause: The updater endpoint uses HTTP instead of HTTPS in a production build.
Fix: Use an HTTPS URL for the updater endpoint. HTTP is only allowed in dev mode.
`
