---
title: IPC Protocol
description: How JavaScript and Go communicate in LightShell — message format, requests, responses, and events.
---

LightShell uses an IPC (Inter-Process Communication) protocol for communication between your JavaScript code and the Go runtime. This page documents the protocol for developers who want to understand how the `lightshell.*` APIs work internally.

## Overview

The IPC layer sits between the webview (where your JavaScript runs) and the Go backend (where native operations execute). Communication flows in both directions:

- **JS to Go**: API calls (e.g., read a file, show a dialog)
- **Go to JS**: Responses to API calls, and push events (e.g., window resized, file changed)

## Transport

Messages travel over a **Unix domain socket** (UDS) using a length-prefixed JSON protocol.

Each message is framed as:
```
[4-byte uint32 big-endian length][JSON payload]
```

The 4-byte length prefix tells the receiver how many bytes to read for the JSON payload. This avoids delimiter-based parsing issues with newlines or special characters in data.

The Unix domain socket is created in the system temp directory with a unique name per app instance. It is cleaned up on shutdown.

## Message Types

There are three message types: **Request**, **Response**, and **Event**.

### Request (JS to Go)

When you call a `lightshell.*` API, the client library creates a request:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "fs.readFile",
  "params": {
    "path": "/tmp/test.txt",
    "encoding": "utf-8"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | A UUID v4 that uniquely identifies this request. Used to match responses. |
| `method` | string | The API method name, in `module.method` format. |
| `params` | object | Method parameters. Structure varies by method. |

### Response (Go to JS)

The Go runtime processes the request and sends back a response with the same `id`:

**Success:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "result": "file contents here",
  "error": null
}
```

**Error:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "result": null,
  "error": "open /tmp/test.txt: no such file or directory"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Matches the request `id`. |
| `result` | any | The return value on success. Type depends on the method. `null` on error. |
| `error` | string or null | Error message on failure. `null` on success. |

### Event (Go to JS, no request ID)

Events are push notifications from Go to JavaScript. They have no `id` because they are not responses to requests.

```json
{
  "event": "window.resize",
  "data": {
    "width": 1024,
    "height": 768
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `event` | string | Event name in `module.eventName` format. |
| `data` | object | Event payload. Structure varies by event type. |

## Message Flow

Here is the complete flow for a `lightshell.fs.readFile()` call:

```
Your JS                  lightshell.js              Go Runtime
  │                          │                          │
  │ readFile('/tmp/f.txt')   │                          │
  ├─────────────────────────>│                          │
  │                          │ {id, method, params}     │
  │                          ├─────────────────────────>│
  │                          │                          │ os.ReadFile()
  │                          │                          │
  │                          │ {id, result, error}      │
  │                          │<─────────────────────────┤
  │ Promise resolves         │                          │
  │<─────────────────────────┤                          │
```

1. Your code calls `lightshell.fs.readFile('/tmp/f.txt')`
2. The client library generates a UUID, stores the Promise callbacks in a pending map, and sends the request via `window.webkit.messageHandlers.lightshell.postMessage()`
3. Go receives the raw message string through the webview's message handler
4. The IPC router parses the JSON, extracts the `method`, and dispatches to the registered handler
5. The handler executes `os.ReadFile("/tmp/f.txt")`
6. The result (or error) is wrapped in a response JSON
7. Go calls `webview.Eval('__lightshell_receive(' + json + ')')` to execute JavaScript in the webview
8. The `__lightshell_receive` function looks up the pending Promise by `id` and resolves or rejects it

## Event Flow

Events work differently — they are initiated by Go, not by a JS request:

```
Go Runtime                 lightshell.js              Your JS
  │                          │                          │
  │ window resize detected   │                          │
  │                          │                          │
  │ {event, data}            │                          │
  ├─────────────────────────>│                          │
  │                          │ invoke all listeners     │
  │                          ├─────────────────────────>│
  │                          │                          │ callback(data)
```

To listen for events in your code:

```js
lightshell.window.onResize((data) => {
  console.log(data.width, data.height)
})
```

The `onResize` function registers a callback in the listeners map under the `"window.resize"` event key. When Go pushes a resize event, `__lightshell_receive` invokes all registered callbacks.

## Method Registry

The Go-side IPC router maps method names to handler functions:

| Method | Handler | Description |
|--------|---------|-------------|
| `window.setTitle` | WindowAPI | Set window title |
| `window.setSize` | WindowAPI | Resize window |
| `window.getSize` | WindowAPI | Get window dimensions |
| `window.setPosition` | WindowAPI | Move window |
| `window.getPosition` | WindowAPI | Get window position |
| `window.minimize` | WindowAPI | Minimize window |
| `window.maximize` | WindowAPI | Maximize window |
| `window.fullscreen` | WindowAPI | Enter fullscreen |
| `window.restore` | WindowAPI | Restore window |
| `window.close` | WindowAPI | Close window |
| `fs.readFile` | FSAPI | Read file contents |
| `fs.writeFile` | FSAPI | Write file contents |
| `fs.readDir` | FSAPI | List directory |
| `fs.exists` | FSAPI | Check path existence |
| `fs.stat` | FSAPI | Get file metadata |
| `fs.mkdir` | FSAPI | Create directory |
| `fs.remove` | FSAPI | Delete file/directory |
| `fs.watch` | FSAPI | Watch for changes |
| `dialog.open` | DialogAPI | File open dialog |
| `dialog.save` | DialogAPI | File save dialog |
| `dialog.message` | DialogAPI | Message dialog |
| `dialog.confirm` | DialogAPI | Confirmation dialog |
| `dialog.prompt` | DialogAPI | Text input dialog |
| `clipboard.read` | ClipboardAPI | Read clipboard |
| `clipboard.write` | ClipboardAPI | Write clipboard |
| `shell.open` | ShellAPI | Open URL/file |
| `notify.send` | NotifyAPI | System notification |
| `tray.set` | TrayAPI | Set tray icon/menu |
| `tray.remove` | TrayAPI | Remove tray icon |
| `menu.set` | MenuAPI | Set app menu |
| `system.platform` | SystemAPI | Get OS name |
| `system.arch` | SystemAPI | Get CPU arch |
| `system.homeDir` | SystemAPI | Get home directory |
| `system.tempDir` | SystemAPI | Get temp directory |
| `system.hostname` | SystemAPI | Get hostname |
| `app.quit` | AppAPI | Quit application |
| `app.version` | AppAPI | Get app version |
| `app.dataDir` | AppAPI | Get data directory |

## Event Types

| Event | Trigger | Data |
|-------|---------|------|
| `window.resize` | Window resized | `{ width, height }` |
| `window.move` | Window moved | `{ x, y }` |
| `window.focus` | Window focused | `{}` |
| `window.blur` | Window lost focus | `{}` |
| `fs.watch` | Watched file changed | `{ path, event }` |
| `tray.click` | Tray menu item clicked | `{ id }` |
| `menu.click` | App menu item clicked | `{ id }` |

## Concurrency

The IPC server handles multiple concurrent requests. Each request is processed independently — you can fire multiple API calls in parallel:

```js
// These run concurrently
const [size, position, platform] = await Promise.all([
  lightshell.window.getSize(),
  lightshell.window.getPosition(),
  lightshell.system.platform()
])
```

Responses are matched to requests by their `id`, so order does not matter.

## Performance

Typical latencies for IPC round-trips:

| Operation | Latency |
|-----------|---------|
| System info (platform, arch) | < 1ms |
| Window operations (getSize, setTitle) | < 2ms |
| File read (small file) | < 5ms |
| File dialog (user interaction) | varies |
| Clipboard read/write | < 2ms |

The Unix domain socket avoids TCP overhead. JSON serialization is fast for the small payloads involved. The bottleneck is almost always the native operation itself (e.g., disk I/O), not the IPC layer.
