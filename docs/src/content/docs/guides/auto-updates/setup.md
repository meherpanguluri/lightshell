---
title: Update Setup
description: Configure auto-updates in your LightShell app.
---

This guide walks through the complete setup for adding auto-updates to your LightShell app. By the end, your app will check for updates on startup and prompt the user to install them.

## Step 1: Add Updater Config

Open your `lightshell.json` and add the `updater` section:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "width": 900,
    "height": 600,
    "title": "My App"
  },
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.example.com/latest.json",
    "interval": "24h"
  }
}
```

### Config Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Set to `true` to enable the updater. |
| `endpoint` | string | Yes | URL to your update manifest JSON file. Must be HTTPS in production. |
| `interval` | string | No | How often to check for updates in the background. Examples: `"1h"`, `"12h"`, `"24h"`, `"7d"`. Defaults to `"24h"`. Set to `"0"` to disable background checks. |

The `endpoint` is the only URL the updater fetches on its own. The actual binary download URL comes from the manifest itself, so it can be hosted anywhere.

## Step 2: Create the Update Manifest

Create a `latest.json` file on your server. This tells LightShell what the latest version is and where to download it.

```json
{
  "version": "1.1.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-08-01T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.example.com/v1.1.0/my-app-darwin-arm64.tar.gz",
      "sha256": "a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678"
    },
    "darwin-x64": {
      "url": "https://releases.example.com/v1.1.0/my-app-darwin-x64.tar.gz",
      "sha256": "b4e3c9d2e5f67890abcdef1234567890abcdef1234567890abcdef12345679"
    },
    "linux-x64": {
      "url": "https://releases.example.com/v1.1.0/my-app-linux-x64.tar.gz",
      "sha256": "c5f4d0e3f6a78901bcdef01234567890abcdef1234567890abcdef1234567a"
    }
  }
}
```

See [Update Manifest](/guides/auto-updates/update-manifest/) for the full format specification.

## Step 3: Add Update Check to Your App

Add an update check to your app's startup code. Here is a complete `src/app.js` example:

```js
// Check for updates on startup
async function checkForUpdates() {
  try {
    const update = await lightshell.updater.check()
    if (!update) return // no update available

    const confirmed = await lightshell.dialog.confirm(
      'Update Available',
      `Version ${update.version} is available (you have ${update.currentVersion}).\n\n${update.notes}\n\nInstall now?`
    )

    if (confirmed) {
      await lightshell.updater.install()
    }
  } catch (err) {
    // Don't block the app if update check fails
    console.warn('Update check failed:', err.message)
  }
}

// Run after a short delay so the app loads first
setTimeout(checkForUpdates, 2000)
```

This pattern is deliberate: the update check runs after the app has loaded and rendered, so the user is never staring at a blank screen waiting for a network request. The `try/catch` ensures a failed update check (no network, server down, etc.) does not break the app.

## Step 4: Test Locally

During development, the updater allows HTTP endpoints so you can test without deploying. Start a local server that serves your manifest:

```bash
# Create a test manifest
cat > /tmp/latest.json << 'EOF'
{
  "version": "99.0.0",
  "notes": "Test update",
  "pub_date": "2025-01-01T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "http://localhost:8080/my-app-darwin-arm64.tar.gz",
      "sha256": "abc123"
    }
  }
}
EOF

# Serve it
cd /tmp && python3 -m http.server 8080
```

Set your endpoint to `http://localhost:8080/latest.json` in `lightshell.json` and run `lightshell dev`. The `check()` call should return the test update info. The `install()` call will fail because the SHA256 will not match, which is expected and confirms that verification works.

## Step 5: Deploy

When you are ready to ship updates:

1. Build your app: `lightshell build`
2. Compress the binary into a `.tar.gz` archive
3. Generate the SHA256 hash: `shasum -a 256 my-app-darwin-arm64.tar.gz`
4. Update your `latest.json` with the new version, URL, and hash
5. Upload both the archive and the manifest to your server

See [Hosting Releases](/guides/auto-updates/hosting-releases/) for detailed hosting options including GitHub Releases and S3.

## Background Checks

If you set an `interval` in the updater config, LightShell checks for updates automatically on app startup when the interval has elapsed since the last check. If an update is found, it emits an `updater.available` event to your JavaScript code:

```js
lightshell.events.on('updater.available', (update) => {
  // Show a non-intrusive banner or badge
  showUpdateBanner(update.version, update.notes)
})
```

Background checks never auto-install. They only notify your code that an update exists. You decide how and when to present it to the user.

## Complete Example

Here is a minimal but complete app with auto-updates:

**lightshell.json:**
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

**src/index.html:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>My App</title>
</head>
<body>
  <h1>My App</h1>
  <p id="status">Ready</p>
  <script src="app.js"></script>
</body>
</html>
```

**src/app.js:**
```js
async function checkForUpdates() {
  try {
    const update = await lightshell.updater.check()
    if (!update) return

    const confirmed = await lightshell.dialog.confirm(
      'Update Available',
      `Version ${update.version} is ready. Install now?`
    )

    if (confirmed) {
      document.getElementById('status').textContent = 'Installing update...'

      lightshell.updater.onProgress((p) => {
        document.getElementById('status').textContent =
          `Downloading: ${p.percent}%`
      })

      await lightshell.updater.install()
    }
  } catch (err) {
    console.warn('Update check failed:', err.message)
  }
}

setTimeout(checkForUpdates, 2000)
```
