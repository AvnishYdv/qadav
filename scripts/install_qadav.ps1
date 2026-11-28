# scripts/install_qadav.ps1
# Qadav One-Click Windows System Installer
# Installs Qadav compiler, tooling, stdlib, examples, VS Code extension, and registers PATH + file types
# Works from both the full Git repository and the standalone extracted SDK bundle

[CmdletBinding()]
param (
    [Parameter(Mandatory = $false)]
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Qadav"),

    [Parameter(Mandatory = $false)]
    [switch]$SkipPath,

    [Parameter(Mandatory = $false)]
    [switch]$SkipExtension
)

$ErrorActionPreference = 'Stop'
$sourceRoot = (Get-Item -Path $PSScriptRoot).Parent.FullName

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV (.ydv) - ONE-CLICK WINDOWS SYSTEM INSTALLER" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Source Location:  $sourceRoot" -ForegroundColor Gray
Write-Host "Target Directory: $InstallDir" -ForegroundColor Gray

# 1. Create Target Directory Tree
$binDir = Join-Path $InstallDir "bin"
$stdlibDir = Join-Path $InstallDir "stdlib"
$examplesDir = Join-Path $InstallDir "examples"
$scriptsDir = Join-Path $InstallDir "scripts"
$ideDir = Join-Path $InstallDir "ide"
$runtimeDir = Join-Path $InstallDir "runtime"

Write-Host "`n[STEP 1/5] Creating directory tree..." -ForegroundColor Yellow
New-Item -ItemType Directory -Path $binDir -Force | Out-Null
New-Item -ItemType Directory -Path $stdlibDir -Force | Out-Null
New-Item -ItemType Directory -Path $examplesDir -Force | Out-Null
New-Item -ItemType Directory -Path $scriptsDir -Force | Out-Null
New-Item -ItemType Directory -Path $ideDir -Force | Out-Null
New-Item -ItemType Directory -Path $runtimeDir -Force | Out-Null
Write-Host "  [OK] Directory tree initialized at $InstallDir" -ForegroundColor Green

# 2. Copy Binaries, Stdlib, Examples, Runtime, Scripts & IDE
Write-Host "`n[STEP 2/5] Copying binaries and runtime assets..." -ForegroundColor Yellow
Copy-Item (Join-Path $sourceRoot "bin\qyn.exe") $binDir -Force
if (Test-Path (Join-Path $sourceRoot "bin\qpk.exe")) {
    Copy-Item (Join-Path $sourceRoot "bin\qpk.exe") $binDir -Force
}
if (Test-Path (Join-Path $sourceRoot "stdlib")) {
    Copy-Item (Join-Path $sourceRoot "stdlib\*") $stdlibDir -Recurse -Force
}
if (Test-Path (Join-Path $sourceRoot "examples")) {
    Copy-Item (Join-Path $sourceRoot "examples\*") $examplesDir -Recurse -Force
}

# Runtime Objects (SDK layout: runtime\; Repo layout: TMP\runtime_objs\)
if (Test-Path (Join-Path $sourceRoot "runtime")) {
    Copy-Item (Join-Path $sourceRoot "runtime\*") $runtimeDir -Recurse -Force
} elseif (Test-Path (Join-Path $sourceRoot "TMP\runtime_objs")) {
    Copy-Item (Join-Path $sourceRoot "TMP\runtime_objs\*") $runtimeDir -Recurse -Force
}

# Ensure kernel32.lib is available locally in installed runtime
$localKernelLib = Join-Path $runtimeDir "kernel32.lib"
if (-not (Test-Path $localKernelLib)) {
    $kernelCandidates = @(
        (Join-Path $sourceRoot "runtime\kernel32.lib"),
        (Join-Path $env:USERPROFILE ".qadav\runtime\kernel32.lib"),
        (Join-Path $env:LOCALAPPDATA "Qadav\runtime\kernel32.lib")
    )
    $srcKernel = $kernelCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
    if ($srcKernel) {
        Copy-Item $srcKernel $localKernelLib -Force
    }
}

Copy-Item (Join-Path $sourceRoot "scripts\new_ai_project.ps1") $scriptsDir -Force
Copy-Item (Join-Path $sourceRoot "ide\vscode-qadav") $ideDir -Recurse -Force
Write-Host "  [OK] Compiler (qyn.exe), tools, stdlib, runtime, and IDE assets copied" -ForegroundColor Green

# 3. Create or Copy qadav-new.bat Helper in bin
Write-Host "`n[STEP 3/5] Configuring global CLI helper (qadav-new.bat)..." -ForegroundColor Yellow
$qadavNewBat = @"
@echo off
powershell -ExecutionPolicy Bypass -File "%~dp0..\scripts\new_ai_project.ps1" %*
"@
Set-Content -Path (Join-Path $binDir "qadav-new.bat") -Value $qadavNewBat -Encoding ASCII
Write-Host "  [OK] Global CLI helper created: $binDir\qadav-new.bat" -ForegroundColor Green

# 4. Idempotently Add bin to User PATH
if (-not $SkipPath) {
    Write-Host "`n[STEP 4/5] Updating User PATH environment variable..." -ForegroundColor Yellow
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $paths = ($userPath -split ';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
    $normalizedTarget = [System.IO.Path]::GetFullPath($binDir).TrimEnd('\')

    $alreadyInPath = $false
    foreach ($p in $paths) {
        if ([System.IO.Path]::GetFullPath($p).TrimEnd('\') -eq $normalizedTarget) {
            $alreadyInPath = $true
            break
        }
    }

    if (-not $alreadyInPath) {
        $newUserPath = ($paths + $binDir) -join ';'
        [Environment]::SetEnvironmentVariable("PATH", $newUserPath, "User")
        Write-Host "  [OK] Added '$binDir' to User PATH" -ForegroundColor Green
    } else {
        Write-Host "  [OK] '$binDir' is already present in User PATH (idempotent)" -ForegroundColor Gray
    }

    # Also update current session PATH
    if ($env:PATH -split ';' -notcontains $binDir) {
        $env:PATH = "$binDir;$env:PATH"
    }
} else {
    Write-Host "`n[STEP 4/5] Skipping PATH update as requested" -ForegroundColor Gray
}

# 5. Register .ydv File Type Metadata in HKCU
Write-Host "`n[STEP 5/5] Registering .ydv file type metadata in Windows Registry..." -ForegroundColor Yellow
try {
    $regBase = "HKCU:\Software\Classes"
    # .ydv extension mapping
    New-Item -Path "$regBase\.ydv" -Force | Out-Null
    Set-ItemProperty -Path "$regBase\.ydv" -Name "(Default)" -Value "QadavSourceFile"
    Set-ItemProperty -Path "$regBase\.ydv" -Name "PerceivedType" -Value "text"
    Set-ItemProperty -Path "$regBase\.ydv" -Name "Content Type" -Value "text/plain"

    # QadavSourceFile definition
    New-Item -Path "$regBase\QadavSourceFile" -Force | Out-Null
    Set-ItemProperty -Path "$regBase\QadavSourceFile" -Name "(Default)" -Value "Qadav Quantum-Native Source File (.ydv)"
    
    $iconPath = Join-Path $ideDir "vscode-qadav\icons\qadav-bansuri.png"
    if (Test-Path $iconPath) {
        New-Item -Path "$regBase\QadavSourceFile\DefaultIcon" -Force | Out-Null
        Set-ItemProperty -Path "$regBase\QadavSourceFile\DefaultIcon" -Name "(Default)" -Value "$iconPath,0"
    }

    New-Item -Path "$regBase\QadavSourceFile\shell\open\command" -Force | Out-Null
    Set-ItemProperty -Path "$regBase\QadavSourceFile\shell\open\command" -Name "(Default)" -Value "notepad.exe `"%1`""
    Write-Host "  [OK] Registered HKCU:\Software\Classes\.ydv -> QadavSourceFile" -ForegroundColor Green
} catch {
    Write-Warning "Failed to set registry keys: $_"
}

# 6. Auto-detect and Install VS Code / Cursor Extensions
if (-not $SkipExtension) {
    Write-Host "`n[BONUS] Auto-detecting IDE extensions directories..." -ForegroundColor Yellow
    $extSource = Join-Path $ideDir "vscode-qadav"
    $extTargetName = "qadav.qadav-lang-1.0.0"

    $vscodeExtDir = Join-Path $env:USERPROFILE ".vscode\extensions"
    if (Test-Path (Split-Path $vscodeExtDir -Parent)) {
        New-Item -ItemType Directory -Path $vscodeExtDir -Force | Out-Null
        $targetVscode = Join-Path $vscodeExtDir $extTargetName
        Copy-Item $extSource $targetVscode -Recurse -Force
        Write-Host "  [OK] Installed Qadav extension to VS Code: $targetVscode" -ForegroundColor Green
    }

    $cursorExtDir = Join-Path $env:USERPROFILE ".cursor\extensions"
    if (Test-Path (Split-Path $cursorExtDir -Parent)) {
        New-Item -ItemType Directory -Path $cursorExtDir -Force | Out-Null
        $targetCursor = Join-Path $cursorExtDir $extTargetName
        Copy-Item $extSource $targetCursor -Recurse -Force
        Write-Host "  [OK] Installed Qadav extension to Cursor: $targetCursor" -ForegroundColor Green
    }
}

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV INSTALLATION SUCCESSFUL! [READY FOR USE]" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Installed components:" -ForegroundColor White
Write-Host "  - Compiler:  $binDir\qyn.exe" -ForegroundColor Gray
Write-Host "  - Package:   $binDir\qpk.exe" -ForegroundColor Gray
Write-Host "  - CLI Tool:  $binDir\qadav-new.bat" -ForegroundColor Gray
Write-Host "  - Runtime:   $runtimeDir (62 compiled modules + kernel32.lib)" -ForegroundColor Gray
Write-Host "  - Stdlib:    $stdlibDir" -ForegroundColor Gray
Write-Host "  - IDE Ext:   $ideDir\vscode-qadav" -ForegroundColor Gray
Write-Host ""
Write-Host "Quickstart test command in any terminal:" -ForegroundColor Yellow
Write-Host "  qyn --version" -ForegroundColor Cyan
Write-Host "  qadav-new my_agent" -ForegroundColor Cyan
Write-Host ""
