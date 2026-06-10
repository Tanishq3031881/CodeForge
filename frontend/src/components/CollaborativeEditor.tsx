import { useEffect, useRef } from 'react'
import Editor, { type OnMount } from '@monaco-editor/react'
import { MonacoBinding } from 'y-monaco'
import * as Y from 'yjs'
import type { WebsocketProvider } from 'y-websocket'

interface Props {
  doc: Y.Doc
  provider: WebsocketProvider | null
  language: string
  readOnly?: boolean
  // Seeded into the shared doc by the first client that finds it empty, so a
  // file saved before realtime existed still shows its content. Guarded by a
  // length check, so once the sidecar seeds server-side (Stage 7) this is inert.
  initialContent?: string
  onRun?: () => void
}

const monacoLang: Record<string, string> = {
  plaintext: 'plaintext',
  python: 'python',
  javascript: 'javascript',
  typescript: 'typescript',
  go: 'go',
  rust: 'rust',
}

export function CollaborativeEditor({
  doc,
  provider,
  language,
  readOnly = false,
  initialContent,
  onRun,
}: Props) {
  const bindingRef = useRef<MonacoBinding | null>(null)
  const onRunRef = useRef(onRun)
  useEffect(() => {
    onRunRef.current = onRun
  }, [onRun])

  const handleMount: OnMount = (editor, monaco) => {
    const yText = doc.getText('content')
    const model = editor.getModel()
    if (!model) return

    // Bind Monaco's model to the shared Y.Text. Remote edits flow into the
    // model directly (bypassing readOnly), so viewers still see live updates.
    bindingRef.current = new MonacoBinding(
      yText,
      model,
      new Set([editor]),
      provider?.awareness ?? null,
    )

    // Ctrl/Cmd+Enter runs the file (wired in Stage 8).
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      onRunRef.current?.()
    })

    if (!readOnly && initialContent && yText.length === 0) {
      // Seed once. If two clients race here the CRDT merges both inserts into
      // duplicated text, but for a fresh doc with one editor this is safe; the
      // server-side seed in Stage 7 removes the race entirely.
      const seedOnce = (synced: boolean) => {
        if (synced && yText.length === 0 && initialContent) {
          yText.insert(0, initialContent)
        }
      }
      if (provider?.synced) seedOnce(true)
      else provider?.once('sync', seedOnce)
    }
  }

  useEffect(() => {
    return () => {
      bindingRef.current?.destroy()
      bindingRef.current = null
    }
  }, [])

  return (
    <Editor
      height="100%"
      theme="vs-dark"
      language={monacoLang[language] ?? 'plaintext'}
      onMount={handleMount}
      options={{
        fontSize: 14,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        automaticLayout: true,
        readOnly,
      }}
    />
  )
}
