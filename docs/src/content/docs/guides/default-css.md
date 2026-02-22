---
title: Default Styles
description: The built-in CSS that ships with every LightShell app.
---

Every LightShell app ships with a small default stylesheet that provides system fonts, sensible colors, form resets, scrollbar styling, and dark mode support. The defaults are designed to make your app look native and consistent across macOS and Linux without writing any CSS.

## What Ships by Default

The default stylesheet is injected before your HTML loads. It includes:

- **System font stack** -- uses the native system font on each platform
- **Light and dark mode colors** -- automatic via `light-dark()`, follows OS preference
- **Form element reset** -- consistent buttons, inputs, selects, and textareas
- **Scrollbar styling** -- subtle, consistent scrollbars on all platforms
- **Focus ring normalization** -- visible focus rings for keyboard navigation, hidden for mouse clicks
- **Stable scrollbar gutter** -- prevents layout shift when scrollbars appear or disappear

## CSS Variables

LightShell defines a set of CSS custom properties that you can use and override. All variables use the `--ls-` prefix.

### Colors

```css
:root {
  --ls-bg:              light-dark(#ffffff, #1a1a2e);
  --ls-bg-secondary:    light-dark(#f5f5f7, #16213e);
  --ls-bg-tertiary:     light-dark(#e8e8ed, #0f3460);
  --ls-text:            light-dark(#1d1d1f, #e8e8ed);
  --ls-text-secondary:  light-dark(#6e6e73, #a0a0b0);
  --ls-text-tertiary:   light-dark(#999999, #707080);
  --ls-border:          light-dark(#d2d2d7, #2a2a40);
  --ls-accent:          light-dark(#0071e3, #64b5f6);
  --ls-accent-hover:    light-dark(#0077ed, #90caf9);
  --ls-danger:          light-dark(#ff3b30, #ff6b6b);
  --ls-success:         light-dark(#34c759, #69f0ae);
  --ls-warning:         light-dark(#ff9500, #ffb74d);
}
```

### Typography

```css
:root {
  --ls-font-family:   -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
                      Helvetica, Arial, sans-serif, "Apple Color Emoji", "Noto Color Emoji";
  --ls-font-mono:     ui-monospace, "SF Mono", "JetBrains Mono", Menlo, Consolas,
                      "Liberation Mono", monospace;
  --ls-font-size:     14px;
  --ls-font-size-sm:  12px;
  --ls-font-size-lg:  16px;
  --ls-font-size-xl:  20px;
  --ls-line-height:   1.5;
}
```

### Spacing

```css
:root {
  --ls-space-xs:  4px;
  --ls-space-sm:  8px;
  --ls-space-md:  16px;
  --ls-space-lg:  24px;
  --ls-space-xl:  32px;
}
```

### Borders and Radius

```css
:root {
  --ls-radius-sm:  4px;
  --ls-radius-md:  8px;
  --ls-radius-lg:  12px;
  --ls-radius-full: 9999px;
  --ls-border-width: 1px;
}
```

### Shadows

```css
:root {
  --ls-shadow-sm:  0 1px 2px rgba(0, 0, 0, 0.05);
  --ls-shadow-md:  0 4px 6px rgba(0, 0, 0, 0.07);
  --ls-shadow-lg:  0 10px 15px rgba(0, 0, 0, 0.1);
}
```

### Scrollbar

```css
:root {
  --ls-scrollbar-width:       8px;
  --ls-scrollbar-track:       transparent;
  --ls-scrollbar-thumb:       rgba(128, 128, 128, 0.4);
  --ls-scrollbar-thumb-hover: rgba(128, 128, 128, 0.6);
}
```

## Customizing

Override any variable in your own CSS to customize the look:

```css
:root {
  --ls-accent: #8b5cf6;
  --ls-accent-hover: #a78bfa;
  --ls-radius-md: 4px;
  --ls-font-size: 15px;
}
```

This works because the defaults are injected first, and your styles load after. Standard CSS cascade rules apply -- your overrides win.

### Example: Custom Theme

```css
:root {
  /* Purple accent */
  --ls-accent: #8b5cf6;
  --ls-accent-hover: #a78bfa;

  /* Rounder corners */
  --ls-radius-sm: 6px;
  --ls-radius-md: 12px;
  --ls-radius-lg: 16px;

  /* Tighter spacing */
  --ls-space-md: 12px;
  --ls-space-lg: 20px;
}
```

### Example: Dark-Only App

If your app is always dark, override the light-dark values directly:

```css
:root {
  color-scheme: dark;
  --ls-bg: #0d1117;
  --ls-bg-secondary: #161b22;
  --ls-text: #c9d1d9;
  --ls-border: #30363d;
}
```

## Disabling Default Styles

If you want full control over your CSS and do not want the defaults at all, disable them in `lightshell.json`:

```json
{
  "defaults": {
    "css": false
  }
}
```

When disabled, no default stylesheet is injected. You are responsible for all styling, including form resets, scrollbars, and font stacks. The CSS variables (`--ls-*`) will not be defined unless you define them yourself.

## Dark Mode

The default styles use `light-dark()` for all color values, which automatically follows the OS color scheme preference. No JavaScript is needed -- the CSS adapts when the user changes their system appearance setting.

```css
/* These adapt automatically */
body {
  background: var(--ls-bg);
  color: var(--ls-text);
}
```

The `color-scheme` property is set on the root element:

```css
:root {
  color-scheme: light dark;
}
```

### Forcing a Color Scheme

To force your app to always use light or dark mode regardless of OS preference:

```js
await lightshell.window.setColorScheme('dark')   // always dark
await lightshell.window.setColorScheme('light')   // always light
await lightshell.window.setColorScheme('system')  // follow OS (default)
```

Or in CSS:

```css
:root {
  color-scheme: dark;  /* force dark */
}
```

## Form Element Reset

The default stylesheet resets form elements to look consistent:

```css
input, select, textarea, button {
  -webkit-appearance: none;
  appearance: none;
  font-family: inherit;
  font-size: inherit;
}
```

This removes platform-specific styling from buttons, inputs, dropdowns, and textareas. They inherit the app font and size instead of using the OS default widget styles.

## Scrollbar Styling

Scrollbars are styled to be subtle and consistent:

```css
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(128, 128, 128, 0.4); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: rgba(128, 128, 128, 0.6); }
```

The `scrollbar-gutter: stable` rule on the root element prevents layout shifts when scrollbar visibility changes.

## Focus Ring Normalization

Focus rings are visible for keyboard navigation but hidden for mouse clicks:

```css
:focus-visible {
  outline: 2px solid highlight;
  outline-offset: 2px;
}
:focus:not(:focus-visible) {
  outline: none;
}
```

This provides accessibility for keyboard users without showing distracting outlines when clicking with a mouse.

## Platform-Specific Notes

- On **macOS**, `-apple-system` resolves to San Francisco. The system font looks native by default.
- On **Linux**, `"Noto Sans"` is the primary fallback. Most Linux distributions ship Noto Sans. If it is missing, the browser falls back to Helvetica or Arial.
- The `light-dark()` function is supported in WebKit 17.4+ (macOS) and WebKitGTK 2.44+ (Linux). On older WebKitGTK versions, the dark mode values may not apply -- consider setting explicit dark mode colors if you need to support older distributions.
