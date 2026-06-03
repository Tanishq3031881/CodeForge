import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../lib/api'
import { useAuth, type AuthUser } from '../lib/store'

export function Signup() {
  const navigate = useNavigate()
  const setAuth = useAuth((s) => s.setAuth)
  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const res = await api<{ token: string; user: AuthUser }>('/api/auth/signup', {
        method: 'POST',
        body: JSON.stringify({ email, username, password }),
      })
      setAuth(res.token, res.user)
      navigate('/dashboard')
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 360, margin: '4rem auto', fontFamily: 'monospace' }}>
      <h1>Sign up</h1>
      <form onSubmit={submit} style={{ display: 'grid', gap: 12 }}>
        <input type="email" placeholder="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input type="text" placeholder="username" value={username} onChange={(e) => setUsername(e.target.value)} required />
        <input type="password" placeholder="password (≥8 chars)" value={password} onChange={(e) => setPassword(e.target.value)} required />
        <button type="submit" disabled={loading}>{loading ? '...' : 'Sign up'}</button>
        {error && <div style={{ color: 'red' }}>{error}</div>}
      </form>
      <p>Have an account? <Link to="/login">Log in</Link></p>
    </div>
  )
}
