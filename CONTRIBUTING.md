# Contributing to LightShell

Thanks for your interest in contributing! LightShell is an open-source desktop app framework and we welcome contributions of all kinds.

## Getting Started

### Prerequisites

- **Go 1.23+** — [install](https://go.dev/dl/)
- **Node.js 20+** — [install](https://nodejs.org/)
- **macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **Linux**: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`

### Setup

```bash
git clone https://github.com/meherpanguluri/lightshell.git
cd lightshell
go build ./cmd/lightshell
```

### Running locally

```bash
# Build the CLI
go build -o lightshell ./cmd/lightshell

# Create a test app
./lightshell init test-app
cd test-app
../lightshell dev
```

### Running tests

```bash
go test ./tests/...
```

## Project Structure

```
cmd/lightshell/     CLI entry point
internal/
  runtime/          Core app lifecycle + webview management
  webview/          WKWebView (macOS) / WebKitGTK (Linux)
  ipc/              Unix domain socket IPC server
  api/              Native API handlers (fs, dialog, clipboard, etc.)
  cli/              CLI commands (init, dev, build, doctor)
  compat/           Cross-platform compatibility layer
  security/         Permission system
client/             JS client library (injected into webview)
templates/          Project templates
packaging/          .app and AppImage bundlers
website/            Landing page (static HTML/CSS/JS)
docs/               Documentation site (Astro/Starlight)
tests/              Test suite
```

## What to Contribute

### Good first issues

- Add new compatibility rules to `internal/compat/rules.go`
- Improve error messages in CLI commands
- Add examples to the documentation
- Fix CSS issues in the website

### Bigger contributions

- New native API handlers (e.g., screen capture, global shortcuts)
- Linux AppImage packaging improvements
- New project templates
- Performance optimizations

## Development Guidelines

### Code style

- **Go**: Run `gofmt` before committing. Follow standard Go conventions.
- **JS/CSS**: No build tools. Plain vanilla JS and CSS. No frameworks.
- **Comments**: Explain *why*, not *what*.

### Commit messages

Keep them short and descriptive:

```
Fix hot reload not triggering on CSS changes
Add clipboard permission check to fs.writeFile
Update getting-started docs with Linux instructions
```

### Pull requests

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `go build ./cmd/lightshell` to verify it compiles
4. Run `go test ./tests/...` to verify tests pass
5. Open a PR with a clear description of what changed and why

### Architecture decisions

- **No Windows code** — Windows support is planned for v2
- **No bundled browsers** — we use system webviews only
- **No Node.js at runtime** — the JS client uses `window.lightshell.*` APIs
- **No npm dependencies at runtime** — external libs via CDN only
- **Go build tags** for platform-specific code (`//go:build darwin` / `//go:build linux`)

## Reporting Bugs

Open an issue at [github.com/meherpanguluri/lightshell/issues](https://github.com/meherpanguluri/lightshell/issues) with:

- What you expected to happen
- What actually happened
- Your OS and version
- Steps to reproduce

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
