---
title: Cross-Platform Rendering
description: How LightShell achieves visual consistency across macOS and Linux.
---

LightShell uses the system webview on each platform: WKWebView on macOS and WebKitGTK on Linux. Both are WebKit-based, which means your HTML, CSS, and JavaScript run on the same engine family. But "same engine" does not mean "identical rendering." This page explains the differences, what LightShell does to close the gap, and what you should expect.

## The Two Webviews

| | macOS (WKWebView) | Linux (WebKitGTK) |
|---|---|---|
| Engine | WebKit (Apple's fork) | WebKit (GNOME's port) |
| Version updates | With macOS updates | With distro packages |
| Minimum version | Ships with macOS 11+ | 2.40+ required |
| JavaScript engine | JavaScriptCore | JavaScriptCore |
| Rendering | CoreGraphics + Metal | Cairo + OpenGL/Vulkan |
| Text rendering | CoreText | FreeType + HarfBuzz |

Both webviews share the same core engine (WebKit) and JavaScript runtime (JavaScriptCore). The differences are in the platform integration layer: how text is rasterized, how graphics are composited, and how quickly new web platform features are adopted.

## The Consistency Target

LightShell aims for **85-90% visual consistency** between macOS and Linux. That number is realistic because:

- Layout, box model, flexbox, grid, and positioning work identically (same engine)
- Colors, borders, shadows, and transforms work identically
- Font metrics differ slightly (CoreText vs FreeType use different hinting)
- Form elements look different by default (each platform has its own native style)
- Some newer CSS features land in WKWebView before WebKitGTK

The remaining 10-15% difference is what you would expect from any cross-platform app: fonts render slightly differently, scrollbars look different, and native UI elements follow platform conventions.

## What LightShell Does Automatically

LightShell injects a set of normalizations and polyfills before your code runs. These are embedded in the binary and add less than 2KB to the app.

### CSS Normalization

A CSS reset is injected to make form elements and scrollbars consistent:

```css
/* Form element reset */
input, select, textarea, button {
  -webkit-appearance: none;
  appearance: none;
  font-family: inherit;
  font-size: inherit;
}

/* Scrollbar normalization */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(128, 128, 128, 0.4); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: rgba(128, 128, 128, 0.6); }

/* Prevent layout shift from scrollbar appearance */
html { scrollbar-gutter: stable; }

/* System font stack */
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif, "Apple Color Emoji", "Noto Color Emoji";
}

/* Focus ring normalization */
:focus-visible {
  outline: 2px solid highlight;
  outline-offset: 2px;
}
:focus:not(:focus-visible) {
  outline: none;
}
```

Without the form element reset, a `<button>` on macOS looks like a rounded macOS button while on Linux it looks like a GTK button. With the reset, both platforms render a plain, unstyled button that you control with your own CSS.

### JavaScript Polyfills

LightShell polyfills JavaScript APIs that are missing on older WebKitGTK versions but present on WKWebView:

| API | Missing on WebKitGTK < | Polyfill |
|-----|------------------------|----------|
| `structuredClone` | 2.40 | JSON round-trip (no transfer support) |
| `Array.prototype.group` | 2.44 | Reduce-based implementation |
| `Promise.withResolvers` | 2.44 | Manual promise construction |
| `Set.prototype.union/intersection/difference` | 2.44 | Iterative set operations |
| `Object.groupBy` | 2.44 | Iterative grouping |

Each polyfill checks for the native implementation first (`if (typeof structuredClone === 'undefined')`) and only activates if the native API is missing. On macOS, where these APIs are always present, no polyfill code runs.

### Platform CSS Classes

LightShell adds CSS classes to the `<html>` element so you can write platform-specific styles:

```html
<!-- macOS -->
<html class="platform-darwin arch-arm64">

<!-- Linux -->
<html class="platform-linux arch-x64">
```

Use these for targeted adjustments:

```css
/* Slightly larger text on Linux to compensate for different font metrics */
.platform-linux body {
  font-size: 15px;
}

/* macOS-specific title bar spacing */
.platform-darwin .titlebar {
  padding-top: 28px;
}
```

## What Cannot Be Fixed

Some differences are fundamental to the platform and cannot be normalized by CSS or JavaScript.

### Font Rasterization

macOS uses CoreText with subpixel antialiasing optimized for Retina displays. Linux uses FreeType with configurable hinting (usually slight or no hinting on modern desktops). The same font at the same size will look slightly different:

- **macOS:** Smoother, slightly heavier glyphs, optimized for high-DPI
- **Linux:** Sharper edges, varies by distribution's FreeType configuration

You cannot and should not try to make fonts render identically. Use the system font stack (injected by LightShell) and accept that text will look "native" on each platform, which is actually what users expect.

### GPU Compositing

macOS uses Metal for GPU-accelerated compositing. Linux uses OpenGL or Vulkan depending on the system. This affects:

- CSS `backdrop-filter` performance (can be slow on some Linux GPU drivers)
- Complex `box-shadow` rendering (minor differences in shadow blur)
- CSS `filter` effects (identical visually, but different performance characteristics)

### Color Management

macOS has system-wide color management (ColorSync). Most Linux desktops do not. This means:

- Colors in your CSS look the same numerically but may appear slightly different perceptually
- Images with embedded color profiles render correctly on macOS and are shown as-is on Linux
- For most apps, this difference is invisible. For color-critical apps (design tools, photo editors), test on both platforms.

## Scanner Warnings

The `lightshell doctor` command scans your HTML, CSS, and JavaScript for APIs that are known to have cross-platform issues:

```bash
lightshell doctor
```

Example output:

```
Scanning src/ for cross-platform issues...

⚠ src/app.js:42 — Intl.Segmenter is unavailable on Linux.
  Use Intl.BreakIterator or split text manually.

⚠ src/styles.css:18 — :has() selector has limited support on WebKitGTK < 2.42.
  Use JavaScript to toggle classes instead.

⚠ src/styles.css:55 — CSS nesting requires WebKitGTK 2.44+.
  Flatten nested CSS rules for Linux compatibility.

3 warnings, 0 errors
```

### APIs That Cannot Be Polyfilled

These APIs are missing on WebKitGTK and too large or too complex to polyfill:

| API | Why no polyfill | Alternative |
|-----|----------------|-------------|
| `Intl.Segmenter` | Polyfill is ~200KB | Use `Intl.BreakIterator` or split manually |
| Navigation API | Not in any webview | Use History API or `lightshell.window` |
| View Transitions API | Missing on WebKitGTK | Use CSS animations |
| `:has()` selector | Missing on WebKitGTK < 2.42 | Toggle classes with JavaScript |
| CSS Nesting | Missing on WebKitGTK < 2.44 | Flatten nested rules |
| `@container` queries | Partial on older WebKitGTK | Test thoroughly on Linux |
| File System Access API | Not in any webview | Use `lightshell.dialog.open()` + `lightshell.fs` |
| Web USB/Bluetooth/Serial | Not in any webview | Hardware APIs are not available in webview contexts |

The scanner catches these and tells you what to use instead.

## Practical Recommendations

**Test on both platforms early.** Do not build your entire app on macOS and then try Linux at the end. Run `lightshell dev` on both platforms regularly, or use CI with Linux runners to catch issues.

**Use the system font stack.** Do not specify a single font like `font-family: "SF Pro"`. The injected font stack falls through gracefully on each platform.

**Avoid cutting-edge CSS.** Features that shipped in Safari in the last 6 months may not be in WebKitGTK yet. Stick to features supported by WebKitGTK 2.40+ (see the [WebKitGTK releases page](https://webkitgtk.org/releases/) for details).

**Embrace platform conventions.** Your app does not need to look identical on both platforms. macOS users expect rounded corners and light drop shadows. Linux users expect sharper edges and flatter UI. Platform CSS classes let you adjust for each.

**Use `lightshell doctor` before shipping.** It catches the most common cross-platform issues in seconds.
