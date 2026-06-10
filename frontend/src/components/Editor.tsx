import { useEffect, useRef } from 'react'
import Editor, { type OnMount } from '@monaco-editor/react'

interface Props {
  language: string
  value: string
  readOnly?: boolean
  onChange: (value: string) => void
  onSave: () => void
}

// Monaco's language IDs mostly match ours; map the ones that differ.
const monacoLang: Record<string, string> = {
  plaintext: 'plaintext',
  python: 'python',
  javascript: 'javascript',
  typescript: 'typescript',
  go: 'go',
  rust: 'rust',
}

export function CodeEditor({ language, value, readOnly = false, onChange, onSave }: Props) {
  // Keep the latest onSave in a ref so the keybinding (registered once at mount)
  // never calls a stale closure.
  const saveRef = useRef(onSave)
  useEffect(() => {
    saveRef.current = onSave
  }, [onSave])

  const handleMount: OnMount = (editor, monaco) => {
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      saveRef.current()
    })
  }

  return (
    <Editor
      height="100%"
      theme="vs-dark"
      language={monacoLang[language] ?? 'plaintext'}
      value={value}
      onChange={(v) => onChange(v ?? '')}
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
