package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// AppVersion is the current version of the application.
// Update this on each release.
const AppVersion = "1.0.0"

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

		// Try to find a Windows installer/exe asset, fallback to HTML URL
		info.DownloadURL = release.HTMLURL
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi") || strings.HasSuffix(name, ".zip") {
				info.DownloadURL = asset.BrowserDownloadURL
				break
			}
		}
	} else {
		info.LatestVersion = latestVersion
	}

	return info
}

// OpenDownloadURL opens the given URL in the default browser.
func (a *App) OpenDownloadURL(url string) string {
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
