package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx context.Context
	dbs map[string]*Database
	mu  sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		dbs: make(map[string]*Database),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ConnectResult struct to return both ID and success status
type ConnectResult struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

func (a *App) ConnectDB(config DBConfig) ConnectResult {
	newDB := NewDatabase()
	err := newDB.Connect(config)
	if err != nil {
		return ConnectResult{Error: fmt.Sprintf("Error: %s", err.Error())}
	}

	id := uuid.New().String()

	a.mu.Lock()
	a.dbs[id] = newDB
	a.mu.Unlock()

	return ConnectResult{ID: id}
}

func (a *App) DisconnectDB(connectionID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if db, ok := a.dbs[connectionID]; ok {
		err := db.Disconnect()
		delete(a.dbs, connectionID)
		if err != nil {
			return fmt.Sprintf("Error: %s", err.Error())
		}
		return "Success"
	}
	return "Connection not found"
}

func (a *App) GetTables(connectionID string) []string {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return []string{}
	}

	tables, err := db.GetTables()
	if err != nil {
		return []string{}
	}
	return tables
}

// Result struct to return both data and error message if any
type QueryResult struct {
	Data    []map[string]interface{} `json:"data"`
	Columns []string                 `json:"columns"`
	Error   string                   `json:"error"`
}

func (a *App) ExecuteQuery(connectionID string, query string) QueryResult {
	a.mu.Lock()
	db, ok := a.dbs[connectionID]
	a.mu.Unlock()

	if !ok {
		return QueryResult{Error: "Connection not found"}
	}

	data, columns, err := db.ExecuteQuery(query)
	if err != nil {
		return QueryResult{Error: err.Error()}
	}
	return QueryResult{Data: data, Columns: columns}
}
