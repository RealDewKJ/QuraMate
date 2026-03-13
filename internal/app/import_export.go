package app

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

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

	identityInsertEnabled := false
	disableIdentityInsert := func() error {
		if !identityInsertEnabled {
			return nil
		}
		identityInsertEnabled = false
		if _, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s OFF", tableName)); err != nil {
			return fmt.Errorf("setting IDENTITY_INSERT OFF: %w", err)
		}
		return nil
	}

	if enableIdentityInsert && db.Type == "mssql" {
		_, err := tx.Exec(fmt.Sprintf("SET IDENTITY_INSERT %s ON", tableName))
		if err != nil {
			return fmt.Sprintf("Error setting IDENTITY_INSERT ON: %s", err.Error())
		}
		identityInsertEnabled = true
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
