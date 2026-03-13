package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (d *Database) GetTables() ([]string, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'"
	case "mysql", "mariadb", "databend":
		query = "SHOW TABLES"
	case "mssql":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE'"
	case "sqlite", "libsql":
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	case "duckdb":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('information_schema', 'pg_catalog') AND table_type = 'BASE TABLE'"
	default:
		return nil, fmt.Errorf("unsupported database type for getting tables")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (d *Database) GetViews() ([]string, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = "SELECT table_name FROM information_schema.views WHERE table_schema NOT IN ('information_schema', 'pg_catalog')"
	case "mysql", "mariadb", "databend":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.VIEWS WHERE TABLE_SCHEMA = DATABASE()"
	case "mssql":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.VIEWS"
	case "sqlite", "libsql":
		query = "SELECT name FROM sqlite_master WHERE type='view'"
	case "duckdb":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('information_schema', 'pg_catalog') AND table_type = 'VIEW'"
	default:
		return nil, fmt.Errorf("unsupported database type for getting views")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := []string{}
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return views, nil
}

func (d *Database) GetStoredProcedures() ([]string, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = "SELECT routine_name FROM information_schema.routines WHERE routine_type = 'PROCEDURE' AND routine_schema NOT IN ('information_schema', 'pg_catalog')"
	case "mysql", "mariadb", "databend":
		query = "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_TYPE = 'PROCEDURE' AND ROUTINE_SCHEMA = DATABASE()"
	case "mssql":
		query = "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_TYPE = 'PROCEDURE'"
	case "sqlite", "libsql", "duckdb":
		return []string{}, nil // SQLite and DuckDB don't support stored procedures in this context
	default:
		return nil, fmt.Errorf("unsupported database type for getting stored procedures")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	procs := []string{}
	for rows.Next() {
		var proc string
		if err := rows.Scan(&proc); err != nil {
			return nil, err
		}
		procs = append(procs, proc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return procs, nil
}

func (d *Database) GetFunctions() ([]string, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = "SELECT routine_name FROM information_schema.routines WHERE routine_type = 'FUNCTION' AND routine_schema NOT IN ('information_schema', 'pg_catalog')"
	case "mysql", "mariadb", "databend":
		query = "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_TYPE = 'FUNCTION' AND ROUTINE_SCHEMA = DATABASE()"
	case "mssql":
		query = "SELECT ROUTINE_NAME FROM INFORMATION_SCHEMA.ROUTINES WHERE ROUTINE_TYPE = 'FUNCTION'"
	case "sqlite", "libsql", "duckdb":
		return []string{}, nil // SQLite and DuckDB don't support stored functions in this way
	default:
		return nil, fmt.Errorf("unsupported database type for getting functions")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	funcs := []string{}
	for rows.Next() {
		var fn string
		if err := rows.Scan(&fn); err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return funcs, nil
}

func (d *Database) GetRoutineDefinition(name string, routineType string) (string, error) {
	if d.conn == nil {
		return "", fmt.Errorf("no database connection pool")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		// PostgreSQL: routine_definition might be null for some functions (like internal ones),
		// but for user-defined ones it should be there.
		// Alternatively, pg_get_functiondef() is more reliable for Postgres.
		query = fmt.Sprintf("SELECT pg_get_functiondef('%s'::regproc)", name)
	case "mysql", "mariadb", "databend":
		if routineType == "PROCEDURE" {
			query = fmt.Sprintf("SHOW CREATE PROCEDURE `%s`", name)
		} else {
			query = fmt.Sprintf("SHOW CREATE FUNCTION `%s`", name)
		}
	case "mssql":
		query = fmt.Sprintf("SELECT definition FROM sys.sql_modules WHERE object_id = OBJECT_ID('%s')", name)
	default:
		return "", fmt.Errorf("unsupported database type for getting routine definition")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		if d.Type == "mysql" || d.Type == "mariadb" || d.Type == "databend" {
			// MySQL's SHOW CREATE ... returns columns like (Procedure, sql_mode, Create Procedure, character_set_client, ...)
			// The definition is in the 3rd column (index 2)
			cols, _ := rows.Columns()
			dest := make([]interface{}, len(cols))
			for i := range dest {
				var v interface{}
				dest[i] = &v
			}
			if err := rows.Scan(dest...); err != nil {
				return "", err
			}
			if len(dest) >= 3 {
				val := *(dest[2].(*interface{}))
				if b, ok := val.([]byte); ok {
					return string(b), nil
				}
				return fmt.Sprintf("%v", val), nil
			}
			return "", fmt.Errorf("unexpected number of columns from SHOW CREATE")
		}

		var definition string
		if err := rows.Scan(&definition); err != nil {
			return "", err
		}
		return definition, nil
	}

	return "", fmt.Errorf("routine definition not found")
}

func (d *Database) GetPrimaryKeys(tableName string) ([]string, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = fmt.Sprintf(`
			SELECT a.attname
			FROM   pg_index i
			JOIN   pg_attribute a ON a.attrelid = i.indrelid
								 AND a.attnum = ANY(i.indkey)
			WHERE  i.indrelid = '%s'::regclass
			AND    i.indisprimary`, tableName)
	case "mysql", "mariadb", "databend":
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
	case "sqlite", "libsql", "duckdb":
		// SQLite and DuckDB support parsing "PRAGMA table_info(tableName)"
		// We can't use a simple SELECT for this in the same way, so we handle it differently
		return d.getSqlitePrimaryKeys(tableName)
	default:
		return nil, fmt.Errorf("unsupported database type for getting primary keys")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pks := []string{}
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pks, nil
}

func (d *Database) getSqlitePrimaryKeys(tableName string) ([]string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := d.conn.QueryContext(context.Background(), query)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pks, nil
}

func (d *Database) GetForeignKeys(tableName string) ([]ForeignKey, error) {
	if d.persistentConn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
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
	case "mysql", "mariadb", "databend":
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
	case "sqlite", "libsql":
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

	fks := []ForeignKey{}
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

	fks := []ForeignKey{}
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

type ColumnDefinition struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Nullable      bool        `json:"nullable"`
	DefaultValue  interface{} `json:"defaultValue"`
	PrimaryKey    bool        `json:"primaryKey"`
	AutoIncrement bool        `json:"autoIncrement"`
}

func (d *Database) GetTableDefinition(tableName string) ([]ColumnDefinition, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	var query string
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query = fmt.Sprintf(`
			SELECT 
				column_name, 
				data_type, 
				is_nullable, 
				column_default 
			FROM information_schema.columns 
			WHERE table_name = '%s'
			ORDER BY ordinal_position`, tableName)
	case "mysql", "mariadb", "databend":
		query = fmt.Sprintf(`
			SELECT 
				COLUMN_NAME, 
				COLUMN_TYPE, 
				IS_NULLABLE, 
				COLUMN_DEFAULT, 
				EXTRA 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION`, tableName)
	case "mssql":
		query = fmt.Sprintf(`
			SELECT 
				c.COLUMN_NAME, 
				c.DATA_TYPE, 
				c.CHARACTER_MAXIMUM_LENGTH,
				c.NUMERIC_PRECISION,
				c.NUMERIC_SCALE,
				c.DATETIME_PRECISION,
				c.IS_NULLABLE, 
				c.COLUMN_DEFAULT,
				COLUMNPROPERTY(OBJECT_ID(c.TABLE_NAME), c.COLUMN_NAME, 'IsIdentity') as IsIdentity
			FROM INFORMATION_SCHEMA.COLUMNS c
			WHERE c.TABLE_NAME = '%s'
			ORDER BY c.ORDINAL_POSITION`, tableName)
	case "sqlite", "libsql":
		return d.getSqliteTableDefinition(tableName)
	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get PKs to mark columns
	pks, _ := d.GetPrimaryKeys(tableName)
	pkMap := make(map[string]bool)
	for _, pk := range pks {
		pkMap[pk] = true
	}

	columns := []ColumnDefinition{}
	for rows.Next() {
		var name, dataType, isNullableStr string
		var defaultValue interface{}
		var extra string

		var err error
		switch d.Type {
		case "mysql", "mariadb", "databend":
			err = rows.Scan(&name, &dataType, &isNullableStr, &defaultValue, &extra)
		case "mssql":
			var charMax sql.NullInt64
			var numericPrecision sql.NullInt64
			var numericScale sql.NullInt64
			var datetimePrecision sql.NullInt64
			var isIdentity int
			err = rows.Scan(
				&name,
				&dataType,
				&charMax,
				&numericPrecision,
				&numericScale,
				&datetimePrecision,
				&isNullableStr,
				&defaultValue,
				&isIdentity,
			)
			if err == nil && isIdentity == 1 {
				extra = "identity" // Marker for later
			}
			if err == nil {
				baseType := strings.ToUpper(strings.TrimSpace(dataType))
				switch baseType {
				case "DECIMAL", "NUMERIC":
					precision := int64(18)
					scale := int64(0)
					if numericPrecision.Valid {
						precision = numericPrecision.Int64
					}
					if numericScale.Valid {
						scale = numericScale.Int64
					}
					dataType = fmt.Sprintf("%s(%d,%d)", baseType, precision, scale)
				case "VARCHAR", "NVARCHAR", "CHAR", "NCHAR", "BINARY", "VARBINARY":
					if charMax.Valid {
						if charMax.Int64 == -1 {
							dataType = fmt.Sprintf("%s(MAX)", baseType)
						} else {
							dataType = fmt.Sprintf("%s(%d)", baseType, charMax.Int64)
						}
					} else {
						dataType = baseType
					}
				case "DATETIME2", "DATETIMEOFFSET", "TIME":
					if datetimePrecision.Valid {
						dataType = fmt.Sprintf("%s(%d)", baseType, datetimePrecision.Int64)
					} else {
						dataType = baseType
					}
				default:
					dataType = baseType
				}
			}
		default:
			err = rows.Scan(&name, &dataType, &isNullableStr, &defaultValue)
		}

		if err != nil {
			return nil, err
		}

		isNullable := false
		if strings.ToUpper(isNullableStr) == "YES" || isNullableStr == "1" || isNullableStr == "true" {
			isNullable = true
		}

		col := ColumnDefinition{
			Name:         name,
			Type:         dataType,
			Nullable:     isNullable,
			DefaultValue: defaultValue,
			PrimaryKey:   pkMap[name],
		}

		if (d.Type == "mysql" || d.Type == "mariadb" || d.Type == "databend") && strings.Contains(strings.ToLower(extra), "auto_increment") {
			col.AutoIncrement = true
		}
		if d.Type == "mssql" && extra == "identity" {
			col.AutoIncrement = true
		}
		// Basic auto-increment detection for Postgres (serial types)
		if (d.Type == "postgres" || d.Type == "greenplum" || d.Type == "redshift" || d.Type == "cockroachdb") && defaultValue != nil {
			defStr := fmt.Sprintf("%v", defaultValue)
			if strings.Contains(defStr, "nextval") {
				col.AutoIncrement = true
			}
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func (d *Database) getSqliteTableDefinition(tableName string) ([]ColumnDefinition, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := d.conn.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ColumnDefinition{}
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

		columns = append(columns, ColumnDefinition{
			Name:          name,
			Type:          typeName,
			Nullable:      notnull == 0,
			DefaultValue:  dfltValue,
			PrimaryKey:    pk > 0,
			AutoIncrement: false, // Hard to detect reliably in SQLite without parsing DDL often, but can infer if INTEGER PRIMARY KEY
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

// IndexDefinition struct

type IndexDefinition struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

func (d *Database) GetTableIndexes(tableName string) ([]IndexDefinition, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("no database connection")
	}

	indexes := []IndexDefinition{}
	// Helper map to group columns by index name
	indexMap := make(map[string]*IndexDefinition)
	// Maintain order of index names as they are encountered
	var indexOrder []string

	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		query := fmt.Sprintf(`
			SELECT
				i.relname as index_name,
				a.attname as column_name,
				ix.indisunique,
				ix.indisprimary
			FROM
				pg_class t,
				pg_class i,
				pg_index ix,
				pg_attribute a
			WHERE
				t.oid = ix.indrelid
				AND i.oid = ix.indexrelid
				AND a.attrelid = t.oid
				AND a.attnum = ANY(ix.indkey)
				AND t.relkind = 'r'
				AND t.relname = '%s'`, tableName)
		rows, err := d.conn.QueryContext(context.Background(), query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var indexName, columnName string
			var isUnique, isPrimary bool
			if err := rows.Scan(&indexName, &columnName, &isUnique, &isPrimary); err != nil {
				return nil, err
			}

			if idx, exists := indexMap[indexName]; exists {
				idx.Columns = append(idx.Columns, columnName)
			} else {
				indexOrder = append(indexOrder, indexName)
				indexMap[indexName] = &IndexDefinition{
					Name:    indexName,
					Columns: []string{columnName},
					Unique:  isUnique,
					Primary: isPrimary,
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

	case "mysql", "mariadb", "databend":
		query := fmt.Sprintf(`
            SELECT 
                INDEX_NAME, 
                COLUMN_NAME, 
                NON_UNIQUE,
				SEQ_IN_INDEX
            FROM INFORMATION_SCHEMA.STATISTICS 
            WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '%s'
            ORDER BY INDEX_NAME, SEQ_IN_INDEX`, tableName)

		rows, err := d.conn.QueryContext(context.Background(), query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var indexName, columnName string
			var nonUnique, seqInIndex int
			if err := rows.Scan(&indexName, &columnName, &nonUnique, &seqInIndex); err != nil {
				return nil, err
			}

			if idx, exists := indexMap[indexName]; exists {
				idx.Columns = append(idx.Columns, columnName)
			} else {
				indexOrder = append(indexOrder, indexName)
				indexMap[indexName] = &IndexDefinition{
					Name:    indexName,
					Columns: []string{columnName},
					Unique:  nonUnique == 0,
					Primary: indexName == "PRIMARY",
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

	case "mssql":
		query := fmt.Sprintf(`
			SELECT 
				i.name AS IndexName,
				c.name AS ColumnName,
				i.is_unique,
				i.is_primary_key
			FROM 
				sys.indexes i
			INNER JOIN 
				sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			INNER JOIN 
				sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
			WHERE 
				i.object_id = OBJECT_ID('%s')
			ORDER BY 
				i.name, ic.index_column_id`, tableName)
		rows, err := d.conn.QueryContext(context.Background(), query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var indexName, columnName string
			var isUnique, isPrimary bool
			if err := rows.Scan(&indexName, &columnName, &isUnique, &isPrimary); err != nil {
				return nil, err
			}

			if idx, exists := indexMap[indexName]; exists {
				idx.Columns = append(idx.Columns, columnName)
			} else {
				indexOrder = append(indexOrder, indexName)
				indexMap[indexName] = &IndexDefinition{
					Name:    indexName,
					Columns: []string{columnName},
					Unique:  isUnique,
					Primary: isPrimary,
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

	case "sqlite", "libsql":
		query := fmt.Sprintf("PRAGMA index_list(%s)", tableName)
		rows, err := d.conn.QueryContext(context.Background(), query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		type sqliteIndexInfo struct {
			Name   string
			Unique bool
			Origin string
		}
		var sqliteIndexes []sqliteIndexInfo

		for rows.Next() {
			var seq int
			var name string
			var unique int
			var origin string
			var partial int
			if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				return nil, err
			}
			sqliteIndexes = append(sqliteIndexes, sqliteIndexInfo{Name: name, Unique: unique != 0, Origin: origin})
		}
		if err := closeRowsWithError(rows); err != nil {
			return nil, err
		}

		for _, idxInfo := range sqliteIndexes {
			infoQuery := fmt.Sprintf("PRAGMA index_info(%s)", idxInfo.Name)
			infoRows, err := d.conn.QueryContext(context.Background(), infoQuery)
			if err != nil {
				continue
			}

			var cols []string
			for infoRows.Next() {
				var seqno, cid int
				var name string
				if err := infoRows.Scan(&seqno, &cid, &name); err != nil {
					_ = infoRows.Close()
					return nil, err
				}
				cols = append(cols, name)
			}
			if err := closeRowsWithError(infoRows); err != nil {
				return nil, err
			}

			indexOrder = append(indexOrder, idxInfo.Name)
			indexMap[idxInfo.Name] = &IndexDefinition{
				Name:    idxInfo.Name,
				Columns: cols,
				Unique:  idxInfo.Unique,
				Primary: idxInfo.Origin == "pk",
			}
		}

	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	for _, name := range indexOrder {
		if idx, ok := indexMap[name]; ok {
			indexes = append(indexes, *idx)
		}
	}

	return indexes, nil
}

// TableChanges struct represents modifications to a table
