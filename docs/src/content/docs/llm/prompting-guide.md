---
title: Prompting Guide
description: How to prompt AI to build LightShell apps effectively.
---

This guide covers how to write prompts that produce working LightShell apps on the first attempt. The key is providing the right context, being specific about what you want, and mentioning which LightShell APIs to use.

## The Prompt Template

Use this structure as a starting point for any LightShell prompt:

```
Build a LightShell desktop app that [description of what the app does].

Use lightshell.* APIs for all native features:
- [list specific APIs the app needs]

The app should:
- [feature 1]
- [feature 2]
- [feature 3]

Put everything in src/index.html, src/app.js, and src/style.css.
Use vanilla JS only — no frameworks, no npm, no build tools.
All lightshell.* calls are async — always use await.
```

This template works because it tells the AI the framework (LightShell), the APIs (specific namespaces), the output structure (which files to create), and the constraints (no Node.js, no frameworks, async).

## The Three Rules

### 1. Include Context

The AI needs to know what LightShell is and what APIs are available. Without context, it may default to Node.js APIs or browser-only code.

**Best approach — paste the full reference:**

```
Here is the LightShell API reference:
https://lightshell.dev/llms-full.txt

Build a todo app that persists tasks across restarts.
```

**Good enough — mention key APIs explicitly:**

```
Build a LightShell desktop app. LightShell provides these APIs at window.lightshell.*:
- lightshell.store.set(key, value) / .get(key) — persistent key-value storage
- lightshell.dialog.confirm(title, message) — native confirmation dialog
- lightshell.window.setTitle(title) — set the window title

Build a todo app that saves tasks using lightshell.store.
```

**Insufficient — just saying "LightShell":**

```
Build a LightShell app that manages todos.
```

This usually fails because the AI does not know the API surface and will guess incorrectly.

### 2. Be Specific About Features

Vague prompts produce vague apps. List the exact features you want.

**Vague:**
```
Build a note-taking app with LightShell.
```

**Specific:**
```
Build a LightShell note-taking app with:
- A sidebar listing all saved notes by title
- A text editor area (plain text, not rich text)
- Save button that writes the note to disk using lightshell.fs.writeFile
- Open button using lightshell.dialog.open() to load an existing .txt file
- The last opened file path should be remembered using lightshell.store
- Window title should show the current file name via lightshell.window.setTitle
- Keyboard shortcut: Cmd/Ctrl+S to save
- Clean, minimal UI with a monospace font in the editor
```

### 3. Mention LightShell Explicitly

Always include "LightShell" in your prompt. This ensures the AI uses the correct APIs and patterns.

Additionally, include these constraints to prevent common mistakes:

```
Constraints:
- Use lightshell.* APIs, not Node.js (no require, no process, no __dirname)
- No npm or node_modules — use CDN script tags if you need a library
- All lightshell.* calls are async — always use await
- Use lightshell.shell.open() for external URLs, not window.open()
- Use lightshell.dialog.open() for file pickers, not showOpenFilePicker()
- File paths must be absolute — use lightshell.app.dataDir() or lightshell.system.homeDir() as base
```

## Tips for Better Results

### Ask for Single-File Apps for Prototypes

For quick prototypes, ask the AI to put everything in one file:

```
Build this as a single-file LightShell app — put all HTML, CSS, and JS in src/index.html.
```

This reduces the chance of the AI generating incorrect import paths or module structures.

### Specify Which APIs to Use

Do not leave API choice to the AI. If you want persistence, say `lightshell.store`. If you want file access, say `lightshell.fs`. This prevents the AI from inventing APIs that do not exist.

```
Use lightshell.store for persistence (not localStorage, not fs).
Use lightshell.http.fetch for API calls (not browser fetch — it has CORS issues).
Use lightshell.dialog.open() to pick files (not an <input type="file">).
```

### Mention the Platform

If you are targeting a specific platform, say so. This helps the AI make better CSS and UX decisions.

```
Target macOS primarily. Use -apple-system font stack.
The app should feel native on macOS — use a sidebar layout with vibrancy-style styling.
```

### Ask for Error Handling

AI-generated code often skips error handling. Ask for it explicitly:

```
Handle errors gracefully:
- Wrap lightshell.fs calls in try/catch
- Check for null from lightshell.dialog.open() (user may cancel)
- Show lightshell.dialog.message() on errors instead of console.log
```

### Request the lightshell.json Config

The AI sometimes forgets the configuration file. Ask for it:

```
Also generate the lightshell.json config file with the app name, version, and window size.
```

## What Works Well

These app types are straightforward for AI to generate with LightShell:

| App Type | Key APIs | Complexity |
|----------|----------|-----------|
| Todo list with persistence | `store`, `dialog.confirm` | Low |
| Note editor with file save/load | `fs`, `dialog.open`, `dialog.save` | Low |
| API dashboard | `http.fetch`, `store` | Low |
| Clipboard manager | `clipboard`, `shortcuts`, `store` | Medium |
| File explorer / browser | `fs.readDir`, `fs.stat`, `dialog`, `shell.open` | Medium |
| System info viewer | `system`, `process.exec` | Low |
| RSS/feed reader | `http.fetch`, `store` | Medium |
| Settings/preferences UI | `store`, `dialog`, `window` | Low |
| Markdown previewer | `fs`, `dialog` | Low (with CDN markdown lib) |

## What Needs Iteration

These patterns usually require a follow-up prompt to get right:

- **Complex multi-panel layouts** — the AI may produce CSS that does not resize well. Ask for flexbox-based layouts explicitly.
- **Real-time file watching** — `lightshell.fs.watch` exists but the AI may not use it correctly. Provide the method signature.
- **Global shortcuts** — the accelerator string format (`CommandOrControl+Shift+P`) needs to be exact. Provide examples.
- **Cross-platform styling** — if you need the app to look good on both macOS and Linux, ask the AI to use the `.platform-darwin` and `.platform-linux` CSS classes.

## Example Prompts

See the [Example Prompts](/llm/example-prompts/) page for 8 complete, copy-paste-ready prompts that produce working LightShell apps.

## When the AI Gets It Wrong

See [Common AI Mistakes](/llm/common-mistakes/) for a catalog of frequent errors (using `require('fs')`, forgetting `await`, using `window.open()`) and their fixes.
