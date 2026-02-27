---
title: React & Svelte
description: Use React or Svelte with LightShell via Vite integration.
---

LightShell works great with vanilla HTML/CSS/JS, but you can also use **React** or **Svelte** with full Vite-powered HMR in dev mode and optimized builds for production.

## How It Works

LightShell owns the native shell (webview, system APIs, packaging). Vite owns the web compilation (JSX, Svelte components, HMR). When you use a framework template:

- `lightshell dev` starts Vite's dev server, then loads it in the native webview
- `lightshell build` runs `vite build`, then embeds the static output into the native binary
- The `lightshell.*` APIs are injected into the webview automatically — no import needed

## React

### Create a React project

```bash
lightshell init my-app --template react
cd my-app
npm install
lightshell dev
```

This creates a project with Vite + React pre-configured:

```
my-app/
  lightshell.json       # LightShell config (entry: dist/index.html)
  package.json          # React + Vite dependencies
  vite.config.js        # Vite config with React plugin
  index.html            # Vite entry point
  src/
    main.jsx            # React root
    App.jsx             # Your app component
    App.css             # Styles
```

### Using LightShell APIs in React

The `lightshell` global is available in any component — no import required:

```jsx
import { useState, useEffect } from 'react'

function App() {
  const [platform, setPlatform] = useState('')

  useEffect(() => {
    lightshell.system.platform().then(setPlatform)
  }, [])

  return <p>Running on {platform}</p>
}
```

### Build for production

```bash
lightshell build
```

This runs `npm run build` (Vite) first, then packages the output into a native `.app` or AppImage.

## Svelte

### Create a Svelte project

```bash
lightshell init my-app --template svelte
cd my-app
npm install
lightshell dev
```

Project structure:

```
my-app/
  lightshell.json
  package.json          # Svelte 5 + Vite dependencies
  vite.config.js        # Vite config with Svelte plugin
  index.html
  src/
    main.js             # Svelte mount
    App.svelte          # Your app component
    app.css             # Styles
```

### Using LightShell APIs in Svelte

```svelte
<script>
  let platform = $state('')

  $effect(() => {
    lightshell.system.platform().then(p => platform = p)
  })
</script>

<p>Running on {platform}</p>
```

### Build for production

```bash
lightshell build
```

## How `lightshell.*` works with frameworks

LightShell injects the client library (`lightshell.js`) as a user script in the webview. This runs **before** any page scripts, so `window.lightshell` is always available by the time your React or Svelte code executes.

You do not need to import or install anything — just use the `lightshell` global directly.

## Configuration

Framework projects use two extra fields in `lightshell.json`:

```json
{
  "entry": "dist/index.html",
  "devCommand": "npm run dev -- --port 5188",
  "buildCommand": "npm run build"
}
```

| Field | Description |
|-------|-------------|
| `entry` | Path to the build output HTML (Vite outputs to `dist/` by default) |
| `devCommand` | Command to start the dev server. LightShell runs this and loads its URL in the webview. |
| `buildCommand` | Command to build for production. LightShell runs this before packaging. |

These fields are optional. If omitted, LightShell uses its built-in static file server (vanilla mode).

## Custom Vite config

The generated `vite.config.js` works out of the box. You can customize it freely — just keep these constraints:

- `build.outDir` must match the directory in `lightshell.json`'s `entry` path (default: `dist`)
- `server.strictPort: true` is recommended so port mismatches fail loudly
- The `--port` in `devCommand` must match your Vite config

## Using pnpm or yarn

The templates default to npm, but you can use any package manager. Just update the `devCommand` and `buildCommand` in `lightshell.json`:

```json
{
  "devCommand": "pnpm dev --port 5188",
  "buildCommand": "pnpm build"
}
```

## Adding TypeScript

Both templates use JavaScript by default. To add TypeScript:

1. Install TypeScript: `npm install -D typescript`
2. Rename `.jsx` files to `.tsx` (React) or add `lang="ts"` to `<script>` tags (Svelte)
3. Add a `tsconfig.json` (Vite handles compilation automatically)

For the `lightshell` global, add a type declaration file:

```ts
// src/lightshell.d.ts
declare const lightshell: {
  system: {
    platform(): Promise<string>
    arch(): Promise<string>
    homeDir(): Promise<string>
    tempDir(): Promise<string>
    hostname(): Promise<string>
  }
  window: {
    setTitle(title: string): Promise<void>
    setSize(width: number, height: number): Promise<void>
    getSize(): Promise<{ width: number; height: number }>
    // ... see API reference for full types
  }
  fs: {
    readFile(path: string, encoding?: string): Promise<string>
    writeFile(path: string, data: string): Promise<void>
    // ... see API reference for full types
  }
  // ... other namespaces
}
```

## Troubleshooting

**"node_modules not found"** — Run `npm install` before `lightshell dev` or `lightshell build`.

**"dev server did not start within 30s"** — Check that the port in `devCommand` isn't already in use. Try changing the port number.

**"lightshell is not defined"** — This can happen if your code runs before the webview injects the client library. In React, use `useEffect`; in Svelte, use `$effect` or `onMount`. Do not call `lightshell.*` at the module top level.
