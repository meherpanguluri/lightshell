---
title: Security & Permissions
description: Understand LightShell's permission system, CSP, and path validation.
---

LightShell provides a layered security model: a permission system for API access, Content Security Policy for the webview, path traversal protection for file operations, and secure IPC via Unix domain sockets. This guide explains how each layer works and how to configure them.

## Permission Modes

LightShell has two permission modes, set by the presence or absence of a `permissions` key in `lightshell.json`.

### Permissive Mode (Default)

If your `lightshell.json` has no `permissions` key, everything is allowed. This is the default for development and simple apps that do not need sandboxing.

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html"
}
```

In permissive mode, all `lightshell.*` APIs work without restriction. File system access, process execution, HTTP requests -- all allowed.

### Restricted Mode

Add a `permissions` key to lock down your app. Only explicitly allowed actions are permitted. Everything else is denied.

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/**"],
      "write": ["$APP_DATA/**"]
    },
    "process": {
      "exec": [
        { "cmd": "git", "args": ["status", "log", "diff"] },
        { "cmd": "python3", "args": ["*"] }
      ]
    },
    "http": {
      "allow": ["https://api.github.com/**", "https://api.example.com/**"],
      "deny": ["http://**"]
    }
  }
}
```

## File System Permissions

Control which paths your app can read from and write to.

```json
{
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/**", "$HOME/Pictures/**"],
      "write": ["$APP_DATA/**", "$TEMP/**"]
    }
  }
}
```

### Path Variables

Use these variables in your permission patterns. They resolve to platform-specific paths at runtime:

| Variable | macOS | Linux |
|----------|-------|-------|
| `$APP_DATA` | `~/Library/Application Support/{app-name}` | `~/.config/{app-name}` |
| `$HOME` | `/Users/{user}` | `/home/{user}` |
| `$TEMP` | `/tmp` | `/tmp` |
| `$RESOURCE` | `{app-bundle}/Contents/Resources` | `{appimage-mount}/resources` |
| `$DOWNLOADS` | `~/Downloads` | `~/Downloads` |
| `$DESKTOP` | `~/Desktop` | `~/Desktop` |

### Glob Patterns

`*` matches any characters within a single directory level. `**` matches recursively through any number of directory levels.

```json
{
  "fs": {
    "read": [
      "$HOME/Documents/**",
      "$HOME/Pictures/*.png",
      "$APP_DATA/**"
    ]
  }
}
```

- `$HOME/Documents/**` -- matches all files and subdirectories under Documents
- `$HOME/Pictures/*.png` -- matches only PNG files directly in Pictures (not subdirectories)
- `$APP_DATA/**` -- matches everything under the app's data directory

## Process Execution Permissions

Control which commands your app can run and with which arguments.

```json
{
  "permissions": {
    "process": {
      "exec": [
        { "cmd": "git", "args": ["status", "log", "diff", "add", "commit"] },
        { "cmd": "python3", "args": ["*"] },
        { "cmd": "ls" }
      ]
    }
  }
}
```

- `"args": ["status", "log"]` -- only those specific first arguments are allowed
- `"args": ["*"]` -- any arguments are allowed for that command
- No `args` key -- any arguments are allowed for that command

Commands are executed directly via `exec.Command`, never through a shell. This prevents shell injection attacks -- characters like `;`, `|`, `&&`, and backticks have no special meaning.

## HTTP Permissions

Control which URLs your app can reach.

```json
{
  "permissions": {
    "http": {
      "allow": [
        "https://api.github.com/**",
        "https://api.example.com/v1/**"
      ],
      "deny": [
        "http://**"
      ]
    }
  }
}
```

The `deny` list takes priority over `allow`. In the example above, all plain HTTP requests are blocked even if they match an allow pattern.

## Path Traversal Protection

Path traversal protection is always active, in both permissive and restricted modes. It cannot be disabled.

Every file system operation goes through path validation:

1. The requested path is resolved to an absolute path
2. All symlinks are resolved to their real target
3. The real path is checked against allowed patterns

This means attempts like `../../etc/passwd` or symlinks pointing outside allowed directories are always blocked:

```js
// This will fail even in permissive mode if the symlink
// resolves to a path outside the expected scope
await lightshell.fs.readFile('/tmp/my-app/data/../../../etc/passwd')
// Error: Permission denied
```

## Content Security Policy

LightShell injects a Content Security Policy into every HTML page to prevent XSS and code injection attacks.

### Production CSP

In production builds (`lightshell build`), a strict CSP is injected:

```
default-src 'self' lightshell:;
script-src 'self';
style-src 'self' 'unsafe-inline';
img-src 'self' data: blob:;
font-src 'self' data:;
connect-src 'self';
object-src 'none';
frame-ancestors 'none'
```

This means:
- Scripts can only load from your app bundle (no inline scripts, no eval)
- Styles can be inline (commonly needed for dynamic styling)
- Images can load from the app bundle, data URIs, and blob URIs
- No plugins or embedded objects
- No iframing of your app

### Development CSP

In dev mode (`lightshell dev`), the CSP is relaxed to allow inline scripts, eval, and localhost connections:

```
default-src 'self' 'unsafe-inline' 'unsafe-eval' lightshell: http://localhost:*
```

### Custom CSP

Override the default CSP in `lightshell.json` if your app needs to load external resources:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'self' 'unsafe-inline'; connect-src 'self' https://api.example.com"
  }
}
```

## IPC Security

LightShell uses Unix domain sockets for IPC between the webview and the Go backend. This is more secure than the localhost WebSocket approach used by some frameworks.

- **Socket permissions:** The socket file is created with `0600` permissions (owner-only read/write). No other user or process on the system can connect.
- **Socket path:** Includes the process PID and a random token: `/tmp/lightshell-{pid}-{random}.sock`. This prevents other processes from guessing the socket path.
- **Cleanup:** The socket file is deleted on shutdown via deferred cleanup and signal handlers.

## Error Messages

When a permission check fails, LightShell returns a detailed, structured error message designed to be useful for both humans and AI:

```
LightShell Error [fs.readFile]: Permission denied
  -> Attempted to read: /etc/passwd
  -> Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  -> To allow this path, update permissions.fs.read in lightshell.json
  -> Docs: https://lightshell.dev/docs/api/permissions#fs
```

These messages include:
- The exact API method that was denied
- The specific path or resource that was requested
- The current allowed patterns
- A concrete fix (which config key to update)
- A link to the relevant documentation

This format makes it straightforward for an AI code assistant to diagnose and fix permission issues automatically.

## Complete Restricted Example

Here is a full `lightshell.json` for a note-taking app with locked-down permissions:

```json
{
  "name": "secure-notes",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "Secure Notes",
    "width": 800,
    "height": 600
  },
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/Notes/**"],
      "write": ["$APP_DATA/**", "$HOME/Documents/Notes/**"]
    },
    "process": {
      "exec": []
    },
    "http": {
      "allow": [],
      "deny": ["**"]
    }
  },
  "security": {
    "csp": "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'"
  }
}
```

This configuration:
- Allows file access only to the app's data directory and a specific Notes folder
- Blocks all process execution (empty array means nothing is allowed)
- Blocks all HTTP requests
- Uses a strict CSP with no external resource loading

## Best Practices

**Start permissive, then restrict.** Develop your app in permissive mode to move fast. Before distributing, add a `permissions` key and whitelist only what your app actually needs.

**Use `$APP_DATA` for app files.** Store configuration, caches, and user data under `$APP_DATA`. This path is scoped to your app and works across platforms.

**Scope process execution tightly.** If your app runs `git`, only allow the specific git subcommands it uses. Do not use `"args": ["*"]` unless truly needed.

**Block plain HTTP in production.** Add `"deny": ["http://**"]` to your HTTP permissions to prevent accidental unencrypted requests.

**Test restricted mode before release.** Run your app with the restricted config and exercise every feature. Permission errors will surface any missed paths or commands.
