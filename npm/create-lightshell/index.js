#!/usr/bin/env node

// npm create lightshell@latest my-app
// Creates a new LightShell project without requiring lightshell to be installed.

const { execSync, spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const name = process.argv[2];

if (!name) {
  console.log(`
Usage: npm create lightshell@latest <app-name>

Example:
  npm create lightshell@latest my-app
  cd my-app
  npx lightshell dev
`);
  process.exit(1);
}

// Validate name
if (!/^[a-z0-9-]+$/.test(name)) {
  console.error(
    `Error: app name must contain only lowercase letters, numbers, and hyphens.`
  );
  process.exit(1);
}

const dir = path.resolve(name);
if (fs.existsSync(dir)) {
  console.error(`Error: directory "${name}" already exists.`);
  process.exit(1);
}

const title = name
  .split("-")
  .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
  .join(" ");

// Create project structure
fs.mkdirSync(path.join(dir, "src"), { recursive: true });

// lightshell.json
fs.writeFileSync(
  path.join(dir, "lightshell.json"),
  JSON.stringify(
    {
      name,
      version: "1.0.0",
      entry: "src/index.html",
      window: {
        title,
        width: 1024,
        height: 768,
        minWidth: 400,
        minHeight: 300,
        resizable: true,
      },
      permissions: ["fs", "dialog", "clipboard", "shell", "notification"],
      build: {
        appId: `com.lightshell.${name}`,
      },
    },
    null,
    2
  ) + "\n"
);

// index.html
fs.writeFileSync(
  path.join(dir, "src", "index.html"),
  `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${title}</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <main>
    <h1>${title}</h1>
    <p>Edit <code>src/</code> to get started.</p>
    <div id="info"></div>
  </main>
  <script src="app.js"></script>
</body>
</html>
`
);

// style.css
fs.writeFileSync(
  path.join(dir, "src", "style.css"),
  `* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
               Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  background: #f5f5f7;
  color: #1d1d1f;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}

main { text-align: center; padding: 2rem; }
h1 { font-size: 2rem; font-weight: 600; margin-bottom: 0.5rem; }
p { color: #6e6e73; margin-bottom: 1.5rem; }
code { background: #e8e8ed; padding: 2px 6px; border-radius: 3px; font-size: 0.875em; }
#info { font-family: monospace; color: #86868b; font-size: 0.875rem; }
`
);

// app.js
fs.writeFileSync(
  path.join(dir, "src", "app.js"),
  `async function init() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  document.getElementById('info').textContent = \`Running on \${platform}/\${arch}\`
}

init()
`
);

console.log(`
Created ${name}!

  cd ${name}
  npx lightshell dev

Happy building!
`);
