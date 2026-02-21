---
title: HTTP API
description: Complete reference for lightshell.http — CORS-free HTTP requests.
---

The `lightshell.http` module makes HTTP requests through the Go backend, completely bypassing CORS restrictions. Unlike `fetch()` in the webview (which is subject to browser CORS policies), `lightshell.http.fetch()` can call any API endpoint without CORS headers. This is one of the most important advantages of a desktop app over a web app. All methods are async and return Promises.

## Methods

### fetch(url, options?)

Make an HTTP request to any URL. The request is executed by the Go backend using `net/http`, so CORS does not apply.

**Parameters:**
- `url` (string) — the full URL to request
- `options` (object, optional):
  - `method` (string) — HTTP method: `"GET"`, `"POST"`, `"PUT"`, `"PATCH"`, `"DELETE"`, `"HEAD"` (default: `"GET"`)
  - `headers` (object) — key-value pairs of request headers
  - `body` (string) — request body (for POST, PUT, PATCH)
  - `timeout` (number) — request timeout in milliseconds (default: `30000`)

**Returns:** `Promise<{ status: number, headers: object, body: string }>` — the response object:
  - `status` — HTTP status code (e.g., `200`, `404`)
  - `headers` — response headers as key-value pairs
  - `body` — response body as a string (parse with `JSON.parse()` for JSON APIs)

**Example:**
```js
// Simple GET request
const response = await lightshell.http.fetch('https://api.github.com/zen')
console.log(response.body) // "Encourage flow."

// GET with headers
const data = await lightshell.http.fetch('https://api.github.com/user', {
  headers: {
    'Authorization': 'Bearer ghp_xxxxx',
    'Accept': 'application/json'
  }
})
const user = JSON.parse(data.body)
console.log(user.login)

// POST with JSON body
const result = await lightshell.http.fetch('https://api.example.com/items', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'New Item', priority: 'high' })
})
console.log(result.status) // 201
```

**Errors:** Rejects on network errors (DNS failure, connection refused, timeout). HTTP error status codes (4xx, 5xx) do NOT cause a rejection — check `response.status` instead.

---

### download(url, options?)

Download a file directly to disk. Unlike `fetch()`, this does not load the entire response into memory, making it suitable for large files.

**Parameters:**
- `url` (string) — the URL of the file to download
- `options` (object, optional):
  - `saveTo` (string) — destination file path. Supports path variables like `$DOWNLOADS`, `$TEMP`, `$DESKTOP`.
  - `headers` (object) — request headers (e.g., for authentication)
  - `onProgress` (function) — callback fired during download with progress info

**Returns:** `Promise<{ path: string, size: number }>` — the saved file path and size in bytes

**Example:**
```js
const result = await lightshell.http.download(
  'https://example.com/report.pdf',
  {
    saveTo: '$DOWNLOADS/report.pdf',
    onProgress: (p) => {
      console.log(`${p.percent}% (${p.bytesDownloaded}/${p.totalBytes})`)
    }
  }
)
console.log(`Saved to ${result.path} (${result.size} bytes)`)
```

**Errors:** Rejects on network errors or if the destination path is not writable.

---

## Common Patterns

### Calling REST APIs

```js
async function fetchTodos() {
  const response = await lightshell.http.fetch('https://api.myapp.com/todos', {
    headers: {
      'Authorization': `Bearer ${apiToken}`,
      'Accept': 'application/json'
    }
  })

  if (response.status !== 200) {
    throw new Error(`API error: ${response.status}`)
  }

  return JSON.parse(response.body)
}

async function createTodo(title) {
  const response = await lightshell.http.fetch('https://api.myapp.com/todos', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${apiToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ title, completed: false })
  })

  if (response.status !== 201) {
    throw new Error(`Failed to create todo: ${response.status}`)
  }

  return JSON.parse(response.body)
}
```

### API Key Authentication

```js
// Create a reusable API client
function createApiClient(baseUrl, apiKey) {
  async function request(path, options = {}) {
    const response = await lightshell.http.fetch(`${baseUrl}${path}`, {
      ...options,
      headers: {
        'X-API-Key': apiKey,
        'Content-Type': 'application/json',
        ...options.headers
      }
    })
    return {
      status: response.status,
      data: JSON.parse(response.body)
    }
  }

  return {
    get: (path) => request(path),
    post: (path, body) => request(path, {
      method: 'POST',
      body: JSON.stringify(body)
    }),
    put: (path, body) => request(path, {
      method: 'PUT',
      body: JSON.stringify(body)
    }),
    delete: (path) => request(path, { method: 'DELETE' })
  }
}

// Usage
const api = createApiClient('https://api.example.com', 'sk_live_xxxxx')
const { data } = await api.get('/users/me')
await api.post('/messages', { text: 'Hello!' })
```

### Download with Progress UI

```js
async function downloadFile(url, filename) {
  const progressBar = document.getElementById('progress')
  const statusText = document.getElementById('status')

  statusText.textContent = 'Downloading...'
  progressBar.style.width = '0%'

  try {
    const result = await lightshell.http.download(url, {
      saveTo: `$DOWNLOADS/${filename}`,
      onProgress: (p) => {
        progressBar.style.width = `${p.percent}%`
        statusText.textContent = `${p.percent}% — ${formatBytes(p.bytesDownloaded)} / ${formatBytes(p.totalBytes)}`
      }
    })

    statusText.textContent = `Downloaded to ${result.path}`
    await lightshell.notify.send('Download Complete', `${filename} saved to Downloads.`)
  } catch (err) {
    statusText.textContent = `Download failed: ${err.message}`
  }
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1048576).toFixed(1)} MB`
}
```

## Platform Notes

- Requests are made by the Go backend using `net/http`, not by the webview. This means CORS headers are not required on the target server.
- The response `body` is always a string. For JSON responses, parse with `JSON.parse()`. For binary responses, the body is base64-encoded.
- There is no streaming support in v1 — the entire response is buffered in memory for `fetch()`. Use `download()` for large files.
- `download()` writes directly to disk and does not load the file into memory.
- Path variables in `saveTo`: `$DOWNLOADS` resolves to `~/Downloads`, `$TEMP` to `/tmp`, `$DESKTOP` to `~/Desktop`.
- In restricted permission mode, URLs must match patterns defined in `permissions.http.allow` in `lightshell.json`. Without a `permissions` key, all URLs are allowed.
- HTTPS is enforced for production builds in the updater, but `lightshell.http.fetch()` allows both HTTP and HTTPS in all modes.
- WebSocket support is not available in v1. Use polling or `lightshell.http.fetch()` for real-time data.
