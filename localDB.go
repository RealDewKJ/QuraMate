package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

type LocalDB struct {
	conn *sql.DB
}

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
	l.CleanupOldQueries()
	return l, nil
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
		return "", nil // Return empty string if not found
	}
	return value, err
}

func (l *LocalDB) SaveQuery(query string, dbType string) error {
	_, err := l.conn.Exec(`INSERT INTO query_history (query, db_type, is_favorite) VALUES (?, ?, 0)`, query, dbType)
	return err
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

func (l *LocalDB) CleanupOldQueries() error {
	// Keep favorites, delete others older than 30 days
	_, err := l.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0 AND timestamp < datetime('now', '-30 days')`)
	if err != nil {
		log.Printf("Failed to cleanup old queries: %v", err)
	}
	return err
}

func (l *LocalDB) Close() error {
	return l.conn.Close()
}
