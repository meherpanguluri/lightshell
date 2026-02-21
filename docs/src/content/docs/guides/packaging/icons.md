---
title: App Icons
description: Create and configure app icons for macOS and Linux.
---

A good app icon gives your application a professional identity. LightShell takes a single PNG image and converts it into the correct format for each platform — `.icns` for macOS and standard icon paths for Linux.

## Setting the Icon

Specify your icon in `lightshell.json`:

```json
{
  "build": {
    "icon": "assets/icon.png"
  }
}
```

The path is relative to your project root. The icon file should be a PNG image.

## Recommended Specifications

| Property | Recommendation |
|----------|---------------|
| Format | PNG |
| Size | 1024x1024 pixels |
| Minimum size | 512x512 pixels |
| Shape | Square |
| Color depth | 32-bit (RGBA) |
| Transparency | Supported |
| File size | Under 500KB |

A 1024x1024 source image ensures sharp rendering at all display sizes and resolutions, including Retina displays on macOS.

## How Icons Are Used

### macOS

On macOS, LightShell converts your PNG to `.icns` format during `lightshell build`. The `.icns` file contains the icon at multiple resolutions:

| Size | Scale | Pixels | Used For |
|------|-------|--------|----------|
| 16x16 | 1x | 16x16 | Finder list view, menu bar |
| 16x16 | 2x | 32x32 | Retina Finder list view |
| 32x32 | 1x | 32x32 | Finder icon view (small) |
| 32x32 | 2x | 64x64 | Retina Finder icon view (small) |
| 128x128 | 1x | 128x128 | Finder icon view (medium) |
| 128x128 | 2x | 256x256 | Retina Finder icon view (medium) |
| 256x256 | 1x | 256x256 | Finder icon view (large) |
| 256x256 | 2x | 512x512 | Retina Finder icon view (large) |
| 512x512 | 1x | 512x512 | Finder icon view (extra large) |
| 512x512 | 2x | 1024x1024 | Retina Finder icon view (extra large) |

All these resolutions are generated automatically from your single source image. The icon appears in:

- The Dock
- Finder
- Spotlight results
- the About dialog
- the application switcher (Cmd+Tab)

### Linux

On Linux, the icon PNG is used as-is. It is placed at the standard icon path for desktop integration:

- **AppImage** — embedded in the AppImage alongside the `.desktop` entry
- **.deb** — installed to `/usr/share/icons/hicolor/256x256/apps/myapp.png`
- **.rpm** — installed to `/usr/share/icons/hicolor/256x256/apps/myapp.png`

The icon appears in the application launcher, taskbar, and window decorations (depending on the desktop environment).

## Creating an Icon

### From Scratch

Use any image editor that can export PNG:

- **GIMP** (free) — create a 1024x1024 canvas, design your icon, export as PNG
- **Figma** (free tier) — design at 1024x1024 in a frame, export as PNG at 1x
- **Inkscape** (free) — design as SVG, export as 1024x1024 PNG
- **Affinity Designer** — work in a 1024x1024 pixel document, export as PNG
- **Photoshop** — 1024x1024 canvas, Save As PNG

### Design Guidelines

**Keep it simple.** Your icon will be displayed as small as 16x16 pixels in some contexts. Fine details disappear at small sizes. Use bold shapes and high contrast.

**Test at small sizes.** After designing at 1024x1024, zoom out or resize to 32x32 and 16x16 to verify it is still recognizable. If critical details vanish, simplify.

**Use a distinct silhouette.** The icon should be identifiable by its outline alone. Avoid relying solely on color or text to convey meaning.

**Consider transparency.** macOS icons traditionally use a transparent background with a rounded-rectangle shape (macOS automatically applies the rounded mask in some contexts). On Linux, the desktop environment may display icons with or without a background depending on the theme.

**Avoid text in icons.** Text becomes unreadable at small sizes. If you must include text, limit it to one or two large characters.

## Default Icon

If no icon is specified in `lightshell.json`, LightShell uses a default icon — a simple geometric mark that identifies it as a LightShell app. Replace it with your own icon before distributing.

## Project Structure

A typical icon setup in your project:

```
my-app/
  assets/
    icon.png              # 1024x1024 source icon
  src/
    index.html
    app.js
    style.css
  lightshell.json         # "build": { "icon": "assets/icon.png" }
```

You can place the icon file anywhere in your project — just update the path in `lightshell.json` to match.

## Multiple Source Files

If you want to provide pre-rendered icons at specific sizes instead of relying on automatic downscaling, you can place them in your assets directory:

```
assets/
  icon.png                # 1024x1024 (primary — referenced in lightshell.json)
  icon-512.png            # 512x512 (optional)
  icon-256.png            # 256x256 (optional)
```

Currently, LightShell uses only the single icon file specified in `lightshell.json` and generates all other sizes from it. Multi-size source support may be added in a future version.

## Troubleshooting

### Icon not appearing on macOS

- Verify the path in `lightshell.json` is correct and the file exists
- Ensure the PNG is a valid image (open it in Preview to check)
- After rebuilding, macOS may cache the old icon. Clear the icon cache:

```bash
sudo rm -rf /Library/Caches/com.apple.iconservices.store
sudo find /private/var/folders/ -name com.apple.iconservices -exec rm -rf {} + 2>/dev/null
killall Dock
```

### Icon looks blurry

Your source image is too small. Use at least 512x512, ideally 1024x1024.

### Icon looks wrong at small sizes

Simplify the design. Remove fine details, increase line thickness, and use higher contrast colors. Test by viewing the PNG at 32x32 pixels in your image editor.
