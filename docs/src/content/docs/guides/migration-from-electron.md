---
title: Migrating from Electron
description: Move your Electron app to LightShell — what changes, what stays the same.
---

LightShell and Electron solve the same problem -- building desktop apps with web technology -- but with very different approaches. This guide walks through what changes when you migrate, and how to translate Electron concepts to LightShell equivalents.

## Concept Mapping

| Electron | LightShell | Notes |
|----------|-----------|-------|
| `package.json` | `lightshell.json` | App metadata and config |
| `main.js` (main process) | Not needed | No separate main process |
| `preload.js` | Not needed | APIs available directly on `window.lightshell` |
| `renderer.js` | Your JS files in `src/` | Standard browser JavaScript |
| `BrowserWindow` | `lightshell.json` `window` config | Declarative, not programmatic |
| `ipcMain` / `ipcRenderer` | Automatic | All APIs go through IPC transparently |
| `require('electron')` | `window.lightshell` | Global object, no imports |
| `node_modules/` | Not applicable | No Node.js runtime |
| `electron-builder` / `electron-forge` | `lightshell build` | Single command, no config file |
| ~150MB app bundle | ~5MB app bundle | System webview, no bundled Chromium |

## Step-by-Step Migration

### 1. Create lightshell.json

Replace your `package.json` + Electron config with a single `lightshell.json`:

**Before (Electron package.json):**

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "main": "main.js",
  "scripts": {
    "start": "electron .",
    "build": "electron-builder"
  },
  "devDependencies": {
    "electron": "^28.0.0",
    "electron-builder": "^24.0.0"
  }
}
```

**After (LightShell lightshell.json):**

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "My App",
    "width": 1200,
    "height": 800,
    "minWidth": 400,
    "minHeight": 300
  }
}
```

### 2. Move Your Renderer Code

In Electron, your UI code lives in renderer files that are loaded by `BrowserWindow`. In LightShell, move those files into `src/`:

**Electron structure:**

```
my-app/
  main.js
  preload.js
  renderer/
    index.html
    style.css
    app.js
  node_modules/
  package.json
```

**LightShell structure:**

```
my-app/
  lightshell.json
  src/
    index.html
    style.css
    app.js
```

Your HTML, CSS, and frontend JavaScript move over with minimal changes.

### 3. Replace Electron APIs

Remove all `require('electron')` calls and `window.electronAPI` references. Replace them with the `window.lightshell` equivalents.

#### Window Management

```js
// Electron (main.js)
const { BrowserWindow } = require('electron')
const win = new BrowserWindow({ width: 800, height: 600 })
win.setTitle('My App')
win.maximize()

// LightShell (your JS)
await lightshell.window.setTitle('My App')
await lightshell.window.maximize()
```

#### File Dialogs

```js
// Electron (main.js + preload bridge)
const { dialog } = require('electron')
const result = await dialog.showOpenDialog({
  properties: ['openFile'],
  filters: [{ name: 'Text', extensions: ['txt'] }]
})
const filePath = result.filePaths[0]

// LightShell
const filePath = await lightshell.dialog.open({
  filters: [{ name: 'Text', extensions: ['txt'] }]
})
```

#### File System

```js
// Electron (using Node.js fs)
const fs = require('fs').promises
const content = await fs.readFile('/path/to/file', 'utf-8')
await fs.writeFile('/path/to/file', content)

// LightShell
const content = await lightshell.fs.readFile('/path/to/file')
await lightshell.fs.writeFile('/path/to/file', content)
```

#### Clipboard

```js
// Electron
const { clipboard } = require('electron')
clipboard.writeText('Hello')
const text = clipboard.readText()

// LightShell
await lightshell.clipboard.write('Hello')
const text = await lightshell.clipboard.read()
```

#### Shell / Open External

```js
// Electron
const { shell } = require('electron')
shell.openExternal('https://example.com')
shell.openPath('/path/to/file')

// LightShell
await lightshell.shell.open('https://example.com')
await lightshell.shell.open('/path/to/file')
```

#### Notifications

```js
// Electron
new Notification({ title: 'Hello', body: 'World' }).show()

// LightShell
await lightshell.notify.send({ title: 'Hello', body: 'World' })
```

#### App Data Path

```js
// Electron
const { app } = require('electron')
const dataPath = app.getPath('userData')

// LightShell
const dataPath = await lightshell.app.dataDir()
```

#### System Info

```js
// Electron
const os = require('os')
const platform = process.platform
const arch = process.arch
const homeDir = os.homedir()

// LightShell
const platform = await lightshell.system.platform()  // "darwin" or "linux"
const arch = await lightshell.system.arch()           // "arm64" or "x64"
const homeDir = await lightshell.system.homeDir()
```

### 4. Remove Node.js-isms

LightShell does not include a Node.js runtime. Remove or replace these patterns:

```js
// Remove these -- they will not work
const path = require('path')
const fs = require('fs')
const http = require('http')
const { exec } = require('child_process')
process.env.MY_VAR
__dirname
__filename

// Replace with LightShell equivalents
// path.join -> string concatenation or template literals
const filePath = `${baseDir}/${fileName}`

// fs -> lightshell.fs
await lightshell.fs.readFile(path)

// http -> lightshell.http
const response = await lightshell.http.fetch('https://api.example.com/data')

// child_process -> lightshell.process
const result = await lightshell.process.exec('git', ['status'])
```

### 5. Handle Data Persistence

Replace Electron's `electron-store` or raw `fs` persistence with `lightshell.store`:

```js
// Electron (with electron-store)
const Store = require('electron-store')
const store = new Store()
store.set('theme', 'dark')
const theme = store.get('theme')

// LightShell
await lightshell.store.set('theme', 'dark')
const theme = await lightshell.store.get('theme')
```

Note that `lightshell.store` methods are async, while `electron-store` methods are synchronous. Add `await` to all store calls.

### 6. Test with lightshell dev

```bash
cd my-app
lightshell dev
```

This starts your app with hot reloading. Fix any remaining API mismatches by checking the error messages -- LightShell errors include the API method name and a link to the relevant docs.

### 7. Build

```bash
lightshell build
```

This produces a native `.app` bundle on macOS or an AppImage on Linux. No `electron-builder` config, no signing certificate setup (unless you want code signing), no waiting for Chromium to compile.

## What You Lose

Be aware of these Electron features that LightShell does not provide:

| Electron Feature | LightShell Status |
|-----------------|-------------------|
| Node.js in renderer | Not available -- use lightshell.* APIs |
| Native Node modules (sharp, better-sqlite3) | Not available -- use Go backend equivalents |
| Multi-window | Not in v1 |
| `<webview>` tag / BrowserView | Not available |
| Chrome DevTools Protocol | Limited -- webview inspector in dev mode only |
| Chrome extensions | Not available |
| Custom protocol handlers (file://) | `lightshell:` protocol for bundled assets |
| `session` / cookie management | Not available |
| `screen` API (display enumeration) | Not in v1 |
| `powerMonitor` / `powerSaveBlocker` | Not available |
| Windows support | LightShell v2 |

## What You Gain

| Benefit | Details |
|---------|---------|
| **App size** | ~5MB vs ~150MB. No bundled Chromium. |
| **Build speed** | Seconds, not minutes. No Webpack/Vite step required. |
| **Memory usage** | System webview shares memory with OS. No separate Chromium process. |
| **Simplicity** | One config file, one command to build, no main/renderer split. |
| **Security** | Permission system, CSP injection, path traversal protection built in. |
| **AI-friendly** | Error messages include fix instructions. APIs designed for AI code generation. |
| **No node_modules** | Zero npm dependencies required to build a LightShell app. |

## Common Migration Pitfalls

**IPC bridge removal.** In Electron, you set up `contextBridge.exposeInMainWorld` in a preload script to expose APIs to the renderer. In LightShell, all APIs are already available on `window.lightshell`. Remove all IPC bridge code.

**Sync vs async.** Some Electron APIs are synchronous (e.g., `clipboard.readText()`). All LightShell APIs are async. Add `await` where needed.

**Node.js module imports.** Any `require()` or `import` of Node.js built-in modules will fail. Replace with LightShell API calls or remove the functionality.

**NPM packages.** If you use npm packages for UI (React, Vue, etc.), you can still include them via a CDN `<script>` tag or a bundled JS file. You just cannot use `require()` to load them.

**Path separators.** Electron apps sometimes use `path.join()` for cross-platform path building. Since LightShell only runs on macOS and Linux (both use `/`), simple string concatenation works fine.
