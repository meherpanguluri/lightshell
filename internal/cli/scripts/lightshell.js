(() => {
  const pending = new Map()
  const listeners = new Map()

  function call(method, params) {
    if (params === undefined) params = {}
    return new Promise((resolve, reject) => {
      const id = crypto.randomUUID()
      pending.set(id, { resolve, reject })
      const msg = JSON.stringify({ id, method, params })
      if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.lightshell) {
        window.webkit.messageHandlers.lightshell.postMessage(msg)
      } else {
        reject(new Error('LightShell runtime not available'))
        pending.delete(id)
      }
    })
  }

  function on(event, cb) {
    if (!listeners.has(event)) listeners.set(event, [])
    listeners.get(event).push(cb)
    return () => {
      const cbs = listeners.get(event)
      if (cbs) {
        const idx = cbs.indexOf(cb)
        if (idx !== -1) cbs.splice(idx, 1)
      }
    }
  }

  window.__lightshell_receive = function(json) {
    let msg
    if (typeof json === 'string') {
      msg = JSON.parse(json)
    } else {
      msg = json
    }
    if (msg.id && pending.has(msg.id)) {
      const { resolve, reject } = pending.get(msg.id)
      pending.delete(msg.id)
      if (msg.error) {
        reject(new Error(msg.error))
      } else {
        resolve(msg.result)
      }
    } else if (msg.event) {
      const cbs = listeners.get(msg.event)
      if (cbs) cbs.forEach(cb => cb(msg.data))
    }
  }

  window.lightshell = {
    window: {
      setTitle: (t) => call('window.setTitle', { title: t }),
      setSize: (w, h) => call('window.setSize', { width: w, height: h }),
      getSize: () => call('window.getSize'),
      setPosition: (x, y) => call('window.setPosition', { x, y }),
      getPosition: () => call('window.getPosition'),
      minimize: () => call('window.minimize'),
      maximize: () => call('window.maximize'),
      fullscreen: () => call('window.fullscreen'),
      restore: () => call('window.restore'),
      close: () => call('window.close'),
      setContentProtection: (enabled) => call('window.setContentProtection', { enabled }),
      setVibrancy: (style) => call('window.setVibrancy', { style }),
      setColorScheme: (scheme) => call('window.setColorScheme', { scheme }),
      onFileDrop: (cb) => { call('window.enableFileDrop'); return on('window.fileDrop', cb) },
      onResize: (cb) => on('window.resize', cb),
      onMove: (cb) => on('window.move', cb),
      onFocus: (cb) => on('window.focus', cb),
      onBlur: (cb) => on('window.blur', cb),
    },
    fs: {
      readFile: (path, enc) => call('fs.readFile', { path, encoding: enc || 'utf-8' }),
      writeFile: (path, data) => call('fs.writeFile', { path, data }),
      readDir: (path) => call('fs.readDir', { path }),
      exists: (path) => call('fs.exists', { path }),
      stat: (path) => call('fs.stat', { path }),
      mkdir: (path) => call('fs.mkdir', { path }),
      remove: (path) => call('fs.remove', { path }),
      watch: (path, cb) => { call('fs.watch', { path }); return on('fs.watch', cb) },
    },
    dialog: {
      open: (opts) => call('dialog.open', opts || {}),
      save: (opts) => call('dialog.save', opts || {}),
      message: (title, msg) => call('dialog.message', { title, message: msg }),
      confirm: (title, msg) => call('dialog.confirm', { title, message: msg }),
      prompt: (title, def) => call('dialog.prompt', { title, default: def || '' }),
    },
    clipboard: {
      read: () => call('clipboard.read'),
      write: (text) => call('clipboard.write', { text }),
    },
    shell: {
      open: (url) => call('shell.open', { url }),
    },
    notify: {
      send: (title, body, opts) => call('notify.send', Object.assign({ title, body }, opts || {})),
    },
    tray: {
      set: (opts) => call('tray.set', opts),
      remove: () => call('tray.remove'),
      onClick: (cb) => on('tray.click', cb),
    },
    menu: {
      set: (template) => call('menu.set', { template }),
    },
    system: {
      platform: () => call('system.platform'),
      arch: () => call('system.arch'),
      homeDir: () => call('system.homeDir'),
      tempDir: () => call('system.tempDir'),
      hostname: () => call('system.hostname'),
    },
    app: {
      quit: () => call('app.quit'),
      version: () => call('app.version'),
      dataDir: () => call('app.dataDir'),
      setBadgeCount: (count) => call('app.setBadgeCount', { count }),
      enableSingleInstance: () => call('app.enableSingleInstance'),
      onSecondInstance: (cb) => on('app.secondInstance', cb),
      onProtocol: (cb) => on('app.openUrl', cb),
    },
    store: {
      get:    (key)        => call('store.get', { key }),
      set:    (key, value) => call('store.set', { key, value }),
      delete: (key)        => call('store.delete', { key }),
      has:    (key)        => call('store.has', { key }),
      keys:   (prefix)     => call('store.keys', { prefix: prefix || '' }),
      clear:  ()           => call('store.clear'),
    },
    http: {
      fetch: (url, opts) => call('http.fetch', Object.assign({ url }, opts || {})),
      download: (url, opts) => {
        const p = call('http.download', Object.assign({ url }, opts || {}))
        if (opts && opts.onProgress) on('http.download.progress', opts.onProgress)
        return p
      },
    },
    process: {
      exec: (cmd, args, opts) => call('process.exec', Object.assign({ cmd, args: args || [] }, opts || {})),
    },
    shortcuts: {
      register:     (combo, cb) => { call('shortcuts.register', { combo }); return on('shortcut.' + combo, cb) },
      unregister:   (combo)     => call('shortcuts.unregister', { combo }),
      unregisterAll: ()         => call('shortcuts.unregisterAll'),
      isRegistered: (combo)     => call('shortcuts.isRegistered', { combo }),
    },
    updater: {
      check:           ()   => call('updater.check'),
      install:         ()   => call('updater.install'),
      checkAndInstall: ()   => call('updater.checkAndInstall'),
      onProgress:      (cb) => on('updater.progress', cb),
    },
    on,
  }
})()
