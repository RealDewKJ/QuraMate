package database

import (
	"context"
	"fmt"
	"strings"
)

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
