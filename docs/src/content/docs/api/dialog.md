---
title: Dialog API
description: Complete reference for lightshell.dialog — native file pickers, alerts, confirmations, and prompts.
---

The `lightshell.dialog` module provides native OS dialogs for file selection, messages, confirmations, and text input. All methods are async and return Promises.

## File Dialogs

### open(options?)

Show a native file open dialog. The user selects a file and the path is returned.

**Parameters:**
- `options` (object, optional):
  - `title` (string) — dialog window title
  - `defaultPath` (string) — initial directory or file path
  - `filters` (array) — file type filters, each with `name` (string) and `extensions` (array of strings)
  - `multiple` (boolean) — allow selecting multiple files (default: `false`)

**Returns:** `Promise<string | string[] | null>` — the selected file path, an array of paths if `multiple` is true, or `null` if the user cancelled

**Example:**
```js
// Simple open
const path = await lightshell.dialog.open()

// With filters
const imagePath = await lightshell.dialog.open({
  title: 'Select an Image',
  filters: [
    { name: 'Images', extensions: ['png', 'jpg', 'gif', 'webp'] },
    { name: 'All Files', extensions: ['*'] }
  ]
})

if (imagePath) {
  const data = await lightshell.fs.readFile(imagePath)
  // process the file
}
```

**Multiple selection:**
```js
const paths = await lightshell.dialog.open({
  title: 'Select Files',
  multiple: true,
  filters: [
    { name: 'Documents', extensions: ['txt', 'md', 'pdf'] }
  ]
})

if (paths && paths.length > 0) {
  for (const p of paths) {
    console.log('Selected:', p)
  }
}
```

---

### save(options?)

Show a native file save dialog. The user chooses a destination path.

**Parameters:**
- `options` (object, optional):
  - `title` (string) — dialog window title
  - `defaultPath` (string) — suggested file name or full path
  - `filters` (array) — file type filters, same format as `open()`

**Returns:** `Promise<string | null>` — the chosen save path, or `null` if cancelled

**Example:**
```js
const savePath = await lightshell.dialog.save({
  title: 'Save Document',
  defaultPath: 'report.txt',
  filters: [
    { name: 'Text Files', extensions: ['txt'] },
    { name: 'Markdown', extensions: ['md'] }
  ]
})

if (savePath) {
  await lightshell.fs.writeFile(savePath, documentContent)
}
```

---

## Message Dialogs

### message(title, message)

Show an informational message dialog with an OK button. Blocks until the user dismisses it.

**Parameters:**
- `title` (string) — dialog title
- `message` (string) — dialog body text

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.dialog.message('Export Complete', 'Your data has been exported to /tmp/export.csv')
```

---

### confirm(title, message)

Show a confirmation dialog with OK and Cancel buttons.

**Parameters:**
- `title` (string) — dialog title
- `message` (string) — dialog body text

**Returns:** `Promise<boolean>` — `true` if the user clicked OK, `false` if cancelled

**Example:**
```js
const confirmed = await lightshell.dialog.confirm(
  'Delete File',
  'Are you sure you want to permanently delete this file? This cannot be undone.'
)

if (confirmed) {
  await lightshell.fs.remove(filePath)
}
```

---

### prompt(title, defaultValue?)

Show a text input dialog.

**Parameters:**
- `title` (string) — dialog title / prompt text
- `defaultValue` (string, optional) — pre-filled text in the input field, defaults to `''`

**Returns:** `Promise<string | null>` — the entered text, or `null` if cancelled

**Example:**
```js
const fileName = await lightshell.dialog.prompt('New File', 'untitled.txt')
if (fileName) {
  await lightshell.fs.writeFile(`/tmp/${fileName}`, '')
  console.log(`Created ${fileName}`)
}
```

---

## Common Patterns

### Open and Read a File

```js
async function openDocument() {
  const path = await lightshell.dialog.open({
    filters: [
      { name: 'Text', extensions: ['txt', 'md'] },
      { name: 'All Files', extensions: ['*'] }
    ]
  })
  if (!path) return null

  const content = await lightshell.fs.readFile(path)
  return { path, content }
}
```

### Save with Confirmation

```js
async function saveDocument(path, content) {
  if (await lightshell.fs.exists(path)) {
    const overwrite = await lightshell.dialog.confirm(
      'File Exists',
      `${path} already exists. Do you want to overwrite it?`
    )
    if (!overwrite) return false
  }
  await lightshell.fs.writeFile(path, content)
  return true
}
```

### Unsaved Changes Guard

```js
let hasUnsavedChanges = false

async function beforeClose() {
  if (!hasUnsavedChanges) return true

  const save = await lightshell.dialog.confirm(
    'Unsaved Changes',
    'You have unsaved changes. Do you want to save before closing?'
  )
  if (save) {
    await saveFile()
  }
  return true
}
```

## Platform Notes

- On macOS, dialogs use the native `NSOpenPanel`, `NSSavePanel`, and `NSAlert` APIs
- On Linux, dialogs use GTK's `GtkFileChooserDialog` and `GtkMessageDialog`
- The visual appearance follows each platform's native look
- File filters work on both platforms but may display slightly differently
