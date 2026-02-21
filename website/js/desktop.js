// desktop.js — Desktop icons, taskbar, window content loading
(() => {
  // Clock
  function updateClock() {
    const el = document.querySelector('.taskbar-clock')
    if (!el) return
    const now = new Date()
    const h = now.getHours()
    const m = String(now.getMinutes()).padStart(2, '0')
    const ampm = h >= 12 ? 'PM' : 'AM'
    const h12 = h % 12 || 12
    el.textContent = `${h12}:${m} ${ampm}`
  }
  updateClock()
  setInterval(updateClock, 10000)

  // Desktop icon definitions
  const icons = [
    { id: 'getting-started', label: 'getting-started.md', icon: '\uD83D\uDCD6', action: openGettingStarted },
    { id: 'examples',        label: 'examples/',          icon: '\uD83D\uDDC2\uFE0F', action: openExamples },
    { id: 'playground',      label: 'playground.app',     icon: '\u26A1',       action: openPlayground },
    { id: 'docs',            label: 'docs/',              icon: '\uD83D\uDCDA', action: () => window.open('https://lightshell.sh/docs/', '_blank') },
    { id: 'showcase',        label: 'showcase/',          icon: '\uD83C\uDFC6', action: openShowcase },
    { id: 'prompts',         label: 'prompts.txt',        icon: '\uD83E\uDD16', action: openPrompts },
    { id: 'benchmarks',      label: 'benchmarks.csv',     icon: '\uD83D\uDCC8', action: openBenchmarks },
    { id: 'about',           label: 'about.txt',          icon: '\uD83D\uDCA1', action: openAbout },
    { id: 'github',          label: 'github.link',        icon: '<img src="assets/icons/github.svg" width="36" height="36" alt="GitHub" style="filter: brightness(0.2)">', action: () => window.open('https://github.com/meherpanguluri/lightshell', '_blank') },
  ]

  // Render desktop icons
  const iconGrid = document.querySelector('.desktop-icons')
  if (iconGrid) {
    icons.forEach(def => {
      const el = document.createElement('div')
      el.className = 'desktop-icon'
      el.setAttribute('role', 'button')
      el.setAttribute('tabindex', '0')
      el.setAttribute('aria-label', def.label)
      el.innerHTML = `<div class="icon-img">${def.icon}</div><div class="icon-label">${def.label}</div>`
      el.addEventListener('dblclick', def.action)
      el.addEventListener('keydown', (e) => { if (e.key === 'Enter') def.action() })
      // Single click on mobile
      let clickTimer = null
      el.addEventListener('click', () => {
        if (window.innerWidth <= 768) { def.action(); return }
        if (clickTimer) { clearTimeout(clickTimer); clickTimer = null; return }
        clickTimer = setTimeout(() => { clickTimer = null }, 400)
      })
      iconGrid.appendChild(el)
    })
  }

  // Open welcome window on load
  setTimeout(() => {
    WindowManager.open({
      id: 'welcome',
      title: 'Welcome to LightShell',
      icon: '\u26A1',
      width: 520,
      height: 340,
      content: document.getElementById('content-welcome')?.innerHTML || '',
      resizable: true,
    })
  }, 300)

  // Window content openers
  function openGettingStarted() {
    WindowManager.open({
      id: 'getting-started',
      title: 'getting-started.md',
      icon: '\uD83D\uDCC4',
      width: 660,
      height: 520,
      x: 140,
      y: 60,
      content: document.getElementById('content-getting-started')?.innerHTML || '',
    })
  }

  function openExamples() {
    WindowManager.open({
      id: 'examples',
      title: 'examples/',
      icon: '\uD83D\uDCC1',
      width: 560,
      height: 440,
      x: 200,
      y: 100,
      content: document.getElementById('content-examples')?.innerHTML || '',
    })
  }

  function openPlayground() {
    window.open('playground.html', '_blank')
  }

  function openPrompts() {
    WindowManager.open({
      id: 'prompts',
      title: 'prompts.txt',
      icon: '\uD83E\uDD16',
      width: 600,
      height: 540,
      x: 240,
      y: 50,
      content: document.getElementById('content-prompts')?.innerHTML || '',
    })
  }

  function openShowcase() {
    WindowManager.open({
      id: 'showcase',
      title: 'showcase/',
      icon: '\uD83C\uDFC6',
      width: 620,
      height: 520,
      x: 160,
      y: 70,
      content: document.getElementById('content-showcase')?.innerHTML || '',
    })
  }

  function openBenchmarks() {
    WindowManager.open({
      id: 'benchmarks',
      title: 'benchmarks.csv',
      icon: '\uD83D\uDCCA',
      width: 640,
      height: 380,
      x: 180,
      y: 120,
      content: document.getElementById('content-benchmarks')?.innerHTML || '',
    })
  }

  function openAbout() {
    WindowManager.open({
      id: 'about',
      title: 'about.txt',
      icon: '\uD83D\uDCC4',
      width: 560,
      height: 480,
      x: 220,
      y: 80,
      content: document.getElementById('content-about')?.innerHTML || '',
    })
  }

  // Copy prompt text to clipboard
  window.copyPrompt = function(btn) {
    const text = btn.parentElement.querySelector('.prompt-text p').textContent
    navigator.clipboard.writeText(text).then(() => {
      const orig = btn.textContent
      btn.textContent = 'Copied!'
      setTimeout(() => { btn.textContent = orig }, 1500)
    })
  }

  // Expose for playground link in welcome window
  window.openPlayground = openPlayground
  window.openGettingStarted = openGettingStarted
})()
