import { useState } from 'react'
import { LANGUAGES, type FileMeta } from '../lib/types'

interface Props {
  files: FileMeta[]
  selectedId: string | null
  canEdit: boolean
  onSelect: (file: FileMeta) => void
  onCreate: (path: string, language: string) => Promise<void>
  onDelete: (fileId: string) => Promise<void>
}

export function FileTree({ files, selectedId, canEdit, onSelect, onCreate, onDelete }: Props) {
  const [adding, setAdding] = useState(false)
  const [path, setPath] = useState('')
  const [language, setLanguage] = useState<string>('python')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setBusy(true)
    try {
      await onCreate(path.trim(), language)
      setPath('')
      setAdding(false)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={header}>
        <span>FILES</span>
        {canEdit && (
          <button title="New file" onClick={() => setAdding((v) => !v)} style={iconBtn}>
            +
          </button>
        )}
      </div>

      {adding && (
        <form onSubmit={submit} style={{ padding: '8px', display: 'grid', gap: 6 }}>
          <input
            autoFocus
            placeholder="path e.g. main.py"
            value={path}
            onChange={(e) => setPath(e.target.value)}
            required
          />
          <select value={language} onChange={(e) => setLanguage(e.target.value)}>
            {LANGUAGES.map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
          <button type="submit" disabled={busy}>
            {busy ? '...' : 'Add file'}
          </button>
          {error && <div style={{ color: 'salmon', fontSize: 12 }}>{error}</div>}
        </form>
      )}

      <ul style={{ listStyle: 'none', margin: 0, padding: 0, overflowY: 'auto', flex: 1 }}>
        {files.length === 0 && <li style={empty}>No files yet</li>}
        {files.map((f) => (
          <li
            key={f.id}
            onClick={() => onSelect(f)}
            style={{
              ...row,
              background: f.id === selectedId ? '#094771' : 'transparent',
            }}
          >
            <span style={{ overflow: 'hidden', textOverflow: 'ellipsis' }}>{f.path}</span>
            {canEdit && (
              <button
                title="Delete file"
                onClick={(e) => {
                  e.stopPropagation()
                  void onDelete(f.id)
                }}
                style={iconBtn}
              >
                ×
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}

const header: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '8px 12px',
  fontSize: 12,
  letterSpacing: 1,
  color: '#999',
  borderBottom: '1px solid #333',
}

const row: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '6px 12px',
  cursor: 'pointer',
  fontSize: 13,
}

const empty: React.CSSProperties = { padding: '12px', color: '#666', fontSize: 13 }

const iconBtn: React.CSSProperties = {
  background: 'transparent',
  border: 'none',
  color: '#aaa',
  cursor: 'pointer',
  fontSize: 16,
  lineHeight: 1,
  padding: '0 4px',
}
