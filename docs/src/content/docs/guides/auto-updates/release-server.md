---
title: Release Server
description: How to set up and deploy the lightshell-server for hosting app updates.
---

The **lightshell-server** is a lightweight Hono-based TypeScript server purpose-built for hosting signed LightShell releases. It handles update manifests, binary storage, Ed25519 signature verification, and provides a dashboard for monitoring downloads.

You do not need this server to ship updates -- you can host update manifests on any static file server or [GitHub Releases](/docs/guides/auto-updates/github-releases/). The release server adds convenience features like automatic manifest generation, a web dashboard, and download analytics.

## One-Click Deploy

Deploy the release server to your preferred platform in one click. Each button will fork the repo, prompt you for environment variables, and deploy automatically.

| Platform | Deploy | Free Tier | Notes |
|----------|--------|-----------|-------|
| **Railway** | [![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/Kvwoxh?referralCode=oLNtQ-) | $5/mo credit | Persistent disk, automatic HTTPS |
| **Render** | [![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/lightshell-dev/lightshell-server) | Free tier available | Auto-deploy from GitHub, managed TLS |
| **Cloudflare Workers** | -- | 100K req/day free | Fastest option, see [manual setup](#cloudflare-workers-recommended) below |
| **Fly.io** | -- | 3 shared VMs free | See [manual setup](#flyio) below |

After deploying, you will need to set two environment variables:

- **`LIGHTSHELL_API_KEY`** -- generate with `openssl rand -hex 32`
- **`PUBLIC_KEY`** -- your Ed25519 public key (see [Signing Keys](/docs/guides/auto-updates/signing-keys/))

### What the one-click deploy does

1. Forks `lightshell-dev/lightshell-server` into your account
2. Prompts you for required environment variables
3. Builds and deploys the server
4. Gives you a public HTTPS URL to use as your updater endpoint

---

## Manual Deploy

If you prefer CLI deployments or need more control over configuration, use one of the methods below.

### Cloudflare Workers (Recommended)

```bash
# Clone the release server
git clone https://github.com/lightshell-dev/lightshell-server
cd lightshell-server

# Install dependencies
npm install

# Configure
cp wrangler.example.toml wrangler.toml
# Edit wrangler.toml with your account details

# Set secrets
npx wrangler secret put LIGHTSHELL_API_KEY
npx wrangler secret put PUBLIC_KEY

# Deploy
npx wrangler deploy
```

Your server will be live at `https://lightshell-server.<your-subdomain>.workers.dev`.

### Railway

```bash
git clone https://github.com/lightshell-dev/lightshell-server
cd lightshell-server

# Install Railway CLI: https://docs.railway.com/guides/cli
railway login
railway init
railway up

# Set environment variables in the Railway dashboard
```

Or use the Railway dashboard directly:

1. Go to [railway.com/new](https://railway.com/new)
2. Select **Deploy from GitHub repo**
3. Connect `lightshell-dev/lightshell-server` (or your fork)
4. Add environment variables: `LIGHTSHELL_API_KEY`, `PUBLIC_KEY`, `STORAGE_BACKEND=local`, `STORAGE_PATH=/data/releases`
5. Add a volume mounted at `/data` for persistent storage
6. Deploy

### Render

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/lightshell-dev/lightshell-server)

Or manually:

1. Fork `lightshell-dev/lightshell-server` on GitHub
2. Create a new **Web Service** on [render.com](https://render.com)
3. Connect your forked repository
4. Set **Build Command** to `npm install && npm run build`
5. Set **Start Command** to `npm start`
6. Add environment variables in the Render dashboard
7. Add a **Disk** mounted at `/data` for local storage (if not using S3)
8. Deploy

### Fly.io

```bash
git clone https://github.com/lightshell-dev/lightshell-server
cd lightshell-server

fly launch
fly secrets set LIGHTSHELL_API_KEY=your-secret-key
fly secrets set PUBLIC_KEY=your-ed25519-public-key
fly deploy
```

### Self-Hosted

Run the server on any machine with Node.js 20+ or Bun:

```bash
git clone https://github.com/lightshell-dev/lightshell-server
cd lightshell-server
npm install
npm run build

# Set environment variables
export LIGHTSHELL_API_KEY="your-secret-key"
export PUBLIC_KEY="your-ed25519-public-key"
export STORAGE_BACKEND="local"
export STORAGE_PATH="/var/lib/lightshell-releases"
export PORT=3000

npm start
```

Put it behind a reverse proxy (Caddy, nginx) with HTTPS. The updater requires HTTPS in production builds.

---

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `LIGHTSHELL_API_KEY` | Yes | API key for authenticating upload requests from CI/CD. Generate with `openssl rand -hex 32`. |
| `PUBLIC_KEY` | Yes | Your Ed25519 public key for verifying release signatures. See [Signing Keys](/docs/guides/auto-updates/signing-keys/). |
| `STORAGE_BACKEND` | No | Where to store release binaries. Options: `r2` (Cloudflare R2, default for Workers), `s3`, `local`. |
| `STORAGE_PATH` | No | Path for local storage backend. Default: `./releases`. |
| `S3_BUCKET` | No | S3 bucket name (when using `s3` backend). |
| `S3_REGION` | No | S3 region (when using `s3` backend). |
| `R2_BUCKET` | No | R2 bucket name (when using `r2` backend). Default: `lightshell-releases`. |
| `PORT` | No | Server port for self-hosted deployments. Default: `3000`. |

## Connecting Your App

Once the server is running, point your app's updater config to it.

**lightshell.json:**
```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "updater": {
    "enabled": true,
    "endpoint": "https://your-release-server.example.com/api/latest/my-app",
    "interval": "24h"
  }
}
```

The server generates the update manifest automatically from uploaded releases. The endpoint URL follows the pattern:

```
https://<server>/api/latest/<app-name>
```

## Uploading Releases

Upload signed release binaries using the CLI or a direct HTTP request.

### With the CLI

```bash
# Build your app
lightshell build

# Upload (reads LIGHTSHELL_RELEASE_SERVER and LIGHTSHELL_RELEASE_TOKEN from env)
lightshell release --server https://your-release-server.example.com
```

### With curl

```bash
curl -X POST https://your-release-server.example.com/api/releases \
  -H "Authorization: Bearer your-api-key" \
  -F "app=my-app" \
  -F "version=1.2.0" \
  -F "platform=darwin-arm64" \
  -F "notes=Bug fixes and performance improvements" \
  -F "signature=base64-encoded-ed25519-signature" \
  -F "file=@dist/my-app-darwin-arm64.tar.gz"
```

The server computes the SHA256 hash automatically and stores it in the manifest.

## Dashboard

The release server includes a web dashboard at the root URL (`https://your-server.example.com/`). The dashboard shows:

- **Current version** for each platform
- **Download counts** per version and platform
- **Release history** with timestamps and release notes
- **Health status** of the storage backend

The dashboard is read-only and does not require authentication. It does not expose API keys, private keys, or download URLs -- only version numbers, release notes, and aggregate statistics.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/latest/:app` | Returns the update manifest for the latest version. This is what the updater fetches. |
| `POST` | `/api/releases` | Upload a new release binary. Requires `Authorization` header. |
| `GET` | `/api/releases/:app` | List all releases for an app (version history). |
| `DELETE` | `/api/releases/:app/:version` | Delete a specific release. Requires `Authorization` header. |
| `GET` | `/health` | Health check endpoint. Returns `200 OK`. |

## Storage Backends

### Cloudflare R2

The default for Cloudflare Workers deployments. R2 provides S3-compatible storage with no egress fees. Binaries are stored in an R2 bucket bound to the Worker.

### Amazon S3

For self-hosted or Railway/Render deployments. Set `STORAGE_BACKEND=s3` and provide `S3_BUCKET` and `S3_REGION`. The server uses the AWS SDK, so standard credential resolution applies (environment variables, IAM role, etc.).

### Local Filesystem

For development or simple self-hosted setups. Set `STORAGE_BACKEND=local` and `STORAGE_PATH` to a directory with sufficient disk space. Make sure the directory is backed up -- losing it means losing all release binaries.

## Next Steps

- [Signing Keys](/docs/guides/auto-updates/signing-keys/) -- generate Ed25519 keys for signing releases
- [CI/CD](/docs/guides/auto-updates/ci-cd/) -- automate the build-sign-upload pipeline with GitHub Actions
- [Update Security](/docs/guides/auto-updates/security/) -- understand the security model
