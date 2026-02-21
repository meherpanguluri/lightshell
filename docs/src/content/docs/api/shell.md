---
title: Shell API
description: Complete reference for lightshell.shell — open URLs and files with system defaults.
---

The `lightshell.shell` module opens URLs in the default browser and files in their default applications. This is the correct way to handle external links in a LightShell app — never use `window.open()` for external URLs, as it will attempt to navigate the webview. All methods are async and return Promises.

## Methods

### open(url)

Open a URL in the user's default browser, or a file path in the default application for that file type. This uses the operating system's native open mechanism.

**Parameters:**
- `url` (string) — a URL (e.g., `https://example.com`) or an absolute file path (e.g., `/Users/me/document.pdf`)

**Returns:** `Promise<void>`

**Example:**
```js
// Open a website in the default browser
await lightshell.shell.open('https://lightshell.dev')

// Open a file in the default application
await lightshell.shell.open('/Users/me/document.pdf')

// Open a directory in the file manager
await lightshell.shell.open('/Users/me/Documents')
```

**Errors:** Rejects if the URL is malformed or the system cannot find a handler for the given URL or file type.

---

## Common Patterns

### External Link Handler

Intercept all external links in your app and open them in the system browser instead of navigating the webview.

```js
// Make all <a> tags with http/https hrefs open in system browser
document.addEventListener('click', (e) => {
  const link = e.target.closest('a[href^="http"]')
  if (link) {
    e.preventDefault()
    lightshell.shell.open(link.href)
  }
})
```

### Open Files in Default Apps

```js
// Open an image with the system image viewer
async function previewFile(filePath) {
  await lightshell.shell.open(filePath)
}

// Open a folder in Finder (macOS) or Files (Linux)
async function revealInFileManager(dirPath) {
  await lightshell.shell.open(dirPath)
}
```

### Help and Support Links

```js
function setupHelpMenu() {
  document.getElementById('docs-link').addEventListener('click', () => {
    lightshell.shell.open('https://lightshell.dev/docs')
  })

  document.getElementById('report-bug').addEventListener('click', () => {
    lightshell.shell.open('https://github.com/myapp/myapp/issues/new')
  })

  document.getElementById('email-support').addEventListener('click', () => {
    lightshell.shell.open('mailto:support@myapp.com')
  })
}
```

### Open Generated Files

```js
async function exportAndOpen() {
  const tmp = await lightshell.system.tempDir()
  const outputPath = `${tmp}/report.html`

  await lightshell.fs.writeFile(outputPath, generateReport())

  // Open the report in the default browser
  await lightshell.shell.open(outputPath)
}
```

## Platform Notes

- On macOS, uses the `open` command internally (via `NSWorkspace`)
- On Linux, uses `xdg-open` to delegate to the desktop environment's default handler
- `mailto:` links are supported and will open the default email client
- File paths must be absolute — relative paths are not supported
- Opening a directory opens the platform's file manager at that location
