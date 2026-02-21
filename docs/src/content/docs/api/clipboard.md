---
title: Clipboard API
description: Complete reference for lightshell.clipboard — read and write system clipboard text.
---

The `lightshell.clipboard` module provides access to the system clipboard for reading and writing text. All methods are async and return Promises.

## Methods

### read()

Read the current text content from the system clipboard.

**Parameters:** none

**Returns:** `Promise<string>` — the clipboard text content, or an empty string if the clipboard is empty or contains non-text data

**Example:**
```js
const text = await lightshell.clipboard.read()
console.log('Clipboard contains:', text)
```

---

### write(text)

Write text to the system clipboard, replacing any existing content.

**Parameters:**
- `text` (string) — the text to place on the clipboard

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.clipboard.write('Hello from LightShell!')

// Verify it was written
const check = await lightshell.clipboard.read()
console.log(check) // "Hello from LightShell!"
```

---

## Common Patterns

### Copy Button

```js
document.getElementById('copy-btn').addEventListener('click', async () => {
  const content = document.getElementById('output').textContent
  await lightshell.clipboard.write(content)

  // Visual feedback
  const btn = document.getElementById('copy-btn')
  btn.textContent = 'Copied!'
  setTimeout(() => { btn.textContent = 'Copy' }, 2000)
})
```

### Paste Button

```js
document.getElementById('paste-btn').addEventListener('click', async () => {
  const text = await lightshell.clipboard.read()
  document.getElementById('editor').value += text
})
```

### Copy Rich Content as Text

```js
async function copyAsMarkdown(htmlElement) {
  // Convert HTML to a simple text representation
  const text = htmlElement.innerText
  await lightshell.clipboard.write(text)
}
```

### Clipboard History

```js
const clipboardHistory = []

async function checkClipboard() {
  const current = await lightshell.clipboard.read()
  if (current && current !== clipboardHistory[0]) {
    clipboardHistory.unshift(current)
    if (clipboardHistory.length > 20) clipboardHistory.pop()
    renderHistory()
  }
}

// Poll every 2 seconds
setInterval(checkClipboard, 2000)
```

### Copy Formatted Data

```js
// Copy as tab-separated values for spreadsheet pasting
async function copyTableData(rows) {
  const tsv = rows.map(row => row.join('\t')).join('\n')
  await lightshell.clipboard.write(tsv)
}

// Example
await copyTableData([
  ['Name', 'Age', 'City'],
  ['Alice', '30', 'NYC'],
  ['Bob', '25', 'LA']
])
```

## Platform Notes

- On macOS, uses `NSPasteboard` (the system pasteboard)
- On Linux, uses GTK's clipboard API which interacts with the X11 or Wayland clipboard
- Only plain text is supported in v1.0 — image and rich text clipboard operations are not available
- The clipboard is shared with all other applications on the system
