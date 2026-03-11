package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// App struct
type App struct {
	ctx              context.Context
	dbs              map[string]*Database
	mu               sync.Mutex
	queryCancelFuncs map[string]context.CancelFunc
	muQueries        sync.Mutex
	appLogs          []LogEntry
	muLogs           sync.Mutex
	localDB          *LocalDB
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

const aiKeyringService = "QuraMate-AI"

func NewApp() *App {
	ldb, err := NewLocalDB()
	if err != nil {
		fmt.Printf("Failed to initialize LocalDB: %v\n", err)
	}

	return &App{
		dbs:              make(map[string]*Database),
		queryCancelFuncs: make(map[string]context.CancelFunc),
		localDB:          ldb,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logEvent("INFO", "Application started")

	// Auto-check for updates after a short delay
	go func() {
		time.Sleep(3 * time.Second)
		info := a.CheckForUpdates()
		if info.Available {
			runtime.EventsEmit(a.ctx, "app:update-available", info)
		}
	}()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ConnectResult struct to return both ID and success status
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

	code := classifyConnectionErrorCode(err)
	if code != "" {
		return fmt.Sprintf("Error [%s]: %s", code, err.Error())
	}

	return fmt.Sprintf("Error: %s", err.Error())
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

	newDB := NewDatabase(a.logEvent)
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

	newDB := NewDatabase(a.logEvent)
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

func (a *App) logEvent(level string, msg string) {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()

	entry := LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   level,
		Message: msg,
	}

	a.appLogs = append(a.appLogs, entry)

	// Keep only last 1000 logs
	if len(a.appLogs) > 1000 {
		a.appLogs = a.appLogs[len(a.appLogs)-1000:]
	}
}

func (a *App) GetAppLogs() []LogEntry {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()
	return a.appLogs
}

func (a *App) ClearAppLogs() string {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()
	a.appLogs = []LogEntry{}
	return "Success"
}

func (a *App) LogClientEvent(level string, message string) string {
	cleanLevel := strings.ToUpper(strings.TrimSpace(level))
	if cleanLevel == "" {
		cleanLevel = "INFO"
	}

	cleanMessage := strings.TrimSpace(message)
	if cleanMessage == "" {
		return "Message is required"
	}
	if len(cleanMessage) > 2000 {
		cleanMessage = cleanMessage[:2000]
	}

	a.logEvent(cleanLevel, cleanMessage)
	return "Success"
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

		const SafeBufferLimit = 10000
		startTime := time.Now()

		if db.persistentConn == nil {
			runtime.EventsEmit(a.ctx, "query:error:"+queryID, "No database connection")
			return
		}

		rows, err := db.persistentConn.QueryContext(ctx, query)
		if err != nil {
			errMsg := err.Error()
			if err == context.Canceled {
				errMsg = "Query cancelled by user"
			}
			runtime.EventsEmit(a.ctx, "query:error:"+queryID, errMsg)
			return
		}

		// Execution Time: Time to get the rows object (query executed)
		executionDuration := time.Since(startTime).Milliseconds()
		runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
			"phase": "execution",
			"time":  executionDuration,
		})

		defer rows.Close()

		resultSetIdx := 0

		for {
			if err := ctx.Err(); err != nil {
				return
			}

			columns, _ := rows.Columns()
			colTypes, _ := rows.ColumnTypes()
			var columnMetas []ColumnMetadata
			if len(colTypes) > 0 {
				columnMetas = make([]ColumnMetadata, len(colTypes))
				for i, ct := range colTypes {
					length, _ := ct.Length()
					nullable, _ := ct.Nullable()
					columnMetas[i] = ColumnMetadata{
						Name:     ct.Name(),
						Type:     ct.DatabaseTypeName(),
						Length:   length,
						Nullable: nullable,
					}
				}
			}

			if len(columns) > 0 {
				nCols := len(columns)
				scanTargets := make([]interface{}, nCols)
				scanPtrs := make([]*interface{}, nCols)
				for i := 0; i < nCols; i++ {
					scanPtrs[i] = new(interface{})
					scanTargets[i] = scanPtrs[i]
				}

				buffer := make([][]interface{}, 0, SafeBufferLimit)
				isStreaming := false
				batchSize := 500
				rowCount := 0

				for rows.Next() {
					rowCount++
					if err := ctx.Err(); err != nil {
						return
					}

					if err := rows.Scan(scanTargets...); err != nil {
						runtime.EventsEmit(a.ctx, "query:error:"+queryID, err.Error())
						return
					}

					row := make([]interface{}, nCols)
					for i := 0; i < nCols; i++ {
						val := *scanPtrs[i]
						if b, ok := val.([]byte); ok {
							row[i] = decodeValue(b, db.Encoding)
						} else if s, ok := val.(string); ok {
							row[i] = decodeStringValue(s, db.Encoding)
						} else {
							row[i] = val
						}
					}

					if !isStreaming {
						buffer = append(buffer, row)
						if len(buffer) >= SafeBufferLimit {
							isStreaming = true

							// Fetch time so far
							fetchDuration := time.Since(startTime).Milliseconds() - executionDuration
							runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
								"rows":      SafeBufferLimit,
								"time":      executionDuration, // Keep original exec time
								"fetchTime": fetchDuration,
								"partial":   true,
								"phase":     "fetch",
							})

							for i := 0; i < len(buffer); i += batchSize {
								end := i + batchSize
								if end > len(buffer) {
									end = len(buffer)
								}
								runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
									Columns:      columns,
									ColumnTypes:  columnMetas,
									Rows:         buffer[i:end],
									ResultSetIdx: resultSetIdx,
									BatchIndex:   i / batchSize,
								})
							}
							buffer = nil
						}
					} else {
						if buffer == nil {
							buffer = make([][]interface{}, 0, batchSize)
						}
						buffer = append(buffer, row)

						if len(buffer) >= batchSize {
							runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
								Columns:      columns,
								ColumnTypes:  columnMetas,
								Rows:         buffer,
								ResultSetIdx: resultSetIdx,
								BatchIndex:   -1,
							})
							buffer = nil
						}
					}
				}

				// Loop finished
				totalDuration := time.Since(startTime).Milliseconds()
				fetchDuration := totalDuration - executionDuration

				if !isStreaming {
					// Final stats for this result set
					runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
						"rows":      rowCount,
						"time":      executionDuration,
						"fetchTime": time.Since(startTime).Milliseconds() - executionDuration,
						"partial":   false,
						"phase":     "fetch",
					})

					// Even if buffer is empty, we must send at least one batch if we have columns
					// to let the frontend know this is a data result set.
					if len(buffer) == 0 {
						runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
							Columns:      columns,
							ColumnTypes:  columnMetas,
							Rows:         [][]interface{}{},
							ResultSetIdx: resultSetIdx,
							BatchIndex:   0,
						})
					} else {
						for i := 0; i < len(buffer); i += batchSize {
							end := i + batchSize
							if end > len(buffer) {
								end = len(buffer)
							}
							runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
								Columns:      columns,
								ColumnTypes:  columnMetas,
								Rows:         buffer[i:end],
								ResultSetIdx: resultSetIdx,
								BatchIndex:   i / batchSize,
							})
						}
					}
				} else {
					if len(buffer) > 0 {
						runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
							Columns:      columns,
							ColumnTypes:  columnMetas,
							Rows:         buffer,
							ResultSetIdx: resultSetIdx,
							BatchIndex:   -1,
						})
					}
					runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
						"rows":      rowCount,
						"time":      executionDuration,
						"fetchTime": fetchDuration,
						"partial":   false,
						"phase":     "fetch",
					})
				}
			} else {
				duration := time.Since(startTime).Milliseconds()
				runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
					"rows":    0,
					"time":    duration,
					"partial": false,
				})
				runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
					Columns:      nil,
					ColumnTypes:  nil,
					Rows:         nil,
					ResultSetIdx: resultSetIdx,
					BatchIndex:   0,
				})
			}

			resultSetIdx++
			if !rows.NextResultSet() {
				break
			}
		}

		if err = rows.Err(); err != nil {
			runtime.EventsEmit(a.ctx, "query:error:"+queryID, err.Error())
		} else {
			runtime.EventsEmit(a.ctx, "query:done:"+queryID)
		}
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

// ForeignKey struct to hold FK details
type ForeignKey struct {
	Table      string `json:"table"`
	Column     string `json:"column"`
	RefTable   string `json:"refTable"`
	RefColumn  string `json:"refColumn"`
	Constraint string `json:"constraint"`
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
func (a *App) ExportTable(connectionID string, tableName string, format string, filePath string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	// Fetch all data
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	resultSets, err := db.ExecuteQuery(context.Background(), query)
	if err != nil {
		return fmt.Sprintf("Error fetching data: %s", err.Error())
	}
	if len(resultSets) == 0 {
		return "Error: No data returned"
	}
	data := resultSets[0].Rows
	columns := resultSets[0].Columns

	var exportErr error
	switch strings.ToLower(format) {
	case "json":
		exportErr = a.exportToJSON(data, columns, filePath)
	case "csv":
		exportErr = a.exportToCSV(data, columns, filePath)
	case "sql":
		exportErr = a.exportToSQL(tableName, data, columns, filePath)
	case "excel":
		exportErr = a.exportToExcel(data, columns, filePath)
	default:
		return fmt.Sprintf("Unsupported format: %s", format)
	}

	if exportErr != nil {
		return fmt.Sprintf("Error exporting: %s", exportErr.Error())
	}

	return "Success"
}

// ExportDatabase exports all tables in a database to a folder
func (a *App) ExportDatabase(connectionID string, format string, folderPath string) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	tables, err := db.GetTables()
	if err != nil {
		return fmt.Sprintf("Error fetching tables: %s", err.Error())
	}

	// Create folder if it doesn't exist
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Sprintf("Error creating folder: %s", err.Error())
	}

	for _, table := range tables {
		fileName := fmt.Sprintf("%s.%s", table, strings.ToLower(format))
		// For SQL, we might want to append to a single file instead?
		// But for now, separate files is easier.
		filePath := filepath.Join(folderPath, fileName)
		res := a.ExportTable(connectionID, table, format, filePath)
		if res != "Success" {
			return fmt.Sprintf("Error exporting table %s: %s", table, res)
		}
	}

	return "Success"
}

func (a *App) exportToJSON(data [][]interface{}, columns []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert to map format for JSON output (user-facing file)
	mapData := make([]map[string]interface{}, len(data))
	for i, row := range data {
		m := make(map[string]interface{}, len(columns))
		for j, col := range columns {
			if j < len(row) {
				m[col] = row[j]
			}
		}
		mapData[i] = m
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(mapData)
}

func (a *App) exportToCSV(data [][]interface{}, columns []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(columns); err != nil {
		return err
	}

	// Write data
	for _, row := range data {
		record := make([]string, len(columns))
		for i := range columns {
			if i < len(row) && row[i] != nil {
				record[i] = fmt.Sprintf("%v", row[i])
			}
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) exportToSQL(tableName string, data [][]interface{}, columns []string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, row := range data {
		var cols []string
		var vals []string
		for i, col := range columns {
			var val interface{}
			if i < len(row) {
				val = row[i]
			}
			if val != nil {
				cols = append(cols, col)
				strVal := fmt.Sprintf("%v", val)
				// Escape single quotes
				strVal = strings.ReplaceAll(strVal, "'", "''")
				vals = append(vals, fmt.Sprintf("'%s'", strVal))
			}
		}
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n", tableName, strings.Join(cols, ", "), strings.Join(vals, ", "))
		if _, err := file.WriteString(query); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) exportToExcel(data [][]interface{}, columns []string, filePath string) error {
	f := excelize.NewFile()
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// Write headers
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col)
	}

	// Write data
	for i, row := range data {
		for j := range columns {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			if j < len(row) {
				f.SetCellValue(sheetName, cell, row[j])
			}
		}
	}

	return f.SaveAs(filePath)
}

// ==== Startup Arguments ====

func (a *App) debugLog(msg string) {
	execPath, err := os.Executable()
	logDir := "."
	if err == nil {
		logDir = filepath.Dir(execPath)
	}
	logFile := filepath.Join(logDir, "debug_open.log")
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}

func (a *App) GetStartupFile() string {
	a.debugLog(fmt.Sprintf("GetStartupFile called, os.Args=%v", os.Args))
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if _, err := os.Stat(arg); err == nil {
			a.debugLog(fmt.Sprintf("GetStartupFile returning: %s", arg))
			return arg
		}
		a.debugLog(fmt.Sprintf("GetStartupFile arg not a file: %s", arg))
	}
	a.debugLog("GetStartupFile returning empty")
	return ""
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	a.debugLog(fmt.Sprintf("onSecondInstanceLaunch called with args: %v", secondInstanceData.Args))

	// Bring window to front
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	a.debugLog("Window brought to front")

	if len(secondInstanceData.Args) > 0 {
		for _, arg := range secondInstanceData.Args {
			a.debugLog(fmt.Sprintf("Checking arg: %s", arg))
			if _, err := os.Stat(arg); err == nil {
				// Write to pending file for the running instance to pick up
				pendingPath := a.getPendingFilePath()
				writeErr := os.WriteFile(pendingPath, []byte(arg), 0644)
				a.debugLog(fmt.Sprintf("Wrote pending file: %s -> %s (err=%v)", arg, pendingPath, writeErr))
				// Also emit event so frontend can react immediately (no polling needed)
				runtime.EventsEmit(a.ctx, "app:open-file", arg)
				break
			}
		}
	} else {
		a.debugLog("No file arg in second instance args")
	}
}

func (a *App) getPendingFilePath() string {
	execPath, err := os.Executable()
	if err != nil {
		return "pending_open.txt"
	}
	appDir := filepath.Dir(execPath)
	if strings.Contains(appDir, "Temp") || strings.Contains(appDir, "tmp") {
		appDir, _ = os.Getwd()
	}
	return filepath.Join(appDir, "pending_open.txt")
}

func (a *App) CheckPendingFile() string {
	pendingPath := a.getPendingFilePath()
	content, err := os.ReadFile(pendingPath)
	if err != nil {
		return ""
	}
	// Delete immediately after reading
	os.Remove(pendingPath)
	filePath := strings.TrimSpace(string(content))
	if filePath == "" {
		a.debugLog("CheckPendingFile: file was empty")
		return ""
	}
	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		a.debugLog(fmt.Sprintf("CheckPendingFile: file not found: %s", filePath))
		return ""
	}
	a.debugLog(fmt.Sprintf("CheckPendingFile: returning %s", filePath))
	return filePath
}

func (a *App) ReadTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Detect UTF-16 LE BOM (FF FE) or BE BOM (FE FF)
	if len(content) >= 2 {
		if content[0] == 0xFF && content[1] == 0xFE {
			// UTF-16 LE
			return decodeUTF16(content[2:], false), nil
		} else if content[0] == 0xFE && content[1] == 0xFF {
			// UTF-16 BE
			return decodeUTF16(content[2:], true), nil
		}
	}

	// Also check if there's an overwhelming amount of NULL bytes (UTF-16 without BOM)
	// Just sample the first few bytes. Wait, safer to just return as string for UTF-8 normally.
	// We'll strip UTF-8 BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return string(content[3:]), nil
	}

	return string(content), nil
}

func (a *App) WriteTextFile(path string, content string) string {
	if strings.TrimSpace(path) == "" {
		return "file path is required"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err.Error()
	}

	return ""
}

func decodeUTF16(b []byte, isBE bool) string {
	if len(b)%2 != 0 {
		// Just append a null byte to make it even
		b = append(b, 0)
	}
	u16s := make([]uint16, len(b)/2)
	for i := 0; i < len(u16s); i++ {
		if isBE {
			u16s[i] = uint16(b[i*2])<<8 | uint16(b[i*2+1])
		} else {
			u16s[i] = uint16(b[i*2+1])<<8 | uint16(b[i*2])
		}
	}
	// Convert array of uint16 to string by casting it to a slice of runes
	runes := make([]rune, len(u16s))
	for i, v := range u16s {
		runes[i] = rune(v)
	}
	return string(runes)
}

// ==== Settings Wails Bindings ====

func normalizeAIProvider(provider string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return "", fmt.Errorf("provider is required")
	}

	for _, ch := range normalized {
		isAlphaNum := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if !isAlphaNum && ch != '-' && ch != '_' {
			return "", fmt.Errorf("provider contains invalid characters")
		}
	}

	return normalized, nil
}

func sanitizeUserSettingsForStorage(value string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		// Keep original payload if it isn't JSON (backward compatibility).
		return value
	}

	if aiRaw, ok := parsed["ai"]; ok {
		if aiMap, ok := aiRaw.(map[string]any); ok {
			delete(aiMap, "apiKey")
			delete(aiMap, "apiKeys")
		}
	}

	sanitized, err := json.Marshal(parsed)
	if err != nil {
		return value
	}

	return string(sanitized)
}

func (a *App) SaveSetting(key string, value string) string {
	if a.localDB == nil {
		return "Error: LocalDB is not initialized"
	}
	if key == "user_settings" {
		value = sanitizeUserSettingsForStorage(value)
	}
	err := a.localDB.SaveSetting(key, value)
	if err != nil {
		return fmt.Sprintf("Error saving setting: %s", err.Error())
	}
	return "Success"
}

func (a *App) LoadSetting(key string) string {
	if a.localDB == nil {
		return ""
	}
	value, err := a.localDB.LoadSetting(key)
	if err != nil {
		fmt.Printf("Error loading setting %s: %v\n", key, err)
		return ""
	}
	return value
}

func (a *App) SaveAIProviderKey(provider string, apiKey string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" {
		if err := keyring.Delete(aiKeyringService, normalizedProvider); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return fmt.Sprintf("Error deleting key: %s", err.Error())
		}
		return "Success"
	}

	if err := keyring.Set(aiKeyringService, normalizedProvider, trimmedKey); err != nil {
		return fmt.Sprintf("Error saving key: %s", err.Error())
	}
	return "Success"
}

func (a *App) LoadAIProviderKey(provider string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return ""
	}

	value, err := keyring.Get(aiKeyringService, normalizedProvider)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ""
		}
		fmt.Printf("Error loading AI provider key for %s: %v\n", normalizedProvider, err)
		return ""
	}
	return value
}

func (a *App) DeleteAIProviderKey(provider string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	if err := keyring.Delete(aiKeyringService, normalizedProvider); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Sprintf("Error deleting key: %s", err.Error())
	}

	return "Success"
}

// ImportTable imports data from a file to a table
func (a *App) ImportTable(connectionID string, tableName string, format string, filePath string, enableIdentityInsert bool) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Sprintf("Error starting transaction: %s", err.Error())
	}
	defer tx.Rollback() // Rollback if not committed

	if enableIdentityInsert && db.Type == "mssql" {
		_, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", tableName))
		if err != nil {
			return fmt.Sprintf("Error setting IDENTITY_INSERT ON: %s", err.Error())
		}
	}

	var importErr error
	switch strings.ToLower(format) {
	case "json":
		importErr = a.importFromJSON(db, tx, tableName, filePath)
	case "csv":
		importErr = a.importFromCSV(db, tx, tableName, filePath)
	case "sql":
		importErr = a.importFromSQL(tx, filePath)
	case "excel":
		importErr = a.importFromExcel(db, tx, tableName, filePath)
	default:
		importErr = fmt.Errorf("unsupported format: %s", format)
	}

	if importErr != nil {
		return fmt.Sprintf("Error importing: %s", importErr.Error())
	}

	if enableIdentityInsert && db.Type == "mssql" {
		_, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s OFF", tableName))
		if err != nil {
			return fmt.Sprintf("Error setting IDENTITY_INSERT OFF: %s", err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Sprintf("Error committing transaction: %s", err.Error())
	}

	return "Success"
}

func (a *App) importFromJSON(db *Database, tx *sql.Tx, tableName string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var data []map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	for _, row := range data {
		if err := db.InsertRecordTx(tx, tableName, row); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) importFromCSV(db *Database, tx *sql.Tx, tableName string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		row := make(map[string]interface{})
		for i, val := range record {
			if i < len(headers) {
				row[headers[i]] = val
			}
		}
		if err := db.InsertRecordTx(tx, tableName, row); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) importFromSQL(tx *sql.Tx, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	query := string(content)
	statements := strings.Split(query, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) importFromExcel(db *Database, tx *sql.Tx, tableName string, filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Assume data is in the first sheet
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return fmt.Errorf("excel file must have a header row and at least one data row")
	}

	headers := rows[0]
	for i := 1; i < len(rows); i++ {
		rowVals := rows[i]
		row := make(map[string]interface{})
		for j, val := range rowVals {
			if j < len(headers) {
				row[headers[j]] = val
			}
		}
		// Fill missing columns with nil or empty if row is shorter than headers

		if err := db.InsertRecordTx(tx, tableName, row); err != nil {
			return err
		}
	}
	return nil
}

// SelectExportFile opens a save file dialog and returns the selected path
func (a *App) SelectExportFile(defaultFilename string) string {
	selection, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Data",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "All Supported Files", Pattern: "*.json;*.csv;*.sql;*.xlsx"},
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "SQL Files (*.sql)", Pattern: "*.sql"},
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		return ""
	}
	return selection
}

// SelectImportFile opens an open file dialog and returns the selected path
func (a *App) SelectImportFile() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import Data",
		Filters: []runtime.FileFilter{
			{DisplayName: "All Supported Files", Pattern: "*.json;*.csv;*.sql;*.xlsx"},
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "SQL Files (*.sql)", Pattern: "*.sql"},
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		return ""
	}
	return selection
}

// SelectSqliteFile opens an open file dialog to select a local database file (SQLite or DuckDB)
func (a *App) SelectSqliteFile() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Local Database File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Local Databases", Pattern: "*.sqlite;*.db;*.duckdb;*.ddb"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// SelectFolder opens a directory dialog and returns the selected path
func (a *App) SelectFolder() string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder",
	})
	if err != nil {
		return ""
	}
	return selection
}

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

func (a *App) SaveQueryHistory(query string, dbType string) string {
	if a.localDB == nil {
		return "LocalDB not initialized"
	}
	err := a.localDB.SaveQuery(query, dbType)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) GetQueryHistory(dbType string) []QueryHistoryEntry {
	if a.localDB == nil {
		return []QueryHistoryEntry{}
	}
	entries, err := a.localDB.GetQueries(dbType)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to get query history: %v", err))
		return []QueryHistoryEntry{}
	}
	return entries
}

func (a *App) SearchQueryHistory(queryText string, dbType string, favoritesOnly bool, dateRange string, sortMode string, limit int) []QueryHistoryEntry {
	if a.localDB == nil {
		return []QueryHistoryEntry{}
	}
	entries, err := a.localDB.SearchQueries(queryText, dbType, favoritesOnly, dateRange, sortMode, limit)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to search query history: %v", err))
		return []QueryHistoryEntry{}
	}
	return entries
}

func (a *App) ToggleFavoriteQuery(id int, isFavorite bool) string {
	if a.localDB == nil {
		return "LocalDB not initialized"
	}
	err := a.localDB.ToggleFavorite(id, isFavorite)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) DeleteQueryHistory(id int) string {
	if a.localDB == nil {
		return "LocalDB not initialized"
	}
	err := a.localDB.DeleteQuery(id)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) ClearQueryHistory() string {
	if a.localDB == nil {
		return "LocalDB not initialized"
	}
	_, err := a.localDB.conn.Exec(`DELETE FROM query_history WHERE is_favorite = 0`)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}
