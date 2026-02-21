---
title: Updater API
description: Complete reference for lightshell.updater — auto-update your app.
---

The `lightshell.updater` module checks for and installs application updates. It uses a simple JSON manifest hosted at an HTTPS endpoint, with mandatory SHA256 verification for all downloads. No Sparkle, no electron-updater complexity — just a manifest, a hash check, and a binary replacement. All methods are async and return Promises.

## Setup

Before using the updater API, configure the update endpoint in `lightshell.json`:

```json
{
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.myapp.com/latest.json",
    "interval": "24h"
  }
}
```

The endpoint must serve a JSON manifest:

```json
{
  "version": "1.2.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-07-15T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-darwin-arm64.tar.gz",
      "sha256": "a1b2c3d4e5f6..."
    },
    "darwin-amd64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-darwin-amd64.tar.gz",
      "sha256": "e5f6g7h8i9j0..."
    },
    "linux-amd64": {
      "url": "https://releases.myapp.com/v1.2.0/myapp-linux-amd64.tar.gz",
      "sha256": "i9j0k1l2m3n4..."
    }
  }
}
```

Platform keys follow the format `{GOOS}-{GOARCH}`: `darwin-arm64`, `darwin-amd64`, `linux-amd64`.

## Methods

### check()

Check for available updates by fetching the manifest from the configured endpoint. Compares the manifest version to the current app version using semver.

**Parameters:** none

**Returns:** `Promise<UpdateInfo | null>` — update info if an update is available, or `null` if the app is up to date

The `UpdateInfo` object:
- `version` (string) — the new version available (e.g., `"1.2.0"`)
- `currentVersion` (string) — the currently running version (e.g., `"1.1.0"`)
- `notes` (string) — release notes from the manifest
- `pubDate` (string) — publication date as an ISO 8601 string

**Example:**
```js
const update = await lightshell.updater.check()
if (update) {
  console.log(`Update available: v${update.version}`)
  console.log(`Current version: v${update.currentVersion}`)
  console.log(`Release notes: ${update.notes}`)
} else {
  console.log('App is up to date')
}
```

---

### install()

Download and install the latest update. This downloads the update archive, verifies the SHA256 hash against the manifest, extracts it, and replaces the current binary. The app will prompt for a restart after installation.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.updater.install()
// App will prompt for restart
```

**Errors:** Rejects if no update is available (call `check()` first), if the download fails, or if the SHA256 hash does not match the manifest.

---

### checkAndInstall()

Convenience method that checks for an update and installs it if available. No-op if no update is available.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
// One-liner for simple apps
await lightshell.updater.checkAndInstall()
```

---

### onProgress(callback)

Listen for download progress events during `install()` or `checkAndInstall()`.

**Parameters:**
- `callback` (function) — receives a progress object:
  - `percent` (number) — download progress from 0 to 100
  - `bytesDownloaded` (number) — bytes downloaded so far
  - `totalBytes` (number) — total file size in bytes

**Returns:** unsubscribe function

**Example:**
```js
const unsubscribe = lightshell.updater.onProgress((p) => {
  console.log(`${p.percent}% — ${p.bytesDownloaded}/${p.totalBytes} bytes`)
})

await lightshell.updater.install()
unsubscribe()
```

---

## Common Patterns

### Check on Startup

```js
async function checkForUpdatesOnStartup() {
  try {
    const update = await lightshell.updater.check()
    if (update) {
      const install = await lightshell.dialog.confirm(
        'Update Available',
        `Version ${update.version} is available (you have ${update.currentVersion}).\n\n${update.notes}\n\nWould you like to update now?`
      )
      if (install) {
        await lightshell.updater.install()
      }
    }
  } catch (err) {
    // Don't block app startup on update check failure
    console.warn('Update check failed:', err.message)
  }
}

// Run after app initializes
checkForUpdatesOnStartup()
```

### Update Dialog with Progress Bar

```js
async function showUpdateDialog(update) {
  const install = await lightshell.dialog.confirm(
    `Update to v${update.version}`,
    `${update.notes}\n\nDownload and install now?`
  )
  if (!install) return

  const progressEl = document.getElementById('update-progress')
  const statusEl = document.getElementById('update-status')

  progressEl.style.display = 'block'
  statusEl.textContent = 'Downloading update...'

  lightshell.updater.onProgress((p) => {
    progressEl.value = p.percent
    const mb = (p.bytesDownloaded / 1048576).toFixed(1)
    const totalMb = (p.totalBytes / 1048576).toFixed(1)
    statusEl.textContent = `Downloading: ${mb} MB / ${totalMb} MB (${p.percent}%)`
  })

  try {
    await lightshell.updater.install()
    statusEl.textContent = 'Update installed. Restarting...'
  } catch (err) {
    statusEl.textContent = `Update failed: ${err.message}`
    progressEl.style.display = 'none'
  }
}
```

### Menu-Driven Update Check

```js
// Add "Check for Updates" to your Help menu
lightshell.on('menu.click', async (event) => {
  if (event.id === 'check-updates') {
    const update = await lightshell.updater.check()
    if (update) {
      await showUpdateDialog(update)
    } else {
      await lightshell.dialog.message(
        'No Updates',
        'You are running the latest version.'
      )
    }
  }
})
```

## Security

- **SHA256 verification is mandatory.** Every downloaded archive is verified against the hash in the manifest. If the hash does not match, the update is rejected and an error is returned. There is no way to skip this check.
- **HTTPS is required in production builds.** The update endpoint must use HTTPS. HTTP endpoints are allowed in development mode for local testing but are rejected in production.
- **The manifest itself should be served over HTTPS** to prevent man-in-the-middle attacks from modifying version info or SHA256 hashes.

## Platform Notes

- On macOS, the updater replaces the binary inside the `.app/Contents/MacOS/` directory. The app bundle structure is preserved.
- On Linux, the updater replaces the AppImage file or the binary at its installed location (e.g., `/usr/bin/myapp`).
- The `interval` config (`"24h"`, `"12h"`, `"1h"`) controls automatic background checks. When the interval has elapsed since the last check, the app emits an `updater.available` event to JavaScript on startup. Background checks never auto-install.
- Delta/differential updates are not supported in v1. The entire binary is re-downloaded.
- The update archive format is `.tar.gz`. The updater extracts it to a temporary directory, verifies the hash, then performs the binary replacement.
