---
title: Error Codes
description: Complete reference for LightShell error codes and troubleshooting.
---

LightShell returns structured, AI-friendly error messages designed to be actionable. Every error includes what was attempted, what went wrong, and how to fix it. This page documents all error types, their codes, and solutions.

## Error Format

All LightShell errors are thrown as standard JavaScript `Error` objects with two additional properties: `code` (a machine-readable error code string) and `method` (the API method that produced the error).

```js
try {
  await lightshell.fs.readFile('/etc/passwd')
} catch (err) {
  console.log(err.code)    // "FS_PERMISSION_DENIED"
  console.log(err.method)  // "fs.readFile"
  console.log(err.message)
  // LightShell Error [fs.readFile]: Permission denied
  //   -> Attempted to read: /etc/passwd
  //   -> Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  //   -> To allow this path, update permissions.fs.read in lightshell.json
  //   -> Docs: https://lightshell.dev/docs/api/config#permissions
}
```

The `message` property always follows this structure:

```
LightShell Error [{method}]: {summary}
  -> Attempted: {what was tried}
  -> {context-specific details}
  -> To fix: {actionable suggestion}
  -> Docs: {link to relevant documentation}
```

---

## Error Code Catalog

### File System Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `FS_NOT_FOUND` | `fs.readFile`, `fs.stat`, `fs.readDir`, `fs.remove` | The file or directory does not exist at the specified path. |
| `FS_PERMISSION_DENIED` | `fs.readFile`, `fs.writeFile`, `fs.readDir`, `fs.mkdir`, `fs.remove`, `fs.watch` | The path is not in the allowed list (restricted mode). |
| `FS_IS_DIRECTORY` | `fs.readFile`, `fs.writeFile` | Expected a file but the path points to a directory. |
| `FS_IS_FILE` | `fs.readDir`, `fs.mkdir` | Expected a directory but the path points to a file. |
| `FS_PARENT_NOT_FOUND` | `fs.writeFile` | The parent directory of the target path does not exist. |
| `FS_PATH_TRAVERSAL` | All `fs.*` methods | The path resolves outside allowed directories after following symlinks and `..` segments. |
| `FS_NOT_EMPTY` | `fs.remove` | Attempted to remove a non-empty directory without the `recursive` option. |
| `FS_ALREADY_EXISTS` | `fs.mkdir` | The directory already exists (only when `existsOk` is `false`). |
| `FS_DISK_FULL` | `fs.writeFile`, `fs.mkdir` | No space left on the device. |

---

### Permission Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `PERMISSION_DENIED` | Any API method | The operation is blocked by the permission engine. The error message includes the specific permission rule that was violated. |
| `PERMISSION_NOT_ENABLED` | Any API method requiring a specific capability | The API requires a permission that is not declared in `lightshell.json`. For example, calling `process.exec` when `permissions.process.exec` is not configured. |

---

### HTTP Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `HTTP_TIMEOUT` | `http.fetch`, `http.download` | The request exceeded the timeout period. Default timeout is 30 seconds. |
| `HTTP_NETWORK_ERROR` | `http.fetch`, `http.download` | Network-level failure: DNS resolution failed, connection refused, TLS error, or no internet. |
| `HTTP_INVALID_URL` | `http.fetch`, `http.download` | The provided URL is not a valid HTTP or HTTPS URL. |
| `HTTP_URL_NOT_ALLOWED` | `http.fetch`, `http.download` | The URL does not match any pattern in `permissions.http.allow` (restricted mode). |

---

### Process Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `PROCESS_TIMEOUT` | `process.exec` | The command did not complete within the specified timeout. |
| `PROCESS_NOT_FOUND` | `process.exec` | The command binary was not found in the system PATH. |
| `PROCESS_EXIT_CODE` | `process.exec` | The command exited with a non-zero status code. The `code` property on the result object contains the exit code; this error is only thrown when the process fails to start, not for non-zero exits (those are returned in the result). |
| `PROCESS_NOT_ALLOWED` | `process.exec` | The command is not in the `permissions.process.exec` allow list (restricted mode). |

---

### Platform Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `PLATFORM_UNSUPPORTED` | Various | The called method is not supported on the current platform. For example, `window.setVibrancy()` on Linux or `app.setBadgeCount()` on Linux. |

---

### Updater Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `UPDATER_SHA256_MISMATCH` | `updater.install` | The SHA256 hash of the downloaded archive does not match the hash in the manifest. |
| `UPDATER_SIGNATURE_INVALID` | `updater.install` | The Ed25519 signature verification failed. |
| `UPDATER_HTTPS_REQUIRED` | `updater.check`, `updater.install` | The endpoint or download URL uses HTTP in a production build. |
| `UPDATER_MANIFEST_INVALID` | `updater.check` | The manifest JSON is malformed or missing required fields. |
| `UPDATER_NO_PLATFORM` | `updater.check` | The manifest does not include an entry for the current platform/architecture. |
| `UPDATER_NETWORK_ERROR` | `updater.check`, `updater.install` | Failed to fetch the manifest or download the archive. |

---

### General Errors

| Code | Method(s) | Meaning |
|------|-----------|---------|
| `INVALID_ARGUMENT` | Any | A required parameter is missing, has the wrong type, or has an invalid value. |
| `INTERNAL_ERROR` | Any | An unexpected error occurred in the Go backend. This is a bug -- please report it. |

---

## Detailed Error Reference

### Permission Errors

#### fs.readFile / fs.writeFile / fs.readDir -- Permission denied

The app attempted to access a file path that is not in the allowed list.

**Code:** `FS_PERMISSION_DENIED`

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

#### fs.readFile / fs.writeFile -- Path traversal blocked

The app attempted to access a path that resolves outside the allowed directories after resolving symlinks and `..` segments.

**Code:** `FS_PATH_TRAVERSAL`

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

#### process.exec -- Command not allowed

The app attempted to run a command that is not in the allowed list.

**Code:** `PROCESS_NOT_ALLOWED`

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

#### http.fetch -- URL not allowed

The app attempted to make an HTTP request to a URL not in the allowed list.

**Code:** `HTTP_URL_NOT_ALLOWED`

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
      "allow": ["https://api.example.com/**"]
    }
  }
}
```

---

### File System Errors

#### fs.readFile -- File not found

**Code:** `FS_NOT_FOUND`

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

#### fs.writeFile -- Parent directory does not exist

**Code:** `FS_PARENT_NOT_FOUND`

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

#### fs.readFile -- Is a directory

**Code:** `FS_IS_DIRECTORY`

**Error:**
```
LightShell Error [fs.readFile]: Is a directory
  -> Path: /Users/alice/Documents
  -> Use lightshell.fs.readDir() to list directory contents
```

**Cause:** `readFile` was called on a path that is a directory, not a file.

**Solution:** Use `lightshell.fs.readDir()` to list directory contents, or check the path:
```js
const info = await lightshell.fs.stat(path)
if (info.isDir) {
  const entries = await lightshell.fs.readDir(path)
} else {
  const content = await lightshell.fs.readFile(path)
}
```

---

### Network Errors

#### http.fetch -- Network error

**Code:** `HTTP_NETWORK_ERROR`

**Error:**
```
LightShell Error [http.fetch]: Network error
  -> URL: https://api.example.com/data
  -> Cause: connection refused
  -> Check that the server is running and accessible
```

**Cause:** The HTTP request failed at the network level -- DNS resolution failure, connection refused, connection timeout, or TLS error.

**Solution:** Verify the URL is correct, the server is running, and your network connection is active. Check for firewall rules that might block outgoing connections.

---

#### http.fetch -- Request timeout

**Code:** `HTTP_TIMEOUT`

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

#### http.fetch -- Invalid URL

**Code:** `HTTP_INVALID_URL`

**Error:**
```
LightShell Error [http.fetch]: Invalid URL
  -> URL: not-a-url
  -> Provide a valid HTTP or HTTPS URL
```

**Cause:** The provided URL string is not a valid URL or does not use the HTTP/HTTPS protocol.

**Solution:** Ensure the URL starts with `http://` or `https://` and is properly formatted.

---

### Process Errors

#### process.exec -- Command not found

**Code:** `PROCESS_NOT_FOUND`

**Error:**
```
LightShell Error [process.exec]: Command not found
  -> Command: ffmpeg
  -> The command was not found in the system PATH
  -> Install the command or provide an absolute path
```

**Cause:** The command binary does not exist in the standard system PATH directories.

**Solution:** Install the required command, or use an absolute path to the binary:
```js
const result = await lightshell.process.exec('/usr/local/bin/ffmpeg', ['-version'])
```

---

#### process.exec -- Timeout

**Code:** `PROCESS_TIMEOUT`

**Error:**
```
LightShell Error [process.exec]: Command timed out
  -> Command: python3 long-script.py
  -> Timeout: 10000ms
  -> The process was killed after exceeding the timeout
```

**Cause:** The command did not complete within the specified timeout.

**Solution:** Increase the timeout or investigate why the command is slow:
```js
const result = await lightshell.process.exec('python3', ['long-script.py'], {
  timeout: 60000
})
```

---

### Updater Errors

#### updater.install -- SHA256 verification failed

**Code:** `UPDATER_SHA256_MISMATCH`

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

#### updater.check -- HTTPS required

**Code:** `UPDATER_HTTPS_REQUIRED`

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

#### updater.install -- Signature invalid

**Code:** `UPDATER_SIGNATURE_INVALID`

**Error:**
```
LightShell Error [updater.install]: Signature verification failed
  -> The Ed25519 signature does not match the archive contents
  -> The update has been rejected
  -> Ensure the release was signed with the correct private key
```

**Cause:** The Ed25519 signature in the manifest does not verify against the downloaded archive and the embedded public key. This may indicate a key mismatch, a tampered archive, or a corrupted signature.

**Solution:** Verify you are signing releases with the private key that matches the `updater.publicKey` in `lightshell.json`. See [Signing Keys](/docs/guides/auto-updates/signing-keys/).

---

### General Errors

#### Invalid argument

**Code:** `INVALID_ARGUMENT`

**Error:**
```
LightShell Error [window.setSize]: Invalid argument
  -> Parameter "width" must be a positive number, got: -100
```

**Cause:** A required parameter is missing, has the wrong type, or has an invalid value.

**Solution:** Check the API documentation for the correct parameter types and values.

---

#### Platform unsupported

**Code:** `PLATFORM_UNSUPPORTED`

**Error:**
```
LightShell Error [window.setVibrancy]: Platform unsupported
  -> setVibrancy is only available on macOS
  -> On Linux, this method is a no-op
```

**Cause:** The called method is not available on the current operating system.

**Solution:** Use `lightshell.system.platform()` to check the platform before calling platform-specific methods:
```js
const platform = await lightshell.system.platform()
if (platform === 'darwin') {
  await lightshell.window.setVibrancy('sidebar')
}
```

---

#### Internal error

**Code:** `INTERNAL_ERROR`

**Error:**
```
LightShell Error [fs.readFile]: Internal error
  -> An unexpected error occurred in the backend
  -> Please report this at https://github.com/lightshell-dev/lightshell/issues
```

**Cause:** An unexpected error in the Go backend that does not fit any known category. This is a bug.

**Solution:** Report the issue at [github.com/lightshell-dev/lightshell/issues](https://github.com/lightshell-dev/lightshell/issues) with the full error message and steps to reproduce.

---

## Common Patterns

### Error Handling by Code

Use the `code` property for reliable error handling instead of parsing the message string:

```js
try {
  await lightshell.fs.readFile(path)
} catch (err) {
  switch (err.code) {
    case 'FS_NOT_FOUND':
      console.log('File does not exist, creating default...')
      await lightshell.fs.writeFile(path, defaultContent)
      break
    case 'FS_PERMISSION_DENIED':
      await lightshell.dialog.message('Permission Error',
        'Cannot access this file. Check app permissions.')
      break
    case 'FS_IS_DIRECTORY':
      const entries = await lightshell.fs.readDir(path)
      showDirectoryListing(entries)
      break
    default:
      console.error('Unexpected error:', err.message)
  }
}
```

### Error Handling Wrapper

```js
async function safeCall(fn, fallback = null) {
  try {
    return await fn()
  } catch (err) {
    console.error(err.message)

    if (err.code?.startsWith('FS_PERMISSION') || err.code === 'PERMISSION_DENIED') {
      await lightshell.dialog.message(
        'Permission Error',
        'This action is not allowed by the app permissions. Check lightshell.json.'
      )
    } else if (err.code === 'FS_NOT_FOUND') {
      // Silently return fallback for missing files
    } else if (err.code?.startsWith('HTTP_')) {
      await lightshell.dialog.message('Network Error',
        'Could not connect to the server. Check your internet connection.')
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

### Categorizing Errors by Code Prefix

```js
function getErrorCategory(err) {
  if (!err.code) return 'unknown'
  if (err.code.startsWith('FS_')) return 'filesystem'
  if (err.code.startsWith('HTTP_')) return 'network'
  if (err.code.startsWith('PROCESS_')) return 'process'
  if (err.code.startsWith('UPDATER_')) return 'updater'
  if (err.code.startsWith('PERMISSION')) return 'permission'
  if (err.code === 'PLATFORM_UNSUPPORTED') return 'platform'
  if (err.code === 'INVALID_ARGUMENT') return 'validation'
  if (err.code === 'INTERNAL_ERROR') return 'internal'
  return 'unknown'
}
```
