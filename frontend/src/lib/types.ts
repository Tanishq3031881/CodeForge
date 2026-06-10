export interface Room {
  id: string
  slug: string
  name: string
  owner_id: string
  is_public: boolean
  created_at: string
}

export interface FileMeta {
  id: string
  room_id: string
  path: string
  language: string
  updated_at: string
}

export interface RoomWithFiles {
  room: Room
  files: FileMeta[]
}

// Languages the editor + (later) sandbox understand. Mirrors the backend's
// `supported` set in files/service.go.
export const LANGUAGES = [
  'python',
  'javascript',
  'typescript',
  'go',
  'rust',
  'plaintext',
] as const

export type Language = (typeof LANGUAGES)[number]
