---
title: Single-File Apps
description: Build a complete LightShell app in a single HTML file.
---

LightShell apps do not require a build step, a bundler, or multiple files. You can build a fully functional desktop app with a single `index.html` file and a minimal `lightshell.json`. This makes LightShell ideal for prototyping, AI-generated apps, and quick utility tools.

## Minimal Project Structure

A single-file app needs just two files:

```
my-app/
  lightshell.json
  src/
    index.html
```

The `lightshell.json` is as simple as it gets:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "entry": "src/index.html"
}
```

Everything else -- styles, scripts, and markup -- lives inside `index.html`.

## Example: Calculator

A working calculator in a single HTML file:

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Calculator</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, sans-serif;
      background: #1c1c1e;
      color: #fff;
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
      user-select: none;
    }

    .calculator {
      width: 280px;
    }

    .display {
      background: #2c2c2e;
      padding: 20px;
      text-align: right;
      font-size: 36px;
      border-radius: 12px 12px 0 0;
      min-height: 80px;
      overflow: hidden;
      word-break: break-all;
    }

    .buttons {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 1px;
    }

    button {
      padding: 20px;
      font-size: 20px;
      border: none;
      cursor: pointer;
      background: #3a3a3c;
      color: #fff;
      transition: background 0.1s;
    }

    button:hover { background: #48484a; }
    button:active { background: #636366; }
    button.op { background: #ff9f0a; color: #000; }
    button.op:hover { background: #ffb340; }
    button.clear { background: #636366; }
    button.equals { background: #30d158; color: #000; border-radius: 0 0 12px 0; }
    button.zero { grid-column: span 2; border-radius: 0 0 0 12px; }
  </style>
</head>
<body>
  <div class="calculator">
    <div class="display" id="display">0</div>
    <div class="buttons">
      <button class="clear" onclick="clearDisplay()">C</button>
      <button class="clear" onclick="toggleSign()">+/-</button>
      <button class="clear" onclick="percent()">%</button>
      <button class="op" onclick="setOp('/')">÷</button>
      <button onclick="input('7')">7</button>
      <button onclick="input('8')">8</button>
      <button onclick="input('9')">9</button>
      <button class="op" onclick="setOp('*')">x</button>
      <button onclick="input('4')">4</button>
      <button onclick="input('5')">5</button>
      <button onclick="input('6')">6</button>
      <button class="op" onclick="setOp('-')">-</button>
      <button onclick="input('1')">1</button>
      <button onclick="input('2')">2</button>
      <button onclick="input('3')">3</button>
      <button class="op" onclick="setOp('+')">+</button>
      <button class="zero" onclick="input('0')">0</button>
      <button onclick="input('.')">.</button>
      <button class="equals" onclick="calculate()">=</button>
    </div>
  </div>

  <script>
    let current = '0'
    let previous = null
    let op = null
    let resetNext = false
    const display = document.getElementById('display')

    function input(ch) {
      if (resetNext) { current = ''; resetNext = false }
      if (ch === '.' && current.includes('.')) return
      if (current === '0' && ch !== '.') current = ''
      current += ch
      display.textContent = current
    }

    function setOp(nextOp) {
      if (previous !== null && !resetNext) calculate()
      previous = parseFloat(current)
      op = nextOp
      resetNext = true
    }

    function calculate() {
      if (previous === null || op === null) return
      const b = parseFloat(current)
      let result
      switch (op) {
        case '+': result = previous + b; break
        case '-': result = previous - b; break
        case '*': result = previous * b; break
        case '/': result = b !== 0 ? previous / b : 'Error'; break
      }
      current = String(result)
      display.textContent = current
      previous = null
      op = null
      resetNext = true
    }

    function clearDisplay() {
      current = '0'
      previous = null
      op = null
      display.textContent = '0'
    }

    function toggleSign() {
      current = String(-parseFloat(current))
      display.textContent = current
    }

    function percent() {
      current = String(parseFloat(current) / 100)
      display.textContent = current
    }
  </script>
</body>
</html>
```

Run it with:

```bash
lightshell dev
```

Build it with:

```bash
lightshell build
```

The output is a native app bundle under 5MB.

## Example: Color Picker

A color utility tool using LightShell APIs to copy the selected color to the clipboard:

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Color Picker</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }

    body {
      font-family: -apple-system, BlinkMacSystemFont, sans-serif;
      background: #f5f5f7;
      padding: 32px;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 24px;
      height: 100vh;
    }

    h1 { font-size: 18px; color: #333; }

    .preview {
      width: 200px;
      height: 200px;
      border-radius: 16px;
      border: 2px solid #ddd;
      transition: background 0.15s;
    }

    input[type="color"] {
      width: 200px;
      height: 44px;
      border: none;
      cursor: pointer;
      border-radius: 8px;
    }

    .values {
      display: flex;
      flex-direction: column;
      gap: 8px;
      width: 200px;
    }

    .value-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #fff;
      padding: 8px 12px;
      border-radius: 8px;
      cursor: pointer;
      border: 1px solid #e0e0e0;
      transition: background 0.1s;
    }

    .value-row:hover { background: #eef; }

    .label { font-size: 12px; color: #888; text-transform: uppercase; }
    .code { font-family: monospace; font-size: 14px; color: #333; }
    .hint { font-size: 12px; color: #aaa; }
  </style>
</head>
<body>
  <h1>Color Picker</h1>
  <div class="preview" id="preview"></div>
  <input type="color" id="picker" value="#3b82f6">
  <div class="values" id="values"></div>
  <div class="hint">Click a value to copy it</div>

  <script>
    const picker = document.getElementById('picker')
    const preview = document.getElementById('preview')
    const values = document.getElementById('values')

    function hexToRgb(hex) {
      const r = parseInt(hex.slice(1, 3), 16)
      const g = parseInt(hex.slice(3, 5), 16)
      const b = parseInt(hex.slice(5, 7), 16)
      return { r, g, b }
    }

    function rgbToHsl(r, g, b) {
      r /= 255; g /= 255; b /= 255
      const max = Math.max(r, g, b), min = Math.min(r, g, b)
      let h, s, l = (max + min) / 2
      if (max === min) { h = s = 0 }
      else {
        const d = max - min
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
        switch (max) {
          case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break
          case g: h = ((b - r) / d + 2) / 6; break
          case b: h = ((r - g) / d + 4) / 6; break
        }
      }
      return {
        h: Math.round(h * 360),
        s: Math.round(s * 100),
        l: Math.round(l * 100)
      }
    }

    function update() {
      const hex = picker.value
      const { r, g, b } = hexToRgb(hex)
      const { h, s, l } = rgbToHsl(r, g, b)

      preview.style.background = hex

      const formats = [
        { label: 'HEX', value: hex.toUpperCase() },
        { label: 'RGB', value: `rgb(${r}, ${g}, ${b})` },
        { label: 'HSL', value: `hsl(${h}, ${s}%, ${l}%)` },
      ]

      values.innerHTML = formats.map(f => `
        <div class="value-row" onclick="copy('${f.value}')">
          <span class="label">${f.label}</span>
          <span class="code">${f.value}</span>
        </div>
      `).join('')
    }

    async function copy(text) {
      await lightshell.clipboard.write(text)
      await lightshell.notify.send({
        title: 'Copied',
        body: text
      })
    }

    picker.addEventListener('input', update)
    update()
  </script>
</body>
</html>
```

## When to Use Single-File vs Multi-File

### Single-file is great for:

- **Prototyping.** Get an idea running in minutes with no setup.
- **AI-generated apps.** An AI can produce one complete HTML file more reliably than coordinating multiple files.
- **Small utilities.** Calculators, converters, timers, color pickers, quick-reference tools.
- **Learning.** Everything is in one place, easy to understand.

### Switch to multi-file when:

- **Your CSS exceeds ~200 lines.** Extract it into a separate `.css` file for readability.
- **Your JavaScript exceeds ~300 lines.** Split into modules or separate `.js` files.
- **You have images or assets.** Put them in a `src/assets/` directory.
- **You want to use a CSS framework.** Link to a local CSS file or use a CDN.
- **Multiple people are working on the project.** Separate files make collaboration easier.

## Multi-File Equivalent

When you outgrow a single file, the transition is straightforward. Split the inline code into separate files:

```
my-app/
  lightshell.json
  src/
    index.html
    style.css
    app.js
```

```html
<!-- src/index.html -->
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>My App</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <div id="app"></div>
  <script src="app.js"></script>
</body>
</html>
```

No config changes needed. The `entry` still points to `src/index.html`, and relative paths to CSS and JS files work as expected.

## Tips

**Use inline `<style>` freely.** LightShell's CSP allows `unsafe-inline` for styles in both dev and production. No need to extract CSS unless you want to.

**Inline `<script>` works in dev mode** but is blocked by the default production CSP. For production single-file apps, either move your script to a separate `.js` file, or customize the CSP in `lightshell.json`:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
  }
}
```

**Keep the HTML file under 500 lines.** Beyond that, readability suffers and a multi-file structure becomes worth the small overhead.

**Use CDN libraries when needed.** For a quick prototype, load a library from a CDN directly in your HTML. Update the CSP to allow it:

```json
{
  "security": {
    "csp": "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'"
  }
}
```

```html
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
```

**File size comparison.** A single-file calculator app compiles to the same ~5MB app bundle as a multi-file app. The HTML/CSS/JS is embedded in the binary either way. The difference is purely in developer experience, not output size.
