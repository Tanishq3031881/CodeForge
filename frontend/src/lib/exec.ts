import { useAuth } from './store'

export interface ExecHandlers {
  onOutput: (stream: 'stdout' | 'stderr', data: string) => void
  onExit: (code: number, timedOut: boolean) => void
  onError: (msg: string) => void
  onClose?: () => void
}

// runCode registers an execution (POST), then opens the WebSocket the backend
// streams stdout/stderr/exit over. Returns the socket so the caller can close
// it (e.g. on unmount). The two-step flow guarantees no output is missed: the
// container isn't run until this WS connects.
export async function runCode(
  slug: string,
  fileId: string,
  code: string,
  h: ExecHandlers,
): Promise<WebSocket | null> {
  const token = useAuth.getState().token
  const res = await fetch(`/api/rooms/${slug}/run`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ file_id: fileId, code }),
  })
  if (!res.ok) {
    let msg = `run failed (${res.status})`
    try {
      const b = await res.json()
      if (b?.error) msg = b.error
    } catch {}
    h.onError(msg)
    h.onClose?.()
    return null
  }

  const { execution_id } = await res.json()
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const ws = new WebSocket(`${proto}://${location.host}/ws/exec/${execution_id}?token=${token ?? ''}`)

  ws.onmessage = (e) => {
    const m = JSON.parse(e.data as string)
    if (m.type === 'stdout' || m.type === 'stderr') h.onOutput(m.type, m.data ?? '')
    else if (m.type === 'exit') h.onExit(m.code, !!m.timed_out)
    else if (m.type === 'error') h.onError(m.data ?? 'execution error')
  }
  ws.onerror = () => h.onError('connection error')
  ws.onclose = () => h.onClose?.()
  return ws
}
