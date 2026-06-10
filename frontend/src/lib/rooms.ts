import { api } from './api'
import type { FileMeta, Room, RoomWithFiles } from './types'

export function listRooms() {
  return api<Room[]>('/api/rooms')
}

export function createRoom(name: string, isPublic = false) {
  return api<Room>('/api/rooms', {
    method: 'POST',
    body: JSON.stringify({ name, is_public: isPublic }),
  })
}

export function getRoom(slug: string) {
  return api<RoomWithFiles>(`/api/rooms/${slug}`)
}

export function deleteRoom(slug: string) {
  return api<void>(`/api/rooms/${slug}`, { method: 'DELETE' })
}

export function createFile(slug: string, path: string, language: string) {
  return api<FileMeta>(`/api/rooms/${slug}/files`, {
    method: 'POST',
    body: JSON.stringify({ path, language }),
  })
}

export function deleteFile(slug: string, fileId: string) {
  return api<void>(`/api/rooms/${slug}/files/${fileId}`, { method: 'DELETE' })
}

export function getFileContent(slug: string, fileId: string) {
  return api<{ content: string }>(`/api/rooms/${slug}/files/${fileId}/content`)
}

export function saveFileContent(slug: string, fileId: string, content: string) {
  return api<void>(`/api/rooms/${slug}/files/${fileId}/content`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
  })
}
