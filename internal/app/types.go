package app

import (
	"QuraMate/internal/database"
	"QuraMate/internal/storage"
	updatepkg "QuraMate/internal/updater"
)

type DBConfig = database.DBConfig
type Database = database.Database
type ResultSet = database.ResultSet
type ServerProcess = database.ServerProcess
type ColumnMetadata = database.ColumnMetadata
type StreamBatch = database.StreamBatch
type DatabaseInfo = database.DatabaseInfo
type ColumnDefinition = database.ColumnDefinition
type IndexDefinition = database.IndexDefinition
type TableChanges = database.TableChanges
type ForeignKey = database.ForeignKey
type QueryHistoryEntry = storage.QueryHistoryEntry
type QueryHistorySummary = storage.QueryHistorySummary
type UpdateInfo = updatepkg.UpdateInfo

type TableExportOptions struct {
	ConnectionID  string `json:"connectionId"`
	TableName     string `json:"tableName"`
	Format        string `json:"format"`
	FilePath      string `json:"filePath"`
	IncludeSchema bool   `json:"includeSchema"`
	IncludeData   bool   `json:"includeData"`
	DropIfExists  bool   `json:"dropIfExists"`
}
