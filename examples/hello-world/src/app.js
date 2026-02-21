async function init() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  const hostname = await lightshell.system.hostname()
  const version = await lightshell.app.version()

  const info = document.getElementById('info')
  info.innerHTML = `
    <div class="info-card">
      <span class="label">Platform</span>
      <span class="value">${platform}</span>
    </div>
    <div class="info-card">
      <span class="label">Architecture</span>
      <span class="value">${arch}</span>
    </div>
    <div class="info-card">
      <span class="label">Hostname</span>
      <span class="value">${hostname}</span>
    </div>
    <div class="info-card">
      <span class="label">App Version</span>
      <span class="value">${version}</span>
    </div>
  `
}

init()
