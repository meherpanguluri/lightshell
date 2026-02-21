---
title: Error Codes
description: Complete reference for LightShell error codes and troubleshooting.
---

LightShell returns structured, AI-friendly error messages designed to be actionable. Every error includes what was attempted, what went wrong, and how to fix it. This page documents all error types and their solutions.

## Error Format

All LightShell errors follow this format:

```
LightShell Error [{method}]: {message}
  -> Attempted: {what was tried}
  -> {context-specific details}
  -> To fix: {actionable suggestion}
  -> Docs: {link to relevant documentation}
```

In JavaScript, errors are thrown as standard `Error` objects. The `message` property contains the full formatted error text.

```js
try {
  await lightshell.fs.readFile('/etc/passwd')
} catch (err) {
  console.error(err.message)
  // LightShell Error [fs.readFile]: Permission denied
  //   -> Attempted to read: /etc/passwd
  //   -> Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  //   -> To allow this path, update permissions.fs.read in lightshell.json
  //   -> Docs: https://lightshell.dev/docs/api/permissions#fs
}
```

---

## Permission Errors

### fs.readFile / fs.writeFile / fs.readDir — Permission denied

The app attempted to access a file path that is not in the allowed list.

**Error:**
```
LightShell Error [fs.readFile]: Permission denied
  -> Attempted to read: /etc/passwd
  -> Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  -> To allow this path, update permissions.fs.read in lightshell.json
  -> Docs: https://lightshell.dev/docs/api/config#permissions
```

**Cause:** The `permissions.fs.read` (or `permissions.fs.write`) array in `lightshell.json` does not include a pattern that matches the requested path. This only occurs in restricted mode (when a `permissions` key exists).

**Solution:** Add the path or a matching glob pattern to the appropriate permission array in `lightshell.json`:
```json
{
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/**", "/etc/passwd"]
    }
  }
}
```

Or remove the `permissions` key entirely to run in permissive mode.

---

### fs.readFile / fs.writeFile — Path traversal blocked

The app attempted to access a path that resolves outside the allowed directories after resolving symlinks and `..` segments.

**Error:**
```
LightShell Error [fs.readFile]: Path traversal blocked
  -> Requested: $APP_DATA/../../../etc/shadow
  -> Resolved to: /etc/shadow
  -> This path is outside all allowed directories
  -> Path traversal protection cannot be disabled
```

**Cause:** The path contains `..` segments or symlinks that resolve to a location outside the allowed directories. This is a security mechanism that is always active, even in permissive mode.

**Solution:** Use direct absolute paths instead of paths with `..` segments. If you need to access the target path, add it explicitly to the permissions.

---

### process.exec — Command not allowed

The app attempted to run a command that is not in the allowed list.

**Error:**
```
LightShell Error [process.exec]: Command not allowed
  -> Attempted: rm -rf /
  -> Allowed commands: git (status, log, diff), python3 (*)
  -> To allow this command, update permissions.process.exec in lightshell.json
  -> Docs: https://lightshell.dev/docs/api/config#permissions
```

**Cause:** The `permissions.process.exec` array does not include an entry for the requested command and arguments.

**Solution:** Add the command to the permission list:
```json
{
  "permissions": {
    "process": {
      "exec": [
        { "cmd": "rm", "args": ["-rf"] }
      ]
    }
  }
}
```

---

### http.fetch — URL not allowed

The app attempted to make an HTTP request to a URL not in the allowed list.

**Error:**
```
LightShell Error [http.fetch]: URL not allowed
  -> Attempted: https://evil.com/steal-data
  -> Allowed URL patterns: https://api.example.com/**
  -> To allow this URL, update permissions.http.allow in lightshell.json
```

**Cause:** The `permissions.http.allow` array does not include a pattern matching the requested URL, or the URL matches a pattern in `permissions.http.deny`.

**Solution:** Add the URL pattern to `permissions.http.allow`:
```json
{
  "permissions": {
    "http": {
      "allow": ["https://api.example.com/**", "https://evil.com/**"]
    }
  }
}
```

---

## File System Errors

### fs.readFile — File not found

**Error:**
```
LightShell Error [fs.readFile]: File not found
  -> Path: /Users/alice/missing.txt
  -> The file does not exist at this path
```

**Cause:** The specified file does not exist.

**Solution:** Check the path for typos. Use `lightshell.fs.exists()` before reading:
```js
if (await lightshell.fs.exists(path)) {
  const content = await lightshell.fs.readFile(path)
}
```

---

### fs.writeFile — Parent directory does not exist

**Error:**
```
LightShell Error [fs.writeFile]: Parent directory does not exist
  -> Path: /Users/alice/nonexistent/dir/file.txt
  -> Create the directory first with lightshell.fs.mkdir()
```

**Cause:** The parent directory of the target file does not exist.

**Solution:** Create the directory first:
```js
await lightshell.fs.mkdir('/Users/alice/nonexistent/dir')
await lightshell.fs.writeFile('/Users/alice/nonexistent/dir/file.txt', 'content')
```

---

## Updater Errors

### updater.install — SHA256 verification failed

**Error:**
```
LightShell Error [updater.install]: SHA256 verification failed
  -> Expected: a1b2c3d4e5f6...
  -> Got: x9y8z7w6v5u4...
  -> The downloaded file may be corrupted or tampered with
  -> The update has been rejected. Try again or verify the manifest.
```

**Cause:** The SHA256 hash of the downloaded archive does not match the hash in the update manifest. This could indicate a corrupted download, a man-in-the-middle attack, or an incorrectly generated manifest.

**Solution:** Regenerate the update manifest with correct hashes, verify the download URL serves the correct file, and ensure the endpoint uses HTTPS.

---

### updater.check — HTTPS required

**Error:**
```
LightShell Error [updater.check]: HTTPS required
  -> Endpoint: http://releases.myapp.com/latest.json
  -> Production builds require HTTPS for update endpoints
  -> Change the endpoint URL to use https://
```

**Cause:** The update endpoint in `lightshell.json` uses HTTP instead of HTTPS, and this is a production build.

**Solution:** Change the endpoint to HTTPS in `lightshell.json`:
```json
{
  "updater": {
    "endpoint": "https://releases.myapp.com/latest.json"
  }
}
```

---

## Network Errors

### http.fetch — Network error

**Error:**
```
LightShell Error [http.fetch]: Network error
  -> URL: https://api.example.com/data
  -> Cause: connection refused
  -> Check that the server is running and accessible
```

**Cause:** The HTTP request failed at the network level — DNS resolution failure, connection refused, connection timeout, or TLS error.

**Solution:** Verify the URL is correct, the server is running, and your network connection is active. Check for firewall rules that might block outgoing connections.

---

### http.fetch — Request timeout

**Error:**
```
LightShell Error [http.fetch]: Request timeout
  -> URL: https://api.example.com/slow-endpoint
  -> Timeout: 30000ms
  -> Increase the timeout or check the server
```

**Cause:** The server did not respond within the timeout period.

**Solution:** Increase the timeout:
```js
const response = await lightshell.http.fetch('https://api.example.com/slow', {
  timeout: 60000 // 60 seconds
})
```

---

## Common Patterns

### Error Handling Wrapper

```js
async function safeCall(fn, fallback = null) {
  try {
    return await fn()
  } catch (err) {
    console.error(err.message)

    if (err.message.includes('Permission denied')) {
      await lightshell.dialog.message(
        'Permission Error',
        'This action is not allowed by the app permissions. Check lightshell.json.'
      )
    } else if (err.message.includes('File not found')) {
      // Silently return fallback for missing files
    } else {
      await lightshell.dialog.message('Error', err.message)
    }

    return fallback
  }
}

// Usage
const content = await safeCall(
  () => lightshell.fs.readFile('/path/to/file.txt'),
  'default content'
)
```

### Categorizing Errors

```js
function getErrorType(err) {
  const msg = err.message
  if (msg.includes('Permission denied') || msg.includes('not allowed')) return 'permission'
  if (msg.includes('not found') || msg.includes('does not exist')) return 'not_found'
  if (msg.includes('Path traversal')) return 'security'
  if (msg.includes('Network error') || msg.includes('timeout')) return 'network'
  if (msg.includes('SHA256')) return 'integrity'
  return 'unknown'
}
```
