const startTime = Date.now()

async function init() {
  // Fetch system info
  const [platform, arch, hostname, version, homeDir, tempDir, dataDir] =
    await Promise.all([
      lightshell.system.platform(),
      lightshell.system.arch(),
      lightshell.system.hostname(),
      lightshell.app.version(),
      lightshell.system.homeDir(),
      lightshell.system.tempDir(),
      lightshell.app.dataDir(),
    ])

  // System cards
  document.getElementById('platform').textContent = platform
  document.getElementById('arch').textContent = arch
  document.getElementById('hostname').textContent = hostname
  document.getElementById('version').textContent = version

  // Directories
  document.getElementById('home-dir').textContent = homeDir
  document.getElementById('temp-dir').textContent = tempDir
  document.getElementById('data-dir').textContent = dataDir

  // List home directory
  try {
    const entries = await lightshell.fs.readDir(homeDir)
    const list = document.getElementById('file-list')
    list.innerHTML = entries
      .sort((a, b) => {
        if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
        return a.name.localeCompare(b.name)
      })
      .slice(0, 30)
      .map(entry => {
        const cls = entry.isDir ? 'dir' : 'file'
        const prefix = entry.isDir ? '/' : ''
        return `<div class="file-entry ${cls}">${prefix}${entry.name}</div>`
      })
      .join('')
  } catch (err) {
    document.getElementById('file-list').textContent = 'Unable to read directory'
  }

  // Actions
  document.getElementById('btn-clipboard').addEventListener('click', async () => {
    const info = `Platform: ${platform}\nArch: ${arch}\nHostname: ${hostname}\nHome: ${homeDir}`
    await lightshell.clipboard.write(info)
    lightshell.notify.send('Copied', 'System info copied to clipboard')
  })

  document.getElementById('btn-notify').addEventListener('click', () => {
    lightshell.notify.send('Hello!', 'This is a test notification from System Monitor')
  })

  document.getElementById('btn-open-home').addEventListener('click', () => {
    lightshell.shell.open(homeDir)
  })
}

// Uptime ticker
function updateUptime() {
  const seconds = Math.floor((Date.now() - startTime) / 1000)
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  document.getElementById('uptime').textContent =
    `uptime ${m}:${s.toString().padStart(2, '0')}`
}

setInterval(updateUptime, 1000)
updateUptime()

init()
