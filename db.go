package main

import (
	"database/sql"
	"fmt"

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
}

type Database struct {
	conn *sql.DB
	Type string
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

	err = conn.Ping()
	if err != nil {
		return err
	}

	d.conn = conn
	d.Type = config.Type
	return nil
}

func (d *Database) Disconnect() error {
	if d.conn != nil {
		err := d.conn.Close()
		d.conn = nil
		return err
	}
	return nil
}

func (d *Database) GetTables() ([]string, error) {
	if d.conn == nil {
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

	rows, err := d.conn.Query(query)
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

func (d *Database) ExecuteQuery(query string) ([]map[string]interface{}, []string, error) {
	if d.conn == nil {
		return nil, nil, fmt.Errorf("no database connection")
	}

	rows, err := d.conn.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := *(values[i].(*interface{}))

			// Handle []byte (common for string/blob data in some drivers)
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}
	return results, columns, nil
}

func (d *Database) GetPrimaryKeys(tableName string) ([]string, error) {
	if d.conn == nil {
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

	rows, err := d.conn.Query(query)
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
	rows, err := d.conn.Query(query)
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
	if d.conn == nil {
		return fmt.Errorf("no database connection")
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

	// We iterate over the map. Since map iteration order is random,
	// let's sort keys to be deterministic if needed,
	// but for now standard random iteration is fine as long as query and args match.
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, getPlaceholder()))
		args = append(args, val)
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

	_, err := d.conn.Exec(query, args...)
	return err
}
