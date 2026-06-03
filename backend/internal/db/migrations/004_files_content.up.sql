-- Stage 5: single-user editing stores raw file content. This is a temporary
-- mechanism — Stage 6 replaces it with Yjs CRDT state in the yjs_state column.
ALTER TABLE files ADD COLUMN content TEXT NOT NULL DEFAULT '';
