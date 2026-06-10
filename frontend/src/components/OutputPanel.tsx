import { useEffect, useRef } from 'react'

export interface OutputChunk {
  stream: 'stdout' | 'stderr'
  data: string
}

export type ExecState =
  | { kind: 'idle' }
  | { kind: 'running' }
  | { kind: 'exited'; code: number; timedOut: boolean }
  | { kind: 'error'; message: string }

interface Props {
  chunks: OutputChunk[]
  state: ExecState
  onClear: () => void
}

export function OutputPanel({ chunks, state, onClear }: Props) {
  const endRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    endRef.current?.scrollIntoView({ block: 'end' })
  }, [chunks, state])

  return (
    <div style={wrap}>
      <div style={header}>
        <span style={{ color: '#ccc' }}>Output</span>
        <span style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <Status state={state} />
          <button onClick={onClear} disabled={state.kind === 'running'} style={{ fontSize: 11 }}>
            clear
          </button>
        </span>
      </div>
      <pre style={body}>
        {chunks.length === 0 && state.kind === 'idle' && (
          <span style={{ color: '#666' }}>Press Run (or Ctrl/Cmd+Enter) to execute.</span>
        )}
        {chunks.map((c, i) => (
          <span key={i} style={{ color: c.stream === 'stderr' ? '#ff8a8a' : '#d4d4d4' }}>
            {c.data}
          </span>
        ))}
        <div ref={endRef} />
      </pre>
    </div>
  )
}

function Status({ state }: { state: ExecState }) {
  switch (state.kind) {
    case 'running':
      return <span style={{ color: '#e2c08d', fontSize: 12 }}>● running…</span>
    case 'exited':
      if (state.timedOut) return <span style={{ color: '#ff8a8a', fontSize: 12 }}>● timed out (killed at 5s)</span>
      return (
        <span style={{ color: state.code === 0 ? '#6cc070' : '#ff8a8a', fontSize: 12 }}>
          ● exited {state.code}
        </span>
      )
    case 'error':
      return <span style={{ color: '#ff8a8a', fontSize: 12 }}>● {state.message}</span>
    default:
      return null
  }
}

const wrap: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  height: 200,
  borderTop: '1px solid #333',
  background: '#181818',
  fontFamily: 'monospace',
}
const header: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '4px 12px',
  borderBottom: '1px solid #2a2a2a',
  fontSize: 13,
}
const body: React.CSSProperties = {
  flex: 1,
  margin: 0,
  padding: '8px 12px',
  overflow: 'auto',
  fontSize: 13,
  lineHeight: 1.45,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
}
