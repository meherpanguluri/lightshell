---
title: File System
description: Complete guide to reading, writing, and managing files with lightshell.fs.
---

The `lightshell.fs` module provides full file system access from JavaScript. All methods are async and work with absolute OS paths. This guide covers every operation with practical examples.

## Reading Files

### readFile(path, encoding?)

Read the contents of a file as a string.

```js
// Read a text file (UTF-8 by default)
const content = await lightshell.fs.readFile('/Users/me/notes/todo.txt')

// Specify encoding explicitly
const data = await lightshell.fs.readFile('/tmp/data.csv', 'utf-8')
```

If the file does not exist, the promise rejects with an error. Always handle this:

```js
try {
  const content = await lightshell.fs.readFile(path)
  // use content
} catch (err) {
  console.error('Could not read file:', err.message)
}
```

## Writing Files

### writeFile(path, data)

Write a string to a file. Creates the file if it does not exist. Overwrites existing content.

```js
// Write text to a file
await lightshell.fs.writeFile('/tmp/output.txt', 'Hello, world!')

// Write JSON data
const config = { theme: 'dark', fontSize: 14 }
await lightshell.fs.writeFile(
  '/tmp/config.json',
  JSON.stringify(config, null, 2)
)
```

The parent directory must exist. Use `mkdir` first if needed:

```js
await lightshell.fs.mkdir('/tmp/my-app/data')
await lightshell.fs.writeFile('/tmp/my-app/data/settings.json', '{}')
```

## Checking File Existence

### exists(path)

Returns `true` if the path exists (file or directory), `false` otherwise.

```js
const hasConfig = await lightshell.fs.exists('/tmp/my-app/config.json')
if (!hasConfig) {
  await lightshell.fs.writeFile('/tmp/my-app/config.json', '{}')
}
```

## File Metadata

### stat(path)

Get metadata about a file or directory.

```js
const info = await lightshell.fs.stat('/tmp/report.pdf')
console.log(info)
// {
//   name: "report.pdf",
//   size: 245760,        // bytes
//   isDir: false,
//   modified: "2024-01-15T10:30:00Z"
// }
```

Useful for checking file size before reading, or displaying last modified dates.

## Listing Directories

### readDir(path)

List the contents of a directory. Returns an array of entries.

```js
const entries = await lightshell.fs.readDir('/Users/me/Documents')
// [
//   { name: "notes", isDir: true, size: 0 },
//   { name: "report.pdf", isDir: false, size: 245760 },
//   { name: "photo.jpg", isDir: false, size: 1048576 }
// ]
```

Filter by type:

```js
const entries = await lightshell.fs.readDir('/Users/me/Documents')
const folders = entries.filter(e => e.isDir)
const files = entries.filter(e => !e.isDir)
```

Build a file browser:

```js
async function renderFileList(dirPath) {
  const entries = await lightshell.fs.readDir(dirPath)
  const list = document.getElementById('file-list')
  list.innerHTML = ''

  for (const entry of entries) {
    const li = document.createElement('li')
    li.textContent = entry.isDir ? `📁 ${entry.name}` : `📄 ${entry.name}`
    li.addEventListener('click', () => {
      if (entry.isDir) {
        renderFileList(`${dirPath}/${entry.name}`)
      } else {
        openFile(`${dirPath}/${entry.name}`)
      }
    })
    list.appendChild(li)
  }
}
```

## Creating Directories

### mkdir(path)

Create a directory. Creates intermediate directories as needed (like `mkdir -p`).

```js
// Creates both 'my-app' and 'data' if they don't exist
await lightshell.fs.mkdir('/tmp/my-app/data')
```

## Removing Files and Directories

### remove(path)

Delete a file or directory.

```js
// Remove a file
await lightshell.fs.remove('/tmp/old-data.txt')

// Remove a directory (must be empty, or removes recursively)
await lightshell.fs.remove('/tmp/my-app/cache')
```

Always confirm with the user before deleting:

```js
async function deleteFile(path) {
  const confirmed = await lightshell.dialog.confirm(
    'Delete File',
    `Are you sure you want to delete ${path}?`
  )
  if (confirmed) {
    await lightshell.fs.remove(path)
  }
}
```

## Watching for Changes

### watch(path, callback)

Watch a file or directory for changes. The callback fires when the file is modified, created, or deleted.

```js
const unwatch = lightshell.fs.watch('/tmp/data.json', (event) => {
  console.log('File changed:', event)
  // Reload the file
  loadData()
})
```

The `watch` function returns an unsubscribe function. Call it to stop watching:

```js
// Start watching
const unwatch = lightshell.fs.watch(path, onChange)

// Later, stop watching
unwatch()
```

## Common Patterns

### App Data Storage

Use `lightshell.app.dataDir()` for persistent app data:

```js
async function loadSettings() {
  const dataDir = await lightshell.app.dataDir()
  const settingsPath = `${dataDir}/settings.json`

  const exists = await lightshell.fs.exists(settingsPath)
  if (exists) {
    const raw = await lightshell.fs.readFile(settingsPath)
    return JSON.parse(raw)
  }
  return { theme: 'light', fontSize: 14 } // defaults
}

async function saveSettings(settings) {
  const dataDir = await lightshell.app.dataDir()
  await lightshell.fs.mkdir(dataDir) // ensure it exists
  await lightshell.fs.writeFile(
    `${dataDir}/settings.json`,
    JSON.stringify(settings, null, 2)
  )
}
```

### Recent Files List

Track recently opened files:

```js
async function addToRecent(filePath) {
  const dataDir = await lightshell.app.dataDir()
  const recentPath = `${dataDir}/recent.json`

  let recent = []
  if (await lightshell.fs.exists(recentPath)) {
    recent = JSON.parse(await lightshell.fs.readFile(recentPath))
  }

  // Add to front, deduplicate, limit to 10
  recent = [filePath, ...recent.filter(p => p !== filePath)].slice(0, 10)
  await lightshell.fs.writeFile(recentPath, JSON.stringify(recent))
}
```

### Safe File Writing

Write to a temporary file first, then rename, to prevent corruption:

```js
async function safeWrite(path, data) {
  const tmpPath = `${path}.tmp`
  await lightshell.fs.writeFile(tmpPath, data)
  await lightshell.fs.remove(path)
  // Note: atomic rename not yet available — this is a best-effort approach
  await lightshell.fs.writeFile(path, data)
  await lightshell.fs.remove(tmpPath)
}
```

## Error Handling

All `fs` methods reject on error. Common errors:

| Error | Cause |
|-------|-------|
| File not found | Path does not exist |
| Permission denied | No read/write access |
| Is a directory | Tried to `readFile` on a directory |
| Not a directory | Tried to `readDir` on a file |
| Directory not empty | Tried to `remove` a non-empty directory |

Always wrap file operations in try/catch:

```js
try {
  await lightshell.fs.writeFile(path, data)
} catch (err) {
  await lightshell.dialog.message('Error', `Could not save: ${err.message}`)
}
```
