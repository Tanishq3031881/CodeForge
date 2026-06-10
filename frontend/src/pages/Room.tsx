import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useAuth } from '../lib/store'
import { createFile, deleteFile, getFileContent, getRoom } from '../lib/rooms'
import type { FileMeta, Room as RoomType } from '../lib/types'
import { FileTree } from '../components/FileTree'
import { CollaborativeEditor } from '../components/CollaborativeEditor'
import { EditorToolbar } from '../components/EditorToolbar'
import { OutputPanel, type OutputChunk, type ExecState } from '../components/OutputPanel'
import { useYjsRoom } from '../hooks/useYjsRoom'
import { runCode } from '../lib/exec'

export function Room() {
  const { slug = '' } = useParams()
  const user = useAuth((s) => s.user)

  const [room, setRoom] = useState<RoomType | null>(null)
  const [files, setFiles] = useState<FileMeta[]>([])
  const [selected, setSelected] = useState<FileMeta | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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
          ) : (
            // Keyed by file id so switching files remounts the pane — each file
            // gets its own Y.Doc + provider, and no content bleeds across files.
            <EditorPane key={selected.id} slug={slug} file={selected} canEdit={canEdit} />
          )}
        </main>
      </div>
    </div>
  )
}

// EditorPane owns the realtime lifecycle for a single file. It connects a
// Y.Doc to the sidecar (via the authenticated Go proxy) and seeds the doc from
// the file's saved content the first time it's opened.
function EditorPane({ slug, file, canEdit }: { slug: string; file: FileMeta; canEdit: boolean }) {
  const { doc, provider, status } = useYjsRoom(slug, file.id)
  const [initialContent, setInitialContent] = useState<string | null>(null)

  // Execution state.
  const [chunks, setChunks] = useState<OutputChunk[]>([])
  const [execState, setExecState] = useState<ExecState>({ kind: 'idle' })
  const wsRef = useRef<WebSocket | null>(null)
  const canRun = file.language === 'python'

  useEffect(() => {
    let alive = true
    getFileContent(slug, file.id)
      .then((res) => alive && setInitialContent(res.content))
      .catch(() => alive && setInitialContent(''))
    return () => {
      alive = false
    }
  }, [slug, file.id])

  // Tear down any live execution socket when switching files / unmounting.
  useEffect(() => () => wsRef.current?.close(), [])

  async function handleRun() {
    if (!canRun || execState.kind === 'running') return
    const code = doc.getText('content').toString()
    setChunks([])
    setExecState({ kind: 'running' })
    wsRef.current = await runCode(slug, file.id, code, {
      onOutput: (stream, data) => setChunks((cs) => [...cs, { stream, data }]),
      onExit: (codeNum, timedOut) => setExecState({ kind: 'exited', code: codeNum, timedOut }),
      onError: (message) => setExecState({ kind: 'error', message }),
    })
  }

  // Wait for the saved content to load before mounting the binding, so the
  // seed-if-empty check sees the real starting text.
  if (initialContent === null) return <Placeholder>Loading {file.path}…</Placeholder>

  const running = execState.kind === 'running'

  return (
    <>
      <EditorToolbar
        path={file.path}
        language={file.language}
        status={status}
        canEdit={canEdit}
        canRun={canRun}
        running={running}
        onRun={handleRun}
      />
      <div style={{ flex: 1, minHeight: 0 }}>
        <CollaborativeEditor
          doc={doc}
          provider={provider}
          language={file.language}
          readOnly={!canEdit}
          initialContent={initialContent}
          onRun={handleRun}
        />
      </div>
      <OutputPanel chunks={chunks} state={execState} onClear={() => { setChunks([]); setExecState({ kind: 'idle' }) }} />
    </>
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
