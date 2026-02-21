---
title: Hosting Releases
description: Host your update files and manifest for LightShell auto-updates.
---

The LightShell updater needs two things from your server: a `latest.json` manifest and the binary archives it points to. You can host these anywhere that serves static files over HTTPS. This guide covers the most common options.

## Requirements

Your hosting must:

- Serve files over HTTPS (required in production builds; HTTP allowed in dev only)
- Return correct `Content-Type` headers (though not strictly required, it avoids edge cases)
- Be publicly accessible (or accessible to your users' networks)

CORS headers are not needed. The Go backend fetches the manifest and downloads, not the browser, so CORS does not apply.

## Option 1: GitHub Releases

The simplest option for open-source and small commercial apps. Free, reliable, and already integrated with your build workflow.

### Manual Setup

1. Build your app for each platform:

```bash
lightshell build                     # builds for current platform
```

2. Compress the output:

```bash
# macOS arm64
tar -czf my-app-darwin-arm64.tar.gz -C dist MyApp.app

# Linux x64
tar -czf my-app-linux-x64.tar.gz -C dist MyApp.AppImage
```

3. Generate SHA256 hashes:

```bash
shasum -a 256 my-app-darwin-arm64.tar.gz
# a3f2b8c1d4e5... my-app-darwin-arm64.tar.gz

shasum -a 256 my-app-linux-x64.tar.gz
# c5f4d0e3f6a7... my-app-linux-x64.tar.gz
```

4. Create a `latest.json`:

```json
{
  "version": "1.1.0",
  "notes": "Bug fixes and performance improvements",
  "pub_date": "2025-08-01T00:00:00Z",
  "platforms": {
    "darwin-arm64": {
      "url": "https://github.com/youruser/myapp/releases/download/v1.1.0/my-app-darwin-arm64.tar.gz",
      "sha256": "a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678"
    },
    "linux-x64": {
      "url": "https://github.com/youruser/myapp/releases/download/v1.1.0/my-app-linux-x64.tar.gz",
      "sha256": "c5f4d0e3f6a78901bcdef01234567890abcdef1234567890abcdef1234567a"
    }
  }
}
```

5. Create a GitHub Release for the tag `v1.1.0` and upload the `.tar.gz` files and `latest.json` as release assets.

6. Set your endpoint to the raw asset URL:

```json
{
  "updater": {
    "endpoint": "https://github.com/youruser/myapp/releases/latest/download/latest.json"
  }
}
```

The `/releases/latest/download/` URL always redirects to the most recent release's assets, so you do not need to update the endpoint URL when you publish new versions.

### Automated with GitHub Actions

Create `.github/workflows/release.yml`:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install LightShell
        run: curl -fsSL https://lightshell.dev/install.sh | sh

      - name: Build
        run: lightshell build

      - name: Package
        run: |
          tar -czf my-app-darwin-arm64.tar.gz -C dist MyApp.app
          shasum -a 256 my-app-darwin-arm64.tar.gz > my-app-darwin-arm64.tar.gz.sha256

      - uses: actions/upload-artifact@v4
        with:
          name: macos-build
          path: |
            my-app-darwin-arm64.tar.gz
            my-app-darwin-arm64.tar.gz.sha256

  build-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install dependencies
        run: sudo apt-get install -y libwebkit2gtk-4.1-dev libgtk-3-dev

      - name: Install LightShell
        run: curl -fsSL https://lightshell.dev/install.sh | sh

      - name: Build
        run: lightshell build

      - name: Package
        run: |
          tar -czf my-app-linux-x64.tar.gz -C dist MyApp.AppImage
          shasum -a 256 my-app-linux-x64.tar.gz > my-app-linux-x64.tar.gz.sha256

      - uses: actions/upload-artifact@v4
        with:
          name: linux-build
          path: |
            my-app-linux-x64.tar.gz
            my-app-linux-x64.tar.gz.sha256

  release:
    needs: [build-macos, build-linux]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4

      - name: Generate manifest
        run: |
          VERSION=${GITHUB_REF_NAME#v}
          MAC_SHA=$(cut -d' ' -f1 macos-build/my-app-darwin-arm64.tar.gz.sha256)
          LINUX_SHA=$(cut -d' ' -f1 linux-build/my-app-linux-x64.tar.gz.sha256)
          REPO_URL="https://github.com/${{ github.repository }}/releases/download/${GITHUB_REF_NAME}"

          cat > latest.json << EOF
          {
            "version": "${VERSION}",
            "notes": "Release ${VERSION}",
            "pub_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
            "platforms": {
              "darwin-arm64": {
                "url": "${REPO_URL}/my-app-darwin-arm64.tar.gz",
                "sha256": "${MAC_SHA}"
              },
              "linux-x64": {
                "url": "${REPO_URL}/my-app-linux-x64.tar.gz",
                "sha256": "${LINUX_SHA}"
              }
            }
          }
          EOF

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            macos-build/my-app-darwin-arm64.tar.gz
            linux-build/my-app-linux-x64.tar.gz
            latest.json
```

Push a tag like `v1.1.0` and the workflow builds for both platforms, generates the manifest with correct hashes, and uploads everything as a GitHub Release.

## Option 2: S3 or R2 Bucket

For apps that need more control over hosting or want to avoid GitHub's rate limits.

### File Structure

```
s3://my-app-releases/
  latest.json
  v1.0.0/
    my-app-darwin-arm64.tar.gz
    my-app-darwin-x64.tar.gz
    my-app-linux-x64.tar.gz
  v1.1.0/
    my-app-darwin-arm64.tar.gz
    my-app-darwin-x64.tar.gz
    my-app-linux-x64.tar.gz
```

### Upload Script

```bash
#!/bin/bash
VERSION="1.1.0"
BUCKET="s3://my-app-releases"

# Upload binaries
aws s3 cp my-app-darwin-arm64.tar.gz "$BUCKET/v$VERSION/"
aws s3 cp my-app-linux-x64.tar.gz "$BUCKET/v$VERSION/"

# Upload manifest (overwrites previous)
aws s3 cp latest.json "$BUCKET/latest.json"
```

Set the bucket to allow public reads. CORS configuration is not needed because the Go backend makes the requests, not the browser.

### Cloudflare R2

Identical to S3 but with no egress fees. Use the S3-compatible API:

```bash
aws s3 cp latest.json s3://my-app-releases/latest.json \
  --endpoint-url https://YOUR_ACCOUNT.r2.cloudflarestorage.com
```

## Option 3: Any Static Host

Any server that can serve static files over HTTPS works. This includes Netlify, Vercel, your own nginx, a CDN, or even a GitHub Pages site.

Example with a simple nginx config:

```nginx
server {
    listen 443 ssl;
    server_name releases.myapp.com;

    root /var/www/releases;

    location / {
        autoindex off;
        try_files $uri =404;
    }
}
```

Upload your files to `/var/www/releases/latest.json` and `/var/www/releases/v1.1.0/...` and set the endpoint accordingly.

## Endpoint URL in lightshell.json

Whichever hosting option you choose, set the endpoint to the direct URL of your `latest.json`:

```json
{
  "updater": {
    "endpoint": "https://releases.myapp.com/latest.json"
  }
}
```

The endpoint URL is baked into the binary at build time. If you need to change it, you must release a new version of your app with the updated URL. Choose a URL you control and that will not change.
