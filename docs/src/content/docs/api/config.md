---
title: Configuration
description: Complete reference for lightshell.json — app configuration file.
---

Every LightShell app has a `lightshell.json` file in the project root. This file defines the app's metadata, window behavior, permissions, security policy, build options, and updater settings.

## Full Example

```json
{
  "name": "My App",
  "version": "1.0.0",
  "entry": "index.html",
  "window": {
    "title": "My App",
    "width": 1024,
    "height": 768,
    "minWidth": 400,
    "minHeight": 300,
    "resizable": true,
    "frameless": false
  },
  "tray": {
    "icon": "tray-icon.png",
    "tooltip": "My App"
  },
  "build": {
    "icon": "icon.png",
    "appId": "com.example.myapp"
  },
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
      "allow": ["https://api.example.com/**"],
      "deny": []
    }
  },
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.example.com"
  },
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.myapp.com/latest.json",
    "interval": "24h"
  }
}
```

## Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | — | Application display name |
| `version` | string | yes | — | Application version (semver recommended, e.g., `"1.0.0"`) |
| `entry` | string | no | `"index.html"` | Path to the main HTML file, relative to the project root |

---

### window

Controls the initial window appearance and behavior.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `title` | string | value of `name` | Initial window title bar text |
| `width` | number | `1024` | Initial window width in pixels |
| `height` | number | `768` | Initial window height in pixels |
| `minWidth` | number | `0` | Minimum window width (0 = no minimum) |
| `minHeight` | number | `0` | Minimum window height (0 = no minimum) |
| `resizable` | boolean | `true` | Whether the user can resize the window |
| `frameless` | boolean | `false` | Remove the native title bar and window chrome |

When `frameless` is `true`, you must implement your own title bar in HTML/CSS. Add `-webkit-app-region: drag` to your custom title bar element to make it draggable.

---

### tray

Optional. If present, a system tray icon is created on app startup.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `icon` | string | — | Path to the tray icon PNG, relative to the project root |
| `tooltip` | string | value of `name` | Tooltip text shown on hover |

The tray menu is set programmatically via `lightshell.tray.set()`. The config only sets the initial icon and tooltip.

---

### build

Build and packaging configuration.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `icon` | string | — | Path to the app icon PNG (512x512 recommended), relative to project root |
| `appId` | string | — | Reverse-domain application identifier (e.g., `"com.example.myapp"`) |
| `mac.identity` | string | — | macOS code signing identity (e.g., `"Developer ID Application: Name (TEAMID)"`) |
| `mac.entitlements` | object | — | macOS entitlements as key-value pairs |

The `appId` determines the app data directory path and the macOS bundle identifier. It should be unique to your application.

**macOS code signing example:**
```json
{
  "build": {
    "icon": "icon.png",
    "appId": "com.example.myapp",
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)",
      "entitlements": {
        "com.apple.security.network.client": true,
        "com.apple.security.files.user-selected.read-write": true
      }
    }
  }
}
```

---

### permissions

Optional. When present, the app runs in **restricted mode** and only the specified operations are allowed. When the entire `permissions` key is absent, the app runs in **permissive mode** where all operations are allowed.

#### permissions.fs

File system access controls using glob patterns.

| Field | Type | Description |
|-------|------|-------------|
| `read` | string[] | Glob patterns for allowed read paths |
| `write` | string[] | Glob patterns for allowed write paths |

**Path variables:**

| Variable | macOS | Linux |
|----------|-------|-------|
| `$APP_DATA` | `~/Library/Application Support/{appId}` | `~/.config/{appId}` |
| `$HOME` | `/Users/{user}` | `/home/{user}` |
| `$TEMP` | `/tmp` | `/tmp` |
| `$RESOURCE` | `{app-bundle}/Contents/Resources` | `{appimage-mount}/resources` |
| `$DOWNLOADS` | `~/Downloads` | `~/Downloads` |
| `$DESKTOP` | `~/Desktop` | `~/Desktop` |

**Glob patterns:** `*` matches within a single directory, `**` matches recursively across directories.

```json
{
  "permissions": {
    "fs": {
      "read": ["$APP_DATA/**", "$HOME/Documents/**", "$TEMP/**"],
      "write": ["$APP_DATA/**", "$TEMP/**"]
    }
  }
}
```

#### permissions.process

Controls which system commands can be executed via `lightshell.process.exec()`.

| Field | Type | Description |
|-------|------|-------------|
| `exec` | array | Array of allowed command definitions |

Each entry in the `exec` array:
- `cmd` (string) — the command name
- `args` (string[], optional) — allowed first arguments. If omitted or `["*"]`, any arguments are allowed.

```json
{
  "permissions": {
    "process": {
      "exec": [
        { "cmd": "git", "args": ["status", "log", "diff"] },
        { "cmd": "python3", "args": ["*"] },
        { "cmd": "ls" }
      ]
    }
  }
}
```

#### permissions.http

Controls which URLs can be accessed via `lightshell.http.fetch()` and `lightshell.http.download()`.

| Field | Type | Description |
|-------|------|-------------|
| `allow` | string[] | URL patterns that are permitted |
| `deny` | string[] | URL patterns that are blocked (takes precedence over allow) |

```json
{
  "permissions": {
    "http": {
      "allow": ["https://api.example.com/**", "https://cdn.example.com/**"],
      "deny": ["https://api.example.com/admin/**"]
    }
  }
}
```

---

### security

Security hardening options.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `csp` | string | *(see below)* | Custom Content Security Policy for the webview |

**Default CSP (production builds):**
```
default-src 'self' lightshell:; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'none'
```

**Default CSP (dev mode):**
```
default-src 'self' 'unsafe-inline' 'unsafe-eval' lightshell: http://localhost:*
```

Setting a custom `csp` completely replaces the default. Make sure to include `'self'` and any sources your app needs.

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'self' 'unsafe-inline'"
  }
}
```

---

### updater

Auto-update configuration. See the [Updater API](/docs/api/updater/) for the full JavaScript API.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable the auto-updater |
| `endpoint` | string | — | URL to the JSON update manifest |
| `interval` | string | `"24h"` | How often to check for updates in the background (e.g., `"1h"`, `"12h"`, `"24h"`) |

The endpoint must be HTTPS in production builds. HTTP is allowed in dev mode for local testing.

```json
{
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.myapp.com/latest.json",
    "interval": "12h"
  }
}
```

---

## Minimal Configuration

The smallest valid `lightshell.json`:

```json
{
  "name": "My App",
  "version": "1.0.0"
}
```

This uses all defaults: `index.html` as the entry point, a 1024x768 resizable window, permissive mode (no permission restrictions), default CSP, and no updater.
