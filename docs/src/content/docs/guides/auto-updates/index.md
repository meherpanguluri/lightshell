---
title: Auto-Updates Overview
description: Add automatic updates to your LightShell app.
---

LightShell includes a built-in auto-update system. It is simple by design: you host a JSON manifest, your app checks for updates, and LightShell handles downloading, verifying, and replacing the binary. No Sparkle, no electron-updater, no complex infrastructure.

## How It Works

The update system has three parts:

1. **A JSON manifest** you host on any static server (GitHub Releases, S3, your own CDN)
2. **A config entry** in your `lightshell.json` pointing to the manifest
3. **A JS API call** in your app to check for and install updates

When your app calls `lightshell.updater.check()`, the Go backend fetches your manifest, compares the version to the running app version using semver, and returns update info if a newer version is available. When you call `lightshell.updater.install()`, it downloads the archive, verifies the SHA256 hash, extracts the new binary, and replaces the running one.

```
Your Server                        LightShell App
  │                                    │
  │   GET /latest.json                 │
  │<───────────────────────────────────┤  lightshell.updater.check()
  │                                    │
  │   { version, sha256, url }         │
  ├───────────────────────────────────>│
  │                                    │  compare semver
  │                                    │  → update available
  │                                    │
  │   GET /myapp-darwin-arm64.tar.gz   │
  │<───────────────────────────────────┤  lightshell.updater.install()
  │                                    │
  │   binary archive                   │
  ├───────────────────────────────────>│
  │                                    │  verify SHA256
  │                                    │  replace binary
  │                                    │  restart (optional)
```

## Quick Start

**1. Add updater config to `lightshell.json`:**

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.example.com/latest.json",
    "interval": "24h"
  }
}
```

**2. Host a manifest file** at your endpoint URL:

```json
{
  "version": "1.1.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-08-01T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.example.com/v1.1.0/my-app-darwin-arm64.tar.gz",
      "sha256": "a1b2c3d4e5f6..."
    },
    "linux-x64": {
      "url": "https://releases.example.com/v1.1.0/my-app-linux-x64.tar.gz",
      "sha256": "f6e5d4c3b2a1..."
    }
  }
}
```

**3. Check for updates in your app:**

```js
const update = await lightshell.updater.check()
if (update) {
  const confirmed = await lightshell.dialog.confirm(
    'Update Available',
    `Version ${update.version} is available. Install now?`
  )
  if (confirmed) {
    await lightshell.updater.install()
  }
}
```

That is the entire integration. Three steps, no external tools, no package manager plugins.

## Security

Every update is verified with a SHA256 hash before installation. If the hash in the manifest does not match the downloaded archive, the update is rejected and the existing binary is preserved. The manifest endpoint must use HTTPS in production builds (HTTP is allowed in dev mode for local testing).

See [Update Security](/guides/auto-updates/security/) for the full security model.

## What Gets Updated

LightShell replaces the app binary itself. On macOS, it replaces the executable inside `.app/Contents/MacOS/`. On Linux, it replaces the AppImage or the binary in `/usr/bin/`. Your app data (stored via `lightshell.store` or in `$APP_DATA`) is never touched during an update.

## Detailed Guides

- [Update Setup](/guides/auto-updates/setup/) -- configure auto-updates step by step
- [Hosting Releases](/guides/auto-updates/hosting-releases/) -- where and how to host your update files
- [Update Manifest](/guides/auto-updates/update-manifest/) -- the JSON manifest format in detail
- [Update Flow](/guides/auto-updates/update-flow/) -- how downloading and installing works internally
- [Update Security](/guides/auto-updates/security/) -- SHA256 verification and HTTPS requirements
- [Update UI Patterns](/guides/auto-updates/ui-patterns/) -- common patterns for presenting updates to users
