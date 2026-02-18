package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

type DBConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	ReadOnly bool   `json:"readOnly"`
}

type Database struct {
	conn           *sql.DB
	persistentConn *sql.Conn
	Type           string
	ReadOnly       bool
}

func NewDatabase() *Database {
	return &Database{}
}

func (d *Database) Connect(config DBConfig) error {
	if d.conn != nil {
		d.Disconnect()
	}

	var dsn string
	var driverName string

	switch config.Type {
	case "postgres":
		driverName = "pgx"
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.User, config.Password, config.Host, config.Port, config.Database)
	case "mysql":
		driverName = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.User, config.Password, config.Host, config.Port, config.Database)
	case "mssql":
		driverName = "sqlserver"
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable", config.User, config.Password, config.Host, config.Port, config.Database)
	case "sqlite":
		driverName = "sqlite"
		dsn = config.Database // Path to DB file
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}

	// Ping to ensure connectivity
	err = conn.Ping()
	if err != nil {
		return err
	}

	// Acquire a dedicated connection for this session
	// This ensures that all queries executed by this Database instance share the same underlying connection,
	// preserving transaction state and session-level settings.
	persistentConn, err := conn.Conn(context.Background())
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to acquire dedicated connection: %w", err)
	}

	d.conn = conn
	d.persistentConn = persistentConn
	d.Type = config.Type
	d.ReadOnly = config.ReadOnly
	return nil
}

func (d *Database) Disconnect() error {
	var err error
	if d.persistentConn != nil {
		err = d.persistentConn.Close()
		d.persistentConn = nil
	}
	if d.conn != nil {
		if closeErr := d.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		d.conn = nil
	}
	return err
}

func (d *Database) BeginTransaction() (*sql.Tx, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return d.persistentConn.BeginTx(context.Background(), nil)
}

func (d *Database) SetReadOnly(readOnly bool) {
	d.ReadOnly = readOnly
}

func (d *Database) GetTables() ([]string, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres":
		query = "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'"
	case "mysql":
		query = "SHOW TABLES"
	case "mssql":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'"
	case "sqlite":
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	default:
		return nil, fmt.Errorf("unsupported database type for getting tables")
	}

	rows, err := d.persistentConn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Message string          `json:"message,omitempty"`
}

func (d *Database) ExecuteQuery(ctx context.Context, query string) ([]ResultSet, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	// Always use QueryContext to support multiple result sets (even for INSERT/UPDATE which might return results or just be part of a batch)
	// We need to handle the case where the driver doesn't support multiple result sets gracefully if possible,
	// but standard database/sql logic is to just use NextResultSet().

	rows, err := d.persistentConn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultSets []ResultSet

	for {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		columns, err := rows.Columns()
		if err != nil {
			// This might happen if it's a result set without columns (like an UPDATE result in some drivers?)
			// primarily checking if we can proceed.
			// Some drivers might verify columns availability.
			// If no columns, maybe it's just a command result?
			// But QueryContext usually expects rows.
			// Let's defer to seeing if we can scan.
			// Actually, if Columns() fails, we might just be at the end or it's not a select.
			// However, in Go `database/sql`, Exec results are not easily retrieveable via Query.
			// BUT, for SQL Server "WAITFOR...; SELECT..." and "UPDATE...; SELECT...", QueryContext IS the way to go to get the SELECT part.
			// If the first part is NOT a select, rows.Columns() might be empty or error depending on driver.
		}

		var currentSet ResultSet
		currentSet.Columns = columns

		// If we have columns, let's scan rows
		if len(columns) > 0 {
			nCols := len(columns)
			// Pre-allocate scan targets once, reuse for every row
			scanTargets := make([]interface{}, nCols)
			scanPtrs := make([]*interface{}, nCols)
			for i := 0; i < nCols; i++ {
				scanPtrs[i] = new(interface{})
				scanTargets[i] = scanPtrs[i]
			}
			// Pre-allocate rows slice
			currentSet.Rows = make([][]interface{}, 0, 256)

			for rows.Next() {
				// Check for cancellation during row processing
				if err := ctx.Err(); err != nil {
					return nil, err
				}

				if err := rows.Scan(scanTargets...); err != nil {
					return nil, err
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
				currentSet.Rows = append(currentSet.Rows, row)
			}
		}

		// Even if loop didn't run (no rows), we might have columns (empty result set).
		// Or if no columns, maybe it was an exec.

		resultSets = append(resultSets, currentSet)

		if !rows.NextResultSet() {
			break
		}
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return resultSets, nil
}

// ExecuteTransientQuery executes a query using the connection pool (d.conn) instead of the persistent connection.
// This allows it to run concurrently with other queries on the persistent connection, but it won't see
// uncommitted transactions or session-local state from the persistent connection.
func (d *Database) ExecuteTransientQuery(ctx context.Context, query string) ([]ResultSet, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	// Use d.conn directly
	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultSets []ResultSet

	for {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		columns, err := rows.Columns()
		if err != nil {
			// Handle error or empty columns
		}

		var currentSet ResultSet
		currentSet.Columns = columns

		if len(columns) > 0 {
			nCols := len(columns)
			scanTargets := make([]interface{}, nCols)
			scanPtrs := make([]*interface{}, nCols)
			for i := 0; i < nCols; i++ {
				scanPtrs[i] = new(interface{})
				scanTargets[i] = scanPtrs[i]
			}
			currentSet.Rows = make([][]interface{}, 0, 256)

			for rows.Next() {
				if err := ctx.Err(); err != nil {
					return nil, err
				}

				if err := rows.Scan(scanTargets...); err != nil {
					return nil, err
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
				currentSet.Rows = append(currentSet.Rows, row)
			}
		}

		resultSets = append(resultSets, currentSet)

		if !rows.NextResultSet() {
			break
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return resultSets, nil
}

// StreamBatch represents a batch of rows sent during streaming
type StreamBatch struct {
	Columns      []string        `json:"columns"`
	Rows         [][]interface{} `json:"rows"`
	ResultSetIdx int             `json:"resultSetIdx"`
	BatchIndex   int             `json:"batchIndex"`
}

// ExecuteQueryStream executes a query and streams results in batches via the onBatch callback.
// batchSize controls how many rows are buffered before emitting a batch.
func (d *Database) ExecuteQueryStream(ctx context.Context, query string, batchSize int, onBatch func(batch StreamBatch)) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	rows, err := d.persistentConn.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	resultSetIdx := 0

	for {
		if err := ctx.Err(); err != nil {
			return err
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

			batch := make([][]interface{}, 0, batchSize)
			batchIndex := 0

			for rows.Next() {
				if err := ctx.Err(); err != nil {
					return err
				}

				if err := rows.Scan(scanTargets...); err != nil {
					return err
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
				batch = append(batch, row)

				if len(batch) >= batchSize {
					onBatch(StreamBatch{
						Columns:      columns,
						Rows:         batch,
						ResultSetIdx: resultSetIdx,
						BatchIndex:   batchIndex,
					})
					batch = make([][]interface{}, 0, batchSize)
					batchIndex++
				}
			}

			// Emit remaining rows
			onBatch(StreamBatch{
				Columns:      columns,
				Rows:         batch,
				ResultSetIdx: resultSetIdx,
				BatchIndex:   batchIndex,
			})
		} else {
			// No-column result set (e.g., UPDATE/INSERT)
			onBatch(StreamBatch{
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
		return err
	}

	return nil
}

func (d *Database) GetPrimaryKeys(tableName string) ([]string, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres":
		query = fmt.Sprintf(`
			SELECT a.attname
			FROM   pg_index i
			JOIN   pg_attribute a ON a.attrelid = i.indrelid
								 AND a.attnum = ANY(i.indkey)
			WHERE  i.indrelid = '%s'::regclass
			AND    i.indisprimary`, tableName)
	case "mysql":
		query = fmt.Sprintf(`
			SELECT COLUMN_NAME
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = '%s'
			AND COLUMN_KEY = 'PRI'`, tableName)
	case "mssql":
		query = fmt.Sprintf(`
			SELECT COLUMN_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE OBJECTPROPERTY(OBJECT_ID(CONSTRAINT_SCHEMA + '.' + CONSTRAINT_NAME), 'IsPrimaryKey') = 1
			AND TABLE_NAME = '%s'`, tableName)
	case "sqlite":
		// SQLite requires parsing "PRAGMA table_info(tableName)"
		// We can't use a simple SELECT for this in the same way, so we handle it differently
		return d.getSqlitePrimaryKeys(tableName)
	default:
		return nil, fmt.Errorf("unsupported database type for getting primary keys")
	}

	rows, err := d.persistentConn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pks []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	return pks, nil
}

func (d *Database) getSqlitePrimaryKeys(tableName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := d.persistentConn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pks []string
	for rows.Next() {
		var cid int
		var name string
		var typeName string
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &typeName, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		if pk > 0 {
			pks = append(pks, name)
		}
	}
	return pks, nil
}

func (d *Database) UpdateRecord(tableName string, updates map[string]interface{}, conditions map[string]interface{}) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	if len(updates) == 0 {
		return nil
	}

	var setClauses []string
	var args []interface{}
	var paramCount int

	getPlaceholder := func() string {
		paramCount++
		switch d.Type {
		case "postgres":
			return fmt.Sprintf("$%d", paramCount)
		case "mssql":
			return fmt.Sprintf("@p%d", paramCount)
		default:
			return "?"
		}
	}

	for col, val := range updates {
		// Check for DEFAULT value marker
		isDefault := false
		if m, ok := val.(map[string]interface{}); ok {
			if _, hasKey := m["_vaultdb_sql_default"]; hasKey {
				isDefault = true
			}
		}

		if isDefault {
			setClauses = append(setClauses, fmt.Sprintf("%s = DEFAULT", col))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, getPlaceholder()))
			args = append(args, val)
		}
	}

	var whereClauses []string
	for col, val := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", col, getPlaceholder()))
		args = append(args, val)
	}

	setStr := ""
	for i, clause := range setClauses {
		if i > 0 {
			setStr += ", "
		}
		setStr += clause
	}

	whereStr := ""
	if len(whereClauses) > 0 {
		whereStr = " WHERE "
		for i, clause := range whereClauses {
			if i > 0 {
				whereStr += " AND "
			}
			whereStr += clause
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s%s", tableName, setStr, whereStr)

	_, err := d.persistentConn.ExecContext(context.Background(), query, args...)
	return err
}

func (d *Database) InsertRecord(tableName string, data map[string]interface{}) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	if len(data) == 0 {
		return nil
	}

	var columns []string
	var placeholders []string
	var args []interface{}
	var paramCount int

	getPlaceholder := func() string {
		paramCount++
		switch d.Type {
		case "postgres":
			return fmt.Sprintf("$%d", paramCount)
		case "mssql":
			return fmt.Sprintf("@p%d", paramCount)
		default:
			return "?"
		}
	}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, getPlaceholder())
		args = append(args, val)
	}

	colsStr := strings.Join(columns, ", ")
	valsStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, colsStr, valsStr)

	_, err := d.persistentConn.ExecContext(context.Background(), query, args...)
	return err
}

// InsertRecordTx is used within a transaction, so it uses the passed *sql.Tx
// This does NOT use d.persistentConn directly, but the Tx itself is tied to it.
func (d *Database) InsertRecordTx(tx *sql.Tx, tableName string, data map[string]interface{}) error {
	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	if len(data) == 0 {
		return nil
	}

	var columns []string
	var placeholders []string
	var args []interface{}
	var paramCount int

	getPlaceholder := func() string {
		paramCount++
		switch d.Type {
		case "postgres":
			return fmt.Sprintf("$%d", paramCount)
		case "mssql":
			return fmt.Sprintf("@p%d", paramCount)
		default:
			return "?"
		}
	}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, getPlaceholder())
		args = append(args, val)
	}

	colsStr := strings.Join(columns, ", ")
	valsStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, colsStr, valsStr)

	_, err := tx.Exec(query, args...)
	return err
}

func (d *Database) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres":
		query = fmt.Sprintf(`
			SELECT
				tc.table_name, kcu.column_name,
				ccu.table_name AS foreign_table_name,
				ccu.column_name AS foreign_column_name,
				tc.constraint_name
			FROM
				information_schema.table_constraints AS tc
				JOIN information_schema.key_column_usage AS kcu
				  ON tc.constraint_name = kcu.constraint_name
				  AND tc.table_schema = kcu.table_schema
				JOIN information_schema.constraint_column_usage AS ccu
				  ON ccu.constraint_name = tc.constraint_name
				  AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY' 
            AND (tc.table_name = '%s' OR ccu.table_name = '%s');`, tableName, tableName)
	case "mysql":
		query = fmt.Sprintf(`
			SELECT
				TABLE_NAME, COLUMN_NAME,
				REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME,
				CONSTRAINT_NAME
			FROM
				INFORMATION_SCHEMA.KEY_COLUMN_USAGE
			WHERE
				REFERENCED_TABLE_SCHEMA = DATABASE()
				AND (TABLE_NAME = '%s' OR REFERENCED_TABLE_NAME = '%s');`, tableName, tableName)
	case "mssql":
		query = fmt.Sprintf(`
			SELECT
				tp.name AS TableName,
				cp.name AS ColumnName,
				tr.name AS ReferencedTableName,
				cr.name AS ReferencedColumnName,
				fk.name AS ConstraintName
			FROM
				sys.foreign_keys AS fk
				INNER JOIN sys.tables AS tp ON fk.parent_object_id = tp.object_id
				INNER JOIN sys.tables AS tr ON fk.referenced_object_id = tr.object_id
				INNER JOIN sys.foreign_key_columns AS fkc ON fkc.constraint_object_id = fk.object_id
				INNER JOIN sys.columns AS cp ON fkc.parent_column_id = cp.column_id AND fkc.parent_object_id = cp.object_id
				INNER JOIN sys.columns AS cr ON fkc.referenced_column_id = cr.column_id AND fkc.referenced_object_id = cr.object_id
			WHERE
				tp.name = '%s' OR tr.name = '%s';`, tableName, tableName)
	case "sqlite":
		query = fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
		// SQLite returns id, seq, table, from, to, on_update, on_delete, match
		// We'll handle this separately as the columns are different
		return d.getSqliteForeignKeys(tableName)
	default:
		return nil, fmt.Errorf("unsupported database type for getting foreign keys")
	}

	rows, err := d.persistentConn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []ForeignKey
	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Table, &fk.Column, &fk.RefTable, &fk.RefColumn, &fk.Constraint); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, nil
}

func (d *Database) getSqliteForeignKeys(tableName string) ([]ForeignKey, error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	rows, err := d.persistentConn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []ForeignKey
	for rows.Next() {
		var id, seq int
		var table, from, to, on_update, on_delete, match string
		// id, seq, table, from, to, on_update, on_delete, match
		if err := rows.Scan(&id, &seq, &table, &from, &to, &on_update, &on_delete, &match); err != nil {
			return nil, err
		}
		fks = append(fks, ForeignKey{
			Table:      tableName,
			Column:     from,
			RefTable:   table,
			RefColumn:  to,
			Constraint: fmt.Sprintf("FK_%s_%d", tableName, id),
		})
	}
	return fks, nil
}

// ExplainQuery retrieves the execution plan for the given query
func (d *Database) ExplainQuery(ctx context.Context, query string) (string, error) {
	if d.persistentConn == nil {
		return "", fmt.Errorf("no database connection")
	}

	var explainQuery string

	switch d.Type {
	case "postgres":
		explainQuery = "EXPLAIN " + query
	case "mysql":
		explainQuery = "EXPLAIN " + query
	case "sqlite":
		explainQuery = "EXPLAIN QUERY PLAN " + query
	case "mssql":
		// MSSQL requires SET SHOWPLAN_TEXT to be the only statement in the batch.
		// We must execute them separately on the persistent connection.
		if _, err := d.persistentConn.ExecContext(ctx, "SET SHOWPLAN_TEXT ON"); err != nil {
			return "", fmt.Errorf("failed to enable showplan: %w", err)
		}

		// Ensure we turn it off even if the query fails
		defer func() {
			d.persistentConn.ExecContext(context.Background(), "SET SHOWPLAN_TEXT OFF")
		}()

		explainQuery = query
	default:
		return "", fmt.Errorf("explain not supported for database type: %s", d.Type)
	}

	rows, err := d.persistentConn.QueryContext(ctx, explainQuery)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Handle multiple result sets.
	// For MSSQL with batching, the first result might be from SET (empty) or the actual plan.
	// We need to iterate through result sets to find the one with rows.

	var planBuilder strings.Builder

	foundPlan := false
	for {
		columns, err := rows.Columns()
		if err != nil {
			// Could happen if it's just a command result
		} else if len(columns) > 0 {
			foundPlan = true
			nCols := len(columns)
			scanTargets := make([]interface{}, nCols)
			scanPtrs := make([]*interface{}, nCols)
			for i := 0; i < nCols; i++ {
				scanPtrs[i] = new(interface{})
				scanTargets[i] = scanPtrs[i]
			}

			// Add header if useful? Maybe not for simple text plan
			// planBuilder.WriteString(strings.Join(columns, " | ") + "\n")
			// planBuilder.WriteString(strings.Repeat("-", 20) + "\n")

			for rows.Next() {
				if err := rows.Scan(scanTargets...); err != nil {
					return "", err
				}

				var rowStrs []string
				for i := 0; i < nCols; i++ {
					val := *scanPtrs[i]
					if b, ok := val.([]byte); ok {
						rowStrs = append(rowStrs, string(b))
					} else {
						rowStrs = append(rowStrs, fmt.Sprintf("%v", val))
					}
				}
				planBuilder.WriteString(strings.Join(rowStrs, " | ") + "\n")
			}
		}

		if !rows.NextResultSet() {
			break
		}
	}

	if !foundPlan {
		return "No execution plan returned.", nil
	}

	return planBuilder.String(), nil
}
