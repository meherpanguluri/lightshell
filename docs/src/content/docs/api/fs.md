---
title: File System API
description: Complete reference for lightshell.fs — read, write, and manage files.
---

The `lightshell.fs` module provides file system access. All paths must be absolute OS paths. All methods are async and return Promises.

## Methods

### readFile(path, encoding?)

Read the entire contents of a file as a string.

**Parameters:**
- `path` (string) — absolute path to the file
- `encoding` (string, optional) — character encoding, defaults to `'utf-8'`

**Returns:** `Promise<string>` — the file contents

**Example:**
```js
const content = await lightshell.fs.readFile('/Users/me/notes.txt')
console.log(content)
```

**Errors:** Rejects if the file does not exist, is a directory, or is not readable.

---

### writeFile(path, data)

Write a string to a file. Creates the file if it does not exist. Overwrites existing content entirely.

**Parameters:**
- `path` (string) — absolute path to the file
- `data` (string) — the content to write

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.fs.writeFile('/tmp/output.txt', 'Hello, world!')

// Write JSON
const data = { count: 42, items: ['a', 'b', 'c'] }
await lightshell.fs.writeFile('/tmp/data.json', JSON.stringify(data, null, 2))
```

**Errors:** Rejects if the parent directory does not exist or the path is not writable.

---

### readDir(path)

List the contents of a directory.

**Parameters:**
- `path` (string) — absolute path to the directory

**Returns:** `Promise<Array<{ name: string, isDir: boolean, size: number }>>` — array of directory entries

**Example:**
```js
const entries = await lightshell.fs.readDir('/Users/me/Documents')
for (const entry of entries) {
  if (entry.isDir) {
    console.log(`Directory: ${entry.name}`)
  } else {
    console.log(`File: ${entry.name} (${entry.size} bytes)`)
  }
}
```

**Errors:** Rejects if the path does not exist or is not a directory.

---

### exists(path)

Check whether a file or directory exists at the given path.

**Parameters:**
- `path` (string) — absolute path to check

**Returns:** `Promise<boolean>` — `true` if the path exists, `false` otherwise

**Example:**
```js
const configExists = await lightshell.fs.exists('/tmp/app/config.json')
if (!configExists) {
  await lightshell.fs.writeFile('/tmp/app/config.json', '{}')
}
```

---

### stat(path)

Get metadata about a file or directory.

**Parameters:**
- `path` (string) — absolute path to the file or directory

**Returns:** `Promise<{ name: string, size: number, isDir: boolean, modified: string }>` — file metadata. The `modified` field is an ISO 8601 timestamp string.

**Example:**
```js
const info = await lightshell.fs.stat('/Users/me/photo.jpg')
console.log(`Size: ${info.size} bytes`)
console.log(`Last modified: ${info.modified}`)
console.log(`Is directory: ${info.isDir}`)
```

**Errors:** Rejects if the path does not exist.

---

### mkdir(path)

Create a directory, including any intermediate directories that do not exist (equivalent to `mkdir -p`).

**Parameters:**
- `path` (string) — absolute path of the directory to create

**Returns:** `Promise<void>`

**Example:**
```js
// Creates both 'my-app' and 'data' if they don't exist
await lightshell.fs.mkdir('/tmp/my-app/data')
```

**Errors:** Rejects if the path is not writable.

---

### remove(path)

Delete a file or directory.

**Parameters:**
- `path` (string) — absolute path to delete

**Returns:** `Promise<void>`

**Example:**
```js
// Remove a file
await lightshell.fs.remove('/tmp/old-data.txt')

// Remove a directory and its contents
await lightshell.fs.remove('/tmp/cache')
```

**Errors:** Rejects if the path does not exist or is not writable.

---

### watch(path, callback)

Watch a file or directory for changes. The callback is invoked when the watched path is modified, created, or deleted.

**Parameters:**
- `path` (string) — absolute path to watch
- `callback` (function) — receives an event object with change details

**Returns:** unsubscribe function — call it to stop watching

**Example:**
```js
const unwatch = lightshell.fs.watch('/tmp/data.json', (event) => {
  console.log('File changed:', event)
  reloadData()
})

// Later, stop watching
unwatch()
```

---

## Common Patterns

### Read JSON Configuration

```js
async function loadConfig(path) {
  try {
    const raw = await lightshell.fs.readFile(path)
    return JSON.parse(raw)
  } catch (err) {
    console.warn('Could not load config, using defaults:', err.message)
    return { theme: 'light' }
  }
}
```

### Save with User Dialog

```js
async function saveAs(content) {
  const path = await lightshell.dialog.save({
    title: 'Save File',
    defaultPath: 'document.txt',
    filters: [{ name: 'Text', extensions: ['txt'] }]
  })
  if (path) {
    await lightshell.fs.writeFile(path, content)
  }
}
```

### Recursive Directory Listing

```js
async function listRecursive(dirPath, prefix = '') {
  const entries = await lightshell.fs.readDir(dirPath)
  for (const entry of entries) {
    console.log(`${prefix}${entry.isDir ? '/' : ''}${entry.name}`)
    if (entry.isDir) {
      await listRecursive(`${dirPath}/${entry.name}`, prefix + '  ')
    }
  }
}
```
