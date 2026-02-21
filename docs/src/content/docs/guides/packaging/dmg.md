---
title: DMG Installer
description: Create a drag-to-install DMG disk image for macOS distribution.
---

A DMG (Disk Image) is the standard macOS distribution format. When users open it, they see your app alongside an Applications folder shortcut and simply drag to install. LightShell automates the entire DMG creation process.

## Building a DMG

```bash
lightshell build --target dmg
```

Output:

```
✓ Built my-app in 1.2s → 5.1MB
✓ Created MyApp.app
✓ Packaged DMG → dist/MyApp-1.0.0.dmg
```

The DMG file appears in `dist/` and is ready to distribute.

## What Users See

When a user opens `MyApp-1.0.0.dmg`, macOS mounts it as a virtual disk and displays a Finder window with:

- **Your app** (MyApp.app) on the left
- **Applications folder shortcut** on the right

The user drags the app to Applications. That is the entire installation process. After installing, they eject the DMG and optionally delete it.

## DMG Contents

Inside the DMG, LightShell creates:

```
MyApp-1.0.0.dmg (mounted volume)
  MyApp.app              # Your application bundle
  Applications           # Symlink to /Applications
```

The volume name matches your app name from `lightshell.json`. The DMG is compressed as a read-only disk image to minimize download size.

## How It Works

Under the hood, `lightshell build --target dmg` performs these steps:

1. Builds the `.app` bundle (same as `lightshell build`)
2. Creates a temporary writable DMG using `hdiutil create`
3. Copies the `.app` bundle into the mounted volume
4. Creates a symbolic link to `/Applications`
5. Sets the volume name and icon
6. Configures the Finder window layout (icon positions and window size)
7. Converts to a compressed, read-only DMG using `hdiutil convert`

The final output is a compressed DMG that is typically the same size as the `.app` bundle itself, since the binary is already stripped and optimized.

## Combining with Code Signing

For distribution outside of direct file sharing, you should sign your app. Signing eliminates the Gatekeeper "unidentified developer" warning:

```bash
lightshell build --target dmg --sign
```

This signs the `.app` bundle before packaging it into the DMG. You must have a Developer ID certificate configured in your `lightshell.json`:

```json
{
  "build": {
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)"
    }
  }
}
```

See [Code Signing](/guides/packaging/code-signing/) for the full setup process.

## Adding Notarization

For the smoothest experience, notarize the DMG after signing. Notarization tells macOS that Apple has scanned your app and found no malicious content:

```bash
lightshell build --target dmg --sign --notarize
```

This submits the DMG to Apple's notarization service, waits for approval, and staples the notarization ticket to the DMG. Users will see no security warnings at all when opening your app.

Notarization requires your Apple ID and Team ID in the configuration. See [Code Signing](/guides/packaging/code-signing/) for details.

## Typical Distribution Workflow

A complete release workflow looks like this:

```bash
# 1. Build and sign the DMG
lightshell build --target dmg --sign

# 2. (Optional) Notarize for zero-warning distribution
lightshell build --target dmg --sign --notarize

# 3. Upload to your distribution channel
#    - GitHub Releases
#    - Your website's download page
#    - A CDN
```

For CI automation with GitHub Actions:

```yaml
jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build DMG
        run: lightshell build --target dmg --sign
        env:
          # Store signing identity in GitHub Secrets
          APPLE_IDENTITY: ${{ secrets.APPLE_IDENTITY }}
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: macos-dmg
          path: dist/*.dmg
```

## DMG Configuration

The DMG is configured through your `lightshell.json`:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "build": {
    "appId": "com.example.myapp",
    "icon": "assets/icon.png"
  }
}
```

- **Volume name** is derived from the app `name`
- **Filename** follows the pattern `{AppName}-{version}.dmg`
- **Volume icon** uses the same icon as the app

## Verifying the DMG

After building, verify the DMG before distributing:

```bash
# Mount and inspect
hdiutil attach dist/MyApp-1.0.0.dmg

# Check that the app launches from the mounted volume
open /Volumes/MyApp/MyApp.app

# Verify code signature (if signed)
codesign --verify --deep --strict /Volumes/MyApp/MyApp.app

# Unmount
hdiutil detach /Volumes/MyApp
```

## Size

A DMG is typically the same size as the `.app` bundle it contains (approximately 5MB for a basic app). The DMG format uses built-in compression, so the download size closely matches the raw binary size.
