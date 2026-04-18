CREATE TABLE files (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id    UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    path       TEXT NOT NULL,
    language   TEXT NOT NULL,
    yjs_state  BYTEA,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(room_id, path)
);
CREATE INDEX idx_files_room ON files(room_id);
