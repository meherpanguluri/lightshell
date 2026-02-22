---
title: Deep Linking
description: Register custom URL protocols for your app.
---

Deep linking lets other apps and websites launch your LightShell app by opening a URL like `myapp://open/doc/123`. When the URL is opened, your app starts (or comes to the foreground if already running) and receives the full URL in JavaScript.

## Configuration

Register your custom URL scheme in `lightshell.json`:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html",
  "protocols": {
    "schemes": ["myapp"]
  }
}
```

You can register multiple schemes:

```json
{
  "protocols": {
    "schemes": ["myapp", "myapp-dev"]
  }
}
```

The scheme name should be lowercase, use only letters, numbers, and hyphens. Avoid generic names like `http`, `mailto`, or `file` -- these are reserved by the OS.

## Handling URLs in JavaScript

Use `lightshell.app.onProtocol()` to receive URLs when your app is opened via a deep link:

```js
lightshell.app.onProtocol((url) => {
  console.log('Opened with URL:', url)
  // url = "myapp://open/doc/123"

  const parsed = new URL(url)
  console.log(parsed.hostname) // "open"
  console.log(parsed.pathname) // "/doc/123"

  // Route to the appropriate view
  navigateTo(parsed.pathname)
})
```

The callback fires in two scenarios:

1. **App was not running** -- the app launches, and the callback fires after the page loads
2. **App was already running** -- the app comes to the foreground, and the callback fires immediately

### Parsing the URL

Deep link URLs follow the standard URL format:

```
myapp://action/parameter1/parameter2?key=value
```

Use the built-in `URL` constructor to parse them:

```js
lightshell.app.onProtocol((url) => {
  const parsed = new URL(url)

  const action = parsed.hostname      // "action"
  const path = parsed.pathname        // "/parameter1/parameter2"
  const params = parsed.searchParams  // URLSearchParams

  switch (action) {
    case 'open':
      openDocument(path.slice(1))  // remove leading slash
      break
    case 'settings':
      openSettings(params.get('tab'))
      break
    case 'auth':
      handleOAuthCallback(params.get('code'))
      break
  }
})
```

## How It Works

### macOS

LightShell registers the URL scheme in the app bundle's `Info.plist` under `CFBundleURLTypes`:

```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleURLName</key>
    <string>com.example.myapp</string>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>myapp</string>
    </array>
  </dict>
</array>
```

When the OS receives a `myapp://` URL, it launches the `.app` bundle and delivers the URL through the `application:openURLs:` delegate method. The Go backend forwards the URL to JavaScript via the IPC bridge.

The registration happens automatically when the app is built with `lightshell build`. During development (`lightshell dev`), the scheme is registered temporarily and may not persist across system restarts.

### Linux

On Linux, the URL scheme is registered in the `.desktop` file:

```ini
[Desktop Entry]
Name=My App
Exec=/usr/bin/my-app %u
Type=Application
MimeType=x-scheme-handler/myapp
```

The `%u` argument in the `Exec` line tells the desktop environment to pass the URL as a command-line argument. LightShell reads the URL from `os.Args` and delivers it to JavaScript.

For `.deb` and `.rpm` packages, the `.desktop` file is included automatically. For AppImage, the scheme is registered when the user runs the app for the first time.

## Examples

### Open a Specific Document

A notes app that opens a specific note when a link is clicked in a browser:

```js
// In your app
lightshell.app.onProtocol(async (url) => {
  const parsed = new URL(url)

  if (parsed.hostname === 'note') {
    const noteId = parsed.pathname.slice(1)
    const note = await lightshell.store.get(`notes.${noteId}`)

    if (note) {
      showNote(note)
    } else {
      await lightshell.dialog.message('Not Found', `Note ${noteId} does not exist.`)
    }
  }
})
```

Link from a website or another app:

```html
<a href="myapp://note/abc123">Open in My App</a>
```

### OAuth Callback

Handle OAuth2 authorization code flow by using a deep link as the redirect URI:

```js
// Step 1: Open the OAuth page in the browser
const clientId = 'your-client-id'
const redirectUri = encodeURIComponent('myapp://auth/callback')
const authUrl = `https://github.com/login/oauth/authorize?client_id=${clientId}&redirect_uri=${redirectUri}`

await lightshell.shell.open(authUrl)

// Step 2: Handle the callback
lightshell.app.onProtocol(async (url) => {
  const parsed = new URL(url)

  if (parsed.hostname === 'auth' && parsed.pathname === '/callback') {
    const code = parsed.searchParams.get('code')

    // Exchange the code for an access token
    const res = await lightshell.http.fetch('https://github.com/login/oauth/access_token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify({
        client_id: clientId,
        client_secret: 'your-client-secret',
        code: code
      })
    })

    const data = JSON.parse(res.body)
    await lightshell.store.set('github.token', data.access_token)
    showLoggedInState()
  }
})
```

### Share Content Into Your App

A link sharing flow where another app sends content to your app:

```js
lightshell.app.onProtocol((url) => {
  const parsed = new URL(url)

  if (parsed.hostname === 'share') {
    const text = parsed.searchParams.get('text')
    const title = parsed.searchParams.get('title')

    createNewEntry({ title, text })
  }
})
```

Shareable link format: `myapp://share?title=Hello&text=World`

## Testing Deep Links

### macOS

Open a terminal and run:

```bash
open "myapp://open/doc/123"
```

### Linux

```bash
xdg-open "myapp://open/doc/123"
```

### During Development

Deep links work with `lightshell dev`, but the scheme registration may not survive system restarts. For reliable testing during development, build the app with `lightshell build` and open the resulting `.app` bundle or AppImage at least once to register the scheme.

## Best Practices

**Validate all URLs.** Treat deep link URLs as untrusted input. An attacker could craft a malicious URL to exploit your app:

```js
lightshell.app.onProtocol((url) => {
  try {
    const parsed = new URL(url)
    // Validate hostname, path, and parameters before acting
    if (!['open', 'settings', 'auth'].includes(parsed.hostname)) {
      console.warn('Unknown deep link action:', parsed.hostname)
      return
    }
    handleDeepLink(parsed)
  } catch (err) {
    console.error('Invalid deep link URL:', url)
  }
})
```

**Choose a unique scheme name.** If two apps register the same scheme, the OS behavior is unpredictable. Use your app name or a name based on your domain (e.g., `com-example-myapp`).

**Handle the case where the app is already running.** When a deep link opens while your app is in the foreground, the `onProtocol` callback fires immediately. Make sure your app can handle being navigated to a different view at any time.

**Provide a fallback for users who do not have your app installed.** On websites, consider using a landing page that checks if the app is installed and falls back to a web version or download link:

```html
<a href="myapp://note/abc123"
   onclick="setTimeout(() => { window.location = 'https://myapp.com/download'; }, 500)">
  Open in My App
</a>
```
