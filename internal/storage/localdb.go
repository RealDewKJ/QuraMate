package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

type QueryHistoryEntry struct {
	ID         int    `json:"id"`
	Query      string `json:"query"`
	DBType     string `json:"db_type"`
	Timestamp  string `json:"timestamp"`
	IsFavorite bool   `json:"is_favorite"`
}

type QueryHistorySummary struct {
	Total   int      `json:"total"`
	DBTypes []string `json:"db_types"`
}

type LocalDB struct {
	conn    *sql.DB
	hasFTS5 bool
}

const defaultQueryHistoryRetentionDays = 30

var sensitiveQuotedValuePattern = regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[_-]?key|access[_-]?key|private[_-]?key)\b\s*(=|:)\s*('(?:''|[^'])*'|"(?:\\"|[^"])*")`)
var sensitiveBareValuePattern = regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[_-]?key|access[_-]?key|private[_-]?key)\b\s*(=|:)\s*([^\s,;)\]}]+)`)
var sensitiveURLCredentialPattern = regexp.MustCompile(`(?i)\b((?:postgres(?:ql)?|mysql|mssql|sqlserver):\/\/[^:\s\/]+:)([^@\s\/]+)@`)

func NewLocalDB() (*LocalDB, error) {
	// Store next to the executable
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not get executable path: %v", err)
	}
	appDir := filepath.Dir(execPath)

	// Handles `go run` or `wails dev` which runs from a temp directory
	if strings.Contains(appDir, "Temp") || strings.Contains(appDir, "tmp") {
		appDir, _ = os.Getwd()
	}

	dbPath := filepath.Join(appDir, "quramate.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	initialized := false
	defer func() {
		if !initialized {
			_ = conn.Close()
		}
	}()
	if chmodErr := os.Chmod(dbPath, 0o600); chmodErr != nil {
		log.Printf("Unable to tighten local DB file permissions: %v", chmodErr)
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS query_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL,
			db_type TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_favorite BOOLEAN DEFAULT 0
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	l := &LocalDB{conn: conn}
	if err := l.initQueryHistoryIndexes(); err != nil {
		return nil, err
	}
	l.initQueryHistoryFTS()
	l.CleanupOldQueries(defaultQueryHistoryRetentionDays)
	initialized = true
	return l, nil
}

func (l *LocalDB) initQueryHistoryIndexes() error {
	_, err := l.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_query_history_db_type ON query_history(db_type);
		CREATE INDEX IF NOT EXISTS idx_query_history_timestamp ON query_history(timestamp);
		CREATE INDEX IF NOT EXISTS idx_query_history_is_favorite ON query_history(is_favorite);
	`)
	if err != nil {
		return fmt.Errorf("create query history indexes: %w", err)
	}
	return nil
}

func (l *LocalDB) initQueryHistoryFTS() {
	_, err := l.conn.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS query_history_fts
		USING fts5(query, db_type, content='query_history', content_rowid='id');
	`)
	if err != nil {
		// FTS5 can be unavailable on some builds; keep search working via LIKE fallback.
		log.Printf("Query history FTS disabled: %v", err)
		l.hasFTS5 = false
		return
	}

	_, err = l.conn.Exec(`
		CREATE TRIGGER IF NOT EXISTS query_history_ai AFTER INSERT ON query_history BEGIN
			INSERT INTO query_history_fts(rowid, query, db_type) VALUES (new.id, new.query, new.db_type);
		END;
		CREATE TRIGGER IF NOT EXISTS query_history_ad AFTER DELETE ON query_history BEGIN
			INSERT INTO query_history_fts(query_history_fts, rowid, query, db_type) VALUES ('delete', old.id, old.query, old.db_type);
		END;
		CREATE TRIGGER IF NOT EXISTS query_history_au AFTER UPDATE OF query, db_type ON query_history BEGIN
			INSERT INTO query_history_fts(query_history_fts, rowid, query, db_type) VALUES ('delete', old.id, old.query, old.db_type);
			INSERT INTO query_history_fts(rowid, query, db_type) VALUES (new.id, new.query, new.db_type);
		END;
	`)
	if err != nil {
		log.Printf("Query history FTS triggers disabled: %v", err)
		l.hasFTS5 = false
		return
	}

	_, err = l.conn.Exec(`INSERT INTO query_history_fts(query_history_fts) VALUES('rebuild')`)
	if err != nil {
		log.Printf("Query history FTS rebuild failed: %v", err)
		l.hasFTS5 = false
		return
	}

	l.hasFTS5 = true
}

func (l *LocalDB) SaveSetting(key string, value string) error {
	_, err := l.conn.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value
	`, key, value)
	return err
}

func (l *LocalDB) LoadSetting(key string) (string, error) {
	var value string
	err := l.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (l *LocalDB) SaveQuery(query string, dbType string, retentionDays int) error {
	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		return nil
	}
	redactedQuery := redactSensitiveQuery(normalizedQuery)
	_, err := l.conn.Exec(`INSERT INTO query_history (query, db_type, is_favorite) VALUES (?, ?, 0)`, redactedQuery, dbType)
	if err != nil {
		return err
	}

	if cleanupErr := l.CleanupOldQueries(retentionDays); cleanupErr != nil {
		log.Printf("Failed to cleanup query history: %v", cleanupErr)
	}
	return err
}

func redactSensitiveQuery(query string) string {
	redacted := sensitiveQuotedValuePattern.ReplaceAllString(query, `$1$2'[REDACTED]'`)
	redacted = sensitiveBareValuePattern.ReplaceAllString(redacted, `$1$2[REDACTED]`)
	redacted = sensitiveURLCredentialPattern.ReplaceAllString(redacted, `$1[REDACTED]@`)
	return redacted
}

func (l *LocalDB) GetQueries(dbType string) ([]QueryHistoryEntry, error) {
	var rows *sql.Rows
	var err error

	if dbType == "" || dbType == "all" {
		rows, err = l.conn.Query(`SELECT id, query, db_type, timestamp, is_favorite FROM query_history ORDER BY id DESC`)
	} else {
		rows, err = l.conn.Query(`SELECT id, query, db_type, timestamp, is_favorite FROM query_history WHERE db_type = ? ORDER BY id DESC`, dbType)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []QueryHistoryEntry
	for rows.Next() {
		var entry QueryHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.Query, &entry.DBType, &entry.Timestamp, &entry.IsFavorite); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (l *LocalDB) SearchQueries(queryText string, dbType string, favoritesOnly bool, dateRange string, sortMode string, limit int) ([]QueryHistoryEntry, error) {
	baseQuery := `SELECT id, query, db_type, timestamp, is_favorite FROM query_history`
	var conditions []string
	var args []interface{}

	if dbType != "" && dbType != "all" {
		conditions = append(conditions, "db_type = ?")
		args = append(args, dbType)
	}

	if favoritesOnly {
		conditions = append(conditions, "is_favorite = 1")
	}

	switch dateRange {
	case "today":
		conditions = append(conditions, "datetime(timestamp, 'localtime') >= datetime('now', 'localtime', 'start of day')")
	case "7d":
		conditions = append(conditions, "datetime(timestamp, 'localtime') >= datetime('now', 'localtime', '-7 days')")
	case "30d":
		conditions = append(conditions, "datetime(timestamp, 'localtime') >= datetime('now', 'localtime', '-30 days')")
	}

	cleanedQuery := strings.TrimSpace(queryText)
	if cleanedQuery != "" {
		appliedTextFilter := false
		if l.hasFTS5 {
			ftsQuery := buildFTSQuery(cleanedQuery)
			if ftsQuery != "" {
				conditions = append(conditions, "id IN (SELECT rowid FROM query_history_fts WHERE query_history_fts MATCH ?)")
				args = append(args, ftsQuery)
				appliedTextFilter = true
			}
		}
		if !appliedTextFilter {
			conditions = append(conditions, "(LOWER(query) LIKE LOWER(?) OR LOWER(db_type) LIKE LOWER(?))")
			likePattern := "%" + cleanedQuery + "%"
			args = append(args, likePattern, likePattern)
		}
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)
	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	orderBy := " ORDER BY is_favorite DESC, id DESC"
	if sortMode == "oldest" {
		orderBy = " ORDER BY is_favorite DESC, id ASC"
	}
	queryBuilder.WriteString(orderBy)

	if limit > 0 {
		queryBuilder.WriteString(" LIMIT ?")
		args = append(args, limit)
	}

	rows, err := l.conn.Query(queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []QueryHistoryEntry
	for rows.Next() {
		var entry QueryHistoryEntry
		if err := rows.Scan(&entry.ID, &entry.Query, &entry.DBType, &entry.Timestamp, &entry.IsFavorite); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func buildFTSQuery(searchText string) string {
	tokens := strings.Fields(searchText)
	var parts []string
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		escaped := strings.ReplaceAll(trimmed, `"`, `""`)
		parts = append(parts, fmt.Sprintf(`"%s"*`, escaped))
	}
	return strings.Join(parts, " AND ")
}

func (l *LocalDB) ToggleFavorite(id int, isFavorite bool) error {
	val := 0
	if isFavorite {
		val = 1
	}
	_, err := l.conn.Exec(`UPDATE query_history SET is_favorite = ? WHERE id = ?`, val, id)
	return err
}

func (l *LocalDB) DeleteQuery(id int) error {
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE id = ?`, id)
	return err
}

func (l *LocalDB) ClearNonFavoriteQueries() error {
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0`)
	return err
}

func (l *LocalDB) GetQueryHistorySummary() (QueryHistorySummary, error) {
	summary := QueryHistorySummary{
		DBTypes: []string{},
	}

	if err := l.conn.QueryRow(`SELECT COUNT(*) FROM query_history`).Scan(&summary.Total); err != nil {
		return summary, err
	}

	rows, err := l.conn.Query(`SELECT DISTINCT db_type FROM query_history ORDER BY db_type ASC`)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var dbType string
		if err := rows.Scan(&dbType); err != nil {
			return summary, err
		}
		summary.DBTypes = append(summary.DBTypes, dbType)
	}

	if err := rows.Err(); err != nil {
		return summary, err
	}

	return summary, nil
}

func (l *LocalDB) CleanupOldQueries(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = defaultQueryHistoryRetentionDays
	}
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0 AND timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	return err
}

func (l *LocalDB) Close() error {
	return l.conn.Close()
}
