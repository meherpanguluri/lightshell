---
title: Tray API
description: Complete reference for lightshell.tray — system tray icons and menus.
---

The `lightshell.tray` module manages a system tray icon (menu bar on macOS, system tray on Linux). Use it to keep your app accessible when the window is closed, show status information, or provide quick actions via a context menu. All methods are async and return Promises.

## Methods

### set(options)

Create or update the system tray icon. If a tray icon already exists, calling `set()` again replaces it.

**Parameters:**
- `options` (object):
  - `icon` (string) — path to the tray icon image. Should be a PNG file, ideally 22x22 pixels (44x44 for Retina). Supports `$RESOURCE` path variable to reference bundled assets.
  - `tooltip` (string, optional) — text shown when the user hovers over the tray icon
  - `menu` (array, optional) — array of menu items for the tray's context menu

**Menu item properties:**

| Property | Type | Description |
|----------|------|-------------|
| `label` | string | The text displayed in the menu item |
| `id` | string | Unique identifier, emitted in click events |
| `enabled` | boolean | Whether the item is clickable (default: `true`) |
| `checked` | boolean | Show a checkmark next to the item (default: `false`) |
| `type` | string | `"normal"`, `"separator"`, or `"checkbox"` (default: `"normal"`) |

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.tray.set({
  icon: '$RESOURCE/tray-icon.png',
  tooltip: 'My App — Running',
  menu: [
    { label: 'Show Window', id: 'show' },
    { label: 'Status: Online', id: 'status', enabled: false },
    { type: 'separator' },
    { label: 'Start at Login', id: 'autostart', type: 'checkbox', checked: false },
    { type: 'separator' },
    { label: 'Quit', id: 'quit' }
  ]
})
```

---

### remove()

Remove the tray icon from the system tray. After calling this, the tray icon is no longer visible and no tray events will fire.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.tray.remove()
```

---

### onClick(callback)

Listen for clicks on tray menu items. The callback receives the `id` of the clicked menu item.

**Parameters:**
- `callback` (function) — receives an object with `{ id: string }` identifying which menu item was clicked

**Returns:** unsubscribe function

**Example:**
```js
const unsubscribe = lightshell.tray.onClick(async (event) => {
  switch (event.id) {
    case 'show':
      await lightshell.window.restore()
      break
    case 'autostart':
      toggleAutoStart()
      break
    case 'quit':
      await lightshell.app.quit()
      break
  }
})

// Later, stop listening
unsubscribe()
```

---

## Common Patterns

### Background App with Tray

Keep your app running in the background with a tray icon when the user closes the window.

```js
async function setupBackgroundMode() {
  // Set up the tray icon
  await lightshell.tray.set({
    icon: '$RESOURCE/tray-icon.png',
    tooltip: 'My App',
    menu: [
      { label: 'Open My App', id: 'open' },
      { type: 'separator' },
      { label: 'Quit', id: 'quit' }
    ]
  })

  // Handle tray menu clicks
  lightshell.tray.onClick(async (event) => {
    if (event.id === 'open') {
      await lightshell.window.restore()
    } else if (event.id === 'quit') {
      await lightshell.tray.remove()
      await lightshell.app.quit()
    }
  })
}
```

### Dynamic Tray Updates

Update the tray icon and menu based on application state.

```js
async function updateTrayStatus(isConnected) {
  const status = isConnected ? 'Online' : 'Offline'
  const icon = isConnected
    ? '$RESOURCE/tray-online.png'
    : '$RESOURCE/tray-offline.png'

  await lightshell.tray.set({
    icon,
    tooltip: `My App — ${status}`,
    menu: [
      { label: `Status: ${status}`, id: 'status', enabled: false },
      { type: 'separator' },
      { label: 'Show Window', id: 'show' },
      { label: 'Quit', id: 'quit' }
    ]
  })
}
```

### Tray with Notification Count

```js
let unreadCount = 0

async function updateTrayBadge(count) {
  unreadCount = count
  const label = count > 0 ? `My App (${count} new)` : 'My App'

  await lightshell.tray.set({
    icon: '$RESOURCE/tray-icon.png',
    tooltip: label,
    menu: [
      { label: count > 0 ? `${count} unread messages` : 'No new messages', id: 'info', enabled: false },
      { type: 'separator' },
      { label: 'Open', id: 'open' },
      { label: 'Mark All Read', id: 'mark-read', enabled: count > 0 },
      { type: 'separator' },
      { label: 'Quit', id: 'quit' }
    ]
  })
}
```

## Platform Notes

- On macOS, the tray icon appears in the menu bar (top-right of the screen). Use a template image (monochrome PNG with transparency) for proper light/dark mode support.
- On Linux, the tray icon appears in the system tray area, which varies by desktop environment (GNOME, KDE, XFCE).
- Tray icon images should be PNG format. Recommended size is 22x22 pixels (44x44 for Retina/HiDPI).
- The `$RESOURCE` path variable resolves to the app bundle's resources directory, making it easy to reference bundled icon files.
- Only one tray icon is supported per application. Calling `set()` multiple times replaces the previous tray icon.
- The tray icon persists across window close/restore cycles. Call `remove()` explicitly before `quit()` if you want a clean shutdown.
