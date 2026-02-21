---
title: LLM Integration
description: Use AI to build LightShell apps — llms.txt, prompting guides, and IDE integration.
---

LightShell is designed from the ground up so that AI agents can generate working desktop apps. The entire API surface is small, consistent, and discoverable — there are no hidden configuration files, no complex build steps, and no framework-specific abstractions. An AI that knows the 15 `lightshell.*` namespaces can build a functional desktop app in a single prompt.

This section covers the three integration points that make this work.

## Why LightShell Works Well with AI

Most desktop frameworks have steep learning curves — complex build systems, multiple process models, and large API surfaces. LightShell keeps things simple:

- **One global object.** Every API lives under `window.lightshell.*` — no imports, no modules, no bundlers.
- **Every call is async.** No callback patterns, no sync/async split. Always `await lightshell.something()`.
- **Plain HTML/CSS/JS.** The entry point is `src/index.html`. No JSX, no templates, no compilation step.
- **AI-friendly errors.** Permission denials and API errors include the exact fix needed, so an AI can self-correct.

## Integration Points

### 1. llms.txt — Context Injection

The fastest way to give an AI knowledge of LightShell. Two files are available:

- **[llms.txt](/llm/llms-txt/)** — a compact summary (~30 lines) with links to full documentation. Best for quick context in chat-based tools.
- **llms-full.txt** — the complete API reference (~800 lines). Best for comprehensive code generation where the AI needs every method signature and parameter.

Paste the URL into your AI conversation or add it to a system prompt.

### 2. Prompting Guide — How to Ask

Not all prompts are equal. The **[prompting guide](/llm/prompting-guide/)** covers how to structure requests so an AI produces working LightShell apps on the first try. Includes a template, tips, and guidance on what works well versus what needs iteration.

### 3. IDE Rules — Cursor and AGENTS.md

For AI-assisted editors, project-level rule files prevent common mistakes before they happen:

- **[Cursor Rules](/llm/cursor-rules/)** — a `.cursorrules` file that teaches Cursor about LightShell APIs, constraints, and platform differences.
- **[AGENTS.md](/llm/agents-md/)** — a project-level instruction file for Claude Code, Devin, and other agent-based tools.

## Quick Start

The fastest path from zero to a working app:

1. Open your AI tool (Claude, ChatGPT, Cursor, Copilot).
2. Paste this context URL: `https://lightshell.dev/llms-full.txt`
3. Use one of the [example prompts](/llm/example-prompts/) or write your own.
4. Save the generated files into a LightShell project (`lightshell init my-app`).
5. Run `lightshell dev` to see it live.

If the AI makes a mistake, check the [common mistakes](/llm/common-mistakes/) page for the fix — most issues are the same patterns repeated.

## All Pages in This Section

| Page | What It Covers |
|------|---------------|
| [llms.txt Specification](/llm/llms-txt/) | The llms.txt and llms-full.txt files for AI context injection |
| [Prompting Guide](/llm/prompting-guide/) | How to structure prompts for best results |
| [Cursor Rules](/llm/cursor-rules/) | .cursorrules file for Cursor IDE |
| [AGENTS.md](/llm/agents-md/) | Project-level instructions for Claude Code and other agents |
| [Example Prompts](/llm/example-prompts/) | Ready-to-use prompts for common app types |
| [Common AI Mistakes](/llm/common-mistakes/) | Frequent errors and their fixes |
