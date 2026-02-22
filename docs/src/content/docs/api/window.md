---
title: Window API
description: Complete reference for lightshell.window — manage the application window.
---

The `lightshell.window` module controls the application window: title, size, position, and state. All methods are async and return Promises.

## Methods

### setTitle(title)

Set the window title bar text.

**Parameters:**
- `title` (string) — the new window title

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.setTitle('My App — document.txt')
```

---

### setSize(width, height)

Resize the window to the specified dimensions in pixels.

**Parameters:**
- `width` (number) — window width in pixels
- `height` (number) — window height in pixels

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.setSize(1280, 720)
```

---

### getSize()

Get the current window dimensions.

**Parameters:** none

**Returns:** `Promise<{ width: number, height: number }>`

**Example:**
```js
const { width, height } = await lightshell.window.getSize()
console.log(`Window is ${width}x${height}`)
```

---

### setPosition(x, y)

Move the window to the specified screen coordinates. The origin (0, 0) is the top-left corner of the primary display.

**Parameters:**
- `x` (number) — horizontal position in pixels
- `y` (number) — vertical position in pixels

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.setPosition(100, 100)
```

---

### getPosition()

Get the current window position on screen.

**Parameters:** none

**Returns:** `Promise<{ x: number, y: number }>`

**Example:**
```js
const { x, y } = await lightshell.window.getPosition()
console.log(`Window is at (${x}, ${y})`)
```

---

### minimize()

Minimize the window to the dock (macOS) or taskbar (Linux).

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.minimize()
```

---

### maximize()

Maximize the window to fill the screen, keeping the title bar and dock/taskbar visible.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.maximize()
```

---

### fullscreen()

Enter true fullscreen mode. On macOS, this uses the native fullscreen transition. On Linux, this uses GTK's fullscreen mode.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.fullscreen()
```

---

### restore()

Restore the window from minimized, maximized, or fullscreen state to its previous size and position.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.restore()
```

---

### close()

Close the application window. This typically triggers app shutdown unless a tray icon keeps the process alive.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.window.close()
```

---

## Events

### onResize(callback)

Fired when the window is resized by the user or programmatically.

**Parameters:**
- `callback` (function) — receives `{ width: number, height: number }`

**Returns:** unsubscribe function

**Example:**
```js
const unsubscribe = lightshell.window.onResize((data) => {
  console.log(`New size: ${data.width}x${data.height}`)
  adjustLayout(data.width, data.height)
})

// Later, stop listening
unsubscribe()
```

---

### onMove(callback)

Fired when the window is moved.

**Parameters:**
- `callback` (function) — receives `{ x: number, y: number }`

**Returns:** unsubscribe function

**Example:**
```js
lightshell.window.onMove((data) => {
  console.log(`Moved to (${data.x}, ${data.y})`)
})
```

---

### onFocus(callback)

Fired when the window gains focus.

**Parameters:**
- `callback` (function) — receives no arguments

**Returns:** unsubscribe function

**Example:**
```js
lightshell.window.onFocus(() => {
  document.title = 'My App'
  // Resume animations, refresh data, etc.
})
```

---

### onBlur(callback)

Fired when the window loses focus.

**Parameters:**
- `callback` (function) — receives no arguments

**Returns:** unsubscribe function

**Example:**
```js
lightshell.window.onBlur(() => {
  // Pause expensive animations when not focused
  pauseAnimations()
})
```

---

## Window Configuration

Initial window properties are set in `lightshell.json`:

```json
{
  "window": {
    "title": "My App",
    "width": 1024,
    "height": 768,
    "minWidth": 400,
    "minHeight": 300,
    "resizable": true,
    "frameless": false
  }
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `title` | string | app name | Initial window title |
| `width` | number | 1024 | Initial width in pixels |
| `height` | number | 768 | Initial height in pixels |
| `minWidth` | number | 0 | Minimum resize width |
| `minHeight` | number | 0 | Minimum resize height |
| `resizable` | boolean | true | Whether the user can resize |
| `frameless` | boolean | false | Hide the native title bar |

The `frameless` option removes the native window chrome. When using frameless mode, you need to implement your own title bar and window controls in HTML/CSS. Use `-webkit-app-region: drag` on your custom title bar element to make it draggable.

---

## Additional Methods

### setContentProtection(enabled)

Prevent the window contents from being captured by screen recording or screenshot tools. When enabled, the window appears as a black rectangle in recordings and screenshots.

**Parameters:**
- `enabled` (boolean) — `true` to hide window from capture, `false` to allow capture

**Returns:** `Promise<void>`

**Example:**
```js
// Enable content protection for sensitive data
await lightshell.window.setContentProtection(true)

// Disable when no longer needed
await lightshell.window.setContentProtection(false)
```

**Platform Notes:**
- macOS: Uses `NSWindow.sharingType = .none`. Effective against screen recording, screenshots, and screen sharing.
- Linux: Uses `gdk_window_set_shadow_type` where supported. Effectiveness varies by compositor and desktop environment.

---

### setVibrancy(style)

Apply a translucent vibrancy effect to the window background. This lets the desktop wallpaper and other windows show through with a blur effect. **macOS only.**

**Parameters:**
- `style` (string) — the vibrancy style. Options: `'titlebar'`, `'selection'`, `'menu'`, `'popover'`, `'sidebar'`, `'header'`, `'sheet'`, `'window'`, `'hud'`, `'fullscreen-ui'`, `'tooltip'`, `'content'`, `'under-window'`, `'under-page'`, `null` (to remove vibrancy)

**Returns:** `Promise<void>`

**Example:**
```js
// Apply sidebar vibrancy (common for navigation panels)
await lightshell.window.setVibrancy('sidebar')

// Apply under-window vibrancy (translucent window background)
await lightshell.window.setVibrancy('under-window')

// Remove vibrancy
await lightshell.window.setVibrancy(null)
```

To see the vibrancy effect, your page background must be transparent:

```css
html, body {
  background: transparent;
}
```

**Platform Notes:**
- macOS: Full support. Uses `NSVisualEffectView` behind the webview.
- Linux: Not supported. The call is a no-op and does not throw an error.

---

### onFileDrop(callback)

Handle files dragged and dropped onto the application window. When the user drags files from Finder (macOS) or a file manager (Linux) onto your window, the callback receives the list of dropped file paths.

**Parameters:**
- `callback` (function) — receives `{ paths: string[], x: number, y: number }` where `paths` is an array of absolute file paths, and `x`/`y` are the drop coordinates relative to the window

**Returns:** unsubscribe function

**Example:**
```js
const unsubscribe = lightshell.window.onFileDrop(async (data) => {
  console.log(`Dropped ${data.paths.length} files at (${data.x}, ${data.y})`)

  for (const path of data.paths) {
    const content = await lightshell.fs.readFile(path)
    processFile(path, content)
  }
})

// Later, stop listening
unsubscribe()
```

**Example: Image Drop Zone**
```js
lightshell.window.onFileDrop(async ({ paths }) => {
  const images = paths.filter(p =>
    p.endsWith('.png') || p.endsWith('.jpg') || p.endsWith('.jpeg')
  )

  if (images.length === 0) {
    await lightshell.dialog.message('Invalid Files', 'Please drop image files (.png, .jpg)')
    return
  }

  for (const img of images) {
    addImageToGallery(img)
  }
})
```

**Note:** The dropped file paths are real filesystem paths. You can read them directly with `lightshell.fs.readFile()`. In restricted permission mode, the dropped paths must still match the allowed read patterns in `lightshell.json`, or you need to add `$HOME/**` or the relevant directories to `permissions.fs.read`.

---

### setColorScheme(scheme)

Force the app to use a specific color scheme, overriding the OS preference. This affects CSS media queries (`prefers-color-scheme`) and the `light-dark()` function in the default stylesheet.

**Parameters:**
- `scheme` (string) — one of `'light'`, `'dark'`, or `'system'`

**Returns:** `Promise<void>`

**Example:**
```js
// Force dark mode
await lightshell.window.setColorScheme('dark')

// Force light mode
await lightshell.window.setColorScheme('light')

// Follow OS setting (default)
await lightshell.window.setColorScheme('system')
```

**Example: Theme Toggle**
```js
let currentScheme = 'system'

async function toggleTheme() {
  if (currentScheme === 'dark') {
    currentScheme = 'light'
  } else {
    currentScheme = 'dark'
  }

  await lightshell.window.setColorScheme(currentScheme)
  await lightshell.store.set('theme', currentScheme)
}

// Restore saved theme on startup
async function initTheme() {
  const saved = await lightshell.store.get('theme')
  if (saved) {
    currentScheme = saved
    await lightshell.window.setColorScheme(currentScheme)
  }
}
```
