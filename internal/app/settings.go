package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

func (a *App) SaveSetting(key string, value string) string {
	if a.localDB == nil {
		return "Error: LocalDB is not initialized"
	}
	if key == "user_settings" {
		value = sanitizeUserSettingsForStorage(value)
	}
	err := a.localDB.SaveSetting(key, value)
	if err != nil {
		return fmt.Sprintf("Error saving setting: %s", err.Error())
	}
	return "Success"
}

func (a *App) LoadSetting(key string) string {
	if a.localDB == nil {
		return ""
	}
	value, err := a.localDB.LoadSetting(key)
	if err != nil {
		fmt.Printf("Error loading setting %s: %v\n", key, err)
		return ""
	}
	return value
}

func (a *App) GetLocalDataEncryptionEnabled() bool {
	if a.localDB == nil {
		return false
	}
	return a.localDB.IsEncryptionEnabled()
}

func (a *App) SetLocalDataEncryptionEnabled(enabled bool) string {
	if a.localDB == nil {
		return "Error: LocalDB is not initialized"
	}
	if err := a.localDB.SetEncryptionEnabled(enabled); err != nil {
		return fmt.Sprintf("Error updating local data encryption: %s", err.Error())
	}
	return "Success"
}

func (a *App) SaveAIProviderKey(provider string, apiKey string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" {
		if err := keyring.Delete(aiKeyringService, normalizedProvider); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return fmt.Sprintf("Error deleting key: %s", err.Error())
		}
		return "Success"
	}

	if err := keyring.Set(aiKeyringService, normalizedProvider, trimmedKey); err != nil {
		return fmt.Sprintf("Error saving key: %s", err.Error())
	}
	return "Success"
}

func (a *App) LoadAIProviderKey(provider string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return ""
	}

	value, err := keyring.Get(aiKeyringService, normalizedProvider)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return ""
		}
		fmt.Printf("Error loading AI provider key for %s: %v\n", normalizedProvider, err)
		return ""
	}
	return value
}

func (a *App) DeleteAIProviderKey(provider string) string {
	normalizedProvider, err := normalizeAIProvider(provider)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	if err := keyring.Delete(aiKeyringService, normalizedProvider); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Sprintf("Error deleting key: %s", err.Error())
	}

	return "Success"
}
