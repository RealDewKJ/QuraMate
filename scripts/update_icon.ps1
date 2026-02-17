param (
    [string]$InputPng = "..\build\appicon.png",
    [string]$OutputIco = "..\build\windows\icon.ico"
)

$ErrorActionPreference = "Stop"
$startPath = Get-Location
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $scriptPath

try {
    Write-Host "Loading System.Drawing..."
    Add-Type -AssemblyName System.Drawing

    $srcPath = Resolve-Path $InputPng
    Write-Host "Source Image: $srcPath"
    
    # Load the image
    $bitmap = [System.Drawing.Bitmap]::FromFile($srcPath)
    Write-Host "Image loaded. Size: $($bitmap.Width)x$($bitmap.Height)"
    
    # Create icon
    # Note: GetHicon creates a cursor/icon handle. This is a simple conversion.
    # consistently for usage as the app icon.
    $hIcon = $bitmap.GetHicon()
    $icon = [System.Drawing.Icon]::FromHandle($hIcon)
    
    $destPath = Join-Path $scriptPath $OutputIco
    $destFile = [System.IO.FileInfo]::new($destPath)
    $destDir = $destFile.Directory
    if (-not $destDir.Exists) {
        New-Item -ItemType Directory -Path $destDir.FullName -Force
    }

    Write-Host "Saving icon to: $destPath"
    $fs = New-Object System.IO.FileStream($destPath, [System.IO.FileMode]::Create)
    $icon.Save($fs)
    $fs.Close()
    
    # Cleanup
    [System.Runtime.InteropServices.Marshal]::DestroyIcon($hIcon)
    $bitmap.Dispose()
    $icon.Dispose()

    Write-Host "Icon updated successfully."
} catch {
    Write-Error "Failed to update icon: $_"
    exit 1
} finally {
    Set-Location $startPath
}
