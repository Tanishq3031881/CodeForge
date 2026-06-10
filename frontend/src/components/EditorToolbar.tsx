import type { SyncStatus } from '../hooks/useYjsRoom'

interface Props {
  path: string
  language: string
  status: SyncStatus
  canEdit: boolean
}

const statusLabel: Record<SyncStatus, { text: string; color: string }> = {
  connecting: { text: '● connecting', color: '#e2c08d' },
  connected: { text: '● live', color: '#6cc070' },
  disconnected: { text: '● offline', color: 'salmon' },
}

export function EditorToolbar({ path, language, status, canEdit }: Props) {
  const s = statusLabel[status]
  return (
    <div style={bar}>
      <span style={{ color: '#ccc' }}>
        {path}
        {!canEdit && <span style={{ color: '#888' }}> (read-only)</span>}
      </span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ color: '#888', fontSize: 12 }}>{language}</span>
        <span title="Realtime sync status" style={{ color: s.color, fontSize: 12 }}>
          {s.text}
        </span>
      </div>
    </div>
  )
}

const bar: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '6px 12px',
  borderBottom: '1px solid #333',
  background: '#252526',
  fontFamily: 'monospace',
  fontSize: 13,
}
