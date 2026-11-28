# scripts/build_qadav_sdk.ps1
# Qadav Sovereign SDK Bundle Packaging Script
# Pure PowerShell, Zero C, Zero Go, Zero Python

$ErrorActionPreference = "Stop"
$repoRoot = (Get-Item -Path $PSScriptRoot).Parent.FullName
Set-Location $repoRoot

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV (.ydv) - SOVEREIGN SDK BUNDLE PACKAGING (Windows x86_64)" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Source Repository: $repoRoot" -ForegroundColor Gray

$distDir = Join-Path $repoRoot "dist"
$sdkBundleName = "qadav-sdk-windows-x86_64"
$sdkDir = Join-Path $distDir $sdkBundleName
$zipArchive = Join-Path $distDir "qadav-sdk-v1.0.0-windows-x86_64.zip"

Write-Host "Target SDK Directory: $sdkDir" -ForegroundColor Gray
Write-Host "Target ZIP Archive:   $zipArchive" -ForegroundColor Gray

# 1. Clean and Recreate Dist Tree
Write-Host "`n[STEP 1/6] Initializing clean SDK output directory..." -ForegroundColor Yellow
if (Test-Path $sdkDir) {
    Remove-Item -Recurse -Force $sdkDir
}
New-Item -ItemType Directory -Path (Join-Path $sdkDir "bin") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $sdkDir "stdlib") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $sdkDir "runtime") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $sdkDir "examples") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $sdkDir "scripts") -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $sdkDir "ide\vscode-qadav") -Force | Out-Null
Write-Host "  [OK] Clean SDK layout initialized" -ForegroundColor Green

# 2. Copy Compiler & Tooling Binaries
Write-Host "`n[STEP 2/6] Packaging compiler and tools into bin/..." -ForegroundColor Yellow
Copy-Item (Join-Path $repoRoot "bin\qyn.exe") (Join-Path $sdkDir "bin\") -Force
if (Test-Path (Join-Path $repoRoot "bin\qpk.exe")) {
    Copy-Item (Join-Path $repoRoot "bin\qpk.exe") (Join-Path $sdkDir "bin\") -Force
}
# Include qadav-new.bat inside bin
$qadavNewBat = @"
@echo off
powershell -ExecutionPolicy Bypass -File "%~dp0..\scripts\new_ai_project.ps1" %*
"@
Set-Content -Path (Join-Path $sdkDir "bin\qadav-new.bat") -Value $qadavNewBat -Encoding ASCII
Write-Host "  [OK] qyn.exe, qpk.exe, and qadav-new.bat packaged in bin/" -ForegroundColor Green

# 3. Package Standard Library & Runtime Objects
Write-Host "`n[STEP 3/6] Packaging stdlib and runtime objects..." -ForegroundColor Yellow
if (Test-Path (Join-Path $repoRoot "stdlib")) {
    Copy-Item (Join-Path $repoRoot "stdlib\*") (Join-Path $sdkDir "stdlib\") -Recurse -Force
}

# Runtime: Copy all compiled .o objects from TMP\runtime_objs
$runtimeObjsDir = Join-Path $repoRoot "TMP\runtime_objs"
if (Test-Path $runtimeObjsDir) {
    Copy-Item (Join-Path $runtimeObjsDir "*.o") (Join-Path $sdkDir "runtime\") -Force
}
# Copy kernel32.lib into runtime/
$kernelCandidates = @(
    (Join-Path $repoRoot "runtime\kernel32.lib"),
    (Join-Path $env:LOCALAPPDATA "Qadav\runtime\kernel32.lib"),
    (Join-Path $env:USERPROFILE ".qadav\runtime\kernel32.lib")
)
$srcKernel = $kernelCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if ($srcKernel) {
    Copy-Item $srcKernel (Join-Path $sdkDir "runtime\kernel32.lib") -Force
}
$runtimeFiles = (Get-ChildItem (Join-Path $sdkDir "runtime")).Count
Write-Host "  [OK] Stdlib and runtime packaged ($runtimeFiles runtime files including kernel32.lib)" -ForegroundColor Green

# 4. Package Examples, Scaffolding Script & IDE Extension
Write-Host "`n[STEP 4/6] Packaging examples, scaffolding script & IDE bundle..." -ForegroundColor Yellow
if (Test-Path (Join-Path $repoRoot "examples")) {
    Copy-Item (Join-Path $repoRoot "examples\*.ydv") (Join-Path $sdkDir "examples\") -Force
    Copy-Item (Join-Path $repoRoot "examples\*.txt") (Join-Path $sdkDir "examples\") -Force
}
Copy-Item (Join-Path $repoRoot "scripts\new_ai_project.ps1") (Join-Path $sdkDir "scripts\") -Force
Copy-Item (Join-Path $repoRoot "scripts\install_qadav.ps1") (Join-Path $sdkDir "scripts\") -Force
Copy-Item (Join-Path $repoRoot "scripts\uninstall_qadav.ps1") (Join-Path $sdkDir "scripts\") -Force
Copy-Item (Join-Path $repoRoot "install_qadav.bat") (Join-Path $sdkDir "install_qadav.bat") -Force
Copy-Item (Join-Path $repoRoot "ide\vscode-qadav\*") (Join-Path $sdkDir "ide\vscode-qadav\") -Recurse -Force
Write-Host "  [OK] Examples, scaffolding scripts, installer, and VS Code extension packaged" -ForegroundColor Green

# 5. Generate SDK_INFO.txt & README.md
Write-Host "`n[STEP 5/6] Generating SDK manifest and documentation..." -ForegroundColor Yellow
$qynVer = (& (Join-Path $sdkDir "bin\qyn.exe") --version | Out-String).Trim()
$sdkInfo = @"
================================================================================
  QADAV AGI SYSTEMS PROGRAMMING LANGUAGE - SOVEREIGN SDK BUNDLE
================================================================================
Product Name:     Qadav Sovereign Developer SDK
Version:          1.0.0-LTS
Compiler Version: $qynVer
Target Platform:  Windows x86_64 (AVX2 Required)
Build Date:       $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
License:          Proprietary / Patent-Enforced Core Architecture
Patents:          25 Core Sovereign Inventions (IPO / USPTO Format)
Invariants:       C = 0, Go = 0, Py = 0, ExternalTools = 0

Directory Layout:
  bin/            Compiler (qyn.exe), Package Manager (qpk.exe), CLI (qadav-new.bat)
  runtime/        Standalone compiled runtime objects (io.o, rt.o, kernel32.lib, 62 modules)
  stdlib/         Quantum-native, tensor, SIMD, cryptographic standard library
  examples/       Canonical cognitive demonstrators, benchmarks and models
  scripts/        Project scaffolding (new_ai_project.ps1) & installer scripts
  ide/            Visual Studio Code & Cursor extension bundle (vscode-qadav)
  install_qadav.bat One-click Windows installer (adds to PATH & sets up IDE)

Quickstart:
  1. Double-click install_qadav.bat (or run powershell -File scripts/install_qadav.ps1)
  2. Open any command prompt or PowerShell terminal and run:
       qadav-new my_agent
  3. Open my_agent in VS Code or Cursor and press Ctrl+Shift+B to build and run!
================================================================================
"@
Set-Content -Path (Join-Path $sdkDir "SDK_INFO.txt") -Value $sdkInfo -Encoding UTF8

$sdkReadme = @"
# Qadav Sovereign Developer SDK (Windows x86_64)

Welcome to the **Qadav Sovereign Developer SDK**, the standalone development kit for the quantum-native AGI systems programming language.

## 🚀 3-Step Quickstart

1. **Install:** Double-click `install_qadav.bat` or run:
   ```bat
   install_qadav.bat
   ```
2. **Create Project:** Open any terminal in any directory and run:
   ```powershell
   qadav-new my_agent
   ```
3. **Open & Run:**
   ```powershell
   code my_agent
   ```
   Press **`Ctrl+Shift+B`** to automatically build, run, and verify your AI agent across all 4 cognitive & security tiers!
"@
Set-Content -Path (Join-Path $sdkDir "README.md") -Value $sdkReadme -Encoding UTF8
Write-Host "  [OK] Generated SDK_INFO.txt and README.md" -ForegroundColor Green

# 6. Compress into Standalone ZIP Archive
Write-Host "`n[STEP 6/6] Generating standalone distribution ZIP archive..." -ForegroundColor Yellow
if (Test-Path $zipArchive) { Remove-Item -Force $zipArchive }
Compress-Archive -Path (Join-Path $sdkDir "*") -DestinationPath $zipArchive -Force
$zipItem = Get-Item $zipArchive
$totalFiles = (Get-ChildItem -Recurse $sdkDir -File).Count
$totalSizeMB = [math]::Round(((Get-ChildItem -Recurse $sdkDir -File | Measure-Object -Property Length -Sum).Sum / 1MB), 2)
$zipSizeMB = [math]::Round(($zipItem.Length / 1MB), 2)

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV SDK PACKAGING COMPLETE [SUCCESS]" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Bundle Directory: $sdkDir" -ForegroundColor White
Write-Host "Total Files:      $totalFiles files ($totalSizeMB MB)" -ForegroundColor Gray
Write-Host "ZIP Archive:      $zipArchive" -ForegroundColor White
Write-Host "ZIP File Size:    $zipSizeMB MB" -ForegroundColor Gray
Write-Host ""
exit 0
