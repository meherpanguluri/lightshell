---
title: "2. Native APIs"
description: Use file system, dialogs, and clipboard APIs to interact with the operating system.
---

LightShell provides native operating system capabilities through the `window.lightshell` global object. Every API is async and returns a Promise. In this tutorial, you will use the file system, dialog, and clipboard APIs to build a simple note-taking feature.

## The lightshell Global

When your app runs, LightShell injects a client library before your code executes. This makes `window.lightshell` (or just `lightshell`) available everywhere in your JavaScript. You never need to import anything.

```js
// All of these work — lightshell is a global
const data = await lightshell.fs.readFile('/tmp/test.txt')
const data2 = await window.lightshell.fs.readFile('/tmp/test.txt')
```

Every method is async. Always use `await` or `.then()`.

## Reading and Writing Files

The `lightshell.fs` module provides full file system access:

```js
// Write a file
await lightshell.fs.writeFile('/tmp/note.txt', 'Hello from LightShell!')

// Read it back
const content = await lightshell.fs.readFile('/tmp/note.txt')
console.log(content) // "Hello from LightShell!"

// Check if a file exists
const exists = await lightshell.fs.exists('/tmp/note.txt')
console.log(exists) // true
```

You can also list directories, get file info, and create directories:

```js
// List directory contents
const files = await lightshell.fs.readDir('/tmp')
console.log(files) // [{name: "note.txt", isDir: false, size: 22}, ...]

// Get file metadata
const info = await lightshell.fs.stat('/tmp/note.txt')
console.log(info.size)     // 22
console.log(info.modified)  // "2024-01-15T10:30:00Z"

// Create directories (including intermediate ones)
await lightshell.fs.mkdir('/tmp/my-app/data')
```

## Opening Files with Dialogs

Instead of hardcoding paths, use native file dialogs to let users choose files:

```js
// Open a file picker dialog
const path = await lightshell.dialog.open({
  title: 'Open a Note',
  filters: [
    { name: 'Text Files', extensions: ['txt', 'md'] },
    { name: 'All Files', extensions: ['*'] }
  ]
})

if (path) {
  // User selected a file — read it
  const content = await lightshell.fs.readFile(path)
  document.getElementById('editor').value = content
}
```

The dialog returns `null` if the user cancels. Always check for this.

For saving files, use `lightshell.dialog.save()`:

```js
const savePath = await lightshell.dialog.save({
  title: 'Save Note',
  defaultPath: 'untitled.txt',
  filters: [
    { name: 'Text Files', extensions: ['txt'] }
  ]
})

if (savePath) {
  const content = document.getElementById('editor').value
  await lightshell.fs.writeFile(savePath, content)
  await lightshell.dialog.message('Saved', `File saved to ${savePath}`)
}
```

## Message Dialogs

Show alerts, confirmations, and prompts:

```js
// Simple message
await lightshell.dialog.message('Hello', 'Welcome to the app!')

// Confirmation dialog — returns true/false
const confirmed = await lightshell.dialog.confirm(
  'Delete File',
  'Are you sure you want to delete this file?'
)

if (confirmed) {
  await lightshell.fs.remove(currentPath)
}

// Prompt dialog — returns the entered text or null
const name = await lightshell.dialog.prompt('New File', 'untitled.txt')
```

## Clipboard Access

Read from and write to the system clipboard:

```js
// Write text to clipboard
await lightshell.clipboard.write('Copied from LightShell!')

// Read text from clipboard
const text = await lightshell.clipboard.read()
console.log(text) // "Copied from LightShell!"
```

## Putting It Together: A Mini Note Editor

Here is a complete example combining file system, dialogs, and clipboard:

```js
let currentPath = null

async function openFile() {
  const path = await lightshell.dialog.open({
    filters: [{ name: 'Text', extensions: ['txt', 'md'] }]
  })
  if (!path) return

  currentPath = path
  const content = await lightshell.fs.readFile(path)
  document.getElementById('editor').value = content
  await lightshell.window.setTitle(`LightShell Notes — ${path}`)
}

async function saveFile() {
  if (!currentPath) {
    currentPath = await lightshell.dialog.save({
      defaultPath: 'note.txt',
      filters: [{ name: 'Text', extensions: ['txt'] }]
    })
    if (!currentPath) return
  }

  const content = document.getElementById('editor').value
  await lightshell.fs.writeFile(currentPath, content)
}

async function copyAll() {
  const content = document.getElementById('editor').value
  await lightshell.clipboard.write(content)
  await lightshell.dialog.message('Copied', 'All text copied to clipboard.')
}

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'o') { e.preventDefault(); openFile() }
  if ((e.metaKey || e.ctrlKey) && e.key === 's') { e.preventDefault(); saveFile() }
})
```

## Listening for Events

Some APIs emit events you can listen to. For example, window resize:

```js
lightshell.window.onResize((data) => {
  console.log(`Window resized to ${data.width}x${data.height}`)
})

lightshell.window.onFocus(() => {
  console.log('Window focused')
})
```

You can also watch files for changes:

```js
const unwatch = lightshell.fs.watch('/tmp/note.txt', (event) => {
  console.log('File changed:', event)
  // Reload file contents
})
```

## Key Points

- All APIs live under `window.lightshell` — no imports needed
- Every method is async — always `await` the result
- Dialogs return `null` when the user cancels
- File paths are absolute OS paths, not URLs
- Use `lightshell.app.dataDir()` to get a persistent storage directory for your app

Next, let's learn how to style your app for a native look.
