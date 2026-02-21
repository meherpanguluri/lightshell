---
title: Code Signing
description: Sign your LightShell app for trusted distribution on macOS.
---

Code signing tells macOS that your app comes from a known developer and has not been tampered with. Without a signature, macOS Gatekeeper blocks your app and users see an "unidentified developer" warning. Signing eliminates this friction. Notarization goes one step further — Apple scans your app and confirms it is free of malicious content, removing all security warnings.

## Why Sign Your App

Unsigned apps trigger Gatekeeper on macOS:

> "MyApp" can't be opened because Apple cannot check it for malicious software.

Users must right-click and select "Open" to bypass this, which is confusing and reduces trust. A signed app opens immediately with no warnings. A signed and notarized app opens with no warnings on any macOS version, including the latest.

If you are distributing your app to anyone beyond yourself, signing is strongly recommended.

## Requirements

To sign a macOS app, you need:

1. **An Apple Developer account** — enroll at [developer.apple.com](https://developer.apple.com/programs/) ($99/year for individuals)
2. **A Developer ID Application certificate** — create one in your Apple Developer account under Certificates, Identifiers & Profiles
3. **The certificate installed in your Keychain** — download and double-click the certificate file to install it

Your signing identity looks like this:

```
Developer ID Application: Your Name (TEAMID)
```

You can find your identity by running:

```bash
security find-identity -v -p codesigning
```

This lists all valid code signing identities in your Keychain.

## Configuration

Add your signing identity to `lightshell.json`:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "build": {
    "appId": "com.example.myapp",
    "icon": "assets/icon.png",
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)"
    }
  }
}
```

## Signing a Build

Use the `--sign` flag with any macOS build target:

```bash
# Sign a .app bundle
lightshell build --sign

# Sign and package as DMG
lightshell build --target dmg --sign
```

Output:

```
✓ Built my-app in 1.2s → 5.1MB
✓ Created MyApp.app
✓ Signed MyApp.app with "Developer ID Application: Your Name (TEAMID)"
✓ Packaged DMG → dist/MyApp-1.0.0.dmg
```

## What Happens During Signing

When you build with `--sign`, LightShell runs:

```bash
codesign --deep --force --options runtime --sign "Developer ID Application: Your Name (TEAMID)" MyApp.app
```

The flags:

- `--deep` signs the app bundle and all nested code (frameworks, helpers)
- `--force` replaces any existing signature
- `--options runtime` enables the hardened runtime, which is required for notarization
- `--sign` specifies the signing identity

After signing, LightShell verifies the signature:

```bash
codesign --verify --deep --strict MyApp.app
```

If verification fails, the build stops and reports the error.

## Entitlements

Entitlements declare what system capabilities your app needs. LightShell configures these in `lightshell.json`:

```json
{
  "build": {
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)",
      "entitlements": {
        "com.apple.security.network.client": true,
        "com.apple.security.files.user-selected.read-write": true
      }
    }
  }
}
```

Common entitlements for LightShell apps:

| Entitlement | Purpose |
|-------------|---------|
| `com.apple.security.network.client` | Make outbound network requests (needed for `lightshell.http.fetch`) |
| `com.apple.security.files.user-selected.read-write` | Read/write files selected via open/save dialogs |
| `com.apple.security.files.downloads.read-write` | Access the Downloads folder |
| `com.apple.security.automation.apple-events` | Send Apple Events to other apps |

LightShell generates an `entitlements.plist` from this configuration and passes it to `codesign` during signing. If you do not specify entitlements, LightShell uses sensible defaults (`network.client` and `files.user-selected.read-write`).

## Notarization

Notarization submits your signed app to Apple for automated security scanning. Once approved, Apple records that your app is safe, and macOS will open it without any warnings on any Mac.

### Setup

Notarization requires your Apple ID and Team ID. Add them to your build configuration or pass them as environment variables:

```json
{
  "build": {
    "mac": {
      "identity": "Developer ID Application: Your Name (TEAMID)",
      "appleId": "your@email.com",
      "teamId": "TEAMID"
    }
  }
}
```

You also need an app-specific password for your Apple ID. Generate one at [appleid.apple.com](https://appleid.apple.com/) under Security > App-Specific Passwords. Store it in your Keychain:

```bash
xcrun notarytool store-credentials "lightshell-notarize" \
  --apple-id "your@email.com" \
  --team-id "TEAMID" \
  --password "xxxx-xxxx-xxxx-xxxx"
```

### Running Notarization

```bash
lightshell build --target dmg --sign --notarize
```

This performs the following steps:

1. Builds and signs the `.app` bundle
2. Packages it into a DMG
3. Submits the DMG to Apple's notarization service via `xcrun notarytool submit`
4. Waits for Apple to complete the scan (typically 1-5 minutes)
5. Staples the notarization ticket to the DMG via `xcrun stapler staple`

Output:

```
✓ Built my-app in 1.2s → 5.1MB
✓ Signed MyApp.app
✓ Packaged DMG → dist/MyApp-1.0.0.dmg
✓ Submitted for notarization... waiting
✓ Notarized successfully (ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
✓ Stapled notarization ticket to MyApp-1.0.0.dmg
```

## Verifying a Signed App

Check the signature and notarization status:

```bash
# Verify the code signature
codesign --verify --deep --strict --verbose=2 dist/MyApp.app

# Check notarization status
spctl --assess --verbose=2 dist/MyApp.app

# Check a DMG
spctl --assess --type open --verbose=2 dist/MyApp-1.0.0.dmg
```

A properly signed and notarized app shows:

```
dist/MyApp.app: accepted
source=Notarized Developer ID
```

## Troubleshooting

### "No identity found"

Your signing certificate is not in the Keychain. Verify with:

```bash
security find-identity -v -p codesigning
```

If empty, download your certificate from the Apple Developer portal and install it.

### "The signature is invalid"

The app was modified after signing. Rebuild and sign again. Ensure no post-processing step modifies the `.app` bundle after signing.

### Notarization rejected

Check the detailed log:

```bash
xcrun notarytool log {submission-id} --keychain-profile "lightshell-notarize"
```

Common reasons for rejection:

- Missing hardened runtime (`--options runtime`) — LightShell includes this automatically
- Unsigned nested binaries — LightShell uses `--deep` to sign everything
- Missing entitlements for capabilities the app uses

### "errSecInternalComponent"

This usually means the Keychain is locked. Unlock it:

```bash
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

In CI environments, you may need to create and unlock a temporary Keychain.

## CI Integration

For automated signing in GitHub Actions:

```yaml
jobs:
  build-signed:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - name: Import certificate
        env:
          CERTIFICATE_P12: ${{ secrets.CERTIFICATE_P12 }}
          CERTIFICATE_PASSWORD: ${{ secrets.CERTIFICATE_PASSWORD }}
        run: |
          echo "$CERTIFICATE_P12" | base64 --decode > certificate.p12
          security create-keychain -p "" build.keychain
          security import certificate.p12 -k build.keychain -P "$CERTIFICATE_PASSWORD" -T /usr/bin/codesign
          security set-key-partition-list -S apple-tool:,apple: -s -k "" build.keychain
          security default-keychain -s build.keychain
      - name: Build and sign
        run: lightshell build --target dmg --sign
```

Store your `.p12` certificate file as a base64-encoded GitHub Secret.
