import { useEffect, useState } from 'react'
import { api } from './lib/api'

function App() {
  const [status, setStatus] = useState<string>('loading...')

  useEffect(() => {
    api<{ status: string }>('/health')
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('error'))
  }, [])

  return (
    <div style={{ padding: '2rem', fontFamily: 'monospace' }}>
      <h1>CodeForge</h1>
      <p>Backend says: {status}</p>
    </div>
  )
}

export default App
