CREATE TABLE pets (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    species    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
