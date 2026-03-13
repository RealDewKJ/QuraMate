package app

import (
	"QuraMate/internal/storage"
	updatepkg "QuraMate/internal/updater"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx              context.Context
	dbs              map[string]*Database
	mu               sync.Mutex
	approvedFileMu   sync.Mutex
	approvedRead     map[string]struct{}
	approvedWrite    map[string]struct{}
	queryCancelFuncs map[string]context.CancelFunc
	muQueries        sync.Mutex
	appLogs          []LogEntry
	muLogs           sync.Mutex
	localDB          *storage.LocalDB
	updater          *updatepkg.Service
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

const aiKeyringService = "QuraMate-AI"

const sshDialTimeout = 15 * time.Second

func NewApp() *App {
	ldb, err := storage.NewLocalDB()
	if err != nil {
		fmt.Printf("Failed to initialize LocalDB: %v\n", err)
	}

	return &App{
		dbs:              make(map[string]*Database),
		approvedRead:     make(map[string]struct{}),
		approvedWrite:    make(map[string]struct{}),
		queryCancelFuncs: make(map[string]context.CancelFunc),
		localDB:          ldb,
		updater:          updatepkg.NewService(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logEvent("INFO", "Application started")

	// Auto-check for updates after a short delay
	go func() {
		time.Sleep(3 * time.Second)
		info := a.updater.CheckForUpdates()
		if info.Available {
			runtime.EventsEmit(a.ctx, "app:update-available", info)
		}
	}()
}

// Greet returns a greeting for the given name

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ConnectResult struct to return both ID and success status

func (a *App) logEvent(level string, msg string) {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()

	entry := LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   level,
		Message: msg,
	}

	a.appLogs = append(a.appLogs, entry)

	// Keep only last 1000 logs
	if len(a.appLogs) > 1000 {
		a.appLogs = a.appLogs[len(a.appLogs)-1000:]
	}
}

func (a *App) GetAppLogs() []LogEntry {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()
	return a.appLogs
}

func (a *App) ClearAppLogs() string {
	a.muLogs.Lock()
	defer a.muLogs.Unlock()
	a.appLogs = []LogEntry{}
	return "Success"
}

func (a *App) LogClientEvent(level string, message string) string {
	cleanLevel := strings.ToUpper(strings.TrimSpace(level))
	if cleanLevel == "" {
		cleanLevel = "INFO"
	}

	cleanMessage := strings.TrimSpace(message)
	if cleanMessage == "" {
		return "Message is required"
	}
	if len(cleanMessage) > 2000 {
		cleanMessage = cleanMessage[:2000]
	}

	a.logEvent(cleanLevel, cleanMessage)
	return "Success"
}

func (a *App) debugLog(msg string) {
	appDir, err := getAppSupportDir()
	if err != nil {
		return
	}
	logFile := filepath.Join(appDir, "debug_open.log")
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}

func (a *App) GetStartupFile() string {
	a.debugLog(fmt.Sprintf("GetStartupFile called, argCount=%d", len(os.Args)))
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if _, err := os.Stat(arg); err == nil {
			a.debugLog(fmt.Sprintf("GetStartupFile returning file=%s", redactFilePathForLog(arg)))
			return arg
		}
		a.debugLog(fmt.Sprintf("GetStartupFile arg not a file: %s", redactFilePathForLog(arg)))
	}
	a.debugLog("GetStartupFile returning empty")
	return ""
}

func (a *App) OnSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	a.debugLog(fmt.Sprintf("OnSecondInstanceLaunch called, argCount=%d", len(secondInstanceData.Args)))

	// Bring window to front
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	a.debugLog("Window brought to front")

	if len(secondInstanceData.Args) > 0 {
		for _, arg := range secondInstanceData.Args {
			a.debugLog(fmt.Sprintf("Checking arg: %s", redactFilePathForLog(arg)))
			if _, err := os.Stat(arg); err == nil {
				// Write to pending file for the running instance to pick up
				pendingPath := a.getPendingFilePath()
				writeErr := os.WriteFile(pendingPath, []byte(arg), 0600)
				a.debugLog(fmt.Sprintf("Wrote pending file for %s (err=%v)", redactFilePathForLog(arg), writeErr))
				// Also emit event so frontend can react immediately (no polling needed)
				runtime.EventsEmit(a.ctx, "app:open-file", arg)
				break
			}
		}
	} else {
		a.debugLog("No file arg in second instance args")
	}
}

func (a *App) getPendingFilePath() string {
	appDir, err := getAppSupportDir()
	if err != nil {
		return filepath.Join(".", "pending_open.txt")
	}
	return filepath.Join(appDir, "pending_open.txt")
}

func (a *App) CheckPendingFile() string {
	pendingPath := a.getPendingFilePath()
	content, err := os.ReadFile(pendingPath)
	if err != nil {
		return ""
	}
	// Delete immediately after reading
	os.Remove(pendingPath)
	filePath := strings.TrimSpace(string(content))
	if filePath == "" {
		a.debugLog("CheckPendingFile: file was empty")
		return ""
	}
	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		a.debugLog(fmt.Sprintf("CheckPendingFile: file not found: %s", redactFilePathForLog(filePath)))
		return ""
	}
	a.debugLog(fmt.Sprintf("CheckPendingFile: returning %s", redactFilePathForLog(filePath)))
	return filePath
}

func (a *App) ReadTextFile(path string) (string, error) {
	if err := a.ensureApprovedFilePath(path, false); err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Detect UTF-16 LE BOM (FF FE) or BE BOM (FE FF)
	if len(content) >= 2 {
		if content[0] == 0xFF && content[1] == 0xFE {
			// UTF-16 LE
			return decodeUTF16(content[2:], false), nil
		} else if content[0] == 0xFE && content[1] == 0xFF {
			// UTF-16 BE
			return decodeUTF16(content[2:], true), nil
		}
	}

	// Also check if there's an overwhelming amount of NULL bytes (UTF-16 without BOM)
	// Just sample the first few bytes. Wait, safer to just return as string for UTF-8 normally.
	// We'll strip UTF-8 BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return string(content[3:]), nil
	}

	return string(content), nil
}

func (a *App) WriteTextFile(path string, content string) string {
	if strings.TrimSpace(path) == "" {
		return "file path is required"
	}
	if err := a.ensureApprovedFilePath(path, true); err != nil {
		return err.Error()
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return err.Error()
	}

	return ""
}

func (a *App) approveReadPath(path string) {
	a.approveFilePath(path, false)
}

func (a *App) approveWritePath(path string) {
	a.approveFilePath(path, true)
}

func (a *App) approveFilePath(path string, write bool) {
	normalizedPath, err := normalizeApprovedFilePath(path)
	if err != nil {
		return
	}

	a.approvedFileMu.Lock()
	defer a.approvedFileMu.Unlock()

	if write {
		a.approvedWrite[normalizedPath] = struct{}{}
		return
	}

	a.approvedRead[normalizedPath] = struct{}{}
}

func (a *App) ensureApprovedFilePath(path string, write bool) error {
	normalizedPath, err := normalizeApprovedFilePath(path)
	if err != nil {
		return err
	}

	a.approvedFileMu.Lock()
	defer a.approvedFileMu.Unlock()

	if write {
		if _, ok := a.approvedWrite[normalizedPath]; ok {
			return nil
		}
		return fmt.Errorf("file path is not approved for writing; choose the file through the save dialog first")
	}

	if _, ok := a.approvedRead[normalizedPath]; ok {
		return nil
	}
	return fmt.Errorf("file path is not approved for reading; choose the file through the open dialog first")
}

func normalizeApprovedFilePath(path string) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "", fmt.Errorf("file path is required")
	}

	absolutePath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	return filepath.Clean(absolutePath), nil
}

func getAppSupportDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(configDir) == "" {
		fallbackDir, fallbackErr := os.Getwd()
		if fallbackErr != nil {
			return "", fmt.Errorf("unable to determine app support directory")
		}
		configDir = fallbackDir
	}

	appDir := filepath.Join(configDir, "QuraMate")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", fmt.Errorf("unable to create app support directory: %w", err)
	}

	return appDir, nil
}

func redactFilePathForLog(path string) string {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return "<empty>"
	}

	baseName := filepath.Base(trimmedPath)
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		return "<path>"
	}

	return baseName
}

func decodeUTF16(b []byte, isBE bool) string {
	if len(b)%2 != 0 {
		// Just append a null byte to make it even
		b = append(b, 0)
	}
	u16s := make([]uint16, len(b)/2)
	for i := 0; i < len(u16s); i++ {
		if isBE {
			u16s[i] = uint16(b[i*2])<<8 | uint16(b[i*2+1])
		} else {
			u16s[i] = uint16(b[i*2+1])<<8 | uint16(b[i*2])
		}
	}
	// Convert array of uint16 to string by casting it to a slice of runes
	runes := make([]rune, len(u16s))
	for i, v := range u16s {
		runes[i] = rune(v)
	}
	return string(runes)
}

// ==== Settings Wails Bindings ====

func normalizeAIProvider(provider string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return "", fmt.Errorf("provider is required")
	}

	for _, ch := range normalized {
		isAlphaNum := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if !isAlphaNum && ch != '-' && ch != '_' {
			return "", fmt.Errorf("provider contains invalid characters")
		}
	}

	return normalized, nil
}

func sanitizeUserSettingsForStorage(value string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		// Keep original payload if it isn't JSON (backward compatibility).
		return value
	}

	if aiRaw, ok := parsed["ai"]; ok {
		if aiMap, ok := aiRaw.(map[string]any); ok {
			delete(aiMap, "apiKey")
			delete(aiMap, "apiKeys")
		}
	}

	sanitized, err := json.Marshal(parsed)
	if err != nil {
		return value
	}

	return string(sanitized)
}

// ImportTable imports data from a file to a table
