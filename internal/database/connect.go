package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"unicode/utf8"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"

	mysql "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	// _ "github.com/marcboeker/go-duckdb"
	_ "github.com/microsoft/go-mssqldb"
	_ "modernc.org/sqlite"
)

const defaultMySQLCharset = "utf8mb4"

func (d *Database) Connect(config DBConfig) error {
	if d.conn != nil {
		d.Disconnect()
	}
	d.Encoding = ""
	config.Type = normalizeDatabaseTypeAlias(config.Type)

	if isLocalDatabaseType(config.Type) {
		config.SSHEnabled = false
	}

	var dbHost string
	var dbPort int

	if config.SSHEnabled {
		host, port, err := d.openSSHTunnel(config)
		if err != nil {
			return err
		}
		dbHost = host
		dbPort = port
	} else {
		dbHost = config.Host
		dbPort = config.Port
	}

	var dsn string
	var driverName string

	switch config.Type {
	case "postgres", "greenplum", "redshift", "cockroachdb":
		driverName = "pgx"
		dsn = buildPostgresDSN(config, dbHost, dbPort)
	case "mysql", "mariadb", "databend":
		driverName = "mysql"
		mysqlCharset := normalizeMySQLCharset(config.Encoding)
		if mysqlCharset == "" {
			mysqlCharset = detectMySQLSchemaCharset(dbHost, dbPort, config)
		}
		if mysqlCharset == "" {
			mysqlCharset = defaultMySQLCharset
		}
		dsn = buildMySQLDSN(config, dbHost, dbPort, mysqlCharset)
		if strings.TrimSpace(config.Encoding) == "" {
			d.Encoding = mapMySQLCharsetToDecoder(mysqlCharset)
		}
	case "mssql":
		driverName = "sqlserver"
		dsn = buildSQLServerDSN(config, dbHost, dbPort)
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

func buildPostgresDSN(config DBConfig, host string, port int) string {
	dsnURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.User, config.Password),
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   "/" + config.Database,
	}

	query := url.Values{}
	if shouldRequireEncryptedConnection(host, config.SSHEnabled) {
		query.Set("sslmode", "require")
	} else {
		query.Set("sslmode", "disable")
	}
	dsnURL.RawQuery = query.Encode()

	return dsnURL.String()
}

func normalizeDatabaseTypeAlias(dbType string) string {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "supabase":
		return "postgres"
	default:
		return strings.ToLower(strings.TrimSpace(dbType))
	}
}

func buildMySQLDSN(config DBConfig, host string, port int, charset string) string {
	dsnConfig := mysql.NewConfig()
	dsnConfig.User = config.User
	dsnConfig.Passwd = config.Password
	dsnConfig.Net = "tcp"
	dsnConfig.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	dsnConfig.DBName = config.Database
	dsnConfig.Params = map[string]string{
		"charset": charset,
	}
	dsnConfig.ParseTime = true
	dsnConfig.Loc = time.Local

	if shouldRequireEncryptedConnection(host, config.SSHEnabled) {
		dsnConfig.TLSConfig = "true"
	}

	return dsnConfig.FormatDSN()
}

func buildSQLServerDSN(config DBConfig, host string, port int) string {
	dsnURL := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(config.User, config.Password),
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}

	query := url.Values{}
	query.Set("database", config.Database)
	// Preserve legacy SQL Server behavior for compatibility with servers that
	// reject TLS negotiation during prelogin.
	query.Set("encrypt", "disable")
	dsnURL.RawQuery = query.Encode()

	return dsnURL.String()
}

func shouldRequireEncryptedConnection(host string, sshEnabled bool) bool {
	if sshEnabled {
		return false
	}

	return !isLoopbackHost(host)
}

func isLoopbackHost(host string) bool {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return false
	}

	unwrappedHost := strings.Trim(trimmedHost, "[]")
	if strings.EqualFold(unwrappedHost, "localhost") {
		return true
	}

	ip := net.ParseIP(unwrappedHost)
	return ip != nil && ip.IsLoopback()
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
	dsn := buildMySQLDSN(config, host, port, defaultMySQLCharset)
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
