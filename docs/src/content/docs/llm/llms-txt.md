---
title: llms.txt Specification
description: The llms.txt and llms-full.txt files for AI context injection.
---

LightShell publishes two machine-readable documentation files that give AI models the context they need to generate working apps. These follow the [llms.txt](https://llmstxt.org) convention — a standardized format for providing documentation context to large language models.

## The Two Files

### llms.txt — Summary with Links

**URL:** `https://lightshell.dev/llms.txt`

A compact summary of LightShell in approximately 30 lines. It describes what LightShell is, lists the available API namespaces, and links to full documentation for each one. This is enough context for an AI to understand the framework and ask for more detail when needed.

**Best for:**
- Chat conversations where you want quick context without filling the entire context window
- AI tools that can follow links to fetch additional documentation on demand
- Initial exploration — "what can LightShell do?"

**Example content:**

```
# LightShell

> LightShell is a desktop app framework for macOS and Linux. Write HTML, CSS, and JS — get a native binary.

## Docs

- [Getting Started](https://lightshell.dev/docs/getting-started): Install and create your first app
- [API: Window](https://lightshell.dev/docs/api/window): setTitle, setSize, getSize, minimize, maximize, fullscreen, close
- [API: File System](https://lightshell.dev/docs/api/fs): readFile, writeFile, readDir, exists, stat, mkdir, remove, watch
- [API: Dialog](https://lightshell.dev/docs/api/dialog): open, save, message, confirm, prompt
- [API: Store](https://lightshell.dev/docs/api/store): get, set, delete, has, keys, clear
- [API: HTTP](https://lightshell.dev/docs/api/http): fetch (CORS-free), download
...
```

### llms-full.txt — Complete API Reference

**URL:** `https://lightshell.dev/llms-full.txt`

The full API reference in approximately 800 lines. Includes every method signature, every parameter with its type, return types, and a usage example for each. This gives an AI everything it needs to generate complete, working code without guessing.

**Best for:**
- Comprehensive code generation — "build me a full app"
- AI tools that cannot follow links (no browsing capability)
- System prompts where you want the AI to have complete knowledge upfront

**Example content:**

```
# LightShell Full API Reference

## lightshell.fs

### readFile(path, encoding?)
Read a file as a string.
- path (string): absolute file path
- encoding (string, optional): defaults to "utf-8"
- Returns: Promise<string>
- Example: const content = await lightshell.fs.readFile('/path/to/file.txt')

### writeFile(path, content)
Write a string to a file. Creates the file if it does not exist. Overwrites if it does.
- path (string): absolute file path
- content (string): the text to write
- Returns: Promise<void>
- Example: await lightshell.fs.writeFile('/path/to/file.txt', 'hello')
...
```

## How to Use

### In a Chat Conversation

Paste the URL directly into your message to Claude, ChatGPT, or any other AI:

```
Here is the LightShell API reference: https://lightshell.dev/llms-full.txt

Build me a desktop note-taking app that saves notes to disk.
```

Some AI tools will fetch the URL automatically. For those that do not, copy-paste the file contents instead.

### In Cursor IDE

Use the `@` mention to include the file as context:

1. Open a file in your LightShell project
2. Open Cursor's AI chat (Cmd+L)
3. Type `@https://lightshell.dev/llms-full.txt` to include the full reference
4. Ask your question or describe what you want to build

Alternatively, add it to your project's `.cursorrules` file so every conversation automatically has context. See the [Cursor Rules](/llm/cursor-rules/) page.

### In a System Prompt

If you are building an AI tool or agent that generates LightShell apps, include the llms-full.txt content in the system prompt:

```
You are an AI that builds desktop apps using LightShell.

<lightshell-api>
{contents of llms-full.txt}
</lightshell-api>

When the user describes an app, generate the HTML, CSS, and JS files needed.
Always use lightshell.* APIs for native features. Never use Node.js APIs.
```

### In Claude Code (AGENTS.md)

Claude Code reads project-level instruction files automatically. You can reference the llms-full.txt content in your AGENTS.md file or simply rely on the [AGENTS.md template](/llm/agents-md/) which includes a summary of all APIs.

### In GitHub Copilot

Copilot does not support URL injection, but you can include the reference as a comment block at the top of your file:

```js
/*
 * LightShell API — all methods are async, available at window.lightshell.*
 * fs: readFile, writeFile, readDir, exists, stat, mkdir, remove, watch
 * dialog: open, save, message, confirm, prompt
 * store: get, set, delete, has, keys, clear
 * http: fetch, download
 * window: setTitle, setSize, minimize, maximize, fullscreen, close
 * clipboard: read, write
 * shell: open
 * system: platform, arch, homeDir, tempDir, hostname
 * app: quit, version, dataDir
 * notify: send
 * tray: set, remove, onClick
 * menu: set
 * process: exec
 * shortcuts: register, unregister, unregisterAll, isRegistered
 * updater: check, install, checkAndInstall, onProgress
 */
```

Copilot will use this as context for completions within the file.

## Which File Should I Use?

| Scenario | Use |
|----------|-----|
| Quick chat question | llms.txt |
| "Build me a complete app" | llms-full.txt |
| System prompt for an AI agent | llms-full.txt |
| Cursor `@` mention | llms-full.txt |
| Copilot context comment | Neither — use the short summary block above |
| Checking if an API exists | llms.txt |

## Keeping the Files Updated

The llms.txt and llms-full.txt files are generated from the same source as the documentation site. When a new API is added or an existing one changes, both files are regenerated. The URLs are stable and do not change between versions.

If you are caching the contents (for example, in a system prompt), check the file periodically for updates. The file includes a `Last-Updated` header with the generation date.
