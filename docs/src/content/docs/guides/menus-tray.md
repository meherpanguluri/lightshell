---
title: Menus & System Tray
description: Add system tray icons and application menus to your LightShell app.
---

LightShell supports system tray icons and application menus through the `lightshell.tray` and `lightshell.menu` modules. These are P1 (nice-to-have) APIs — they work on both platforms but may have minor visual differences.

## System Tray

The system tray (menu bar on macOS, notification area on Linux) lets your app stay accessible even when the window is closed or minimized.

### Setting Up the Tray

```js
await lightshell.tray.set({
  tooltip: 'My App — Running',
  menu: [
    { label: 'Open Window', id: 'open' },
    { label: 'New File', id: 'new' },
    { type: 'separator' },
    { label: 'Quit', id: 'quit' }
  ]
})
```

The `tooltip` appears when hovering over the tray icon. The `menu` is an array of items shown when clicking the tray icon.

### Handling Tray Clicks

```js
lightshell.tray.onClick((data) => {
  switch (data.id) {
    case 'open':
      lightshell.window.restore()
      break
    case 'new':
      createNewFile()
      break
    case 'quit':
      lightshell.app.quit()
      break
  }
})
```

### Updating the Tray

Call `tray.set()` again to update the tooltip or menu items:

```js
// Update tooltip to show status
await lightshell.tray.set({
  tooltip: `My App — ${fileCount} files open`,
  menu: [
    { label: `${fileCount} files open`, id: 'status', enabled: false },
    { type: 'separator' },
    { label: 'Quit', id: 'quit' }
  ]
})
```

### Removing the Tray

```js
await lightshell.tray.remove()
```

### Tray Menu Item Properties

| Property | Type | Description |
|----------|------|-------------|
| `label` | string | Text displayed in the menu |
| `id` | string | Identifier passed to the click handler |
| `type` | string | `'separator'` for a divider line |
| `enabled` | boolean | `false` to show as grayed out (default: `true`) |

## Application Menus

The `lightshell.menu` module lets you define the application menu bar (the menu at the top of the screen on macOS, or the window menu on Linux).

### Setting the Menu

```js
await lightshell.menu.set({
  template: [
    {
      label: 'File',
      submenu: [
        { label: 'New', accelerator: 'CmdOrCtrl+N', id: 'file-new' },
        { label: 'Open...', accelerator: 'CmdOrCtrl+O', id: 'file-open' },
        { label: 'Save', accelerator: 'CmdOrCtrl+S', id: 'file-save' },
        { label: 'Save As...', accelerator: 'CmdOrCtrl+Shift+S', id: 'file-save-as' },
        { type: 'separator' },
        { label: 'Quit', accelerator: 'CmdOrCtrl+Q', id: 'app-quit' }
      ]
    },
    {
      label: 'Edit',
      submenu: [
        { label: 'Undo', accelerator: 'CmdOrCtrl+Z', role: 'undo' },
        { label: 'Redo', accelerator: 'CmdOrCtrl+Shift+Z', role: 'redo' },
        { type: 'separator' },
        { label: 'Cut', accelerator: 'CmdOrCtrl+X', role: 'cut' },
        { label: 'Copy', accelerator: 'CmdOrCtrl+C', role: 'copy' },
        { label: 'Paste', accelerator: 'CmdOrCtrl+V', role: 'paste' },
        { label: 'Select All', accelerator: 'CmdOrCtrl+A', role: 'selectAll' }
      ]
    },
    {
      label: 'View',
      submenu: [
        { label: 'Fullscreen', accelerator: 'F11', id: 'view-fullscreen' },
        { label: 'Zoom In', accelerator: 'CmdOrCtrl+=', id: 'view-zoom-in' },
        { label: 'Zoom Out', accelerator: 'CmdOrCtrl+-', id: 'view-zoom-out' }
      ]
    },
    {
      label: 'Help',
      submenu: [
        { label: 'About My App', id: 'help-about' }
      ]
    }
  ]
})
```

### Menu Item Properties

| Property | Type | Description |
|----------|------|-------------|
| `label` | string | Text shown in the menu |
| `id` | string | Identifier for custom click handling |
| `submenu` | array | Nested menu items |
| `accelerator` | string | Keyboard shortcut (e.g., `'CmdOrCtrl+S'`) |
| `type` | string | `'separator'` for a divider |
| `role` | string | Built-in action: `'undo'`, `'redo'`, `'cut'`, `'copy'`, `'paste'`, `'selectAll'` |
| `enabled` | boolean | `false` to gray out the item |

### Accelerator Strings

Use `CmdOrCtrl` for cross-platform shortcuts — it maps to Cmd on macOS and Ctrl on Linux.

| Accelerator | macOS | Linux |
|------------|-------|-------|
| `CmdOrCtrl+S` | Cmd+S | Ctrl+S |
| `CmdOrCtrl+Shift+S` | Cmd+Shift+S | Ctrl+Shift+S |
| `Alt+F4` | Opt+F4 | Alt+F4 |
| `F11` | F11 | F11 |

### Handling Menu Clicks

Menu items with an `id` fire events through the `lightshell.on` handler:

```js
lightshell.on('menu.click', (data) => {
  switch (data.id) {
    case 'file-new':
      createNewFile()
      break
    case 'file-open':
      openFile()
      break
    case 'file-save':
      saveFile()
      break
    case 'app-quit':
      lightshell.app.quit()
      break
    case 'view-fullscreen':
      lightshell.window.fullscreen()
      break
  }
})
```

Items with a `role` are handled automatically by the system — you do not need to add click handlers for cut, copy, paste, etc.

## Complete Example: Notes App with Menus and Tray

```js
async function initMenusAndTray() {
  // App menu
  await lightshell.menu.set({
    template: [
      {
        label: 'File',
        submenu: [
          { label: 'New Note', accelerator: 'CmdOrCtrl+N', id: 'new' },
          { label: 'Open...', accelerator: 'CmdOrCtrl+O', id: 'open' },
          { label: 'Save', accelerator: 'CmdOrCtrl+S', id: 'save' },
          { type: 'separator' },
          { label: 'Quit', accelerator: 'CmdOrCtrl+Q', id: 'quit' }
        ]
      },
      {
        label: 'Edit',
        submenu: [
          { label: 'Undo', role: 'undo', accelerator: 'CmdOrCtrl+Z' },
          { label: 'Redo', role: 'redo', accelerator: 'CmdOrCtrl+Shift+Z' },
          { type: 'separator' },
          { label: 'Cut', role: 'cut', accelerator: 'CmdOrCtrl+X' },
          { label: 'Copy', role: 'copy', accelerator: 'CmdOrCtrl+C' },
          { label: 'Paste', role: 'paste', accelerator: 'CmdOrCtrl+V' },
        ]
      }
    ]
  })

  // System tray
  await lightshell.tray.set({
    tooltip: 'Notes — Ready',
    menu: [
      { label: 'New Note', id: 'new' },
      { label: 'Show Window', id: 'show' },
      { type: 'separator' },
      { label: 'Quit', id: 'quit' }
    ]
  })

  // Handle menu clicks
  lightshell.on('menu.click', handleAction)
  lightshell.tray.onClick(handleAction)
}

function handleAction(data) {
  switch (data.id) {
    case 'new': createNewNote(); break
    case 'open': openNote(); break
    case 'save': saveNote(); break
    case 'show': lightshell.window.restore(); break
    case 'quit': lightshell.app.quit(); break
  }
}

initMenusAndTray()
```

## Platform Differences

| Feature | macOS | Linux |
|---------|-------|-------|
| Tray icon location | Menu bar (top right) | System tray area (varies by DE) |
| App menu location | Screen top (global menu bar) | Window title bar area |
| Accelerator display | Shows symbols | Shows text (Ctrl, Alt) |
| Tray click behavior | Left-click shows menu | Left-click shows menu |

The tray and menu APIs normalize these differences as much as possible, but the visual location follows each platform's conventions.
