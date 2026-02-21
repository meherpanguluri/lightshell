---
title: Cursor Rules
description: Configure Cursor IDE for LightShell development.
---

[Cursor](https://cursor.sh) is an AI-powered code editor that supports project-level rules. These rules are included as context in every AI conversation within the project, preventing the AI from making framework-specific mistakes before they happen.

## Installation

Create a file called `.cursorrules` in the root of your LightShell project:

```bash
cd my-lightshell-app
touch .cursorrules
```

Paste the complete rules file below into `.cursorrules`. Cursor automatically detects this file and includes its contents in every AI interaction within the project.

## The Complete .cursorrules File

```
You are building a LightShell desktop application.

LightShell is a desktop app framework where you write HTML, CSS, and JS and get a
native binary. It uses system webviews (WKWebView on macOS, WebKitGTK on Linux).
There is no Node.js runtime, no Electron, no npm, and no build step.

## Key Facts

- All native APIs are at window.lightshell.* — no imports or require() needed
- Every lightshell.* API call is async — ALWAYS use await
- Entry point is src/index.html, configuration is lightshell.json
- No Node.js APIs are available (no require, no process, no __dirname, no Buffer)
- No npm or node_modules — use CDN <script> tags for third-party libraries
- Target platforms: macOS and Linux only (no Windows)
- File paths must always be absolute, never relative

## Available API Namespaces

lightshell.window    — setTitle, setSize, getSize, setPosition, getPosition,
                       minimize, maximize, fullscreen, restore, close
lightshell.fs        — readFile, writeFile, readDir, exists, stat, mkdir, remove, watch
lightshell.dialog    — open, save, message, confirm, prompt
lightshell.clipboard — read, write
lightshell.shell     — open (URLs in browser, files in default app)
lightshell.notify    — send
lightshell.tray      — set, remove, onClick
lightshell.menu      — set
lightshell.system    — platform, arch, homeDir, tempDir, hostname
lightshell.app       — quit, version, dataDir
lightshell.store     — get, set, delete, has, keys, clear (persistent key-value)
lightshell.http      — fetch (CORS-free), download
lightshell.process   — exec (scoped shell command execution)
lightshell.shortcuts — register, unregister, unregisterAll, isRegistered
lightshell.updater   — check, install, checkAndInstall, onProgress

## Do NOT Use

- require('fs'), require('path'), or any Node.js module
- window.open() for external URLs — use lightshell.shell.open() instead
- window.showOpenFilePicker() — use lightshell.dialog.open() instead
- window.showSaveFilePicker() — use lightshell.dialog.save() instead
- localStorage/sessionStorage for persistence — use lightshell.store instead
- fetch() for cross-origin requests — use lightshell.http.fetch() instead
- process.env, __dirname, __filename — not available
- import/export module syntax for lightshell APIs — they are globals

## File Structure

A LightShell project looks like this:

lightshell.json      — app config (name, version, window size)
src/
  index.html         — entry point, loaded by the webview
  app.js             — application logic (optional, can inline in HTML)
  style.css          — styles (optional, can inline in HTML)

## CSS Guidelines

- Use explicit font stacks: -apple-system, BlinkMacSystemFont, "Segoe UI",
  "Noto Sans", Helvetica, Arial, sans-serif
- Do NOT rely on system-ui alone (inconsistent across Linux distributions)
- Avoid CSS nesting (not supported on WebKitGTK < 2.44)
- Avoid backdrop-filter without a fallback (limited on Linux)
- Use .platform-darwin and .platform-linux body classes for platform-specific CSS
- Avoid :has() selector (not supported on WebKitGTK < 2.42)
- Use flexbox or grid for layouts — both are well-supported

## Common Patterns

// Get app data directory (use this as base for file paths)
const dataDir = await lightshell.app.dataDir()

// Read and write files
const content = await lightshell.fs.readFile(filePath)
await lightshell.fs.writeFile(filePath, content)

// Open file picker (returns null if cancelled)
const path = await lightshell.dialog.open({ filters: [...] })
if (path) { /* user selected a file */ }

// Persistent storage (JSON-serialized automatically)
await lightshell.store.set('key', { any: 'value' })
const val = await lightshell.store.get('key')

// HTTP request (no CORS restrictions)
const res = await lightshell.http.fetch('https://api.example.com/data')
const data = JSON.parse(res.body)

// Open URL in system browser
await lightshell.shell.open('https://example.com')
```

## What Each Rule Does

### "All native APIs are at window.lightshell.*"

Prevents Cursor from generating `import { readFile } from 'lightshell'` or `const ls = require('lightshell')`. LightShell APIs are injected into the global scope automatically.

### "Every lightshell.* API call is async"

Prevents Cursor from writing `const name = lightshell.store.get('name')` without `await`. Every LightShell API returns a Promise.

### "No Node.js APIs are available"

This is the most important rule. Without it, Cursor frequently generates `const fs = require('fs')`, `process.env.HOME`, or `path.join()`. None of these exist in a LightShell app.

### "No npm or node_modules"

Prevents Cursor from suggesting `npm install marked` or generating `import` statements from npm packages. Instead, it will use CDN script tags like `<script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>`.

### "File paths must always be absolute"

LightShell's `fs` module does not resolve relative paths. This rule ensures Cursor generates code that builds paths from `lightshell.app.dataDir()` or `lightshell.system.homeDir()` rather than using `./data/file.txt`.

### "Use lightshell.shell.open() for external URLs"

In a webview, `window.open('https://example.com')` navigates the webview itself instead of opening the system browser. This rule teaches Cursor the correct pattern.

### CSS Guidelines

WebKitGTK (the Linux webview) lags behind Safari's WebKit in CSS feature support. The CSS rules prevent Cursor from generating modern CSS that breaks on Linux: no nesting, no `:has()`, explicit font stacks.

### Platform-Specific CSS Classes

LightShell adds `.platform-darwin` or `.platform-linux` to the `<body>` element automatically. This lets Cursor generate platform-specific styles:

```css
/* Default styles */
.sidebar { background: #f5f5f5; }

/* macOS-specific */
.platform-darwin .sidebar { background: rgba(245, 245, 245, 0.8); }

/* Linux-specific */
.platform-linux .sidebar { background: #f0f0f0; }
```

## Verifying It Works

After creating the `.cursorrules` file:

1. Open your LightShell project in Cursor
2. Open the AI chat (Cmd+L)
3. Ask: "What APIs are available in this project?"
4. Cursor should respond with the LightShell API namespaces, not Node.js or browser APIs

If Cursor still generates Node.js code, make sure the `.cursorrules` file is in the project root (the same directory as `lightshell.json`).

## Combining with llms-full.txt

For even better results, reference the full API documentation in your Cursor chat:

```
@https://lightshell.dev/llms-full.txt

Add a file export feature to this app.
```

The `.cursorrules` file provides the constraints (what not to do), while `llms-full.txt` provides the complete API details (what to do and how).
