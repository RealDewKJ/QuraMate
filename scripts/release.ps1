<#
.SYNOPSIS
    Bumps version across all project files, commits, tags, and pushes.

.DESCRIPTION
    Updates version in: wails.json, frontend/package.json, npm/package.json, updater.go
    Then commits, creates a git tag, and pushes everything.

.EXAMPLE
    .\scripts\release.ps1 1.2.1
    .\scripts\release.ps1 2.0.0-beta.1
#>

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Version
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

# Validate version format (semver)
if ($Version -notmatch '^\d+\.\d+\.\d+') {
    Write-Host "  ❌ Invalid version format: $Version" -ForegroundColor Red
    Write-Host "  Expected: X.Y.Z (e.g., 1.2.1, 2.0.0-beta.1)" -ForegroundColor Yellow
    exit 1
}

$Tag = "v$Version"

Write-Host ""
Write-Host "  🚀 QuraMate Release — $Tag" -ForegroundColor Cyan
Write-Host "  ─────────────────────────────" -ForegroundColor DarkGray
Write-Host ""

# Check for uncommitted changes
$status = git status --porcelain
if ($status) {
    Write-Host "  ⚠️  You have uncommitted changes:" -ForegroundColor Yellow
    git status --short
    Write-Host ""
    $confirm = Read-Host "  Continue anyway? (y/N)"
    if ($confirm -ne 'y') {
        Write-Host "  Aborted." -ForegroundColor Red
        exit 1
    }
}

# Check if tag already exists
$existingTag = git tag -l $Tag
if ($existingTag) {
    Write-Host "  ❌ Tag $Tag already exists!" -ForegroundColor Red
    exit 1
}

# 1. Update wails.json
Write-Host "  📝 wails.json" -ForegroundColor White -NoNewline
$wailsJson = Get-Content "wails.json" -Raw | ConvertFrom-Json
$wailsJson.version = $Version
$wailsJson | ConvertTo-Json -Depth 10 | Set-Content "wails.json" -Encoding UTF8 -NoNewline
# Re-add trailing newline
Add-Content "wails.json" ""
Write-Host " → $Version" -ForegroundColor Green

# 2. Update frontend/package.json
Write-Host "  📝 frontend/package.json" -ForegroundColor White -NoNewline
Push-Location "frontend"
npm version $Version --no-git-tag-version --allow-same-version 2>$null | Out-Null
Pop-Location
Write-Host " → $Version" -ForegroundColor Green

# 3. Update npm/package.json
Write-Host "  📝 npm/package.json" -ForegroundColor White -NoNewline
Push-Location "npm"
npm version $Version --no-git-tag-version --allow-same-version 2>$null | Out-Null
Pop-Location
Write-Host " → $Version" -ForegroundColor Green

# 4. Update updater.go (fallback version)
Write-Host "  📝 updater.go" -ForegroundColor White -NoNewline
$updaterContent = Get-Content "updater.go" -Raw
$updaterContent = $updaterContent -replace 'var AppVersion = "[^"]*"', "var AppVersion = `"$Version`""
Set-Content "updater.go" -Value $updaterContent -Encoding UTF8 -NoNewline
Write-Host " → $Version" -ForegroundColor Green

# 5. Update homebrew cask version
$caskFile = "homebrew/Casks/quramate.rb"
if (Test-Path $caskFile) {
    Write-Host "  📝 homebrew/Casks/quramate.rb" -ForegroundColor White -NoNewline
    $caskContent = Get-Content $caskFile -Raw
    $caskContent = $caskContent -replace 'version "[^"]*"', "version `"$Version`""
    Set-Content $caskFile -Value $caskContent -Encoding UTF8 -NoNewline
    Write-Host " → $Version" -ForegroundColor Green
}

Write-Host ""

# 6. Git commit + tag + push
Write-Host "  📦 Committing..." -ForegroundColor White
git add wails.json frontend/package.json npm/package.json updater.go homebrew/Casks/quramate.rb 2>$null
git commit -m "chore: bump version to $Version"

Write-Host "  🏷️  Creating tag $Tag..." -ForegroundColor White
git tag -a $Tag -m "Release $Tag"

Write-Host ""
$push = 'y'
if ($push -eq '' -or $push -eq 'y' -or $push -eq 'Y') {
    git push origin HEAD
    git push origin $Tag
    Write-Host ""
    Write-Host "  ✅ Done! Release $Tag pushed." -ForegroundColor Green
    Write-Host "  CI will now build + publish automatically." -ForegroundColor DarkGray
} else {
    Write-Host ""
    Write-Host "  ✅ Commit + tag created locally." -ForegroundColor Green
    Write-Host "  Run 'git push origin HEAD && git push origin $Tag' when ready." -ForegroundColor Yellow
}

Write-Host ""
