// window-manager.js — Core: drag, resize, stack, minimize, close
const WindowManager = (() => {
  const windows = new Map()
  let topZ = 100
  let activeId = null

  function create(opts) {
    const id = opts.id
    if (windows.has(id)) {
      focus(id)
      return
    }

    const el = document.createElement('div')
    el.className = 'window'
    el.id = 'win-' + id
    el.style.width = (opts.width || 600) + 'px'
    el.style.height = (opts.height || 400) + 'px'
    el.style.zIndex = ++topZ

    // Position: use provided or center
    const x = opts.x != null ? opts.x : (window.innerWidth - (opts.width || 600)) / 2
    const y = opts.y != null ? opts.y : Math.max(40, (window.innerHeight - (opts.height || 400)) / 2 - 40)
    el.style.left = x + 'px'
    el.style.top = y + 'px'

    // Header
    const header = document.createElement('div')
    header.className = 'window-header'

    const controls = document.createElement('div')
    controls.className = 'window-controls'

    const btnClose = document.createElement('button')
    btnClose.className = 'window-control close'
    btnClose.setAttribute('aria-label', 'Close')
    btnClose.onclick = (e) => { e.stopPropagation(); close(id) }

    const btnMin = document.createElement('button')
    btnMin.className = 'window-control minimize'
    btnMin.setAttribute('aria-label', 'Minimize')
    btnMin.onclick = (e) => { e.stopPropagation(); minimize(id) }

    const btnMax = document.createElement('button')
    btnMax.className = 'window-control maximize'
    btnMax.setAttribute('aria-label', 'Maximize')
    btnMax.onclick = (e) => { e.stopPropagation(); toggleMaximize(id) }

    controls.append(btnClose, btnMin, btnMax)

    const title = document.createElement('div')
    title.className = 'window-title'
    title.textContent = opts.title || id

    const spacer = document.createElement('div')
    spacer.className = 'window-header-spacer'

    header.append(controls, title, spacer)

    // Body
    const body = document.createElement('div')
    body.className = 'window-body'
    if (typeof opts.content === 'string') {
      body.innerHTML = opts.content
    } else if (opts.content instanceof HTMLElement) {
      body.appendChild(opts.content)
    }

    el.append(header, body)

    // Resize handles
    if (opts.resizable !== false) {
      const dirs = ['n','s','e','w','ne','nw','se','sw']
      dirs.forEach(d => {
        const handle = document.createElement('div')
        handle.className = 'resize-handle ' + d
        handle.dataset.dir = d
        el.appendChild(handle)
      })
    }

    // Store state
    const state = {
      id,
      el,
      minimized: false,
      maximized: false,
      prevBounds: null,
      title: opts.title || id,
      icon: opts.icon || ''
    }
    windows.set(id, state)

    // Focus on mousedown
    el.addEventListener('mousedown', () => focus(id))

    // Drag by header
    setupDrag(header, el, state)

    // Resize by handles
    if (opts.resizable !== false) {
      setupResize(el, state)
    }

    // Double-click title to maximize
    header.addEventListener('dblclick', (e) => {
      if (e.target.closest('.window-controls')) return
      toggleMaximize(id)
    })

    document.querySelector('.desktop').appendChild(el)
    focus(id)
    updateTaskbar()
  }

  function setupDrag(header, el, state) {
    let startX, startY, origX, origY, dragging = false

    header.addEventListener('mousedown', (e) => {
      if (e.target.closest('.window-controls')) return
      if (state.maximized) return
      dragging = true
      startX = e.clientX
      startY = e.clientY
      origX = el.offsetLeft
      origY = el.offsetTop
      el.style.transition = 'none'
      document.body.style.cursor = 'default'
      e.preventDefault()
    })

    document.addEventListener('mousemove', (e) => {
      if (!dragging) return
      const dx = e.clientX - startX
      const dy = e.clientY - startY
      el.style.left = (origX + dx) + 'px'
      el.style.top = Math.max(0, origY + dy) + 'px'
    })

    document.addEventListener('mouseup', () => {
      if (!dragging) return
      dragging = false
      el.style.transition = ''
      document.body.style.cursor = ''

      // Snap back if dragged mostly off screen
      const rect = el.getBoundingClientRect()
      const vw = window.innerWidth
      const vh = window.innerHeight
      let snapped = false
      let newLeft = parseFloat(el.style.left)
      let newTop = parseFloat(el.style.top)

      if (rect.right < 80) { newLeft = 20; snapped = true }
      if (rect.left > vw - 80) { newLeft = vw - 100; snapped = true }
      if (rect.bottom < 60) { newTop = 20; snapped = true }
      if (rect.top > vh - 80) { newTop = vh - 120; snapped = true }

      if (snapped) {
        el.style.transition = 'left 0.3s cubic-bezier(0.34, 1.56, 0.64, 1), top 0.3s cubic-bezier(0.34, 1.56, 0.64, 1)'
        el.style.left = newLeft + 'px'
        el.style.top = newTop + 'px'
        setTimeout(() => { el.style.transition = '' }, 350)
      }
    })
  }

  function setupResize(el, state) {
    let startX, startY, startW, startH, startL, startT, dir, resizing = false

    el.addEventListener('mousedown', (e) => {
      const handle = e.target.closest('.resize-handle')
      if (!handle || state.maximized) return
      resizing = true
      dir = handle.dataset.dir
      startX = e.clientX
      startY = e.clientY
      startW = el.offsetWidth
      startH = el.offsetHeight
      startL = el.offsetLeft
      startT = el.offsetTop
      el.style.transition = 'none'
      e.preventDefault()
      e.stopPropagation()
    })

    document.addEventListener('mousemove', (e) => {
      if (!resizing) return
      const dx = e.clientX - startX
      const dy = e.clientY - startY
      const minW = 280, minH = 180

      if (dir.includes('e')) el.style.width = Math.max(minW, startW + dx) + 'px'
      if (dir.includes('w')) {
        const newW = Math.max(minW, startW - dx)
        el.style.width = newW + 'px'
        el.style.left = (startL + startW - newW) + 'px'
      }
      if (dir.includes('s')) el.style.height = Math.max(minH, startH + dy) + 'px'
      if (dir.includes('n')) {
        const newH = Math.max(minH, startH - dy)
        el.style.height = newH + 'px'
        el.style.top = (startT + startH - newH) + 'px'
      }
    })

    document.addEventListener('mouseup', () => {
      if (!resizing) return
      resizing = false
      el.style.transition = ''
    })
  }

  function focus(id) {
    const state = windows.get(id)
    if (!state) return
    windows.forEach((s) => s.el.classList.remove('focused'))
    state.el.classList.add('focused')
    state.el.style.zIndex = ++topZ
    activeId = id
    updateTaskbar()
  }

  function close(id) {
    const state = windows.get(id)
    if (!state) return
    state.el.style.transition = 'opacity 0.2s ease, transform 0.2s ease'
    state.el.style.opacity = '0'
    state.el.style.transform = 'scale(0.95)'
    setTimeout(() => {
      state.el.remove()
      windows.delete(id)
      if (activeId === id) activeId = null
      updateTaskbar()
    }, 200)
  }

  function minimize(id) {
    const state = windows.get(id)
    if (!state) return
    state.minimized = true
    state.el.classList.add('minimizing')
    setTimeout(() => {
      state.el.style.display = 'none'
      state.el.classList.remove('minimizing')
    }, 300)
    if (activeId === id) activeId = null
    updateTaskbar()
  }

  function restore(id) {
    const state = windows.get(id)
    if (!state) return
    if (state.minimized) {
      state.minimized = false
      state.el.style.display = ''
      state.el.style.opacity = '1'
      state.el.style.transform = ''
      focus(id)
    }
    updateTaskbar()
  }

  function toggleMaximize(id) {
    const state = windows.get(id)
    if (!state) return

    if (state.maximized) {
      // Restore
      const b = state.prevBounds
      state.el.style.left = b.left + 'px'
      state.el.style.top = b.top + 'px'
      state.el.style.width = b.width + 'px'
      state.el.style.height = b.height + 'px'
      state.el.classList.remove('maximized')
      state.maximized = false
    } else {
      // Maximize
      state.prevBounds = {
        left: state.el.offsetLeft,
        top: state.el.offsetTop,
        width: state.el.offsetWidth,
        height: state.el.offsetHeight
      }
      state.el.style.left = '0'
      state.el.style.top = '0'
      state.el.style.width = '100vw'
      state.el.style.height = 'calc(100vh - 48px)'
      state.el.classList.add('maximized')
      state.maximized = true
    }
    focus(id)
  }

  function updateTaskbar() {
    const container = document.querySelector('.taskbar-windows')
    if (!container) return
    container.innerHTML = ''
    windows.forEach((state, id) => {
      const item = document.createElement('div')
      item.className = 'taskbar-item' + (id === activeId && !state.minimized ? ' active' : '')
      item.innerHTML = `<span class="item-icon">${state.icon}</span> ${state.title}`
      item.onclick = () => {
        if (state.minimized) {
          restore(id)
        } else if (id === activeId) {
          minimize(id)
        } else {
          focus(id)
        }
      }
      container.appendChild(item)
    })
  }

  function isOpen(id) {
    return windows.has(id)
  }

  return { open: create, close, minimize, restore, focus, toggleMaximize, isOpen, updateTaskbar }
})()
