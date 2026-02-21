---
title: "4. Packaging"
description: Build your LightShell app into a distributable native binary.
---

Once your app is ready, `lightshell build` compiles it into a native binary that users can run without installing anything. This tutorial covers the build process, output formats, and size optimization.

## Building Your App

From your project directory, run:

```bash
lightshell build
```

LightShell reads your `lightshell.json`, bundles your `src/` files into a Go binary, and produces a platform-specific package in the `dist/` directory.

```
✓ Built my-app in 1.2s → 2.8MB
✓ Output: dist/MyApp.app
```

## macOS Output: .app Bundle

On macOS, `lightshell build` creates a standard `.app` bundle:

```
MyApp.app/
  Contents/
    Info.plist           # App metadata (generated from lightshell.json)
    MacOS/
      myapp              # The executable binary with embedded assets
    Resources/
      icon.icns          # App icon (converted from PNG if provided)
```

This is a real macOS application bundle. You can:
- Double-click it in Finder to launch
- Drag it to the Applications folder
- Set it as a login item
- Add it to the Dock

The `Info.plist` is generated from your `lightshell.json`:
- `name` becomes the bundle display name
- `build.appId` becomes the bundle identifier (e.g., `com.example.myapp`)
- `version` becomes the bundle version

## Linux Output: AppImage

On Linux, `lightshell build` creates an AppImage:

```
MyApp.AppImage           # Single executable file
```

Inside the AppImage:
- `AppRun` — the entry script
- `myapp` — the Go binary with embedded assets
- `myapp.desktop` — desktop entry file
- `icon.png` — the app icon

AppImages are self-contained. Users download the file, mark it executable (`chmod +x`), and run it. No installation step needed.

## Configuring the Build

The `build` section of `lightshell.json` controls packaging:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "My App",
    "width": 1024,
    "height": 768
  },
  "build": {
    "icon": "assets/icon.png",
    "appId": "com.example.myapp"
  }
}
```

### App Icon

Provide a PNG icon at the path specified in `build.icon`. Recommended size: 512x512 pixels or larger.

- On macOS, it is converted to `.icns` format automatically
- On Linux, it is included as-is in the AppImage

If no icon is provided, a default LightShell icon is used.

### App ID

The `build.appId` should be a reverse-domain identifier unique to your app. This is used for:
- macOS bundle identifier
- Linux desktop entry
- Per-app data directory paths

## Understanding Binary Size

A LightShell binary consists of:

| Component | Typical Size |
|-----------|-------------|
| Go runtime | ~1.5MB |
| Webview bindings | ~200KB |
| LightShell runtime code | ~300KB |
| Polyfills + normalize | ~3KB |
| **Your HTML/CSS/JS** | **varies** |

For a typical app with a few hundred lines of HTML, CSS, and JS, the total is around **2.8MB**.

## Size Optimization Tips

### Keep Assets Small

The biggest factor in binary size is your embedded assets. Tips:

- **Compress images** — use WebP or optimized PNG. Avoid uncompressed assets.
- **Minify CSS and JS** — use any standard minifier before building.
- **Avoid large libraries** — importing a 500KB charting library adds 500KB to your binary. Consider lightweight alternatives.
- **Use CDN for large libraries** — if your app is online-only, load large libraries from a CDN instead of bundling them.

### Fonts

Bundling web fonts adds to binary size (typically 50-200KB per font). The system font stack provided by LightShell's normalization layer looks great on both platforms without any font files.

If you must use a custom font, consider loading it from a CDN.

## Cross-Compilation

LightShell builds for the current platform by default. Cross-compilation support is available through CI:

```yaml
# Example GitHub Actions workflow
jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - run: lightshell build

  build-linux:
    runs-on: ubuntu-latest
    steps:
      - run: sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      - run: lightshell build
```

## Testing Your Build

After building, test the output:

```bash
# macOS
open dist/MyApp.app

# Linux
chmod +x dist/MyApp.AppImage
./dist/MyApp.AppImage
```

Verify that:
- The app launches correctly
- All assets load (images, fonts, etc.)
- Native API calls work (file dialogs, clipboard, etc.)
- The window title and size match your configuration

## Distribution

Once built, you can distribute your app through:

- **Direct download** — host the `.app` or `.AppImage` file on your website
- **GitHub Releases** — attach binaries to a release using GitHub Actions
- **Package managers** — create a Homebrew formula (macOS) or distribute via Flatpak (Linux)

LightShell does not currently handle code signing or notarization. For macOS distribution outside the App Store, users may need to right-click and select "Open" on first launch, or you can sign the app with your own Apple Developer certificate.
