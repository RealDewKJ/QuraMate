package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"unicode/utf8"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	// _ "github.com/marcboeker/go-duckdb"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

type DBConfig struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Encoding string `json:"encoding,omitempty"`
	ReadOnly bool   `json:"readOnly"`

	// SSH Tunnel Config
	SSHEnabled  bool   `json:"sshEnabled"`
	SSHHost     string `json:"sshHost"`
	SSHPort     int    `json:"sshPort"`
	SSHUser     string `json:"sshUser"`
	SSHPassword string `json:"sshPassword"`
	SSHKeyFile  string `json:"sshKeyFile"`
}

type Database struct {
	conn           *sql.DB
	persistentConn *sql.Conn
	sshClient      *ssh.Client
	sshListener    net.Listener
	logf           func(level string, message string)
	Type           string
	Encoding       string
	ReadOnly       bool
	Host           string
	Port           int
	User           string
	DatabaseName   string
	SSHEnabled     bool
	ConnectedAt    time.Time
}

const sshDialTimeout = 15 * time.Second
const defaultMySQLCharset = "utf8mb4"

func NewDatabase(logger ...func(level string, message string)) *Database {
	db := &Database{}
	if len(logger) > 0 {
		db.logf = logger[0]
	}
	return db
}

func (d *Database) log(level string, message string) {
	if d.logf != nil {
		d.logf(level, message)
	}
}

func isLocalDatabaseType(dbType string) bool {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "sqlite", "duckdb", "libsql":
		return true
	default:
		return false
	}
}

func existingKnownHostsFiles() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return nil
	}

	candidates := []string{
		filepath.Join(homeDir, ".ssh", "known_hosts"),
		filepath.Join(homeDir, ".ssh", "known_hosts2"),
	}

	files := make([]string, 0, len(candidates))
	for _, path := range candidates {
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() {
			files = append(files, path)
		}
	}

	return files
}

func buildSSHHostKeyCallback() (ssh.HostKeyCallback, error) {
	knownHostFiles := existingKnownHostsFiles()
	if len(knownHostFiles) == 0 {
		return nil, fmt.Errorf("no SSH known_hosts file found (expected ~/.ssh/known_hosts). Please add the SSH host key first")
	}

	baseCallback, err := knownhosts.New(knownHostFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts: %w", err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := baseCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			gotFingerprint := ssh.FingerprintSHA256(key)
			if len(keyErr.Want) == 0 {
				return fmt.Errorf("ssh host key for %s is not trusted (fingerprint: %s); add it to ~/.ssh/known_hosts", hostname, gotFingerprint)
			}
			return fmt.Errorf("ssh host key mismatch for %s (got: %s); check ~/.ssh/known_hosts", hostname, gotFingerprint)
		}

		return err
	}, nil
}

func proxyBidirectional(localConn net.Conn, remoteConn net.Conn) {
	var closeOnce sync.Once
	closeAll := func() {
		_ = localConn.Close()
		_ = remoteConn.Close()
	}

	relay := func(dst net.Conn, src net.Conn, done chan<- struct{}) {
		_, _ = io.Copy(dst, src)
		closeOnce.Do(closeAll)
		done <- struct{}{}
	}

	done := make(chan struct{}, 2)
	go relay(localConn, remoteConn, done)
	go relay(remoteConn, localConn, done)

	<-done
	<-done
}

func (d *Database) Connect(config DBConfig) error {
	if d.conn != nil {
		d.Disconnect()
	}
	d.Encoding = ""

	if isLocalDatabaseType(config.Type) {
		config.SSHEnabled = false
	}

	var dbHost string
	var dbPort int

	// Handle SSH Tunneling
	if config.SSHEnabled {
		if strings.TrimSpace(config.SSHHost) == "" {
			return fmt.Errorf("ssh host is required")
		}
		if config.SSHPort <= 0 || config.SSHPort > 65535 {
			return fmt.Errorf("ssh port must be between 1 and 65535")
		}
		if strings.TrimSpace(config.SSHUser) == "" {
			return fmt.Errorf("ssh user is required")
		}
		if strings.TrimSpace(config.SSHKeyFile) == "" && config.SSHPassword == "" {
			return fmt.Errorf("ssh password or ssh key file is required")
		}

		hostKeyCallback, err := buildSSHHostKeyCallback()
		if err != nil {
			return err
		}

		// 1. Setup SSH Client Config
		sshConfig := &ssh.ClientConfig{
			User:            config.SSHUser,
			HostKeyCallback: hostKeyCallback,
			Timeout:         sshDialTimeout,
		}

		if strings.TrimSpace(config.SSHKeyFile) != "" {
			key, err := os.ReadFile(config.SSHKeyFile)
			if err != nil {
				return fmt.Errorf("unable to read private key: %v", err)
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				return fmt.Errorf("unable to parse private key: %v", err)
			}
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		} else {
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.Password(config.SSHPassword),
			}
		}

		// 2. Connect to SSH Server
		sshAddr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
		client, err := ssh.Dial("tcp", sshAddr, sshConfig)
		if err != nil {
			return fmt.Errorf("failed to dial ssh: %w", err)
		}
		d.sshClient = client

		// 3. Setup Local Listener for Port Forwarding
		listener, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			client.Close()
			return fmt.Errorf("failed to start local listener: %w", err)
		}
		d.sshListener = listener

		// 4. Start Forwarding
		go func() {
			for {
				localConn, err := listener.Accept()
				if err != nil {
					d.log("INFO", fmt.Sprintf("SSH tunnel listener stopped: %v", err))
					return
				}

				go func(lc net.Conn) {
					remoteAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
					remoteConn, err := client.Dial("tcp", remoteAddr)
					if err != nil {
						d.log("ERROR", fmt.Sprintf("SSH tunnel dial error: %v", err))
						_ = lc.Close()
						return
					}

					proxyBidirectional(lc, remoteConn)
				}(localConn)
			}
		}()

		// 5. Update DB Connection Info to use Local Listener
		tcpAddr := listener.Addr().(*net.TCPAddr)
		dbHost = "localhost"
		dbPort = tcpAddr.Port

	} else {
		// Direct Connection
		dbHost = config.Host
		dbPort = config.Port
	}

	var dsn string
	var driverName string

	switch config.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		driverName = "pgx"
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.User, config.Password, dbHost, dbPort, config.Database)
	case "mysql", "mariadb", "databend":
		driverName = "mysql"
		mysqlCharset := normalizeMySQLCharset(config.Encoding)
		if mysqlCharset == "" {
			mysqlCharset = detectMySQLSchemaCharset(dbHost, dbPort, config)
		}
		if mysqlCharset == "" {
			mysqlCharset = defaultMySQLCharset
		}
		dsn = buildMySQLDSN(config.User, config.Password, dbHost, dbPort, config.Database, mysqlCharset)
		if strings.TrimSpace(config.Encoding) == "" {
			d.Encoding = mapMySQLCharsetToDecoder(mysqlCharset)
		}
	case "mssql":
		driverName = "sqlserver"
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable", config.User, config.Password, dbHost, dbPort, config.Database)
	case "sqlite", "libsql":
		driverName = "sqlite"
		dsn = config.Database // Path to DB file (local)
	case "duckdb":
		driverName = "duckdb"
		dsn = config.Database // Path to DB file (local)
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
		conn.Close() // Close SQL connection
		// If we opened an SSH tunnel, we should close it too since connection failed
		if d.sshListener != nil {
			d.sshListener.Close()
			d.sshListener = nil
		}
		if d.sshClient != nil {
			d.sshClient.Close()
			d.sshClient = nil
		}
		return err
	}

	// Acquire a dedicated connection for this session
	// This ensures that all queries executed by this Database instance share the same underlying connection,
	// preserving transaction state and session-level settings.
	persistentConn, err := conn.Conn(context.Background())
	if err != nil {
		conn.Close()
		if d.sshListener != nil {
			d.sshListener.Close()
			d.sshListener = nil
		}
		if d.sshClient != nil {
			d.sshClient.Close()
			d.sshClient = nil
		}
		return fmt.Errorf("failed to acquire dedicated connection: %w", err)
	}

	d.conn = conn
	d.persistentConn = persistentConn
	d.Type = config.Type
	if strings.TrimSpace(d.Encoding) == "" {
		d.Encoding = config.Encoding
	}
	d.ReadOnly = config.ReadOnly
	d.Host = dbHost
	d.Port = dbPort
	d.User = config.User
	d.DatabaseName = config.Database
	d.SSHEnabled = config.SSHEnabled
	d.ConnectedAt = time.Now()
	return nil
}

func buildMySQLDSN(user string, password string, host string, port int, database string, charset string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", user, password, host, port, database, charset)
}

func normalizeMySQLCharset(encodingName string) string {
	normalized := strings.ToLower(strings.TrimSpace(encodingName))
	switch normalized {
	case "":
		return ""
	case "auto":
		return ""
	case "utf8":
		return "utf8"
	case "utf8mb4":
		return "utf8mb4"
	case "tis620", "tis-620":
		return "tis620"
	case "cp874", "windows-874", "windows874":
		return "cp874"
	default:
		return normalized
	}
}

func mapMySQLCharsetToDecoder(charset string) string {
	switch normalizeMySQLCharset(charset) {
	case "tis620", "cp874":
		return "TIS-620"
	default:
		return ""
	}
}

func detectMySQLSchemaCharset(host string, port int, config DBConfig) string {
	dsn := buildMySQLDSN(config.User, config.Password, host, port, config.Database, defaultMySQLCharset)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return ""
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return ""
	}

	var schemaCharset string
	err = conn.QueryRowContext(
		ctx,
		"SELECT DEFAULT_CHARACTER_SET_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?",
		config.Database,
	).Scan(&schemaCharset)
	if err == nil {
		return normalizeMySQLCharset(schemaCharset)
	}

	var fallbackCharset string
	if err := conn.QueryRowContext(ctx, "SELECT @@character_set_database").Scan(&fallbackCharset); err != nil {
		return ""
	}
	return normalizeMySQLCharset(fallbackCharset)
}

func decodeValue(val []byte, encodingName string) string {
	if len(val) == 0 {
		return ""
	}

	encodingName = strings.ToUpper(encodingName)

	// If it's valid UTF-8, return as is (unless explicitly told otherwise)
	if utf8.Valid(val) {
		if encodingName == "" || encodingName == "UTF-8" {
			return string(val)
		}
	}

	// Dynamic detection if no encoding specified
	if encodingName == "" || encodingName == "AUTO" {
		// Check if it's likely Thai TIS-620
		// Thai characters in TIS-620/Windows-874 are 0xA1-0xFB
		thaiScore := 0
		for _, b := range val {
			if b >= 0xA1 && b <= 0xFB {
				thaiScore++
			}
		}
		if thaiScore > 0 {
			encodingName = "TIS-620"
			// fmt.Printf("Auto-detected Thai TIS-620 for data starting with %X\n", val[:min(len(val), 4)])
		} else {
			return string(val) // Fallback to raw string if no obvious encoding
		}
	}

	// Manual decoding for TIS-620/Windows-874 (Avoids some library inconsistencies)
	if encodingName == "TIS-620" || encodingName == "WINDOWS-874" {
		runes := make([]rune, 0, len(val))
		for _, b := range val {
			if b <= 0x7F {
				runes = append(runes, rune(b))
			} else if b >= 0xA1 && b <= 0xFB {
				runes = append(runes, rune(uint32(b)+0x0D60))
			} else {
				runes = append(runes, rune(b))
			}
		}
		return string(runes)
	}

	// Use ianaindex to look up the encoding by name for other encodings
	e, err := ianaindex.IANA.Encoding(encodingName)
	if err != nil || e == nil {
		return string(val)
	}

	decoded, _, err := transform.Bytes(e.NewDecoder(), val)
	if err != nil {
		return string(val)
	}
	return string(decoded)
}

// decodeStringValue handles string values returned by drivers like go-mssqldb
// which do their own internal code page → Unicode conversion.
//
// Problem: When UTF-8 data is stored in a varchar column with Thai collation
// (Windows-874), the driver interprets each raw byte through the Windows-874
// code page, producing wrong Unicode characters. For example, the UTF-8 bytes
// E2 89 A5 (≥) become โ U+0089 ฅ.
//
// Fix: Reverse the Windows-874 mapping to recover the original raw bytes,
// then check if those bytes form valid UTF-8. If yes, use the UTF-8
// interpretation. If no, the data is genuinely Windows-874 and we keep it.
func decodeStringValue(s string, _ string) string {
	if len(s) == 0 {
		return s
	}

	// Quick scan: does this string contain any Thai Unicode characters?
	// If not, no Windows-874 mis-mapping could have occurred.
	hasThai := false
	hasC1Control := false
	for _, r := range s {
		if r >= 0x0E01 && r <= 0x0E5B {
			hasThai = true
		}
		if r >= 0x0080 && r <= 0x009F {
			hasC1Control = true
		}
	}

	if !hasThai {
		return s // No Thai chars, nothing to fix
	}

	// Try to reverse the Windows-874 → Unicode mapping to recover raw bytes.
	rawBytes := make([]byte, 0, len(s))
	canReverse := true

	for _, r := range s {
		if r <= 0x7F {
			// ASCII: same in both
			rawBytes = append(rawBytes, byte(r))
		} else if r >= 0x0E01 && r <= 0x0E3A {
			// Thai consonants/vowels: reverse Windows-874 mapping
			rawBytes = append(rawBytes, byte(r-0x0D60))
		} else if r >= 0x0E3F && r <= 0x0E5B {
			// Thai currency/digits: reverse Windows-874 mapping
			rawBytes = append(rawBytes, byte(r-0x0D60))
		} else if r >= 0x0080 && r <= 0x009F {
			// C1 control chars: these are unmapped Windows-874 bytes passed through
			rawBytes = append(rawBytes, byte(r))
		} else if r == 0x20AC {
			rawBytes = append(rawBytes, 0x80) // € → 0x80
		} else if r == 0x2026 {
			rawBytes = append(rawBytes, 0x85) // … → 0x85
		} else if r == 0x2018 {
			rawBytes = append(rawBytes, 0x91) // ' → 0x91
		} else if r == 0x2019 {
			rawBytes = append(rawBytes, 0x92) // ' → 0x92
		} else if r == 0x201C {
			rawBytes = append(rawBytes, 0x93) // " → 0x93
		} else if r == 0x201D {
			rawBytes = append(rawBytes, 0x94) // " → 0x94
		} else if r == 0x2022 {
			rawBytes = append(rawBytes, 0x95) // • → 0x95
		} else if r == 0x2013 {
			rawBytes = append(rawBytes, 0x96) // – → 0x96
		} else if r == 0x2014 {
			rawBytes = append(rawBytes, 0x97) // — → 0x97
		} else if r == 0x2039 {
			rawBytes = append(rawBytes, 0x8B) // ‹ → 0x8B
		} else if r == 0x203A {
			rawBytes = append(rawBytes, 0x9B) // › → 0x9B
		} else {
			// Character not in Windows-874 range — can't reverse
			canReverse = false
			break
		}
	}

	if !canReverse {
		return s
	}

	// Check if the reversed bytes form valid UTF-8
	if utf8.Valid(rawBytes) {
		decoded := string(rawBytes)
		// Extra sanity: only use the UTF-8 interpretation if it differs
		// and if the string contained C1 control chars or looks like
		// multi-byte UTF-8 was broken apart
		if decoded != s && (hasC1Control || len(rawBytes) != len([]rune(s))) {
			return decoded
		}
	}

	return s
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

	// Close SSH resources
	if d.sshListener != nil {
		d.sshListener.Close()
		d.sshListener = nil
	}
	if d.sshClient != nil {
		d.sshClient.Close()
		d.sshClient = nil
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
	return pks, nil
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
type ColumnDefinition struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Nullable      bool        `json:"nullable"`
	DefaultValue  interface{} `json:"defaultValue"`
	PrimaryKey    bool        `json:"primaryKey"`
	AutoIncrement bool        `json:"autoIncrement"`
}

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
		d.persistentConn.QueryRowContext(context.Background(), "SELECT current_database()").Scan(&info.DBName)
		d.persistentConn.QueryRowContext(context.Background(), "SELECT version()").Scan(&info.Version)
	case "mysql", "mariadb", "databend":
		d.persistentConn.QueryRowContext(context.Background(), "SELECT DATABASE()").Scan(&info.DBName)
		d.persistentConn.QueryRowContext(context.Background(), "SELECT VERSION()").Scan(&info.Version)
	case "mssql":
		d.persistentConn.QueryRowContext(context.Background(), "SELECT DB_NAME()").Scan(&info.DBName)
		d.persistentConn.QueryRowContext(context.Background(), "SELECT @@VERSION").Scan(&info.Version)
	case "sqlite", "libsql":
		info.DBName = "main"
		d.persistentConn.QueryRowContext(context.Background(), "SELECT sqlite_version()").Scan(&info.Version)
	}
	info.Summary["activeDatabase"] = info.DBName
	info.RuntimeInfo["serverVersion"] = info.Version

	// Get Size
	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		d.persistentConn.QueryRowContext(context.Background(), "SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&info.Size)
	case "mysql", "mariadb", "databend":
		var sizeBytes int64
		d.persistentConn.QueryRowContext(context.Background(), "SELECT SUM(data_length + index_length) FROM information_schema.TABLES WHERE table_schema = DATABASE()").Scan(&sizeBytes)
		info.Size = formatSize(sizeBytes)
	case "mssql":
		// Simple approach for MSSQL
		var databaseSize float64
		d.persistentConn.QueryRowContext(context.Background(), "SELECT SUM(size) * 8 / 1024 FROM sys.master_files WHERE database_id = DB_ID()").Scan(&databaseSize)
		info.Size = fmt.Sprintf("%.2f MB", databaseSize)
	case "sqlite", "libsql":
		var pageCount, pageSize int64
		d.persistentConn.QueryRowContext(context.Background(), "PRAGMA page_count").Scan(&pageCount)
		d.persistentConn.QueryRowContext(context.Background(), "PRAGMA page_size").Scan(&pageSize)
		info.Size = formatSize(pageCount * pageSize)
	}

	// Counts
	tables, _ := d.GetTables()
	info.TableCount = len(tables)

	views, _ := d.GetViews()
	info.ViewCount = len(views)

	procs, _ := d.GetStoredProcedures()
	funcs, _ := d.GetFunctions()
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

func (d *Database) DropDatabase(dbName string) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	// Drop database command depends on DB type
	// For most, it's DROP DATABASE [name]
	// But we need to be careful with active connection
	var query string
	switch d.Type {
	case "mysql", "mariadb", "postgres", "mssql":
		query = fmt.Sprintf("DROP DATABASE %s", d.quoteIdentifier(dbName))
	case "sqlite", "libsql":
		return fmt.Errorf("DROP DATABASE is not supported for SQLite. Please delete the file manually.")
	default:
		return fmt.Errorf("DROP DATABASE not supported for %s", d.Type)
	}

	_, err := d.persistentConn.ExecContext(context.Background(), query)
	return err
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
		rows.Close()

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
					break
				}
				cols = append(cols, name)
			}
			infoRows.Close()

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
type TableChanges struct {
	RenameTable  string             `json:"renameTable"` // New name if renaming
	AddColumns   []ColumnDefinition `json:"addColumns"`
	DropColumns  []string           `json:"dropColumns"`
	AlterColumns []ColumnChange     `json:"alterColumns"`
	AddIndexes   []IndexDefinition  `json:"addIndexes"`
	DropIndexes  []string           `json:"dropIndexes"`
	AddFKs       []ForeignKey       `json:"addFKs"`
	DropFKs      []string           `json:"dropFKs"`
}

type ColumnChange struct {
	OldName       string           `json:"oldName"`
	NewDefinition ColumnDefinition `json:"newDefinition"`
}

func (d *Database) AlterTable(tableName string, changes TableChanges) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	// Transaction wrapper
	tx, err := d.persistentConn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var statements []string
	var errGen error

	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		statements, errGen = d.generatePostgresAlterStatements(tableName, changes)
	case "mysql", "mariadb", "databend":
		statements, errGen = d.generateMysqlAlterStatements(tableName, changes)
	case "mssql":
		statements, errGen = d.generateMssqlAlterStatements(tableName, changes)
	case "sqlite", "libsql":
		statements, errGen = d.generateSqliteAlterStatements(tableName, changes)
	default:
		return fmt.Errorf("unsupported database type")
	}

	if errGen != nil {
		return errGen
	}

	for _, stmt := range statements {
		_, err := tx.ExecContext(context.Background(), stmt)
		if err != nil {
			return fmt.Errorf("error executing '%s': %w", stmt, err)
		}
	}

	return tx.Commit()
}

func (d *Database) generatePostgresAlterStatements(tableName string, changes TableChanges) ([]string, error) {
	var stmts []string

	// Rename table
	if changes.RenameTable != "" && changes.RenameTable != tableName {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tableName, changes.RenameTable))
		tableName = changes.RenameTable // Use new name for subsequent changes
	}

	// Column changes
	for _, col := range changes.DropColumns {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, col))
	}

	for _, col := range changes.AddColumns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.AutoIncrement {
			upperType := strings.ToUpper(col.Type)
			switch {
			case strings.Contains(upperType, "BIGINT"):
				def = fmt.Sprintf("%s BIGSERIAL", col.Name)
			case strings.Contains(upperType, "SMALLINT"):
				def = fmt.Sprintf("%s SMALLSERIAL", col.Name)
			default:
				def = fmt.Sprintf("%s SERIAL", col.Name)
			}
		} else {
			if !col.Nullable {
				def += " NOT NULL"
			}
			if col.DefaultValue != nil {
				def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
			}
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, def))
	}

	for _, change := range changes.AlterColumns {
		// PostgreSQL often needs separate statements for type, nullability, default
		if change.OldName != change.NewDefinition.Name {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, change.OldName, change.NewDefinition.Name))
		}

		// For type change
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", tableName, change.NewDefinition.Name, change.NewDefinition.Type, change.NewDefinition.Name, change.NewDefinition.Type))

		// For Nullable
		if change.NewDefinition.Nullable {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", tableName, change.NewDefinition.Name))
		} else {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", tableName, change.NewDefinition.Name))
		}

		if change.NewDefinition.AutoIncrement {
			seqName := fmt.Sprintf("%s_%s_seq", tableName, change.NewDefinition.Name)
			stmts = append(stmts, fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s", seqName))
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT nextval('%s')", tableName, change.NewDefinition.Name, seqName))
			stmts = append(stmts, fmt.Sprintf("ALTER SEQUENCE %s OWNED BY %s.%s", seqName, tableName, change.NewDefinition.Name))
		}
	}

	// Index changes
	for _, idx := range changes.DropIndexes {
		stmts = append(stmts, fmt.Sprintf("DROP INDEX %s", idx))
	}
	for _, idx := range changes.AddIndexes {
		unique := ""
		if idx.Unique {
			unique = "UNIQUE"
		}
		cols := strings.Join(idx.Columns, ", ")
		stmts = append(stmts, fmt.Sprintf("CREATE %s INDEX %s ON %s (%s)", unique, idx.Name, tableName, cols))
	}

	// FK changes
	for _, fk := range changes.DropFKs {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName, fk))
	}
	for _, fk := range changes.AddFKs {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			tableName, fk.Constraint, fk.Column, fk.RefTable, fk.RefColumn))
	}

	return stmts, nil
}

func (d *Database) generateMysqlAlterStatements(tableName string, changes TableChanges) ([]string, error) {
	var stmts []string

	// Rename
	if changes.RenameTable != "" && changes.RenameTable != tableName {
		stmts = append(stmts, fmt.Sprintf("RENAME TABLE %s TO %s", tableName, changes.RenameTable))
		tableName = changes.RenameTable
	}

	// Helper for column definitions
	getDef := func(col ColumnDefinition) string {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if !col.Nullable {
			def += " NOT NULL"
		} else {
			def += " NULL"
		}
		if col.AutoIncrement {
			def += " AUTO_INCREMENT"
		}
		if col.DefaultValue != nil {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		return def
	}

	var alters []string

	for _, col := range changes.DropColumns {
		alters = append(alters, fmt.Sprintf("DROP COLUMN %s", col))
	}
	for _, col := range changes.AddColumns {
		alters = append(alters, fmt.Sprintf("ADD COLUMN %s", getDef(col)))
	}
	for _, change := range changes.AlterColumns {
		if change.OldName != change.NewDefinition.Name {
			alters = append(alters, fmt.Sprintf("CHANGE COLUMN %s %s", change.OldName, getDef(change.NewDefinition)))
		} else {
			alters = append(alters, fmt.Sprintf("MODIFY COLUMN %s", getDef(change.NewDefinition)))
		}
	}
	for _, idx := range changes.DropIndexes {
		if idx == "PRIMARY" {
			alters = append(alters, "DROP PRIMARY KEY")
		} else {
			alters = append(alters, fmt.Sprintf("DROP INDEX %s", idx))
		}
	}
	for _, idx := range changes.AddIndexes {
		if idx.Primary {
			alters = append(alters, fmt.Sprintf("ADD PRIMARY KEY (%s)", strings.Join(idx.Columns, ", ")))
		} else {
			unique := ""
			if idx.Unique {
				unique = "UNIQUE"
			}
			alters = append(alters, fmt.Sprintf("ADD %s INDEX %s (%s)", unique, idx.Name, strings.Join(idx.Columns, ", ")))
		}
	}
	for _, fk := range changes.DropFKs {
		alters = append(alters, fmt.Sprintf("DROP FOREIGN KEY %s", fk))
	}
	for _, fk := range changes.AddFKs {
		alters = append(alters, fmt.Sprintf("ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			fk.Constraint, fk.Column, fk.RefTable, fk.RefColumn))
	}

	if len(alters) > 0 {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s %s", tableName, strings.Join(alters, ", ")))
	}

	return stmts, nil
}

func (d *Database) generateMssqlAlterStatements(tableName string, changes TableChanges) ([]string, error) {
	var stmts []string

	if changes.RenameTable != "" && changes.RenameTable != tableName {
		stmts = append(stmts, fmt.Sprintf("EXEC sp_rename '%s', '%s'", tableName, changes.RenameTable))
		tableName = changes.RenameTable
	}

	for _, change := range changes.AlterColumns {
		if change.NewDefinition.AutoIncrement {
			return nil, fmt.Errorf("changing auto increment on existing MSSQL columns is not supported")
		}
	}

	// Collect columns that are being dropped or altered
	var affectedColumns []string
	affectedColumns = append(affectedColumns, changes.DropColumns...)
	for _, c := range changes.AlterColumns {
		affectedColumns = append(affectedColumns, c.OldName)
	}

	type mssqlIdxInfo struct {
		Name     string
		IsUnique bool
		IsPK     bool
		Columns  []string
	}
	droppedIndexes := make(map[string]mssqlIdxInfo)
	droppedDefaults := make(map[string]string) // name -> column

	if len(affectedColumns) > 0 {
		// Identify all dependent indexes/constraints for affected columns
		for _, colName := range affectedColumns {
			indexQuery := fmt.Sprintf(`
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
					AND i.index_id IN (
						SELECT index_id 
						FROM sys.index_columns ic2 
						JOIN sys.columns c2 ON ic2.object_id = c2.object_id AND ic2.column_id = c2.column_id
						WHERE ic2.object_id = OBJECT_ID('%s') AND c2.name = '%s'
					)
				ORDER BY 
					i.name, ic.index_column_id`, tableName, tableName, colName)

			rows, err := d.conn.QueryContext(context.Background(), indexQuery)
			if err == nil {
				for rows.Next() {
					var iName, cName string
					var isU, isP bool
					if err := rows.Scan(&iName, &cName, &isU, &isP); err == nil {
						if info, ok := droppedIndexes[iName]; ok {
							// Update existing entry if column not already there (order is important)
							found := false
							for _, c := range info.Columns {
								if c == cName {
									found = true
									break
								}
							}
							if !found {
								info.Columns = append(info.Columns, cName)
								droppedIndexes[iName] = info
							}
						} else {
							droppedIndexes[iName] = mssqlIdxInfo{
								Name:     iName,
								IsUnique: isU,
								IsPK:     isP,
								Columns:  []string{cName},
							}
						}
					}
				}
				rows.Close()
			}

			// Identify default constraints
			defaultQuery := fmt.Sprintf(`
				SELECT d.name
				FROM sys.default_constraints d
				INNER JOIN sys.columns c ON d.parent_object_id = c.object_id AND d.parent_column_id = c.column_id
				WHERE d.parent_object_id = OBJECT_ID('%s') AND c.name = '%s'`, tableName, colName)
			drows, err := d.conn.QueryContext(context.Background(), defaultQuery)
			if err == nil {
				for drows.Next() {
					var dName string
					if err := drows.Scan(&dName); err == nil {
						droppedDefaults[dName] = colName
					}
				}
				drows.Close()
			}
		}
	}

	// 1. Drop all identified dependencies
	for _, idx := range droppedIndexes {
		if idx.IsPK || idx.IsUnique {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName, idx.Name))
		} else {
			stmts = append(stmts, fmt.Sprintf("DROP INDEX %s ON %s", idx.Name, tableName))
		}
	}
	for dName := range droppedDefaults {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName, dName))
	}

	// 2. Perform Drops, Adds, and Alters
	for _, col := range changes.DropColumns {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, col))
	}
	for _, col := range changes.AddColumns {
		safeType := normalizeMssqlColumnType(col.Type)
		upperType := strings.ToUpper(safeType)
		if (strings.HasPrefix(upperType, "VARCHAR") || strings.HasPrefix(upperType, "NVARCHAR")) && !strings.Contains(safeType, "(") {
			safeType += "(255)"
		}

		def := fmt.Sprintf("%s %s", col.Name, safeType)
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.DefaultValue != nil {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD %s", tableName, def))
	}

	renameMap := make(map[string]string)
	for _, change := range changes.AlterColumns {
		if change.OldName != change.NewDefinition.Name {
			stmts = append(stmts, fmt.Sprintf("EXEC sp_rename '%s.%s', '%s', 'COLUMN'", tableName, change.OldName, change.NewDefinition.Name))
			renameMap[change.OldName] = change.NewDefinition.Name
		}

		safeType := normalizeMssqlColumnType(change.NewDefinition.Type)
		upperType := strings.ToUpper(safeType)
		if (strings.HasPrefix(upperType, "VARCHAR") || strings.HasPrefix(upperType, "NVARCHAR")) && !strings.Contains(safeType, "(") {
			safeType += "(255)"
		}

		def := fmt.Sprintf("%s", safeType)
		if !change.NewDefinition.Nullable {
			def += " NOT NULL"
		} else {
			def += " NULL"
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", tableName, change.NewDefinition.Name, def))

		// If it had a default or has a new one, we handle it below in recreation
	}

	// 3. Recreate Dependencies
	recreatedIndexNames := make(map[string]struct{})
	for _, idx := range droppedIndexes {
		// Skip if any of the columns were dropped
		skip := false
		for _, col := range idx.Columns {
			for _, dropped := range changes.DropColumns {
				if col == dropped {
					skip = true
					break
				}
			}
			if skip {
				break
			}
		}
		if skip {
			continue
		}

		newCols := make([]string, len(idx.Columns))
		for i, c := range idx.Columns {
			if nc, ok := renameMap[c]; ok {
				newCols[i] = nc
			} else {
				newCols[i] = c
			}
		}
		colsJoined := strings.Join(newCols, ", ")

		if idx.IsPK {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", tableName, idx.Name, colsJoined))
		} else if idx.IsUnique {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", tableName, idx.Name, colsJoined))
		} else {
			stmts = append(stmts, fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.Name, tableName, colsJoined))
		}
		recreatedIndexNames[idx.Name] = struct{}{}
	}

	// Restore or add default constraints
	for _, change := range changes.AlterColumns {
		if change.NewDefinition.DefaultValue != nil {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD DEFAULT '%v' FOR %s", tableName, change.NewDefinition.DefaultValue, change.NewDefinition.Name))
		}
	}

	// Additional index and FK drops/adds from Change request
	for _, idx := range changes.DropIndexes {
		// Only if not already dropped by dependency logic
		alreadyDropped := false
		for _, di := range droppedIndexes {
			if di.Name == idx {
				alreadyDropped = true
				break
			}
		}
		if !alreadyDropped {
			stmts = append(stmts, fmt.Sprintf("DROP INDEX %s ON %s", idx, tableName))
		}
	}
	for _, idx := range changes.AddIndexes {
		if _, wasRecreated := recreatedIndexNames[idx.Name]; wasRecreated {
			continue
		}
		cols := strings.Join(idx.Columns, ", ")
		if idx.Primary {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", tableName, idx.Name, cols))
		} else if idx.Unique {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)", tableName, idx.Name, cols))
		} else {
			stmts = append(stmts, fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.Name, tableName, cols))
		}
	}

	for _, fk := range changes.DropFKs {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName, fk))
	}
	for _, fk := range changes.AddFKs {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			tableName, fk.Constraint, fk.Column, fk.RefTable, fk.RefColumn))
	}

	return stmts, nil
}

func normalizeMssqlColumnType(typeName string) string {
	cleanType := strings.TrimSpace(typeName)
	if cleanType == "" {
		return cleanType
	}

	upperType := strings.ToUpper(cleanType)
	if strings.HasPrefix(upperType, "TIMESTAMP") {
		return "DATETIME2"
	}

	return cleanType
}

func (d *Database) generateSqliteAlterStatements(tableName string, changes TableChanges) ([]string, error) {
	// SQLite supports basic ALTER TABLE
	// RENAME TABLE, RENAME COLUMN, ADD COLUMN
	// It does NOT support DROP COLUMN, ALTER COLUMN type/nullability directly (requires recreation)

	// Complex check: if we have drop columns or alter column type/nullability, we MUST recreate.
	mustRecreate := len(changes.DropColumns) > 0 || len(changes.AlterColumns) > 0 || len(changes.DropFKs) > 0 || len(changes.AddFKs) > 0
	// Actually adding FKs might require recreation too if constraints are inline? ALTER TABLE ADD CONSTRAINT is NOT supported in SQLite.

	if mustRecreate {
		return nil, fmt.Errorf("complex table alterations (Modify/Drop Column, FKs) are not yet supported for SQLite in this version (requires table recreation)")
	}

	var stmts []string

	if changes.RenameTable != "" && changes.RenameTable != tableName {
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tableName, changes.RenameTable))
		tableName = changes.RenameTable
	}

	for _, col := range changes.AddColumns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		// SQLite ADD COLUMN has limitations (e.g. can't be UNIQUE or PRIMARY KEY usually directly without constraints?)
		// Actually basic ADD COLUMN is fine.
		if !col.Nullable {
			// Adding NOT NULL column to populated table requires DEFAULT
			if col.DefaultValue == nil {
				return nil, fmt.Errorf("cannot add NOT NULL column to SQLite table without DEFAULT value")
			}
			def += " NOT NULL"
		}
		if col.DefaultValue != nil {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, def))
	}

	// Index operations are separate and supported
	for _, idx := range changes.DropIndexes {
		stmts = append(stmts, fmt.Sprintf("DROP INDEX %s", idx))
	}
	for _, idx := range changes.AddIndexes {
		unique := ""
		if idx.Unique {
			unique = "UNIQUE"
		}
		cols := strings.Join(idx.Columns, ", ")
		stmts = append(stmts, fmt.Sprintf("CREATE %s INDEX %s ON %s (%s)", unique, idx.Name, tableName, cols))
	}

	return stmts, nil
}

func (d *Database) CreateTable(tableName string, columns []ColumnDefinition) error {
	if d.persistentConn == nil {
		return fmt.Errorf("no database connection")
	}

	if d.ReadOnly {
		return fmt.Errorf("database is in read-only mode")
	}

	var stmt string
	var errGen error

	switch d.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		stmt, errGen = d.generatePostgresCreateTableStatement(tableName, columns)
	case "mysql", "mariadb", "databend":
		stmt, errGen = d.generateMysqlCreateTableStatement(tableName, columns)
	case "mssql":
		stmt, errGen = d.generateMssqlCreateTableStatement(tableName, columns)
	case "sqlite", "libsql":
		stmt, errGen = d.generateSqliteCreateTableStatement(tableName, columns)
	default:
		return fmt.Errorf("unsupported database type")
	}

	if errGen != nil {
		return errGen
	}

	_, err := d.persistentConn.ExecContext(context.Background(), stmt)
	if err != nil {
		return fmt.Errorf("error executing '%s': %w", stmt, err)
	}

	return nil
}

func (d *Database) generatePostgresCreateTableStatement(tableName string, columns []ColumnDefinition) (string, error) {
	var cols []string
	var pks []string

	for _, col := range columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.AutoIncrement {
			// In Postgres, typically SERIAL is used for auto-increment.
			if strings.Contains(strings.ToLower(col.Type), "int") {
				def = fmt.Sprintf("%s SERIAL", col.Name)
			}
		}
		if !col.Nullable && !col.AutoIncrement {
			def += " NOT NULL"
		}
		if col.DefaultValue != nil {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		cols = append(cols, def)

		if col.PrimaryKey {
			pks = append(pks, col.Name)
		}
	}

	if len(pks) > 0 {
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pks, ", ")))
	}

	return fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", tableName, strings.Join(cols, ",\n  ")), nil
}

func (d *Database) generateMysqlCreateTableStatement(tableName string, columns []ColumnDefinition) (string, error) {
	var cols []string
	var pks []string

	for _, col := range columns {
		def := fmt.Sprintf("`%s` %s", col.Name, col.Type)
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.AutoIncrement {
			def += " AUTO_INCREMENT"
		}
		if col.DefaultValue != nil {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		cols = append(cols, def)

		if col.PrimaryKey {
			pks = append(pks, fmt.Sprintf("`%s`", col.Name))
		}
	}

	if len(pks) > 0 {
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pks, ", ")))
	}

	return fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n)", tableName, strings.Join(cols, ",\n  ")), nil
}

func (d *Database) generateMssqlCreateTableStatement(tableName string, columns []ColumnDefinition) (string, error) {
	var cols []string
	var pks []string

	for _, col := range columns {
		safeType := normalizeMssqlColumnType(col.Type)
		def := fmt.Sprintf("[%s] %s", col.Name, safeType)
		if col.AutoIncrement {
			def += " IDENTITY(1,1)"
		}
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.DefaultValue != nil && !col.AutoIncrement {
			def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
		}
		cols = append(cols, def)

		if col.PrimaryKey {
			pks = append(pks, fmt.Sprintf("[%s]", col.Name))
		}
	}

	if len(pks) > 0 {
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pks, ", ")))
	}

	return fmt.Sprintf("CREATE TABLE [%s] (\n  %s\n)", tableName, strings.Join(cols, ",\n  ")), nil
}

func (d *Database) generateSqliteCreateTableStatement(tableName string, columns []ColumnDefinition) (string, error) {
	var cols []string
	var pks []string

	for _, col := range columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.PrimaryKey {
			if col.AutoIncrement {
				def = fmt.Sprintf("%s INTEGER PRIMARY KEY AUTOINCREMENT", col.Name)
			} else {
				pks = append(pks, col.Name)
				if !col.Nullable {
					def += " NOT NULL"
				}
				if col.DefaultValue != nil {
					def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
				}
			}
		} else {
			if !col.Nullable {
				def += " NOT NULL"
			}
			if col.DefaultValue != nil {
				def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
			}
			cols = append(cols, def)
			// Wait, if it IS NOT a primary key, we should add it to cols here.
			// But if it IS a primary key, we also need to add it to cols.
			// The original logic missed adding to cols if it was a primary key. Let me fix it.
		}
	}

	var correctedCols []string
	var correctedPks []string
	for _, col := range columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.PrimaryKey {
			if col.AutoIncrement {
				def = fmt.Sprintf("%s INTEGER PRIMARY KEY AUTOINCREMENT", col.Name)
			} else {
				correctedPks = append(correctedPks, col.Name)
				if !col.Nullable {
					def += " NOT NULL"
				}
				if col.DefaultValue != nil {
					def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
				}
			}
		} else {
			if !col.Nullable {
				def += " NOT NULL"
			}
			if col.DefaultValue != nil {
				def += fmt.Sprintf(" DEFAULT '%v'", col.DefaultValue)
			}
		}
		correctedCols = append(correctedCols, def)
	}

	if len(correctedPks) > 0 {
		correctedCols = append(correctedCols, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(correctedPks, ", ")))
	}

	return fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", tableName, strings.Join(correctedCols, ",\n  ")), nil
}
