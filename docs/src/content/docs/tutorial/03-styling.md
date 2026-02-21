---
title: "3. Styling"
description: CSS tips for making LightShell apps look native across macOS and Linux.
---

LightShell apps run in system webviews, which means your CSS renders natively on each platform. This tutorial covers how to write CSS that looks great on both macOS and Linux, how to use platform-specific classes, and what to avoid.

## The Normalization Layer

LightShell injects a normalization stylesheet and polyfills before your code runs. These handle:

- **Form element reset** — strips OS-native styling so inputs, buttons, selects, and textareas look consistent
- **Scrollbar normalization** — uniform thin scrollbars on both platforms
- **System font stack** — a cross-platform font fallback chain
- **Focus styles** — consistent `:focus-visible` outlines

You do not need to include a CSS reset in your app. LightShell handles it.

## System Font Stack

LightShell sets a default font stack on `body`:

```css
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
}
```

This selects the best native font on each platform:
- **macOS**: San Francisco (via `-apple-system`)
- **Linux**: Noto Sans, or the system default sans-serif

If you want a specific font, bundle a web font or load one from a CDN. Avoid using `system-ui` alone — it renders differently across platforms.

## Platform-Specific CSS

LightShell adds a CSS class to the `<html>` element based on the current platform:

- `platform-darwin` on macOS
- `platform-linux` on Linux

Use these for targeted adjustments:

```css
/* Slightly larger text on Linux where font metrics differ */
.platform-linux body {
  font-size: 15px;
}

/* macOS-specific vibrancy effect */
.platform-darwin .sidebar {
  backdrop-filter: blur(20px);
  background: rgba(255, 255, 255, 0.7);
}

/* Linux fallback — no backdrop-filter support in WebKitGTK */
.platform-linux .sidebar {
  background: rgba(245, 245, 245, 0.95);
}
```

## Making Your App Look Native

### Use System Colors

Modern CSS provides system color keywords that match OS themes:

```css
body {
  background: Canvas;
  color: CanvasText;
}

button {
  background: ButtonFace;
  color: ButtonText;
  border: 1px solid ButtonBorder;
}

::selection {
  background: Highlight;
  color: HighlightText;
}
```

### Subtle Shadows and Borders

Native apps use subtle depth cues, not heavy drop shadows:

```css
.card {
  background: white;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* Focused cards get slightly more shadow */
.card:focus-within {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}
```

### Proper Spacing

Desktop apps tend to have tighter spacing than web pages:

```css
:root {
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 24px;
}

body {
  margin: 0;
  padding: var(--spacing-md);
}
```

## What to Avoid

### backdrop-filter

The `backdrop-filter` CSS property (for frosted glass effects) has limited support in WebKitGTK on Linux. LightShell includes an automatic fallback that substitutes a solid semi-transparent background, but the effect will not look identical.

```css
/* This works great on macOS but falls back on Linux */
.toolbar {
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.6);
}

/* Better: provide an explicit fallback with platform classes */
.platform-darwin .toolbar {
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.6);
}
.platform-linux .toolbar {
  background: rgba(245, 245, 245, 0.95);
}
```

### CSS Nesting

Native CSS nesting (the `&` selector) may not work on older versions of WebKitGTK. Use flat selectors for maximum compatibility:

```css
/* Avoid on older WebKitGTK */
.card {
  & .title { font-weight: bold; }
}

/* Safe everywhere */
.card .title {
  font-weight: bold;
}
```

### system-ui Font

Using `font-family: system-ui` alone produces different results on macOS vs Linux. Always include a full fallback chain as shown above.

## Scrollbar Styling

LightShell normalizes scrollbars to thin, overlay-style bars:

```css
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.2); border-radius: 4px; }
```

You can override these in your own CSS if you want different scrollbar styling.

## Dark Mode

Detect the user's system preference and adjust:

```css
@media (prefers-color-scheme: dark) {
  body {
    background: #1a1a1a;
    color: #e0e0e0;
  }

  .card {
    background: #2a2a2a;
    border-color: rgba(255, 255, 255, 0.1);
  }
}
```

Or use the system color keywords (`Canvas`, `CanvasText`, etc.) which automatically adapt to dark mode.

## Example: Native-Feeling Layout

Here is a complete example of a native-looking app layout:

```css
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  padding: 12px;
  background: #f5f5f5;
  border-right: 1px solid #e0e0e0;
  overflow-y: auto;
}

.platform-linux .sidebar {
  background: #f0f0f0;
}

.content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.sidebar-item {
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}

.sidebar-item:hover {
  background: rgba(0, 0, 0, 0.06);
}

.sidebar-item.active {
  background: rgba(0, 100, 255, 0.1);
  color: #0064ff;
}
```

This produces a sidebar-content layout that feels native on both platforms, with subtle hover effects and proper spacing.

## Run the Compatibility Scanner

Use `lightshell doctor` to check your CSS and JS for cross-platform issues:

```bash
lightshell doctor
```

This scans your `src/` files and reports any known incompatibilities, along with whether they are auto-polyfilled or need manual attention.

Next, let's learn how to package your app for distribution.
