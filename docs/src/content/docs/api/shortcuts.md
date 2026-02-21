---
title: Shortcuts API
description: Complete reference for lightshell.shortcuts — global keyboard shortcuts.
---

The `lightshell.shortcuts` module registers global keyboard shortcuts that work even when the application window is not focused. This is essential for productivity tools, clipboard managers, quick-launch utilities, and screenshot tools. All methods are async and return Promises.

## Methods

### register(accelerator, callback)

Register a global keyboard shortcut. When the user presses the key combination anywhere on their system, the callback is invoked and the app receives focus.

**Parameters:**
- `accelerator` (string) — the keyboard shortcut string (see modifier table below)
- `callback` (function) — called when the shortcut is triggered

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.shortcuts.register('CommandOrControl+Shift+P', () => {
  console.log('Command palette triggered!')
  showCommandPalette()
})

await lightshell.shortcuts.register('CommandOrControl+Shift+C', async () => {
  const text = await lightshell.clipboard.read()
  processClipboardText(text)
})
```

**Errors:** Rejects if the shortcut is already registered by this app or by another application, or if the accelerator string is invalid.

---

### unregister(accelerator)

Unregister a previously registered global shortcut.

**Parameters:**
- `accelerator` (string) — the same keyboard shortcut string used in `register()`

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.shortcuts.unregister('CommandOrControl+Shift+P')
```

---

### unregisterAll()

Unregister all global shortcuts registered by this application.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.shortcuts.unregisterAll()
```

---

### isRegistered(accelerator)

Check whether a specific shortcut is currently registered by this application.

**Parameters:**
- `accelerator` (string) — the keyboard shortcut string to check

**Returns:** `Promise<boolean>` — `true` if the shortcut is registered, `false` otherwise

**Example:**
```js
const active = await lightshell.shortcuts.isRegistered('CommandOrControl+Shift+P')
console.log(`Shortcut is ${active ? 'active' : 'inactive'}`)
```

---

## Modifier Keys

Accelerator strings are composed of modifier keys and a key name, joined with `+`. The `CommandOrControl` modifier is recommended for cross-platform shortcuts.

| Modifier String | macOS Key | Linux Key |
|----------------|-----------|-----------|
| `CommandOrControl` | Cmd | Ctrl |
| `Command` | Cmd | *(ignored on Linux)* |
| `Control` | Ctrl | Ctrl |
| `Alt` | Option | Alt |
| `Shift` | Shift | Shift |
| `Super` | Cmd | Super/Win |

**Key names:** Standard key names are used — `A` through `Z`, `0` through `9`, `F1` through `F12`, `Space`, `Tab`, `Enter`, `Backspace`, `Delete`, `Escape`, `Up`, `Down`, `Left`, `Right`, `Home`, `End`, `PageUp`, `PageDown`.

**Examples of valid accelerator strings:**
- `CommandOrControl+S` — Cmd+S on macOS, Ctrl+S on Linux
- `CommandOrControl+Shift+N` — Cmd+Shift+N on macOS, Ctrl+Shift+N on Linux
- `Alt+Space` — Option+Space on macOS, Alt+Space on Linux
- `F12` — F12 on both platforms (no modifier)
- `CommandOrControl+Shift+Alt+Z` — all modifiers combined

---

## Common Patterns

### Quick Launch / Command Palette

```js
async function setupGlobalShortcuts() {
  // Show/hide the app with a global shortcut
  await lightshell.shortcuts.register('CommandOrControl+Shift+Space', async () => {
    const { width, height } = await lightshell.window.getSize()
    if (width === 0 && height === 0) {
      await lightshell.window.restore()
    } else {
      await lightshell.window.minimize()
    }
  })
}
```

### Configurable Shortcuts

```js
async function registerUserShortcuts(shortcuts) {
  // Clear existing shortcuts
  await lightshell.shortcuts.unregisterAll()

  for (const [action, combo] of Object.entries(shortcuts)) {
    try {
      await lightshell.shortcuts.register(combo, () => {
        handleAction(action)
      })
    } catch (err) {
      console.warn(`Could not register ${combo} for ${action}: ${err.message}`)
    }
  }
}

// Load shortcuts from user preferences
const userShortcuts = await lightshell.store.get('shortcuts') || {
  'toggle-window': 'CommandOrControl+Shift+Space',
  'new-note': 'CommandOrControl+Shift+N',
  'search': 'CommandOrControl+Shift+F'
}

await registerUserShortcuts(userShortcuts)
```

### Cleanup on Quit

```js
// Always unregister shortcuts before quitting
async function gracefulQuit() {
  await lightshell.shortcuts.unregisterAll()
  await lightshell.app.quit()
}
```

## Platform Notes

- On macOS, global shortcuts are implemented using `CGEventTapCreate` and `NSEvent addGlobalMonitorForEvents`. The app must have Accessibility permissions for some shortcut combinations.
- On Linux, global shortcuts are implemented using X11's `XGrabKey` or `libkeybinder` (GTK-based). Wayland support depends on the compositor.
- Some key combinations are reserved by the operating system (e.g., Cmd+Tab on macOS, Ctrl+Alt+Delete on Linux) and cannot be registered.
- If another application has already registered the same global shortcut, `register()` will reject with an error.
- Global shortcuts are automatically unregistered when the application exits. However, calling `unregisterAll()` explicitly before quitting is good practice.
- Use `CommandOrControl` instead of `Command` or `Control` for shortcuts that should work on both platforms.
