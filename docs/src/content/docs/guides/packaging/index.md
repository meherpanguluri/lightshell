---
title: Packaging Overview
description: Choose the right packaging format for your LightShell app.
---

LightShell can package your app into several formats depending on your platform and distribution needs. By default, `lightshell build` produces the native default for your OS — a `.app` bundle on macOS or an AppImage on Linux. Use the `--target` flag to produce other formats.

## Available Formats

| Format | Platform | Use Case | Typical Size |
|--------|----------|----------|-------------|
| `.app` bundle | macOS | Development, direct distribution | ~5MB |
| DMG | macOS | Drag-to-install distribution | ~5MB |
| AppImage | Linux | Portable, no-install distribution | ~5MB |
| `.deb` | Debian/Ubuntu | System package manager install | ~5MB |
| `.rpm` | Fedora/RHEL | System package manager install | ~5MB |

All formats embed your HTML, CSS, and JS into a single native binary. The size differences between formats are minimal — the packaging overhead is small compared to the binary itself.

## Quick Start

Run `lightshell build` with no flags to get the default format for your platform:

```bash
lightshell build
```

On macOS, this produces `dist/MyApp.app`. On Linux, this produces `dist/MyApp.AppImage`.

## All Build Commands

```bash
# Default format for current OS
lightshell build

# macOS formats
lightshell build --target app          # .app bundle (same as default on macOS)
lightshell build --target dmg          # DMG disk image with drag-to-install
lightshell build --target dmg --sign   # Signed DMG
lightshell build --target dmg --sign --notarize  # Signed + notarized DMG

# Linux formats
lightshell build --target appimage     # AppImage (same as default on Linux)
lightshell build --target deb          # Debian package
lightshell build --target rpm          # RPM package

# Build all formats for current OS
lightshell build --target all

# Include DevTools in production build (for debugging)
lightshell build --devtools
```

You can only build formats for your current platform. To build for both macOS and Linux, use CI with runners for each OS.

## Choosing a Format

**For macOS distribution:**

- Use `.app` during development and for sharing with individuals. Users drag it to Applications or double-click to run.
- Use DMG for public distribution. It gives users a clean drag-to-install experience with an Applications shortcut.
- Add code signing to avoid the "unidentified developer" warning from Gatekeeper.
- Add notarization for the smoothest user experience — no security warnings at all.

**For Linux distribution:**

- Use AppImage for maximum compatibility. It works on any distribution with WebKitGTK installed, requires no root access, and needs no installation step.
- Use `.deb` for Debian-based distributions (Ubuntu, Debian, Linux Mint, Pop!_OS). Users install via `apt` or `dpkg` and get automatic dependency resolution.
- Use `.rpm` for Red Hat-based distributions (Fedora, RHEL, CentOS Stream, openSUSE). Users install via `dnf` or `rpm`.

## Configuration

All packaging formats read from the same `lightshell.json` configuration:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "build": {
    "icon": "assets/icon.png",
    "appId": "com.example.myapp"
  }
}
```

The `name` determines the output filename and display name. The `version` is embedded in the package metadata. The `build.icon` and `build.appId` are used across all formats for icons and identifiers.

For format-specific configuration like code signing, see the individual format guides.

## Detailed Guides

- [macOS App Bundle](/docs/guides/packaging/app-bundle/) — the default macOS format
- [DMG Installer](/docs/guides/packaging/dmg/) — drag-to-install disk image
- [AppImage](/docs/guides/packaging/appimage/) — portable Linux executable
- [.deb Package](/docs/guides/packaging/deb/) — for Debian and Ubuntu
- [.rpm Package](/docs/guides/packaging/rpm/) — for Fedora and RHEL
- [Code Signing](/docs/guides/packaging/code-signing/) — sign and notarize for macOS
- [App Icons](/docs/guides/packaging/icons/) — create and configure icons
- [App Identifier](/docs/guides/packaging/app-id/) — choose a unique app ID
