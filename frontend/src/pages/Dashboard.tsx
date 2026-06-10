import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../lib/store'
import { deleteRoom, listRooms } from '../lib/rooms'
import type { Room } from '../lib/types'
import { CreateRoomModal } from '../components/CreateRoomModal'

export function Dashboard() {
  const navigate = useNavigate()
  const { user, clear } = useAuth()
  const [rooms, setRooms] = useState<Room[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  useEffect(() => {
    listRooms()
      .then(setRooms)
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false))
  }, [])

  function logout() {
    clear()
    navigate('/login')
  }

  async function remove(e: React.MouseEvent, slug: string) {
    e.stopPropagation()
    if (!confirm(`Delete room "${slug}"? This cannot be undone.`)) return
    try {
      await deleteRoom(slug)
      setRooms((rs) => rs.filter((r) => r.slug !== slug))
    } catch (err) {
      setError((err as Error).message)
    }
  }

  return (
    <div style={{ padding: '2rem', fontFamily: 'monospace', color: '#eee' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <h1 style={{ margin: 0 }}>CodeForge</h1>
        <div style={{ fontSize: 13, color: '#aaa' }}>
          {user?.username} · <button onClick={logout} style={linkBtn}>log out</button>
        </div>
      </header>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '1.5rem 0' }}>
        <h2 style={{ margin: 0 }}>Your rooms</h2>
        <button onClick={() => setShowCreate(true)}>+ Create room</button>
      </div>

      {error && <div style={{ color: 'salmon' }}>{error}</div>}
      {loading && <p>Loading…</p>}
      {!loading && rooms.length === 0 && <p style={{ color: '#888' }}>No rooms yet — create one to get started.</p>}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 16 }}>
        {rooms.map((r) => (
          <div key={r.id} onClick={() => navigate(`/r/${r.slug}`)} style={card}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <strong>{r.name}</strong>
              <button title="Delete room" onClick={(e) => remove(e, r.slug)} style={linkBtn}>
                ×
              </button>
            </div>
            <div style={{ fontSize: 12, color: '#888', marginTop: 8 }}>/{r.slug}</div>
            <div style={{ fontSize: 12, color: '#888' }}>{r.is_public ? 'public' : 'private'}</div>
          </div>
        ))}
      </div>

      {showCreate && (
        <CreateRoomModal
          onClose={() => setShowCreate(false)}
          onCreated={(room) => {
            setShowCreate(false)
            navigate(`/r/${room.slug}`)
          }}
        />
      )}
    </div>
  )
}

const card: React.CSSProperties = {
  background: '#252526',
  border: '1px solid #333',
  borderRadius: 8,
  padding: '1rem',
  cursor: 'pointer',
}

const linkBtn: React.CSSProperties = {
  background: 'transparent',
  border: 'none',
  color: '#6cb6ff',
  cursor: 'pointer',
  font: 'inherit',
  padding: 0,
}
