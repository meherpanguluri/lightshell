---
title: "1. Your First App"
description: Build a hello world desktop app with LightShell from scratch.
---

In this tutorial, you will create a simple desktop app, run it in development mode, and see live changes. By the end, you will understand the basic LightShell project structure and workflow.

## Scaffold the Project

Open your terminal and run:

```bash
lightshell init hello-world
cd hello-world
```

This creates a project with three files in `src/` and a configuration file.

## Explore the Project Structure

### lightshell.json

This is your app's configuration file:

```json
{
  "name": "hello-world",
  "version": "1.0.0",
  "entry": "src/index.html",
  "window": {
    "title": "Hello World",
    "width": 1024,
    "height": 768,
    "minWidth": 400,
    "minHeight": 300,
    "resizable": true,
    "frameless": false
  },
  "build": {
    "icon": "assets/icon.png",
    "appId": "com.example.hello-world"
  }
}
```

Key fields:
- **entry** — the HTML file loaded when the app starts
- **window** — initial window dimensions and behavior
- **build** — packaging settings like the app icon and identifier

### src/index.html

The entry point for your app:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Hello World</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <h1>Hello, LightShell!</h1>
  <p>Edit src/index.html and save to see changes.</p>
  <script src="app.js"></script>
</body>
</html>
```

This is standard HTML. Nothing special about it. LightShell injects its runtime scripts automatically before your code runs, so `window.lightshell` is available in `app.js` without any imports.

### src/app.js

Your application logic:

```js
// All lightshell APIs are available via window.lightshell
// They are all async — always use await

async function init() {
  const platform = await lightshell.system.platform()
  console.log(`Running on ${platform}`)
}

init()
```

### src/style.css

Standard CSS. LightShell injects a normalization layer before your styles, so form elements and scrollbars look consistent across macOS and Linux.

## Run the App

```bash
lightshell dev
```

A native window opens showing your HTML. The window title, size, and other properties come from `lightshell.json`.

## Make Changes

With the app still running, open `src/index.html` in your editor. Change the heading:

```html
<h1>My First Desktop App</h1>
```

Save the file. The app reloads automatically — you should see the new heading within a fraction of a second. This is hot reload in action. It watches every file in `src/` and triggers a refresh when anything changes.

## Add Interactivity

Edit `src/app.js` to add a button that shows the current platform:

```js
async function init() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  const hostname = await lightshell.system.hostname()

  document.body.innerHTML = `
    <h1>My First Desktop App</h1>
    <p>Platform: ${platform} (${arch})</p>
    <p>Hostname: ${hostname}</p>
    <button id="greet">Say Hello</button>
    <div id="output"></div>
  `

  document.getElementById('greet').addEventListener('click', async () => {
    const homeDir = await lightshell.system.homeDir()
    document.getElementById('output').textContent = `Hello from ${homeDir}!`
  })
}

init()
```

Save and watch the app update. Click the button to see your home directory path displayed — this is a native API call, not available in a regular browser.

## Recap

You have learned how to:

1. **Scaffold** a project with `lightshell init`
2. **Run** the development server with `lightshell dev`
3. **Edit** files and see hot reload in action
4. **Call native APIs** from JavaScript using `window.lightshell`

Next, we will explore the native APIs in depth.
