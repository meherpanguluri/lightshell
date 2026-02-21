async function init() {
  const platform = await lightshell.system.platform()
  const arch = await lightshell.system.arch()
  const info = document.getElementById('info')
  info.textContent = `Running on ${platform}/${arch}`
}

init()
