package app

import "fmt"

func (a *App) SaveQueryHistory(query string, dbType string, historyEnabled bool, retentionDays int) ActionResult {
	if a.localDB == nil {
		return ActionResult{Success: false, Error: "LocalDB not initialized"}
	}
	if !historyEnabled {
		return ActionResult{Success: true}
	}
	err := a.localDB.SaveQuery(query, dbType, retentionDays)
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}
	return ActionResult{Success: true}
}

func (a *App) GetQueryHistory(dbType string) []QueryHistoryEntry {
	if a.localDB == nil {
		return []QueryHistoryEntry{}
	}
	entries, err := a.localDB.GetQueries(dbType)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to get query history: %v", err))
		return []QueryHistoryEntry{}
	}
	return entries
}

func (a *App) SearchQueryHistory(queryText string, dbType string, favoritesOnly bool, dateRange string, sortMode string, limit int) []QueryHistoryEntry {
	if a.localDB == nil {
		return []QueryHistoryEntry{}
	}
	entries, err := a.localDB.SearchQueries(queryText, dbType, favoritesOnly, dateRange, sortMode, limit)
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to search query history: %v", err))
		return []QueryHistoryEntry{}
	}
	return entries
}

func (a *App) GetQueryHistorySummary() QueryHistorySummary {
	if a.localDB == nil {
		return QueryHistorySummary{DBTypes: []string{}}
	}

	summary, err := a.localDB.GetQueryHistorySummary()
	if err != nil {
		a.logEvent("ERROR", fmt.Sprintf("Failed to get query history summary: %v", err))
		return QueryHistorySummary{DBTypes: []string{}}
	}
	return summary
}

func (a *App) ToggleFavoriteQuery(id int, isFavorite bool) ActionResult {
	if a.localDB == nil {
		return ActionResult{Success: false, Error: "LocalDB not initialized"}
	}
	err := a.localDB.ToggleFavorite(id, isFavorite)
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}
	return ActionResult{Success: true}
}

func (a *App) DeleteQueryHistory(id int) ActionResult {
	if a.localDB == nil {
		return ActionResult{Success: false, Error: "LocalDB not initialized"}
	}
	err := a.localDB.DeleteQuery(id)
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}
	return ActionResult{Success: true}
}

func (a *App) ClearQueryHistory() ActionResult {
	if a.localDB == nil {
		return ActionResult{Success: false, Error: "LocalDB not initialized"}
	}
	err := a.localDB.ClearNonFavoriteQueries()
	if err != nil {
		return ActionResult{Success: false, Error: err.Error()}
	}
	return ActionResult{Success: true}
}
