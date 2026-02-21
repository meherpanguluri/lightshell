---
title: Keyboard Shortcuts
description: Add in-app and global keyboard shortcuts to your LightShell app.
---

LightShell supports two kinds of keyboard shortcuts: **in-app shortcuts** that work when your window is focused, and **global shortcuts** that work system-wide, even when your app is in the background.

## In-App Shortcuts

For shortcuts that only need to work while your app is focused, use standard DOM keyboard event listeners. This is the simplest approach and requires no special API.

```js
document.addEventListener('keydown', (e) => {
  // Cmd+S on macOS, Ctrl+S on Linux
  if ((e.metaKey || e.ctrlKey) && e.key === 's') {
    e.preventDefault()
    saveFile()
  }

  // Cmd+O / Ctrl+O
  if ((e.metaKey || e.ctrlKey) && e.key === 'o') {
    e.preventDefault()
    openFile()
  }

  // Cmd+Shift+P / Ctrl+Shift+P
  if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'P') {
    e.preventDefault()
    toggleCommandPalette()
  }

  // Escape
  if (e.key === 'Escape') {
    closeModal()
  }
})
```

### Cross-Platform Key Detection

On macOS, `e.metaKey` is the Command key. On Linux, `e.ctrlKey` is the Ctrl key. To handle both platforms with one check, test for either:

```js
function isModKey(e) {
  return e.metaKey || e.ctrlKey
}

document.addEventListener('keydown', (e) => {
  if (isModKey(e) && e.key === 'n') {
    e.preventDefault()
    newFile()
  }
})
```

### Building a Shortcut Handler

For apps with many shortcuts, a lookup table is cleaner than a chain of if/else:

```js
const shortcuts = {
  'mod+s': saveFile,
  'mod+o': openFile,
  'mod+n': newFile,
  'mod+shift+p': toggleCommandPalette,
  'mod+,': openSettings,
  'mod+w': closeTab,
  'escape': closeModal,
  'f11': toggleFullscreen,
}

document.addEventListener('keydown', (e) => {
  const parts = []
  if (e.metaKey || e.ctrlKey) parts.push('mod')
  if (e.shiftKey) parts.push('shift')
  if (e.altKey) parts.push('alt')
  parts.push(e.key.toLowerCase())

  const combo = parts.join('+')
  const handler = shortcuts[combo]

  if (handler) {
    e.preventDefault()
    handler()
  }
})
```

## Global Shortcuts

Global shortcuts work even when your app window is not focused. They are registered through `lightshell.shortcuts` and use the system's global hotkey mechanism.

This is essential for productivity tools, clipboard managers, quick launchers, and screenshot utilities.

### Registering a Global Shortcut

```js
lightshell.shortcuts.register('CommandOrControl+Shift+Space', () => {
  lightshell.window.restore()  // bring window to front
  document.getElementById('search').focus()
})
```

### Modifier Key Names

Use these cross-platform modifier names in your shortcut strings:

| Key String | macOS | Linux |
|-----------|-------|-------|
| `CommandOrControl` | Cmd | Ctrl |
| `Command` | Cmd | (ignored on Linux) |
| `Control` | Ctrl | Ctrl |
| `Alt` | Option | Alt |
| `Shift` | Shift | Shift |
| `Super` | Cmd | Super/Win |

`CommandOrControl` is the recommended modifier for most shortcuts. It maps to the primary modifier key on each platform.

### Shortcut String Format

Combine modifiers and a key with `+`. The key part can be a letter, number, function key, or special key:

```
CommandOrControl+Shift+P
Alt+Space
Control+F12
CommandOrControl+1
```

### Unregistering Shortcuts

```js
// Remove a specific shortcut
lightshell.shortcuts.unregister('CommandOrControl+Shift+Space')

// Remove all global shortcuts
lightshell.shortcuts.unregisterAll()
```

### Check if a Shortcut is Registered

```js
const registered = await lightshell.shortcuts.isRegistered('CommandOrControl+Shift+Space')
console.log(registered) // true or false
```

## Common Patterns

### Quick Launcher

A clipboard manager or quick launcher that activates with a global hotkey:

```js
// Register global hotkey to show/hide the app
lightshell.shortcuts.register('CommandOrControl+Shift+Space', () => {
  toggleWindow()
})

async function toggleWindow() {
  // Bring window to front and focus the search input
  await lightshell.window.restore()
  document.getElementById('search-input').focus()
  document.getElementById('search-input').select()
}

// Escape hides the window
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    lightshell.window.minimize()
  }
})
```

### Text Editor Shortcuts

A full set of shortcuts for a text editing app:

```js
document.addEventListener('keydown', (e) => {
  const mod = e.metaKey || e.ctrlKey

  if (mod && e.key === 's') {
    e.preventDefault()
    if (e.shiftKey) {
      saveFileAs()
    } else {
      saveFile()
    }
  }

  if (mod && e.key === 'o') {
    e.preventDefault()
    openFile()
  }

  if (mod && e.key === 'n') {
    e.preventDefault()
    newFile()
  }

  if (mod && e.key === 'f') {
    e.preventDefault()
    showFindBar()
  }

  if (mod && e.key === 'z') {
    e.preventDefault()
    if (e.shiftKey) {
      redo()
    } else {
      undo()
    }
  }
})
```

### Clean Up on Quit

Always unregister global shortcuts when your app exits. Global shortcuts remain active at the OS level until explicitly removed:

```js
// Register shortcuts on startup
function initShortcuts() {
  lightshell.shortcuts.register('CommandOrControl+Shift+L', showApp)
  lightshell.shortcuts.register('CommandOrControl+Shift+K', quickCapture)
}

// Clean up before quitting
window.addEventListener('beforeunload', () => {
  lightshell.shortcuts.unregisterAll()
})

initShortcuts()
```

### User-Configurable Shortcuts

Let users customize their shortcuts by storing the bindings in `lightshell.store`:

```js
const defaultBindings = {
  save: 'mod+s',
  open: 'mod+o',
  find: 'mod+f',
  newFile: 'mod+n',
}

async function loadShortcuts() {
  const saved = await lightshell.store.get('keybindings')
  return saved || defaultBindings
}

async function saveShortcuts(bindings) {
  await lightshell.store.set('keybindings', bindings)
}
```

## Best Practices

**Do not conflict with system shortcuts.** Avoid overriding shortcuts the OS or desktop environment already uses:

| Shortcut | Reserved By |
|----------|------------|
| Cmd+Q (macOS) | Quit application |
| Cmd+H (macOS) | Hide application |
| Cmd+Tab | App switcher |
| Ctrl+Alt+Delete (Linux) | System menu |
| Super (Linux) | Activities/launcher |
| Alt+F4 (Linux) | Close window |

**Use `CommandOrControl` instead of platform-specific modifiers.** This makes your shortcuts work on both macOS and Linux without platform checks.

**Keep global shortcuts minimal.** Only register global shortcuts for actions that genuinely need to work when your app is not focused. Overusing global shortcuts can conflict with other apps.

**Always call `e.preventDefault()`** for in-app shortcuts. Without it, the browser default action (like Cmd+S triggering a save dialog in the webview) may fire alongside your handler.

**Show shortcuts in your UI.** Display the keyboard shortcut next to menu items and buttons so users can discover them. Use `lightshell.system.platform()` to show the correct modifier symbol:

```js
const platform = await lightshell.system.platform()
const modLabel = platform === 'darwin' ? 'Cmd' : 'Ctrl'

document.getElementById('save-btn').title = `Save (${modLabel}+S)`
```

## Platform Notes

| Feature | macOS | Linux |
|---------|-------|-------|
| Global hotkey API | CGEventTap / NSEvent | X11 XGrabKey |
| CommandOrControl | Maps to Cmd | Maps to Ctrl |
| Function keys | May need Fn key | Work directly |
| Media keys | Typically reserved by OS | Varies by DE |
| Global shortcut limit | No hard limit | No hard limit |

Global shortcuts are registered at the OS level. If another application has already claimed a shortcut, registration may fail silently or the other app may take priority. Choose distinctive key combinations to minimize conflicts.
