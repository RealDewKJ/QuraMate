package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/zalando/go-keyring"
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
	conn              *sql.DB
	hasFTS5           bool
	encryptionEnabled bool
	mu                sync.RWMutex
}

const defaultQueryHistoryRetentionDays = 30
const localDataEncryptionMetadataKey = "__local_data_encryption_enabled"
const localDataEncryptionKeyringService = "QuraMate-LocalData"
const localDataEncryptionKeyringAccount = "default"
const localDataEncryptionPrefix = "enc:v1:"

var sensitiveQuotedValuePattern = regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[_-]?key|access[_-]?key|private[_-]?key)\b\s*(=|:)\s*('(?:''|[^'])*'|"(?:\\"|[^"])*")`)
var sensitiveBareValuePattern = regexp.MustCompile(`(?i)\b(password|passwd|pwd|token|secret|api[_-]?key|access[_-]?key|private[_-]?key)\b\s*(=|:)\s*([^\s,;)\]}]+)`)
var sensitiveURLCredentialPattern = regexp.MustCompile(`(?i)\b((?:postgres(?:ql)?|mysql|mssql|sqlserver):\/\/[^:\s\/]+:)([^@\s\/]+)@`)
var localDataKeyringGet = keyring.Get
var localDataKeyringSet = keyring.Set
var localDataKeyringDelete = keyring.Delete

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
	return newLocalDBWithPath(dbPath)
}

func newLocalDBWithPath(dbPath string) (*LocalDB, error) {
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
	if err := l.loadEncryptionMetadataLocked(); err != nil {
		return nil, err
	}
	if err := l.initQueryHistoryIndexes(); err != nil {
		return nil, err
	}
	l.initQueryHistoryFTS()
	l.CleanupOldQueries(defaultQueryHistoryRetentionDays)
	if chmodErr := os.Chmod(dbPath, 0o600); chmodErr != nil {
		log.Printf("Unable to tighten local DB file permissions: %v", chmodErr)
	}
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
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.saveSettingLocked(key, value)
}

func (l *LocalDB) saveSettingLocked(key string, value string) error {
	if key != localDataEncryptionMetadataKey {
		encryptedValue, err := l.encryptValueIfNeededLocked(value)
		if err != nil {
			return err
		}
		value = encryptedValue
	}
	_, err := l.conn.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value
	`, key, value)
	return err
}

func (l *LocalDB) LoadSetting(key string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var value string
	err := l.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if key == localDataEncryptionMetadataKey {
		return value, nil
	}
	return l.decryptValueIfNeededLocked(value)
}

func (l *LocalDB) SaveQuery(query string, dbType string, retentionDays int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		return nil
	}
	redactedQuery := redactSensitiveQuery(normalizedQuery)
	storedQuery, err := l.encryptValueIfNeededLocked(redactedQuery)
	if err != nil {
		return err
	}
	_, err = l.conn.Exec(`INSERT INTO query_history (query, db_type, is_favorite) VALUES (?, ?, 0)`, storedQuery, dbType)
	if err != nil {
		return err
	}

	if cleanupErr := l.cleanupOldQueriesLocked(retentionDays); cleanupErr != nil {
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
	l.mu.RLock()
	defer l.mu.RUnlock()
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
		entry.Query, err = l.decryptValueIfNeededLocked(entry.Query)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (l *LocalDB) SearchQueries(queryText string, dbType string, favoritesOnly bool, dateRange string, sortMode string, limit int) ([]QueryHistoryEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
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
	needsPostFilter := l.encryptionEnabled && cleanedQuery != ""
	if cleanedQuery != "" && !needsPostFilter {
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

	if limit > 0 && !needsPostFilter {
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
		entry.Query, err = l.decryptValueIfNeededLocked(entry.Query)
		if err != nil {
			return nil, err
		}
		if needsPostFilter {
			normalizedSearch := strings.ToLower(cleanedQuery)
			if !strings.Contains(strings.ToLower(entry.Query), normalizedSearch) && !strings.Contains(strings.ToLower(entry.DBType), normalizedSearch) {
				continue
			}
		}
		entries = append(entries, entry)
		if needsPostFilter && limit > 0 && len(entries) >= limit {
			break
		}
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
	l.mu.Lock()
	defer l.mu.Unlock()
	val := 0
	if isFavorite {
		val = 1
	}
	_, err := l.conn.Exec(`UPDATE query_history SET is_favorite = ? WHERE id = ?`, val, id)
	return err
}

func (l *LocalDB) DeleteQuery(id int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE id = ?`, id)
	return err
}

func (l *LocalDB) ClearNonFavoriteQueries() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0`)
	return err
}

func (l *LocalDB) GetQueryHistorySummary() (QueryHistorySummary, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
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
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cleanupOldQueriesLocked(retentionDays)
}

func (l *LocalDB) cleanupOldQueriesLocked(retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = defaultQueryHistoryRetentionDays
	}
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0 AND timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	return err
}

func (l *LocalDB) Close() error {
	return l.conn.Close()
}

func (l *LocalDB) IsEncryptionEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.encryptionEnabled
}

func (l *LocalDB) SetEncryptionEnabled(enabled bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.encryptionEnabled == enabled {
		return nil
	}
	if enabled {
		if _, err := l.getOrCreateEncryptionKeyLocked(); err != nil {
			return err
		}
	}

	tx, err := l.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin encryption migration: %w", err)
	}
	defer tx.Rollback()

	if err := l.migrateSettingsLocked(tx, enabled); err != nil {
		return err
	}
	if err := l.migrateQueriesLocked(tx, enabled); err != nil {
		return err
	}
	if err := l.setEncryptionMetadataTx(tx, enabled); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit encryption migration: %w", err)
	}

	l.encryptionEnabled = enabled
	if !enabled {
		if err := localDataKeyringDelete(localDataEncryptionKeyringService, localDataEncryptionKeyringAccount); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			log.Printf("Unable to remove local data encryption key: %v", err)
		}
	}

	return nil
}

func (l *LocalDB) loadEncryptionMetadataLocked() error {
	var value string
	err := l.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, localDataEncryptionMetadataKey).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		l.encryptionEnabled = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("load encryption metadata: %w", err)
	}
	l.encryptionEnabled = strings.EqualFold(strings.TrimSpace(value), "true")
	return nil
}

func (l *LocalDB) setEncryptionMetadataTx(tx *sql.Tx, enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	_, err := tx.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value
	`, localDataEncryptionMetadataKey, value)
	if err != nil {
		return fmt.Errorf("save encryption metadata: %w", err)
	}
	return nil
}

func (l *LocalDB) migrateSettingsLocked(tx *sql.Tx, enableEncryption bool) error {
	rows, err := tx.Query(`SELECT key, value FROM settings WHERE key <> ?`, localDataEncryptionMetadataKey)
	if err != nil {
		return fmt.Errorf("load settings for encryption migration: %w", err)
	}
	defer rows.Close()

	type settingRow struct {
		key   string
		value string
	}
	var settings []settingRow
	for rows.Next() {
		var row settingRow
		if err := rows.Scan(&row.key, &row.value); err != nil {
			return fmt.Errorf("scan setting for encryption migration: %w", err)
		}
		settings = append(settings, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate settings for encryption migration: %w", err)
	}

	for _, row := range settings {
		convertedValue, err := l.transformValueForMigrationLocked(row.value, enableEncryption)
		if err != nil {
			return fmt.Errorf("migrate setting %s: %w", row.key, err)
		}
		if _, err := tx.Exec(`UPDATE settings SET value = ? WHERE key = ?`, convertedValue, row.key); err != nil {
			return fmt.Errorf("update setting %s: %w", row.key, err)
		}
	}
	return nil
}

func (l *LocalDB) migrateQueriesLocked(tx *sql.Tx, enableEncryption bool) error {
	rows, err := tx.Query(`SELECT id, query FROM query_history`)
	if err != nil {
		return fmt.Errorf("load queries for encryption migration: %w", err)
	}
	defer rows.Close()

	type queryRow struct {
		id    int
		query string
	}
	var queries []queryRow
	for rows.Next() {
		var row queryRow
		if err := rows.Scan(&row.id, &row.query); err != nil {
			return fmt.Errorf("scan query for encryption migration: %w", err)
		}
		queries = append(queries, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate queries for encryption migration: %w", err)
	}

	for _, row := range queries {
		convertedValue, err := l.transformValueForMigrationLocked(row.query, enableEncryption)
		if err != nil {
			return fmt.Errorf("migrate query %d: %w", row.id, err)
		}
		if _, err := tx.Exec(`UPDATE query_history SET query = ? WHERE id = ?`, convertedValue, row.id); err != nil {
			return fmt.Errorf("update query %d: %w", row.id, err)
		}
	}
	return nil
}

func (l *LocalDB) transformValueForMigrationLocked(value string, enableEncryption bool) (string, error) {
	if enableEncryption {
		if strings.HasPrefix(value, localDataEncryptionPrefix) {
			return value, nil
		}
		return l.encryptStringLocked(value)
	}
	return l.decryptValueIfNeededLocked(value)
}

func (l *LocalDB) encryptValueIfNeededLocked(value string) (string, error) {
	if !l.encryptionEnabled {
		return value, nil
	}
	if strings.HasPrefix(value, localDataEncryptionPrefix) {
		return value, nil
	}
	return l.encryptStringLocked(value)
}

func (l *LocalDB) encryptStringLocked(value string) (string, error) {
	key, err := l.getOrCreateEncryptionKeyLocked()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create local encryption cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create local encryption gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate local encryption nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	payload := append(nonce, ciphertext...)
	return localDataEncryptionPrefix + base64.StdEncoding.EncodeToString(payload), nil
}

func (l *LocalDB) decryptValueIfNeededLocked(value string) (string, error) {
	if !strings.HasPrefix(value, localDataEncryptionPrefix) {
		return value, nil
	}
	encodedPayload := strings.TrimPrefix(value, localDataEncryptionPrefix)
	payload, err := base64.StdEncoding.DecodeString(encodedPayload)
	if err != nil {
		return "", fmt.Errorf("decode local encrypted payload: %w", err)
	}
	key, err := l.getEncryptionKeyLocked()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create local decryption cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create local decryption gcm: %w", err)
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("local encrypted payload is too short")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt local data: %w", err)
	}
	return string(plaintext), nil
}

func (l *LocalDB) getOrCreateEncryptionKeyLocked() ([]byte, error) {
	key, err := l.getEncryptionKeyLocked()
	if err == nil {
		return key, nil
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return nil, err
	}
	rawKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, rawKey); err != nil {
		return nil, fmt.Errorf("generate local encryption key: %w", err)
	}
	encodedKey := base64.StdEncoding.EncodeToString(rawKey)
	if err := localDataKeyringSet(localDataEncryptionKeyringService, localDataEncryptionKeyringAccount, encodedKey); err != nil {
		return nil, fmt.Errorf("save local encryption key: %w", err)
	}
	return rawKey, nil
}

func (l *LocalDB) getEncryptionKeyLocked() ([]byte, error) {
	encodedKey, err := localDataKeyringGet(localDataEncryptionKeyringService, localDataEncryptionKeyringAccount)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("load local encryption key: %w", err)
	}
	decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("decode local encryption key: %w", err)
	}
	if len(decodedKey) != 32 {
		return nil, fmt.Errorf("local encryption key has invalid length %d", len(decodedKey))
	}
	return decodedKey, nil
}
