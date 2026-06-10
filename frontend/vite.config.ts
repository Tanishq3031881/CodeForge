  import { defineConfig } from 'vite'
  import react from '@vitejs/plugin-react'

  export default defineConfig({
    plugins: [react()],
    server: {
      proxy: {
        '/api': 'http://localhost:8080',
        '/health': 'http://localhost:8080',
        // WebSocket routes (Yjs sync + code-execution output) must be proxied
        // with ws:true so the browser can reach the Go backend through Vite.
        '/ws': { target: 'ws://localhost:8080', ws: true },
      },
    },
  })