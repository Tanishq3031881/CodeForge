interface Props {
  path: string
  language: string
  dirty: boolean
  saving: boolean
  canEdit: boolean
  onSave: () => void
}

export function EditorToolbar({ path, language, dirty, saving, canEdit, onSave }: Props) {
  return (
    <div style={bar}>
      <span style={{ color: '#ccc' }}>
        {path}
        {dirty && <span title="Unsaved changes" style={{ color: '#e2c08d' }}> ●</span>}
      </span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ color: '#888', fontSize: 12 }}>{language}</span>
        {canEdit && (
          <button onClick={onSave} disabled={saving || !dirty}>
            {saving ? 'Saving…' : 'Save (Ctrl+S)'}
          </button>
        )}
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
