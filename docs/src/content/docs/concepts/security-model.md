---
title: Security Model
description: How LightShell protects your app and your users.
---

LightShell provides multiple layers of security between your JavaScript code and the operating system. This page explains the permission engine, content security policy, path validation, and IPC hardening that protect your app and your users.

## Three Security Layers

```
┌──────────────────────────────────────────┐
│  Your JavaScript Code                    │
├──────────────────────────────────────────┤
│  Layer 1: Content Security Policy (CSP)  │
│  → Controls what the webview can load    │
├──────────────────────────────────────────┤
│  Layer 2: Permission Engine              │
│  → Controls which APIs your app can call │
├──────────────────────────────────────────┤
│  Layer 3: Path Validation                │
│  → Controls which files your app can     │
│    read and write                        │
├──────────────────────────────────────────┤
│  Go Runtime (native operations)          │
└──────────────────────────────────────────┘
```

Each layer is independent. A request must pass all three to succeed.

## Permission Engine

The permission engine controls which LightShell APIs your app can use. It operates in one of two modes.

### Permissive Mode (Default)

If your `lightshell.json` has no `permissions` key, the app runs in **permissive mode**. All APIs are allowed. This is the default for new projects and is appropriate for apps you build and distribute yourself.

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html"
}
```

In permissive mode, `lightshell.fs.readFile('/etc/hosts')` works, `lightshell.process.exec('git', ['status'])` works, and every other API call succeeds (assuming the underlying operation succeeds).

### Restricted Mode

If your `lightshell.json` includes a `permissions` key, the app runs in **restricted mode**. Only the APIs and paths you explicitly allow are permitted.

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
        { "cmd": "git", "args": ["status", "log", "diff"] }
      ]
    },
    "http": {
      "allow": ["https://api.example.com/**"]
    }
  }
}
```

In restricted mode, any API call that does not match the permission rules is rejected with a detailed error.

### Permission Check Flow

Every IPC call goes through the permission engine before the handler executes:

```
JS calls lightshell.fs.readFile('/etc/passwd')
  → IPC router receives request
  → Permission engine checks:
    - Mode is "restricted"
    - Method is "fs.readFile"
    - Path is "/etc/passwd"
    - Allowed read paths: $APP_DATA/**, $HOME/Documents/**
    - /etc/passwd does not match any allowed pattern
  → Permission DENIED
  → Error returned to JavaScript
```

The check adds less than 0.1ms to each IPC call.

### Error Messages

When a permission is denied, the error message tells you exactly what happened and how to fix it:

```
LightShell Error [fs.readFile]: Permission denied
  → Attempted to read: /etc/passwd
  → Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  → To allow this path, update permissions.fs.read in lightshell.json
  → Docs: https://lightshell.dev/docs/api/permissions#fs
```

These error messages are designed to be understood by both humans and AI code generators. An AI agent reading this error can update the `lightshell.json` to fix the permission issue without external documentation.

### Path Variables

Permission patterns support these variables, resolved at runtime:

| Variable | macOS | Linux |
|----------|-------|-------|
| `$APP_DATA` | `~/Library/Application Support/{app-name}` | `~/.config/{app-name}` |
| `$HOME` | `/Users/{user}` | `/home/{user}` |
| `$TEMP` | `/tmp` | `/tmp` |
| `$RESOURCE` | `{app-bundle}/Contents/Resources` | `{appimage-mount}/resources` |
| `$DOWNLOADS` | `~/Downloads` | `~/Downloads` |
| `$DESKTOP` | `~/Desktop` | `~/Desktop` |

Glob patterns work in permission paths: `*` matches within a directory, `**` matches recursively.

## Content Security Policy

LightShell injects a Content Security Policy into every HTML page before any user scripts execute. The CSP controls what the webview is allowed to load: scripts, styles, images, network connections, and more.

### Production CSP

In production builds (`lightshell build`), a strict CSP is applied:

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

This prevents:
- Loading scripts from external origins (XSS mitigation)
- Embedding the app in an iframe (clickjacking mitigation)
- Loading plugins or objects

### Dev Mode CSP

In dev mode (`lightshell dev`), the CSP is relaxed to allow common development patterns:

```
default-src 'self' 'unsafe-inline' 'unsafe-eval' lightshell: http://localhost:*
```

This allows inline scripts, eval (used by some dev tools), and connections to localhost servers (for hot reload, API mocking, etc.).

### Custom CSP

Override the default CSP in `lightshell.json` if your app needs to load resources from external origins:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.example.com; connect-src 'self' https://api.example.com"
  }
}
```

Only override the CSP if you have a specific need. The default is secure for most apps.

## Path Traversal Protection

Path traversal protection is always on and cannot be disabled. It prevents an app (or a bug in an app) from accessing files outside the allowed directories.

Every file system operation goes through path validation:

1. **Resolve to absolute path** -- relative paths like `../../../etc/passwd` are resolved to their absolute equivalent.
2. **Resolve symlinks** -- if the path contains symlinks, they are followed to the real path. A symlink in `$APP_DATA` that points to `/etc/passwd` is caught here.
3. **Check against allowed patterns** -- in restricted mode, the real path must match at least one allowed pattern.

```
Requested path: $APP_DATA/../../etc/passwd
  → Absolute path: /etc/passwd
  → Real path: /etc/passwd
  → Allowed patterns: $APP_DATA/**
  → /etc/passwd does not match
  → DENIED
```

This validation runs even in permissive mode for the restricted set of always-protected paths (system files, other users' home directories).

## IPC Security

The communication channel between JavaScript and Go is secured at the transport level.

### Unix Domain Socket

LightShell uses a Unix domain socket for IPC instead of a localhost TCP connection. This matters because:

- **File permissions:** The socket file is created with `0600` permissions (owner-only read/write). No other user on the system can connect to it.
- **No network exposure:** A Unix domain socket is not accessible over the network, even on localhost. Other processes can only connect if they have filesystem access to the socket file.

Unlike a localhost TCP connection (where any process on the machine can connect), a Unix domain socket is only accessible to processes running as the same user, and the `0600` permissions further restrict access to the owning process.

### Socket Path Randomization

The socket path includes the process ID and a random token:

```
/tmp/lightshell-{pid}-{random}.sock
```

This prevents a malicious process from pre-creating a socket file at a known path (a symlink attack) and intercepting IPC messages.

### Cleanup

The socket file is deleted on shutdown via a deferred cleanup handler and a signal handler (for SIGINT and SIGTERM). If the app crashes without cleanup, the stale socket is detected and removed on the next launch.

## DevTools Control

DevTools (the web inspector) are controlled by the build mode:

| Mode | DevTools | Right-click menu |
|------|----------|-----------------|
| `lightshell dev` | Enabled | Enabled |
| `lightshell build` | Disabled | Disabled |
| `lightshell build --devtools` | Enabled | Enabled |

In production builds, DevTools are disabled by default. This prevents end users from inspecting the DOM, viewing network requests, or executing arbitrary JavaScript in the webview. Use the `--devtools` flag on `lightshell build` for debug builds that need inspection.

## Summary

LightShell's default is permissive because most apps built with LightShell are first-party (you build and distribute them). When you need tighter control — for apps that load third-party content or run plugins — switch to restricted mode by adding a `permissions` key to `lightshell.json`.
