---
title: Dialogs & Prompts
description: Use native file pickers, message boxes, confirmations, and text prompts in your LightShell app.
---

The `lightshell.dialog` module provides native OS dialogs — file pickers, message boxes, confirmations, and text input prompts. All methods are async and return `null` when the user cancels (not an error).

## File Open Dialog

### open(options?)

Show a native file picker for opening files. Returns the selected file path, or `null` if cancelled.

```js
const filePath = await lightshell.dialog.open({
  title: 'Open Document',
  filters: [
    { name: 'Text Files', extensions: ['txt', 'md'] },
    { name: 'All Files', extensions: ['*'] }
  ]
})

if (filePath) {
  const content = await lightshell.fs.readFile(filePath)
  editor.value = content
}
```

### Options

| Property | Type | Description |
|----------|------|-------------|
| `title` | string | Window title for the dialog |
| `filters` | array | File type filters (see below) |
| `defaultPath` | string | Starting directory or default file name |
| `multiple` | boolean | Allow selecting multiple files (returns array) |
| `directory` | boolean | Select a directory instead of a file |

### File Filters

Filters control which files are visible in the picker. Each filter has a display name and a list of extensions:

```js
const filePath = await lightshell.dialog.open({
  filters: [
    { name: 'Images', extensions: ['png', 'jpg', 'jpeg', 'gif', 'webp'] },
    { name: 'Documents', extensions: ['pdf', 'docx', 'txt'] },
    { name: 'All Files', extensions: ['*'] }
  ]
})
```

### Selecting Multiple Files

Set `multiple: true` to allow multi-selection. The return value becomes an array:

```js
const files = await lightshell.dialog.open({
  title: 'Select Images',
  multiple: true,
  filters: [
    { name: 'Images', extensions: ['png', 'jpg', 'jpeg'] }
  ]
})

if (files) {
  for (const filePath of files) {
    await processImage(filePath)
  }
}
```

### Selecting a Directory

```js
const dirPath = await lightshell.dialog.open({
  title: 'Choose Project Folder',
  directory: true
})

if (dirPath) {
  const entries = await lightshell.fs.readDir(dirPath)
  renderFileTree(entries)
}
```

## File Save Dialog

### save(options?)

Show a native save dialog. Returns the chosen file path, or `null` if cancelled. This dialog does not write the file — it only returns the path the user selected.

```js
const savePath = await lightshell.dialog.save({
  title: 'Save Document',
  defaultPath: 'untitled.txt',
  filters: [
    { name: 'Text Files', extensions: ['txt'] },
    { name: 'Markdown', extensions: ['md'] }
  ]
})

if (savePath) {
  await lightshell.fs.writeFile(savePath, editor.value)
}
```

The `defaultPath` can include a directory to start in:

```js
const savePath = await lightshell.dialog.save({
  defaultPath: '/Users/me/Documents/report.pdf'
})
```

## Message Dialog

### message(title, body, type?)

Show a message box with an OK button. Use this for informational messages, warnings, and errors.

```js
// Simple info message
await lightshell.dialog.message('Export Complete', 'Your file has been saved to Downloads.')

// Warning
await lightshell.dialog.message('Warning', 'This action cannot be undone.', 'warning')

// Error
await lightshell.dialog.message('Error', 'Could not connect to the server.', 'error')
```

The `type` parameter controls the dialog icon:

| Type | Description |
|------|-------------|
| `'info'` | Information icon (default) |
| `'warning'` | Warning/caution icon |
| `'error'` | Error/stop icon |

## Confirmation Dialog

### confirm(title, body)

Show a dialog with OK and Cancel buttons. Returns `true` if the user clicked OK, `false` if they clicked Cancel.

```js
const confirmed = await lightshell.dialog.confirm(
  'Delete File',
  'Are you sure you want to delete "report.pdf"? This cannot be undone.'
)

if (confirmed) {
  await lightshell.fs.remove(filePath)
}
```

## Text Prompt

### prompt(title, body, defaultValue?)

Show a dialog with a text input field. Returns the entered string, or `null` if cancelled.

```js
const name = await lightshell.dialog.prompt(
  'New File',
  'Enter a name for the new file:',
  'untitled.txt'
)

if (name) {
  await lightshell.fs.writeFile(`${projectDir}/${name}`, '')
}
```

The third argument sets the default value pre-filled in the input field.

## Common Patterns

### Open-Edit-Save Workflow

A complete file editing flow with open, edit, and save:

```js
let currentPath = null
let hasUnsavedChanges = false

async function openFile() {
  const filePath = await lightshell.dialog.open({
    filters: [{ name: 'Text Files', extensions: ['txt', 'md', 'json'] }]
  })
  if (!filePath) return

  const content = await lightshell.fs.readFile(filePath)
  editor.value = content
  currentPath = filePath
  hasUnsavedChanges = false
  lightshell.window.setTitle(`${filePath.split('/').pop()} - My Editor`)
}

async function saveFile() {
  if (!currentPath) {
    return saveFileAs()
  }
  await lightshell.fs.writeFile(currentPath, editor.value)
  hasUnsavedChanges = false
}

async function saveFileAs() {
  const savePath = await lightshell.dialog.save({
    defaultPath: currentPath || 'untitled.txt',
    filters: [{ name: 'Text Files', extensions: ['txt', 'md'] }]
  })
  if (!savePath) return

  await lightshell.fs.writeFile(savePath, editor.value)
  currentPath = savePath
  hasUnsavedChanges = false
  lightshell.window.setTitle(`${savePath.split('/').pop()} - My Editor`)
}
```

### Unsaved Changes Guard

Prompt the user before closing if there are unsaved changes:

```js
window.addEventListener('beforeunload', async (e) => {
  if (!hasUnsavedChanges) return

  const confirmed = await lightshell.dialog.confirm(
    'Unsaved Changes',
    'You have unsaved changes. Do you want to quit without saving?'
  )

  if (!confirmed) {
    e.preventDefault()
  }
})
```

### Batch File Processing

Select multiple files and process them with progress feedback:

```js
async function batchConvert() {
  const files = await lightshell.dialog.open({
    title: 'Select Files to Convert',
    multiple: true,
    filters: [{ name: 'CSV Files', extensions: ['csv'] }]
  })
  if (!files) return

  for (let i = 0; i < files.length; i++) {
    try {
      const csv = await lightshell.fs.readFile(files[i])
      const json = csvToJson(csv)
      const outPath = files[i].replace(/\.csv$/, '.json')
      await lightshell.fs.writeFile(outPath, JSON.stringify(json, null, 2))
    } catch (err) {
      await lightshell.dialog.message(
        'Conversion Error',
        `Failed to convert ${files[i]}:\n${err.message}`,
        'error'
      )
    }
  }

  await lightshell.dialog.message(
    'Done',
    `Converted ${files.length} file(s) to JSON.`
  )
}
```

### Rename Prompt

Ask the user for a new file name:

```js
async function renameFile(oldPath) {
  const oldName = oldPath.split('/').pop()
  const dir = oldPath.substring(0, oldPath.lastIndexOf('/'))

  const newName = await lightshell.dialog.prompt(
    'Rename File',
    `Enter a new name for "${oldName}":`,
    oldName
  )

  if (!newName || newName === oldName) return

  const newPath = `${dir}/${newName}`
  const content = await lightshell.fs.readFile(oldPath)
  await lightshell.fs.writeFile(newPath, content)
  await lightshell.fs.remove(oldPath)
}
```

## Platform Notes

| Feature | macOS | Linux |
|---------|-------|-------|
| File picker | NSOpenPanel / NSSavePanel | GtkFileChooserDialog |
| Message box | NSAlert | GtkMessageDialog |
| Dialog style | Sheets (attached to window) | Standalone window |
| Filter display | Dropdown in dialog | Dropdown in dialog |
| Default path | Sets starting directory | Sets starting directory |

The dialog APIs normalize these differences. The visual appearance follows the native platform style, so dialogs will look at home on both macOS and Linux.

## Important Notes

- Dialog methods are async and block user interaction with the main window until dismissed.
- Cancelled dialogs return `null` (for `open`, `save`, `prompt`) or `false` (for `confirm`). They do not throw errors.
- Always check the return value before using it. A common mistake is assuming the user always picks a file.
- File filters are suggestions, not enforcement. On some platforms, users can override the filter and select any file type.
