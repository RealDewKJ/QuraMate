package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Package database contains connection lifecycle, query execution,
// schema inspection, and database metadata helpers.

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

func closeRowsWithError(rows *sql.Rows) error {
	if err := rows.Err(); err != nil {
		return err
	}
	return rows.Close()
}

func isLocalDatabaseType(dbType string) bool {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "sqlite", "duckdb", "libsql":
		return true
	default:
		return false
	}
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
