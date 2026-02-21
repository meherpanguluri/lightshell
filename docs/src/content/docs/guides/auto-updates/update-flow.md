---
title: Update Flow
description: How LightShell downloads and installs updates.
---

This page documents the internal flow of the LightShell auto-updater, from the initial version check through binary replacement and restart. Understanding this flow helps you design the right update experience for your app.

## Overview

The update process has two distinct phases: **check** and **install**. They are separate API calls so you can check for updates without installing, show the user information about the update, and let them decide when to proceed.

```
check()                          install()
  │                                │
  ├─ Fetch manifest                ├─ Download archive
  ├─ Parse JSON                    ├─ Verify SHA256
  ├─ Compare versions              ├─ Extract binary
  └─ Return update info            ├─ Replace running binary
                                   └─ Restart (optional)
```

## Phase 1: Check

When your code calls `lightshell.updater.check()`, here is what happens on the Go side:

**1. Fetch the manifest.** The Go backend makes an HTTP GET request to the endpoint URL configured in `lightshell.json`. This request goes through Go's `net/http` client, not the webview, so it is not subject to CORS or CSP restrictions.

**2. Parse the JSON.** The response body is parsed into the manifest structure. If the JSON is malformed, the check returns an error.

**3. Find the platform entry.** The updater constructs a platform key from `runtime.GOOS` and `runtime.GOARCH` (e.g., `darwin-arm64`, `linux-x64`) and looks it up in the manifest's `platforms` map. If the current platform is not in the manifest, the check returns `null` -- no update available for this platform.

**4. Compare versions.** The manifest version is compared to the app's built-in version (from `lightshell.json` at compile time) using semver rules. If the manifest version is greater, an update is available.

**5. Return the result.** If an update is available, the check returns an object with the update details:

```js
const update = await lightshell.updater.check()
// Returns null if no update, or:
// {
//   version: "1.2.0",
//   currentVersion: "1.1.0",
//   notes: "Bug fixes and performance improvements",
//   pubDate: "2025-07-15T00:00:00Z"
// }
```

The check does not download anything beyond the manifest (a few hundred bytes). It is safe to call frequently.

## Phase 2: Install

When your code calls `lightshell.updater.install()`, the Go backend performs the download and replacement:

**1. Download the archive.** The Go backend downloads the archive from the platform-specific URL in the manifest. The download goes to a temporary directory (`$TEMP/lightshell-update-{random}/`). During download, progress events are emitted to JavaScript:

```js
lightshell.updater.onProgress((p) => {
  console.log(`${p.percent}% (${p.bytesDownloaded}/${p.totalBytes})`)
})

await lightshell.updater.install()
```

**2. Verify SHA256.** After the download completes, the Go backend computes the SHA256 hash of the downloaded file and compares it to the hash in the manifest. If they do not match, the downloaded file is deleted, the temp directory is cleaned up, and an error is returned. The existing app binary is never touched.

```
Downloaded file SHA256: a3f2b8c1d4e5...
Manifest SHA256:        a3f2b8c1d4e5...
→ Match: proceed with installation

Downloaded file SHA256: ffffffffffffffff...
Manifest SHA256:        a3f2b8c1d4e5...
→ Mismatch: reject update, delete download, return error
```

**3. Extract the archive.** The `.tar.gz` archive is extracted to the temp directory. The updater expects to find the app binary inside.

**4. Replace the running binary.** This step is platform-specific:

- **macOS:** The updater replaces the contents of the `.app` bundle. The new binary is written to `.app/Contents/MacOS/`, and updated resources (if any) are placed in `.app/Contents/Resources/`. The replacement uses a rename operation for atomicity -- the old binary is moved to a backup path, the new binary is moved into place, and then the backup is deleted.

- **Linux (AppImage):** The updater replaces the AppImage file. The old file is renamed to a backup, the new file is moved into place, and the backup is deleted.

- **Linux (installed via deb/rpm):** The updater replaces the binary at `/usr/bin/{appname}`. This may require appropriate file permissions.

**5. Clean up.** The temp directory and downloaded archive are deleted regardless of success or failure.

**6. Restart (optional).** After a successful binary replacement, the updater can restart the app. By default, it does not restart automatically -- it returns successfully and your code decides what to do:

```js
await lightshell.updater.install()
// Update installed successfully

const restart = await lightshell.dialog.confirm(
  'Update Installed',
  'Restart now to use the new version?'
)

if (restart) {
  // The app quits and the OS relaunches it
  await lightshell.app.quit()
}
```

The new version takes effect the next time the app starts.

## Background Check Flow

If you configure an `interval` in the updater config, LightShell runs automatic checks:

**1. On app startup,** the updater checks when the last update check occurred (stored in the app data directory).

**2. If the interval has elapsed,** it performs a check in the background (same as calling `lightshell.updater.check()` but initiated by the runtime, not by your code).

**3. If an update is available,** the runtime emits an `updater.available` event to JavaScript. It does not download or install anything.

```js
lightshell.events.on('updater.available', (update) => {
  // update = { version: "1.2.0", notes: "...", currentVersion: "1.1.0" }
  showUpdateBanner(update)
})
```

**4. The timestamp of the last check is recorded** so the next check happens after the configured interval.

Background checks run on a separate goroutine and do not block app startup. Network failures are silently ignored -- the check will be retried at the next startup.

## The checkAndInstall Shortcut

For apps that want the simplest possible integration, `checkAndInstall()` combines both phases:

```js
await lightshell.updater.checkAndInstall()
```

This calls `check()` and, if an update is available, immediately calls `install()`. If no update is available, it returns without doing anything. This is useful for internal tools where you always want the latest version without user interaction.

## Error Handling

The updater can fail at several points. All errors are returned as rejected promises:

| Error | Phase | Cause |
|-------|-------|-------|
| Network error | Check | Cannot reach the manifest endpoint |
| Invalid manifest | Check | JSON is malformed or missing required fields |
| Platform not found | Check | Current platform not in manifest (returns null, not error) |
| Download failed | Install | Network error during archive download |
| SHA256 mismatch | Install | Downloaded file does not match expected hash |
| Extract failed | Install | Archive is corrupted or in wrong format |
| Replace failed | Install | Cannot write to the binary path (permissions issue) |
| HTTP in production | Check | Manifest URL or download URL uses HTTP instead of HTTPS |

Always wrap update calls in try/catch:

```js
try {
  const update = await lightshell.updater.check()
  if (update) {
    await lightshell.updater.install()
  }
} catch (err) {
  console.error('Update failed:', err.message)
  // The app continues running with the current version
}
```

On any failure during install, the existing binary is preserved. The updater never leaves the app in a broken state.

## Temp Directory Cleanup

The updater creates a temporary directory for each install attempt. This directory is always cleaned up:

- On successful install: deleted after binary replacement
- On failed install: deleted immediately after the error
- On app crash during install: cleaned up on next app startup (the updater checks for stale temp directories)

No disk space is permanently consumed by failed updates.
