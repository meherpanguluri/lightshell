---
title: AI-Native Design
description: How LightShell is designed for AI agents to build apps.
---

LightShell is designed so that an AI agent (Claude, GPT, Copilot, or any code-generating model) can produce a working desktop app from a single prompt. This is not a marketing claim -- it is a concrete design constraint that influenced every API decision. This page explains the principles and how they manifest in the framework.

## What "AI-Native" Means

An AI-native API is one that a language model can use correctly on the first try, without documentation lookup, without examples, and without debugging cycles. This requires:

1. **Predictable naming** -- if you know `lightshell.fs.readFile`, you can guess `lightshell.fs.writeFile`, `lightshell.fs.readDir`, `lightshell.fs.exists`.
2. **Flat structure** -- no method chaining, no builder patterns, no class hierarchies. Every API is a function call that takes arguments and returns a result.
3. **Consistent patterns** -- every API is async, every API returns a value or throws an error, every API uses the same IPC mechanism.
4. **Errors that explain the fix** -- not just "permission denied" but "permission denied, here is what you tried, here is what is allowed, here is the config to change."

## Design Principle 1: One Way to Do Things

In many frameworks, there are multiple ways to accomplish the same task. For reading a file, you might have streams, callbacks, promises, synchronous reads, and buffer-based reads. An AI model must choose between them, and it often chooses wrong or mixes patterns.

LightShell has one way:

```js
const content = await lightshell.fs.readFile('/tmp/data.txt')
```

No streams. No callbacks. No synchronous variant. No buffer mode. One function, one signature, one return type. An AI model that has seen the API name once will generate correct code.

The same principle applies across the framework:

```js
// One way to show a dialog
await lightshell.dialog.message('Title', 'Body')

// One way to read the clipboard
const text = await lightshell.clipboard.read()

// One way to persist data
await lightshell.store.set('key', value)

// One way to make an HTTP request
const res = await lightshell.http.fetch(url, { method: 'GET' })
```

## Design Principle 2: Errors That Explain the Fix

A typical framework error:

```
Error: EACCES: permission denied, open '/etc/passwd'
```

A LightShell error:

```
LightShell Error [fs.readFile]: Permission denied
  → Attempted to read: /etc/passwd
  → Allowed read paths: $APP_DATA/**, $HOME/Documents/**
  → To allow this path, update permissions.fs.read in lightshell.json
  → Docs: https://lightshell.dev/docs/api/permissions#fs
```

The LightShell error tells you:
- Which API was called
- What was attempted
- What is currently allowed
- How to fix it (which config key to change)
- Where to read more

When an AI agent encounters this error in a feedback loop, it can parse the error, modify `lightshell.json`, and retry -- without needing to search documentation or guess at solutions.

## Design Principle 3: Zero-Config Defaults

A new LightShell project works immediately:

```bash
lightshell init my-app
cd my-app
lightshell dev
```

No `package.json`. No `node_modules`. No webpack config. No vite config. No tsconfig. No babel. The app runs with a single HTML file and zero configuration.

The `lightshell.json` file has sensible defaults:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html"
}
```

Everything else is optional. Window size defaults to 800x600. Permissions default to permissive. CSP defaults to secure. The updater is disabled until you add the config. An AI agent generating a new app does not need to configure anything to get a working result.

When you do need to customize, every option has one place to set it:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": { "width": 1200, "height": 800, "title": "My App" },
  "permissions": { "fs": { "read": ["$APP_DATA/**"] } },
  "updater": { "enabled": true, "endpoint": "https://..." }
}
```

One file. Flat keys. No inheritance, no profiles, no environment overrides.

## Design Principle 4: Small API Surface

LightShell has 15 API namespaces covering the core needs of desktop apps:

| Namespace | Purpose |
|-----------|---------|
| `window` | Window management |
| `fs` | File system |
| `dialog` | Native dialogs |
| `clipboard` | Clipboard |
| `shell` | Open URLs/files |
| `notify` | System notifications |
| `tray` | System tray |
| `menu` | App menu |
| `system` | Platform info |
| `app` | App lifecycle |
| `store` | Key-value persistence |
| `http` | CORS-free HTTP client |
| `process` | Command execution |
| `shortcuts` | Global hotkeys |
| `updater` | Auto-updates |

This covers 95% of what desktop apps need. The total API surface is roughly 50 methods. An AI model can internalize the entire API from a single context document.

Compare this to Electron, which exposes hundreds of APIs across `BrowserWindow`, `ipcMain`, `ipcRenderer`, `app`, `dialog`, `shell`, `clipboard`, `screen`, `globalShortcut`, `Menu`, `MenuItem`, `Tray`, `Notification`, `nativeTheme`, `powerMonitor`, `session`, `webContents`, `webFrame`, and more. An AI model generating Electron code must choose the right module, the right method, and the right process (main vs renderer) for each operation. With LightShell, it is always `lightshell.{namespace}.{method}()`.

## Design Principle 5: No Build Tools Required

LightShell apps are plain HTML, CSS, and JavaScript. There is no compilation step, no bundling step, no transpilation step.

```html
<!DOCTYPE html>
<html>
<head>
  <title>My App</title>
  <link rel="stylesheet" href="styles.css">
</head>
<body>
  <h1>Hello</h1>
  <script src="app.js"></script>
</body>
</html>
```

This is a complete LightShell app. No JSX. No imports map. No module bundler. An AI model generating this code does not need to know about build toolchains.

If you want to use React, Vue, Svelte, or TypeScript, you can -- include them via CDN or set up your own build pipeline. But LightShell does not require it.

## llms.txt and llms-full.txt

LightShell ships with machine-readable documentation for AI context injection:

- **`llms.txt`** -- a concise summary of every API with signatures and one-line descriptions. Small enough to fit in a system prompt alongside other context.
- **`llms-full.txt`** -- the complete API documentation with examples, error descriptions, and configuration options. Use this when the AI has a large context window.

These files are designed to be pasted into an AI prompt or system message:

```
You are building a desktop app using LightShell.
Here is the API reference:

{contents of llms.txt}

Build me a markdown note-taking app with file save/open.
```

The AI reads the API reference, understands the available methods, and generates a working app. No guessing, no hallucinating APIs that do not exist.

## Example: One-Prompt App

Here is a prompt and the resulting working app:

**Prompt:**
> Build a LightShell app that's a simple todo list. Store todos persistently. Let me add, complete, and delete todos.

**Result (generated by AI):**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Todos</title>
  <style>
    body { font-family: system-ui; max-width: 500px; margin: 40px auto; padding: 0 20px; }
    input { width: 100%; padding: 8px; font-size: 16px; box-sizing: border-box; }
    ul { list-style: none; padding: 0; }
    li { display: flex; align-items: center; padding: 8px 0; border-bottom: 1px solid #eee; }
    li.done span { text-decoration: line-through; color: #999; }
    li span { flex: 1; cursor: pointer; }
    li button { background: none; border: none; color: #c00; cursor: pointer; font-size: 18px; }
  </style>
</head>
<body>
  <h1>Todos</h1>
  <input id="input" placeholder="What needs to be done?" autofocus>
  <ul id="list"></ul>
  <script>
    let todos = []

    async function load() {
      todos = (await lightshell.store.get('todos')) || []
      render()
    }

    async function save() {
      await lightshell.store.set('todos', todos)
    }

    function render() {
      const list = document.getElementById('list')
      list.innerHTML = ''
      todos.forEach((todo, i) => {
        const li = document.createElement('li')
        if (todo.done) li.className = 'done'
        li.innerHTML = `<span>${todo.text}</span><button>&times;</button>`
        li.querySelector('span').onclick = () => toggle(i)
        li.querySelector('button').onclick = () => remove(i)
        list.appendChild(li)
      })
    }

    function toggle(i) { todos[i].done = !todos[i].done; save(); render() }
    function remove(i) { todos.splice(i, 1); save(); render() }

    document.getElementById('input').addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && e.target.value.trim()) {
        todos.push({ text: e.target.value.trim(), done: false })
        e.target.value = ''
        save()
        render()
      }
    })

    load()
  </script>
</body>
</html>
```

This app works on the first try. The AI used `lightshell.store.get` and `lightshell.store.set` for persistence because the API is simple enough to use correctly from the name alone. No file paths, no serialization logic, no database setup. The AI did not need to decide between localStorage (not persistent across builds), IndexedDB (complex API), or file-based storage (path management). There is one obvious choice: `lightshell.store`.

## Why This Matters

Desktop app development has historically been complex. Electron requires understanding Node.js, the main process / renderer process split, IPC, and a large API surface. Tauri requires Rust knowledge for the backend. Native frameworks require platform-specific languages.

LightShell reduces the problem to: write HTML/CSS/JS, call `lightshell.*` when you need native features. This is a problem that AI models solve well. The result is that anyone -- regardless of programming experience -- can describe a desktop app and get a working binary.
