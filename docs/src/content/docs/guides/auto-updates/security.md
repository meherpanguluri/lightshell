---
title: Update Security
description: Security guarantees for LightShell auto-updates.
---

The LightShell auto-updater is designed to be simple and secure. Every update is verified before installation, and the existing binary is never modified until verification passes. This page documents the security model and what protections are in place.

## SHA256 Verification

Every platform entry in the update manifest includes a `sha256` field -- the hex-encoded SHA256 hash of the archive file. This hash is **mandatory**. If it is missing from the manifest, the updater refuses to download.

After downloading the archive, the Go backend computes the SHA256 hash of the downloaded file and compares it character-by-character to the hash in the manifest:

```
Expected (from manifest): a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678
Computed (from download):  a3f2b8c1d4e5f67890abcdef1234567890abcdef1234567890abcdef12345678
→ Match: proceed with installation
```

If the hashes do not match, the update is rejected:

1. The downloaded file is deleted
2. The temp directory is cleaned up
3. An error is returned to JavaScript
4. The existing app binary is untouched

This protects against:

- **Corrupted downloads** -- partial downloads, network errors, or CDN issues that produce an incomplete file
- **Tampered archives** -- an attacker who modifies the archive on the server or in transit cannot produce a file that matches the expected hash
- **CDN cache poisoning** -- even if an attacker replaces the file on a CDN edge, the hash check catches it

## HTTPS Requirement

In production builds (output of `lightshell build`), the updater enforces HTTPS for both the manifest endpoint and download URLs:

- The `endpoint` URL in `lightshell.json` must use `https://`
- Every `url` field in the manifest's platform entries must use `https://`
- If any URL uses `http://`, the updater rejects it with an error

This protects against man-in-the-middle attacks. Without HTTPS, an attacker on the same network could intercept the manifest request and return a modified manifest pointing to a malicious binary with a matching hash.

**In dev mode** (`lightshell dev`), HTTP URLs are allowed. This lets you test the update flow against a local server without setting up TLS certificates. Do not ship apps that use HTTP endpoints.

## What the Updater Does Not Do

The updater is deliberately simple. It does **not**:

- **Execute scripts during update** -- there are no pre-install or post-install hooks. The updater downloads, verifies, extracts, and replaces. Nothing else runs.
- **Auto-install without consent** -- background checks only emit events. The `install()` call must be made explicitly by your code, which means you can require user confirmation.
- **Modify app data** -- the updater only touches the app binary. Your data directory (`$APP_DATA`), store database, and user files are never read or written during an update.
- **Downgrade** -- the updater only installs versions newer than the current one (by semver comparison). A manifest with an older version is treated as "no update available."

## Binary Replacement Safety

The binary replacement step uses a rename-based atomic swap to prevent corruption:

1. The new binary is written to a temp location on the same filesystem
2. The old binary is renamed to `{binary}.backup`
3. The new binary is renamed to the original path
4. The backup is deleted

If step 3 fails (e.g., disk full, permissions error), step 2 is rolled back -- the backup is renamed back to the original path. The app is never left without a working binary.

On macOS, the `.app` bundle structure is preserved. Only the files that have changed are replaced. The `Info.plist`, icon, and other resources are updated if the new archive includes them.

## Manifest Integrity

The manifest itself is fetched over HTTPS, which provides transport-layer integrity. However, the manifest is not independently signed -- its integrity depends on the security of your hosting and HTTPS certificate.

Recommendations for manifest security:

- **Use a domain you control** for the endpoint URL. Avoid third-party URL shorteners or redirects you do not manage.
- **Enable access logging** on your server or S3 bucket so you can detect unauthorized manifest modifications.
- **Use GitHub Releases** if possible -- GitHub's infrastructure provides reliable HTTPS and access controls.
- **Pin the endpoint** -- the endpoint URL is baked into the binary at build time, so an attacker cannot redirect it without modifying the binary itself.

## Threat Model

Here is what the update system protects against and what it does not:

| Threat | Protected | How |
|--------|-----------|-----|
| Corrupted download | Yes | SHA256 verification |
| Network MITM (modify archive in transit) | Yes | HTTPS + SHA256 |
| Tampered archive on server | Yes | SHA256 verification (attacker must also change manifest) |
| Compromised manifest + matching archive | No | If attacker controls both manifest and archive, they can deliver a malicious update. Protect your hosting. |
| Compromised build pipeline | No | If the attacker can modify your build output, they produce both the archive and the hash. Use CI/CD security best practices. |
| DNS hijacking | Partially | HTTPS certificate validation prevents serving content from a fake server, but DNS attacks can deny service. |

The security model assumes you control your hosting infrastructure and build pipeline. The updater verifies that what you published is what the user receives.

## Recommendations

**For all apps:**
- Always use HTTPS endpoints
- Generate SHA256 hashes as part of your build pipeline, not manually
- Verify the manifest is correct after uploading (fetch and check the hashes yourself)

**For apps with sensitive data:**
- Host the manifest on infrastructure you control (not a shared hosting service)
- Monitor the manifest endpoint for unauthorized changes
- Consider adding a version check in your app that alerts you if the manifest version jumps unexpectedly

**For apps distributed to enterprises:**
- Document the manifest endpoint URL so IT teams can whitelist it
- Use a stable, dedicated domain for releases (e.g., `releases.yourapp.com`)
- Keep old versions available for rollback (do not delete old archives from your server)

## Verification Failure Handling

When SHA256 verification fails, the updater:

1. Deletes the downloaded archive immediately
2. Cleans up the temp directory
3. Returns an error to your JavaScript code

```js
try {
  await lightshell.updater.install()
} catch (err) {
  // err.message: "SHA256 verification failed: expected a3f2b8..., got fff..."
  console.error('Update verification failed:', err.message)
  await lightshell.dialog.message(
    'Update Failed',
    'The update could not be verified and was not installed. Please try again later.'
  )
}
```

A verification failure is not normal. If it happens, it means the archive on your server does not match the hash in your manifest. Common causes:

- You updated the archive but forgot to update the hash in `latest.json`
- Your CDN is serving a cached (stale) version of the archive
- The upload was interrupted and the file on the server is incomplete
- (Rare) The file was tampered with
