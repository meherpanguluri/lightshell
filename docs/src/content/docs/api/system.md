---
title: System API
description: Complete reference for lightshell.system — operating system information and paths.
---

The `lightshell.system` module provides information about the operating system, CPU architecture, and standard system paths. All methods are async and return Promises.

## Methods

### platform()

Get the current operating system identifier.

**Parameters:** none

**Returns:** `Promise<string>` — `"darwin"` on macOS, `"linux"` on Linux

**Example:**
```js
const os = await lightshell.system.platform()
if (os === 'darwin') {
  console.log('Running on macOS')
} else if (os === 'linux') {
  console.log('Running on Linux')
}
```

---

### arch()

Get the CPU architecture.

**Parameters:** none

**Returns:** `Promise<string>` — `"arm64"` or `"amd64"` (also known as x64)

**Example:**
```js
const arch = await lightshell.system.arch()
console.log(`Architecture: ${arch}`) // "arm64" or "amd64"
```

---

### homeDir()

Get the current user's home directory path.

**Parameters:** none

**Returns:** `Promise<string>` — absolute path to the home directory

**Example:**
```js
const home = await lightshell.system.homeDir()
console.log(home) // "/Users/alice" on macOS, "/home/alice" on Linux
```

---

### tempDir()

Get the system temporary directory path.

**Parameters:** none

**Returns:** `Promise<string>` — absolute path to the temp directory

**Example:**
```js
const tmp = await lightshell.system.tempDir()
await lightshell.fs.writeFile(`${tmp}/scratch.txt`, 'temporary data')
```

---

### hostname()

Get the system hostname.

**Parameters:** none

**Returns:** `Promise<string>` — the hostname of the machine

**Example:**
```js
const name = await lightshell.system.hostname()
console.log(`Running on: ${name}`)
```

---

## Common Patterns

### Platform-Specific Behavior

```js
async function getDefaultSavePath(filename) {
  const platform = await lightshell.system.platform()
  const home = await lightshell.system.homeDir()

  if (platform === 'darwin') {
    return `${home}/Documents/${filename}`
  } else {
    return `${home}/${filename}`
  }
}
```

### System Info Panel

```js
async function getSystemInfo() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  const hostname = await lightshell.system.hostname()
  const home = await lightshell.system.homeDir()
  const temp = await lightshell.system.tempDir()

  return {
    os: platform === 'darwin' ? 'macOS' : 'Linux',
    arch,
    hostname,
    home,
    temp
  }
}

// Display in UI
const info = await getSystemInfo()
document.getElementById('sys-info').innerHTML = `
  <p>OS: ${info.os} (${info.arch})</p>
  <p>Hostname: ${info.hostname}</p>
  <p>Home: ${info.home}</p>
`
```

### Platform-Conditional UI

```js
async function setupUI() {
  const platform = await lightshell.system.platform()

  if (platform === 'darwin') {
    // macOS uses Cmd key
    document.getElementById('shortcut-hint').textContent = 'Cmd+S to save'
  } else {
    // Linux uses Ctrl key
    document.getElementById('shortcut-hint').textContent = 'Ctrl+S to save'
  }
}
```

## Platform Notes

- On macOS, `platform()` returns `"darwin"` (the underlying OS name), not `"macos"`
- `arch()` returns Go-style architecture names: `"arm64"` for Apple Silicon, `"amd64"` for Intel
- `homeDir()` returns `/Users/{user}` on macOS and `/home/{user}` on Linux
- `tempDir()` returns `/tmp` on both platforms
- All return values are strings, not objects — they resolve directly to the value
