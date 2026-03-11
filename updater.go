package main

import (
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

// AppVersion is the current version of the application.
// This is overridden at build time via: -ldflags "-X main.AppVersion=x.y.z"
var AppVersion = "1.1.8"

// GitHubRepo is the GitHub repository for update checks.
const GitHubRepo = "RealDewKJ/QuraMate"

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadURL    string `json:"downloadURL"`
	PublishedAt    string `json:"publishedAt"`
}

// githubRelease represents the JSON structure from GitHub API.
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

// GetCurrentVersion returns the current app version.
func (a *App) GetCurrentVersion() string {
	return AppVersion
}

// CheckForUpdates checks GitHub for a newer release.
func (a *App) CheckForUpdates() UpdateInfo {
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

		// Try to find the correct asset for the current OS.
		info.DownloadURL = release.HTMLURL
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if goRuntime.GOOS == "windows" {
				// Prefer installer assets so we can run a silent update flow.
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

// OpenDownloadURL opens the given URL in the default browser.
func (a *App) OpenDownloadURL(url string) string {
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

func (a *App) emitUpdateProgress(stage string, percent int, message string) {
	if a.ctx == nil {
		return
	}

	wailsRuntime.EventsEmit(a.ctx, "app:update-progress", map[string]interface{}{
		"stage":   stage,
		"percent": percent,
		"message": message,
	})
}

// compareVersions compares two semver strings.
// Returns 1 if a > b, -1 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// Pad to equal length
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

// parseVersionPart parses a version part string to an integer.
func parseVersionPart(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break // stop at first non-digit (e.g. "1-beta")
		}
	}
	return n
}

// PerformUpdate downloads the update from the provided URL and runs a silent installer.
func (a *App) PerformUpdate(downloadURL string) error {
	if !strings.HasPrefix(downloadURL, "https://github.com/RealDewKJ/QuraMate/releases/download/") {
		return fmt.Errorf("invalid download URL source: must be an official update URL")
	}

	a.emitUpdateProgress("preparing", 5, "Preparing update package...")

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

	a.emitUpdateProgress("downloading", 10, "Downloading update...")

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		if _, err = io.Copy(tmpFile, resp.Body); err != nil {
			return fmt.Errorf("failed to save update file: %w", err)
		}
	} else {
		buf := make([]byte, 64*1024)
		var written int64

		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
					return fmt.Errorf("failed to save update file: %w", writeErr)
				}

				written += int64(n)
				downloadPercent := int((written * 100) / contentLength)
				if downloadPercent > 100 {
					downloadPercent = 100
				}

				// Keep room for install steps by mapping download progress to 10-85.
				uiPercent := 10 + int(float64(downloadPercent)*0.75)
				if uiPercent > 85 {
					uiPercent = 85
				}
				a.emitUpdateProgress("downloading", uiPercent, "Downloading update...")
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

	a.emitUpdateProgress("installing", 90, "Installing update...")

	installerPath := tmpFile.Name()
	lowerPath := strings.ToLower(installerPath)

	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "windows":
		if strings.HasSuffix(lowerPath, ".msi") {
			cmd = exec.Command("msiexec", "/i", installerPath, "/qn", "/norestart")
		} else if strings.HasSuffix(lowerPath, ".exe") {
			// Wails NSIS installers accept /S for silent mode.
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

	a.emitUpdateProgress("finalizing", 98, "Finalizing update...")

	time.Sleep(600 * time.Millisecond)
	os.Exit(0)

	return nil
}
