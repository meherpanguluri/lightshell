---
title: ".rpm Package"
description: Build an RPM package for Fedora and RHEL-based distributions.
---

An `.rpm` package is the native package format for Red Hat-based Linux distributions including Fedora, RHEL, CentOS Stream, Rocky Linux, AlmaLinux, and openSUSE. Users install it with `dnf` or `rpm`, and the system handles dependencies, menu integration, and clean uninstallation.

## Building

```bash
lightshell build --target rpm
```

Output:

```
✓ Built my-app in 1.4s → 5.2MB
✓ Created myapp binary
✓ Packaged RPM → dist/myapp-1.0.0-1.x86_64.rpm
```

The `.rpm` file appears in the `dist/` directory.

## Installing

Users can install the package using `dnf` or `rpm`:

```bash
# Using dnf (recommended — resolves dependencies automatically)
sudo dnf install ./myapp-1.0.0-1.x86_64.rpm

# Using rpm (manual dependency resolution)
sudo rpm -i myapp-1.0.0-1.x86_64.rpm
```

After installation:

- The binary is placed at `/usr/bin/myapp`
- A `.desktop` entry is created at `/usr/share/applications/myapp.desktop`
- The app icon is installed to `/usr/share/icons/hicolor/256x256/apps/myapp.png`
- The app appears in the GNOME or KDE application launcher

## Uninstalling

```bash
sudo dnf remove myapp
```

This removes the binary, desktop entry, and icon. User data in `~/.local/share/{appId}` is preserved.

## Package Contents

LightShell generates an RPM containing:

```
/usr/bin/myapp                                          # Application binary
/usr/share/applications/myapp.desktop                   # Desktop entry
/usr/share/icons/hicolor/256x256/apps/myapp.png         # App icon
```

## Spec File

LightShell generates an RPM spec file from your `lightshell.json` configuration. The spec file defines the package metadata, dependencies, and file list:

```
Name:       myapp
Version:    1.0.0
Release:    1
Summary:    A short description of your app
License:    MIT
URL:        https://example.com
Requires:   webkit2gtk4.1 >= 2.38, gtk3

%description
A short description of your app

%files
/usr/bin/myapp
/usr/share/applications/myapp.desktop
/usr/share/icons/hicolor/256x256/apps/myapp.png
```

LightShell uses `rpmbuild` if available on the build system, or falls back to `fpm` (Effing Package Management) as an alternative RPM builder.

## Dependencies

LightShell RPM packages declare two runtime dependencies:

- **webkit2gtk4.1** (>= 2.38) — the WebKitGTK webview
- **gtk3** — the GTK 3 toolkit

On Fedora and RHEL desktop installations, these are typically pre-installed. If missing, `dnf install` resolves them automatically.

## Configuration

The RPM metadata is derived from your `lightshell.json`:

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

| Field | RPM Spec Field |
|-------|---------------|
| `name` | `Name` (lowercased, hyphens removed) |
| `version` | `Version` |
| `description` | `Summary` and `%description` |
| `author` | Used in the changelog |
| `homepage` | `URL` |
| `build.icon` | Installed to hicolor icons |
| `build.appId` | Used in `.desktop` entry |

## Inspecting the Package

Verify the RPM contents before distributing:

```bash
# View package metadata
rpm -qpi dist/myapp-1.0.0-1.x86_64.rpm

# List all files in the package
rpm -qpl dist/myapp-1.0.0-1.x86_64.rpm

# Show dependencies
rpm -qpR dist/myapp-1.0.0-1.x86_64.rpm
```

Example `rpm -qpi` output:

```
Name        : myapp
Version     : 1.0.0
Release     : 1
Architecture: x86_64
Size        : 5242880
Summary     : A short description of your app
URL         : https://example.com
Description : A short description of your app
```

## Build Requirements

To build RPM packages, your build system needs one of:

- **rpmbuild** — the standard RPM build tool, part of the `rpm-build` package
- **fpm** — a multi-format package builder that can produce RPMs

On Fedora/RHEL:

```bash
sudo dnf install rpm-build
```

On Ubuntu (for cross-building):

```bash
sudo apt-get install rpm
```

## Architecture

LightShell builds for the architecture of the current machine:

- **x86_64** — standard AMD/Intel 64-bit systems
- **aarch64** — ARM 64-bit systems

The architecture is detected automatically and used in the package filename and spec file.

## CI Integration

Build RPM packages in GitHub Actions:

```yaml
jobs:
  build-rpm:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install build dependencies
        run: |
          sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev rpm
      - name: Build .rpm
        run: lightshell build --target rpm
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: linux-rpm
          path: dist/*.rpm
```

For native Fedora builds:

```yaml
jobs:
  build-rpm:
    runs-on: ubuntu-latest
    container: fedora:latest
    steps:
      - uses: actions/checkout@v4
      - name: Install build dependencies
        run: dnf install -y gtk3-devel webkit2gtk4.1-devel rpm-build
      - name: Build .rpm
        run: lightshell build --target rpm
```

## Hosting in a DNF Repository

For ongoing distribution, host your RPMs in a YUM/DNF repository so users can install and receive updates through the package manager. You can use Fedora COPR (Community Projects) for free hosting of RPM packages, or set up a self-hosted repository with `createrepo`.
