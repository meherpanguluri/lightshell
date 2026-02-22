---
title: CI/CD
description: Automate releases with GitHub Actions.
---

This guide provides a complete GitHub Actions workflow for building, signing, and uploading LightShell releases on every tag push. The workflow produces binaries for macOS (arm64 and x64) and Linux (x64), signs them with your Ed25519 key, and uploads them to your release server.

## Prerequisites

Before setting up CI/CD, make sure you have:

1. A release server running (see [Release Server](/docs/guides/auto-updates/release-server/)) or a GitHub Releases setup (see [GitHub Releases](/docs/guides/auto-updates/github-releases/))
2. An Ed25519 signing key pair (see [Signing Keys](/docs/guides/auto-updates/signing-keys/))
3. A GitHub repository with your LightShell app

## Secrets

Add these secrets to your GitHub repository (Settings > Secrets and variables > Actions):

| Secret | Description |
|--------|-------------|
| `SIGNING_PRIVATE_KEY` | Contents of `~/.lightshell/signing-key.pem`. The full PEM file contents. |
| `RELEASE_SERVER` | URL of your release server (e.g., `https://releases.example.com`). |
| `RELEASE_TOKEN` | Your `LIGHTSHELL_API_KEY` value for authenticating uploads. |

To add a secret:

```bash
# Copy your signing key to clipboard
cat ~/.lightshell/signing-key.pem | pbcopy   # macOS
cat ~/.lightshell/signing-key.pem | xclip    # Linux

# Then paste it into the GitHub secret form
```

## Complete Workflow

Create `.github/workflows/release.yml` in your repository:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: macos-latest
            platform: darwin-arm64
            target_arch: arm64
          - os: macos-13
            platform: darwin-x64
            target_arch: x64
          - os: ubuntu-latest
            platform: linux-x64
            target_arch: x64

    runs-on: ${{ matrix.os }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install LightShell
        run: npm install -g lightshell

      - name: Install Linux dependencies
        if: runner.os == 'Linux'
        run: |
          sudo apt-get update
          sudo apt-get install -y libwebkit2gtk-4.1-dev libgtk-3-dev

      - name: Build
        run: lightshell build

      - name: Package archive
        run: |
          cd dist
          tar -czf ../my-app-${{ matrix.platform }}.tar.gz .
          cd ..

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: build-${{ matrix.platform }}
          path: my-app-${{ matrix.platform }}.tar.gz

  sign-and-release:
    needs: build
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install LightShell
        run: npm install -g lightshell

      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts
          merge-multiple: true

      - name: Sign and upload releases
        env:
          LIGHTSHELL_SIGNING_KEY: ${{ secrets.SIGNING_PRIVATE_KEY }}
          LIGHTSHELL_RELEASE_SERVER: ${{ secrets.RELEASE_SERVER }}
          LIGHTSHELL_RELEASE_TOKEN: ${{ secrets.RELEASE_TOKEN }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}

          for archive in artifacts/*.tar.gz; do
            PLATFORM=$(echo "$archive" | sed 's/.*my-app-\(.*\)\.tar\.gz/\1/')

            lightshell release \
              --server "$LIGHTSHELL_RELEASE_SERVER" \
              --archive "$archive" \
              --platform "$PLATFORM" \
              --version "$VERSION" \
              --sign \
              --notes "Release v${VERSION}"
          done
```

## How It Works

The workflow runs in two stages:

### Stage 1: Build (parallel)

Three jobs run in parallel, one for each target platform:

- **macOS arm64** on `macos-latest` (Apple Silicon runners)
- **macOS x64** on `macos-13` (Intel runners)
- **Linux x64** on `ubuntu-latest`

Each job:
1. Checks out the code
2. Installs the LightShell CLI
3. Installs platform-specific dependencies (Linux only: WebKitGTK and GTK)
4. Runs `lightshell build` to produce the binary
5. Packages the output into a `.tar.gz` archive
6. Uploads the archive as a GitHub Actions artifact

### Stage 2: Sign and Release (sequential)

After all builds complete:
1. Downloads all build artifacts
2. For each archive, runs `lightshell release` which:
   - Computes the SHA256 hash
   - Signs the archive with the Ed25519 private key
   - Uploads the archive and metadata to the release server

## Triggering a Release

Create a git tag and push it:

```bash
# Update version in lightshell.json first
# Then tag and push
git tag v1.2.0
git push origin v1.2.0
```

The workflow triggers on any tag matching `v*`. The version number is extracted from the tag name (stripping the `v` prefix).

## Release Notes

The workflow uses a generic release note (`Release v1.2.0`). For richer release notes, you have two options:

### From a CHANGELOG

```yaml
- name: Extract release notes
  id: notes
  run: |
    # Extract notes for this version from CHANGELOG.md
    VERSION=${GITHUB_REF#refs/tags/v}
    NOTES=$(sed -n "/^## ${VERSION}/,/^## /p" CHANGELOG.md | head -n -1)
    echo "notes<<EOF" >> $GITHUB_OUTPUT
    echo "$NOTES" >> $GITHUB_OUTPUT
    echo "EOF" >> $GITHUB_OUTPUT

- name: Sign and upload
  run: |
    lightshell release \
      --notes "${{ steps.notes.outputs.notes }}" \
      # ... rest of args
```

### From the Git Tag Message

```bash
# Create an annotated tag with release notes
git tag -a v1.2.0 -m "Bug fixes and performance improvements

- Fixed crash on startup with large files
- Improved search performance by 3x
- Updated dependencies"

git push origin v1.2.0
```

```yaml
- name: Get tag message
  id: tag
  run: |
    NOTES=$(git tag -l --format='%(contents)' ${GITHUB_REF#refs/tags/})
    echo "notes<<EOF" >> $GITHUB_OUTPUT
    echo "$NOTES" >> $GITHUB_OUTPUT
    echo "EOF" >> $GITHUB_OUTPUT
```

## Customizing the Workflow

### Adding macOS Code Signing

If you have an Apple Developer ID, add code signing to the macOS build step:

```yaml
- name: Build (signed)
  if: runner.os == 'macOS'
  env:
    APPLE_CERTIFICATE: ${{ secrets.APPLE_CERTIFICATE }}
    APPLE_CERTIFICATE_PASSWORD: ${{ secrets.APPLE_CERTIFICATE_PASSWORD }}
    APPLE_IDENTITY: ${{ secrets.APPLE_SIGNING_IDENTITY }}
  run: |
    # Import certificate
    echo "$APPLE_CERTIFICATE" | base64 -d > cert.p12
    security create-keychain -p "" build.keychain
    security import cert.p12 -k build.keychain -P "$APPLE_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
    security set-key-partition-list -S apple-tool:,apple: -s -k "" build.keychain
    security default-keychain -s build.keychain

    # Build with signing
    lightshell build --sign
```

### Custom Build Arguments

Pass additional build arguments for specific platforms:

```yaml
- name: Build
  run: |
    if [ "${{ matrix.platform }}" = "linux-x64" ]; then
      lightshell build --target appimage
    else
      lightshell build
    fi
```

## Verifying the Pipeline

After your first automated release, verify everything is correct:

1. Check your release server dashboard for the new version
2. Run `lightshell updater.check()` in a dev build of the previous version
3. Confirm the update downloads and installs successfully
4. Check the SHA256 hash matches: `shasum -a 256 downloaded-archive.tar.gz`
