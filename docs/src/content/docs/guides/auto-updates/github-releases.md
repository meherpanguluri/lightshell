---
title: GitHub Releases
description: Alternative -- publish updates to GitHub Releases instead of lightshell-server.
---

If you do not want to run a separate release server, you can publish updates directly to GitHub Releases. LightShell's updater can fetch manifests from GitHub and download binaries from release assets.

## When to Use GitHub Releases

**Choose GitHub Releases when:**
- Your project is already on GitHub
- You want zero infrastructure to manage
- You do not need download analytics or a release dashboard
- Your release cadence is moderate (not multiple times per day)

**Choose [lightshell-server](/docs/guides/auto-updates/release-server/) when:**
- You want a dashboard with download stats and audit logs
- You need to manage releases for multiple apps in one place
- You want automatic manifest generation from uploaded binaries
- You need fine-grained access control over release management

## Configuration

Point your updater endpoint to the GitHub API:

```json
{
  "updater": {
    "enabled": true,
    "endpoint": "https://api.github.com/repos/your-name/your-app/releases/latest",
    "provider": "github",
    "interval": "24h"
  }
}
```

The `provider: "github"` field tells the updater to parse the GitHub Releases API response format instead of the standard LightShell manifest format.

## Publishing with the CLI

Build your app and publish to GitHub Releases in one step:

```bash
lightshell release --github --repo your-name/your-app
```

This command:

1. Reads the version from `lightshell.json`
2. Creates a GitHub Release tagged `v{version}`
3. Uploads the build artifact as a release asset
4. Generates and uploads a `latest.json` manifest as a release asset
5. Signs the archive if a signing key is available

You need a `GITHUB_TOKEN` with `contents: write` permission:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
lightshell release --github --repo your-name/your-app
```

## GitHub Actions Workflow

Here is a complete workflow for publishing to GitHub Releases on every tag push:

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
          - os: macos-13
            platform: darwin-x64
          - os: ubuntu-latest
            platform: linux-x64

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

  release:
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

      - name: Create release and upload assets
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          LIGHTSHELL_SIGNING_KEY: ${{ secrets.SIGNING_PRIVATE_KEY }}
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}

          # Create the GitHub Release
          gh release create "v${VERSION}" \
            --title "v${VERSION}" \
            --generate-notes \
            artifacts/*.tar.gz

          # Generate and upload the update manifest
          lightshell release --github \
            --repo ${{ github.repository }} \
            --version "$VERSION" \
            --sign \
            --artifacts artifacts/
```

## Manifest Format

When using the GitHub provider, the updater reads the GitHub Release API response and looks for release assets with platform-specific naming:

```
my-app-darwin-arm64.tar.gz
my-app-darwin-x64.tar.gz
my-app-linux-x64.tar.gz
```

The naming convention is `{app-name}-{platform}.tar.gz` where platform is one of `darwin-arm64`, `darwin-x64`, or `linux-x64`.

Alternatively, you can attach a `latest.json` manifest file as a release asset for full control over the update metadata:

```json
{
  "version": "1.2.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-08-01T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://github.com/your-name/your-app/releases/download/v1.2.0/my-app-darwin-arm64.tar.gz",
      "sha256": "a1b2c3d4...",
      "signature": "base64-ed25519-signature..."
    },
    "darwin-x64": {
      "url": "https://github.com/your-name/your-app/releases/download/v1.2.0/my-app-darwin-x64.tar.gz",
      "sha256": "e5f6g7h8...",
      "signature": "base64-ed25519-signature..."
    },
    "linux-x64": {
      "url": "https://github.com/your-name/your-app/releases/download/v1.2.0/my-app-linux-x64.tar.gz",
      "sha256": "i9j0k1l2...",
      "signature": "base64-ed25519-signature..."
    }
  }
}
```

## Rate Limits

The GitHub API has rate limits that apply to unauthenticated requests:

- **Unauthenticated:** 60 requests per hour per IP
- **Authenticated:** 5,000 requests per hour

For most apps, the default `24h` check interval stays well within the unauthenticated limit. If you have many users on the same network (e.g., a corporate deployment), consider using a longer interval or switching to [lightshell-server](/docs/guides/auto-updates/release-server/).

## Private Repositories

For private repositories, users need a GitHub token to access release assets. This is not practical for distributed apps. If your repository is private, use one of these alternatives:

- Make the repository public (release assets are publicly downloadable)
- Use [lightshell-server](/docs/guides/auto-updates/release-server/) to host releases separately
- Use a public repository just for releases (keep your source code private in a separate repo)

## Limitations

Compared to lightshell-server, GitHub Releases has some limitations:

- **No dashboard** -- you can see releases on GitHub, but there is no unified view with download counts per platform
- **No audit log** -- GitHub provides release event logs but not per-download tracking
- **No download analytics** -- GitHub shows total download counts per asset, but not per-version trends or geographic data
- **Rate limits** -- the GitHub API rate limit can be an issue for apps with many simultaneous users
- **No automatic manifest** -- you need to generate and attach the `latest.json` yourself (the CLI does this for you)
