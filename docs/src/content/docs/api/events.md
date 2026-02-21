---
title: Events
description: Complete reference for LightShell events — window, file system, tray, menu, and update events.
---

LightShell uses an event system to communicate asynchronous changes from the native backend to your JavaScript code. Events cover window state changes, file system watches, tray and menu interactions, shortcuts, and updater progress.

## Listening to Events

### lightshell.on(event, callback)

Subscribe to a named event. The callback is invoked each time the event fires.

**Parameters:**
- `event` (string) — the event name (see the full list below)
- `callback` (function) — receives an event data object specific to each event type

**Returns:** unsubscribe function — call it to stop listening

**Example:**
```js
const unsubscribe = lightshell.on('window.resize', (data) => {
  console.log(`New size: ${data.width}x${data.height}`)
})

// Later, stop listening
unsubscribe()
```

---

## Event Reference

### Window Events

#### window.resize

Fired when the window is resized by the user or programmatically.

**Data:** `{ width: number, height: number }`

```js
lightshell.on('window.resize', (data) => {
  console.log(`${data.width}x${data.height}`)
  adjustLayout(data.width, data.height)
})
```

---

#### window.move

Fired when the window is moved to a new position.

**Data:** `{ x: number, y: number }`

```js
lightshell.on('window.move', (data) => {
  console.log(`Window at (${data.x}, ${data.y})`)
})
```

---

#### window.focus

Fired when the window gains focus (becomes the active window).

**Data:** none

```js
lightshell.on('window.focus', () => {
  document.title = 'My App'
  refreshData()
})
```

---

#### window.blur

Fired when the window loses focus.

**Data:** none

```js
lightshell.on('window.blur', () => {
  pauseAnimations()
})
```

---

### File System Events

#### fs.watch

Fired when a watched file or directory changes. This event is triggered by `lightshell.fs.watch()` and is documented here for completeness. Use `lightshell.fs.watch()` directly for file watching.

**Data:** `{ path: string, event: string }`

- `path` — the path of the changed file
- `event` — the type of change: `"create"`, `"write"`, `"remove"`, `"rename"`

```js
lightshell.fs.watch('/tmp/data.json', (event) => {
  console.log(`${event.path} was ${event.event}d`)
  reloadData()
})
```

---

### Menu Events

#### menu.click

Fired when a menu bar item is clicked. The `id` field matches the `id` you assigned to the menu item in `lightshell.menu.set()`.

**Data:** `{ id: string }`

```js
lightshell.on('menu.click', (event) => {
  switch (event.id) {
    case 'file-new': createDocument(); break
    case 'file-open': openDocument(); break
    case 'file-save': saveDocument(); break
  }
})
```

---

### Tray Events

#### tray.click

Fired when a tray menu item is clicked. Equivalent to using `lightshell.tray.onClick()`.

**Data:** `{ id: string }`

```js
lightshell.on('tray.click', (event) => {
  if (event.id === 'quit') {
    lightshell.app.quit()
  }
})
```

---

### Shortcut Events

#### shortcut.{accelerator}

Fired when a registered global shortcut is triggered. The event name includes the accelerator string. These events are managed internally by `lightshell.shortcuts.register()` — you typically use the callback API rather than listening directly.

**Data:** none

```js
// These two are equivalent:
lightshell.shortcuts.register('CommandOrControl+Shift+P', () => {
  showPalette()
})

lightshell.on('shortcut.CommandOrControl+Shift+P', () => {
  showPalette()
})
```

---

### Updater Events

#### updater.available

Fired when a background update check detects a new version. This only fires for automatic background checks (controlled by the `interval` setting in `lightshell.json`), not for manual `lightshell.updater.check()` calls.

**Data:** `{ version: string, currentVersion: string, notes: string, pubDate: string }`

```js
lightshell.on('updater.available', (update) => {
  showUpdateBanner(`Version ${update.version} is available!`)
})
```

---

#### updater.progress

Fired during an update download to report progress. Equivalent to using `lightshell.updater.onProgress()`.

**Data:** `{ percent: number, bytesDownloaded: number, totalBytes: number }`

```js
lightshell.on('updater.progress', (p) => {
  progressBar.value = p.percent
})
```

---

## Common Patterns

### Global Event Logger

```js
// Debug helper — log all events
const events = [
  'window.resize', 'window.move', 'window.focus', 'window.blur',
  'menu.click', 'tray.click',
  'updater.available', 'updater.progress'
]

const unsubscribers = events.map(name =>
  lightshell.on(name, (data) => {
    console.log(`[${name}]`, data)
  })
)

// Stop all logging
function stopLogging() {
  unsubscribers.forEach(fn => fn())
}
```

### Event-Driven State Management

```js
const appState = {
  focused: true,
  windowSize: { width: 1024, height: 768 }
}

lightshell.on('window.focus', () => {
  appState.focused = true
  render()
})

lightshell.on('window.blur', () => {
  appState.focused = false
  render()
})

lightshell.on('window.resize', (size) => {
  appState.windowSize = size
  render()
})

function render() {
  const { width } = appState.windowSize
  document.body.className = width < 768 ? 'compact' : 'full'

  if (!appState.focused) {
    document.title = '(inactive) My App'
  } else {
    document.title = 'My App'
  }
}
```

### Respond to Update Availability

```js
lightshell.on('updater.available', async (update) => {
  // Show a non-intrusive banner
  const banner = document.getElementById('update-banner')
  banner.textContent = `Version ${update.version} is available — ${update.notes}`
  banner.style.display = 'block'

  banner.querySelector('.install-btn').addEventListener('click', async () => {
    banner.textContent = 'Installing update...'
    lightshell.on('updater.progress', (p) => {
      banner.textContent = `Installing update... ${p.percent}%`
    })
    await lightshell.updater.install()
  })
})
```

## Platform Notes

- Events are delivered asynchronously from the Go backend to JavaScript via the IPC bridge.
- Multiple listeners can be registered for the same event. They are called in the order they were registered.
- The unsubscribe function returned by `lightshell.on()` removes only that specific listener. Other listeners for the same event are unaffected.
- Window events (`resize`, `move`, `focus`, `blur`) also have dedicated helper methods on `lightshell.window` (e.g., `lightshell.window.onResize()`). Both approaches are equivalent.
- There is no `once()` helper in v1. To listen for a single event, unsubscribe inside the callback.
