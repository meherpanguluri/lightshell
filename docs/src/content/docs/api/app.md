---
title: App API
description: Complete reference for lightshell.app — application lifecycle, version, and data paths.
---

The `lightshell.app` module manages the application lifecycle, provides version information, and resolves app-specific data paths. All methods are async and return Promises.

## Methods

### quit()

Quit the application. This closes the window and terminates the process. If a tray icon is active, this still terminates the process entirely.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
document.getElementById('quit-btn').addEventListener('click', async () => {
  const confirmed = await lightshell.dialog.confirm('Quit', 'Are you sure?')
  if (confirmed) {
    await lightshell.app.quit()
  }
})
```

---

### version()

Get the application version as defined in `lightshell.json`.

**Parameters:** none

**Returns:** `Promise<string>` — the version string (e.g., `"1.0.0"`)

**Example:**
```js
const ver = await lightshell.app.version()
document.getElementById('version').textContent = `v${ver}`
```

---

### dataDir()

Get the persistent data directory for this application. This is a platform-appropriate location for storing user data, settings, and caches.

**Parameters:** none

**Returns:** `Promise<string>` — absolute path to the app's data directory

**Example:**
```js
const dataDir = await lightshell.app.dataDir()
console.log(dataDir)
// macOS: ~/Library/Application Support/com.example.myapp/
// Linux: ~/.config/com.example.myapp/
```

The directory path is derived from the `build.appId` field in `lightshell.json`. The directory is not created automatically — use `lightshell.fs.mkdir()` on first run if it does not exist.

---

## Common Patterns

### App Initialization

```js
async function initApp() {
  // Ensure data directory exists
  const dataDir = await lightshell.app.dataDir()
  await lightshell.fs.mkdir(dataDir)

  // Load or create settings
  const settingsPath = `${dataDir}/settings.json`
  let settings = { theme: 'light', recentFiles: [] }

  if (await lightshell.fs.exists(settingsPath)) {
    const raw = await lightshell.fs.readFile(settingsPath)
    settings = JSON.parse(raw)
  }

  return settings
}
```

### About Dialog

```js
async function showAbout() {
  const version = await lightshell.app.version()
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()

  await lightshell.dialog.message(
    'About My App',
    `Version: ${version}\nPlatform: ${platform} (${arch})`
  )
}
```

### Graceful Quit with Cleanup

```js
let hasUnsavedChanges = false

async function gracefulQuit() {
  if (hasUnsavedChanges) {
    const save = await lightshell.dialog.confirm(
      'Unsaved Changes',
      'You have unsaved changes. Save before quitting?'
    )
    if (save) {
      await saveCurrentDocument()
    }
  }

  // Persist application state
  const dataDir = await lightshell.app.dataDir()
  await lightshell.fs.writeFile(
    `${dataDir}/state.json`,
    JSON.stringify({
      lastOpen: Date.now(),
      windowSize: await lightshell.window.getSize(),
      windowPos: await lightshell.window.getPosition()
    })
  )

  await lightshell.app.quit()
}
```

### Version Check on Startup

```js
async function checkVersion() {
  const current = await lightshell.app.version()
  const dataDir = await lightshell.app.dataDir()
  const statePath = `${dataDir}/last-version.txt`

  let lastVersion = null
  if (await lightshell.fs.exists(statePath)) {
    lastVersion = await lightshell.fs.readFile(statePath)
  }

  if (lastVersion && lastVersion !== current) {
    await lightshell.dialog.message(
      'Updated',
      `Welcome to version ${current}! See what's new in the changelog.`
    )
  }

  await lightshell.fs.writeFile(statePath, current)
}
```

---

### setBadgeCount(count)

Set the badge count on the app's dock icon (macOS). This is used to indicate unread items, pending tasks, or other numeric notifications. Pass `0` to clear the badge.

**Parameters:**
- `count` (number) — the badge number to display. `0` clears the badge.

**Returns:** `Promise<void>`

**Example:**
```js
// Show 5 unread messages
await lightshell.app.setBadgeCount(5)

// Clear the badge
await lightshell.app.setBadgeCount(0)
```

**Example: Track Unread Count**
```js
let unreadCount = 0

function onNewMessage(message) {
  unreadCount++
  lightshell.app.setBadgeCount(unreadCount)
  showNotification(message)
}

function onMessagesRead() {
  unreadCount = 0
  lightshell.app.setBadgeCount(0)
}
```

**Platform Notes:**
- macOS: Displays a red badge with the number on the dock icon. Uses `NSApp.dockTile.badgeLabel`.
- Linux: Not supported on most desktop environments. The call is a no-op.

---

### onProtocol(callback)

Handle custom URL protocol opens. When the user opens a URL like `myapp://action/data` in their browser or another app, your app launches (or comes to the foreground) and the callback receives the full URL.

Requires the `protocols.schemes` field in `lightshell.json`:
```json
{
  "protocols": {
    "schemes": ["myapp"]
  }
}
```

**Parameters:**
- `callback` (function) — receives the full URL string (e.g., `"myapp://open/doc/123"`)

**Returns:** unsubscribe function

**Example:**
```js
const unsubscribe = lightshell.app.onProtocol((url) => {
  console.log('Received URL:', url)
  const parsed = new URL(url)

  switch (parsed.hostname) {
    case 'open':
      openDocument(parsed.pathname.slice(1))
      break
    case 'settings':
      showSettings(parsed.searchParams.get('tab'))
      break
  }
})
```

**Example: OAuth Callback**
```js
lightshell.app.onProtocol(async (url) => {
  const parsed = new URL(url)
  if (parsed.hostname === 'auth' && parsed.pathname === '/callback') {
    const code = parsed.searchParams.get('code')
    await exchangeCodeForToken(code)
  }
})
```

See the [Deep Linking](/guides/deep-linking/) guide for detailed examples and platform-specific behavior.

---

### onSecondInstance(callback)

Handle duplicate app launches. When the user tries to open a second instance of your app (e.g., double-clicking the app icon while it is already running), the existing instance receives this callback instead of a second window opening.

**Parameters:**
- `callback` (function) — receives `{ args: string[], cwd: string }` where `args` is the command-line arguments from the second launch and `cwd` is its working directory

**Returns:** unsubscribe function

**Example:**
```js
lightshell.app.onSecondInstance(async ({ args, cwd }) => {
  // Bring the existing window to front
  await lightshell.window.restore()

  // If launched with a file argument, open it
  if (args.length > 1) {
    const filePath = args[1]
    const content = await lightshell.fs.readFile(filePath)
    openInEditor(content)
  }
})
```

**Example: Single Instance with File Association**
```js
lightshell.app.onSecondInstance(async ({ args }) => {
  await lightshell.window.restore()

  // Filter for file paths in arguments
  const files = args.slice(1).filter(arg => !arg.startsWith('-'))
  for (const file of files) {
    if (await lightshell.fs.exists(file)) {
      openFile(file)
    }
  }
})
```

**Note:** Single-instance behavior is the default in production builds. When the user opens the app while it is already running, the first instance receives the `onSecondInstance` event and the second process exits. In development (`lightshell dev`), multiple instances are allowed.

## Platform Notes

- On macOS, `dataDir()` resolves to `~/Library/Application Support/{appId}/`
- On Linux, `dataDir()` resolves to `~/.config/{appId}/`
- The `appId` is read from `build.appId` in `lightshell.json` (e.g., `"com.example.myapp"`)
- `quit()` terminates the entire process, including any background tasks or tray icons
- `version()` reads the `version` field from `lightshell.json` at build time — it is baked into the binary
- `setBadgeCount()` only works on macOS. On Linux it is a no-op.
- `onProtocol()` requires `protocols.schemes` in `lightshell.json` and a built app (`.app` or packaged format)
- `onSecondInstance()` only applies to production builds. During development, multiple instances can run simultaneously.
