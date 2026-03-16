package app

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

func (a *App) ExportTable(connectionID string, tableName string, format string, filePath string) string {
	result := a.ExportTableAdvanced(TableExportOptions{
		ConnectionID:  connectionID,
		TableName:     tableName,
		Format:        format,
		FilePath:      filePath,
		IncludeData:   true,
		IncludeSchema: false,
		DropIfExists:  false,
	})
	if !result.Success {
		return result.Error
	}
	return "Success"
}

func (a *App) ExportTableAdvanced(options TableExportOptions) ActionResult {
	a.mu.Lock()
	db, ok := a.dbs[options.ConnectionID]
	a.mu.Unlock()

	if !ok {
		return ActionResult{Success: false, Error: "Connection not found"}
	}

	format, err := normalizeImportExportFormat(options.Format)
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}
	if !options.IncludeData && !options.IncludeSchema {
		return ActionResult{Success: false, Error: "Select schema, data, or both before exporting"}
	}
	if strings.TrimSpace(options.FilePath) == "" {
		return ActionResult{Success: false, Error: "Export file path is required"}
	}

	safeTableName, err := resolveExistingTableName(db, options.TableName)
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}

	var (
		data           [][]interface{}
		columns        []string
		dropTableSQL   string
		schemaSQL      string
		fetchDataErr   error
		fetchSchemaErr error
	)

	if options.IncludeData {
		query := fmt.Sprintf("SELECT * FROM %s", db.QuoteIdentifier(safeTableName))
		resultSets, queryErr := db.ExecuteQuery(context.Background(), query)
		if queryErr != nil {
			fetchDataErr = fmt.Errorf("error fetching data: %w", queryErr)
		} else if len(resultSets) == 0 {
			fetchDataErr = fmt.Errorf("no data returned")
		} else {
			data = resultSets[0].Rows
			columns = resultSets[0].Columns
		}
	}

	if options.IncludeSchema {
		schemaSQL, fetchSchemaErr = db.GenerateCreateTableStatement(safeTableName)
		if fetchSchemaErr != nil {
			fetchSchemaErr = fmt.Errorf("error generating schema: %w", fetchSchemaErr)
		}
		if fetchSchemaErr == nil && options.DropIfExists && format == "sql" {
			dropTableSQL, fetchSchemaErr = db.GenerateDropTableIfExistsStatement(safeTableName)
			if fetchSchemaErr != nil {
				fetchSchemaErr = fmt.Errorf("error generating drop statement: %w", fetchSchemaErr)
			}
		}
	}

	if fetchDataErr != nil {
		return ActionResult{Success: false, Error: fetchDataErr.Error()}
	}
	if fetchSchemaErr != nil {
		return ActionResult{Success: false, Error: fetchSchemaErr.Error()}
	}

	var exportErr error
	switch format {
	case "json":
		if options.IncludeData {
			exportErr = a.exportToJSON(data, columns, options.FilePath)
		}
	case "csv":
		if options.IncludeData {
			exportErr = a.exportToCSV(data, columns, options.FilePath)
		}
	case "sql":
		exportErr = a.exportToSQL(db, safeTableName, data, columns, options.FilePath, dropTableSQL, schemaSQL, options.IncludeSchema, options.IncludeData)
	case "excel":
		if options.IncludeData {
			exportErr = a.exportToExcel(data, columns, options.FilePath)
		}
	default:
		exportErr = fmt.Errorf("unsupported format: %s", format)
	}

	if exportErr == nil && options.IncludeSchema && format != "sql" {
		exportErr = writeSchemaSidecar(options.FilePath, schemaSQL)
	}

	if exportErr != nil {
		return ActionResult{Success: false, Error: fmt.Sprintf("Error exporting: %s", exportErr.Error())}
	}

	return ActionResult{Success: true}
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

func (a *App) exportToSQL(db *Database, tableName string, data [][]interface{}, columns []string, filePath string, dropTableSQL string, schemaSQL string, includeSchema bool, includeData bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if includeSchema && schemaSQL != "" {
		if strings.TrimSpace(dropTableSQL) != "" {
			if _, err := file.WriteString(strings.TrimSpace(dropTableSQL) + ";\n"); err != nil {
				return err
			}
		}
		if _, err := file.WriteString(strings.TrimSpace(schemaSQL) + ";\n\n"); err != nil {
			return err
		}
	}

	if !includeData {
		return nil
	}

	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = db.QuoteIdentifier(col)
	}
	quotedTableName := db.QuoteIdentifier(tableName)

	for _, row := range data {
		var vals []string
		for i := range columns {
			var val interface{}
			if i < len(row) {
				val = row[i]
			}
			vals = append(vals, toSQLLiteral(val))
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s);\n",
			quotedTableName,
			strings.Join(quotedColumns, ", "),
			strings.Join(vals, ", "),
		)
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

func (a *App) ImportTable(connectionID string, tableName string, format string, filePath string, enableIdentityInsert bool) string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return "Connection not found"
	}

	safeTableName, err := resolveExistingTableName(db, tableName)
	if err != nil {
		return err.Error()
	}

	normalizedFormat, err := normalizeImportExportFormat(format)
	if err != nil {
		return err.Error()
	}

	tx, err := db.BeginTransaction()
	if err != nil {
		return fmt.Sprintf("Error starting transaction: %s", err.Error())
	}
	defer tx.Rollback() // Rollback if not committed

	identityInsertEnabled := false
	disableIdentityInsert := func() error {
		if !identityInsertEnabled {
			return nil
		}
		identityInsertEnabled = false
		if _, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s OFF", db.QuoteIdentifier(safeTableName))); err != nil {
			return fmt.Errorf("setting IDENTITY_INSERT OFF: %w", err)
		}
		return nil
	}

	if enableIdentityInsert && db.Type == "mssql" {
		_, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", db.QuoteIdentifier(safeTableName)))
		if err != nil {
			return fmt.Sprintf("Error setting IDENTITY_INSERT ON: %s", err.Error())
		}
		identityInsertEnabled = true
	}

	var importErr error
	switch normalizedFormat {
	case "json":
		importErr = a.importFromJSON(db, tx, safeTableName, filePath)
	case "csv":
		importErr = a.importFromCSV(db, tx, safeTableName, filePath)
	case "sql":
		importErr = a.importFromSQL(tx, safeTableName, filePath)
	case "excel":
		importErr = a.importFromExcel(db, tx, safeTableName, filePath)
	default:
		importErr = fmt.Errorf("unsupported format: %s", normalizedFormat)
	}

	if importErr != nil {
		if err := disableIdentityInsert(); err != nil {
			return fmt.Sprintf("Error importing: %s (cleanup failed: %s)", importErr.Error(), err.Error())
		}
		return fmt.Sprintf("Error importing: %s", importErr.Error())
	}

	if err := disableIdentityInsert(); err != nil {
		return fmt.Sprintf("Error setting IDENTITY_INSERT OFF: %s", err.Error())
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

func (a *App) importFromSQL(tx *sql.Tx, tableName string, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	statements := splitSQLStatements(string(content))
	targetName := canonicalObjectName(tableName)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := validateImportStatement(stmt, targetName); err != nil {
			return err
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

func normalizeImportExportFormat(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "json", nil
	case "csv":
		return "csv", nil
	case "sql":
		return "sql", nil
	case "excel", "xlsx":
		return "excel", nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func resolveExistingTableName(db *Database, tableName string) (string, error) {
	tables, err := db.GetTables()
	if err != nil {
		return "", fmt.Errorf("error validating table: %w", err)
	}

	target := canonicalObjectName(tableName)
	if target == "" {
		return "", fmt.Errorf("table name is required")
	}

	var match string
	for _, table := range tables {
		if canonicalObjectName(table) != target {
			continue
		}
		if match != "" && match != table {
			return "", fmt.Errorf("table name is ambiguous: %s", tableName)
		}
		match = table
	}

	if match == "" {
		return "", fmt.Errorf("table not found: %s", tableName)
	}

	return match, nil
}

func writeSchemaSidecar(filePath string, schemaSQL string) error {
	if strings.TrimSpace(schemaSQL) == "" {
		return fmt.Errorf("schema script is empty")
	}

	base := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	schemaPath := base + ".schema.sql"
	return os.WriteFile(schemaPath, []byte(strings.TrimSpace(schemaSQL)+";\n"), 0644)
}

func toSQLLiteral(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case json.Number:
		return v.String()
	case time.Time:
		return quoteSQLString(v.Format("2006-01-02 15:04:05.999999999"))
	case []byte:
		if utf8.Valid(v) {
			return quoteSQLString(string(v))
		}
		return quoteSQLString(hex.EncodeToString(v))
	default:
		return quoteSQLString(fmt.Sprintf("%v", value))
	}
}

func quoteSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func splitSQLStatements(query string) []string {
	var (
		statements          []string
		current             strings.Builder
		inSingleQuote       bool
		inDoubleQuote       bool
		inBacktickQuote     bool
		inBracketIdentifier bool
		inLineComment       bool
		inBlockComment      bool
	)

	for i := 0; i < len(query); i++ {
		ch := query[i]
		next := byte(0)
		if i+1 < len(query) {
			next = query[i+1]
		}

		if inLineComment {
			current.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			current.WriteByte(ch)
			if ch == '*' && next == '/' {
				current.WriteByte(next)
				i++
				inBlockComment = false
			}
			continue
		}
		if !inSingleQuote && !inDoubleQuote && !inBacktickQuote && !inBracketIdentifier {
			if ch == '-' && next == '-' {
				current.WriteByte(ch)
				current.WriteByte(next)
				i++
				inLineComment = true
				continue
			}
			if ch == '/' && next == '*' {
				current.WriteByte(ch)
				current.WriteByte(next)
				i++
				inBlockComment = true
				continue
			}
			if ch == ';' {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
				continue
			}
		}

		current.WriteByte(ch)

		switch ch {
		case '\'':
			if !inDoubleQuote && !inBacktickQuote && !inBracketIdentifier {
				if inSingleQuote && next == '\'' {
					current.WriteByte(next)
					i++
					continue
				}
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote && !inBacktickQuote && !inBracketIdentifier {
				if inDoubleQuote && next == '"' {
					current.WriteByte(next)
					i++
					continue
				}
				inDoubleQuote = !inDoubleQuote
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote && !inBracketIdentifier {
				inBacktickQuote = !inBacktickQuote
			}
		case '[':
			if !inSingleQuote && !inDoubleQuote && !inBacktickQuote {
				inBracketIdentifier = true
			}
		case ']':
			if inBracketIdentifier && !inSingleQuote && !inDoubleQuote && !inBacktickQuote {
				inBracketIdentifier = false
			}
		}
	}

	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

var (
	createTablePrefixPattern = regexp.MustCompile(`(?is)^\s*CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?`)
	insertIntoPrefixPattern  = regexp.MustCompile(`(?is)^\s*INSERT\s+INTO\s+`)
)

func validateImportStatement(statement string, targetTable string) error {
	createRest := createTablePrefixPattern.ReplaceAllString(statement, "")
	if createRest != statement {
		objectName := readQualifiedIdentifier(createRest)
		if canonicalObjectName(objectName) != targetTable {
			return fmt.Errorf("SQL import only allows CREATE TABLE for the selected table")
		}
		return nil
	}

	insertRest := insertIntoPrefixPattern.ReplaceAllString(statement, "")
	if insertRest != statement {
		objectName := readQualifiedIdentifier(insertRest)
		if canonicalObjectName(objectName) != targetTable {
			return fmt.Errorf("SQL import only allows INSERT INTO for the selected table")
		}
		return nil
	}

	return fmt.Errorf("unsafe SQL statement blocked during import")
}

func readQualifiedIdentifier(value string) string {
	value = strings.TrimLeftFunc(value, unicode.IsSpace)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < len(value); {
		ch := value[i]
		if unicode.IsSpace(rune(ch)) || ch == '(' {
			break
		}

		switch ch {
		case '"':
			end := consumeDelimitedIdentifier(value, i, '"')
			builder.WriteString(value[i:end])
			i = end
		case '`':
			end := consumeDelimitedIdentifier(value, i, '`')
			builder.WriteString(value[i:end])
			i = end
		case '[':
			end := consumeBracketIdentifier(value, i)
			builder.WriteString(value[i:end])
			i = end
		default:
			if isIdentifierChar(ch) || ch == '.' {
				builder.WriteByte(ch)
				i++
				continue
			}
			return strings.TrimSpace(builder.String())
		}
	}

	return strings.TrimSpace(builder.String())
}

func consumeDelimitedIdentifier(value string, start int, delimiter byte) int {
	i := start + 1
	for i < len(value) {
		if value[i] == delimiter {
			i++
			if i < len(value) && value[i] == delimiter {
				i++
				continue
			}
			return i
		}
		i++
	}
	return len(value)
}

func consumeBracketIdentifier(value string, start int) int {
	i := start + 1
	for i < len(value) {
		if value[i] == ']' {
			i++
			if i < len(value) && value[i] == ']' {
				i++
				continue
			}
			return i
		}
		i++
	}
	return len(value)
}

func isIdentifierChar(ch byte) bool {
	return ch == '_' || ch == '$' || unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch))
}

func canonicalObjectName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	parts := strings.Split(value, ".")
	last := strings.TrimSpace(parts[len(parts)-1])
	last = strings.TrimPrefix(last, `"`)
	last = strings.TrimSuffix(last, `"`)
	last = strings.TrimPrefix(last, "`")
	last = strings.TrimSuffix(last, "`")
	last = strings.TrimPrefix(last, "[")
	last = strings.TrimSuffix(last, "]")
	return strings.ToLower(strings.TrimSpace(last))
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
	a.approveWritePath(selection)
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
	a.approveReadPath(selection)
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
	a.approveReadPath(selection)
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
