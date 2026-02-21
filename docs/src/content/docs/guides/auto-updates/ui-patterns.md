---
title: Update UI Patterns
description: Common UI patterns for presenting updates to users.
---

How you present updates to your users matters. A forced update dialog on startup is annoying. A silent update that changes behavior without warning is confusing. This page shows three common patterns with complete code examples so you can pick the right one for your app.

## Pattern 1: Banner Notification

Check for updates silently on startup. If one is available, show a dismissible banner at the top of the app. The user can install when they are ready.

**Best for:** consumer apps, creative tools, any app where interrupting the user is unwelcome.

```html
<!-- Add this to your index.html -->
<div id="update-banner" style="display:none">
  <div id="update-banner-inner">
    <span id="update-message"></span>
    <button id="update-install-btn">Install</button>
    <button id="update-dismiss-btn">Later</button>
  </div>
</div>
```

```css
#update-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 9999;
  background: #1a73e8;
  color: white;
  padding: 8px 16px;
  font-size: 14px;
}
#update-banner-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: 800px;
  margin: 0 auto;
}
#update-message { flex: 1; }
#update-install-btn {
  background: white;
  color: #1a73e8;
  border: none;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
}
#update-dismiss-btn {
  background: transparent;
  color: white;
  border: 1px solid rgba(255,255,255,0.5);
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
}
```

```js
async function initUpdateBanner() {
  try {
    const update = await lightshell.updater.check()
    if (!update) return

    const banner = document.getElementById('update-banner')
    const message = document.getElementById('update-message')
    const installBtn = document.getElementById('update-install-btn')
    const dismissBtn = document.getElementById('update-dismiss-btn')

    message.textContent = `Version ${update.version} is available: ${update.notes}`
    banner.style.display = 'block'

    dismissBtn.addEventListener('click', () => {
      banner.style.display = 'none'
    })

    installBtn.addEventListener('click', async () => {
      installBtn.disabled = true
      installBtn.textContent = 'Installing...'

      lightshell.updater.onProgress((p) => {
        installBtn.textContent = `${p.percent}%`
      })

      try {
        await lightshell.updater.install()
        message.textContent = 'Update installed. Restart to apply.'
        installBtn.textContent = 'Restart'
        installBtn.disabled = false
        installBtn.onclick = () => lightshell.app.quit()
      } catch (err) {
        message.textContent = `Update failed: ${err.message}`
        installBtn.textContent = 'Retry'
        installBtn.disabled = false
      }
    })
  } catch (err) {
    console.warn('Update check failed:', err.message)
  }
}

// Check after the app has loaded and rendered
setTimeout(initUpdateBanner, 3000)
```

## Pattern 2: Confirm Dialog with Progress

Show a native confirmation dialog when an update is found. If the user confirms, show download progress in the app UI before prompting a restart.

**Best for:** productivity apps, tools used daily, apps where users expect to stay current.

```js
async function checkForUpdates() {
  try {
    const update = await lightshell.updater.check()
    if (!update) return

    const confirmed = await lightshell.dialog.confirm(
      'Update Available',
      `Version ${update.version} is available (you have ${update.currentVersion}).\n\n${update.notes}\n\nDownload and install?`
    )
    if (!confirmed) return

    // Show progress in the app
    const overlay = document.createElement('div')
    overlay.id = 'update-overlay'
    overlay.innerHTML = `
      <div class="update-dialog">
        <h3>Downloading Update</h3>
        <div class="progress-bar">
          <div class="progress-fill" id="progress-fill"></div>
        </div>
        <p id="progress-text">Starting download...</p>
      </div>
    `
    overlay.style.cssText = `
      position: fixed; inset: 0; z-index: 10000;
      background: rgba(0,0,0,0.5);
      display: flex; align-items: center; justify-content: center;
    `
    document.body.appendChild(overlay)

    const style = document.createElement('style')
    style.textContent = `
      .update-dialog {
        background: white; border-radius: 8px; padding: 24px;
        min-width: 320px; text-align: center;
      }
      .update-dialog h3 { margin: 0 0 16px; }
      .progress-bar {
        height: 8px; background: #e0e0e0; border-radius: 4px;
        overflow: hidden; margin: 12px 0;
      }
      .progress-fill {
        height: 100%; background: #1a73e8; border-radius: 4px;
        width: 0%; transition: width 0.3s;
      }
      #progress-text { color: #666; font-size: 13px; margin: 8px 0 0; }
    `
    document.head.appendChild(style)

    const fill = document.getElementById('progress-fill')
    const text = document.getElementById('progress-text')

    lightshell.updater.onProgress((p) => {
      fill.style.width = `${p.percent}%`
      const mb = (n) => (n / 1024 / 1024).toFixed(1)
      text.textContent = `${mb(p.bytesDownloaded)} MB / ${mb(p.totalBytes)} MB`
    })

    await lightshell.updater.install()

    // Update installed, prompt restart
    overlay.remove()
    const restart = await lightshell.dialog.confirm(
      'Update Installed',
      `Version ${update.version} has been installed. Restart now?`
    )
    if (restart) {
      await lightshell.app.quit()
    }
  } catch (err) {
    const overlay = document.getElementById('update-overlay')
    if (overlay) overlay.remove()
    console.error('Update failed:', err.message)
  }
}

setTimeout(checkForUpdates, 2000)
```

## Pattern 3: Silent Auto-Update

Check and install with no user interaction. The app updates itself and restarts. The user sees the new version next time they open the app.

**Best for:** internal tools, kiosk apps, apps where you control the deployment and want everyone on the latest version.

```js
// One line. No UI, no prompts.
async function autoUpdate() {
  try {
    await lightshell.updater.checkAndInstall()
    // If we get here, either no update was available,
    // or the update was installed successfully.
    // The new version takes effect on next launch.
  } catch (err) {
    // Silently ignore -- the app continues with the current version
    console.warn('Auto-update failed:', err.message)
  }
}

autoUpdate()
```

If you want the app to restart immediately after an update:

```js
async function autoUpdateAndRestart() {
  try {
    const update = await lightshell.updater.check()
    if (!update) return

    await lightshell.updater.install()
    await lightshell.app.quit()
  } catch (err) {
    console.warn('Auto-update failed:', err.message)
  }
}

autoUpdateAndRestart()
```

## Using Background Check Events

If you configured an `interval` in your updater config, LightShell checks for updates automatically on startup. You can listen for the result instead of calling `check()` yourself:

```js
lightshell.events.on('updater.available', (update) => {
  // Show a badge on a settings icon, update a status bar, etc.
  const badge = document.getElementById('update-badge')
  badge.textContent = `v${update.version}`
  badge.style.display = 'inline'

  badge.addEventListener('click', async () => {
    const confirmed = await lightshell.dialog.confirm(
      'Update Available',
      `Install version ${update.version}?`
    )
    if (confirmed) {
      await lightshell.updater.install()
    }
  })
})
```

This approach is less intrusive than checking on a timer. The event fires once per app session at most, and only if an update is actually available.

## Showing Release Notes

The `notes` field in the update object contains the release notes from your manifest. Display them so the user knows what they are getting:

```js
const update = await lightshell.updater.check()
if (update) {
  await lightshell.dialog.message(
    `What's New in ${update.version}`,
    update.notes
  )
}
```

For longer release notes, render them in your app UI instead of a dialog:

```js
const update = await lightshell.updater.check()
if (update) {
  const panel = document.getElementById('release-notes')
  panel.innerHTML = `
    <h3>Version ${update.version}</h3>
    <p>${update.notes}</p>
    <button onclick="installUpdate()">Install Now</button>
  `
  panel.style.display = 'block'
}
```

## Best Practices

**Never force-update without consent.** Even for internal tools, give users a way to dismiss or postpone. A forced update during an unsaved task loses trust.

**Do not check on every page load.** Check once on startup (with a delay) or rely on the background interval. Checking too often wastes bandwidth and slows down manifest servers.

**Show progress for large downloads.** Even a 5MB download takes a noticeable moment on slow connections. A progress bar tells the user something is happening.

**Preserve the user's work.** If your app has unsaved state, check for that before calling `install()` or `quit()`:

```js
if (hasUnsavedChanges()) {
  const save = await lightshell.dialog.confirm(
    'Unsaved Changes',
    'Save your work before updating?'
  )
  if (save) {
    await saveCurrentWork()
  }
}
await lightshell.updater.install()
```

**Handle offline gracefully.** If the update check fails because the user is offline, do not show an error. Just continue running the current version. Updates are a nice-to-have, not a requirement for the app to function.
