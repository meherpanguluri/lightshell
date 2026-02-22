---
title: Getting Started
description: Install LightShell and create your first desktop app in under 5 minutes.
---

LightShell lets you build native desktop apps using only JavaScript, HTML, and CSS. No Rust, no Go knowledge needed. This guide gets you from zero to a running desktop app in under 5 minutes.

## Prerequisites

- **macOS** (arm64 or x64) or **Linux** (x64 or arm64 with WebKitGTK 2.40+)
- **Node.js 18+** and npm

You do **not** need Go installed. The pre-compiled LightShell binary is downloaded automatically via npm.

## Install

```bash
npm install -g @lightshell/cli
```

This installs the `lightshell` CLI globally. The correct platform-specific binary is selected automatically.

Or create a project directly without a global install:

```bash
npx @lightshell/create my-app
```

## Create a New Project

```bash
lightshell init my-app
cd my-app
```

This creates a new directory with the following structure:

```
my-app/
  lightshell.json       # App configuration
  src/
    index.html          # Entry point
    app.js              # Application logic
    style.css           # Styles
```

The `lightshell.json` file defines your app's name, window size, and other settings. The `src/` directory contains your web code.

## Run in Development Mode

```bash
lightshell dev
```

This launches your app in a native window with:

- **Hot reload** — edit any file in `src/` and the app refreshes instantly
- **DevTools** — enabled by default in dev mode for debugging
- A local HTTP server serving your files

You should see a native window open with your app running inside it.

## Build for Distribution

```bash
lightshell build
```

This compiles your app into a native binary:

- **macOS**: produces a `.app` bundle in `dist/`
- **Linux**: produces an AppImage in `dist/`

The build output includes the final binary size. A typical app comes in at **~2.8MB**.

```
✓ Built my-app in 1.2s → 2.8MB
✓ Output: dist/MyApp.app
```

## What Just Happened?

When you ran `lightshell build`, here is what happened behind the scenes:

1. Your HTML, CSS, and JS files were embedded into a Go binary using `embed.FS`
2. The binary includes a thin runtime that opens a native window with the system webview (WKWebView on macOS, WebKitGTK on Linux)
3. Your code runs inside the webview with access to native APIs via `window.lightshell.*`
4. The result is a single native executable with no external dependencies

You never wrote Go. You never configured a build system. You wrote JS and got a native app.

## AI-Assisted Development with MCP

LightShell includes a built-in MCP server that lets AI agents see, inspect, and control your running app. If you use Claude Code, Cursor, or any MCP-compatible tool:

**1. Add MCP config** to your project root (`.mcp.json`):

```json
{
  "mcpServers": {
    "lightshell": {
      "command": "lightshell",
      "args": ["mcp"]
    }
  }
}
```

**2. Your AI agent can now:**

- Create and edit project files
- Start/stop the dev server
- Take screenshots of your running app
- Read console logs and errors
- Inspect the DOM
- Execute JS in the webview
- Hot-reload after changes
- Build for production

This enables a write-screenshot-fix loop where the AI iterates visually on your app.

## Next Steps

- [Build your first real app](/docs/tutorial/01-your-first-app/) — a step-by-step tutorial
- [Use native APIs](/docs/tutorial/02-native-apis/) — file system, dialogs, clipboard, and more
- [API Reference](/docs/api/window/) — complete documentation for every API
- [MCP Server Reference](/docs/api/cli/#lightshell-mcp) — all 16 MCP tools for AI agents
