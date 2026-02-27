import { useState, useEffect } from 'react'

function App() {
  const [platform, setPlatform] = useState('')
  const [arch, setArch] = useState('')

  useEffect(() => {
    async function init() {
      const p = await lightshell.system.platform()
      const a = await lightshell.system.arch()
      setPlatform(p)
      setArch(a)
    }
    init()
  }, [])

  return (
    <main>
      <h1>{{TITLE}}</h1>
      <p>Your LightShell + React app is running.</p>
      {platform && <div className="info">{platform}/{arch}</div>}
    </main>
  )
}

export default App
