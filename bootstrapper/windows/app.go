package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"
)

var AppVersion = "dev"

const githubRepo = "RealDewKJ/QuraMate"
const uninstallRegistryRoot = `Software\Microsoft\Windows\CurrentVersion\Uninstall`

type LaunchContext struct {
	Mode           string `json:"mode"`
	Version        string `json:"version"`
	InstallerPath  string `json:"installerPath"`
	InstallerURL   string `json:"installerUrl"`
	ExecutablePath string `json:"executablePath"`
	InstallDir     string `json:"installDir"`
}

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetLaunchContext() LaunchContext {
	return LaunchContext{
		Mode:           detectMode(),
		Version:        detectVersion(),
		InstallerPath:  detectInstallerPath(),
		InstallerURL:   detectInstallerURL(),
		ExecutablePath: detectExecutablePath(),
		InstallDir:     detectInstallDir(),
	}
}

func (a *App) BeginInstall(installerPath string, executablePath string) error {
	if a.ctx == nil {
		return fmt.Errorf("bootstrapper is not ready")
	}

	resolvedInstallerPath := strings.TrimSpace(installerPath)
	if resolvedInstallerPath == "" {
		resolvedInstallerPath = detectInstallerPath()
	}
	if resolvedInstallerPath == "" {
		downloadedInstallerPath, downloadErr := downloadInstallerAsset(detectInstallerURL())
		if downloadErr != nil {
			return downloadErr
		}
		resolvedInstallerPath = downloadedInstallerPath
	}

	if _, err := os.Stat(resolvedInstallerPath); err != nil {
		return fmt.Errorf("installer not found: %w", err)
	}

	resolvedExecutablePath := strings.TrimSpace(executablePath)
	if resolvedExecutablePath == "" {
		resolvedExecutablePath = detectExecutablePath()
	}

	runtime.EventsEmit(a.ctx, "bootstrapper:status", map[string]string{
		"stage":   "preparing",
		"message": "Preparing QuraMate setup...",
	})

	helperPath, err := createBootstrapperHelper(resolvedInstallerPath, resolvedExecutablePath, os.Getpid())
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden",
		"-File", helperPath,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch install helper: %w", err)
	}

	go func() {
		runtime.EventsEmit(a.ctx, "bootstrapper:status", map[string]string{
			"stage":   "handoff",
			"message": "Handing off to background installer...",
		})

		time.Sleep(900 * time.Millisecond)
		runtime.Quit(a.ctx)
	}()

	return nil
}

func createBootstrapperHelper(installerPath string, executablePath string, parentPID int) (string, error) {
	escapedInstallerPath := strings.ReplaceAll(installerPath, "'", "''")
	escapedExecutablePath := strings.ReplaceAll(executablePath, "'", "''")

	script := fmt.Sprintf(`$ErrorActionPreference = 'SilentlyContinue'
$parentPid = %d
$installer = '%s'
$exePath = '%s'

while (Get-Process -Id $parentPid -ErrorAction SilentlyContinue) {
    Start-Sleep -Milliseconds 300
}

Start-Sleep -Milliseconds 500

$installProcess = Start-Process -FilePath $installer -ArgumentList '/S' -Verb RunAs -Wait -PassThru

Start-Sleep -Milliseconds 900

if ($exePath -ne '' -and (Test-Path $exePath)) {
    Start-Process -FilePath $exePath
}

exit $installProcess.ExitCode
`, parentPID, escapedInstallerPath, escapedExecutablePath)

	helperFile, err := os.CreateTemp("", "QuraMate-Bootstrapper-*.ps1")
	if err != nil {
		return "", fmt.Errorf("failed to create bootstrapper helper: %w", err)
	}
	defer helperFile.Close()

	if _, err := helperFile.WriteString(script); err != nil {
		return "", fmt.Errorf("failed to write bootstrapper helper: %w", err)
	}

	return helperFile.Name(), nil
}

func detectVersion() string {
	for _, arg := range os.Args[1:] {
		if value, ok := strings.CutPrefix(arg, "--version="); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	if envVersion := strings.TrimSpace(os.Getenv("QURAMATE_BOOTSTRAPPER_VERSION")); envVersion != "" {
		return envVersion
	}
	return AppVersion
}

func detectMode() string {
	for _, arg := range os.Args[1:] {
		if value, ok := strings.CutPrefix(arg, "--mode="); ok {
			mode := strings.ToLower(strings.TrimSpace(value))
			if mode == "update" {
				return "update"
			}
		}

		if strings.HasPrefix(arg, "--installer-url=") {
			return "update"
		}
	}

	return "install"
}

func detectInstallerPath() string {
	currentExePath := currentExecutablePath()
	exeDir := filepath.Dir(currentExePath)
	workingDir, _ := os.Getwd()
	candidates := []string{
		filepath.Join(exeDir, "QuraMate-amd64-package.exe"),
		filepath.Join(exeDir, "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(exeDir, "..", "..", "..", "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(exeDir, "..", "..", "..", "..", "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(exeDir, "..", "..", "..", "..", "..", "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(workingDir, "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(workingDir, "..", "build", "bin", "QuraMate-amd64-package.exe"),
		filepath.Join(repoRootFromWorkingDir(), "build", "bin", "QuraMate-amd64-package.exe"),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		resolvedCandidate := filepath.Clean(candidate)
		if currentExePath != "" && strings.EqualFold(resolvedCandidate, currentExePath) {
			continue
		}
		if _, err := os.Stat(resolvedCandidate); err == nil {
			return resolvedCandidate
		}
	}

	return ""
}

func detectExecutablePath() string {
	for _, arg := range os.Args[1:] {
		if value, ok := strings.CutPrefix(arg, "--executable-path="); ok {
			candidate := strings.TrimSpace(value)
			if candidate != "" {
				return filepath.Clean(candidate)
			}
		}
	}

	if envPath := strings.TrimSpace(os.Getenv("QURAMATE_BOOTSTRAPPER_EXECUTABLE_PATH")); envPath != "" {
		return filepath.Clean(envPath)
	}

	installDir := detectInstallDir()
	if installDir == "" {
		return ""
	}

	return filepath.Join(installDir, "QuraMate.exe")
}

func detectInstallerURL() string {
	for _, arg := range os.Args[1:] {
		if value, ok := strings.CutPrefix(arg, "--installer-url="); ok {
			return strings.TrimSpace(value)
		}
	}

	if envURL := strings.TrimSpace(os.Getenv("QURAMATE_BOOTSTRAPPER_INSTALLER_URL")); envURL != "" {
		return envURL
	}

	version := strings.TrimSpace(detectVersion())
	if version == "" {
		return ""
	}

	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/QuraMate-amd64-package.exe", githubRepo, version)
}

func downloadInstallerAsset(installerURL string) (string, error) {
	if strings.TrimSpace(installerURL) == "" {
		return "", fmt.Errorf("installer package could not be resolved")
	}

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(installerURL)
	if err != nil {
		return "", fmt.Errorf("failed to download installer asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected installer download status: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "QuraMate-Bootstrapper-Installer-*.exe")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary installer file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save installer asset: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize installer asset: %w", err)
	}

	return tmpFile.Name(), nil
}

func detectInstallDir() string {
	if installDir := readInstallLocationFromRegistry(); installDir != "" {
		return installDir
	}

	if programFiles := strings.TrimSpace(os.Getenv("ProgramFiles")); programFiles != "" {
		return filepath.Join(programFiles, "QuraMate")
	}
	return ""
}

func readInstallLocationFromRegistry() string {
	keyNames := []string{
		"QuraMateQuraMate",
		"QuraMate",
	}

	roots := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}
	for _, root := range roots {
		for _, keyName := range keyNames {
			keyPath := uninstallRegistryRoot + `\` + keyName
			key, err := registry.OpenKey(root, keyPath, registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			installLocation, _, err := key.GetStringValue("InstallLocation")
			key.Close()
			if err != nil {
				continue
			}

			trimmed := strings.TrimSpace(installLocation)
			if trimmed != "" {
				return filepath.Clean(trimmed)
			}
		}
	}

	return ""
}

func currentExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return path
}

func repoRootFromWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return filepath.Clean(filepath.Join(workingDir, "..", ".."))
}
