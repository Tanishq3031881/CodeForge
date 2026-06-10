// CodeForge Yjs sidecar.
//
// A tiny WebSocket server that speaks the Yjs sync protocol via y-websocket.
// It trusts every connection it receives: authentication and authorization
// happen in the Go backend, which verifies the JWT and room access before
// proxying the connection here. Therefore this server MUST only listen on
// localhost (dev) or an internal Docker network (prod) — never the public
// internet.
//
// Doc names are file IDs (UUIDs), set by the Go proxy when it rewrites the
// request path.

const http = require('http')
const WebSocket = require('ws')
const { setupWSConnection, setPersistence } = require('y-websocket/bin/utils')
const { makePersistence } = require('./persistence')

// Persist documents to Postgres through the Go backend's internal API. Load
// failures are tolerated (a missing backend just means no prior state), so
// enabling this unconditionally is safe even in a backend-less dev run.
setPersistence(makePersistence())

const host = process.env.HOST || '127.0.0.1'
const port = parseInt(process.env.PORT || '1234', 10)

// Plain HTTP requests get a health response so docker/compose can probe us.
const server = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'application/json' })
  res.end(JSON.stringify({ status: 'ok' }))
})

const wss = new WebSocket.Server({ server })

wss.on('connection', (conn, req) => {
  // Path is /<docName>; strip the leading slash and any query string.
  const docName = req.url.slice(1).split('?')[0]
  if (!docName) {
    conn.close()
    return
  }
  setupWSConnection(conn, req, { docName })
})

server.listen(port, host, () => {
  console.log(`yjs sidecar listening on ${host}:${port}`)
})
