package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type QueryHistoryEntry struct {
	ID         int    `json:"id"`
	Query      string `json:"query"`
	DBType     string `json:"db_type"`
	Timestamp  string `json:"timestamp"`
	IsFavorite bool   `json:"is_favorite"`
}

type HistoryDB struct {
	conn *sql.DB
}

func NewHistoryDB() (*HistoryDB, error) {
	// Store in user config dir
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get config dir: %v", err)
	}

	appDir := filepath.Join(configDir, "QuraDB")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create app dir: %v", err)
	}

	dbPath := filepath.Join(appDir, "history.db")
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
		)
	`)
	if err != nil {
		return nil, err
	}

	h := &HistoryDB{conn: conn}
	h.CleanupOldQueries()
	return h, nil
}

func (h *HistoryDB) SaveQuery(query string, dbType string) error {
	_, err := h.conn.Exec(`INSERT INTO query_history (query, db_type, is_favorite) VALUES (?, ?, 0)`, query, dbType)
	return err
}

func (h *HistoryDB) GetQueries(dbType string) ([]QueryHistoryEntry, error) {
	var rows *sql.Rows
	var err error

	if dbType == "" || dbType == "all" {
		rows, err = h.conn.Query(`SELECT id, query, db_type, timestamp, is_favorite FROM query_history ORDER BY id DESC`)
	} else {
		rows, err = h.conn.Query(`SELECT id, query, db_type, timestamp, is_favorite FROM query_history WHERE db_type = ? ORDER BY id DESC`, dbType)
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

func (h *HistoryDB) ToggleFavorite(id int, isFavorite bool) error {
	val := 0
	if isFavorite {
		val = 1
	}
	_, err := h.conn.Exec(`UPDATE query_history SET is_favorite = ? WHERE id = ?`, val, id)
	return err
}

func (h *HistoryDB) DeleteQuery(id int) error {
	_, err := h.conn.Exec(`DELETE FROM query_history WHERE id = ?`, id)
	return err
}

func (h *HistoryDB) CleanupOldQueries() error {
	// Keep favorites, delete others older than 30 days
	_, err := h.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0 AND timestamp < datetime('now', '-30 days')`)
	if err != nil {
		log.Printf("Failed to cleanup old queries: %v", err)
	}
	return err
}

func (h *HistoryDB) Close() error {
	return h.conn.Close()
}
