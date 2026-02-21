---
title: Cross-Platform Development
description: Understand platform differences and use lightshell doctor to find compatibility issues.
---

LightShell targets macOS and Linux using system webviews — WKWebView on macOS and WebKitGTK on Linux. While both are WebKit-based, there are differences. LightShell aims for **~85-90% visual consistency**, not pixel-perfect matching. This guide covers what differs, what is auto-fixed, and what you need to handle manually.

## How LightShell Handles Differences

LightShell addresses cross-platform differences at three levels:

### 1. Auto-Polyfilled (You Do Nothing)

These are handled automatically by the normalization layer injected before your code:

- **Form element styling** — inputs, buttons, selects, textareas are reset to a consistent base
- **Scrollbar appearance** — thin, overlay-style scrollbars on both platforms
- **Font stack** — system font fallback chain that selects the best native font
- **Focus outlines** — consistent `:focus-visible` styles
- **`structuredClone`** — polyfilled on older WebKitGTK versions
- **Platform classes** — `platform-darwin` or `platform-linux` on `<html>`

### 2. Detectable by Scanner (lightshell doctor Warns You)

The compatibility scanner flags these with actionable advice:

- **`backdrop-filter`** — limited in WebKitGTK, auto-fallback injected but visually different
- **CSS nesting** — not supported in older WebKitGTK
- **`system-ui` font** — renders differently, scanner suggests explicit font stack
- **Container Queries** — version-dependent in WebKitGTK
- **`color-mix()`** — not available in older WebKitGTK
- **`:has()` selector** — limited support in WebKitGTK

### 3. Unfixable from Web Layer (Accept and Document)

These differences come from the rendering engine and OS, and cannot be fixed from JavaScript or CSS:

- **Font rasterization** — macOS uses CoreText, Linux uses FreeType. Text will look slightly different in weight and spacing.
- **GPU compositing** — different GPUs and drivers produce subtly different blending.
- **Color management** — macOS uses ColorSync ICC profiles; Linux color management varies by setup.
- **Subpixel rendering** — macOS and Linux handle subpixel antialiasing differently.

## Using lightshell doctor

The `lightshell doctor` command scans your source files and reports compatibility issues:

```bash
lightshell doctor
```

Example output:

```
LightShell Compatibility Report
================================

src/style.css
  ⚠  line 45: backdrop-filter — limited on Linux (WebKitGTK)
     → Auto-polyfill: fallback background injected at runtime

  ⚠  line 12: system-ui font — renders differently across platforms
     → Recommendation: use explicit font stack or bundle a web font

src/app.js
  ⚠  line 28: structuredClone() — missing on WebKitGTK < 2.40
     → Auto-polyfill: JSON-based clone injected at runtime

  ✗  line 55: Navigation API — not available on either platform webview
     → Use standard History API or lightshell.window instead

Summary: 1 error, 3 warnings (2 auto-polyfilled)
```

Severity levels:
- **Error** (✗) — will not work on one or both platforms, must fix
- **Warning** (⚠) — works but with differences, some are auto-polyfilled
- **Info** (ℹ) — minor difference, usually cosmetic

## Platform Detection

### In JavaScript

```js
const platform = await lightshell.system.platform()
// "darwin" on macOS, "linux" on Linux

if (platform === 'darwin') {
  // macOS-specific behavior
} else {
  // Linux-specific behavior
}
```

### In CSS

Use the platform classes added to `<html>`:

```css
/* Applied on all platforms */
.toolbar {
  height: 40px;
}

/* macOS-only */
.platform-darwin .toolbar {
  height: 38px; /* Slightly shorter to match native toolbar height */
}

/* Linux-only */
.platform-linux .toolbar {
  height: 42px; /* GTK toolbars tend to be slightly taller */
}
```

## Known Differences Catalog

### CSS Features

| Feature | macOS (WKWebView) | Linux (WebKitGTK 2.40+) | Notes |
|---------|-------------------|--------------------------|-------|
| `backdrop-filter` | Full support | Limited/no support | Auto-fallback provided |
| CSS Nesting (`&`) | Supported | Version-dependent | Use flat selectors for safety |
| `:has()` selector | Supported | Limited | Avoid for critical layout |
| `color-mix()` | Supported | Version-dependent | Use pre-computed colors as fallback |
| Container Queries | Supported | Version-dependent | Use media queries as fallback |
| View Transitions | Not supported | Not supported | Neither webview supports this |

### JavaScript APIs

| API | macOS | Linux | Notes |
|-----|-------|-------|-------|
| `structuredClone()` | Available | WebKitGTK 2.40+ | Auto-polyfilled |
| `Intl.Segmenter` | Available | Not available | Console warning emitted |
| `Navigation API` | Not available | Not available | Use History API |
| `showOpenFilePicker` | Not available | Not available | Use `lightshell.dialog.open()` |
| `Web USB/Bluetooth` | Not available | Not available | Not in webviews |

### Visual Differences

| Aspect | macOS | Linux | Impact |
|--------|-------|-------|--------|
| Font rendering | CoreText, smooth | FreeType, sharper | Text looks slightly different |
| Default font | San Francisco | Noto Sans / DejaVu | Similar weight and metrics |
| Scrollbar style | Native overlay | Custom (normalized) | Close match |
| Window chrome | macOS title bar | GTK title bar | Different but both native |
| Selection color | System blue | System blue (varies by theme) | May differ slightly |

## Writing Robust Cross-Platform Code

### Test on Both Platforms

The most reliable approach is to test on both macOS and Linux. Use CI to catch issues:

```yaml
jobs:
  test-macos:
    runs-on: macos-latest
    steps:
      - run: lightshell build

  test-linux:
    runs-on: ubuntu-latest
    steps:
      - run: sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      - run: lightshell build
```

### Use Progressive Enhancement

Do not rely on cutting-edge CSS features for critical layout. Use them as enhancements:

```css
/* Base — works everywhere */
.glass-panel {
  background: rgba(245, 245, 245, 0.95);
}

/* Enhancement — only on macOS where backdrop-filter works */
.platform-darwin .glass-panel {
  backdrop-filter: blur(20px);
  background: rgba(255, 255, 255, 0.6);
}
```

### Avoid Browser-Specific APIs

Do not use APIs that only exist in full browsers:

```js
// WRONG — not available in webviews
const handle = await window.showOpenFilePicker()

// RIGHT — use LightShell's native API
const path = await lightshell.dialog.open()
```

### Handle Missing APIs Gracefully

```js
// Check before using optional APIs
if (typeof Intl.Segmenter !== 'undefined') {
  const segmenter = new Intl.Segmenter('en', { granularity: 'word' })
  // use segmenter
} else {
  // fallback: split by spaces
  const words = text.split(/\s+/)
}
```

## WebKitGTK Version Requirements

LightShell requires WebKitGTK 2.40 or later on Linux. This version is available in:

- Ubuntu 23.10+
- Fedora 38+
- Arch Linux (rolling release, always current)
- Debian 13 (trixie)

For older distributions, users may need to install a newer WebKitGTK from a PPA or build from source. The `lightshell doctor` command checks the installed WebKitGTK version and warns if it is too old.
