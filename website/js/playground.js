// playground.js — Code editor setup, mock lightshell API, examples
(() => {
  // Examples database
  const examples = {
    'hello-world': {
      name: 'Hello World',
      html: `<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: -apple-system, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      background: #f5f5f7;
      color: #1d1d1f;
      margin: 0;
    }
    main { text-align: center; }
    h1 { font-size: 2rem; margin-bottom: 0.5rem; }
    p { color: #6e6e73; }
    #info {
      margin-top: 1rem;
      font-family: monospace;
      font-size: 0.875rem;
      background: #e8e8ed;
      padding: 8px 16px;
      border-radius: 6px;
      color: #86868b;
    }
  </style>
</head>
<body>
  <main>
    <h1>Hello, LightShell!</h1>
    <p>Your desktop app is running.</p>
    <div id="info"></div>
  </main>
  <script>
    async function init() {
      const platform = await lightshell.system.platform()
      const arch = await lightshell.system.arch()
      document.getElementById('info').textContent =
        'Running on ' + platform + '/' + arch
    }
    init()
  </script>
</body>
</html>`,
    },
    'todo-app': {
      name: 'Todo App',
      html: `<!DOCTYPE html>
<html>
<head>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, sans-serif; background: #f5f5f7; padding: 2rem; }
    h1 { font-size: 1.5rem; margin-bottom: 1rem; color: #1d1d1f; }
    .input-row { display: flex; gap: 8px; margin-bottom: 1rem; }
    input {
      flex: 1; padding: 10px 12px; border: 1px solid #d2d2d7; border-radius: 8px;
      font-size: 0.9rem; outline: none; font-family: inherit;
    }
    input:focus { border-color: #1a1a1a; }
    button {
      padding: 10px 20px; background: #1a1a1a; color: white; border: none;
      border-radius: 8px; cursor: pointer; font-size: 0.9rem; font-family: inherit;
    }
    button:hover { background: #333; }
    .todo-list { list-style: none; }
    .todo-item {
      display: flex; align-items: center; gap: 10px; padding: 10px 12px;
      background: white; border-radius: 8px; margin-bottom: 6px;
      border: 1px solid rgba(0,0,0,0.06);
    }
    .todo-item.done span { text-decoration: line-through; color: #86868b; }
    .todo-item input[type="checkbox"] { width: 18px; height: 18px; cursor: pointer; }
    .todo-item span { flex: 1; font-size: 0.9rem; }
    .todo-item .delete { background: none; color: #ff6b6b; padding: 4px 8px; font-size: 0.8rem; }
    .count { font-size: 0.8rem; color: #86868b; margin-top: 8px; }
  </style>
</head>
<body>
  <h1>Todo</h1>
  <div class="input-row">
    <input id="input" placeholder="Add a task..." autofocus>
    <button onclick="addTodo()">Add</button>
  </div>
  <ul class="todo-list" id="list"></ul>
  <div class="count" id="count"></div>
  <script>
    let todos = []

    document.getElementById('input').addEventListener('keydown', e => {
      if (e.key === 'Enter') addTodo()
    })

    function addTodo() {
      const input = document.getElementById('input')
      const text = input.value.trim()
      if (!text) return
      todos.push({ text, done: false })
      input.value = ''
      render()
    }

    function toggle(i) {
      todos[i].done = !todos[i].done
      render()
    }

    function remove(i) {
      todos.splice(i, 1)
      render()
    }

    function render() {
      const list = document.getElementById('list')
      list.innerHTML = todos.map((t, i) =>
        '<li class="todo-item' + (t.done ? ' done' : '') + '">' +
        '<input type="checkbox" ' + (t.done ? 'checked' : '') + ' onchange="toggle(' + i + ')">' +
        '<span>' + t.text + '</span>' +
        '<button class="delete" onclick="remove(' + i + ')">remove</button>' +
        '</li>'
      ).join('')
      const remaining = todos.filter(t => !t.done).length
      document.getElementById('count').textContent =
        remaining + ' task' + (remaining !== 1 ? 's' : '') + ' remaining'
    }
    render()
  </script>
</body>
</html>`,
    },
    'markdown-previewer': {
      name: 'Markdown Previewer',
      html: `<!DOCTYPE html>
<html>
<head>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, sans-serif; display: flex; height: 100vh; }
    .pane { flex: 1; display: flex; flex-direction: column; }
    .pane-header {
      padding: 8px 16px; font-size: 0.75rem; font-weight: 600;
      text-transform: uppercase; letter-spacing: 0.04em;
      color: #86868b; background: #f5f5f7; border-bottom: 1px solid #e8e8ed;
    }
    textarea {
      flex: 1; padding: 16px; border: none; outline: none; resize: none;
      font-family: 'SF Mono', monospace; font-size: 0.875rem; line-height: 1.6;
      background: #fafafa;
    }
    .divider { width: 1px; background: #e8e8ed; }
    #preview {
      flex: 1; padding: 16px; overflow: auto; font-size: 0.9rem; line-height: 1.65;
    }
    #preview h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
    #preview h2 { font-size: 1.25rem; margin: 1rem 0 0.5rem; }
    #preview p { margin-bottom: 0.75rem; }
    #preview code { background: #f0f0f3; padding: 2px 6px; border-radius: 3px; font-size: 0.85em; }
    #preview ul { padding-left: 1.5rem; margin-bottom: 0.75rem; }
    #preview li { margin-bottom: 0.25rem; }
  </style>
</head>
<body>
  <div class="pane">
    <div class="pane-header">Markdown</div>
    <textarea id="editor" spellcheck="false"># Hello, Markdown!

This is a **live preview** of your markdown.

## Features
- Real-time rendering
- Clean, native-feeling UI
- Built with LightShell

## Code
Use \`lightshell.fs\` to open and save files:

\`lightshell.dialog.open()\` - native file picker
\`lightshell.fs.readFile(path)\` - read file contents
\`lightshell.fs.writeFile(path, data)\` - save to disk</textarea>
  </div>
  <div class="divider"></div>
  <div class="pane">
    <div class="pane-header">Preview</div>
    <div id="preview"></div>
  </div>
  <script>
    const editor = document.getElementById('editor')
    const preview = document.getElementById('preview')

    function renderMarkdown(md) {
      return md
        .replace(/^### (.+)$/gm, '<h3>$1</h3>')
        .replace(/^## (.+)$/gm, '<h2>$1</h2>')
        .replace(/^# (.+)$/gm, '<h1>$1</h1>')
        .replace(/\\*\\*(.+?)\\*\\*/g, '<strong>$1</strong>')
        .replace(/\\*(.+?)\\*/g, '<em>$1</em>')
        .replace(/\\\`([^\\\`]+)\\\`/g, '<code>$1</code>')
        .replace(/^- (.+)$/gm, '<li>$1</li>')
        .replace(/(<li>.*<\\/li>)/s, '<ul>$1</ul>')
        .replace(/^(?!<[hulo])(\\S.+)$/gm, '<p>$1</p>')
        .replace(/&mdash;/g, '\\u2014')
    }

    function update() {
      preview.innerHTML = renderMarkdown(editor.value)
    }

    editor.addEventListener('input', update)
    update()
  </script>
</body>
</html>`,
    },
    'color-picker': {
      name: 'Color Picker',
      html: `<!DOCTYPE html>
<html>
<head>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, sans-serif; background: #f5f5f7;
      display: flex; align-items: center; justify-content: center;
      min-height: 100vh; padding: 2rem;
    }
    .card {
      background: white; border-radius: 16px; padding: 2rem;
      box-shadow: 0 4px 24px rgba(0,0,0,0.08); width: 320px;
    }
    h1 { font-size: 1.25rem; margin-bottom: 1.5rem; text-align: center; }
    .preview {
      width: 100%; height: 120px; border-radius: 12px; margin-bottom: 1rem;
      transition: background 0.2s;
    }
    .sliders { display: flex; flex-direction: column; gap: 12px; margin-bottom: 1rem; }
    .slider-row { display: flex; align-items: center; gap: 8px; }
    .slider-row label { width: 16px; font-weight: 600; font-size: 0.875rem; }
    .slider-row input[type="range"] { flex: 1; }
    .slider-row .value { width: 32px; text-align: right; font-family: monospace; font-size: 0.8rem; color: #86868b; }
    .hex-row {
      display: flex; align-items: center; justify-content: center; gap: 8px;
      padding: 12px; background: #f5f5f7; border-radius: 8px;
    }
    .hex-value { font-family: monospace; font-size: 1.1rem; font-weight: 600; }
    .copy-btn {
      padding: 6px 12px; background: #1a1a1a; color: white; border: none;
      border-radius: 6px; cursor: pointer; font-size: 0.8rem; font-family: inherit;
    }
    .copy-btn:hover { background: #333; }
    .copied { background: #00b894 !important; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Color Picker</h1>
    <div class="preview" id="preview"></div>
    <div class="sliders">
      <div class="slider-row">
        <label style="color: #ff6b6b">R</label>
        <input type="range" id="r" min="0" max="255" value="108">
        <span class="value" id="rv">108</span>
      </div>
      <div class="slider-row">
        <label style="color: #00b894">G</label>
        <input type="range" id="g" min="0" max="255" value="92">
        <span class="value" id="gv">92</span>
      </div>
      <div class="slider-row">
        <label style="color: #0984e3">B</label>
        <input type="range" id="b" min="0" max="255" value="231">
        <span class="value" id="bv">231</span>
      </div>
    </div>
    <div class="hex-row">
      <span class="hex-value" id="hex">#6C5CE7</span>
      <button class="copy-btn" id="copyBtn" onclick="copyColor()">Copy</button>
    </div>
  </div>
  <script>
    const rs = document.getElementById('r')
    const gs = document.getElementById('g')
    const bs = document.getElementById('b')

    function update() {
      const r = rs.value, g = gs.value, b = bs.value
      document.getElementById('rv').textContent = r
      document.getElementById('gv').textContent = g
      document.getElementById('bv').textContent = b
      const hex = '#' + [r,g,b].map(v => Number(v).toString(16).padStart(2,'0').toUpperCase()).join('')
      document.getElementById('hex').textContent = hex
      document.getElementById('preview').style.background = hex
    }

    rs.addEventListener('input', update)
    gs.addEventListener('input', update)
    bs.addEventListener('input', update)
    update()

    async function copyColor() {
      const hex = document.getElementById('hex').textContent
      await lightshell.clipboard.write(hex)
      const btn = document.getElementById('copyBtn')
      btn.textContent = 'Copied!'
      btn.classList.add('copied')
      setTimeout(() => { btn.textContent = 'Copy'; btn.classList.remove('copied') }, 1500)
    }
  </script>
</body>
</html>`,
    },
    'system-info': {
      name: 'System Info Dashboard',
      html: `<!DOCTYPE html>
<html>
<head>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: -apple-system, sans-serif; background: #1d1d2e; color: #e0e0e6; padding: 2rem; }
    h1 { font-size: 1.25rem; margin-bottom: 1.5rem; color: #fff; }
    .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
    .card {
      background: rgba(255,255,255,0.06); border-radius: 12px; padding: 16px;
      border: 1px solid rgba(255,255,255,0.08);
    }
    .card .label {
      font-size: 0.6875rem; text-transform: uppercase; letter-spacing: 0.05em;
      color: rgba(255,255,255,0.4); margin-bottom: 6px;
    }
    .card .value { font-size: 1.25rem; font-weight: 600; color: #fff; }
    .card .value.mono { font-family: 'SF Mono', monospace; font-size: 1rem; }
    .accent { color: #1a1a1a; }
    .green { color: #00b894; }
  </style>
</head>
<body>
  <h1>System Info</h1>
  <div class="grid">
    <div class="card">
      <div class="label">Platform</div>
      <div class="value" id="platform">...</div>
    </div>
    <div class="card">
      <div class="label">Architecture</div>
      <div class="value" id="arch">...</div>
    </div>
    <div class="card">
      <div class="label">Hostname</div>
      <div class="value mono" id="hostname">...</div>
    </div>
    <div class="card">
      <div class="label">Home Directory</div>
      <div class="value mono" id="home" style="font-size:0.8rem;word-break:break-all">...</div>
    </div>
    <div class="card">
      <div class="label">Temp Directory</div>
      <div class="value mono" id="temp" style="font-size:0.8rem;word-break:break-all">...</div>
    </div>
    <div class="card">
      <div class="label">App Version</div>
      <div class="value accent" id="version">...</div>
    </div>
  </div>
  <script>
    async function load() {
      document.getElementById('platform').textContent = await lightshell.system.platform()
      document.getElementById('arch').textContent = await lightshell.system.arch()
      document.getElementById('hostname').textContent = await lightshell.system.hostname()
      document.getElementById('home').textContent = await lightshell.system.homeDir()
      document.getElementById('temp').textContent = await lightshell.system.tempDir()
      document.getElementById('version').textContent = await lightshell.app.version()
    }
    load()
  </script>
</body>
</html>`,
    },
  }

  // Mock lightshell API for browser playground
  const mockFS = new Map()
  mockFS.set('/tmp/test.txt', 'Hello from LightShell playground!')

  const mockLightshell = {
    window: {
      setTitle: async (t) => { document.title = t },
      setSize: async () => {},
      getSize: async () => ({ width: 800, height: 600 }),
      setPosition: async () => {},
      getPosition: async () => ({ x: 100, y: 100 }),
      minimize: async () => {},
      maximize: async () => {},
      fullscreen: async () => {},
      restore: async () => {},
      close: async () => { consoleLog('warn', 'window.close() — would close the app') },
      onResize: () => () => {},
      onMove: () => () => {},
      onFocus: () => () => {},
      onBlur: () => () => {},
    },
    fs: {
      readFile: async (path) => {
        if (mockFS.has(path)) return mockFS.get(path)
        throw new Error('ENOENT: ' + path)
      },
      writeFile: async (path, data) => { mockFS.set(path, data) },
      readDir: async () => [
        { name: 'index.html', isDir: false, size: 1024 },
        { name: 'app.js', isDir: false, size: 512 },
        { name: 'style.css', isDir: false, size: 256 },
      ],
      exists: async (path) => mockFS.has(path),
      stat: async (path) => ({ name: path.split('/').pop(), size: 1024, isDir: false, modTime: new Date().toISOString(), mode: '-rw-r--r--' }),
      mkdir: async () => {},
      remove: async () => {},
      watch: () => () => {},
    },
    dialog: {
      open: async () => { consoleLog('info', 'dialog.open() — would show native file picker'); return '/mock/selected-file.txt' },
      save: async () => { consoleLog('info', 'dialog.save() — would show native save dialog'); return '/mock/saved-file.txt' },
      message: async (title, msg) => alert(title + '\n\n' + msg),
      confirm: async (title, msg) => confirm(title + '\n\n' + msg),
      prompt: async (title, def) => prompt(title, def),
    },
    clipboard: {
      read: async () => {
        try { return await navigator.clipboard.readText() }
        catch { return '' }
      },
      write: async (text) => {
        try { await navigator.clipboard.writeText(text) }
        catch { consoleLog('warn', 'Clipboard write blocked by browser') }
      },
    },
    shell: {
      open: async (url) => window.open(url, '_blank'),
    },
    notify: {
      send: async (title, body) => {
        if ('Notification' in window && Notification.permission === 'granted') {
          new Notification(title, { body })
        } else {
          consoleLog('info', 'Notification: ' + title + ' — ' + body)
        }
      },
    },
    tray: {
      set: async () => consoleLog('info', 'tray.set() — not available in browser'),
      remove: async () => {},
      onClick: () => () => {},
    },
    menu: {
      set: async () => consoleLog('info', 'menu.set() — not available in browser'),
    },
    system: {
      platform: async () => navigator.platform.includes('Mac') ? 'darwin' : 'linux',
      arch: async () => 'arm64',
      homeDir: async () => '/Users/demo',
      tempDir: async () => '/tmp',
      hostname: async () => 'playground-demo',
    },
    app: {
      quit: async () => consoleLog('warn', 'app.quit() — would close the app'),
      version: async () => '1.0.0',
      dataDir: async () => '/Users/demo/Library/Application Support/playground',
    },
    on: () => () => {},
  }

  // Console log helper
  const consoleLogs = []
  function consoleLog(level, msg) {
    consoleLogs.push({ level, msg })
    renderConsole()
  }

  function renderConsole() {
    const el = document.querySelector('.playground-console')
    if (!el) return
    el.classList.add('visible')
    el.innerHTML = consoleLogs.map(l =>
      `<div class="console-line ${l.level}">${l.msg}</div>`
    ).join('')
    el.scrollTop = el.scrollHeight
  }

  // Editor setup
  let editor = null
  let currentExample = 'hello-world'

  async function initEditor() {
    const editorEl = document.querySelector('.playground-editor')
    if (!editorEl) return

    // Try to load CodeMirror; fall back to textarea
    if (window.EditorView) {
      setupCodeMirror(editorEl)
    } else {
      setupFallbackEditor(editorEl)
    }

    loadExample('hello-world')
  }

  function setupCodeMirror(container) {
    const { EditorView, basicSetup } = window.CM || {}
    const { javascript } = window.CMLangJS || {}

    if (!EditorView) {
      setupFallbackEditor(container)
      return
    }

    editor = new EditorView({
      extensions: [
        basicSetup,
        javascript ? javascript() : [],
        EditorView.theme({
          '&': { height: '100%' },
          '.cm-scroller': { overflow: 'auto' },
        }),
      ],
      parent: container,
    })
  }

  function setupFallbackEditor(container) {
    const ta = document.createElement('textarea')
    ta.style.cssText = 'width:100%;height:100%;background:#1e1e2e;color:#cdd6f4;border:none;padding:16px;font-family:var(--font-mono);font-size:13px;resize:none;outline:none;tab-size:2;'
    ta.spellcheck = false
    container.appendChild(ta)
    editor = {
      _ta: ta,
      dispatch: (tr) => { if (tr.changes) ta.value = tr.changes },
      state: { doc: { toString: () => ta.value } },
    }
  }

  function getCode() {
    if (editor._ta) return editor._ta.value
    if (editor?.state) return editor.state.doc.toString()
    return ''
  }

  function setCode(code) {
    if (editor._ta) {
      editor._ta.value = code
      return
    }
    if (editor?.dispatch) {
      editor.dispatch({
        changes: { from: 0, to: editor.state.doc.length, insert: code }
      })
    }
  }

  function loadExample(id) {
    const ex = examples[id]
    if (!ex) return
    currentExample = id
    setCode(ex.html)
    consoleLogs.length = 0
    const consoleEl = document.querySelector('.playground-console')
    if (consoleEl) { consoleEl.innerHTML = ''; consoleEl.classList.remove('visible') }

    // Highlight active in sidebar
    document.querySelectorAll('.example-list li').forEach(li => {
      li.classList.toggle('active', li.dataset.id === id)
    })

    runCode()
  }

  function runCode() {
    const code = getCode()
    const iframe = document.querySelector('.preview-frame')
    if (!iframe) return

    // Inject mock lightshell into iframe
    const mockScript = `<script>window.lightshell = window.parent.__playgroundMockLS</` + `script>`
    const html = code.replace('</head>', mockScript + '</head>')

    iframe.srcdoc = html

    // Update estimated size
    const sizeEl = document.querySelector('.toolbar-size')
    if (sizeEl) {
      const kb = new Blob([code]).size / 1024
      const estMB = (2.1 + kb / 1024).toFixed(1)
      sizeEl.textContent = 'Est. binary: ~' + estMB + 'MB'
    }
  }

  // Expose mock API for iframe
  window.__playgroundMockLS = mockLightshell

  // Wire up UI
  document.addEventListener('DOMContentLoaded', () => {
    initEditor()

    // Sidebar example list
    const list = document.querySelector('.example-list')
    if (list) {
      Object.entries(examples).forEach(([id, ex]) => {
        const li = document.createElement('li')
        li.textContent = ex.name
        li.dataset.id = id
        li.onclick = () => loadExample(id)
        list.appendChild(li)
      })
    }

    // Run button
    const runBtn = document.querySelector('.toolbar-btn.run')
    if (runBtn) runBtn.onclick = runCode

    // Keyboard shortcut: Cmd/Ctrl+Enter to run
    document.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault()
        runCode()
      }
    })
  })

  // Expose for inline use
  window.loadExample = loadExample
  window.runCode = runCode
})()
