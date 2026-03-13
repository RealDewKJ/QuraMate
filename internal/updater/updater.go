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

var AppVersion = "1.2.0"

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
				if strings.HasSuffix(name, "-installer.exe") || strings.HasSuffix(name, ".msi") || strings.Contains(name, "setup") {
					info.DownloadURL = asset.BrowserDownloadURL
					break
				}
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

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		if strings.HasSuffix(lowerPath, ".msi") {
			cmd = exec.Command("msiexec", "/i", installerPath, "/qn", "/norestart")
		} else if strings.HasSuffix(lowerPath, ".exe") {
			cmd = exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
				"Start-Process -FilePath $args[0] -ArgumentList '/S' -Verb RunAs", installerPath)
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
