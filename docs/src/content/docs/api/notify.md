---
title: Notifications API
description: Complete reference for lightshell.notify — system notifications.
---

The `lightshell.notify` module sends native operating system notifications. Notifications appear in the system notification center and follow platform conventions for display and dismissal. All methods are async and return Promises.

## Methods

### send(title, body, options?)

Show a system notification with a title and body text.

**Parameters:**
- `title` (string) — the notification title, displayed prominently
- `body` (string) — the notification body text
- `options` (object, optional) — additional notification options:
  - `silent` (boolean) — suppress the notification sound (default: `false`)

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.notify.send('Download Complete', 'report.pdf has been saved.')
```

```js
// Silent notification
await lightshell.notify.send('Sync Complete', '14 files updated.', {
  silent: true
})
```

**Errors:** Rejects if the operating system denies notification permission or the notification system is unavailable.

---

## Common Patterns

### Task Completion Notification

```js
async function processFiles(files) {
  let processed = 0
  for (const file of files) {
    await processFile(file)
    processed++
  }

  await lightshell.notify.send(
    'Processing Complete',
    `${processed} file${processed === 1 ? '' : 's'} processed successfully.`
  )
}
```

### Background Event Alert

```js
// Notify when a long-running operation finishes while the window is not focused
async function runWithNotification(taskName, taskFn) {
  const result = await taskFn()

  // Only notify if the app is in the background
  lightshell.window.onBlur(() => {
    lightshell.notify.send(taskName, 'Task completed. Click to view results.')
  })

  return result
}
```

### Timer / Reminder

```js
function setReminder(message, delayMinutes) {
  const ms = delayMinutes * 60 * 1000
  setTimeout(async () => {
    await lightshell.notify.send('Reminder', message)
  }, ms)
}

// Usage
setReminder('Stand up and stretch!', 30)
```

## Platform Notes

- On macOS 11+, notifications use `UNUserNotificationCenter`. The older `NSUserNotification` API is deprecated and not used by LightShell.
- On macOS, the user must grant notification permission to the app. If permission is denied, `send()` will reject with an error.
- On Linux, notifications are delivered via `libnotify` / the desktop environment's notification daemon (e.g., GNOME, KDE).
- Notification appearance (duration, grouping, action buttons) is controlled by the operating system and cannot be customized from LightShell.
- The `options` parameter is reserved for future expansion (e.g., action buttons, icons). Currently only `silent` is supported.
- Notifications persist in the system notification center on macOS. On Linux, behavior depends on the desktop environment.
