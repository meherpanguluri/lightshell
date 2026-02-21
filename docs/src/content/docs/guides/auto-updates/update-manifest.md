---
title: Update Manifest
description: The JSON manifest format for LightShell auto-updates.
---

The update manifest is a JSON file that tells LightShell what the latest version of your app is, where to download it, and how to verify the download. You host this file at a URL that your app checks periodically.

## Full Example

```json
{
  "version": "1.2.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-07-15T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.example.com/v1.2.0/myapp-darwin-arm64.tar.gz",
      "sha256": "a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678"
    },
    "darwin-x64": {
      "url": "https://releases.example.com/v1.2.0/myapp-darwin-x64.tar.gz",
      "sha256": "b4e3c9d2e5f67890abcdef1234567890abcdef1234567890abcdef12345679"
    },
    "linux-x64": {
      "url": "https://releases.example.com/v1.2.0/myapp-linux-x64.tar.gz",
      "sha256": "c5f4d0e3f6a78901bcdef01234567890abcdef1234567890abcdef1234567a"
    },
    "linux-arm64": {
      "url": "https://releases.example.com/v1.2.0/myapp-linux-arm64.tar.gz",
      "sha256": "d6a5e1f4a7b89012cdef01234567890abcdef1234567890abcdef1234567b"
    }
  }
}
```

## Field Reference

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | The version of the latest release. Must be valid semver (e.g., `"1.2.0"`, `"2.0.0-beta.1"`). |
| `notes` | string | Yes | Release notes shown to the user. Plain text or a short description. |
| `pub_date` | string | Yes | ISO 8601 timestamp of the release. Used for display purposes only. |
| `platforms` | object | Yes | Map of platform keys to download info. |

### Platform Keys

Each key in the `platforms` object identifies a target platform using the format `{os}-{arch}`:

| Key | OS | Architecture |
|-----|----|-------------|
| `darwin-arm64` | macOS | Apple Silicon (M1, M2, M3, M4) |
| `darwin-x64` | macOS | Intel |
| `linux-x64` | Linux | x86_64 / AMD64 |
| `linux-arm64` | Linux | ARM64 / AArch64 |

You only need to include the platforms you support. If a user runs your app on a platform not listed in the manifest, `lightshell.updater.check()` returns `null` (no update available for their platform).

### Platform Object Fields

Each platform entry contains:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | Direct download URL for the archive. Must be HTTPS in production builds. |
| `sha256` | string | Yes | SHA256 hash of the archive file. Used to verify the download. |

## Version Comparison

LightShell compares the manifest `version` against the app's current version (from `lightshell.json` at build time) using semantic versioning rules:

- `1.2.0` > `1.1.0` -- update available
- `1.1.0` = `1.1.0` -- no update
- `1.0.0` < `1.1.0` -- no update (manifest is older)
- `2.0.0-beta.1` > `1.9.0` -- update available
- `1.1.0-rc.1` < `1.1.0` -- no update (pre-release is less than release)

The comparison follows the semver 2.0.0 specification. Pre-release versions have lower precedence than the associated release version.

## SHA256 Hash

The SHA256 hash is mandatory. It protects against corrupted downloads and tampering. Generate it with:

```bash
# macOS or Linux
shasum -a 256 myapp-darwin-arm64.tar.gz
# Output: a3f2b8c1d4e5f67890abcdef... myapp-darwin-arm64.tar.gz

# Or with openssl
openssl dgst -sha256 myapp-darwin-arm64.tar.gz
```

Copy the hex string (64 characters) into the manifest. If the hash in the manifest does not match the downloaded file, the update is rejected.

## Archive Format

The download URL must point to a `.tar.gz` archive containing the app binary:

**macOS:** The archive should contain the `.app` bundle:

```
myapp-darwin-arm64.tar.gz
  └── MyApp.app/
      └── Contents/
          ├── MacOS/
          │   └── myapp          (the binary)
          ├── Resources/
          │   └── icon.icns
          └── Info.plist
```

Create it with:

```bash
tar -czf myapp-darwin-arm64.tar.gz -C dist MyApp.app
```

**Linux:** The archive should contain the AppImage or raw binary:

```
myapp-linux-x64.tar.gz
  └── MyApp.AppImage
```

Create it with:

```bash
tar -czf myapp-linux-x64.tar.gz -C dist MyApp.AppImage
```

## URL Requirements

- **Production builds:** URLs must use HTTPS. If the manifest contains an HTTP URL, the updater rejects it and emits an error.
- **Dev mode** (`lightshell dev`): HTTP URLs are allowed for local testing.
- **Redirects:** The updater follows HTTP redirects (e.g., GitHub Releases redirect through a CDN). The final URL must be HTTPS.

## Minimal Manifest

If you only support one platform, the manifest can be quite small:

```json
{
  "version": "1.0.1",
  "notes": "Fixed a crash on startup",
  "pub_date": "2025-08-15T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://releases.example.com/v1.0.1/myapp-darwin-arm64.tar.gz",
      "sha256": "a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678"
    }
  }
}
```

## Generating the Manifest

Here is a shell script that builds the manifest from your release artifacts:

```bash
#!/bin/bash
set -euo pipefail

VERSION="$1"
BASE_URL="https://releases.example.com/v${VERSION}"

manifest="{\"version\":\"${VERSION}\",\"notes\":\"Release ${VERSION}\",\"pub_date\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"platforms\":{"

first=true
for file in myapp-*.tar.gz; do
  platform=$(echo "$file" | sed 's/myapp-\(.*\)\.tar\.gz/\1/')
  sha=$(shasum -a 256 "$file" | cut -d' ' -f1)

  if [ "$first" = true ]; then
    first=false
  else
    manifest+=","
  fi

  manifest+="\"${platform}\":{\"url\":\"${BASE_URL}/${file}\",\"sha256\":\"${sha}\"}"
done

manifest+="}}"

echo "$manifest" | python3 -m json.tool > latest.json
echo "Generated latest.json for version ${VERSION}"
```

Usage:

```bash
./generate-manifest.sh 1.2.0
```
