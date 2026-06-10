import { useEffect, useState } from 'react'
import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'
import { useAuth } from '../lib/store'

export type SyncStatus = 'connecting' | 'connected' | 'disconnected'

// Manages the Y.Doc + WebsocketProvider lifecycle for one file.
//
// The Y.Doc is created once per hook instance (useState initializer) and kept
// alive across reconnects — creating a fresh doc on reconnect is the classic
// way to end up with duplicated content after a network blip. Callers remount
// the consuming component (via key={fileId}) when switching files, which gives
// each file its own doc.
export function useYjsRoom(slug: string, fileId: string) {
  const token = useAuth((s) => s.token)
  const [doc] = useState(() => new Y.Doc())
  const [provider, setProvider] = useState<WebsocketProvider | null>(null)
  const [status, setStatus] = useState<SyncStatus>('connecting')

  useEffect(() => {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    // The backend authenticates this connection (JWT via query param, since
    // browsers can't set WebSocket headers) and proxies it to the sidecar.
    // The doc name is the file ID; the slug rides as a query param for authz.
    const p = new WebsocketProvider(`${proto}://${location.host}/ws/yjs`, fileId, doc, {
      params: { token: token ?? '', slug },
    })
    p.on('status', (e: { status: SyncStatus }) => setStatus(e.status))
    setProvider(p)
    return () => {
      p.destroy()
    }
  }, [slug, fileId, doc, token])

  useEffect(() => () => doc.destroy(), [doc])

  return { doc, provider, status }
}
