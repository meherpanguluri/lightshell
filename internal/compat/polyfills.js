(() => {
  const isLinux = navigator.platform.includes('Linux')

  // Platform class for targeted CSS
  document.documentElement.classList.add(
    isLinux ? 'platform-linux' : 'platform-darwin'
  )

  // structuredClone polyfill (missing in older WebKitGTK)
  if (typeof structuredClone === 'undefined') {
    window.structuredClone = (obj) => JSON.parse(JSON.stringify(obj))
  }

  // Intl.Segmenter warning (missing in WebKitGTK)
  if (typeof Intl !== 'undefined' && !Intl.Segmenter) {
    console.warn('[LightShell] Intl.Segmenter not available on this platform')
  }

  // backdrop-filter fallback for WebKitGTK
  if (isLinux && typeof CSS !== 'undefined' && CSS.supports && !CSS.supports('backdrop-filter', 'blur(1px)')) {
    const style = document.createElement('style')
    style.textContent = '[style*="backdrop-filter"]{background-color:rgba(255,255,255,0.9)!important}'
    document.head.appendChild(style)
  }
})()
