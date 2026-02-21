# LightShell

Build desktop apps with JavaScript. Ship them under 5MB.

---

## Quick Start

```bash
npx @lightshell/create my-app
cd my-app
npx @lightshell/cli dev
```

Or install globally:

```bash
npm install -g @lightshell/cli
lightshell init my-app
cd my-app
lightshell dev
```

That's it. No Rust toolchain. No C++ compiler. No 10-minute setup. You write JS/HTML/CSS, LightShell handles the rest.

## Why LightShell?

| | Electron | Tauri | LightShell |
|---|---------|-------|------------|
| You write | JS + Node | JS + Rust | **JS only** |
| Binary size | ~180MB | ~8MB | **~2.8MB** |
| Startup time | ~1.2s | ~0.5s | **~0.3s** |
| Memory usage | ~120MB | ~45MB | **~30MB** |
| Setup time | 2 min | 10 min | **30 sec** |

LightShell apps are pure JS/HTML/CSS with a simple API surface — making them easy for AI to generate end-to-end.

## How It Works

LightShell uses **system webviews** (WKWebView on macOS, WebKitGTK on Linux) instead of bundling Chromium. Your HTML/CSS/JS is embedded into a single Go binary via `embed.FS` at compile time.

The result: a native app with no runtime dependencies, no bundled browser, and a binary under 5MB.

```
Your JS/HTML/CSS
      |
  [ LightShell CLI ]
      |
  Single native binary (~2.8MB)
      |
  System webview (WKWebView / WebKitGTK)
      |
  Native desktop app
```

You never see Go. You never write Go. You never configure Go. It's an implementation detail, like how esbuild uses Go but nobody cares.

## Native APIs -- All from JavaScript

```js
// Read and write files
const content = await lightshell.fs.readFile('./data.json')
await lightshell.fs.writeFile('./output.txt', 'Hello from LightShell')

// Native file dialogs
const file = await lightshell.dialog.open({
  filters: [{ name: 'JSON', extensions: ['json'] }]
})

// Save dialogs
const path = await lightshell.dialog.save({ defaultPath: 'output.txt' })

// System notifications
lightshell.notify.send('Saved!', 'Your file has been saved')

// Clipboard
await lightshell.clipboard.write('Copied to clipboard')
const text = await lightshell.clipboard.read()

// Open URLs in default browser
lightshell.shell.open('https://lightshell.sh')

// Window management
await lightshell.window.setTitle('My App')
await lightshell.window.setSize(1200, 800)

// System info
const platform = await lightshell.system.platform()  // 'darwin' or 'linux'
const home = await lightshell.system.homeDir()
```

Every API call returns a Promise. No callbacks, no event emitters, no Node.js globals.

## Project Structure

```
my-app/
  lightshell.json      # App config (name, window size, icon)
  src/
    index.html         # Entry point
    app.js             # Your application code
    style.css          # Your styles
```

**lightshell.json:**

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "My App",
    "width": 1024,
    "height": 768,
    "minWidth": 400,
    "minHeight": 300,
    "resizable": true
  },
  "build": {
    "icon": "assets/icon.png",
    "appId": "com.example.myapp"
  }
}
```

## CLI Commands

```bash
lightshell init [name]     # Create a new project
lightshell dev             # Run with hot reload
lightshell build           # Build for current platform
lightshell doctor          # Check cross-platform compatibility
lightshell version         # Print version
```

## Platform Support

| Platform | Architecture | Webview | Package Format |
|----------|-------------|---------|----------------|
| macOS | arm64, x64 | WKWebView | .app bundle |
| Linux | x64, arm64 | WebKitGTK 2.40+ | AppImage |

Windows support is planned for v2.

## Development

```bash
# Clone the repo
git clone https://github.com/meherpanguluri/lightshell.git
cd lightshell

# Build
go build ./cmd/lightshell

# Run tests
go test ./...

# Linux dependencies
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

## Documentation

- [Getting Started](https://lightshell.sh/docs/getting-started) -- Install and build your first app in 5 minutes
- [API Reference](https://lightshell.sh/docs/api/window) -- Complete API documentation
- [Tutorial](https://lightshell.sh/docs/tutorial/01-your-first-app) -- Build a real app step by step
- [Cross-Platform Guide](https://lightshell.sh/docs/guides/cross-platform) -- Handle platform differences
- [Playground](https://lightshell.sh/playground) -- Try LightShell in your browser

## License

[MIT](LICENSE)
