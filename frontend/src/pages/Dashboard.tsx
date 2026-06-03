import { useNavigate } from 'react-router-dom'
import { useAuth } from '../lib/store'

export function Dashboard() {
  const navigate = useNavigate()
  const { user, clear } = useAuth()

  function logout() {
    clear()
    navigate('/login')
  }

  return (
    <div style={{ padding: '2rem', fontFamily: 'monospace' }}>
      <h1>CodeForge</h1>
      <p>Logged in as <strong>{user?.username}</strong> ({user?.email})</p>
      <button onClick={logout}>Log out</button>
    </div>
  )
}
