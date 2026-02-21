---
title: Error Handling
description: Handle errors gracefully in your LightShell app.
---

All `lightshell.*` APIs are async and reject their promises on failure. This guide covers how to handle errors, what to expect from each API, and how to build resilient apps.

## Basic Error Handling

Every LightShell API call should be wrapped in a try/catch block:

```js
try {
  const content = await lightshell.fs.readFile('/path/to/file.txt')
  // use content
} catch (err) {
  console.error('Failed to read file:', err.message)
}
```

The error object has a `message` property with a human-readable description of what went wrong.

## Error Format

LightShell errors follow a structured, AI-friendly format:

```
LightShell Error [fs.readFile]: Permission denied
  -> Attempted to read: /etc/passwd
  -> Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  -> To allow this path, update permissions.fs.read in lightshell.json
  -> Docs: https://lightshell.dev/docs/api/permissions#fs
```

Each error includes:
- The API method that failed (`fs.readFile`)
- What was attempted
- What is currently allowed (for permission errors)
- A concrete fix
- A link to the relevant documentation

This format makes errors easy to diagnose manually and straightforward for AI assistants to fix automatically.

## Common Errors by API

### File System Errors

```js
// File not found
try {
  await lightshell.fs.readFile('/nonexistent/file.txt')
} catch (err) {
  // err.message: "File not found: /nonexistent/file.txt"
}

// Permission denied (restricted mode)
try {
  await lightshell.fs.readFile('/etc/shadow')
} catch (err) {
  // err.message: "LightShell Error [fs.readFile]: Permission denied ..."
}

// Writing to a read-only location
try {
  await lightshell.fs.writeFile('/usr/bin/test', 'data')
} catch (err) {
  // err.message: OS-level permission error
}

// Parent directory does not exist
try {
  await lightshell.fs.writeFile('/tmp/nonexistent-dir/file.txt', 'data')
} catch (err) {
  // Create the directory first, then retry
  await lightshell.fs.mkdir('/tmp/nonexistent-dir')
  await lightshell.fs.writeFile('/tmp/nonexistent-dir/file.txt', 'data')
}
```

### Dialog Returns (Not Errors)

Dialog methods do not throw when the user cancels. They return `null` or `false`:

```js
// open() returns null if the user cancels
const filePath = await lightshell.dialog.open()
if (filePath === null) {
  // User cancelled -- not an error, just do nothing
  return
}

// save() returns null if cancelled
const savePath = await lightshell.dialog.save()
if (!savePath) return

// confirm() returns false if cancelled
const confirmed = await lightshell.dialog.confirm('Delete?', 'Are you sure?')
if (!confirmed) return

// prompt() returns null if cancelled
const input = await lightshell.dialog.prompt('Name', 'Enter your name:')
if (input === null) return
```

A common mistake is wrapping dialog calls in try/catch expecting an error on cancel. The cancellation is not an error -- always check the return value instead.

### HTTP Errors

```js
try {
  const response = await lightshell.http.fetch('https://api.example.com/data')

  // A non-2xx status is NOT an error -- the promise resolves with the status code
  if (response.status === 404) {
    console.log('Resource not found')
    return
  }
  if (response.status >= 400) {
    console.log('Server error:', response.status, response.body)
    return
  }

  const data = JSON.parse(response.body)
} catch (err) {
  // True errors: network failure, DNS resolution failure, timeout, permission denied
  console.error('Request failed:', err.message)
}
```

Note that HTTP status codes like 404 or 500 do not cause a rejection. The promise resolves with the response, and you check `response.status` yourself. The promise only rejects for actual failures like network errors, timeouts, or permission denials.

### Process Execution Errors

```js
try {
  const result = await lightshell.process.exec('git', ['status'])

  // A non-zero exit code is NOT an error -- check result.code
  if (result.code !== 0) {
    console.error('git failed:', result.stderr)
    return
  }

  console.log(result.stdout)
} catch (err) {
  // True errors: command not found, permission denied, timeout
  console.error('Could not run command:', err.message)
}
```

Similar to HTTP, a command that runs but exits with a non-zero code is not an error from LightShell's perspective. The promise resolves with `{ stdout, stderr, code }`. Only check the `code` property.

### Store Errors

```js
// get() returns null if the key does not exist -- not an error
const value = await lightshell.store.get('nonexistent-key')
// value === null

// has() returns false for missing keys
const exists = await lightshell.store.has('nonexistent-key')
// exists === false

// delete() does not error if the key does not exist
await lightshell.store.delete('nonexistent-key')
// no error
```

The store API is designed to be forgiving. Missing keys return `null` or `false` rather than throwing.

## Global Error Handling

Catch unhandled promise rejections across your entire app:

```js
window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled error:', event.reason)

  // Show a user-friendly error dialog
  lightshell.dialog.message(
    'Something went wrong',
    event.reason?.message || String(event.reason),
    'error'
  )

  // Prevent the default browser behavior (console error)
  event.preventDefault()
})
```

This catches any rejected promise that does not have a `.catch()` or try/catch handler. It is a safety net, not a replacement for proper per-call error handling.

### Global Error Handler for Synchronous Errors

```js
window.addEventListener('error', (event) => {
  console.error('Runtime error:', event.message)
  lightshell.dialog.message('Error', event.message, 'error')
})
```

## User-Friendly Error Dialogs

Use `lightshell.dialog.message` to show errors in a native dialog instead of silently logging to the console:

```js
async function loadFile(path) {
  try {
    const content = await lightshell.fs.readFile(path)
    return content
  } catch (err) {
    await lightshell.dialog.message(
      'Could not open file',
      `Failed to read "${path}":\n${err.message}`,
      'error'
    )
    return null
  }
}
```

### Error Reporting Pattern

For apps that need to report errors, build a helper function:

```js
async function reportError(action, err) {
  const message = [
    `Action: ${action}`,
    `Error: ${err.message}`,
    `Time: ${new Date().toISOString()}`,
  ].join('\n')

  console.error(message)

  await lightshell.dialog.message('Error', message, 'error')
}

// Usage
try {
  await lightshell.fs.writeFile(path, data)
} catch (err) {
  await reportError('Save file', err)
}
```

## Defensive Patterns

### Provide Fallback Defaults

When loading settings or data, always fall back to defaults if the read fails:

```js
async function loadSettings() {
  try {
    const raw = await lightshell.fs.readFile(settingsPath)
    return JSON.parse(raw)
  } catch {
    // File missing, corrupted, or permission denied -- use defaults
    return { theme: 'light', fontSize: 14, sidebarOpen: true }
  }
}
```

### Retry Logic

For network operations that may fail transiently:

```js
async function fetchWithRetry(url, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const response = await lightshell.http.fetch(url)
      if (response.status >= 500 && attempt < maxRetries) {
        await new Promise(r => setTimeout(r, 1000 * attempt))
        continue
      }
      return response
    } catch (err) {
      if (attempt === maxRetries) throw err
      await new Promise(r => setTimeout(r, 1000 * attempt))
    }
  }
}
```

### Safe JSON Parsing

Wrap `JSON.parse` when dealing with user data or file contents:

```js
function safeParseJSON(str, fallback = null) {
  try {
    return JSON.parse(str)
  } catch {
    return fallback
  }
}

const data = safeParseJSON(await lightshell.fs.readFile(path), [])
```

### Guard Against Null Dialog Returns

Always check dialog results before using them:

```js
// Bad -- will crash if user cancels
const path = await lightshell.dialog.open()
const content = await lightshell.fs.readFile(path)  // path is null!

// Good
const path = await lightshell.dialog.open()
if (!path) return
const content = await lightshell.fs.readFile(path)
```

## Best Practices

**Wrap every `fs` call in try/catch.** File operations are the most common source of runtime errors (missing files, permission issues, disk full).

**Check dialog returns for null.** Cancelled dialogs are normal user behavior, not errors. Never assume the user will always pick a file or confirm an action.

**Distinguish error types for HTTP and process.** Promise rejections mean the operation could not be performed at all. Non-2xx status codes and non-zero exit codes mean the operation ran but returned an unsuccessful result. Handle both.

**Use a global error handler as a safety net.** The `unhandledrejection` listener catches anything you missed. Log it and show a dialog so the user knows something went wrong.

**Show errors in native dialogs, not just the console.** Users cannot see the developer console. Use `lightshell.dialog.message` with `'error'` type to surface problems visibly.

**Provide defaults for everything.** If loading saved data fails, fall back to sensible defaults. The app should always start, even if its data is missing or corrupted.
