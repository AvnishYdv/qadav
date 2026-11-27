# scripts/uninstall_qadav.ps1
# Qadav Clean Windows Uninstaller

[CmdletBinding()]
param (
    [Parameter(Mandatory = $false)]
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Qadav")
)

$ErrorActionPreference = 'Stop'

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV (.ydv) - WINDOWS SYSTEM UNINSTALLER" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Target Install Directory: $InstallDir" -ForegroundColor Gray

# 1. Remove from User PATH
Write-Host "`n[STEP 1/4] Removing Qadav from User PATH..." -ForegroundColor Yellow
$binDir = Join-Path $InstallDir "bin"
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath) {
    $normalizedTarget = [System.IO.Path]::GetFullPath($binDir).TrimEnd('\')
    $paths = ($userPath -split ';') | Where-Object {
        $_ -and $_.Trim().Length -gt 0 -and [System.IO.Path]::GetFullPath($_).TrimEnd('\') -ne $normalizedTarget
    }
    $newUserPath = $paths -join ';'
    [Environment]::SetEnvironmentVariable("PATH", $newUserPath, "User")
    Write-Host "  [OK] Cleaned User PATH" -ForegroundColor Green
}

# 2. Unregister .ydv Registry Keys
Write-Host "`n[STEP 2/4] Removing .ydv Windows Registry associations..." -ForegroundColor Yellow
try {
    if (Test-Path "HKCU:\Software\Classes\.ydv") {
        Remove-Item "HKCU:\Software\Classes\.ydv" -Recurse -Force | Out-Null
    }
    if (Test-Path "HKCU:\Software\Classes\QadavSourceFile") {
        Remove-Item "HKCU:\Software\Classes\QadavSourceFile" -Recurse -Force | Out-Null
    }
    Write-Host "  [OK] Registry associations cleaned" -ForegroundColor Green
} catch {
    Write-Warning "Registry cleanup note: $_"
}

# 3. Remove IDE Extensions
Write-Host "`n[STEP 3/4] Removing installed IDE extensions..." -ForegroundColor Yellow
$extTargetName = "qadav.qadav-lang-1.0.0"
$vscodeExt = Join-Path $env:USERPROFILE ".vscode\extensions\$extTargetName"
if (Test-Path $vscodeExt) {
    Remove-Item $vscodeExt -Recurse -Force | Out-Null
    Write-Host "  [OK] Removed VS Code extension: $vscodeExt" -ForegroundColor Green
}
$cursorExt = Join-Path $env:USERPROFILE ".cursor\extensions\$extTargetName"
if (Test-Path $cursorExt) {
    Remove-Item $cursorExt -Recurse -Force | Out-Null
    Write-Host "  [OK] Removed Cursor extension: $cursorExt" -ForegroundColor Green
}

# 4. Remove Install Directory
Write-Host "`n[STEP 4/4] Removing installation directory tree..." -ForegroundColor Yellow
if (Test-Path $InstallDir) {
    Remove-Item $InstallDir -Recurse -Force | Out-Null
    Write-Host "  [OK] Removed $InstallDir" -ForegroundColor Green
}

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV UNINSTALLATION COMPLETE" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
