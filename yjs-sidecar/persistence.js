// Postgres persistence for the Yjs sidecar, via the Go backend's internal API.
//
// y-websocket's persistence hook gives us two callbacks per document:
//   bindState(docName, ydoc)  — first connection: load saved state, then watch
//                               for updates and flush them on a debounce.
//   writeState(docName, ydoc) — last connection closed: flush a final time.
//
// docName is the file's UUID. State is round-tripped as a Yjs update
// (Y.encodeStateAsUpdate); we also send the decoded plain text so the backend
// can keep files.content current for the content endpoint and the sandbox.

const Y = require('yjs')

const BACKEND = process.env.BACKEND_URL || 'http://127.0.0.1:8080'
const KEY = process.env.INTERNAL_KEY || 'dev-internal-key'
const DEBOUNCE_MS = parseInt(process.env.PERSIST_DEBOUNCE_MS || '2000', 10)

async function loadState(docName) {
  const res = await fetch(`${BACKEND}/internal/files/${docName}/yjs-state`, {
    headers: { 'X-Internal-Key': KEY },
  })
  if (res.status === 204) return null
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const buf = Buffer.from(await res.arrayBuffer())
  return buf.length ? new Uint8Array(buf) : null
}

async function saveState(docName, ydoc) {
  const state = Y.encodeStateAsUpdate(ydoc)
  const text = ydoc.getText('content').toString()
  const res = await fetch(`${BACKEND}/internal/files/${docName}/yjs-state`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Internal-Key': KEY },
    body: JSON.stringify({ state: Buffer.from(state).toString('base64'), text }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}

function makePersistence() {
  // One pending-flush timer per document.
  const timers = new Map()

  return {
    provider: 'codeforge-backend',

    bindState: async (docName, ydoc) => {
      try {
        const state = await loadState(docName)
        if (state) Y.applyUpdate(ydoc, state)
      } catch (e) {
        console.error(`[persist] load ${docName} failed: ${e.message}`)
      }
      // Coalesce bursts of keystrokes into one save per DEBOUNCE_MS window.
      ydoc.on('update', () => {
        if (timers.has(docName)) return
        timers.set(
          docName,
          setTimeout(async () => {
            timers.delete(docName)
            try {
              await saveState(docName, ydoc)
            } catch (e) {
              console.error(`[persist] save ${docName} failed: ${e.message}`)
            }
          }, DEBOUNCE_MS),
        )
      })
    },

    writeState: async (docName, ydoc) => {
      const t = timers.get(docName)
      if (t) {
        clearTimeout(t)
        timers.delete(docName)
      }
      try {
        await saveState(docName, ydoc)
      } catch (e) {
        console.error(`[persist] final save ${docName} failed: ${e.message}`)
      }
    },
  }
}

module.exports = { makePersistence }
