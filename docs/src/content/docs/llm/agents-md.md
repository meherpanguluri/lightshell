---
title: AGENTS.md
description: Configure Claude Code and other AI agents for LightShell projects.
---

AGENTS.md is a project-level instruction file that AI coding agents read automatically. When an agent like Claude Code, Devin, or SWE-Agent opens your project, it looks for an AGENTS.md file at the repository root and follows the instructions inside. This is the most effective way to prevent AI agents from making LightShell-specific mistakes.

## Installation

Create an `AGENTS.md` file in the root of your LightShell project:

```bash
cd my-lightshell-app
touch AGENTS.md
```

Paste the complete template below into the file. Claude Code reads AGENTS.md automatically on every conversation start — no additional configuration is needed.

## The Complete AGENTS.md Template

````markdown
# LightShell Desktop App

## Project Overview

This is a desktop application built with [LightShell](https://lightshell.dev).
LightShell apps are written in HTML, CSS, and JavaScript and compile to native
binaries using system webviews (WKWebView on macOS, WebKitGTK on Linux).

There is no Node.js runtime, no Electron, no npm, and no build step.
All native APIs are available as async functions on the global `window.lightshell` object.

## Tech Stack

- **Frontend:** Vanilla HTML, CSS, JavaScript (no framework)
- **Backend:** Go runtime (you never write or see Go code)
- **Native APIs:** `window.lightshell.*` — injected automatically
- **Persistence:** `lightshell.store` (key-value) or `lightshell.fs` (file system)
- **HTTP:** `lightshell.http.fetch` (CORS-free, goes through Go backend)
- **Packaging:** `lightshell build` produces a native binary

## File Structure

```
lightshell.json          # App configuration (name, version, window size, permissions)
src/
  index.html             # Entry point — loaded by the webview
  app.js                 # Application logic
  style.css              # Styles
  assets/                # Images, fonts, etc.
```

## API Reference

All methods are async. Always use `await`.

### File System (`lightshell.fs`)
- `readFile(path, encoding?)` — read file as string
- `writeFile(path, content)` — write string to file
- `readDir(path)` — list directory contents, returns `[{name, isDir}]`
- `exists(path)` — check if path exists, returns boolean
- `stat(path)` — get file info `{size, isDir, modified}`
- `mkdir(path)` — create directory (recursive)
- `remove(path)` — delete file or directory
- `watch(path, callback)` — watch for changes

### Dialogs (`lightshell.dialog`)
- `open(options?)` — native file open picker, returns path or null
- `save(options?)` — native file save picker, returns path or null
- `message(title, body)` — info message box
- `confirm(title, body)` — yes/no dialog, returns boolean
- `prompt(title, body, defaultValue?)` — text input dialog, returns string or null

### Key-Value Store (`lightshell.store`)
- `get(key)` — get value (JSON-deserialized), null if missing
- `set(key, value)` — set value (JSON-serialized)
- `delete(key)` — delete key
- `has(key)` — check existence, returns boolean
- `keys(prefix?)` — list keys matching prefix
- `clear()` — delete all keys

### HTTP Client (`lightshell.http`)
- `fetch(url, options?)` — CORS-free HTTP request, returns `{status, headers, body}`
- `download(url, options?)` — download file to disk

### Window (`lightshell.window`)
- `setTitle(title)`, `setSize(w, h)`, `getSize()`, `setPosition(x, y)`, `getPosition()`
- `minimize()`, `maximize()`, `fullscreen()`, `restore()`, `close()`

### Other APIs
- `lightshell.clipboard.read()` / `.write(text)` — clipboard access
- `lightshell.shell.open(url)` — open URL in browser or file in default app
- `lightshell.notify.send({title, body})` — system notification
- `lightshell.system.platform()` / `.arch()` / `.homeDir()` / `.tempDir()` / `.hostname()`
- `lightshell.app.quit()` / `.version()` / `.dataDir()`
- `lightshell.process.exec(cmd, args?, options?)` — run system command
- `lightshell.shortcuts.register(combo, callback)` — global keyboard shortcut
- `lightshell.tray.set({title, icon, menu})` — system tray icon
- `lightshell.menu.set(template)` — native app menu

## Build Commands

```bash
lightshell dev              # Start dev server with hot reload
lightshell build            # Build native binary (.app on macOS, AppImage on Linux)
lightshell build --target dmg    # macOS DMG with drag-to-install
lightshell build --target deb    # Debian package
lightshell build --target rpm    # RPM package
lightshell doctor           # Check system dependencies
```

## Common Patterns

### Persistence
```js
// Use lightshell.store for app state
await lightshell.store.set('todos', todoList)
const todos = await lightshell.store.get('todos') || []
```

### File Operations
```js
// Always use absolute paths
const dataDir = await lightshell.app.dataDir()
const filePath = `${dataDir}/notes.json`
await lightshell.fs.writeFile(filePath, JSON.stringify(data))
```

### File Picker
```js
// Always check for null (user may cancel)
const path = await lightshell.dialog.open({
  filters: [{ name: 'Text', extensions: ['txt', 'md'] }]
})
if (path) {
  const content = await lightshell.fs.readFile(path)
}
```

### API Calls
```js
// Use lightshell.http.fetch, NOT browser fetch (CORS issues)
const res = await lightshell.http.fetch('https://api.example.com/data', {
  method: 'GET',
  headers: { 'Authorization': 'Bearer token' }
})
const data = JSON.parse(res.body)
```

### External Links
```js
// Use lightshell.shell.open, NOT window.open
await lightshell.shell.open('https://example.com')
```

## Rules

- NEVER use require() or import for Node.js modules — they do not exist
- NEVER use window.open() for external URLs — use lightshell.shell.open()
- NEVER use localStorage — use lightshell.store
- NEVER use browser fetch() for cross-origin — use lightshell.http.fetch()
- NEVER use relative file paths — always build from dataDir() or homeDir()
- ALWAYS await lightshell.* calls — they return Promises
- ALWAYS check for null from dialog.open() and dialog.save()
- ALWAYS use explicit font stacks in CSS, not system-ui alone
- AVOID CSS nesting (breaks on older Linux WebKitGTK)
- AVOID :has() CSS selector (breaks on older Linux WebKitGTK)

## Testing

Run the app during development:
```bash
lightshell dev
```
Open DevTools with Cmd+Option+I (macOS) or Ctrl+Shift+I (Linux) in dev mode.
Check the console for errors. All lightshell.* errors include the fix in the message.
````

## How Claude Code Uses AGENTS.md

Claude Code reads the AGENTS.md file at the start of every conversation. The instructions become part of the AI's context, so it will:

- Use `lightshell.store` instead of `localStorage` for persistence
- Use `lightshell.http.fetch` instead of browser `fetch` for API calls
- Use `lightshell.shell.open` instead of `window.open` for external links
- Always `await` LightShell API calls
- Generate the correct file structure (`src/index.html`, `src/app.js`, `src/style.css`)
- Run `lightshell dev` for testing, not `npm start` or `node server.js`

## Placement

The file must be at the project root — the same directory as `lightshell.json`:

```
my-lightshell-app/
  AGENTS.md              <-- here
  lightshell.json
  src/
    index.html
    app.js
    style.css
```

## Customizing the Template

The template above is generic. Customize it for your specific project:

- **Add your app's data model** — describe the key entities and how they are stored
- **Add your app's routes/views** — if your app has multiple screens, describe the navigation structure
- **Add your API endpoints** — if your app calls external APIs, list them with expected request/response formats
- **Add your coding conventions** — indentation, naming, comment style

Example additions:

```markdown
## App-Specific Data Model

This app manages a list of bookmarks. Each bookmark has:
- url (string) — the bookmarked URL
- title (string) — display title
- tags (string[]) — categorization tags
- createdAt (number) — Unix timestamp

Bookmarks are stored via lightshell.store with the key 'bookmarks' as a JSON array.

## External APIs

The app fetches metadata from:
- GET https://api.example.com/metadata?url={url} — returns {title, description, image}
- Use lightshell.http.fetch for all requests
- API key is stored in lightshell.store under 'settings.apiKey'
```

## Other Agent Tools

AGENTS.md is also read by:

- **Devin** — reads AGENTS.md from the repository root
- **SWE-Agent** — reads AGENTS.md when starting a task
- **Aider** — can be instructed to read AGENTS.md via its conventions file
- **Cline / Continue** — reads project root files for context

The format is plain Markdown, so any tool that reads project documentation will benefit from it.
