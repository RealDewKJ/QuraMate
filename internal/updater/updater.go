package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppVersion = "dev"

const GitHubRepo = "RealDewKJ/QuraMate"

type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadURL    string `json:"downloadURL"`
	PublishedAt    string `json:"publishedAt"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

func resolveWindowsReleaseDownloadURL(assets []githubAsset) string {
	var installerURL string

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if name == "quramate-amd64-installer.exe" {
			return asset.BrowserDownloadURL
		}
		if strings.Contains(name, "setup") && strings.HasSuffix(name, ".exe") {
			return asset.BrowserDownloadURL
		}
		if installerURL == "" && (strings.HasSuffix(name, "-installer.exe") || strings.HasSuffix(name, ".msi")) {
			installerURL = asset.BrowserDownloadURL
		}
	}

	return installerURL
}

func deriveBootstrapperInstallerURL(downloadURL string) string {
	lowerURL := strings.ToLower(downloadURL)
	if strings.Contains(lowerURL, "quramate-amd64-installer.exe") {
		lastSlash := strings.LastIndex(downloadURL, "/")
		if lastSlash == -1 {
			return ""
		}

		return downloadURL[:lastSlash+1] + "QuraMate-amd64-package.exe"
	}

	if !strings.Contains(lowerURL, "setup") {
		return ""
	}

	lastSlash := strings.LastIndex(downloadURL, "/")
	if lastSlash == -1 {
		return ""
	}

	return downloadURL[:lastSlash+1] + "QuraMate-amd64-package.exe"
}

func extractVersionFromDownloadURL(downloadURL string) string {
	const marker = "/releases/download/v"

	idx := strings.Index(downloadURL, marker)
	if idx == -1 {
		return ""
	}

	start := idx + len(marker)
	remaining := downloadURL[start:]
	end := strings.Index(remaining, "/")
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(remaining[:end])
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetCurrentVersion() string {
	return AppVersion
}

func (s *Service) CheckForUpdates() UpdateInfo {
	info := UpdateInfo{
		Available:      false,
		CurrentVersion: AppVersion,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return info
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "QuraMate-Updater")

	resp, err := client.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return info
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return info
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	if compareVersions(latestVersion, AppVersion) > 0 {
		info.Available = true
		info.LatestVersion = latestVersion
		info.ReleaseNotes = release.Body
		info.PublishedAt = release.PublishedAt
		info.DownloadURL = release.HTMLURL
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if goRuntime.GOOS == "windows" {
				if resolvedURL := resolveWindowsReleaseDownloadURL(release.Assets); resolvedURL != "" {
					info.DownloadURL = resolvedURL
				}
				break
			} else if goRuntime.GOOS == "darwin" {
				if strings.HasSuffix(name, "macos-universal.zip") || strings.HasSuffix(name, "darwin.zip") || strings.HasSuffix(name, ".dmg") {
					info.DownloadURL = asset.BrowserDownloadURL
					break
				}
			} else if goRuntime.GOOS == "linux" {
				if strings.HasSuffix(name, "linux.zip") || strings.HasSuffix(name, ".tar.gz") {
					info.DownloadURL = asset.BrowserDownloadURL
					break
				}
			}
		}
	} else {
		info.LatestVersion = latestVersion
	}

	return info
}

func (s *Service) OpenDownloadURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "Error opening URL: Invalid URL scheme. Only http and https are allowed."
	}

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("Error opening URL: %s", err.Error())
	}
	return "Success"
}

func (s *Service) EmitProgress(ctx context.Context, stage string, percent int, message string) {
	wailsRuntime.EventsEmit(ctx, "app:update-progress", map[string]interface{}{
		"stage":   stage,
		"percent": percent,
		"message": message,
	})
}

func compareVersions(a string, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for len(aParts) < 3 {
		aParts = append(aParts, "0")
	}
	for len(bParts) < 3 {
		bParts = append(bParts, "0")
	}
	for i := 0; i < 3; i++ {
		aNum := parseVersionPart(aParts[i])
		bNum := parseVersionPart(bParts[i])
		if aNum > bNum {
			return 1
		}
		if aNum < bNum {
			return -1
		}
	}
	return 0
}

func parseVersionPart(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}

func createWindowsUpdateHelper(installerPath string, parentPID int, executablePath string) (string, error) {
	escapedInstallerPath := strings.ReplaceAll(installerPath, "'", "''")
	escapedExecutablePath := strings.ReplaceAll(executablePath, "'", "''")

	script := fmt.Sprintf(`$ErrorActionPreference = 'SilentlyContinue'
$parentPid = %d
$installer = '%s'
$exePath = '%s'

while (Get-Process -Id $parentPid -ErrorAction SilentlyContinue) {
    Start-Sleep -Milliseconds 300
}

Start-Sleep -Milliseconds 700

$installerArgs = '/S'

$installProcess = Start-Process -FilePath $installer -ArgumentList $installerArgs -Verb RunAs -Wait -PassThru

Start-Sleep -Seconds 1

if ($exePath -ne '' -and (Test-Path $exePath)) {
    $runningApp = Get-Process -Name 'QuraMate' -ErrorAction SilentlyContinue
    if (-not $runningApp) {
        Start-Process -FilePath $exePath
    }
}

exit $installProcess.ExitCode
`, parentPID, escapedInstallerPath, escapedExecutablePath)

	helperFile, err := os.CreateTemp("", "QuraMate-Update-Helper-*.ps1")
	if err != nil {
		return "", fmt.Errorf("failed to create update helper: %w", err)
	}
	defer helperFile.Close()

	if _, err := helperFile.WriteString(script); err != nil {
		return "", fmt.Errorf("failed to write update helper: %w", err)
	}

	return helperFile.Name(), nil
}

func (s *Service) PerformUpdate(downloadURL string) error {
	if !strings.HasPrefix(downloadURL, "https://github.com/RealDewKJ/QuraMate/releases/download/") {
		return fmt.Errorf("invalid download URL source: must be an official update URL")
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code downloading update: %d", resp.StatusCode)
	}

	ext := filepath.Ext(strings.ToLower(downloadURL))
	if ext == "" || len(ext) > 6 {
		ext = ".bin"
	}

	tmpFile, err := os.CreateTemp("", "QuraMate-Update-*"+ext)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		if _, err = io.Copy(tmpFile, resp.Body); err != nil {
			return fmt.Errorf("failed to save update file: %w", err)
		}
	} else {
		buf := make([]byte, 64*1024)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
					return fmt.Errorf("failed to save update file: %w", writeErr)
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return fmt.Errorf("failed to save update file: %w", readErr)
			}
		}
	}

	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to flush update file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close update file: %w", err)
	}

	installerPath := tmpFile.Name()
	lowerPath := strings.ToLower(installerPath)
	currentExecutablePath, currentExecutableErr := os.Executable()
	currentPID := os.Getpid()

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		if strings.HasSuffix(lowerPath, ".msi") {
			if currentExecutableErr != nil {
				cmd = exec.Command("msiexec", "/i", installerPath, "/qn", "/norestart")
			} else {
				cmd = exec.Command(
					"powershell",
					"-NoProfile",
					"-ExecutionPolicy", "Bypass",
					"-Command",
					"Start-Process -FilePath 'msiexec.exe' -ArgumentList @('/i', $args[0], '/qn', '/norestart') -Verb RunAs -Wait; Start-Sleep -Milliseconds 800; Start-Process -FilePath $args[1]",
					installerPath,
					currentExecutablePath,
				)
			}
		} else if strings.HasSuffix(lowerPath, ".exe") {
			executablePath := ""
			if currentExecutableErr == nil {
				executablePath = currentExecutablePath
			}
			installerURL := deriveBootstrapperInstallerURL(downloadURL)
			targetVersion := extractVersionFromDownloadURL(downloadURL)

			if installerURL != "" {
				bootstrapperArgs := []string{
					fmt.Sprintf(`--installer-url=%s`, installerURL),
					"--mode=update",
				}
				if targetVersion != "" {
					bootstrapperArgs = append(bootstrapperArgs, fmt.Sprintf(`--version=%s`, targetVersion))
				}
				if executablePath != "" {
					bootstrapperArgs = append(bootstrapperArgs, fmt.Sprintf(`--executable-path=%s`, executablePath))
				}

				cmd = exec.Command(
					"powershell",
					"-NoProfile",
					"-ExecutionPolicy", "Bypass",
					"-WindowStyle", "Hidden",
					"-Command",
					"Start-Process -FilePath $args[0] -ArgumentList $args[1..($args.Length-1)] -WindowStyle Normal",
					installerPath,
				)
				cmd.Args = append(cmd.Args, bootstrapperArgs...)
			} else {
				helperPath, helperErr := createWindowsUpdateHelper(installerPath, currentPID, executablePath)
				if helperErr != nil {
					return helperErr
				}

				cmd = exec.Command(
					"powershell",
					"-NoProfile",
					"-ExecutionPolicy", "Bypass",
					"-WindowStyle", "Hidden",
					"-File",
					helperPath,
				)
			}
		} else {
			cmd = exec.Command("cmd.exe", "/c", "start", "", installerPath)
		}
	case "darwin":
		cmd = exec.Command("open", installerPath)
	default:
		cmd = exec.Command("xdg-open", installerPath)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start installer: %w", err)
	}

	time.Sleep(600 * time.Millisecond)
	os.Exit(0)

	return nil
}
