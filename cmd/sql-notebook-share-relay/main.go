package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"QuraMate/internal/notebookshare/relay"
)

func main() {
	port := envOrDefault("QURAMATE_SHARE_RELAY_PORT", "8787")
	dbPath := envOrDefault("QURAMATE_SHARE_RELAY_DB", defaultDBPath())

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("create relay data directory: %v", err)
	}

	server, err := relay.NewServer(dbPath)
	if err != nil {
		log.Fatalf("start relay: %v", err)
	}
	defer server.Close()

	address := ":" + port
	log.Printf("SQL Notebook share relay listening on %s", address)
	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatalf("serve relay: %v", err)
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func defaultDBPath() string {
	return filepath.Join(".", "data", "sql-notebook-share-relay.db")
}
