package app

func (a *App) GetCurrentVersion() string {
	return a.updater.GetCurrentVersion()
}

func (a *App) CheckForUpdates() UpdateInfo {
	return a.updater.CheckForUpdates()
}

func (a *App) OpenDownloadURL(url string) string {
	return a.updater.OpenDownloadURL(url)
}

func (a *App) PerformUpdate(downloadURL string) error {
	return a.updater.PerformUpdate(downloadURL)
}

func (a *App) emitUpdateProgress(stage string, percent int, message string) {
	if a.ctx == nil {
		return
	}

	a.updater.EmitProgress(a.ctx, stage, percent, message)
}
