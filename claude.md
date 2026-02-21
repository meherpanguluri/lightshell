---

## Agent 7: Security, Packaging & Platform Services

**Goal:** Harden the framework with a permission system, expand packaging to production-grade installers, add an auto-updater, and ship the APIs that real apps need but Agents 1-4 didn't cover — key-value storage, CORS-free HTTP, scoped process execution, and global shortcuts.

**Why this agent exists:** Agents 1-5 build a working framework. Agent 7 makes it shippable. Without this layer, LightShell is a toy — a cool demo that can't survive contact with real users distributing real apps. This is the difference between "weekend project" and "I'd actually use this."

### Responsibilities

#### 1. Permission System

A flat, single-section permission model in `lightshell.json`. Inspired by Tauri's capabilities but without the complexity of per-window capability files.

**Design principle:** Permissive by default (everything allowed), restrictive on opt-in. When a developer adds a `permissions` key, the framework switches to allowlist mode.

```json
// lightshell.json — no permissions key = everything allowed (prototyping mode)
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html"
}
```

```json
// lightshell.json — with permissions = allowlist mode (production mode)
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/**"],
      "write": ["$APP_DATA/**"]
    },
    "dialog": true,
    "clipboard": true,
    "notifications": true,
    "shell": {
      "open": true
    },
    "process": {
      "exec": [
        { "cmd": "git", "args": ["status", "log", "diff"] },
        { "cmd": "ls" }
      ]
    },
    "http": {
      "allow": ["https://api.myapp.com/*", "https://api.github.com/*"],
      "deny": []
    },
    "shortcuts": true,
    "store": true,
    "tray": true,
    "menu": true,
    "system": true,
    "updater": true
  }
}
```

**Path variables** (resolved at runtime):

| Variable | macOS | Linux |
|----------|-------|-------|
| `$APP_DATA` | `~/Library/Application Support/{app-name}` | `~/.config/{app-name}` |
| `$HOME` | `/Users/{user}` | `/home/{user}` |
| `$TEMP` | `/tmp` | `/tmp` |
| `$RESOURCE` | `{app-bundle}/Contents/Resources` | `{appimage-mount}/resources` |
| `$DOWNLOADS` | `~/Downloads` | `~/Downloads` |
| `$DESKTOP` | `~/Desktop` | `~/Desktop` |

**Glob patterns:** `*` matches within directory, `**` matches recursively.

**Enforcement layer** (`internal/security/permissions.go`):

```go
type PermissionEngine struct {
    mode     string // "permissive" | "restricted"
    rules    PermissionConfig
    appPaths ResolvedPaths
}

// Called by EVERY IPC handler before executing
func (p *PermissionEngine) Check(method string, params json.RawMessage) error {
    if p.mode == "permissive" {
        return nil // everything allowed
    }
    switch {
    case strings.HasPrefix(method, "fs."):
        return p.checkFS(method, params)
    case strings.HasPrefix(method, "process."):
        return p.checkProcess(method, params)
    case strings.HasPrefix(method, "http."):
        return p.checkHTTP(method, params)
    default:
        return p.checkSimple(method)
    }
}
```

**Error on permission denial** (AI-friendly):

```
LightShell Error [fs.readFile]: Permission denied
  → Attempted to read: /etc/passwd
  → Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  → To allow this path, update permissions.fs.read in lightshell.json
  → Docs: https://lightshell.sh/docs/api/permissions#fs
```

#### 2. Security Hardening

**Content Security Policy** — auto-injected into every HTML page:

```go
// internal/security/csp.go

// Default CSP for production builds
const defaultCSP = `default-src 'self' lightshell:; ` +
    `script-src 'self'; ` +
    `style-src 'self' 'unsafe-inline'; ` +   // inline styles commonly needed
    `img-src 'self' data: blob:; ` +
    `font-src 'self' data:; ` +
    `connect-src 'self'; ` +
    `object-src 'none'; ` +
    `frame-ancestors 'none'`

// Dev mode CSP is more relaxed
const devCSP = `default-src 'self' 'unsafe-inline' 'unsafe-eval' lightshell: http://localhost:*`
```

CSP is injected as a `<meta>` tag at the top of the document, before any user scripts execute. Configurable in `lightshell.json`:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.example.com"
  }
}
```

**Path traversal protection** — always on, cannot be disabled:

```go
// internal/security/paths.go

// Canonicalize and validate every path before any fs operation
func ValidatePath(requested string, allowedPatterns []string) error {
    // 1. Resolve to absolute path
    abs, err := filepath.Abs(requested)
    if err != nil {
        return fmt.Errorf("invalid path: %s", requested)
    }

    // 2. Resolve symlinks to real path
    real, err := filepath.EvalSymlinks(abs)
    if err != nil {
        return fmt.Errorf("cannot resolve path: %s", requested)
    }

    // 3. Check for traversal (real path must be within allowed dirs)
    for _, pattern := range allowedPatterns {
        resolved := resolvePathVariable(pattern)
        if matchGlob(real, resolved) {
            return nil
        }
    }

    return &PermissionError{
        Method:  "fs",
        Path:    requested,
        Allowed: allowedPatterns,
    }
}
```

**IPC channel security:**
- Unix domain socket with `0600` permissions (owner-only access). Neutralinojs's localhost WebSocket = any process can connect. Ours = only the app process can connect.
- Socket path includes PID and random token: `/tmp/lightshell-{pid}-{random}.sock`
- Socket file deleted on shutdown (deferred cleanup + signal handler)

**DevTools control:**
- Enabled by default in `lightshell dev`
- Disabled by default in `lightshell build` output
- Override: `lightshell build --devtools` for debug builds
- No right-click context menu in production builds unless explicitly enabled

#### 3. Extended Packaging

Expand Agent 3's packaging to cover production distribution formats.

**macOS — DMG with drag-to-install:**

```go
// internal/packaging/darwin/dmg.go

func BuildDMG(appPath string, config DMGConfig) (string, error) {
    // 1. Create temporary DMG from .app bundle
    // 2. Set volume name, icon, background image
    // 3. Create symlink to /Applications inside DMG
    // 4. Set window size and icon positions via AppleScript/DS_Store
    // 5. Convert to compressed read-only DMG
    // Uses: hdiutil create, hdiutil convert
    // Output: dist/MyApp-1.0.0.dmg
}
```

The DMG includes:
- The .app bundle
- Shortcut to /Applications
- Clean layout (app on left, Applications on right)
- Volume icon matching app icon

**macOS — Code signing:**

```json
// lightshell.json
{
  "build": {
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)",
      "entitlements": {
        "com.apple.security.network.client": true,
        "com.apple.security.files.user-selected.read-write": true
      }
    }
  }
}
```

```go
// internal/packaging/darwin/sign.go

func SignApp(appPath string, identity string, entitlements map[string]bool) error {
    // 1. Generate entitlements.plist from config
    // 2. codesign --deep --force --options runtime --sign {identity} {app}
    // 3. Verify: codesign --verify --deep --strict {app}
}

func Notarize(dmgPath string, appleID string, teamID string) error {
    // 1. xcrun notarytool submit {dmg} --apple-id {id} --team-id {team} --wait
    // 2. xcrun stapler staple {dmg}
    // Returns: notarization UUID and status
}
```

CLI: `lightshell build --target dmg --sign` / `lightshell build --target dmg --sign --notarize`

**Linux — .deb package:**

```go
// internal/packaging/linux/deb.go

func BuildDeb(binaryPath string, config DebConfig) (string, error) {
    // Build structure:
    // staging/
    //   DEBIAN/
    //     control      (package metadata)
    //     postinst     (post-install script — optional)
    //   usr/
    //     bin/
    //       myapp      (binary)
    //     share/
    //       applications/
    //         myapp.desktop    (.desktop entry)
    //       icons/hicolor/
    //         256x256/apps/
    //           myapp.png      (icon)
    //
    // Run: dpkg-deb --build staging dist/myapp_1.0.0_amd64.deb

    return debPath, nil
}
```

Control file template:

```
Package: {{.Name}}
Version: {{.Version}}
Section: utils
Priority: optional
Architecture: {{.Arch}}
Depends: libwebkit2gtk-4.1-0 (>= 2.38), libgtk-3-0
Maintainer: {{.Author}} <{{.Email}}>
Description: {{.Description}}
Homepage: {{.Homepage}}
```

**Linux — .rpm package:**

```go
// internal/packaging/linux/rpm.go

func BuildRPM(binaryPath string, config RPMConfig) (string, error) {
    // Uses rpmbuild or fpm if available
    // Spec file:
    //   Name, Version, Release, Summary
    //   Requires: webkit2gtk4.1, gtk3
    //   %files: /usr/bin/myapp, .desktop, icons
    // Output: dist/myapp-1.0.0-1.x86_64.rpm
}
```

**Expanded CLI:**

```
lightshell build                        # .app (macOS) or AppImage (Linux) — default
lightshell build --target dmg           # .dmg with drag-to-install
lightshell build --target dmg --sign    # signed .dmg
lightshell build --target deb           # Debian package
lightshell build --target rpm           # RPM package
lightshell build --target all           # all formats for current OS
lightshell build --devtools             # include DevTools in production build
```

#### 4. Auto-Updater

A simple, secure update mechanism. No Sparkle, no electron-updater complexity. Just a JSON manifest, SHA256 verification, and binary replacement.

**Developer-side** — host a JSON file:

```json
// https://releases.myapp.com/latest.json
{
  "version": "1.2.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-07-15T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-darwin-arm64.tar.gz",
      "sha256": "a1b2c3d4..."
    },
    "darwin-x64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-darwin-x64.tar.gz",
      "sha256": "e5f6g7h8..."
    },
    "linux-x64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-linux-x64.tar.gz",
      "sha256": "i9j0k1l2..."
    }
  }
}
```

**Config** in `lightshell.json`:

```json
{
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.myapp.com/latest.json",
    "interval": "24h"
  }
}
```

**JS API:**

```js
// Check for updates (returns null if no update)
const update = await lightshell.updater.check()
// { version: "1.2.0", notes: "Bug fixes...", currentVersion: "1.1.0" }

// Download and install (replaces binary, prompts restart)
await lightshell.updater.install()

// Progress events
lightshell.updater.onProgress((p) => {
  console.log(`${p.percent}% — ${p.bytesDownloaded}/${p.totalBytes}`)
})

// Or one-liner for simple apps
await lightshell.updater.checkAndInstall()
```

**Go implementation** (`internal/api/updater.go`):

```go
func handleUpdaterCheck(params json.RawMessage) (any, error) {
    // 1. GET the endpoint URL
    // 2. Parse JSON manifest
    // 3. Compare manifest version to current app version (semver)
    // 4. Find matching platform key: runtime.GOOS + "-" + runtime.GOARCH
    // 5. Return update info or null
}

func handleUpdaterInstall(params json.RawMessage) (any, error) {
    // 1. Download the archive to temp directory
    // 2. Verify SHA256 hash matches manifest
    // 3. Extract archive
    // 4. Replace current binary:
    //    macOS: replace .app/Contents/MacOS/binary
    //    Linux: replace AppImage or binary in /usr/bin
    // 5. Trigger app restart (optional, configurable)
}
```

**Background check** (optional, runs on startup if interval has elapsed):

```go
func StartBackgroundUpdater(config UpdaterConfig) {
    // On app startup, check if interval has passed since last check
    // If update available, emit "updater.available" event to JS
    // Never auto-installs without user/developer consent
}
```

**Security:** SHA256 verification is mandatory. If hash doesn't match, the update is rejected and an error emitted. The endpoint MUST be HTTPS in production builds (HTTP allowed in dev for testing).

#### 5. Additional APIs

APIs that real apps need, designed with the same AI-native principles as the existing ones.

##### 5a. Key-Value Store (`lightshell.store`)

Persistent key-value storage backed by SQLite (embedded via CGO or pure-Go port). The user never sees SQL — it's just `get`/`set`/`delete`.

**JS API:**

```js
// Simple key-value
await lightshell.store.set('user.name', 'Alice')
const name = await lightshell.store.get('user.name')    // "Alice"
await lightshell.store.delete('user.name')

// JSON values (automatically serialized/deserialized)
await lightshell.store.set('settings', { theme: 'dark', fontSize: 14 })
const settings = await lightshell.store.get('settings')  // { theme: 'dark', fontSize: 14 }

// List keys by prefix
const keys = await lightshell.store.keys('user.*')       // ["user.name", "user.email"]

// Check existence
const exists = await lightshell.store.has('user.name')    // true

// Clear all
await lightshell.store.clear()
```

**Go implementation** (`internal/api/store.go`):

```go
// Store backed by a single SQLite file at $APP_DATA/store.db
// Table: CREATE TABLE IF NOT EXISTS kv (key TEXT PRIMARY KEY, value TEXT, updated_at INTEGER)
// Values stored as JSON strings

func handleStoreSet(params json.RawMessage) (any, error) {
    var p struct {
        Key   string          `json:"key"`
        Value json.RawMessage `json:"value"`
    }
    json.Unmarshal(params, &p)
    // UPSERT into kv table
    return nil, nil
}

func handleStoreGet(params json.RawMessage) (any, error) {
    var p struct { Key string `json:"key"` }
    json.Unmarshal(params, &p)
    // SELECT value FROM kv WHERE key = ?
    // Return parsed JSON value, or null if not found
}
```

**Why this matters for AI:** An AI generating a todo app, a note-taking app, a settings panel — all need persistence. `lightshell.store` lets the AI write `await lightshell.store.set('todos', todos)` instead of figuring out file paths, serialization, atomicity, and error handling. Covers 80% of persistence needs in one line.

**SQLite choice:** Use `modernc.org/sqlite` (pure Go, no CGO required) to avoid complicating the build. ~2MB addition to binary but eliminates CGO dependency chain for SQLite. Alternatively, use a simpler bolt-style key-value store (`bbolt`) for even smaller size (~200KB) but lose SQL query capability for future expansion.

Recommend: **bbolt for v1** (smallest footprint, key-value is all we need), migrate to SQLite if structured queries ever needed.

##### 5b. CORS-Free HTTP Client (`lightshell.http`)

Desktop apps need to call APIs without CORS restrictions. The webview's `fetch()` is subject to CORS — the Go backend's `net/http` is not. This is a genuine killer feature.

**JS API:**

```js
const response = await lightshell.http.fetch('https://api.github.com/user', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ghp_xxxxx',
    'Accept': 'application/json'
  }
})
// response = { status: 200, headers: {...}, body: "..." }

// POST with body
const result = await lightshell.http.fetch('https://api.example.com/data', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ query: 'test' })
})

// The response body is always a string. Parse as needed:
const data = JSON.parse(result.body)
```

**Go implementation** (`internal/api/http.go`):

```go
func handleHTTPFetch(params json.RawMessage) (any, error) {
    var p struct {
        URL     string            `json:"url"`
        Method  string            `json:"method"`
        Headers map[string]string `json:"headers"`
        Body    string            `json:"body"`
        Timeout int               `json:"timeout"` // ms, default 30000
    }
    json.Unmarshal(params, &p)

    // 1. Check permission: is this URL in the allowed list?
    // 2. Build http.Request
    // 3. Execute with timeout
    // 4. Read response body
    // 5. Return { status, headers, body }

    // Important: respect permissions.http.allow/deny patterns
    // Default (no permissions key): all URLs allowed
    // With permissions: only matching URLs allowed
}
```

**Why not just proxy through the Go backend?** This IS the proxy. But it's exposed as a clean API that feels like `fetch()` rather than asking users to set up a proxy server. The Go backend makes the real HTTP call and returns the result via IPC.

**Limitations to document:**
- Response body is returned as string (not streaming). For large downloads, use `lightshell.http.download()` (writes directly to disk).
- Binary responses are base64-encoded.
- WebSocket support: out of scope for v1.

```js
// For large files — downloads directly to disk without loading into memory
await lightshell.http.download('https://example.com/large-file.zip', {
  saveTo: '$DOWNLOADS/file.zip',
  onProgress: (p) => console.log(`${p.percent}%`)
})
```

##### 5c. Scoped Process Execution (`lightshell.process`)

Run system commands from JavaScript. Unlike Neutralinojs (no scoping at all), commands must be declared in the permissions config when in restricted mode.

**JS API:**

```js
const result = await lightshell.process.exec('git', ['status'])
// result = { stdout: "On branch main...", stderr: "", code: 0 }

const result2 = await lightshell.process.exec('python3', ['script.py'], {
  cwd: '/path/to/project',
  env: { PYTHONPATH: '/custom/path' },
  timeout: 10000  // ms
})
```

**Go implementation** (`internal/api/process.go`):

```go
func handleProcessExec(params json.RawMessage) (any, error) {
    var p struct {
        Cmd     string            `json:"cmd"`
        Args    []string          `json:"args"`
        Cwd     string            `json:"cwd"`
        Env     map[string]string `json:"env"`
        Timeout int               `json:"timeout"`
    }
    json.Unmarshal(params, &p)

    // 1. Check permission: is this command + these args allowed?
    //    Permission model matches Tauri's shell scoping:
    //    { "cmd": "git", "args": ["status", "log", "diff"] }
    //    means only git status, git log, git diff are allowed
    //    If args is omitted or ["*"], any args are allowed for that command
    //
    // 2. Resolve command path (no shell expansion — exec directly, not via sh -c)
    //    This prevents shell injection attacks
    //
    // 3. Execute with timeout via context.WithTimeout
    // 4. Capture stdout, stderr, exit code
    // 5. Return result

    cmd := exec.CommandContext(ctx, p.Cmd, p.Args...)
    // ...
}
```

**Security notes:**
- Commands are executed directly (`exec.Command`), NEVER via shell (`sh -c`). This prevents shell injection.
- In restricted mode, command+args must match an entry in `permissions.process.exec`
- PATH is restricted to standard system paths, not user-modifiable
- No `eval()` or arbitrary command string execution

##### 5d. Global Keyboard Shortcuts (`lightshell.shortcuts`)

Register keyboard shortcuts that work even when the app window isn't focused (global hotkeys). Essential for productivity tools, clipboard managers, screenshot tools.

**JS API:**

```js
// Register a global shortcut
lightshell.shortcuts.register('CommandOrControl+Shift+P', () => {
  console.log('Command palette triggered!')
})

// Unregister
lightshell.shortcuts.unregister('CommandOrControl+Shift+P')

// Unregister all
lightshell.shortcuts.unregisterAll()

// Check if registered
const registered = await lightshell.shortcuts.isRegistered('CommandOrControl+Shift+P')
```

**Modifier key naming** (cross-platform):

| Key String | macOS | Linux |
|-----------|-------|-------|
| `CommandOrControl` | ⌘ Cmd | Ctrl |
| `Command` | ⌘ Cmd | (ignored on Linux) |
| `Control` | Ctrl | Ctrl |
| `Alt` | ⌥ Option | Alt |
| `Shift` | ⇧ Shift | Shift |
| `Super` | ⌘ Cmd | Super/Win |

**Go implementation:**
- macOS: `CGEventTapCreate` + `NSEvent addGlobalMonitorForEvents`
- Linux: X11 `XGrabKey` or libkeybinder (GTK-based global hotkeys)

##### 5e. Deep Links / Protocol Handler (`lightshell.app.onOpenUrl`)

Register a custom URL protocol so other apps and websites can launch your app.

**Config:**

```json
// lightshell.json
{
  "protocol": "myapp"
}
```

**JS API:**

```js
// When someone opens myapp://settings/theme in their browser,
// your app launches (or focuses) and receives the URL
lightshell.app.onOpenUrl((url) => {
  console.log(url) // "myapp://settings/theme"
  // route to appropriate view
})
```

**Implementation:**
- macOS: Register `CFBundleURLTypes` in Info.plist, handle `application:openURLs:` in app delegate
- Linux: Register as handler in `.desktop` file (`MimeType=x-scheme-handler/myapp`), handle via command-line argument

#### 6. Enhanced Polyfills (Refined)

Replace Agent 4's current polyfill list with the researched, accurate set. The polyfills.js file should be < 2KB and cover only the gaps that are (a) actually broken and (b) cheaply fixable.

**Must polyfill** (< 2KB total):

```js
// structuredClone — missing on WebKitGTK < 2.40
if (typeof structuredClone === 'undefined') {
  window.structuredClone = (obj, options) => {
    if (options?.transfer?.length) {
      throw new DOMException('Transfer not supported in polyfill', 'DataCloneError')
    }
    return JSON.parse(JSON.stringify(obj))
  }
}

// Array.prototype.group / groupBy — missing on WebKitGTK < 2.44
if (!Array.prototype.group) {
  Array.prototype.group = function(fn) {
    return this.reduce((acc, item, i) => {
      const key = fn(item, i, this)
      ;(acc[key] ??= []).push(item)
      return acc
    }, {})
  }
}

// Promise.withResolvers — missing on WebKitGTK < 2.44
if (!Promise.withResolvers) {
  Promise.withResolvers = function() {
    let resolve, reject
    const promise = new Promise((res, rej) => { resolve = res; reject = rej })
    return { promise, resolve, reject }
  }
}

// Set methods (union, intersection, difference) — missing on WebKitGTK < 2.44
if (!Set.prototype.union) {
  Set.prototype.union = function(other) {
    const result = new Set(this)
    for (const item of other) result.add(item)
    return result
  }
  Set.prototype.intersection = function(other) {
    const result = new Set()
    for (const item of this) if (other.has(item)) result.add(item)
    return result
  }
  Set.prototype.difference = function(other) {
    const result = new Set()
    for (const item of this) if (!other.has(item)) result.add(item)
    return result
  }
  Set.prototype.symmetricDifference = function(other) {
    const result = new Set(this)
    for (const item of other) {
      if (result.has(item)) result.delete(item)
      else result.add(item)
    }
    return result
  }
  Set.prototype.isSubsetOf = function(other) {
    for (const item of this) if (!other.has(item)) return false
    return true
  }
  Set.prototype.isSupersetOf = function(other) {
    for (const item of other) if (!this.has(item)) return false
    return true
  }
  Set.prototype.isDisjointFrom = function(other) {
    for (const item of this) if (other.has(item)) return false
    return true
  }
}

// Object.groupBy — missing on WebKitGTK < 2.44
if (!Object.groupBy) {
  Object.groupBy = function(iterable, fn) {
    const result = Object.create(null)
    let i = 0
    for (const item of iterable) {
      const key = fn(item, i++)
      ;(result[key] ??= []).push(item)
    }
    return result
  }
}
```

**Must warn** (scanner catches, no polyfill):

| API | Reason | Scanner Message |
|-----|--------|----------------|
| `Intl.Segmenter` | Polyfill is ~200KB, too large | "Intl.Segmenter unavailable on Linux. Use Intl.BreakIterator or split manually." |
| `Navigation API` | Not in ANY webview | "Use History API or lightshell.window for navigation." |
| `View Transitions API` | Missing on WebKitGTK | "View transitions not supported on Linux. Implement with CSS animations." |
| `:has()` selector | Missing on WebKitGTK < 2.42 | "Use JS to toggle classes instead of :has() for Linux support." |
| CSS Nesting | Missing on WebKitGTK < 2.44 | "Flatten nested CSS rules for Linux compatibility." |
| `@container` queries | Partial on older WebKitGTK | "Container queries have limited support on older Linux. Test thoroughly." |
| `File System Access API` | Not in any webview | "Use lightshell.dialog.open() and lightshell.fs instead." |
| `Web USB/Bluetooth/Serial` | Not in any webview | "Hardware APIs not available in webview context." |

**Must normalize** (CSS injected always):

```css
/* Form element reset — consistent across platforms */
input, select, textarea, button {
  -webkit-appearance: none;
  appearance: none;
  font-family: inherit;
  font-size: inherit;
}

/* Scrollbar normalization */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(128, 128, 128, 0.4); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: rgba(128, 128, 128, 0.6); }

/* Prevent layout shift from scrollbar appearance */
html { scrollbar-gutter: stable; }

/* System font stack — explicit, not relying on system-ui resolution */
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif, "Apple Color Emoji", "Noto Color Emoji";
}

/* Focus ring normalization */
:focus-visible {
  outline: 2px solid highlight;
  outline-offset: 2px;
}
:focus:not(:focus-visible) {
  outline: none;
}
```

### Repository Additions

```
lightshell/
├── internal/
│   ├── security/
│   │   ├── permissions.go      # Permission engine (Agent 7)
│   │   ├── permissions_test.go # Permission tests (Agent 7)
│   │   ├── csp.go              # CSP injection (Agent 7)
│   │   └── paths.go            # Path validation & traversal protection (Agent 7)
│   │
│   ├── api/
│   │   ├── store.go            # Key-value store API (Agent 7)
│   │   ├── http.go             # CORS-free HTTP client API (Agent 7)
│   │   ├── process.go          # Scoped process execution (Agent 7)
│   │   ├── shortcuts.go        # Global keyboard shortcuts (Agent 7)
│   │   ├── shortcuts_darwin.go # macOS hotkey impl (Agent 7)
│   │   ├── shortcuts_linux.go  # Linux hotkey impl (Agent 7)
│   │   ├── updater.go          # Auto-updater API (Agent 7)
│   │   └── protocol.go         # Deep link / URL protocol handler (Agent 7)
│   │
│   └── store/
│       ├── store.go            # bbolt-backed key-value store (Agent 7)
│       └── store_test.go       # Store tests (Agent 7)
│
├── packaging/
│   ├── darwin/
│   │   ├── dmg.go              # DMG builder (Agent 7)
│   │   ├── sign.go             # Code signing (Agent 7)
│   │   └── notarize.go         # Notarization (Agent 7)
│   └── linux/
│       ├── deb.go              # .deb package builder (Agent 7)
│       └── rpm.go              # .rpm package builder (Agent 7)
│
└── tests/
    ├── security_test.go        # Permission & path validation tests (Agent 7)
    ├── store_test.go           # Store API tests (Agent 7)
    └── updater_test.go         # Updater tests (Agent 7)
```

### Updated Client Library Additions

Agent 7 extends `window.lightshell` with new namespaces. Agent 2 owns the client library file, but Agent 7 provides these additions to be merged:

```js
// Additions to client/lightshell.js

// Key-Value Store
store: {
  get:    (key)        => call('store.get', { key }),
  set:    (key, value) => call('store.set', { key, value }),
  delete: (key)        => call('store.delete', { key }),
  has:    (key)        => call('store.has', { key }),
  keys:   (prefix)     => call('store.keys', { prefix: prefix || '' }),
  clear:  ()           => call('store.clear'),
},

// CORS-Free HTTP
http: {
  fetch:    (url, opts) => call('http.fetch', { url, ...opts }),
  download: (url, opts) => {
    call('http.download', { url, ...opts })
    return on('http.download.progress', opts?.onProgress)
  },
},

// Process Execution
process: {
  exec: (cmd, args, opts) => call('process.exec', { cmd, args: args || [], ...opts }),
},

// Global Shortcuts
shortcuts: {
  register:     (combo, cb) => { call('shortcuts.register', { combo }); return on('shortcut.' + combo, cb) },
  unregister:   (combo)     => call('shortcuts.unregister', { combo }),
  unregisterAll: ()         => call('shortcuts.unregisterAll'),
  isRegistered: (combo)     => call('shortcuts.isRegistered', { combo }),
},

// Auto-Updater
updater: {
  check:           ()   => call('updater.check'),
  install:         ()   => call('updater.install'),
  checkAndInstall: ()   => call('updater.checkAndInstall'),
  onProgress:      (cb) => on('updater.progress', cb),
},
```

### Updated TypeScript Definitions

```typescript
// Additions to client/lightshell.d.ts

interface LightShellStore {
  /** Get a value by key. Returns null if key doesn't exist. */
  get<T = any>(key: string): Promise<T | null>

  /** Set a key-value pair. Values are JSON-serialized. */
  set(key: string, value: any): Promise<void>

  /** Delete a key. No error if key doesn't exist. */
  delete(key: string): Promise<void>

  /** Check if a key exists. */
  has(key: string): Promise<boolean>

  /** List all keys matching a prefix. Use '*' to list all. */
  keys(prefix?: string): Promise<string[]>

  /** Delete all keys. */
  clear(): Promise<void>
}

interface LightShellHTTP {
  /** Make an HTTP request (CORS-free, goes through Go backend). */
  fetch(url: string, options?: {
    method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD'
    headers?: Record<string, string>
    body?: string
    timeout?: number  // ms, default 30000
  }): Promise<{
    status: number
    headers: Record<string, string>
    body: string
  }>

  /** Download a file directly to disk. */
  download(url: string, options?: {
    saveTo: string                        // path, supports $DOWNLOADS etc
    headers?: Record<string, string>
    onProgress?: (progress: {
      percent: number
      bytesDownloaded: number
      totalBytes: number
    }) => void
  }): Promise<{ path: string; size: number }>
}

interface LightShellProcess {
  /** Execute a system command. Respects permissions.process.exec scoping. */
  exec(cmd: string, args?: string[], options?: {
    cwd?: string
    env?: Record<string, string>
    timeout?: number  // ms
  }): Promise<{
    stdout: string
    stderr: string
    code: number
  }>
}

interface LightShellShortcuts {
  /** Register a global keyboard shortcut. Works even when app is not focused. */
  register(accelerator: string, callback: () => void): Promise<void>

  /** Unregister a global shortcut. */
  unregister(accelerator: string): Promise<void>

  /** Unregister all global shortcuts. */
  unregisterAll(): Promise<void>

  /** Check if a shortcut is currently registered by this app. */
  isRegistered(accelerator: string): Promise<boolean>
}

interface UpdateInfo {
  version: string
  currentVersion: string
  notes: string
  pubDate: string
}

interface LightShellUpdater {
  /** Check for updates. Returns null if no update available. */
  check(): Promise<UpdateInfo | null>

  /** Download and install the latest update. Prompts restart. */
  install(): Promise<void>

  /** Check and install in one call. No-op if no update available. */
  checkAndInstall(): Promise<void>

  /** Listen to download progress. */
  onProgress(callback: (progress: {
    percent: number
    bytesDownloaded: number
    totalBytes: number
  }) => void): () => void
}
```

### Integration Points with Other Agents

| Agent 7 Provides | Used By |
|-------------------|---------|
| `PermissionEngine.Check()` | Agent 2's IPC handler — wraps every API call |
| CSP injection | Agent 1's webview loader — injects CSP meta tag before user HTML |
| Path validation | Agent 2's `fs.*` handlers — validates all paths before operations |
| DMG/deb/rpm builders | Agent 3's CLI `build` command — new `--target` options |
| New API handlers | Agent 2's IPC router — registered alongside existing APIs |
| Enhanced polyfills | Agent 4's polyfills.js — replaces current implementation |
| Store database | Agent 1's runtime — initializes bbolt DB on app start |

### Build Order

Agent 7 can work in parallel with existing agents but has these dependencies:

```
Agent 2 (IPC router interface) ──→ Agent 7 (new API handlers plug in)
Agent 3 (build pipeline)       ──→ Agent 7 (new packaging targets plug in)
Agent 4 (polyfills.js)         ──→ Agent 7 (enhanced polyfills replace existing)
Agent 1 (webview loader)       ──→ Agent 7 (CSP injection hooks in)
```

Agent 7 should work against the interfaces Agents 1-4 define. Agent 5 integrates Agent 7's output alongside everything else.

### Success Criteria

**Security:**
- [ ] Permission engine correctly blocks unauthorized fs access in restricted mode
- [ ] Permission engine allows everything in permissive mode (no permissions key)
- [ ] Path traversal attacks blocked (../../../etc/passwd)
- [ ] CSP injected in production builds, relaxed in dev
- [ ] Unix socket permissions are 0600
- [ ] Error messages on permission denial are AI-debuggable

**Packaging:**
- [ ] `lightshell build --target dmg` produces a working DMG with drag-to-install on macOS
- [ ] `lightshell build --target deb` produces an installable .deb on Debian/Ubuntu
- [ ] `lightshell build --target rpm` produces an installable .rpm on Fedora
- [ ] Code signing works when identity is provided
- [ ] All package formats remain under 5MB for the hello-world example

**Auto-Updater:**
- [ ] `lightshell.updater.check()` fetches and parses manifest correctly
- [ ] SHA256 verification rejects tampered downloads
- [ ] Binary replacement works on macOS (.app) and Linux (AppImage)
- [ ] Progress events fire during download
- [ ] HTTP-only endpoints rejected in production builds

**APIs:**
- [ ] `lightshell.store.set/get/delete` persists across app restarts
- [ ] `lightshell.http.fetch` bypasses CORS restrictions
- [ ] `lightshell.process.exec` respects command scoping in restricted mode
- [ ] `lightshell.process.exec` prevents shell injection (no shell expansion)
- [ ] `lightshell.shortcuts.register` works globally on macOS and Linux
- [ ] Deep links launch app and deliver URL via `lightshell.app.onOpenUrl`

**Polyfills:**
- [ ] Total polyfill size < 2KB
- [ ] All polyfilled APIs pass basic functional tests on WebKitGTK 2.38
- [ ] Scanner warns on all known unfixable gaps
- [ ] No polyfill modifies behavior when native implementation exists (feature detection first)

---

## Spec Amendments

### Updated Non-Goals

Replace the existing Non-Goals section:

- Windows support (v2)
- Mobile support (never)
- Custom rendering engine (use system webview)
- Node.js compatibility (this is not Electron)
- npm/node_modules in the app (keep it simple)
- React/Vue/Svelte integration (users can include via CDN if they want)
- Multi-window support (v2)
- Snap / Flatpak packaging (v1.1)
- WebSocket support in lightshell.http (v2)
- Delta updates / differential patching (v2)
- Notarization automation (v1.1 — documented how-to for manual notarization)

**Moved INTO scope (Agent 7):**
- ~~Auto-updater~~ → now in Agent 7
- ~~Code signing~~ → now in Agent 7 (macOS)

### Updated API Table

Agent 2's original 10 APIs + Agent 7's additions:

| # | Namespace | Methods | Owner | Priority |
|---|-----------|---------|-------|----------|
| 1 | window | setTitle, setSize, getSize, setPosition, getPosition, minimize, maximize, fullscreen, restore, close + events | Agent 2 | P0 |
| 2 | fs | readFile, writeFile, readDir, exists, stat, mkdir, remove, watch | Agent 2 | P0 |
| 3 | dialog | open, save, message, confirm, prompt | Agent 2 | P0 |
| 4 | clipboard | read, write | Agent 2 | P0 |
| 5 | shell | open | Agent 2 | P0 |
| 6 | notify | send | Agent 2 | P1 |
| 7 | tray | set, remove, onClick | Agent 2 | P1 |
| 8 | menu | set | Agent 2 | P1 |
| 9 | system | platform, arch, homeDir, tempDir, hostname | Agent 2 | P0 |
| 10 | app | quit, version, dataDir, onOpenUrl | Agent 2 + 7 | P0 |
| 11 | store | get, set, delete, has, keys, clear | Agent 7 | P0 |
| 12 | http | fetch, download | Agent 7 | P0 |
| 13 | process | exec | Agent 7 | P1 |
| 14 | shortcuts | register, unregister, unregisterAll, isRegistered | Agent 7 | P1 |
| 15 | updater | check, install, checkAndInstall, onProgress | Agent 7 | P1 |

### Updated Quality Bar

Add to existing quality bar:

**Security:**
- Permission check overhead: < 0.1ms per IPC call
- Path validation overhead: < 0.5ms per fs operation
- Zero known path traversal bypasses

**Size (updated):**
- Runtime binary: < 4MB (was 3MB — bbolt + http client add ~800KB)
- Polyfills: < 2KB (was 3KB — tighter scoping)