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

## Platform Notes

- On macOS, `dataDir()` resolves to `~/Library/Application Support/{appId}/`
- On Linux, `dataDir()` resolves to `~/.config/{appId}/`
- The `appId` is read from `build.appId` in `lightshell.json` (e.g., `"com.example.myapp"`)
- `quit()` terminates the entire process, including any background tasks or tray icons
- `version()` reads the `version` field from `lightshell.json` at build time — it is baked into the binary
