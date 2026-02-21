---
title: Migrating from Neutralinojs
description: Move your Neutralinojs app to LightShell — what changes, what stays the same.
---

Neutralinojs and LightShell share a similar philosophy: use the system webview, provide a lightweight runtime, and avoid bundling a browser engine. This makes migration straightforward. The core ideas translate directly -- the main changes are in API naming, security model, and IPC transport.

## Concept Mapping

| Neutralinojs | LightShell | Notes |
|-------------|-----------|-------|
| `neutralino.config.json` | `lightshell.json` | App configuration |
| `Neutralino.*` | `lightshell.*` | Global API object |
| `neutralino.js` client library | Built-in (auto-injected) | No script tag needed |
| `Neutralino.init()` | Not needed | LightShell initializes automatically |
| Localhost WebSocket IPC | Unix domain socket IPC | More secure |
| `resources/` directory | `src/` directory | App source files |
| `neu build` | `lightshell build` | Build command |
| `neu run` | `lightshell dev` | Development command |
| `neu create` | `lightshell init` | Project scaffolding |
| ~2MB binary + resources | ~5MB self-contained binary | Resources embedded in binary |
| No permission system | Granular permissions | Optional restricted mode |

## API Mapping

### File System

```js
// Neutralinojs
await Neutralino.filesystem.readFile('/path/to/file')
await Neutralino.filesystem.writeFile('/path/to/file', 'content')
await Neutralino.filesystem.readDirectory('/path')
await Neutralino.filesystem.createDirectory('/path/to/dir')
await Neutralino.filesystem.removeFile('/path/to/file')
await Neutralino.filesystem.removeDirectory('/path/to/dir')

// LightShell
await lightshell.fs.readFile('/path/to/file')
await lightshell.fs.writeFile('/path/to/file', 'content')
await lightshell.fs.readDir('/path')
await lightshell.fs.mkdir('/path/to/dir')
await lightshell.fs.remove('/path/to/file')
await lightshell.fs.remove('/path/to/dir')      // same method for files and dirs
```

### Dialogs

```js
// Neutralinojs
const entries = await Neutralino.os.showOpenDialog('Open File', {
  filters: [{ name: 'Text', extensions: ['txt'] }]
})
const path = entries[0]

const savePath = await Neutralino.os.showSaveDialog('Save', {
  filters: [{ name: 'Text', extensions: ['txt'] }]
})

await Neutralino.os.showMessageBox('Title', 'Message')

// LightShell
const path = await lightshell.dialog.open({
  title: 'Open File',
  filters: [{ name: 'Text', extensions: ['txt'] }]
})

const savePath = await lightshell.dialog.save({
  title: 'Save',
  filters: [{ name: 'Text', extensions: ['txt'] }]
})

await lightshell.dialog.message('Title', 'Message')
```

Note: `Neutralino.os.showOpenDialog` returns an array of paths. `lightshell.dialog.open` returns a single path (or an array if `multiple: true` is set).

### Clipboard

```js
// Neutralinojs
await Neutralino.clipboard.writeText('Hello')
const text = await Neutralino.clipboard.readText()

// LightShell
await lightshell.clipboard.write('Hello')
const text = await lightshell.clipboard.read()
```

### Process Execution

```js
// Neutralinojs -- no scoping, full shell access
const result = await Neutralino.os.execCommand('ls -la /tmp')
console.log(result.stdOut)

// LightShell -- direct execution, no shell, optional scoping
const result = await lightshell.process.exec('ls', ['-la', '/tmp'])
console.log(result.stdout)
```

Key difference: Neutralinojs passes commands through the shell (`sh -c`), which allows shell injection. LightShell executes commands directly via `exec.Command`, which prevents shell injection entirely. Pipe characters, semicolons, and backticks are treated as literal strings, not shell operators.

### Storage

```js
// Neutralinojs -- file-based key-value
await Neutralino.storage.setData('user', JSON.stringify({ name: 'Alice' }))
const raw = await Neutralino.storage.getData('user')
const user = JSON.parse(raw)

// LightShell -- JSON-native key-value store
await lightshell.store.set('user', { name: 'Alice' })    // auto-serialized
const user = await lightshell.store.get('user')           // auto-deserialized
```

LightShell's store automatically handles JSON serialization. You pass objects directly instead of manually calling `JSON.stringify` and `JSON.parse`.

### Window Management

```js
// Neutralinojs
await Neutralino.window.setTitle('My App')
await Neutralino.window.setSize({ width: 800, height: 600 })
await Neutralino.window.minimize()
await Neutralino.window.maximize()
await Neutralino.window.setFullScreen()

// LightShell
await lightshell.window.setTitle('My App')
await lightshell.window.setSize(800, 600)
await lightshell.window.minimize()
await lightshell.window.maximize()
await lightshell.window.fullscreen()
```

### System / OS Info

```js
// Neutralinojs
const info = await Neutralino.computer.getOSInfo()
const platform = NL_OS  // global constant

// LightShell
const platform = await lightshell.system.platform()   // "darwin" or "linux"
const arch = await lightshell.system.arch()            // "arm64" or "x64"
const homeDir = await lightshell.system.homeDir()
```

### App Lifecycle

```js
// Neutralinojs
Neutralino.init()
Neutralino.events.on('windowClose', () => {
  Neutralino.app.exit()
})

// LightShell -- no init needed, quit is explicit
// App starts automatically when index.html loads
lightshell.app.quit()  // call when you want to exit
```

### Open External URLs

```js
// Neutralinojs
await Neutralino.os.open('https://example.com')

// LightShell
await lightshell.shell.open('https://example.com')
```

### Notifications

```js
// Neutralinojs
await Neutralino.os.showNotification('Title', 'Body')

// LightShell
await lightshell.notify.send({ title: 'Title', body: 'Body' })
```

### System Tray

```js
// Neutralinojs
await Neutralino.os.setTray({
  icon: '/resources/icon.png',
  menuItems: [
    { id: 'show', text: 'Show' },
    { id: 'quit', text: 'Quit' }
  ]
})
Neutralino.events.on('trayMenuItemClicked', (e) => {
  if (e.detail.id === 'quit') Neutralino.app.exit()
})

// LightShell
await lightshell.tray.set({
  tooltip: 'My App',
  menu: [
    { label: 'Show', id: 'show' },
    { label: 'Quit', id: 'quit' }
  ]
})
lightshell.tray.onClick((data) => {
  if (data.id === 'quit') lightshell.app.quit()
})
```

## Migration Steps

### 1. Create lightshell.json

Translate your `neutralino.config.json` to `lightshell.json`:

**Before (neutralino.config.json):**

```json
{
  "applicationId": "com.example.myapp",
  "defaultMode": "window",
  "port": 0,
  "url": "/resources/",
  "nativeAllowList": [
    "app.*",
    "os.*",
    "filesystem.*",
    "clipboard.*",
    "window.*"
  ],
  "modes": {
    "window": {
      "title": "My App",
      "width": 1000,
      "height": 700,
      "minWidth": 400,
      "minHeight": 300
    }
  }
}
```

**After (lightshell.json):**

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "My App",
    "width": 1000,
    "height": 700,
    "minWidth": 400,
    "minHeight": 300
  }
}
```

### 2. Move Source Files

Rename `resources/` to `src/` and update any internal paths:

```
# Neutralinojs
resources/
  index.html
  styles/
  scripts/

# LightShell
src/
  index.html
  styles/
  scripts/
```

### 3. Remove Neutralinojs Client Library

In Neutralinojs, you include `<script src="/__neutralino_globals.js"></script>` in your HTML. Remove this tag. LightShell automatically injects its client library -- no script tag needed.

```html
<!-- Remove this line -->
<script src="/__neutralino_globals.js"></script>

<!-- Remove this line too -->
<script>Neutralino.init()</script>
```

### 4. Replace API Calls

Find and replace all `Neutralino.*` calls with `lightshell.*` equivalents using the mapping table above. The most common replacements:

```
Neutralino.filesystem.readFile     -> lightshell.fs.readFile
Neutralino.filesystem.writeFile    -> lightshell.fs.writeFile
Neutralino.os.showOpenDialog       -> lightshell.dialog.open
Neutralino.os.showSaveDialog       -> lightshell.dialog.save
Neutralino.os.showMessageBox       -> lightshell.dialog.message
Neutralino.os.execCommand          -> lightshell.process.exec
Neutralino.clipboard.readText      -> lightshell.clipboard.read
Neutralino.clipboard.writeText     -> lightshell.clipboard.write
Neutralino.storage.setData         -> lightshell.store.set
Neutralino.storage.getData         -> lightshell.store.get
Neutralino.window.setTitle         -> lightshell.window.setTitle
Neutralino.app.exit                -> lightshell.app.quit
Neutralino.os.open                 -> lightshell.shell.open
```

### 5. Update Event Handling

```js
// Neutralinojs events
Neutralino.events.on('windowClose', handler)
Neutralino.events.on('trayMenuItemClicked', handler)
Neutralino.events.on('ready', handler)

// LightShell -- use standard DOM events or lightshell event system
window.addEventListener('beforeunload', handler)
lightshell.tray.onClick(handler)
// No "ready" event needed -- code runs when the script loads
```

### 6. Fix execCommand Calls

This is the change that requires the most attention. Neutralinojs runs commands through the shell as a single string. LightShell splits the command and arguments:

```js
// Neutralinojs (shell string)
await Neutralino.os.execCommand('grep -r "TODO" /project/src')

// LightShell (command + args array)
await lightshell.process.exec('grep', ['-r', 'TODO', '/project/src'])
```

If your Neutralinojs code uses shell features like pipes or redirection, you need to restructure:

```js
// Neutralinojs -- uses shell pipe
await Neutralino.os.execCommand('cat file.txt | grep error | wc -l')

// LightShell -- run commands separately and process in JS
const result = await lightshell.process.exec('cat', ['file.txt'])
const errorLines = result.stdout.split('\n').filter(l => l.includes('error'))
const count = errorLines.length
```

### 7. Test and Build

```bash
lightshell dev     # test in development
lightshell build   # produce the final binary
```

## What You Gain

| Improvement | Details |
|------------|---------|
| **Security (IPC)** | Unix domain socket with 0600 permissions vs localhost WebSocket. Neutralinojs's WebSocket allows any process on the machine to connect to your app's IPC channel. LightShell's socket is owner-only. |
| **Security (permissions)** | Granular permission system with fs, process, and http scoping. Neutralinojs's `nativeAllowList` is module-level only (all-or-nothing per namespace). |
| **Security (process exec)** | Direct execution prevents shell injection. Neutralinojs passes strings to `sh -c`. |
| **Path traversal protection** | Always-on symlink resolution and path validation. Not available in Neutralinojs. |
| **CSP injection** | Automatic Content Security Policy in production builds. |
| **Built-in store** | JSON-native key-value store with `lightshell.store`. Neutralinojs storage writes raw files. |
| **CORS-free HTTP** | `lightshell.http.fetch` bypasses CORS via the Go backend. Neutralinojs requires custom proxy setup. |
| **AI-friendly errors** | Structured error messages with fix instructions and doc links. |
| **Self-contained binary** | Single binary with embedded resources. Neutralinojs requires a binary + resources directory. |

## What Changes

| Neutralinojs Feature | LightShell Equivalent |
|----------------------|----------------------|
| `nativeAllowList` | `permissions` in lightshell.json (more granular) |
| Extensions (child processes) | `lightshell.process.exec` |
| `NL_PORT`, `NL_TOKEN` globals | Not exposed (IPC is transparent) |
| `Neutralino.debug.log` | `console.log` (shown in dev tools) |
| Custom cloud mode | Not available (desktop only) |
| `.neu` project metadata | Not used |
| Binary + `resources.neu` bundle | Single self-contained binary |

## Platform Notes

Both Neutralinojs and LightShell use the system webview. If your Neutralinojs app already works with WebKitGTK on Linux and WKWebView on macOS, your UI code should work in LightShell without changes. The same browser engine quirks and limitations apply to both frameworks.

LightShell includes polyfills for APIs missing in older WebKitGTK versions (`structuredClone`, `Array.prototype.group`, `Promise.withResolvers`, Set methods). If your Neutralinojs app had workarounds for these gaps, you can remove them.
