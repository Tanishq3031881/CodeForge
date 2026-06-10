import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useAuth } from '../lib/store'
import {
  createFile,
  deleteFile,
  getFileContent,
  getRoom,
  saveFileContent,
} from '../lib/rooms'
import type { FileMeta, Room as RoomType } from '../lib/types'
import { FileTree } from '../components/FileTree'
import { CodeEditor } from '../components/Editor'
import { EditorToolbar } from '../components/EditorToolbar'

export function Room() {
  const { slug = '' } = useParams()
  const user = useAuth((s) => s.user)

  const [room, setRoom] = useState<RoomType | null>(null)
  const [files, setFiles] = useState<FileMeta[]>([])
  const [selected, setSelected] = useState<FileMeta | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Editor state for the selected file.
  const [content, setContent] = useState('')
  const [dirty, setDirty] = useState(false)
  const [saving, setSaving] = useState(false)
  const [loadingContent, setLoadingContent] = useState(false)

  useEffect(() => {
    setLoading(true)
    getRoom(slug)
      .then((data) => {
        setRoom(data.room)
        setFiles(data.files)
      })
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false))
  }, [slug])

  // Load content whenever the selected file changes. A request token guards
  // against a slow fetch for an old file overwriting a newer selection.
  const reqToken = useRef(0)
  useEffect(() => {
    if (!selected) {
      setContent('')
      setDirty(false)
      return
    }
    const token = ++reqToken.current
    setLoadingContent(true)
    setDirty(false)
    getFileContent(slug, selected.id)
      .then((res) => {
        if (token === reqToken.current) setContent(res.content)
      })
      .catch((err) => {
        if (token === reqToken.current) setError((err as Error).message)
      })
      .finally(() => {
        if (token === reqToken.current) setLoadingContent(false)
      })
  }, [slug, selected])

  const canEdit = !!room && !!user && room.owner_id === user.id

  async function handleCreate(path: string, language: string) {
    const f = await createFile(slug, path, language)
    setFiles((fs) => [...fs, f].sort((a, b) => a.path.localeCompare(b.path)))
    setSelected(f)
  }

  async function handleDelete(fileId: string) {
    await deleteFile(slug, fileId)
    setFiles((fs) => fs.filter((f) => f.id !== fileId))
    setSelected((s) => (s?.id === fileId ? null : s))
  }

  async function handleSave() {
    if (!selected || !dirty) return
    setSaving(true)
    try {
      await saveFileContent(slug, selected.id, content)
      setDirty(false)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <Centered>Loading…</Centered>
  if (error && !room) return <Centered>{error} · <Link to="/dashboard">back</Link></Centered>
  if (!room) return null

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', fontFamily: 'monospace', color: '#eee' }}>
      <header style={topbar}>
        <Link to="/dashboard" style={{ color: '#6cb6ff', textDecoration: 'none' }}>
          ← rooms
        </Link>
        <strong>{room.name}</strong>
        <span style={{ color: '#888', fontSize: 12 }}>/{room.slug}</span>
        {error && <span style={{ color: 'salmon', fontSize: 12, marginLeft: 'auto' }}>{error}</span>}
      </header>

      <div style={{ display: 'flex', flex: 1, minHeight: 0 }}>
        <aside style={{ width: 240, borderRight: '1px solid #333', background: '#1e1e1e' }}>
          <FileTree
            files={files}
            selectedId={selected?.id ?? null}
            canEdit={canEdit}
            onSelect={setSelected}
            onCreate={handleCreate}
            onDelete={handleDelete}
          />
        </aside>

        <main style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, background: '#1e1e1e' }}>
          {!selected ? (
            <Placeholder>Select a file</Placeholder>
          ) : loadingContent ? (
            <Placeholder>Loading {selected.path}…</Placeholder>
          ) : (
            <>
              <EditorToolbar
                path={selected.path}
                language={selected.language}
                dirty={dirty}
                saving={saving}
                canEdit={canEdit}
                onSave={handleSave}
              />
              <div style={{ flex: 1, minHeight: 0 }}>
                <CodeEditor
                  language={selected.language}
                  value={content}
                  readOnly={!canEdit}
                  onChange={(v) => {
                    setContent(v)
                    setDirty(true)
                  }}
                  onSave={handleSave}
                />
              </div>
            </>
          )}
        </main>
      </div>
    </div>
  )
}

function Placeholder({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#666' }}>
      {children}
    </div>
  )
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', fontFamily: 'monospace', color: '#ccc' }}>
      {children}
    </div>
  )
}

const topbar: React.CSSProperties = {
  display: 'flex',
  gap: 16,
  alignItems: 'center',
  padding: '10px 16px',
  borderBottom: '1px solid #333',
  background: '#252526',
}
