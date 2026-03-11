package main

import (
	"database/sql"
	"strings"
	"testing"
)

func newTestLocalDB(t *testing.T) *LocalDB {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE query_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL,
			db_type TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_favorite BOOLEAN DEFAULT 0
		);
	`)
	if err != nil {
		t.Fatalf("create query_history: %v", err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &LocalDB{conn: conn}
}

func TestRedactSensitiveQuery(t *testing.T) {
	input := `
UPDATE users SET password='my-secret', token: "abc123", api_key=plainKey;
-- DSN
SELECT * FROM x WHERE dsn='postgres://app:pw123@localhost:5432/db';
`

	redacted := redactSensitiveQuery(input)

	for _, secret := range []string{"my-secret", "abc123", "plainKey", "pw123"} {
		if strings.Contains(redacted, secret) {
			t.Fatalf("secret %q should be redacted: %s", secret, redacted)
		}
	}
	if !strings.Contains(redacted, "[REDACTED]") {
		t.Fatalf("expected redaction marker in query: %s", redacted)
	}
}

func TestSaveQueryRetentionKeepsFavorites(t *testing.T) {
	ldb := newTestLocalDB(t)

	_, err := ldb.conn.Exec(`
		INSERT INTO query_history (query, db_type, timestamp, is_favorite) VALUES
		('old non-favorite', 'pg', datetime('now', '-40 days'), 0),
		('old favorite', 'pg', datetime('now', '-40 days'), 1)
	`)
	if err != nil {
		t.Fatalf("seed rows: %v", err)
	}

	if err := ldb.SaveQuery("SELECT * FROM users WHERE password='supersecret'", "pg", 30); err != nil {
		t.Fatalf("save query: %v", err)
	}

	var total int
	if err := ldb.conn.QueryRow(`SELECT COUNT(*) FROM query_history`).Scan(&total); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 rows after retention cleanup, got %d", total)
	}

	var favoriteCount int
	if err := ldb.conn.QueryRow(`SELECT COUNT(*) FROM query_history WHERE is_favorite = 1`).Scan(&favoriteCount); err != nil {
		t.Fatalf("count favorites: %v", err)
	}
	if favoriteCount != 1 {
		t.Fatalf("expected favorite row to remain, got %d", favoriteCount)
	}

	var latestQuery string
	if err := ldb.conn.QueryRow(`SELECT query FROM query_history ORDER BY id DESC LIMIT 1`).Scan(&latestQuery); err != nil {
		t.Fatalf("load latest query: %v", err)
	}
	if strings.Contains(latestQuery, "supersecret") {
		t.Fatalf("expected latest query to be redacted, got %q", latestQuery)
	}
}

func TestSaveQueryUsesDefaultRetentionWhenInvalid(t *testing.T) {
	ldb := newTestLocalDB(t)

	_, err := ldb.conn.Exec(`
		INSERT INTO query_history (query, db_type, timestamp, is_favorite)
		VALUES ('stale query', 'pg', datetime('now', '-40 days'), 0)
	`)
	if err != nil {
		t.Fatalf("seed stale row: %v", err)
	}

	if err := ldb.SaveQuery("SELECT 1", "pg", 0); err != nil {
		t.Fatalf("save query with invalid retention: %v", err)
	}

	var staleCount int
	if err := ldb.conn.QueryRow(`SELECT COUNT(*) FROM query_history WHERE query = 'stale query'`).Scan(&staleCount); err != nil {
		t.Fatalf("count stale rows: %v", err)
	}
	if staleCount != 0 {
		t.Fatalf("expected stale non-favorite row to be removed, got %d", staleCount)
	}
}
