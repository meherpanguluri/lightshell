---
title: Custom Go Handlers
description: Extend your LightShell app with custom Go code — register handlers callable from JavaScript and clean up resources on shutdown.
---

LightShell lets you write custom Go handlers that your JavaScript can call directly. This is useful when you need to run a sidecar process (like an AI model server), perform CPU-intensive computation, or integrate with a Go library.

You never need to touch the LightShell internals. You write your handlers in a single `handlers.go` file in your project root, and the build pipeline wires everything together.

## How It Works

```
Your JS                    handlers.go (Go)
  │                            │
  │ lightshell.invoke(name)    │
  ├───────────────────────────>│
  │                            │ your Go code runs
  │                            │
  │ Promise resolves           │
  │<───────────────────────────┤
```

1. You register named handlers in `handlers.go` using `Handle(name, fn)`
2. JavaScript calls `lightshell.invoke(name, payload)` which returns a Promise
3. The Go handler receives the payload as JSON, processes it, and returns a result
4. The Promise resolves with the returned value

## Quick Start

When you run `lightshell init`, a `handlers.go` file is created in your project root:

```go
//go:build ignore

package main

import "encoding/json"

func customHandlers() {
    // Register your custom handlers here.
    _ = json.RawMessage{} // keep import
}
```

Add your handlers inside `customHandlers()`:

```go
//go:build ignore

package main

import "encoding/json"

func customHandlers() {
    Handle("greet", func(payload json.RawMessage) (any, error) {
        var p struct {
            Name string `json:"name"`
        }
        json.Unmarshal(payload, &p)
        return map[string]any{"message": "Hello, " + p.Name + "!"}, nil
    })
}
```

Call it from JavaScript:

```js
const result = await lightshell.invoke('greet', { name: 'Alice' })
console.log(result.message) // "Hello, Alice!"
```

## API Reference

### JavaScript: `lightshell.invoke(handler, payload)`

Call a custom Go handler by name.

**Parameters:**
- `handler` (string) — the handler name registered with `Handle()`
- `payload` (object, optional) — JSON-serializable data passed to the handler

**Returns:** `Promise<any>` — the value returned by the Go handler

**Errors:**
- Rejects if the handler name is not registered
- Rejects if the handler returns an error

```js
try {
  const result = await lightshell.invoke('compute', { input: [1, 2, 3] })
  console.log(result)
} catch (err) {
  console.error('Handler error:', err.message)
}
```

### Go: `Handle(name, handler)`

Register a named handler callable from JavaScript.

```go
func Handle(name string, handler func(json.RawMessage) (any, error))
```

- `name` — the handler name (must match the first argument to `lightshell.invoke()`)
- `handler` — receives the payload as raw JSON, returns any JSON-serializable value

### Go: `OnShutdown(fn)`

Register a function to run when the app exits. Use this to clean up child processes, close connections, or save state.

```go
func OnShutdown(fn func())
```

Shutdown hooks run when:
- The user closes the window
- `lightshell.app.quit()` is called from JavaScript
- The process receives SIGINT or SIGTERM (e.g., Ctrl+C)

Multiple hooks can be registered. They run in the order they were added.

## Examples

### Run a Sidecar Process

Start a background process (like an AI model server) and kill it on shutdown:

```go
//go:build ignore

package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
)

var llamaCmd *exec.Cmd

func customHandlers() {
    // Start llama-server on app launch
    llamaCmd = exec.Command("llama-server", "--model", "model.gguf", "--port", "8081")
    llamaCmd.Start()
    fmt.Println("Started llama-server on :8081")

    // Kill it when the app exits
    OnShutdown(func() {
        if llamaCmd != nil && llamaCmd.Process != nil {
            llamaCmd.Process.Kill()
            fmt.Println("Killed llama-server")
        }
    })

    // Expose a health check
    Handle("ai.status", func(payload json.RawMessage) (any, error) {
        if llamaCmd == nil || llamaCmd.ProcessState != nil {
            return map[string]any{"running": false}, nil
        }
        return map[string]any{"running": true, "pid": llamaCmd.Process.Pid}, nil
    })
}
```

```js
const status = await lightshell.invoke('ai.status')
if (status.running) {
  console.log(`AI server running (PID ${status.pid})`)
}
```

### Database Operations

Use a Go library for database access:

```go
//go:build ignore

package main

import (
    "database/sql"
    "encoding/json"
    _ "modernc.org/sqlite"
)

var db *sql.DB

func customHandlers() {
    var err error
    db, err = sql.Open("sqlite", "app.db")
    if err != nil {
        panic(err)
    }
    db.Exec("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT, created TEXT)")

    OnShutdown(func() {
        db.Close()
    })

    Handle("notes.list", func(payload json.RawMessage) (any, error) {
        rows, err := db.Query("SELECT id, text, created FROM notes ORDER BY created DESC")
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        var notes []map[string]any
        for rows.Next() {
            var id int
            var text, created string
            rows.Scan(&id, &text, &created)
            notes = append(notes, map[string]any{"id": id, "text": text, "created": created})
        }
        return notes, nil
    })

    Handle("notes.add", func(payload json.RawMessage) (any, error) {
        var p struct {
            Text string `json:"text"`
        }
        json.Unmarshal(payload, &p)
        result, err := db.Exec("INSERT INTO notes (text, created) VALUES (?, datetime('now'))", p.Text)
        if err != nil {
            return nil, err
        }
        id, _ := result.LastInsertId()
        return map[string]any{"id": id}, nil
    })
}
```

```js
// Add a note
await lightshell.invoke('notes.add', { text: 'Remember to ship v2' })

// List all notes
const notes = await lightshell.invoke('notes.list')
notes.forEach(n => console.log(n.text))
```

## The `//go:build ignore` Tag

The `handlers.go` file has a `//go:build ignore` tag at the top. This prevents it from being compiled when you run `go build ./...` on the LightShell repository itself. During `lightshell build`, the build pipeline copies your `handlers.go` into a temporary staging directory and compiles it as part of the generated `main.go` — the build tag is stripped automatically.

You should keep this tag in your `handlers.go` file.

## How It Works Internally

When you run `lightshell build`:

1. The CLI reads your `handlers.go` from the project root
2. It generates a `main.go` that includes `Handle()` and `OnShutdown()` functions
3. Your `handlers.go` is copied alongside `main.go` in a temp staging directory
4. `customHandlers()` is called during app startup, before the window opens
5. Shutdown hooks are called when the app exits (via signal handler or normal close)

If no `handlers.go` exists, a default empty one is generated — your app works fine without custom handlers.

## Dev Mode

In `lightshell dev`, custom handlers registered in `handlers.go` are **not** active because dev mode uses the prebuilt CLI binary rather than compiling your project. Custom handlers only work in built apps (`lightshell build`).

To test custom handlers during development, use `lightshell build && ./dist/your-app`.

## Platform Notes

- `handlers.go` is compiled as part of the final binary — any Go package you import is available
- On macOS, the binary is placed inside the `.app` bundle at `Contents/MacOS/`
- On Linux, the binary is packaged into the AppImage
- Shutdown hooks run on SIGINT, SIGTERM, and normal window close
- Multiple shutdown hooks are supported; they run sequentially in registration order
