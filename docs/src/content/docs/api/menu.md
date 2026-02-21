---
title: Menu API
description: Complete reference for lightshell.menu — application menu bar.
---

The `lightshell.menu` module sets the application's native menu bar. On macOS, this is the menu bar at the top of the screen. On Linux, it is the menu bar within the application window (if the desktop environment supports it). All methods are async and return Promises.

## Methods

### set(template)

Set the entire application menu bar from a template. The template is an array of top-level menus, each containing an array of menu items. Calling `set()` replaces the entire menu bar.

**Parameters:**
- `template` (array) — array of menu objects, each with:
  - `label` (string) — the top-level menu name (e.g., `"File"`, `"Edit"`)
  - `items` (array) — array of menu items within this menu

**Menu item properties:**

| Property | Type | Description |
|----------|------|-------------|
| `label` | string | The text displayed in the menu item |
| `id` | string | Unique identifier, emitted in click events |
| `accelerator` | string | Keyboard shortcut string (see table below) |
| `enabled` | boolean | Whether the item is clickable (default: `true`) |
| `checked` | boolean | Show a checkmark next to the item (default: `false`) |
| `type` | string | `"normal"`, `"separator"`, or `"checkbox"` (default: `"normal"`) |
| `role` | string | Built-in system role (see below) |
| `submenu` | array | Nested array of menu items for a submenu |

**Accelerator strings:**

| String | macOS | Linux |
|--------|-------|-------|
| `CommandOrControl+S` | Cmd+S | Ctrl+S |
| `CommandOrControl+Shift+S` | Cmd+Shift+S | Ctrl+Shift+S |
| `Command+Q` | Cmd+Q | (ignored) |
| `Alt+F4` | Option+F4 | Alt+F4 |
| `F11` | F11 | F11 |

Modifiers can be combined: `CommandOrControl+Shift+Alt+Z`.

**Built-in roles:**

| Role | Description |
|------|-------------|
| `undo` | Undo the last action |
| `redo` | Redo the last undone action |
| `cut` | Cut selection to clipboard |
| `copy` | Copy selection to clipboard |
| `paste` | Paste from clipboard |
| `selectAll` | Select all content |
| `minimize` | Minimize the window |
| `close` | Close the window |
| `quit` | Quit the application |

When a `role` is set, the menu item uses the system's built-in behavior and label. You do not need to set `id` or handle click events for role-based items.

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.menu.set([
  {
    label: 'File',
    items: [
      { label: 'New', id: 'file-new', accelerator: 'CommandOrControl+N' },
      { label: 'Open...', id: 'file-open', accelerator: 'CommandOrControl+O' },
      { label: 'Save', id: 'file-save', accelerator: 'CommandOrControl+S' },
      { type: 'separator' },
      { role: 'quit' }
    ]
  },
  {
    label: 'Edit',
    items: [
      { role: 'undo' },
      { role: 'redo' },
      { type: 'separator' },
      { role: 'cut' },
      { role: 'copy' },
      { role: 'paste' },
      { role: 'selectAll' }
    ]
  }
])
```

Menu item clicks are handled via the `lightshell.on()` event system:

```js
lightshell.on('menu.click', (event) => {
  switch (event.id) {
    case 'file-new':
      createNewDocument()
      break
    case 'file-open':
      openDocument()
      break
    case 'file-save':
      saveDocument()
      break
  }
})
```

---

## Common Patterns

### Standard App Menu

A typical menu bar for a text editor or document-based application.

```js
async function setupMenu() {
  await lightshell.menu.set([
    {
      label: 'File',
      items: [
        { label: 'New', id: 'new', accelerator: 'CommandOrControl+N' },
        { label: 'Open...', id: 'open', accelerator: 'CommandOrControl+O' },
        { type: 'separator' },
        { label: 'Save', id: 'save', accelerator: 'CommandOrControl+S' },
        { label: 'Save As...', id: 'save-as', accelerator: 'CommandOrControl+Shift+S' },
        { type: 'separator' },
        { role: 'quit' }
      ]
    },
    {
      label: 'Edit',
      items: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectAll' }
      ]
    },
    {
      label: 'View',
      items: [
        { label: 'Zoom In', id: 'zoom-in', accelerator: 'CommandOrControl+=' },
        { label: 'Zoom Out', id: 'zoom-out', accelerator: 'CommandOrControl+-' },
        { label: 'Reset Zoom', id: 'zoom-reset', accelerator: 'CommandOrControl+0' },
        { type: 'separator' },
        { label: 'Fullscreen', id: 'fullscreen', accelerator: 'F11' }
      ]
    },
    {
      label: 'Help',
      items: [
        { label: 'Documentation', id: 'help-docs' },
        { label: 'Report Issue', id: 'help-issue' },
        { type: 'separator' },
        { label: 'About', id: 'help-about' }
      ]
    }
  ])

  lightshell.on('menu.click', async (event) => {
    switch (event.id) {
      case 'new': createNewDocument(); break
      case 'open': openDocument(); break
      case 'save': saveDocument(); break
      case 'save-as': saveDocumentAs(); break
      case 'zoom-in': adjustZoom(1); break
      case 'zoom-out': adjustZoom(-1); break
      case 'zoom-reset': resetZoom(); break
      case 'fullscreen': await lightshell.window.fullscreen(); break
      case 'help-docs': lightshell.shell.open('https://myapp.dev/docs'); break
      case 'help-issue': lightshell.shell.open('https://github.com/me/myapp/issues'); break
      case 'help-about': showAbout(); break
    }
  })
}
```

### Menu with Submenus

```js
await lightshell.menu.set([
  {
    label: 'File',
    items: [
      {
        label: 'Export',
        submenu: [
          { label: 'As PDF', id: 'export-pdf' },
          { label: 'As HTML', id: 'export-html' },
          { label: 'As Markdown', id: 'export-md' }
        ]
      },
      { type: 'separator' },
      { role: 'quit' }
    ]
  }
])
```

### Dynamic Menu Updates

Rebuild the menu to reflect current state (e.g., recent files, toggle states).

```js
async function updateRecentFilesMenu(recentFiles) {
  const recentItems = recentFiles.map((file, i) => ({
    label: file.split('/').pop(),
    id: `recent-${i}`
  }))

  await lightshell.menu.set([
    {
      label: 'File',
      items: [
        { label: 'Open...', id: 'open', accelerator: 'CommandOrControl+O' },
        { type: 'separator' },
        { label: 'Recent Files', submenu: recentItems.length > 0
          ? recentItems
          : [{ label: 'No Recent Files', enabled: false }]
        },
        { type: 'separator' },
        { role: 'quit' }
      ]
    }
  ])
}
```

## Platform Notes

- On macOS, the first menu in the template becomes the application menu (shown with the app name). It is standard practice to include `quit` and `About` items in this menu.
- On Linux, the menu bar appears inside the window (depending on desktop environment settings).
- `CommandOrControl` resolves to `Cmd` on macOS and `Ctrl` on Linux. Use this instead of `Control` for cross-platform compatibility.
- Role-based items (`role: 'copy'`, etc.) use the system's native implementation and localized labels automatically.
- Setting an empty template (`[]`) clears the menu bar entirely.
- Separator items (`type: 'separator'`) render as horizontal lines and have no label or click handler.
