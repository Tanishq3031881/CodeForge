import { useState } from 'react'
import { createRoom } from '../lib/rooms'
import type { Room } from '../lib/types'

interface Props {
  onCreated: (room: Room) => void
  onClose: () => void
}

export function CreateRoomModal({ onCreated, onClose }: Props) {
  const [name, setName] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const room = await createRoom(name, isPublic)
      onCreated(room)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={overlay} onClick={onClose}>
      <div style={modal} onClick={(e) => e.stopPropagation()}>
        <h2 style={{ marginTop: 0 }}>New room</h2>
        <form onSubmit={submit} style={{ display: 'grid', gap: 12 }}>
          <input
            autoFocus
            placeholder="room name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
          <label style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
            />
            Public (anyone with the link can view)
          </label>
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <button type="button" onClick={onClose}>
              Cancel
            </button>
            <button type="submit" disabled={loading}>
              {loading ? '...' : 'Create'}
            </button>
          </div>
          {error && <div style={{ color: 'red' }}>{error}</div>}
        </form>
      </div>
    </div>
  )
}

const overlay: React.CSSProperties = {
  position: 'fixed',
  inset: 0,
  background: 'rgba(0,0,0,0.5)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
}

const modal: React.CSSProperties = {
  background: '#1e1e1e',
  color: '#eee',
  padding: '1.5rem',
  borderRadius: 8,
  width: 360,
  fontFamily: 'monospace',
}
