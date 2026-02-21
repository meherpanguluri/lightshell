---
title: CLI
description: Complete reference for the lightshell command-line interface.
---

The `lightshell` CLI creates, develops, builds, and diagnoses LightShell applications. Install it with `go install github.com/lightshell-dev/lightshell@latest` or download a prebuilt binary from the releases page.

## Commands

### lightshell init

Create a new LightShell project with a starter template.

**Usage:**
```bash
lightshell init <project-name>
```

**What it does:**
1. Creates a new directory with the given name
2. Generates `lightshell.json` with sensible defaults
3. Creates a starter `index.html` with a basic UI and the LightShell client library included
4. Creates a minimal CSS file

**Example:**
```bash
lightshell init my-app
cd my-app
lightshell dev
```

**Output structure:**
```
my-app/
  lightshell.json
  index.html
  style.css
```

---

### lightshell dev

Start the development server. Opens a window running your app with hot reload, DevTools enabled, and a relaxed Content Security Policy.

**Usage:**
```bash
lightshell dev
```

**Behavior:**
- Watches the project directory for file changes
- Automatically reloads the webview when HTML, CSS, or JS files change
- DevTools are enabled (right-click to inspect)
- Uses the relaxed dev CSP (`default-src 'self' 'unsafe-inline' 'unsafe-eval' lightshell: http://localhost:*`)
- Console output from `console.log()` is printed to the terminal

**Example:**
```bash
cd my-app
lightshell dev
```

---

### lightshell build

Build a distributable application package. The output format depends on the current platform and the `--target` flag.

**Usage:**
```bash
lightshell build [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--target <format>` | Output format (see table below). Default: `app` on macOS, `appimage` on Linux |
| `--sign` | Code sign the build (macOS only, requires `build.mac.identity` in config) |
| `--notarize` | Notarize the build with Apple (macOS only, requires `--sign`) |
| `--devtools` | Include DevTools in the production build |

**Target formats:**

| Target | Platform | Output | Description |
|--------|----------|--------|-------------|
| `app` | macOS | `.app` bundle | Default macOS format, a standard application bundle |
| `dmg` | macOS | `.dmg` disk image | DMG with drag-to-Applications installer layout |
| `appimage` | Linux | `.AppImage` | Default Linux format, single portable executable |
| `deb` | Linux | `.deb` package | Debian/Ubuntu package for `apt install` |
| `rpm` | Linux | `.rpm` package | Fedora/RHEL package for `dnf install` |
| `all` | both | all formats | Build all formats available for the current OS |

**Examples:**
```bash
# Default build — .app on macOS, AppImage on Linux
lightshell build

# macOS DMG with code signing
lightshell build --target dmg --sign

# macOS DMG with signing and notarization
lightshell build --target dmg --sign --notarize

# Debian package
lightshell build --target deb

# RPM package
lightshell build --target rpm

# All formats for the current platform
lightshell build --target all

# Production build with DevTools for debugging
lightshell build --devtools
```

**Output:**
Build artifacts are placed in the `dist/` directory:
```
dist/
  MyApp.app/                      # macOS .app bundle
  MyApp-1.0.0.dmg                 # macOS DMG
  MyApp-1.0.0-arm64.AppImage      # Linux AppImage
  myapp_1.0.0_amd64.deb           # Debian package
  myapp-1.0.0-1.x86_64.rpm        # RPM package
```

**Build sizes:**
- The hello-world example produces a binary under 5MB
- Final binary size depends on embedded HTML/CSS/JS assets

---

### lightshell doctor

Check the development environment for required dependencies and common issues.

**Usage:**
```bash
lightshell doctor
```

**What it checks:**
- Go version (1.21+ required)
- CGO availability (`CGO_ENABLED=1`)
- Xcode Command Line Tools (macOS)
- WebKitGTK development headers (Linux: `libwebkit2gtk-4.1-dev`)
- GTK3 development headers (Linux: `libgtk-3-dev`)
- Code signing identity (if configured)
- Project structure validity

**Example output:**
```
LightShell Doctor
=================

[ok] Go 1.22.0
[ok] CGO enabled
[ok] Xcode Command Line Tools installed
[ok] lightshell.json found
[ok] Entry file index.html found
[warn] No build.appId set — using default

All checks passed.
```

---

## Common Workflows

### Create and Run a New App

```bash
lightshell init my-app
cd my-app
lightshell dev
```

### Build for Distribution

```bash
# macOS
lightshell build --target dmg --sign

# Linux
lightshell build --target deb
lightshell build --target rpm
lightshell build --target appimage
```

### Check Environment Before Building

```bash
lightshell doctor
lightshell build
```

### Debug a Production Build

```bash
lightshell build --devtools
open dist/MyApp.app
```

## Platform Notes

- On macOS, `lightshell build` requires Xcode Command Line Tools (`xcode-select --install`).
- On Linux, `lightshell build` requires `libwebkit2gtk-4.1-dev` and `libgtk-3-dev` packages.
- The `--sign` and `--notarize` flags are macOS-only. They are silently ignored on Linux.
- `--notarize` requires Apple Developer credentials. Set `APPLE_ID` and `APPLE_TEAM_ID` environment variables or pass them interactively.
- The `dist/` directory is created automatically. Previous builds in `dist/` are not cleaned — remove manually if needed.
- Cross-compilation is not supported in v1. Build on the target platform.
