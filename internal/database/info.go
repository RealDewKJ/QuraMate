package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DatabaseInfo holds statistics and info about the database
type DatabaseInfo struct {
	Size         string `json:"size"`
	TableCount   int    `json:"tableCount"`
	ViewCount    int    `json:"viewCount"`
	RoutineCount int    `json:"routineCount"`
	DBName       string `json:"dbName"`
	Version      string `json:"version"`
	Engine       string `json:"engine"`
	Category     string `json:"category"`

	Summary       map[string]string `json:"summary"`
	Capabilities  map[string]bool   `json:"capabilities"`
	RuntimeInfo   map[string]string `json:"runtimeInfo"`
	EngineDetails map[string]string `json:"engineDetails"`
	Stats         map[string]string `json:"stats"`
}

func (d *Database) GetDatabaseInfo() (DatabaseInfo, error) {
	if d.persistentConn == nil {
		return DatabaseInfo{}, fmt.Errorf("no database connection")
	}

	queryString := func(query string, dest *string) error {
		if err := d.persistentConn.QueryRowContext(context.Background(), query).Scan(dest); err != nil {
			return fmt.Errorf("query %q: %w", query, err)
		}
		return nil
	}

	info := DatabaseInfo{
		Engine:       d.Type,
		Category:     databaseCategory(d.Type),
		Summary:      map[string]string{},
		Capabilities: databaseCapabilities(d.Type),
		RuntimeInfo:  map[string]string{},
		EngineDetails: map[string]string{
			"driverType": d.Type,
		},
		Stats: map[string]string{},
	}

	info.Summary["host"] = d.Host
	info.Summary["port"] = fmt.Sprintf("%d", d.Port)
	info.Summary["database"] = d.DatabaseName
	info.Summary["user"] = d.User
	info.Summary["readOnly"] = fmt.Sprintf("%t", d.ReadOnly)
	info.Summary["sshTunnel"] = fmt.Sprintf("%t", d.SSHEnabled)
	if !d.ConnectedAt.IsZero() {
		info.Summary["connectedAt"] = d.ConnectedAt.Format(time.RFC3339)
	}
	if strings.TrimSpace(d.Encoding) == "" {
		info.RuntimeInfo["appEncodingMode"] = "AUTO"
	} else {
		info.RuntimeInfo["appEncodingMode"] = d.Encoding
	}

	// Get Database Name
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		if err := queryString("SELECT current_database()", &info.DBName); err != nil {
			return DatabaseInfo{}, err
		}
		if err := queryString("SELECT version()", &info.Version); err != nil {
			return DatabaseInfo{}, err
		}
	case "mysql", "mariadb", "databend":
		if err := queryString("SELECT DATABASE()", &info.DBName); err != nil {
			return DatabaseInfo{}, err
		}
		if err := queryString("SELECT VERSION()", &info.Version); err != nil {
			return DatabaseInfo{}, err
		}
	case "mssql":
		if err := queryString("SELECT DB_NAME()", &info.DBName); err != nil {
			return DatabaseInfo{}, err
		}
		if err := queryString("SELECT @@VERSION", &info.Version); err != nil {
			return DatabaseInfo{}, err
		}
	case "sqlite", "libsql":
		info.DBName = "main"
		if err := queryString("SELECT sqlite_version()", &info.Version); err != nil {
			return DatabaseInfo{}, err
		}
	}
	info.Summary["activeDatabase"] = info.DBName
	info.RuntimeInfo["serverVersion"] = info.Version

	// Get Size
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		if err := queryString("SELECT pg_size_pretty(pg_database_size(current_database()))", &info.Size); err != nil {
			return DatabaseInfo{}, err
		}
	case "mysql", "mariadb", "databend":
		var sizeBytes int64
		if err := d.persistentConn.QueryRowContext(context.Background(), "SELECT SUM(data_length + index_length) FROM information_schema.TABLES WHERE table_schema = DATABASE()").Scan(&sizeBytes); err != nil {
			return DatabaseInfo{}, fmt.Errorf("query database size: %w", err)
		}
		info.Size = formatSize(sizeBytes)
	case "mssql":
		// Simple approach for MSSQL
		var databaseSize float64
		if err := d.persistentConn.QueryRowContext(context.Background(), "SELECT SUM(size) * 8 / 1024 FROM sys.master_files WHERE database_id = DB_ID()").Scan(&databaseSize); err != nil {
			return DatabaseInfo{}, fmt.Errorf("query database size: %w", err)
		}
		info.Size = fmt.Sprintf("%.2f MB", databaseSize)
	case "sqlite", "libsql":
		var pageCount, pageSize int64
		if err := d.persistentConn.QueryRowContext(context.Background(), "PRAGMA page_count").Scan(&pageCount); err != nil {
			return DatabaseInfo{}, fmt.Errorf("query page count: %w", err)
		}
		if err := d.persistentConn.QueryRowContext(context.Background(), "PRAGMA page_size").Scan(&pageSize); err != nil {
			return DatabaseInfo{}, fmt.Errorf("query page size: %w", err)
		}
		info.Size = formatSize(pageCount * pageSize)
	}

	// Counts
	tables, err := d.GetTables()
	if err != nil {
		return DatabaseInfo{}, fmt.Errorf("get tables: %w", err)
	}
	info.TableCount = len(tables)

	views, err := d.GetViews()
	if err != nil {
		return DatabaseInfo{}, fmt.Errorf("get views: %w", err)
	}
	info.ViewCount = len(views)

	procs, err := d.GetStoredProcedures()
	if err != nil {
		return DatabaseInfo{}, fmt.Errorf("get stored procedures: %w", err)
	}
	funcs, err := d.GetFunctions()
	if err != nil {
		return DatabaseInfo{}, fmt.Errorf("get functions: %w", err)
	}
	info.RoutineCount = len(procs) + len(funcs)

	info.Stats["size"] = info.Size
	info.Stats["tableCount"] = fmt.Sprintf("%d", info.TableCount)
	info.Stats["viewCount"] = fmt.Sprintf("%d", info.ViewCount)
	info.Stats["routineCount"] = fmt.Sprintf("%d", info.RoutineCount)

	d.enrichDatabaseInfoRuntime(&info)

	return info, nil
}

func databaseCategory(dbType string) string {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "mysql", "mariadb", "postgres", "greenplum", "redshift", "cockroachdb", "mssql", "sqlite", "libsql", "duckdb", "databend":
		return "sql"
	default:
		return "other"
	}
}

func databaseCapabilities(dbType string) map[string]bool {
	caps := map[string]bool{
		"supportsSchemas":          false,
		"supportsTables":           false,
		"supportsViews":            false,
		"supportsRoutines":         false,
		"supportsCollections":      false,
		"supportsCharset":          false,
		"supportsCollation":        false,
		"supportsTransactions":     true,
		"supportsSessionVariables": false,
	}

	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "mysql", "mariadb", "databend":
		caps["supportsTables"] = true
		caps["supportsViews"] = true
		caps["supportsRoutines"] = true
		caps["supportsCharset"] = true
		caps["supportsCollation"] = true
		caps["supportsSessionVariables"] = true
	case "postgres", "greenplum", "redshift", "cockroachdb":
		caps["supportsSchemas"] = true
		caps["supportsTables"] = true
		caps["supportsViews"] = true
		caps["supportsRoutines"] = true
		caps["supportsSessionVariables"] = true
	case "mssql":
		caps["supportsSchemas"] = true
		caps["supportsTables"] = true
		caps["supportsViews"] = true
		caps["supportsRoutines"] = true
	case "sqlite", "libsql", "duckdb":
		caps["supportsTables"] = true
		caps["supportsViews"] = true
		caps["supportsTransactions"] = true
	default:
		caps["supportsTransactions"] = false
	}

	return caps
}

func (d *Database) enrichDatabaseInfoRuntime(info *DatabaseInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	queryString := func(query string, args ...interface{}) string {
		var value sql.NullString
		if err := d.persistentConn.QueryRowContext(ctx, query, args...).Scan(&value); err != nil || !value.Valid {
			return ""
		}
		return strings.TrimSpace(value.String)
	}

	switch strings.ToLower(strings.TrimSpace(d.Type)) {
	case "mysql", "mariadb", "databend":
		pairs := map[string]string{
			"characterSetServer":     "SELECT @@character_set_server",
			"characterSetDatabase":   "SELECT @@character_set_database",
			"characterSetConnection": "SELECT @@character_set_connection",
			"characterSetClient":     "SELECT @@character_set_client",
			"characterSetResults":    "SELECT @@character_set_results",
			"collationServer":        "SELECT @@collation_server",
			"collationDatabase":      "SELECT @@collation_database",
			"collationConnection":    "SELECT @@collation_connection",
			"currentTime":            "SELECT DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s')",
		}
		for key, query := range pairs {
			if v := queryString(query); v != "" {
				info.RuntimeInfo[key] = v
			}
		}
	case "postgres", "greenplum", "redshift", "cockroachdb":
		pairs := map[string]string{
			"serverEncoding": "SHOW server_encoding",
			"timezone":       "SHOW timezone",
			"searchPath":     "SHOW search_path",
			"currentSchema":  "SELECT current_schema()",
			"currentTime":    "SELECT to_char(NOW(), 'YYYY-MM-DD HH24:MI:SS')",
		}
		for key, query := range pairs {
			if v := queryString(query); v != "" {
				info.RuntimeInfo[key] = v
			}
		}
	case "mssql":
		pairs := map[string]string{
			"serverCollation":   "SELECT CAST(SERVERPROPERTY('Collation') AS NVARCHAR(255))",
			"databaseCollation": "SELECT CAST(DATABASEPROPERTYEX(DB_NAME(), 'Collation') AS NVARCHAR(255))",
			"currentTime":       "SELECT CONVERT(VARCHAR(19), GETDATE(), 120)",
		}
		for key, query := range pairs {
			if v := queryString(query); v != "" {
				info.RuntimeInfo[key] = v
			}
		}
	case "sqlite", "libsql":
		if v := queryString("PRAGMA encoding"); v != "" {
			info.RuntimeInfo["encoding"] = v
		}
		if v := queryString("SELECT datetime('now')"); v != "" {
			info.RuntimeInfo["currentTime"] = v
		}
	case "duckdb":
		if v := queryString("SELECT version()"); v != "" {
			info.RuntimeInfo["engineVersion"] = v
		}
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("% d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
