---
title: ".deb Package"
description: Build a Debian package for Ubuntu and Debian-based distributions.
---

A `.deb` package is the native package format for Debian-based Linux distributions including Ubuntu, Linux Mint, Pop!_OS, and Debian itself. Users install it with `apt` or `dpkg`, and the system handles dependencies, menu integration, and clean uninstallation automatically.

## Building

```bash
lightshell build --target deb
```

Output:

```
✓ Built my-app in 1.4s → 5.2MB
✓ Created myapp binary
✓ Packaged .deb → dist/myapp_1.0.0_amd64.deb
```

The `.deb` file appears in the `dist/` directory.

## Installing

Users can install the package using either `apt` or `dpkg`:

```bash
# Using apt (recommended — resolves dependencies automatically)
sudo apt install ./myapp_1.0.0_amd64.deb

# Using dpkg (manual dependency resolution)
sudo dpkg -i myapp_1.0.0_amd64.deb
sudo apt-get install -f   # Install missing dependencies if needed
```

After installation:

- The binary is placed at `/usr/bin/myapp`
- A `.desktop` entry is created at `/usr/share/applications/myapp.desktop`
- The app icon is installed to `/usr/share/icons/hicolor/256x256/apps/myapp.png`
- The app appears in the system application launcher

## Uninstalling

```bash
sudo apt remove myapp
```

This removes the binary, desktop entry, and icon. User data stored in `~/.local/share/{appId}` is preserved.

## Package Structure

LightShell generates the following `.deb` package structure:

```
myapp_1.0.0_amd64.deb
  DEBIAN/
    control                                    # Package metadata
  usr/
    bin/
      myapp                                    # Application binary
    share/
      applications/
        myapp.desktop                          # Desktop entry for app launcher
      icons/
        hicolor/
          256x256/
            apps/
              myapp.png                        # App icon
```

### The Control File

The `DEBIAN/control` file contains package metadata, generated from your `lightshell.json`:

```
Package: myapp
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Depends: libwebkit2gtk-4.1-0 (>= 2.38), libgtk-3-0
Maintainer: Your Name <you@example.com>
Description: A description of your application
Homepage: https://example.com
```

### The Desktop Entry

The `.desktop` file registers your app with the system's application launcher:

```ini
[Desktop Entry]
Type=Application
Name=My App
Exec=/usr/bin/myapp
Icon=myapp
Categories=Utility;
Comment=A description of your application
```

## Dependencies

LightShell `.deb` packages declare two runtime dependencies:

- **libwebkit2gtk-4.1-0** (>= 2.38) — the WebKitGTK webview
- **libgtk-3-0** — the GTK 3 toolkit

These are pre-installed on most Ubuntu and Debian desktop installations. If they are missing, `apt install` resolves them automatically from the system's configured repositories.

## Configuration

The `.deb` package metadata comes from your `lightshell.json`:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "description": "A short description of your app",
  "author": "Your Name",
  "homepage": "https://example.com",
  "build": {
    "icon": "assets/icon.png",
    "appId": "com.example.myapp"
  }
}
```

| Field | Used For |
|-------|----------|
| `name` | Package name (lowercased, hyphens converted to underscores) |
| `version` | Package version |
| `description` | Package description and desktop entry comment |
| `author` | Maintainer field |
| `homepage` | Homepage URL in control file |
| `build.icon` | App icon installed to hicolor icons directory |
| `build.appId` | Desktop entry and data directory name |

## Inspecting the Package

Before distributing, verify the package contents:

```bash
# View package metadata
dpkg-deb --info dist/myapp_1.0.0_amd64.deb

# List all files in the package
dpkg -c dist/myapp_1.0.0_amd64.deb

# Extract without installing (for inspection)
dpkg-deb -x dist/myapp_1.0.0_amd64.deb /tmp/myapp-extracted/
```

Example `dpkg-deb --info` output:

```
Package: myapp
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Depends: libwebkit2gtk-4.1-0 (>= 2.38), libgtk-3-0
Installed-Size: 5200
Maintainer: Your Name <you@example.com>
Description: A short description of your app
```

## Hosting in an APT Repository

For ongoing distribution, you can host your `.deb` packages in an APT repository so users can install and update via `apt`:

```bash
# Users add your repository once
sudo add-apt-repository ppa:yourname/myapp
sudo apt update

# Then install
sudo apt install myapp

# Future updates come through apt
sudo apt upgrade
```

Setting up a PPA (Personal Package Archive) on Launchpad or a self-hosted APT repository is beyond the scope of this guide, but the `.deb` files LightShell produces are fully compatible with any APT repository.

## Architecture

LightShell builds for the architecture of the current machine:

- **amd64** — standard x86_64 systems
- **arm64** — ARM-based systems (Raspberry Pi 4+, cloud ARM instances)

The architecture is automatically detected and set in the package filename and control file.

## CI Integration

Build `.deb` packages in GitHub Actions:

```yaml
jobs:
  build-deb:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install build dependencies
        run: sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      - name: Build .deb
        run: lightshell build --target deb
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: linux-deb
          path: dist/*.deb
```
