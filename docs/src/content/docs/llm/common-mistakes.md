---
title: Common AI Mistakes
description: Mistakes AI agents make when generating LightShell apps, and how to fix them.
---

AI models frequently make the same mistakes when generating LightShell code, especially when they lack sufficient context about the framework. This page catalogs each mistake with the wrong code, an explanation of why it fails, and the correct code.

Use this page to debug AI-generated code, or include it as context in your prompt to prevent these mistakes upfront.

---

## 1. Using require('fs') Instead of lightshell.fs

**Wrong:**

```js
const fs = require('fs')
const content = fs.readFileSync('/path/to/file.txt', 'utf-8')
```

**Why it fails:** There is no Node.js runtime in LightShell. `require` is not defined. The webview runs plain browser JavaScript with LightShell's APIs injected.

**Correct:**

```js
const content = await lightshell.fs.readFile('/path/to/file.txt')
```

This also applies to `require('path')`, `require('os')`, `require('child_process')`, and any other Node.js module.

---

## 2. Using window.open() for External URLs

**Wrong:**

```js
function openDocs() {
  window.open('https://lightshell.dev/docs')
}
```

**Why it fails:** `window.open()` attempts to navigate the webview to the URL or open a popup within the webview context. It does not open the system browser. The user sees the URL loaded inside your app window instead of in Chrome, Safari, or Firefox.

**Correct:**

```js
async function openDocs() {
  await lightshell.shell.open('https://lightshell.dev/docs')
}
```

`lightshell.shell.open()` delegates to the operating system, which opens the URL in the user's default browser.

---

## 3. Forgetting await on lightshell.* Calls

**Wrong:**

```js
const name = lightshell.store.get('user.name')
console.log(name) // Promise {<pending>}

lightshell.fs.writeFile('/tmp/data.txt', content)
// File may not be written before next line executes
```

**Why it fails:** Every `lightshell.*` method returns a Promise. Without `await`, you get a Promise object instead of the actual value, and write operations may not complete before subsequent code runs.

**Correct:**

```js
const name = await lightshell.store.get('user.name')
console.log(name) // "Alice"

await lightshell.fs.writeFile('/tmp/data.txt', content)
// File is guaranteed to be written
```

If you are inside a non-async function, either make it async or use `.then()`:

```js
// Option 1: make the function async
async function loadSettings() {
  const settings = await lightshell.store.get('settings')
  applySettings(settings)
}

// Option 2: use .then() (less preferred)
lightshell.store.get('settings').then(settings => {
  applySettings(settings)
})
```

---

## 4. Using showOpenFilePicker Instead of lightshell.dialog.open

**Wrong:**

```js
const [handle] = await window.showOpenFilePicker({
  types: [{ description: 'Text', accept: { 'text/plain': ['.txt'] } }]
})
const file = await handle.getFile()
const content = await file.text()
```

**Why it fails:** The File System Access API (`showOpenFilePicker`, `showSaveFilePicker`, `showDirectoryPicker`) is not available in any webview. It is a Chrome-only API that requires specific browser permissions not present in WKWebView or WebKitGTK.

**Correct:**

```js
const path = await lightshell.dialog.open({
  title: 'Select a text file',
  filters: [{ name: 'Text', extensions: ['txt'] }]
})
if (path) {
  const content = await lightshell.fs.readFile(path)
}
```

Note the `if (path)` check — `dialog.open()` returns `null` if the user cancels.

---

## 5. Using Node.js __dirname or process.env

**Wrong:**

```js
const configPath = __dirname + '/config.json'
const apiKey = process.env.API_KEY
const home = process.env.HOME
```

**Why it fails:** `__dirname`, `__filename`, and `process` are Node.js globals. They do not exist in a webview context.

**Correct:**

```js
// For app-specific data paths:
const dataDir = await lightshell.app.dataDir()
const configPath = `${dataDir}/config.json`

// For system information:
const home = await lightshell.system.homeDir()
const tmp = await lightshell.system.tempDir()

// For API keys, store them in lightshell.store:
const apiKey = await lightshell.store.get('apiKey')
```

---

## 6. Using system-ui Font Without Fallback

**Wrong:**

```css
body {
  font-family: system-ui;
}
```

**Why it fails:** The `system-ui` generic font resolves inconsistently across Linux distributions. On some systems it maps to a serif font or a fallback that looks nothing like the system UI font. This is a known WebKitGTK issue.

**Correct:**

```css
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif, "Apple Color Emoji", "Noto Color Emoji";
}
```

This explicit stack resolves to the native system font on macOS (`-apple-system`) and falls back to well-known fonts on Linux (`Noto Sans`, `Helvetica`, `Arial`).

---

## 7. Using CSS Nesting

**Wrong:**

```css
.sidebar {
  background: #f0f0f0;

  .item {
    padding: 8px;

    &:hover {
      background: #e0e0e0;
    }
  }
}
```

**Why it fails:** Native CSS nesting is not supported on WebKitGTK versions before 2.44. Many Linux users run older versions. The nested rules are silently ignored, and the styles do not apply.

**Correct:**

```css
.sidebar {
  background: #f0f0f0;
}

.sidebar .item {
  padding: 8px;
}

.sidebar .item:hover {
  background: #e0e0e0;
}
```

Flatten all nested CSS rules into standard selectors for maximum compatibility.

---

## 8. Not Checking for null from dialog.open() and dialog.save()

**Wrong:**

```js
const path = await lightshell.dialog.open()
const content = await lightshell.fs.readFile(path)
// Throws if user cancelled — path is null
```

**Why it fails:** When the user clicks "Cancel" in the file picker, `dialog.open()` returns `null`. Passing `null` to `lightshell.fs.readFile()` causes an error.

**Correct:**

```js
const path = await lightshell.dialog.open()
if (!path) return // User cancelled

const content = await lightshell.fs.readFile(path)
```

The same applies to `dialog.save()` and `dialog.prompt()` — all return `null` on cancellation.

---

## 9. Using Relative Paths with lightshell.fs

**Wrong:**

```js
await lightshell.fs.writeFile('data/settings.json', json)
await lightshell.fs.readFile('./notes/todo.txt')
```

**Why it fails:** LightShell's file system APIs require absolute paths. Relative paths have no reliable base directory in a webview context — there is no `cwd` concept like in Node.js.

**Correct:**

```js
const dataDir = await lightshell.app.dataDir()
await lightshell.fs.writeFile(`${dataDir}/settings.json`, json)

const home = await lightshell.system.homeDir()
await lightshell.fs.readFile(`${home}/notes/todo.txt`)
```

Always build paths from a known absolute base: `lightshell.app.dataDir()` for app-specific data, `lightshell.system.homeDir()` for user files, or `lightshell.system.tempDir()` for temporary files.

---

## 10. Trying to import or require Modules

**Wrong:**

```js
import marked from 'marked'
import { format } from 'date-fns'
```

```js
const marked = require('marked')
```

**Why it fails:** LightShell apps run in a webview, not in Node.js. There is no module bundler, no `node_modules` directory, and no module resolution system. Both `import` (ES modules from bare specifiers) and `require` (CommonJS) will fail.

**Correct:**

Load libraries from a CDN using a `<script>` tag in your HTML:

```html
<script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/date-fns/cdn.min.js"></script>
```

Then use the global variables the libraries expose:

```js
const html = marked.parse('# Hello')
const formatted = dateFns.format(new Date(), 'yyyy-MM-dd')
```

If you need a library that does not provide a UMD/global build, look for an alternative that does, or inline the code directly.

---

## 11. Using backdrop-filter Without Platform Fallback

**Wrong:**

```css
.toolbar {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px);
}
```

**Why it fails:** `backdrop-filter` has limited or no support on older WebKitGTK versions used by Linux. The toolbar ends up as a semi-transparent white rectangle with no blur, which often looks worse than a solid background.

**Correct:**

```css
.toolbar {
  background: #f5f5f5; /* solid fallback */
}

/* Apply blur only on macOS where it works reliably */
.platform-darwin .toolbar {
  background: rgba(255, 255, 255, 0.7);
  -webkit-backdrop-filter: blur(20px);
  backdrop-filter: blur(20px);
}
```

LightShell adds `.platform-darwin` or `.platform-linux` to the `<body>` element, so you can scope platform-specific CSS without JavaScript.

---

## 12. Not Ensuring the Data Directory Exists Before Writing

**Wrong:**

```js
const dataDir = await lightshell.app.dataDir()
await lightshell.fs.writeFile(`${dataDir}/notes/today.txt`, content)
// Fails if the "notes" subdirectory does not exist
```

**Why it fails:** `lightshell.fs.writeFile` creates the file but does not create parent directories. If `notes/` does not exist inside the data directory, the write fails.

**Correct:**

```js
const dataDir = await lightshell.app.dataDir()
const notesDir = `${dataDir}/notes`

// Ensure the directory exists first
const exists = await lightshell.fs.exists(notesDir)
if (!exists) {
  await lightshell.fs.mkdir(notesDir)
}

await lightshell.fs.writeFile(`${notesDir}/today.txt`, content)
```

Or as a reusable helper:

```js
async function ensureDir(dirPath) {
  const exists = await lightshell.fs.exists(dirPath)
  if (!exists) {
    await lightshell.fs.mkdir(dirPath)
  }
}
```

---

## 13. Using localStorage Instead of lightshell.store

**Wrong:**

```js
localStorage.setItem('settings', JSON.stringify(settings))
const settings = JSON.parse(localStorage.getItem('settings'))
```

**Why it fails:** While `localStorage` technically works in a webview, it is unreliable for desktop apps. The storage can be cleared by the OS, is limited to 5-10MB, and is tied to the webview origin which may change between app versions or builds. Data loss is common.

**Correct:**

```js
await lightshell.store.set('settings', settings)
const settings = await lightshell.store.get('settings')
```

`lightshell.store` is backed by a persistent database file at `$APP_DATA/store.db`. It survives app updates, OS cleanups, and webview origin changes. Values are automatically JSON-serialized, so you do not need `JSON.stringify` or `JSON.parse`.

---

## 14. Using browser fetch() for Cross-Origin API Calls

**Wrong:**

```js
const response = await fetch('https://api.github.com/user', {
  headers: { 'Authorization': 'Bearer ghp_xxxxx' }
})
const data = await response.json()
```

**Why it fails:** The browser's `fetch()` is subject to CORS restrictions. The webview enforces the same-origin policy, and most APIs do not include LightShell's origin in their `Access-Control-Allow-Origin` headers. The request either fails outright or is blocked by a preflight check.

**Correct:**

```js
const response = await lightshell.http.fetch('https://api.github.com/user', {
  method: 'GET',
  headers: { 'Authorization': 'Bearer ghp_xxxxx' }
})
const data = JSON.parse(response.body)
```

`lightshell.http.fetch` makes the HTTP request through the Go backend, which is not subject to CORS. Note that the response body is a string, so you need `JSON.parse()` instead of `.json()`.

---

## Quick Reference

| Mistake | Wrong | Correct |
|---------|-------|---------|
| Node.js modules | `require('fs')` | `lightshell.fs.*` |
| External URLs | `window.open(url)` | `lightshell.shell.open(url)` |
| Missing await | `lightshell.store.get(k)` | `await lightshell.store.get(k)` |
| File picker | `showOpenFilePicker()` | `lightshell.dialog.open()` |
| Environment vars | `process.env.HOME` | `await lightshell.system.homeDir()` |
| Font stack | `font-family: system-ui` | Explicit fallback stack |
| CSS nesting | `.a { .b { } }` | `.a .b { }` |
| Cancel handling | No null check | `if (!path) return` |
| Relative paths | `'./data/file.txt'` | `` `${dataDir}/file.txt` `` |
| npm imports | `import x from 'x'` | CDN `<script>` tag |
| Backdrop filter | `backdrop-filter: blur()` | Platform-scoped with fallback |
| Missing directory | Write without mkdir | `mkdir` then write |
| Local storage | `localStorage.setItem()` | `lightshell.store.set()` |
| CORS fetch | `fetch(crossOriginUrl)` | `lightshell.http.fetch()` |
