param (
    [string]$InputPng = "..\build\appicon.png",
    [string]$OutputIco = "..\build\windows\icon.ico"
)

$ErrorActionPreference = "Stop"
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
$startPath = Get-Location
Set-Location $scriptPath

Add-Type -AssemblyName System.Drawing

function Resize-Image {
    param(
        [System.Drawing.Image]$Image,
        [int]$Size
    )
    $destRect = New-Object System.Drawing.Rectangle(0, 0, $Size, $Size)
    $destImage = New-Object System.Drawing.Bitmap($Size, $Size)
    $destImage.SetResolution($Image.HorizontalResolution, $Image.VerticalResolution)

    $graphics = [System.Drawing.Graphics]::FromImage($destImage)
    $graphics.CompositingMode = [System.Drawing.Drawing2D.CompositingMode]::SourceCopy
    $graphics.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality

    $wrapMode = New-Object System.Drawing.Imaging.ImageAttributes
    $wrapMode.SetWrapMode([System.Drawing.Drawing2D.WrapMode]::TileFlipXY)
    $graphics.DrawImage($Image, $destRect, 0, 0, $Image.Width, $Image.Height, [System.Drawing.GraphicsUnit]::Pixel, $wrapMode)
    $graphics.Dispose()
    
    return $destImage
}

try {
    $srcPath = Resolve-Path $InputPng
    Write-Host "Source Image: $srcPath"

    $sourceImage = [System.Drawing.Image]::FromFile($srcPath)
    
    # Define target sizes
    $sizes = @(256, 48, 32, 16)
    $images = @()

    foreach ($s in $sizes) {
        Write-Host "Resizing to ${s}x${s}..."
        $resized = Resize-Image -Image $sourceImage -Size $s
        
        # Convert to PNG byte array
        $ms = New-Object System.IO.MemoryStream
        $resized.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
        $bytes = $ms.ToArray()
        $ms.Close()
        
        $images += @{
            Size = $s
            Bytes = $bytes
        }
        $resized.Dispose()
    }
    $sourceImage.Dispose()

    # Create ICO File
    $destPath = Join-Path $scriptPath $OutputIco
    $destFile = [System.IO.FileInfo]::new($destPath)
    if (-not $destFile.Directory.Exists) {
        New-Item -ItemType Directory -Path $destFile.Directory.FullName -Force
    }
    
    $fs = [System.IO.File]::Create($destPath)
    
    # 1. Header
    # Reserved (2), Type (2), Count (2)
    # 0,0, 1,0, Count,0
    $count = $images.Count
    $header = [byte[]]@(0, 0, 1, 0, [byte]$count, 0)
    $fs.Write($header, 0, $header.Length)
    
    # Calculate offset for first image data
    # Header (6) + (Count * 16 per entry)
    $currentOffset = 6 + ($count * 16)
    
    # 2. Directory Entries
    foreach ($img in $images) {
        $w = if ($img.Size -ge 256) { 0 } else { [byte]$img.Size }
        $h = if ($img.Size -ge 256) { 0 } else { [byte]$img.Size }
        
        $size = $img.Bytes.Length
        
        $entry = [byte[]]@(
            $w, $h, 0, 0,  # W, H, Colors, Res
            1, 0,          # Planes
            32, 0          # BPP
        )
        # Size (4 bytes)
        $entry += [BitConverter]::GetBytes([uint32]$size)
        # Offset (4 bytes)
        $entry += [BitConverter]::GetBytes([uint32]$currentOffset)
        
        $fs.Write($entry, 0, $entry.Length)
        
        $currentOffset += $size
    }
    
    # 3. Image Data
    foreach ($img in $images) {
        $fs.Write($img.Bytes, 0, $img.Bytes.Length)
    }
    
    $fs.Close()
    
    Write-Host "Success! Created multi-size ICO at $destPath"
    
} catch {
    Write-Error "Failed: $_"
    exit 1
} finally {
    Set-Location $startPath
}
