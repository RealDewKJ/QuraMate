package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AppVersion is the current version of the application.
// This is overridden at build time via: -ldflags "-X main.AppVersion=x.y.z"
var AppVersion = "1.1.0"

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

		// Try to find the correct asset for the current OS
		info.DownloadURL = release.HTMLURL
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if runtime.GOOS == "windows" {
				if strings.HasSuffix(name, "-installer.exe") || strings.HasSuffix(name, "windows.zip") || strings.HasSuffix(name, "win.zip") {
					info.DownloadURL = asset.BrowserDownloadURL
					break
				}
			} else if runtime.GOOS == "darwin" {
				if strings.HasSuffix(name, "macos-universal.zip") || strings.HasSuffix(name, "darwin.zip") || strings.HasSuffix(name, ".dmg") {
					info.DownloadURL = asset.BrowserDownloadURL
					break
				}
			} else if runtime.GOOS == "linux" {
				// Add linux matching if needed
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
	switch runtime.GOOS {
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

// PerformUpdate downloads the update from the provided URL, saves it to a temp file, and runs the installer.
func (a *App) PerformUpdate(downloadURL string) error {
	if !strings.HasPrefix(downloadURL, "https://github.com/RealDewKJ/QuraMate/releases/download/") {
		return fmt.Errorf("invalid download URL source: must be an official update URL")
	}

	client := &http.Client{Timeout: 5 * time.Minute} // Large timeout for download
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

	// Create a temporary file to save the installer
	tmpFile, err := os.CreateTemp("", "QuraMate-Update-*.exe")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	// Copy the downloaded data into the temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save update file: %w", err)
	}

	// Make sure everything is written to disk
	tmpFile.Sync()
	tmpFile.Close()

	installerPath := tmpFile.Name()

	// Execute the installer using OS-specific launcher to handle elevation prompts (UAC on Windows)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Use ShellExecute via cmd /c start to ensure UAC prompt is handled and it runs detached
		// Providing the full path and ensuring it's treated as a single argument
		cmd = exec.Command("cmd.exe", "/c", "start", "", installerPath)
	case "darwin":
		cmd = exec.Command("open", installerPath)
	default:
		cmd = exec.Command("xdg-open", installerPath)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start installer: %w", err)
	}

	// Exit the current app immediately so the installer can overwrite the files.
	// We give the OS a very tiny bit of time to start the process before exiting.
	time.Sleep(500 * time.Millisecond)
	os.Exit(0)

	return nil
}
