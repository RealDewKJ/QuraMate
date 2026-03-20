CREATE TABLE IF NOT EXISTS notebook_shares (
    code TEXT PRIMARY KEY,
    scope TEXT NOT NULL,
    sender_label TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notebook_shares_expires_at
    ON notebook_shares (expires_at);

CREATE TABLE IF NOT EXISTS live_notebook_sessions (
    code TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    sender_label TEXT NOT NULL DEFAULT '',
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_live_notebook_sessions_expires_at
    ON live_notebook_sessions (expires_at);
