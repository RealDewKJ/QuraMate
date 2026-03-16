package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"
)

type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Message string          `json:"message,omitempty"`
}

// ServerProcess represents an active session on the database server

type ServerProcess struct {
	SessionID   string `json:"sessionId"`
	User        string `json:"user"`
	Host        string `json:"host"`
	Database    string `json:"database"`
	Command     string `json:"command"`
	Status      string `json:"status"`
	State       string `json:"state"`
	WaitTime    int64  `json:"waitTime"` // in milliseconds
	WaitType    string `json:"waitType"`
	QueryText   string `json:"queryText"`
	ElapsedTime int64  `json:"elapsedTime"` // in milliseconds
	HeadBlock   string `json:"headBlock"`
}

func (d *Database) ExecuteQuery(ctx context.Context, query string) ([]ResultSet, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}
	if err := ensureQueryAllowedInReadOnlyMode(query, d.ReadOnly); err != nil {
		return nil, err
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
						row[i] = decodeValue(b, d.Encoding)
					} else if s, ok := val.(string); ok {
						row[i] = decodeStringValue(s, d.Encoding)
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
	if err := ensureQueryAllowedInReadOnlyMode(query, d.ReadOnly); err != nil {
		return nil, err
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
						row[i] = decodeValue(b, d.Encoding)
					} else if s, ok := val.(string); ok {
						row[i] = decodeStringValue(s, d.Encoding)
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

// ColumnMetadata holds type information for a column

type ColumnMetadata struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Length   int64  `json:"length"`
	Nullable bool   `json:"nullable"`
}

// StreamBatch represents a batch of rows sent during streaming

type StreamBatch struct {
	Columns      []string         `json:"columns"`
	ColumnTypes  []ColumnMetadata `json:"columnTypes"`
	Rows         [][]interface{}  `json:"rows"`
	ResultSetIdx int              `json:"resultSetIdx"`
	BatchIndex   int              `json:"batchIndex"`
}

// ExecuteQueryStream executes a query and streams results in batches via the onBatch callback.
// batchSize controls how many rows are buffered before emitting a batch.

func (d *Database) ExecuteQueryStream(ctx context.Context, query string, batchSize int, onBatch func(batch StreamBatch)) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}
	if err := ensureQueryAllowedInReadOnlyMode(query, d.ReadOnly); err != nil {
		return err
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
						row[i] = decodeValue(b, d.Encoding)
					} else if s, ok := val.(string); ok {
						row[i] = decodeStringValue(s, d.Encoding)
					} else {
						row[i] = val
					}
				}
				batch = append(batch, row)

				if len(batch) >= batchSize {
					onBatch(StreamBatch{
						Columns:      columns,
						ColumnTypes:  columnMetas,
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
				ColumnTypes:  columnMetas,
				Rows:         batch,
				ResultSetIdx: resultSetIdx,
				BatchIndex:   batchIndex,
			})
		} else {
			// No-column result set (e.g., UPDATE/INSERT)
			onBatch(StreamBatch{
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
		return err
	}

	return nil
}

func ensureQueryAllowedInReadOnlyMode(query string, readOnly bool) error {
	if !readOnly {
		return nil
	}

	statements := splitSQLStatements(query)
	if len(statements) == 0 {
		return nil
	}

	for _, statement := range statements {
		if !isReadOnlyStatement(statement) {
			return fmt.Errorf("database is in read-only mode")
		}
	}

	return nil
}

func splitSQLStatements(query string) []string {
	var statements []string
	var current strings.Builder
	depth := 0
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	inBracketIdentifier := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(query); i++ {
		ch := query[i]
		next := byte(0)
		if i+1 < len(query) {
			next = query[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				current.WriteByte(ch)
			}
			continue
		}

		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inSingleQuote {
			if ch == '\'' {
				if next == '\'' {
					i++
					continue
				}
				inSingleQuote = false
			}
			continue
		}

		if inDoubleQuote {
			if ch == '"' {
				if next == '"' {
					i++
					continue
				}
				inDoubleQuote = false
			}
			current.WriteByte(ch)
			continue
		}

		if inBacktick {
			if ch == '`' {
				inBacktick = false
			}
			current.WriteByte(ch)
			continue
		}

		if inBracketIdentifier {
			current.WriteByte(ch)
			if ch == ']' {
				inBracketIdentifier = false
			}
			continue
		}

		switch {
		case ch == '-' && next == '-':
			inLineComment = true
			i++
			continue
		case ch == '/' && next == '*':
			inBlockComment = true
			i++
			continue
		case ch == '\'':
			inSingleQuote = true
			continue
		case ch == '"':
			inDoubleQuote = true
			current.WriteByte(ch)
			continue
		case ch == '`':
			inBacktick = true
			current.WriteByte(ch)
			continue
		case ch == '[':
			inBracketIdentifier = true
			current.WriteByte(ch)
			continue
		case ch == '(':
			depth++
			current.WriteByte(ch)
			continue
		case ch == ')':
			if depth > 0 {
				depth--
			}
			current.WriteByte(ch)
			continue
		case ch == ';' && depth == 0:
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		default:
			current.WriteByte(ch)
		}
	}

	statement := strings.TrimSpace(current.String())
	if statement != "" {
		statements = append(statements, statement)
	}

	return statements
}

func isReadOnlyStatement(statement string) bool {
	tokens := tokenizeSQL(statement)
	if len(tokens) == 0 {
		return true
	}

	first := tokens[0]
	switch first {
	case "select", "show", "describe", "desc", "pragma", "values":
		return true
	case "explain":
		return isExplainReadOnly(tokens[1:])
	case "with":
		return isWithReadOnly(tokens[1:])
	default:
		return false
	}
}

func isExplainReadOnly(tokens []string) bool {
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		switch token {
		case "analyze", "analyse", "verbose", "costs", "settings", "buffers", "wal", "timing", "summary", "format", ",":
			continue
		default:
			filtered = append(filtered, token)
		}
	}

	if len(filtered) == 0 {
		return false
	}

	return isReadOnlyStatement(strings.Join(filtered, " "))
}

func isWithReadOnly(tokens []string) bool {
	depth := 0
	for _, token := range tokens {
		if isMutatingStatementToken(token) {
			return false
		}

		switch token {
		case "(":
			depth++
		case ")":
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && isStatementStarter(token) {
				return isReadOnlyStatement(token)
			}
		}
	}

	return false
}

func isMutatingStatementToken(token string) bool {
	switch token {
	case "insert", "update", "delete", "merge", "call", "exec", "execute", "create", "alter", "drop", "truncate", "grant", "revoke", "set", "use", "begin", "start", "commit", "rollback":
		return true
	default:
		return false
	}
}

func isStatementStarter(token string) bool {
	switch token {
	case "select", "insert", "update", "delete", "merge", "values", "show", "describe", "desc", "pragma", "with", "explain", "call", "exec", "execute", "create", "alter", "drop", "truncate", "grant", "revoke", "set", "use", "begin", "start", "commit", "rollback":
		return true
	default:
		return false
	}
}

func tokenizeSQL(statement string) []string {
	tokens := make([]string, 0)
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(current.String()))
		current.Reset()
	}

	for _, r := range statement {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$':
			current.WriteRune(r)
		case r == '(' || r == ')' || r == ',':
			flush()
			tokens = append(tokens, string(r))
		default:
			flush()
		}
	}

	flush()
	return tokens
}

func (d *Database) GetServerProcesses(ctx context.Context) ([]ServerProcess, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "mssql":
		query = `
			SELECT 
				er.session_id as SessionID,
				es.login_name as [User],
				es.host_name as Host,
				DB_NAME(er.database_id) as [Database],
				er.command as Command,
				er.status as Status,
				es.status as State,
				er.wait_time as WaitTime,
				er.wait_type as WaitType,
				st.text as QueryText,
				er.total_elapsed_time as ElapsedTime,
				CAST(NULLIF(er.blocking_session_id, 0) AS VARCHAR) as HeadBlock
			FROM sys.dm_exec_requests er
			INNER JOIN sys.dm_exec_sessions es ON er.session_id = es.session_id
			CROSS APPLY sys.dm_exec_sql_text(er.sql_handle) st
			WHERE es.is_user_process = 1 AND er.session_id <> @@SPID
		`
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = `
			SELECT 
				pid as SessionID,
				usename as "User",
				client_addr::text as Host,
				datname as "Database",
				'' as Command,
				state as Status,
				state as State,
				0 as WaitTime,
				wait_event_type as WaitType,
				query as QueryText,
				EXTRACT(EPOCH FROM (NOW() - query_start)) * 1000 as ElapsedTime,
				'' as HeadBlock
			FROM pg_stat_activity
			WHERE pid <> pg_backend_pid()
		`
	case "mysql", "mariadb", "databend":
		query = `
			SELECT 
				ID as SessionID,
				USER as "User",
				HOST as Host,
				DB as "Database",
				COMMAND as Command,
				STATE as Status,
				STATE as State,
				TIME * 1000 as WaitTime,
				'' as WaitType,
				INFO as QueryText,
				TIME * 1000 as ElapsedTime,
				'' as HeadBlock
			FROM information_schema.PROCESSLIST
			WHERE ID <> CONNECTION_ID()
		`
	default:
		return nil, fmt.Errorf("unsupported database type for getting server processes")
	}

	rows, err := d.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []ServerProcess
	for rows.Next() {
		var sp ServerProcess
		var queryText, dbName, hostName, command, status, state, waitType, headBlock sql.NullString
		var waitTime, elapsedTime sql.NullInt64

		// Scan generic types to handle nulls
		// SessionID can be string or int depending on DB, but defined as string in struct
		// We'll scan into a string for SessionID
		var sessionID interface{}

		err = rows.Scan(
			&sessionID,
			&sp.User,
			&hostName,
			&dbName,
			&command,
			&status,
			&state,
			&waitTime,
			&waitType,
			&queryText,
			&elapsedTime,
			&headBlock,
		)
		if err != nil {
			return nil, err
		}

		sp.SessionID = fmt.Sprintf("%v", sessionID)
		sp.Host = hostName.String
		sp.Database = dbName.String
		sp.Command = command.String
		sp.Status = status.String
		sp.State = state.String
		sp.WaitType = waitType.String
		sp.QueryText = queryText.String
		sp.HeadBlock = headBlock.String

		if waitTime.Valid {
			sp.WaitTime = waitTime.Int64
		}
		if elapsedTime.Valid {
			sp.ElapsedTime = elapsedTime.Int64
		}

		processes = append(processes, sp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return processes, nil
}

func (d *Database) KillServerProcess(ctx context.Context, sessionID string) error {
	if d.conn == nil {
		return fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "mssql":
		query = fmt.Sprintf("KILL %s", sessionID)
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = fmt.Sprintf("SELECT pg_terminate_backend(%s)", sessionID)
	case "mysql", "mariadb", "databend":
		query = fmt.Sprintf("KILL %s", sessionID)
	default:
		return fmt.Errorf("unsupported database type for killing server processes")
	}

	_, err := d.conn.ExecContext(ctx, query)
	return err
}

func (d *Database) quoteIdentifier(name string) string {
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb", "sqlite", "libsql":
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case "mysql", "mariadb", "databend":
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "mssql":
		return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
	default:
		return name
	}
}

func (d *Database) QuoteIdentifier(name string) string {
	return d.quoteIdentifier(name)
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
		case "postgres", "greenplum", "redshift", "cockroachdb":
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
			if _, hasKey := m["_quramate_sql_default"]; hasKey {
				isDefault = true
			}
		}

		quotedCol := d.quoteIdentifier(col)

		if isDefault {
			setClauses = append(setClauses, fmt.Sprintf("%s = DEFAULT", quotedCol))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", quotedCol, getPlaceholder()))
			args = append(args, val)
		}
	}

	var whereClauses []string
	for col, val := range conditions {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", d.quoteIdentifier(col), getPlaceholder()))
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

	query := fmt.Sprintf("UPDATE %s SET %s%s", d.quoteIdentifier(tableName), setStr, whereStr)

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
		case "postgres", "greenplum", "redshift", "cockroachdb":
			return fmt.Sprintf("$%d", paramCount)
		case "mssql":
			return fmt.Sprintf("@p%d", paramCount)
		default:
			return "?"
		}
	}

	for col, val := range data {
		columns = append(columns, d.quoteIdentifier(col))
		placeholders = append(placeholders, getPlaceholder())
		args = append(args, val)
	}

	colsStr := strings.Join(columns, ", ")
	valsStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", d.quoteIdentifier(tableName), colsStr, valsStr)

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
		case "postgres", "greenplum", "redshift", "cockroachdb":
			return fmt.Sprintf("$%d", paramCount)
		case "mssql":
			return fmt.Sprintf("@p%d", paramCount)
		default:
			return "?"
		}
	}

	for col, val := range data {
		columns = append(columns, d.quoteIdentifier(col))
		placeholders = append(placeholders, getPlaceholder())
		args = append(args, val)
	}

	colsStr := strings.Join(columns, ", ")
	valsStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", d.quoteIdentifier(tableName), colsStr, valsStr)

	_, err := tx.Exec(query, args...)
	return err
}

func (d *Database) ExplainQuery(ctx context.Context, query string) (string, error) {
	if d.persistentConn == nil {
		return "", fmt.Errorf("no database connection")
	}

	var explainQuery string

	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		explainQuery = "EXPLAIN " + query
	case "mysql", "mariadb", "databend":
		explainQuery = "EXPLAIN " + query
	case "sqlite", "libsql":
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

// ColumnDefinition struct
