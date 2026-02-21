---
title: System & App API
description: Complete reference for lightshell.system and lightshell.app — OS info, paths, and app lifecycle.
---

The `lightshell.system` module provides information about the operating system. The `lightshell.app` module manages the application lifecycle and provides app-specific paths. All methods are async and return Promises.

## lightshell.system

### platform()

Get the current operating system identifier.

**Parameters:** none

**Returns:** `Promise<string>` — `"darwin"` on macOS, `"linux"` on Linux

**Example:**
```js
const os = await lightshell.system.platform()
if (os === 'darwin') {
  console.log('Running on macOS')
} else if (os === 'linux') {
  console.log('Running on Linux')
}
```

---

### arch()

Get the CPU architecture.

**Parameters:** none

**Returns:** `Promise<string>` — `"arm64"` or `"amd64"` (also known as x64)

**Example:**
```js
const arch = await lightshell.system.arch()
console.log(`Architecture: ${arch}`) // "arm64" or "amd64"
```

---

### homeDir()

Get the current user's home directory path.

**Parameters:** none

**Returns:** `Promise<string>` — absolute path to the home directory

**Example:**
```js
const home = await lightshell.system.homeDir()
console.log(home) // "/Users/alice" on macOS, "/home/alice" on Linux
```

---

### tempDir()

Get the system temporary directory path.

**Parameters:** none

**Returns:** `Promise<string>` — absolute path to the temp directory

**Example:**
```js
const tmp = await lightshell.system.tempDir()
await lightshell.fs.writeFile(`${tmp}/scratch.txt`, 'temporary data')
```

---

### hostname()

Get the system hostname.

**Parameters:** none

**Returns:** `Promise<string>` — the hostname of the machine

**Example:**
```js
const name = await lightshell.system.hostname()
console.log(`Running on: ${name}`)
```

---

## lightshell.app

### quit()

Quit the application. This closes the window and terminates the process.

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
// Linux: ~/.local/share/com.example.myapp/
```

The directory is derived from the `build.appId` in `lightshell.json`. It is not created automatically — use `lightshell.fs.mkdir()` on first run.

---

## lightshell.shell

### open(url)

Open a URL in the user's default browser, or a file path in the default application.

**Parameters:**
- `url` (string) — a URL (e.g., `https://example.com`) or file path

**Returns:** `Promise<void>`

**Example:**
```js
// Open a website in the default browser
await lightshell.shell.open('https://lightshell.sh')

// Open a file in the default application
await lightshell.shell.open('/Users/me/document.pdf')
```

**Important:** Do not use `window.open()` for external URLs — it will try to navigate the webview. Always use `lightshell.shell.open()` to launch the system browser.

---

## lightshell.notify

### send(title, body, options?)

Show a system notification.

**Parameters:**
- `title` (string) — notification title
- `body` (string) — notification body text
- `options` (object, optional) — additional options (reserved for future use)

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.notify.send('Download Complete', 'report.pdf has been saved.')
```

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

### Platform-Specific Behavior

```js
async function getDefaultSavePath(filename) {
  const platform = await lightshell.system.platform()
  const home = await lightshell.system.homeDir()

  if (platform === 'darwin') {
    return `${home}/Documents/${filename}`
  } else {
    return `${home}/${filename}`
  }
}
```

### External Links

```js
// Make all <a> tags with href starting with http open in system browser
document.addEventListener('click', (e) => {
  const link = e.target.closest('a[href^="http"]')
  if (link) {
    e.preventDefault()
    lightshell.shell.open(link.href)
  }
})
```

### Quit with Cleanup

```js
async function gracefulQuit() {
  // Save state
  const dataDir = await lightshell.app.dataDir()
  await lightshell.fs.writeFile(
    `${dataDir}/state.json`,
    JSON.stringify({ lastOpen: Date.now() })
  )

  // Quit
  await lightshell.app.quit()
}
```
