---
title: API Overview
description: Overview of all LightShell JavaScript APIs.
---

LightShell exposes a complete set of native APIs through the `window.lightshell` object. All API methods are async and return Promises. No imports, no build tools, no node_modules — just call `lightshell.*` from any JavaScript in your app.

## API Namespaces

| # | Namespace | Methods | Priority | Description |
|---|-----------|---------|----------|-------------|
| 1 | [window](/docs/api/window/) | setTitle, setSize, getSize, setPosition, getPosition, minimize, maximize, fullscreen, restore, close | P0 | Window management and state |
| 2 | [fs](/docs/api/fs/) | readFile, writeFile, readDir, exists, stat, mkdir, remove, watch | P0 | File system access |
| 3 | [dialog](/docs/api/dialog/) | open, save, message, confirm, prompt | P0 | Native file pickers and message dialogs |
| 4 | [clipboard](/docs/api/clipboard/) | read, write | P0 | System clipboard text access |
| 5 | [system](/docs/api/system/) | platform, arch, homeDir, tempDir, hostname | P0 | OS information and system paths |
| 6 | [app](/docs/api/app/) | quit, version, dataDir | P0 | Application lifecycle and metadata |
| 7 | [shell](/docs/api/shell/) | open | P0 | Open URLs and files with system defaults |
| 8 | [notify](/docs/api/notify/) | send | P1 | System notifications |
| 9 | [tray](/docs/api/tray/) | set, remove, onClick | P1 | System tray icon and menu |
| 10 | [menu](/docs/api/menu/) | set | P1 | Application menu bar |
| 11 | [store](/docs/api/store/) | get, set, delete, has, keys, clear | P0 | Persistent key-value storage |
| 12 | [http](/docs/api/http/) | fetch, download | P0 | CORS-free HTTP requests |
| 13 | [process](/docs/api/process/) | exec | P1 | Scoped system command execution |
| 14 | [shortcuts](/docs/api/shortcuts/) | register, unregister, unregisterAll, isRegistered | P1 | Global keyboard shortcuts |
| 15 | [updater](/docs/api/updater/) | check, install, checkAndInstall, onProgress | P1 | Auto-update mechanism |

**P0** = core APIs available from day one. **P1** = important APIs that ship in v1 but are secondary to core functionality.

## Quick Example

A small example using multiple APIs together — a note-taking app that saves notes to the store, uses keyboard shortcuts, and sends a notification when done.

```js
// Initialize the app
const settings = await lightshell.store.get('settings') || { theme: 'light' }
const version = await lightshell.app.version()
await lightshell.window.setTitle(`Notes — v${version}`)

// Load saved notes
const notes = await lightshell.store.get('notes') || []
renderNotes(notes)

// Save a new note
async function saveNote(text) {
  notes.push({ id: Date.now(), text, created: new Date().toISOString() })
  await lightshell.store.set('notes', notes)
  renderNotes(notes)
}

// Export notes to a file
async function exportNotes() {
  const path = await lightshell.dialog.save({
    title: 'Export Notes',
    defaultPath: 'notes.json',
    filters: [{ name: 'JSON', extensions: ['json'] }]
  })
  if (path) {
    await lightshell.fs.writeFile(path, JSON.stringify(notes, null, 2))
    await lightshell.notify.send('Export Complete', `Saved ${notes.length} notes to ${path}`)
  }
}

// Register a global shortcut for quick note capture
await lightshell.shortcuts.register('CommandOrControl+Shift+N', () => {
  lightshell.window.restore()
  document.getElementById('note-input').focus()
})

// Open external links in the system browser
document.addEventListener('click', (e) => {
  const link = e.target.closest('a[href^="http"]')
  if (link) {
    e.preventDefault()
    lightshell.shell.open(link.href)
  }
})
```

## Event System

LightShell also provides an event system for reacting to asynchronous changes. See the [Events](/docs/api/events/) reference for the full list.

```js
const unsubscribe = lightshell.on('window.resize', (data) => {
  console.log(`Window resized to ${data.width}x${data.height}`)
})
```

## Configuration

All app settings, permissions, and build options are defined in `lightshell.json`. See the [Configuration](/docs/api/config/) reference.

## Key Principles

- **All APIs are async.** Every method returns a Promise. Use `await` or `.then()`.
- **No imports needed.** The `lightshell` object is globally available on `window` — no import statements, no bundlers.
- **Permissive by default.** Without a `permissions` key in `lightshell.json`, all operations are allowed. Add permissions only when you want to restrict access for security.
- **AI-friendly errors.** Error messages include what was attempted, what went wrong, and how to fix it. They are designed to be useful to both humans and AI coding assistants.
- **Cross-platform.** All APIs work identically on macOS and Linux. Platform differences are handled internally.
