---
title: macOS App Bundle
description: Build and distribute a .app bundle for macOS.
---

On macOS, `lightshell build` produces a `.app` bundle — the standard application format that macOS users expect. A `.app` bundle is actually a directory with a specific structure, but Finder displays it as a single file that users can double-click to launch.

## Building

From your project directory:

```bash
lightshell build
```

Output:

```
✓ Built my-app in 1.2s → 5.1MB
✓ Output: dist/MyApp.app
```

The `.app` bundle appears in the `dist/` directory. You can open it immediately:

```bash
open dist/MyApp.app
```

## Bundle Structure

Inside the `.app` bundle, LightShell creates the standard macOS directory layout:

```
MyApp.app/
  Contents/
    Info.plist              # App metadata (name, version, identifier)
    MacOS/
      myapp                 # The executable binary with embedded assets
    Resources/
      icon.icns             # App icon (converted from your PNG)
```

- **Info.plist** contains metadata that macOS uses to display your app in Finder, Spotlight, and the Dock. LightShell generates this from your `lightshell.json`.
- **MacOS/myapp** is the compiled Go binary with your HTML, CSS, and JS files embedded inside it. This is the only executable in the bundle.
- **Resources/icon.icns** is your app icon converted to the macOS `.icns` format, which contains the icon at multiple resolutions for different display contexts.

## Info.plist Fields

LightShell generates `Info.plist` from your `lightshell.json` configuration:

| lightshell.json field | Info.plist key | Example |
|----------------------|----------------|---------|
| `name` | `CFBundleName` | `"My App"` |
| `version` | `CFBundleShortVersionString` | `"1.0.0"` |
| `build.appId` | `CFBundleIdentifier` | `"com.example.myapp"` |
| `build.icon` | `CFBundleIconFile` | `"icon"` |

Additional fields are set automatically:

- `CFBundleExecutable` — the binary name, derived from your app name
- `CFBundlePackageType` — always `"APPL"`
- `CFBundleInfoDictionaryVersion` — always `"6.0"`
- `NSHighResolutionCapable` — always `true` for Retina display support
- `NSSupportsAutomaticGraphicsSwitching` — always `true`

## Testing Your Build

After building, verify the bundle works correctly:

```bash
# Launch the app
open dist/MyApp.app

# Or double-click MyApp.app in Finder
```

Check the following:

- The app launches and displays your UI
- The window title matches your configuration
- The app icon appears in the Dock
- All native API calls work (file dialogs, clipboard, notifications)
- The app appears with its correct name in the menu bar

## Installing the App

Users can install a `.app` bundle by dragging it to the Applications folder:

1. Open the `dist/` directory in Finder
2. Drag `MyApp.app` to `/Applications`
3. Launch from Spotlight or the Applications folder

The app will also appear in Launchpad once it is in the Applications folder.

## Gatekeeper and Unsigned Apps

macOS Gatekeeper blocks apps from unidentified developers by default. If your app is not code-signed, users will see a warning when they first try to open it:

> "MyApp" can't be opened because Apple cannot check it for malicious software.

To open an unsigned app:

1. Right-click (or Control-click) the app in Finder
2. Select "Open" from the context menu
3. Click "Open" in the dialog that appears

This only needs to be done once. After that, the app opens normally.

For distribution to a wider audience, code signing eliminates this friction entirely. See the [Code Signing](/docs/guides/packaging/code-signing/) guide.

## Customizing the Build

### App Icon

Provide a 512x512 or 1024x1024 PNG image:

```json
{
  "build": {
    "icon": "assets/icon.png"
  }
}
```

LightShell converts it to `.icns` format with multiple resolutions (16x16 through 512x512 at 1x and 2x). See the [App Icons](/docs/guides/packaging/icons/) guide for details.

### App Identifier

Set a reverse-domain identifier for your app:

```json
{
  "build": {
    "appId": "com.yourname.myapp"
  }
}
```

This is used as the macOS bundle identifier. Choose it carefully before your first release — changing it later creates a new identity from the OS perspective. See the [App Identifier](/docs/guides/packaging/app-id/) guide.

## Binary Size

A typical LightShell `.app` bundle is approximately 5MB. The breakdown:

| Component | Size |
|-----------|------|
| Go runtime | ~1.5MB |
| Webview bindings | ~200KB |
| LightShell runtime | ~300KB |
| Polyfills + normalize | ~3KB |
| Your app code | varies |
| bbolt + HTTP client | ~800KB |

The binary is stripped (`-ldflags="-s -w"`) to remove debug symbols.

## Distribution

A `.app` bundle can be distributed by:

- Compressing it into a `.zip` file and sharing directly
- Uploading to GitHub Releases
- Packaging it in a [DMG](/docs/guides/packaging/dmg/) for a professional drag-to-install experience
- Creating a Homebrew Cask formula

For the best user experience when distributing publicly, use a signed DMG. See [DMG Installer](/docs/guides/packaging/dmg/) and [Code Signing](/docs/guides/packaging/code-signing/).
