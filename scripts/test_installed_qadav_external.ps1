# scripts/test_installed_qadav_external.ps1
# Qadav Milestone M13.13: Verification of Installed Toolchain from External Directory

$ErrorActionPreference = "Stop"

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV SYSTEM INSTALLATION & EXTERNAL PATH VERIFICATION" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan

# Refresh PATH from User environment
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$env:PATH = "$userPath;$env:PATH"

# 1. Verify PATH resolution for qyn, qpk, qadav-new
Write-Host "`n[PHASE 1: PATH RESOLUTION AUDIT]" -ForegroundColor Yellow

$qynCmd = Get-Command "qyn" -ErrorAction SilentlyContinue
if (-not $qynCmd) { throw "qyn.exe not found in PATH!" }
Write-Host "  [PASS] 'qyn' resolved to: $($qynCmd.Source)" -ForegroundColor Green

$qpkCmd = Get-Command "qpk" -ErrorAction SilentlyContinue
if (-not $qpkCmd) { throw "qpk.exe not found in PATH!" }
Write-Host "  [PASS] 'qpk' resolved to: $($qpkCmd.Source)" -ForegroundColor Green

$newCmd = Get-Command "qadav-new" -ErrorAction SilentlyContinue
if (-not $newCmd) { throw "qadav-new.bat not found in PATH!" }
Write-Host "  [PASS] 'qadav-new' resolved to: $($newCmd.Source)" -ForegroundColor Green

# 2. Test compiler version
Write-Host "`n[PHASE 2: COMPILER VERSION EXECUTION]" -ForegroundColor Yellow
$verOut = & qyn --version
Write-Host "  $verOut" -ForegroundColor Gray
if ($LASTEXITCODE -ne 0) { throw "qyn --version failed" }
Write-Host "  [PASS] Compiler executed cleanly via PATH" -ForegroundColor Green

# 3. Test Registry file type registration
Write-Host "`n[PHASE 3: WINDOWS REGISTRY FILE TYPE REGISTRATION AUDIT]" -ForegroundColor Yellow
$regVal = (Get-ItemProperty -Path "HKCU:\Software\Classes\.ydv" -ErrorAction SilentlyContinue).'(Default)'
$regType = (Get-ItemProperty -Path "HKCU:\Software\Classes\.ydv" -ErrorAction SilentlyContinue).'PerceivedType'
if ($regVal -ne "QadavSourceFile" -or $regType -ne "text") {
    throw "HKCU:\Software\Classes\.ydv registration missing or invalid (Val: $regVal, Type: $regType)"
}
Write-Host "  [PASS] HKCU:\Software\Classes\.ydv -> $regVal (PerceivedType: $regType)" -ForegroundColor Green

# 4. Test VS Code extension installation
Write-Host "`n[PHASE 4: IDE EXTENSION AUTO-DETECTION AUDIT]" -ForegroundColor Yellow
$vscodeExtPath = Join-Path $env:USERPROFILE ".vscode\extensions\qadav.qadav-lang-1.0.0"
if (-not (Test-Path $vscodeExtPath)) {
    throw "VS Code extension not found at $vscodeExtPath"
}
$extPkg = Join-Path $vscodeExtPath "package.json"
if (-not (Test-Path $extPkg)) { throw "package.json missing in $vscodeExtPath" }
Write-Host "  [PASS] VS Code extension auto-installed at: $vscodeExtPath" -ForegroundColor Green

# 5. External directory execution test
Write-Host "`n[PHASE 5: EXTERNAL DIRECTORY WORKFLOW EXECUTION]" -ForegroundColor Yellow
$extDir = Join-Path ([System.IO.Path]::GetTempPath()) "qadav_ext_test_$(Get-Random)"
New-Item -ItemType Directory -Path $extDir -Force | Out-Null
Set-Location $extDir
Write-Host "  Switched working directory to external temp path: $extDir" -ForegroundColor Gray

# Run qadav-new from external path
Write-Host "  Executing: qadav-new external_agent" -ForegroundColor Gray
& qadav-new external_agent -Destination .
if ($LASTEXITCODE -ne 0) { throw "qadav-new failed in external directory" }

$projDir = Join-Path $extDir "external_agent"
if (-not (Test-Path $projDir)) { throw "Project folder not created at $projDir" }

# Check that .vscode was created
if (-not (Test-Path (Join-Path $projDir ".vscode\tasks.json"))) {
    throw ".vscode/tasks.json missing in scaffolded project"
}
if (-not (Test-Path (Join-Path $projDir ".vscode\settings.json"))) {
    throw ".vscode/settings.json missing in scaffolded project"
}
Write-Host "  [PASS] Project scaffolded with .vscode/tasks.json and settings.json" -ForegroundColor Green

# Run automated tests in external project
Set-Location $projDir
Write-Host "  Running tests/verify.ps1 in external project..." -ForegroundColor Gray
$sw = [System.Diagnostics.Stopwatch]::StartNew()
powershell -ExecutionPolicy Bypass -File (Join-Path $projDir "tests\verify.ps1")
$ec = $LASTEXITCODE
$sw.Stop()

if ($ec -ne 0) {
    throw "External project verification failed with exit code $ec"
}
Write-Host "  [PASS] External project compiled, executed, and verified all 4 tiers in $($sw.ElapsedMilliseconds) ms" -ForegroundColor Green

# Cleanup temp dir
Set-Location $PSScriptRoot
Remove-Item -Recurse -Force $extDir -ErrorAction SilentlyContinue

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  INSTALLATION & EXTERNAL EXECUTION AUDIT: 100% PASS [COMMERCIALLY READY]" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
exit 0
