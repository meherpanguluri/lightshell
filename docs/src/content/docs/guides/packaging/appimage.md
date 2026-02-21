---
title: AppImage
description: Build a portable AppImage for Linux distribution.
---

An AppImage is a single executable file that runs on any Linux distribution without installation. Users download it, mark it executable, and run it. No root access, no package manager, no dependencies to install manually. LightShell produces an AppImage as the default build output on Linux.

## Building

```bash
lightshell build
```

Output:

```
✓ Built my-app in 1.4s → 5.2MB
✓ Output: dist/MyApp.AppImage
```

This is the same as `lightshell build --target appimage` — AppImage is the default on Linux.

## Running an AppImage

Users run your AppImage in two steps:

```bash
# Make it executable
chmod +x MyApp.AppImage

# Run it
./MyApp.AppImage
```

That is the entire process. No installation, no extraction, no dependency resolution.

## How It Works

An AppImage is a self-mounting filesystem image. When executed, it mounts itself using FUSE (Filesystem in Userspace), runs the embedded application, and unmounts when the app exits. The user never sees the mount/unmount process.

## AppImage Structure

Inside the AppImage, LightShell packages these files:

```
MyApp.AppImage (self-extracting archive)
  AppRun                    # Entry script that launches the binary
  myapp                     # The Go binary with embedded web assets
  myapp.desktop             # Desktop entry file for system integration
  icon.png                  # App icon
```

- **AppRun** is the entry point that the AppImage runtime calls. It sets up the environment and launches your binary.
- **myapp** is the compiled binary containing the Go runtime, LightShell runtime, and all your HTML, CSS, and JS files embedded via `embed.FS`.
- **myapp.desktop** is a standard FreeDesktop `.desktop` entry used for system integration.
- **icon.png** is your app icon, included as-is from the path specified in `lightshell.json`.

## System Requirements

AppImages require the following on the user's system:

- **WebKitGTK 2.40+** — the system webview used by LightShell
- **GTK 3** — the underlying toolkit
- **FUSE** — for mounting the AppImage (installed by default on most distributions)

These are the only runtime dependencies. WebKitGTK and GTK 3 are pre-installed on most desktop Linux distributions. If a user's system is missing WebKitGTK, they will see an error at launch with instructions to install it.

### Distribution Compatibility

WebKitGTK 2.40+ is available on:

- Ubuntu 23.10 and later
- Fedora 38 and later
- Arch Linux (always current)
- Debian 13 (trixie) and later
- Linux Mint 22 and later
- openSUSE Tumbleweed

For older distributions, users may need to install a newer version of WebKitGTK from a PPA or alternative repository.

## Desktop Integration

By default, an AppImage is just a file on disk — it does not appear in the application launcher or menus. Users can integrate it in several ways.

### Manual Integration

Users can copy the AppImage to a standard location and create a `.desktop` file:

```bash
# Move to a standard location
mkdir -p ~/.local/bin
mv MyApp.AppImage ~/.local/bin/

# Create a desktop entry
cat > ~/.local/share/applications/myapp.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=My App
Exec=~/.local/bin/MyApp.AppImage
Icon=myapp
Categories=Utility;
EOF
```

### Automatic Integration with appimaged

The [appimaged](https://github.com/probonopd/go-appimage) daemon watches directories like `~/Applications` and `~/Downloads` for AppImage files and automatically creates menu entries for them:

```bash
# Move AppImage to watched directory
mkdir -p ~/Applications
mv MyApp.AppImage ~/Applications/
```

If appimaged is running, the app appears in the application launcher within seconds.

## Distribution

AppImages are well-suited for several distribution channels:

### GitHub Releases

Attach the AppImage to a GitHub Release. Users download and run it directly:

```yaml
jobs:
  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install dependencies
        run: sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
      - name: Build
        run: lightshell build
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: linux-appimage
          path: dist/*.AppImage
```

### Direct Download

Host the AppImage on your website. Include instructions for users:

```
1. Download MyApp.AppImage
2. Open a terminal in the download directory
3. Run: chmod +x MyApp.AppImage && ./MyApp.AppImage
```

### AppImageHub

Submit your AppImage to [AppImageHub](https://appimage.github.io/) for wider discoverability among Linux users who prefer the AppImage format.

## Size

A typical LightShell AppImage is approximately 5MB. This includes:

| Component | Size |
|-----------|------|
| Go runtime | ~1.5MB |
| Webview bindings | ~200KB |
| LightShell runtime | ~300KB |
| AppImage runtime | ~100KB |
| Your app code | varies |

The binary is stripped of debug symbols to minimize size.

## Comparison with .deb and .rpm

| | AppImage | .deb | .rpm |
|-|----------|------|------|
| Installation | None (just run) | `apt install` | `dnf install` |
| Root required | No | Yes | Yes |
| Auto-updates | Via LightShell updater | Via apt | Via dnf |
| System integration | Manual or appimaged | Automatic | Automatic |
| Uninstall | Delete the file | `apt remove` | `dnf remove` |
| Works offline | Yes | Needs dpkg | Needs rpm |

Choose AppImage when you want maximum portability and zero friction. Choose `.deb` or `.rpm` when your users expect system package manager integration.
