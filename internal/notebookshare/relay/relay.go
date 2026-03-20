package relay

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const defaultExpiry = 7 * 24 * time.Hour

type CreateRequest struct {
	Scope       string          `json:"scope"`
	SenderLabel string          `json:"senderLabel,omitempty"`
	Payload     json.RawMessage `json:"payload"`
}

type CreateResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expiresAt"`
}

type ResolveResponse struct {
	Code        string          `json:"code"`
	ExpiresAt   string          `json:"expiresAt"`
	SenderLabel string          `json:"senderLabel,omitempty"`
	Payload     json.RawMessage `json:"payload"`
}

type Server struct {
	db *sql.DB
}

func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open relay database: %w", err)
	}

	if err := initializeDB(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Server{db: db}, nil
}

func (s *Server) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")

	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	switch {
	case request.Method == http.MethodPost && request.URL.Path == "/api/notebook-shares":
		s.handleCreate(writer, request)
		return
	case request.Method == http.MethodGet && strings.HasPrefix(request.URL.Path, "/api/notebook-shares/"):
		s.handleResolve(writer, request)
		return
	default:
		writeError(writer, http.StatusNotFound, "Route not found.")
	}
}

func initializeDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS notebook_shares (
			code TEXT PRIMARY KEY,
			scope TEXT NOT NULL,
			sender_label TEXT NOT NULL DEFAULT '',
			payload_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create notebook_shares table: %w", err)
	}

	return nil
}

func (s *Server) handleCreate(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var payload CreateRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeError(writer, http.StatusBadRequest, "Invalid share payload.")
		return
	}

	scope := strings.ToLower(strings.TrimSpace(payload.Scope))
	if scope != "notebook" && scope != "sql" {
		writeError(writer, http.StatusBadRequest, "Scope must be notebook or sql.")
		return
	}

	if len(payload.Payload) == 0 {
		writeError(writer, http.StatusBadRequest, "Payload is required.")
		return
	}

	trimmedPayload := strings.TrimSpace(string(payload.Payload))
	if trimmedPayload == "" {
		writeError(writer, http.StatusBadRequest, "Payload is required.")
		return
	}
	if len(trimmedPayload) > 1_000_000 {
		writeError(writer, http.StatusRequestEntityTooLarge, "Payload is too large.")
		return
	}

	normalizedPayload := normalizePayloadForReuse(trimmedPayload)
	now := time.Now().UTC()
	expiresAt := now.Add(defaultExpiry)

	existingCode, existingExpiry, err := s.findReusableCode(scope, normalizedPayload, now)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Could not check existing share code.")
		return
	}
	if existingCode != "" {
		writeJSON(writer, http.StatusOK, CreateResponse{
			Code:      existingCode,
			ExpiresAt: existingExpiry.Format(time.RFC3339),
		})
		return
	}

	code, err := s.generateUniqueCode()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Could not generate share code.")
		return
	}

	_, err = s.db.Exec(
		`INSERT INTO notebook_shares (code, scope, sender_label, payload_json, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		code,
		scope,
		strings.TrimSpace(payload.SenderLabel),
		normalizedPayload,
		now.Format(time.RFC3339),
		expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Could not save share payload.")
		return
	}

	writeJSON(writer, http.StatusOK, CreateResponse{
		Code:      code,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (s *Server) handleResolve(writer http.ResponseWriter, request *http.Request) {
	code := normalizeCode(strings.TrimPrefix(request.URL.Path, "/api/notebook-shares/"))
	if code == "" {
		writeError(writer, http.StatusBadRequest, "Share code is required.")
		return
	}

	var response ResolveResponse
	var expiresAtRaw string
	err := s.db.QueryRow(
		`SELECT code, expires_at, sender_label, payload_json FROM notebook_shares WHERE code = ?`,
		code,
	).Scan(&response.Code, &expiresAtRaw, &response.SenderLabel, &response.Payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(writer, http.StatusNotFound, "Share code not found.")
			return
		}
		writeError(writer, http.StatusInternalServerError, "Could not resolve share code.")
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Stored share code expiry is invalid.")
		return
	}

	if time.Now().UTC().After(expiresAt) {
		_, _ = s.db.Exec(`DELETE FROM notebook_shares WHERE code = ?`, code)
		writeError(writer, http.StatusGone, "This share code expired. Ask the sender to generate a new one.")
		return
	}

	response.ExpiresAt = expiresAt.Format(time.RFC3339)
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) findReusableCode(scope string, payloadJSON string, now time.Time) (string, time.Time, error) {
	var code string
	var expiresAtRaw string
	err := s.db.QueryRow(
		`
			SELECT code, expires_at
			FROM notebook_shares
			WHERE scope = ? AND payload_json = ? AND expires_at > ?
			ORDER BY expires_at DESC
			LIMIT 1
		`,
		scope,
		payloadJSON,
		now.Format(time.RFC3339),
	).Scan(&code, &expiresAtRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return "", time.Time{}, nil
	}
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

func normalizePayloadForReuse(payloadJSON string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return payloadJSON
	}

	delete(payload, "exportedAt")

	normalized, err := json.Marshal(payload)
	if err != nil {
		return payloadJSON
	}

	return string(normalized)
}

func (s *Server) generateUniqueCode() (string, error) {
	for range 20 {
		code := randomCode()
		var existing string
		err := s.db.QueryRow(`SELECT code FROM notebook_shares WHERE code = ?`, code).Scan(&existing)
		if errors.Is(err, sql.ErrNoRows) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("could not generate unique code")
}

func randomCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	builder := strings.Builder{}
	builder.Grow(9)
	builder.WriteString("QN-")
	for range 6 {
		builder.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}
	return builder.String()
}

func normalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{
		"error": message,
	})
}
