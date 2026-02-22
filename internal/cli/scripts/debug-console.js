(() => {
  if (window.__lightshell_debug) return

  // --- State ---
  const MAX_ENTRIES = 1000
  const logs = []
  const errors = []
  const ipcCalls = []
  let activeTab = 'console'
  let panelVisible = false
  let panelHeight = 300

  // --- Console interception ---
  const origConsole = {}
  ;['log', 'warn', 'error', 'info', 'debug'].forEach(level => {
    origConsole[level] = console[level].bind(console)
    console[level] = function(...args) {
      origConsole[level](...args)
      addLog(level, args)
    }
  })
  const origClear = console.clear ? console.clear.bind(console) : () => {}
  console.clear = function() {
    origClear()
    logs.length = 0
    renderActive()
  }

  // --- Error interception ---
  window.addEventListener('error', (e) => {
    errors.push({
      time: Date.now(),
      message: e.message,
      source: e.filename,
      line: e.lineno,
      col: e.colno,
      stack: e.error ? e.error.stack : ''
    })
    renderActive()
  })
  window.addEventListener('unhandledrejection', (e) => {
    const reason = e.reason
    errors.push({
      time: Date.now(),
      message: reason instanceof Error ? reason.message : String(reason),
      source: 'Promise',
      line: 0,
      col: 0,
      stack: reason instanceof Error ? reason.stack : ''
    })
    renderActive()
  })

  // --- IPC interception ---
  const origPostMessage = window.webkit &&
    window.webkit.messageHandlers &&
    window.webkit.messageHandlers.lightshell
    ? window.webkit.messageHandlers.lightshell.postMessage.bind(window.webkit.messageHandlers.lightshell)
    : null

  if (origPostMessage) {
    window.webkit.messageHandlers.lightshell.postMessage = function(raw) {
      let parsed
      try { parsed = JSON.parse(raw) } catch(e) { parsed = {} }
      const entry = {
        id: parsed.id,
        method: parsed.method || '?',
        params: parsed.params,
        time: Date.now(),
        result: null,
        error: null,
        duration: null
      }
      ipcCalls.push(entry)
      if (ipcCalls.length > MAX_ENTRIES) ipcCalls.shift()
      renderActive()
      return origPostMessage(raw)
    }
  }

  const origReceive = window.__lightshell_receive
  if (origReceive) {
    window.__lightshell_receive = function(json) {
      let msg = typeof json === 'string' ? JSON.parse(json) : json
      if (msg.id) {
        const entry = ipcCalls.find(e => e.id === msg.id)
        if (entry) {
          entry.duration = Date.now() - entry.time
          entry.result = msg.result !== undefined ? msg.result : null
          entry.error = msg.error || null
          renderActive()
        }
      }
      return origReceive(json)
    }
  }

  // --- Helpers ---
  function addLog(level, args) {
    logs.push({
      time: Date.now(),
      level,
      message: args.map(a => {
        if (a === null) return 'null'
        if (a === undefined) return 'undefined'
        if (typeof a === 'object') {
          try { return JSON.stringify(a, null, 2) } catch(e) { return String(a) }
        }
        return String(a)
      }).join(' ')
    })
    if (logs.length > MAX_ENTRIES) logs.shift()
    renderActive()
  }

  function ts(t) {
    const d = new Date(t)
    return d.toTimeString().split(' ')[0] + '.' + String(d.getMilliseconds()).padStart(3, '0')
  }

  function esc(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  }

  function truncate(s, max) {
    if (!s || s.length <= max) return s || ''
    return s.slice(0, max) + '...'
  }

  // --- UI ---
  function createPanel() {
    const el = document.createElement('div')
    el.id = '__ls-debug'
    el.innerHTML = `
      <div class="__ls-resize"></div>
      <div class="__ls-tabs">
        <button class="__ls-tab __ls-active" data-tab="console">Console</button>
        <button class="__ls-tab" data-tab="errors">Errors <span class="__ls-badge" id="__ls-err-badge">0</span></button>
        <button class="__ls-tab" data-tab="ipc">IPC <span class="__ls-badge" id="__ls-ipc-badge">0</span></button>
        <button class="__ls-tab" data-tab="info">Info</button>
        <div class="__ls-spacer"></div>
        <button class="__ls-tab __ls-clear-btn" id="__ls-clear">Clear</button>
        <button class="__ls-tab __ls-close-btn" id="__ls-close">&times;</button>
      </div>
      <div class="__ls-content" id="__ls-content"></div>
      <div class="__ls-input-row" id="__ls-input-row">
        <span class="__ls-prompt">&gt;</span>
        <input type="text" id="__ls-eval" placeholder="Evaluate expression..." spellcheck="false" autocomplete="off" />
      </div>
    `
    document.body.appendChild(el)
    attachEvents(el)
    return el
  }

  function attachEvents(el) {
    // Tab switching
    el.querySelectorAll('.__ls-tab[data-tab]').forEach(btn => {
      btn.addEventListener('click', () => {
        activeTab = btn.dataset.tab
        el.querySelectorAll('.__ls-tab[data-tab]').forEach(b => b.classList.remove('__ls-active'))
        btn.classList.add('__ls-active')
        el.querySelector('#__ls-input-row').style.display = activeTab === 'console' ? '' : 'none'
        renderActive()
      })
    })

    // Close
    el.querySelector('#__ls-close').addEventListener('click', hide)

    // Clear
    el.querySelector('#__ls-clear').addEventListener('click', () => {
      if (activeTab === 'console') { logs.length = 0 }
      else if (activeTab === 'errors') { errors.length = 0 }
      else if (activeTab === 'ipc') { ipcCalls.length = 0 }
      renderActive()
    })

    // Eval input
    const input = el.querySelector('#__ls-eval')
    const history = []
    let histIdx = -1
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        const expr = input.value.trim()
        if (!expr) return
        history.unshift(expr)
        histIdx = -1
        addLog('info', ['> ' + expr])
        try {
          const result = eval(expr)
          addLog('log', [result])
        } catch(err) {
          addLog('error', [err.message])
        }
        input.value = ''
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        if (histIdx < history.length - 1) {
          histIdx++
          input.value = history[histIdx]
        }
      } else if (e.key === 'ArrowDown') {
        e.preventDefault()
        if (histIdx > 0) {
          histIdx--
          input.value = history[histIdx]
        } else {
          histIdx = -1
          input.value = ''
        }
      }
    })

    // Resize handle
    const handle = el.querySelector('.__ls-resize')
    let resizing = false
    let startY = 0
    let startH = 0
    handle.addEventListener('mousedown', (e) => {
      resizing = true
      startY = e.clientY
      startH = panelHeight
      e.preventDefault()
    })
    document.addEventListener('mousemove', (e) => {
      if (!resizing) return
      const delta = startY - e.clientY
      panelHeight = Math.max(100, Math.min(window.innerHeight - 40, startH + delta))
      el.style.height = panelHeight + 'px'
    })
    document.addEventListener('mouseup', () => { resizing = false })
  }

  function renderActive() {
    const panel = document.getElementById('__ls-debug')
    if (!panel || !panelVisible) return
    const content = panel.querySelector('#__ls-content')
    if (!content) return

    // Update badges
    const errBadge = panel.querySelector('#__ls-err-badge')
    const ipcBadge = panel.querySelector('#__ls-ipc-badge')
    if (errBadge) errBadge.textContent = errors.length
    if (ipcBadge) ipcBadge.textContent = ipcCalls.length

    if (activeTab === 'console') {
      content.innerHTML = logs.map(e =>
        `<div class="__ls-log __ls-${e.level}"><span class="__ls-ts">${ts(e.time)}</span> <span class="__ls-msg">${esc(e.message)}</span></div>`
      ).join('')
    } else if (activeTab === 'errors') {
      content.innerHTML = errors.length === 0
        ? '<div class="__ls-empty">No errors</div>'
        : errors.map(e =>
          `<div class="__ls-log __ls-error">
            <span class="__ls-ts">${ts(e.time)}</span>
            <span class="__ls-msg">${esc(e.message)}</span>
            ${e.source ? `<div class="__ls-src">${esc(e.source)}${e.line ? ':' + e.line : ''}</div>` : ''}
            ${e.stack ? `<pre class="__ls-stack">${esc(e.stack)}</pre>` : ''}
          </div>`
        ).join('')
    } else if (activeTab === 'ipc') {
      content.innerHTML = ipcCalls.length === 0
        ? '<div class="__ls-empty">No IPC calls yet</div>'
        : `<table class="__ls-ipc-table">
            <thead><tr><th>Time</th><th>Method</th><th>Params</th><th>Result</th><th>ms</th></tr></thead>
            <tbody>${ipcCalls.map(e => {
              const params = e.params ? truncate(JSON.stringify(e.params), 60) : ''
              const result = e.error
                ? `<span class="__ls-ipc-err">${esc(truncate(e.error, 40))}</span>`
                : e.result !== null ? esc(truncate(JSON.stringify(e.result), 60)) : '<span class="__ls-pending">...</span>'
              const dur = e.duration !== null ? e.duration : ''
              return `<tr><td>${ts(e.time)}</td><td class="__ls-method">${esc(e.method)}</td><td>${esc(params)}</td><td>${result}</td><td>${dur}</td></tr>`
            }).join('')}</tbody>
          </table>`
    } else if (activeTab === 'info') {
      content.innerHTML = `<div class="__ls-info-grid">
        <div><b>User Agent</b></div><div>${esc(navigator.userAgent)}</div>
        <div><b>Window Size</b></div><div>${window.innerWidth} x ${window.innerHeight}</div>
        <div><b>Device Pixel Ratio</b></div><div>${window.devicePixelRatio}</div>
        <div><b>URL</b></div><div>${esc(location.href)}</div>
        <div><b>Mode</b></div><div>Development</div>
      </div>`
    }

    content.scrollTop = content.scrollHeight
  }

  function show() {
    let panel = document.getElementById('__ls-debug')
    if (!panel) panel = createPanel()
    panel.style.display = ''
    panel.style.height = panelHeight + 'px'
    panelVisible = true
    panel.querySelector('#__ls-input-row').style.display = activeTab === 'console' ? '' : 'none'
    renderActive()
  }

  function hide() {
    const panel = document.getElementById('__ls-debug')
    if (panel) panel.style.display = 'none'
    panelVisible = false
  }

  function toggle() {
    panelVisible ? hide() : show()
  }

  // --- Styles ---
  const style = document.createElement('style')
  style.textContent = `
    #__ls-debug {
      all: initial;
      position: fixed;
      bottom: 0; left: 0; right: 0;
      height: 300px;
      background: #1e1e1e;
      color: #ccc;
      font-family: 'SF Mono', Menlo, Monaco, 'Courier New', monospace;
      font-size: 12px;
      z-index: 2147483647;
      display: none;
      flex-direction: column;
      border-top: 1px solid #444;
      box-sizing: border-box;
    }
    #__ls-debug * { box-sizing: border-box; }
    #__ls-debug .__ls-resize {
      height: 4px; cursor: ns-resize; background: transparent;
      flex-shrink: 0;
    }
    #__ls-debug .__ls-resize:hover { background: #007acc; }
    #__ls-debug .__ls-tabs {
      display: flex; align-items: center; background: #252526;
      border-bottom: 1px solid #444; padding: 0 4px; flex-shrink: 0;
      height: 28px;
    }
    #__ls-debug .__ls-tab {
      all: unset; padding: 4px 10px; cursor: pointer; color: #888;
      font-family: inherit; font-size: 11px; white-space: nowrap;
    }
    #__ls-debug .__ls-tab:hover { color: #ddd; }
    #__ls-debug .__ls-tab.__ls-active { color: #fff; border-bottom: 2px solid #007acc; }
    #__ls-debug .__ls-spacer { flex: 1; }
    #__ls-debug .__ls-close-btn { font-size: 16px; color: #888; padding: 4px 8px; }
    #__ls-debug .__ls-close-btn:hover { color: #fff; }
    #__ls-debug .__ls-clear-btn { color: #888; }
    #__ls-debug .__ls-clear-btn:hover { color: #fff; }
    #__ls-debug .__ls-badge {
      display: inline-block; background: #444; color: #aaa; border-radius: 8px;
      padding: 0 5px; font-size: 10px; margin-left: 4px; min-width: 16px;
      text-align: center;
    }
    #__ls-debug .__ls-content {
      flex: 1; overflow: auto; padding: 4px 8px;
    }
    #__ls-debug .__ls-log {
      padding: 2px 0; border-bottom: 1px solid #2a2a2a; white-space: pre-wrap;
      word-break: break-all; line-height: 1.4;
    }
    #__ls-debug .__ls-ts { color: #666; margin-right: 8px; }
    #__ls-debug .__ls-msg { color: #ccc; }
    #__ls-debug .__ls-log.__ls-warn { background: #332b00; }
    #__ls-debug .__ls-log.__ls-warn .__ls-msg { color: #e6c300; }
    #__ls-debug .__ls-log.__ls-error { background: #2d0000; }
    #__ls-debug .__ls-log.__ls-error .__ls-msg { color: #f48771; }
    #__ls-debug .__ls-log.__ls-info .__ls-msg { color: #75beff; }
    #__ls-debug .__ls-log.__ls-debug .__ls-msg { color: #888; }
    #__ls-debug .__ls-src { color: #888; font-size: 11px; margin-top: 2px; }
    #__ls-debug .__ls-stack {
      color: #888; font-size: 11px; margin: 4px 0 0 0; padding: 0;
      white-space: pre-wrap; font-family: inherit;
    }
    #__ls-debug .__ls-empty { color: #666; padding: 20px; text-align: center; }
    #__ls-debug .__ls-ipc-table {
      width: 100%; border-collapse: collapse;
    }
    #__ls-debug .__ls-ipc-table th {
      text-align: left; color: #888; font-weight: normal; padding: 4px 8px;
      border-bottom: 1px solid #333; font-size: 11px; position: sticky; top: 0;
      background: #1e1e1e;
    }
    #__ls-debug .__ls-ipc-table td {
      padding: 3px 8px; border-bottom: 1px solid #2a2a2a; font-size: 11px;
      white-space: nowrap; max-width: 200px; overflow: hidden; text-overflow: ellipsis;
    }
    #__ls-debug .__ls-method { color: #dcdcaa; }
    #__ls-debug .__ls-ipc-err { color: #f48771; }
    #__ls-debug .__ls-pending { color: #666; }
    #__ls-debug .__ls-info-grid {
      display: grid; grid-template-columns: 160px 1fr; gap: 4px 12px; padding: 8px;
    }
    #__ls-debug .__ls-info-grid b { color: #888; font-weight: normal; }
    #__ls-debug .__ls-input-row {
      display: flex; align-items: center; border-top: 1px solid #444;
      padding: 0 8px; height: 28px; flex-shrink: 0; background: #1e1e1e;
    }
    #__ls-debug .__ls-prompt { color: #007acc; margin-right: 6px; font-weight: bold; }
    #__ls-debug #__ls-eval {
      all: unset; flex: 1; color: #ccc; font-family: inherit; font-size: 12px;
      caret-color: #fff;
    }
  `
  document.head.appendChild(style)

  // When display changes from none, use flex
  const observer = new MutationObserver(() => {
    const panel = document.getElementById('__ls-debug')
    if (panel && panel.style.display === '') {
      panel.style.display = 'flex'
    }
  })

  // Wait for body
  function init() {
    if (document.body) {
      observer.observe(document.body, { childList: true, subtree: true })
    } else {
      document.addEventListener('DOMContentLoaded', () => {
        observer.observe(document.body, { childList: true, subtree: true })
      })
    }
  }
  init()

  window.__lightshell_debug = { toggle, show, hide }
})()
