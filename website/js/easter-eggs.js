// easter-eggs.js — Screensaver, right-click context menu, fun surprises
(() => {
  // Screensaver — Matrix-style falling LightShell API calls
  let screensaverTimer = null
  let screensaverActive = false
  const IDLE_TIMEOUT = 120000 // 2 minutes

  const apiCalls = [
    'lightshell.fs.readFile("notes.md")',
    'lightshell.dialog.open()',
    'lightshell.clipboard.write(text)',
    'lightshell.notify.send("Done!", body)',
    'lightshell.system.platform()',
    'lightshell.window.setTitle(t)',
    'lightshell.shell.open(url)',
    'lightshell.fs.writeFile(path, data)',
    'lightshell.dialog.save()',
    'lightshell.window.setSize(w, h)',
    'lightshell.tray.set({ tooltip })',
    'lightshell.menu.set(template)',
    'lightshell.system.arch()',
    'lightshell.app.dataDir()',
    'lightshell.fs.readDir(".")',
    'lightshell.clipboard.read()',
    'await lightshell.dialog.confirm()',
    'lightshell.window.fullscreen()',
    'lightshell.fs.exists(path)',
    'lightshell.app.version()',
  ]

  function resetIdleTimer() {
    if (screensaverActive) hideScreensaver()
    clearTimeout(screensaverTimer)
    screensaverTimer = setTimeout(showScreensaver, IDLE_TIMEOUT)
  }

  function showScreensaver() {
    if (screensaverActive) return
    screensaverActive = true

    const overlay = document.createElement('div')
    overlay.id = 'screensaver'
    overlay.style.cssText = `
      position: fixed; inset: 0; background: #0a0a12; z-index: 20000;
      cursor: pointer; overflow: hidden;
    `
    overlay.onclick = hideScreensaver

    document.body.appendChild(overlay)

    // Create falling columns
    const cols = Math.floor(window.innerWidth / 140)
    for (let i = 0; i < cols; i++) {
      createFallingColumn(overlay, i * 140 + Math.random() * 60)
    }
  }

  function createFallingColumn(container, x) {
    const col = document.createElement('div')
    col.style.cssText = `
      position: absolute; left: ${x}px; top: -50px;
      font-family: var(--font-mono); font-size: 12px;
      color: rgba(100, 100, 100, 0.6); white-space: nowrap;
      animation: fall ${6 + Math.random() * 8}s linear infinite;
      animation-delay: ${-Math.random() * 8}s;
    `

    // Stack random API calls
    const count = 4 + Math.floor(Math.random() * 6)
    for (let i = 0; i < count; i++) {
      const line = document.createElement('div')
      line.textContent = apiCalls[Math.floor(Math.random() * apiCalls.length)]
      line.style.opacity = (0.3 + Math.random() * 0.7).toFixed(2)
      line.style.marginBottom = '4px'
      col.appendChild(line)
    }

    container.appendChild(col)
  }

  function hideScreensaver() {
    screensaverActive = false
    const el = document.getElementById('screensaver')
    if (el) {
      el.style.transition = 'opacity 0.5s'
      el.style.opacity = '0'
      setTimeout(() => el.remove(), 500)
    }
    resetIdleTimer()
  }

  // Add falling animation keyframe
  const style = document.createElement('style')
  style.textContent = `
    @keyframes fall {
      from { transform: translateY(-100%); }
      to { transform: translateY(100vh); }
    }
  `
  document.head.appendChild(style)

  // Right-click desktop context menu (doesn't override browser's — only on desktop bg)
  document.addEventListener('contextmenu', (e) => {
    const target = e.target
    // Only on the desktop background itself, not on windows or icons
    if (!target.classList.contains('desktop')) return
    e.preventDefault()

    // Remove any existing menu
    const existing = document.getElementById('ctx-menu')
    if (existing) existing.remove()

    const menu = document.createElement('div')
    menu.id = 'ctx-menu'
    menu.style.cssText = `
      position: fixed; left: ${e.clientX}px; top: ${e.clientY}px;
      background: #ffffff;
      border: 1px solid rgba(255,255,255,0.8); border-radius: 16px;
      padding: 6px 0; min-width: 180px; z-index: 10002;
      box-shadow: 0 15px 40px rgba(0,0,0,0.08);
      font-size: 0.8125rem; color: #1a1a1a;
    `

    const items = [
      { label: 'New LightShell App', action: () => window.open('playground.html', '_blank') },
      { label: 'View Source', action: () => window.open('https://github.com/meherpanguluri/lightshell', '_blank') },
      { separator: true },
      { label: 'About LightShell', action: () => {
        if (window.WindowManager) {
          WindowManager.open({
            id: 'about', title: 'about.txt', icon: '\uD83D\uDCC4',
            width: 560, height: 480, x: 220, y: 80,
            content: document.getElementById('content-about')?.innerHTML || '',
          })
        }
      }},
    ]

    items.forEach(item => {
      if (item.separator) {
        const sep = document.createElement('div')
        sep.style.cssText = 'height: 1px; background: rgba(0,0,0,0.08); margin: 4px 8px;'
        menu.appendChild(sep)
        return
      }
      const el = document.createElement('div')
      el.textContent = item.label
      el.style.cssText = 'padding: 6px 14px; cursor: pointer; transition: background 0.1s;'
      el.onmouseenter = () => el.style.background = 'rgba(0, 0, 0, 0.04)'
      el.onmouseleave = () => el.style.background = 'transparent'
      el.onclick = () => { menu.remove(); item.action() }
      menu.appendChild(el)
    })

    document.body.appendChild(menu)

    // Close on click elsewhere
    const closeMenu = (ev) => {
      if (!menu.contains(ev.target)) {
        menu.remove()
        document.removeEventListener('click', closeMenu)
      }
    }
    setTimeout(() => document.addEventListener('click', closeMenu), 0)
  })

  // Konami code
  const konami = [38, 38, 40, 40, 37, 39, 37, 39, 66, 65]
  let konamiIndex = 0

  document.addEventListener('keydown', (e) => {
    if (e.keyCode === konami[konamiIndex]) {
      konamiIndex++
      if (konamiIndex === konami.length) {
        konamiIndex = 0
        spawnHedgehog()
      }
    } else {
      konamiIndex = 0
    }
  })

  function spawnHedgehog() {
    const hog = document.createElement('div')
    hog.style.cssText = `
      position: fixed; bottom: 48px; left: -60px; font-size: 48px;
      z-index: 15000; transition: none; pointer-events: none;
    `
    hog.textContent = '\uD83E\uDD94'
    document.body.appendChild(hog)

    let x = -60
    const walk = () => {
      x += 2
      hog.style.left = x + 'px'
      if (x < window.innerWidth + 60) {
        requestAnimationFrame(walk)
      } else {
        hog.remove()
      }
    }
    requestAnimationFrame(walk)
  }

  // Start idle timer
  ;['mousemove', 'mousedown', 'keydown', 'touchstart', 'scroll'].forEach(evt => {
    document.addEventListener(evt, resetIdleTimer, { passive: true })
  })
  resetIdleTimer()
})()
