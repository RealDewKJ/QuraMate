package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// App struct
type App struct {
	ctx              context.Context
	dbs              map[string]*Database
	mu               sync.Mutex
	queryCancelFuncs map[string]context.CancelFunc
	muQueries        sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		dbs:              make(map[string]*Database),
		queryCancelFuncs: make(map[string]context.CancelFunc),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) ConnectDB(config DBConfig) ConnectResult {
	newDB := NewDatabase()
	err := newDB.Connect(config)
	if err != nil {
		return ConnectResult{Error: fmt.Sprintf("Error: %s", err.Error())}
	}

	id := uuid.New().String()

	a.mu.Lock()
	a.dbs[id] = newDB
	a.mu.Unlock()

	return ConnectResult{ID: id}
}

func (a *App) TestConnection(config DBConfig) string {
	newDB := NewDatabase()
	err := newDB.Connect(config)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	newDB.Disconnect()
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

// Result struct to return both data and error message if any
type QueryResult struct {
	ResultSets []ResultSet `json:"resultSets"` // Changed from Data/Columns
	Error      string      `json:"error"`
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
		return QueryResult{Error: err.Error()}
	}
	return QueryResult{ResultSets: resultSets}
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

				for rows.Next() {
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
							row[i] = string(b)
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
					runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
						"rows":      len(buffer),
						"time":      executionDuration,
						"fetchTime": fetchDuration,
						"partial":   false,
						"phase":     "fetch",
					})

					for i := 0; i < len(buffer); i += batchSize {
						end := i + batchSize
						if end > len(buffer) {
							end = len(buffer)
						}
						runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
							Columns:      columns,
							Rows:         buffer[i:end],
							ResultSetIdx: resultSetIdx,
							BatchIndex:   i / batchSize,
						})
					}
				} else {
					if len(buffer) > 0 {
						runtime.EventsEmit(a.ctx, "query:batch:"+queryID, StreamBatch{
							Columns:      columns,
							Rows:         buffer,
							ResultSetIdx: resultSetIdx,
							BatchIndex:   -1,
						})
					}
					runtime.EventsEmit(a.ctx, "query:stats:"+queryID, map[string]interface{}{
						"rows":      -1,
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
		exportErr = a.exportToExcel(tableName, data, columns, filePath)
	default:
		return fmt.Sprintf("Unsupported format: %s", format)
	}

	if exportErr != nil {
		return fmt.Sprintf("Error exporting: %s", exportErr.Error())
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

func (a *App) exportToExcel(tableName string, data [][]interface{}, columns []string, filePath string) error {
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
			var val interface{}
			if j < len(row) {
				val = row[j]
			}
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	return f.SaveAs(filePath)
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
		importErr = a.importFromSQL(db, tx, filePath)
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

func (a *App) importFromSQL(db *Database, tx *sql.Tx, filePath string) error {
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
