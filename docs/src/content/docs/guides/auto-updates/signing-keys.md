---
title: Signing Keys
description: Ed25519 key management for secure release signing.
---

LightShell uses Ed25519 digital signatures to verify that release binaries were produced by you and have not been tampered with. This page covers key generation, storage, signing workflow, and key rotation.

## Why Signing Matters

SHA256 hashes verify that the file you downloaded matches the file the server has. But they do not prove **who** produced the file. If an attacker gains access to your release server, they can upload a malicious binary and update the SHA256 hash to match.

Ed25519 signatures close this gap. The signature is produced with your **private key** (which only you have) and verified with your **public key** (which is embedded in the app binary). Even if an attacker controls your server, they cannot produce a valid signature without your private key.

## Generate Keys

```bash
lightshell keys generate
```

This creates two files:

| File | Location | Purpose |
|------|----------|---------|
| Private key | `~/.lightshell/signing-key.pem` | Signs release archives. Keep this secret. |
| Public key | Printed to stdout | Verifies signatures. Embed in your app config. |

Output:

```
Generated Ed25519 signing key pair.

Private key saved to: ~/.lightshell/signing-key.pem
  → KEEP THIS SECRET. Do not commit to version control.
  → Back up securely (password manager, encrypted drive).

Public key:
  MCowBQYDK2VwAyEA7K1bN...

Add this public key to your lightshell.json:
  "updater": {
    "publicKey": "MCowBQYDK2VwAyEA7K1bN..."
  }
```

Copy the public key into your `lightshell.json`:

```json
{
  "updater": {
    "enabled": true,
    "endpoint": "https://releases.example.com/api/latest/my-app",
    "publicKey": "MCowBQYDK2VwAyEA7K1bN...",
    "interval": "24h"
  }
}
```

The public key is embedded into the built binary at compile time. It cannot be changed after the app is built, which prevents an attacker from substituting a different key.

## Key Storage

### Private Key

The private key file at `~/.lightshell/signing-key.pem` is created with `0600` permissions (owner read/write only). Treat it like an SSH private key:

- **Never** commit it to version control
- **Never** share it in chat, email, or issue trackers
- Store a backup in a password manager or encrypted drive
- In CI/CD, store it as an encrypted secret (see [CI/CD](/guides/auto-updates/ci-cd/))

If you lose the private key, you cannot sign new releases that match the public key embedded in existing app installations. You would need to ship a new version with a rotated public key (see [Key Rotation](#key-rotation) below).

### Public Key

The public key is **not** secret. It is safe to:

- Commit it to version control (it is in `lightshell.json`)
- Share it publicly
- Embed it in documentation

The public key is only used for verification, not signing.

## How Signing Works

Here is the step-by-step flow when you publish a release:

1. **Build** your app: `lightshell build`
2. **Compress** the output into a `.tar.gz` archive
3. **Sign** the archive: `lightshell release --sign`
   - Reads the private key from `~/.lightshell/signing-key.pem`
   - Computes the Ed25519 signature over the raw archive bytes
   - Base64-encodes the signature
4. **Upload** the archive + signature to your release server

When the updater downloads a release:

1. **Fetch** the update manifest (includes `sha256` and `signature` fields)
2. **Download** the archive
3. **Verify SHA256** -- reject if the hash does not match
4. **Verify signature** -- decode the base64 signature, verify against the archive bytes using the embedded public key
5. **Install** only if both checks pass

```
Archive bytes ──┬── SHA256 hash ──── matches manifest hash?     ── YES ─┐
                │                                                        │
                └── Ed25519 verify(publicKey, signature, bytes)? ── YES ─┤
                                                                         │
                                                             Install ◄───┘
```

## Key Rotation

If your private key is compromised (or you suspect it might be), rotate keys immediately:

### Step 1: Generate a New Key Pair

```bash
# Back up the old key first
mv ~/.lightshell/signing-key.pem ~/.lightshell/signing-key-old.pem

# Generate new key pair
lightshell keys generate
```

### Step 2: Update lightshell.json

Replace the `publicKey` with the new public key:

```json
{
  "updater": {
    "publicKey": "NEW_PUBLIC_KEY_HERE"
  }
}
```

### Step 3: Ship a Transition Release

Build and sign a new release with the **old** private key that contains the **new** public key:

```bash
# Sign with the old key one last time
LIGHTSHELL_SIGNING_KEY=~/.lightshell/signing-key-old.pem lightshell release --sign
```

This release is verifiable by existing installations (they have the old public key), and once installed, the new public key takes effect. All subsequent releases should be signed with the new private key.

### Step 4: Sign Future Releases with the New Key

```bash
# New releases use the new key automatically
lightshell release --sign
```

### Step 5: Clean Up

After enough time has passed for users to update past the transition release:

```bash
rm ~/.lightshell/signing-key-old.pem
```

## Threat Model

| Threat | Protected | How |
|--------|-----------|-----|
| Tampered archive on server | Yes | Signature verification fails without the private key |
| Compromised release server | Yes | Attacker cannot sign with your private key |
| Compromised CDN | Yes | Signature is computed over the archive, not the URL |
| Stolen private key | No | Attacker can sign malicious releases. Rotate keys immediately. |
| Compromised build machine | No | Attacker has access to the private key on that machine. Rotate keys and audit CI. |
| Man-in-the-middle (no HTTPS) | Partially | Signature still valid, but attacker could block updates. Always use HTTPS. |

## CI/CD Secrets

In automated pipelines, store the private key as a CI secret rather than a file:

```yaml
# GitHub Actions example
env:
  LIGHTSHELL_SIGNING_KEY: ${{ secrets.SIGNING_PRIVATE_KEY }}
```

The `lightshell release --sign` command checks for the `LIGHTSHELL_SIGNING_KEY` environment variable before falling back to the file at `~/.lightshell/signing-key.pem`. The variable should contain the full PEM-encoded key contents.

See [CI/CD](/guides/auto-updates/ci-cd/) for a complete GitHub Actions workflow.

## Verifying a Signature Manually

To manually verify that an archive was signed correctly:

```bash
# Extract the signature from the manifest
curl -s https://releases.example.com/api/latest/my-app | jq -r '.platforms["darwin-arm64"].signature'

# Verify (requires openssl 3.0+)
echo -n "<base64-signature>" | base64 -d > sig.bin
openssl pkeyutl -verify \
  -pubin -inkey public-key.pem \
  -sigfile sig.bin \
  -in my-app-darwin-arm64.tar.gz
```

A successful verification prints `Signature Verified Successfully`.
