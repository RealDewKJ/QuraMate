package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const sqlNotebookTestKey = "dashboard_sql_notebooks:mssql:wms_site_as"

func sqlNotebookTestPayload(t *testing.T) string {
	t.Helper()

	payload := sqlNotebookWorkspaceState{
		Version:          1,
		ActiveNotebookID: "notebook_mmub6zjm_y0r2bjb5",
		Notebooks: []sqlNotebookRecord{
			{
				ID:          "notebook_mmub6zjm_y0r2bjb5",
				Title:       "รายละเอียดการคำนวณ Color Visual",
				Description: "Capture reusable SQL workflows, notes, and operational context.",
				Tags:        []string{},
				ConnectionScope: sqlNotebookScope{
					DBType:         "mssql",
					ConnectionName: "WMS Site AS",
				},
				Cells: []sqlNotebookCell{
					{
						ID:    "cell_mmub6zjm_sppw6pqn",
						Type:  "markdown",
						Title: "Notes",
						Content: "# รายละเอียดการคำนวณ Color Visual (MasterBatch)\n" +
							"### 1. รูปแบบข้อมูล\n" +
							"ในตารางแสดงผล (Index) Color Visual จะแสดงสถานะเป็น Pass หรือ Fail โดยอ้างอิงจากตัวเลขที่ผู้ใช้กรอกในหน้า Modal\n\n" +
							"### 2. เงื่อนไขการแสดงผล (Logic)\n" +
							"ระบบใช้เงื่อนไขแบบตายตัว (Hardcoded) ดังนี้:\n" +
							"ค่าที่กรอก (Input) -> 1 = Pass, ค่าอื่น = Fail",
						Collapsed:      false,
						ExecutionState: "idle",
					},
				},
				CreatedAt:    "2026-03-17T07:46:48.560Z",
				UpdatedAt:    "2026-03-17T07:48:05.304Z",
				LastOpenedAt: "2026-03-17T07:48:05.304Z",
			},
		},
	}

	serialized, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal notebook payload: %v", err)
	}

	return string(serialized)
}

func newTestLocalDB(t *testing.T) *LocalDB {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "quramate.db")
	db, err := newLocalDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("create local db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.conn.Close()
		_ = os.Remove(dbPath)
	})

	return db
}

func TestSaveSettingPersistsSQLNotebooksInTables(t *testing.T) {
	t.Parallel()

	db := newTestLocalDB(t)
	if err := db.SaveSetting(sqlNotebookTestKey, sqlNotebookTestPayload(t)); err != nil {
		t.Fatalf("save notebook payload: %v", err)
	}

	var workspaceCount int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM sql_notebook_workspaces WHERE storage_key = ?`, sqlNotebookTestKey).Scan(&workspaceCount); err != nil {
		t.Fatalf("count notebook workspaces: %v", err)
	}
	if workspaceCount != 1 {
		t.Fatalf("expected 1 notebook workspace, got %d", workspaceCount)
	}

	var notebookCount int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM sql_notebooks WHERE storage_key = ?`, sqlNotebookTestKey).Scan(&notebookCount); err != nil {
		t.Fatalf("count notebooks: %v", err)
	}
	if notebookCount != 1 {
		t.Fatalf("expected 1 notebook row, got %d", notebookCount)
	}

	var cellCount int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM sql_notebook_cells`).Scan(&cellCount); err != nil {
		t.Fatalf("count notebook cells: %v", err)
	}
	if cellCount != 1 {
		t.Fatalf("expected 1 notebook cell row, got %d", cellCount)
	}

	var legacySettingsCount int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM settings WHERE key = ?`, sqlNotebookTestKey).Scan(&legacySettingsCount); err != nil {
		t.Fatalf("count legacy notebook settings: %v", err)
	}
	if legacySettingsCount != 0 {
		t.Fatalf("expected migrated notebook to not remain in settings, got %d rows", legacySettingsCount)
	}
}

func TestLoadSettingMigratesLegacyNotebookJSON(t *testing.T) {
	t.Parallel()

	db := newTestLocalDB(t)
	if _, err := db.conn.Exec(`INSERT INTO settings (key, value) VALUES (?, ?)`, sqlNotebookTestKey, sqlNotebookTestPayload(t)); err != nil {
		t.Fatalf("insert legacy notebook setting: %v", err)
	}

	loaded, err := db.LoadSetting(sqlNotebookTestKey)
	if err != nil {
		t.Fatalf("load migrated notebook setting: %v", err)
	}

	var state sqlNotebookWorkspaceState
	if err := json.Unmarshal([]byte(loaded), &state); err != nil {
		t.Fatalf("unmarshal loaded notebook state: %v", err)
	}
	if state.ActiveNotebookID != "notebook_mmub6zjm_y0r2bjb5" {
		t.Fatalf("expected active notebook id to migrate, got %q", state.ActiveNotebookID)
	}
	if len(state.Notebooks) != 1 || len(state.Notebooks[0].Cells) != 1 {
		t.Fatalf("expected notebook and cell to migrate, got notebooks=%d cells=%d", len(state.Notebooks), len(state.Notebooks[0].Cells))
	}

	var workspaceCount int
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM sql_notebook_workspaces WHERE storage_key = ?`, sqlNotebookTestKey).Scan(&workspaceCount); err != nil {
		t.Fatalf("count migrated workspaces: %v", err)
	}
	if workspaceCount != 1 {
		t.Fatalf("expected migrated workspace row, got %d", workspaceCount)
	}
}
