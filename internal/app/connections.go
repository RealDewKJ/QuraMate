package app

import (
	"QuraMate/internal/database"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type ConnectResult struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

type SSHHostKeyInfo struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Pattern     string `json:"pattern"`
	KeyType     string `json:"keyType"`
	Fingerprint string `json:"fingerprint"`
	Error       string `json:"error"`
}

type ActionResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

const (
	connectionErrorCodeSSHHostUntrusted   = "SSH_HOST_UNTRUSTED"
	connectionErrorCodeSSHHostKeyMismatch = "SSH_HOST_KEY_MISMATCH"
)

func classifyConnectionErrorCode(err error) string {
	if err == nil {
		return ""
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "ssh host key for") && strings.Contains(message, "not trusted") {
		return connectionErrorCodeSSHHostUntrusted
	}
	if strings.Contains(message, "ssh host key mismatch") {
		return connectionErrorCodeSSHHostKeyMismatch
	}

	return ""
}

func formatConnectionError(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	code := classifyConnectionErrorCode(err)
	if code != "" {
		return fmt.Sprintf("Error [%s]: %s", code, message)
	}

	lowerMessage := strings.ToLower(message)
	if strings.Contains(lowerMessage, "forcibly closed by the remote host") ||
		strings.Contains(lowerMessage, "wsarecv") {
		return fmt.Sprintf(
			"Error: %s. The database server closed the connection during setup. For SQL Server, check Settings > Trust SQL Server Certificates By Default, verify the server allows encrypted connections, and confirm the server supports the negotiated TLS version.",
			message,
		)
	}

	return fmt.Sprintf("Error: %s", message)
}

func (a *App) ConnectDB(config DBConfig) ConnectResult {
	if config.ID != "" {
		if config.Password == "" {
			if pw, err := keyring.Get("QuraMate", config.ID); err == nil {
				config.Password = pw
			}
		}
		if config.SSHEnabled && config.SSHPassword == "" {
			if sshPw, err := keyring.Get("QuraMate-SSH", config.ID); err == nil {
				config.SSHPassword = sshPw
			}
		}
	}

	newDB := database.NewDatabase(a.logEvent)
	err := newDB.Connect(config)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to connect to %s: %s", config.Type, err.Error()))
		return ConnectResult{Error: formatConnectionError(err)}
	}

	a.logEvent("INFO", fmt.Sprintf("Connected to %s database", config.Type))
	id := uuid.New().String()

	a.mu.Lock()
	a.dbs[id] = newDB
	a.mu.Unlock()

	return ConnectResult{ID: id}
}

func (a *App) TestConnection(config DBConfig) string {
	if config.ID != "" {
		if config.Password == "" {
			if pw, err := keyring.Get("QuraMate", config.ID); err == nil {
				config.Password = pw
			}
		}
		if config.SSHEnabled && config.SSHPassword == "" {
			if sshPw, err := keyring.Get("QuraMate-SSH", config.ID); err == nil {
				config.SSHPassword = sshPw
			}
		}
	}

	newDB := database.NewDatabase(a.logEvent)
	err := newDB.Connect(config)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Connection test failed for %s: %s", config.Type, err.Error()))
		return formatConnectionError(err)
	}
	newDB.Disconnect()
	a.logEvent("INFO", fmt.Sprintf("Connection test successful for %s", config.Type))
	return "Success"
}

func fetchSSHHostKey(host string, port int) (ssh.PublicKey, error) {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return nil, fmt.Errorf("host is required")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}

	var capturedKey ssh.PublicKey
	sshConfig := &ssh.ClientConfig{
		User: "quramate-probe",
		Auth: []ssh.AuthMethod{
			ssh.Password("quramate-probe"),
		},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return fmt.Errorf("host key captured")
		},
		Timeout: sshDialTimeout,
	}

	addr := net.JoinHostPort(trimmedHost, strconv.Itoa(port))
	_, _ = ssh.Dial("tcp", addr, sshConfig)
	if capturedKey == nil {
		return nil, fmt.Errorf("unable to fetch host key from %s", addr)
	}

	return capturedKey, nil
}

func ensureKnownHostsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return "", fmt.Errorf("unable to locate user home directory")
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", fmt.Errorf("unable to create .ssh directory: %w", err)
	}

	knownHostsPath := filepath.Join(sshDir, "known_hosts")
	if _, err := os.Stat(knownHostsPath); errors.Is(err, os.ErrNotExist) {
		file, createErr := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_WRONLY, 0600)
		if createErr != nil {
			return "", fmt.Errorf("unable to create known_hosts: %w", createErr)
		}
		_ = file.Close()
	}

	return knownHostsPath, nil
}

func knownHostsPattern(host string, port int) string {
	trimmedHost := strings.TrimSpace(host)
	if port == 22 {
		return trimmedHost
	}
	return net.JoinHostPort(trimmedHost, strconv.Itoa(port))
}

func (a *App) GetSSHHostKeyInfo(host string, port int) SSHHostKeyInfo {
	key, err := fetchSSHHostKey(host, port)
	if err != nil {
		return SSHHostKeyInfo{Error: fmt.Sprintf("Error: %s", err.Error())}
	}

	trimmedHost := strings.TrimSpace(host)
	return SSHHostKeyInfo{
		Host:        trimmedHost,
		Port:        port,
		Pattern:     knownHostsPattern(trimmedHost, port),
		KeyType:     key.Type(),
		Fingerprint: ssh.FingerprintSHA256(key),
	}
}

func (a *App) TrustSSHHost(host string, port int) string {
	key, err := fetchSSHHostKey(host, port)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	knownHostsPath, err := ensureKnownHostsFilePath()
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	pattern := knownHostsPattern(host, port)
	newLine := knownhosts.Line([]string{pattern}, key)

	existing, err := os.ReadFile(knownHostsPath)
	if err == nil {
		existingText := string(existing)
		if strings.Contains(existingText, newLine) {
			a.logEvent("INFO", fmt.Sprintf("SSH host already trusted: %s", pattern))
			return "Success"
		}
	}

	file, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Sprintf("Error: unable to open known_hosts: %s", err.Error())
	}
	defer file.Close()

	if _, err := file.WriteString(newLine + "\n"); err != nil {
		return fmt.Sprintf("Error: unable to write known_hosts: %s", err.Error())
	}

	a.logEvent("INFO", fmt.Sprintf("Trusted SSH host key for %s (%s)", pattern, ssh.FingerprintSHA256(key)))
	return "Success"
}

func (a *App) DisconnectDB(connectionID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if db, ok := a.dbs[connectionID]; ok {
		err := db.Disconnect()
		delete(a.dbs, connectionID)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error())
		}
		return "Success"
	}
	return "Connection not found"
}

func (a *App) SaveCredential(id string, password string, sshPassword string) error {
	if id == "" {
		return fmt.Errorf("connection ID is required to save credentials")
	}
	var errs []string
	if password != "" {
		if err := keyring.Set("QuraMate", id, password); err != nil {
			errs = append(errs, fmt.Sprintf("failed to save db password: %v", err))
		}
	}
	if sshPassword != "" {
		if err := keyring.Set("QuraMate-SSH", id, sshPassword); err != nil {
			errs = append(errs, fmt.Sprintf("failed to save ssh password: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, ", "))
	}
	return nil
}

func (a *App) DeleteCredential(id string) error {
	if id == "" {
		return nil
	}
	_ = keyring.Delete("QuraMate", id)
	_ = keyring.Delete("QuraMate-SSH", id)
	return nil
}

func (a *App) SetReadOnly(connectionID string, readOnly bool) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if db, ok := a.dbs[connectionID]; ok {
		db.SetReadOnly(readOnly)
		return "Success"
	}
	return "Connection not found"
}

func (a *App) GetTables(connectionID string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	tables, err := db.GetTables()
	if err != nil {
		return []string{}
	}
	return tables
}

func (a *App) GetViews(connectionID string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	views, err := db.GetViews()
	if err != nil {
		return []string{}
	}
	return views
}

func (a *App) GetStoredProcedures(connectionID string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	header, err := db.GetStoredProcedures()
	if err != nil {
		return []string{}
	}
	return header
}

func (a *App) GetFunctions(connectionID string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	funcs, err := db.GetFunctions()
	if err != nil {
		return []string{}
	}
	return funcs
}

func (a *App) GetRoutineDefinition(connectionID string, name string, routineType string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Error: Connection not found"
	}

	def, err := db.GetRoutineDefinition(name, routineType)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return def
}

// Result struct to return both data and error message if any

type QueryResult struct {
	ResultSets []ResultSet `json:"resultSets"` // Changed from Data/Columns
	Error      string      `json:"error"`
}

type BatchInsertResult struct {
	Inserted int      `json:"inserted"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
	Error    string   `json:"error"`
}

func (a *App) ExecuteQuery(connectionID string, query string, queryID string) QueryResult {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return QueryResult{Error: "Connection not found"}
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Store the cancel function
	a.muQueries.Lock()
	a.queryCancelFuncs[queryID] = cancel
	a.muQueries.Unlock()

	// Ensure cleanup
	defer func() {
		a.muQueries.Lock()
		delete(a.queryCancelFuncs, queryID)
		a.muQueries.Unlock()
		cancel()
	}()

	resultSets, err := db.ExecuteQuery(ctx, query)
	if err != nil {
		if err == context.Canceled {
			return QueryResult{Error: "Query cancelled by user"}
		}
		a.logEvent("ERROR", fmt.Sprintf("Query execution failed: %s", err.Error()))
		return QueryResult{Error: err.Error()}
	}
	return QueryResult{ResultSets: resultSets}
}

func (a *App) ExecuteTransientQuery(connectionID string, query string) QueryResult {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return QueryResult{Error: "Connection not found"}
	}

	// Create a context (no cancel func stored for transient queries for now, they are short lived usually)
	ctx := context.Background()

	resultSets, err := db.ExecuteTransientQuery(ctx, query)
	if err != nil {
		return QueryResult{Error: err.Error()}
	}
	return QueryResult{ResultSets: resultSets}
}

func (a *App) InsertRowsBatch(connectionID string, tableName string, rows []map[string]interface{}) BatchInsertResult {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return BatchInsertResult{Error: "Connection not found"}
	}

	if strings.TrimSpace(tableName) == "" {
		return BatchInsertResult{Error: "Table name is required"}
	}

	if len(rows) == 0 {
		return BatchInsertResult{Error: "No rows to insert"}
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return BatchInsertResult{Error: fmt.Sprintf("Error starting transaction: %s", err.Error())}
	}
	defer tx.Rollback()

	inserted := 0
	skipped := 0
	rowErrors := make([]string, 0)

	for idx, row := range rows {
		if len(row) == 0 {
			skipped++
			rowErrors = append(rowErrors, fmt.Sprintf("Row %d skipped: no values to insert", idx+1))
			continue
		}

		if err := db.InsertRecordTx(tx, tableName, row); err != nil {
			rowErrors = append(rowErrors, fmt.Sprintf("Row %d failed: %s", idx+1, err.Error()))
			return BatchInsertResult{
				Inserted: 0,
				Skipped:  skipped,
				Errors:   rowErrors,
				Error:    fmt.Sprintf("Batch insert failed and rolled back: %s", err.Error()),
			}
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return BatchInsertResult{
			Inserted: 0,
			Skipped:  skipped,
			Errors:   rowErrors,
			Error:    fmt.Sprintf("Error committing transaction: %s", err.Error()),
		}
	}

	return BatchInsertResult{
		Inserted: inserted,
		Skipped:  skipped,
		Errors:   rowErrors,
	}
}

func (a *App) CancelQuery(queryID string) string {
	a.muQueries.Lock()
	cancel, ok := a.queryCancelFuncs[queryID]
	a.muQueries.Unlock()

	if ok {
		cancel()
		return "Success"
	}
	return "Query not found or already completed"
}

func (a *App) GetServerProcesses(connectionID string) ([]ServerProcess, error) {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("Connection not found")
	}
	return db.GetServerProcesses(context.Background())
}

func (a *App) KillServerProcess(connectionID string, sessionID string) error {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return fmt.Errorf("Connection not found")
	}
	return db.KillServerProcess(context.Background(), sessionID)
}

// ExecuteQueryStream starts query execution in a goroutine and streams results
// via Wails events. Returns immediately with "" (no error) or an error string.

func (a *App) ExecuteQueryStream(connectionID string, query string, queryID string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	ctx, cancel := context.WithCancel(context.Background())

	a.muQueries.Lock()
	a.queryCancelFuncs[queryID] = cancel
	a.muQueries.Unlock()

	go func() {
		defer func() {
			a.muQueries.Lock()
			delete(a.queryCancelFuncs, queryID)
			a.muQueries.Unlock()
			cancel()
		}()
		startTime := time.Now()
		rowCount := 0
		executionReported := false
		err := db.ExecuteQueryStream(ctx, query, 500, func(batch StreamBatch) {
			if !executionReported {
				executionReported = true
				runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
					"phase": "execution",
					"time":  time.Since(startTime).Milliseconds(),
				})
			}

			rowCount += len(batch.Rows)
			runtime.EventsEmit(a.ctx, "query:batch:"+queryID, batch)
			runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
				"rows":      rowCount,
				"time":      time.Since(startTime).Milliseconds(),
				"fetchTime": time.Since(startTime).Milliseconds(),
				"partial":   false,
				"phase":     "fetch",
			})
		})
		if err != nil {
			errMsg := err.Error()
			if err == context.Canceled {
				errMsg = "Query cancelled by user"
			}
			runtime.EventsEmit(a.ctx, "query:error:"+queryID, errMsg)
			return
		}
		runtime.EventsEmit(a.ctx, "query:done:"+queryID)
	}()

	return ""
}

func (a *App) GetPrimaryKeys(connectionID string, tableName string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	pks, err := db.GetPrimaryKeys(tableName)
	if err != nil {
		return []string{}
	}
	return pks
}

func (a *App) UpdateRecord(connectionID string, tableName string, updates map[string]interface{}, conditions map[string]interface{}) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	err := db.UpdateRecord(tableName, updates, conditions)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) GetForeignKeys(connectionID string, tableName string) []ForeignKey {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []ForeignKey{}
	}

	fks, err := db.GetForeignKeys(tableName)
	if err != nil {
		return []ForeignKey{}
	}
	return fks
}

func (a *App) GetDatabaseInfo(connectionID string) (DatabaseInfo, error) {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return DatabaseInfo{}, fmt.Errorf("no database connection")
	}

	return db.GetDatabaseInfo()
}

func (a *App) DropDatabase(connectionID string, dbName string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "no database connection"
	}

	err := db.DropDatabase(dbName)
	if err != nil {
		return err.Error()
	}
	return ""
}

// ExportTable exports a table to a file

func (a *App) ExplainQuery(connectionID string, query string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	plan, err := db.ExplainQuery(context.Background(), query)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return plan
}

func (a *App) GetTableDefinition(connectionID string, tableName string) []ColumnDefinition {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []ColumnDefinition{}
	}

	cols, err := db.GetTableDefinition(tableName)
	if err != nil {
		return []ColumnDefinition{}
	}
	return cols
}

func (a *App) GetTableIndexes(connectionID string, tableName string) []IndexDefinition {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []IndexDefinition{}
	}

	indexes, err := db.GetTableIndexes(tableName)
	if err != nil {
		return []IndexDefinition{}
	}
	return indexes
}

func (a *App) AlterTable(connectionID string, tableName string, changes TableChanges) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	err := db.AlterTable(tableName, changes)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) CreateTable(connectionID string, tableName string, columns []ColumnDefinition) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	err := db.CreateTable(tableName, columns)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

// --- Query History Methods ---
