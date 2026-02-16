package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
	db  *Database
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		db: NewDatabase(),
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

func (a *App) ConnectDB(config DBConfig) string {
	err := a.db.Connect(config)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) DisconnectDB() string {
	err := a.db.Disconnect()
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}
	return "Success"
}

func (a *App) GetTables() []string {
	tables, err := a.db.GetTables()
	if err != nil {
		return []string{}
	}
	return tables
}

// Result struct to return both data and error message if any
type QueryResult struct {
	Data  []map[string]interface{} `json:"data"`
	Error string                   `json:"error"`
}

func (a *App) ExecuteQuery(query string) QueryResult {
	data, err := a.db.ExecuteQuery(query)
	if err != nil {
		return QueryResult{Error: err.Error()}
	}
	return QueryResult{Data: data}
}
