---
title: "6. Connecting to APIs"
description: Make CORS-free HTTP requests from your desktop app using lightshell.http.
---

Desktop apps often need to call external APIs — fetching data from GitHub, posting to a webhook, downloading files. In a browser, `fetch()` is restricted by CORS policies that block requests to most APIs. Your LightShell app does not have this limitation.

## The CORS Problem

If you try calling an API with the standard `fetch()` in a webview, you will often see this:

```
Access to fetch at 'https://api.example.com/data' from origin 'null'
has been blocked by CORS policy
```

CORS (Cross-Origin Resource Sharing) is a browser security mechanism. It makes sense for websites, but desktop apps should be able to call any API freely — just like `curl` or Postman can.

## How lightshell.http Works

`lightshell.http.fetch()` routes the HTTP request through the Go backend instead of the webview. The Go backend uses `net/http`, which has no CORS restrictions. The result comes back to your JavaScript via IPC.

```
Your JS code → IPC → Go backend → net/http → External API
                                       ↓
Your JS code ← IPC ← Go backend ← Response
```

The API is intentionally similar to `fetch()`, so the learning curve is minimal.

## Building a GitHub User Lookup

Let's build a tool that fetches GitHub user profiles. Create `src/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>GitHub Lookup</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <h1>GitHub User Lookup</h1>
  <div class="search-row">
    <input type="text" id="username" placeholder="Enter a GitHub username">
    <button id="search-btn">Search</button>
  </div>
  <div id="result"></div>
  <script src="app.js"></script>
</body>
</html>
```

### Making the API Call

In `src/app.js`, use `lightshell.http.fetch()` to call the GitHub API:

```js
async function searchUser() {
  const username = document.getElementById('username').value.trim()
  if (!username) return

  const result = document.getElementById('result')
  result.innerHTML = '<p>Loading...</p>'

  try {
    const response = await lightshell.http.fetch(
      `https://api.github.com/users/${username}`,
      {
        method: 'GET',
        headers: { 'Accept': 'application/json' }
      }
    )

    if (response.status === 404) {
      result.innerHTML = '<p>User not found.</p>'
      return
    }

    if (response.status !== 200) {
      result.innerHTML = `<p>Error: HTTP ${response.status}</p>`
      return
    }

    const user = JSON.parse(response.body)
    displayUser(user)
  } catch (err) {
    result.innerHTML = `<p>Request failed: ${err.message}</p>`
  }
}
```

Notice that `response.body` is a string. You parse it yourself with `JSON.parse()`. This is different from the browser `fetch()` where you call `response.json()`. The trade-off is simplicity — no streaming, no response object methods, just a plain string.

### Displaying the Results

```js
function displayUser(user) {
  document.getElementById('result').innerHTML = `
    <div class="user-card">
      <img src="${user.avatar_url}" alt="${user.login}" width="80">
      <div>
        <h2>${user.name || user.login}</h2>
        <p>${user.bio || 'No bio'}</p>
        <p>${user.public_repos} public repos &middot; ${user.followers} followers</p>
        <p>${user.location || ''}</p>
      </div>
    </div>
  `
}
```

### Wiring Up Events

```js
document.getElementById('search-btn').addEventListener('click', searchUser)
document.getElementById('username').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') searchUser()
})
```

Run `lightshell dev`, type a GitHub username, and hit Search. The profile loads instantly — no CORS errors, no proxy server, no browser extensions.

## Adding Authentication Headers

Many APIs require authentication. Pass headers just like you would with `fetch()`:

```js
const response = await lightshell.http.fetch('https://api.github.com/user', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ghp_your_token_here',
    'Accept': 'application/json'
  }
})

const user = JSON.parse(response.body)
console.log(user.login) // your authenticated GitHub username
```

For a real app, you would store the API token using `lightshell.store.set('github_token', token)` and load it on startup, rather than hardcoding it.

## POST Requests

Sending data works the same way. Stringify the body yourself:

```js
const response = await lightshell.http.fetch('https://api.example.com/items', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer my_token'
  },
  body: JSON.stringify({ name: 'New Item', priority: 'high' })
})

const created = JSON.parse(response.body)
console.log(created.id) // ID of the newly created item
```

## Setting a Timeout

By default, requests time out after 30 seconds. You can change this per request:

```js
const response = await lightshell.http.fetch('https://slow-api.example.com/data', {
  method: 'GET',
  timeout: 60000 // 60 seconds
})
```

## Downloading Files

For large files, you do not want to load the entire response into memory as a string. Use `lightshell.http.download()` to save directly to disk:

```js
await lightshell.http.download('https://example.com/report.pdf', {
  saveTo: '$DOWNLOADS/report.pdf',
  onProgress: (p) => {
    console.log(`${p.percent}% downloaded`)
    document.getElementById('progress').style.width = `${p.percent}%`
  }
})

await lightshell.dialog.message('Done', 'Download complete!')
```

The `saveTo` path supports LightShell path variables like `$DOWNLOADS`, `$DESKTOP`, and `$APP_DATA`. The download happens entirely in the Go backend — only progress events are sent to JavaScript.

## Comparing with fetch()

Here is the same request using browser `fetch()` vs `lightshell.http.fetch()`:

```js
// Browser fetch() — blocked by CORS in a webview
const res = await fetch('https://api.github.com/users/torvalds')
const data = await res.json()

// lightshell.http.fetch() — works everywhere, no CORS
const res = await lightshell.http.fetch('https://api.github.com/users/torvalds')
const data = JSON.parse(res.body)
```

The differences:
- `lightshell.http.fetch()` goes through the Go backend, bypassing CORS entirely
- The response body is always a string, not a `Response` object
- Binary responses (images, PDFs) are base64-encoded in the body
- There is no streaming — the full response is returned at once
- For large file downloads, use `lightshell.http.download()` instead

## Restricting Allowed URLs

In permissive mode (the default), your app can call any URL. If you want to lock down which APIs your app can reach, add an `http` section to your permissions in `lightshell.json`:

```json
{
  "permissions": {
    "mode": "restricted",
    "http": {
      "allow": [
        "https://api.github.com/**",
        "https://api.example.com/**"
      ]
    }
  }
}
```

In restricted mode, requests to URLs not matching an `allow` pattern will be rejected with a clear error message explaining which URLs are permitted.

## Putting It Together

Here is the complete `app.js` for the GitHub lookup tool, including error handling, loading states, and persisting the last searched username:

```js
async function searchUser() {
  const input = document.getElementById('username')
  const username = input.value.trim()
  if (!username) return

  const result = document.getElementById('result')
  const btn = document.getElementById('search-btn')

  result.innerHTML = '<p>Loading...</p>'
  btn.disabled = true

  try {
    const response = await lightshell.http.fetch(
      `https://api.github.com/users/${username}`,
      { method: 'GET', headers: { 'Accept': 'application/json' } }
    )

    if (response.status === 404) {
      result.innerHTML = '<p>User not found.</p>'
      return
    }

    const user = JSON.parse(response.body)
    result.innerHTML = `
      <div class="user-card">
        <img src="${user.avatar_url}" width="80">
        <div>
          <h2>${user.name || user.login}</h2>
          <p>${user.bio || 'No bio'}</p>
          <p>${user.public_repos} repos &middot; ${user.followers} followers</p>
        </div>
      </div>
    `

    // Remember the last search
    await lightshell.store.set('lastSearch', username)
  } catch (err) {
    result.innerHTML = `<p>Request failed: ${err.message}</p>`
  } finally {
    btn.disabled = false
  }
}

// Restore last search on startup
async function init() {
  const last = await lightshell.store.get('lastSearch')
  if (last) {
    document.getElementById('username').value = last
  }
}

document.getElementById('search-btn').addEventListener('click', searchUser)
document.getElementById('username').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') searchUser()
})

init()
```

Notice how `lightshell.store` from the previous tutorial and `lightshell.http` work together naturally. Persistence and networking are the two most common needs for desktop apps, and both are one-line operations.

## Recap

You have learned how to:

1. **Make HTTP requests** with `lightshell.http.fetch()` — no CORS restrictions
2. **Parse responses** — the body is a string, use `JSON.parse()` for JSON APIs
3. **Add headers** for authentication and content types
4. **Download files** to disk with `lightshell.http.download()` and progress tracking
5. **Restrict URLs** in production using the permissions config
6. **Combine APIs** — using `lightshell.store` and `lightshell.http` together

Your LightShell app can now persist data locally and communicate with any external API, which covers the needs of most desktop applications.
