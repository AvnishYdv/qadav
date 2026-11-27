# scripts/test_commercial_sdk.ps1
# Qadav Commercial Developer Experience & Toolchain End-to-End Test Suite
# Pure PowerShell (Zero Python, Zero C, Zero Go)
# Validates commercial readiness: Scaffolding, Multi-target Compilation, Engine Execution & Security

[CmdletBinding()]
param ()

$ErrorActionPreference = 'Stop'
$rootDir = (Get-Item -Path $PSScriptRoot).Parent.FullName
Set-Location $rootDir

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV COMMERCIAL DEVELOPER SUITE & SDK READINESS AUDIT (M13.10)" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Host System:    $([System.Environment]::OSVersion)" -ForegroundColor Gray
Write-Host "Root Directory: $rootDir" -ForegroundColor Gray
Write-Host "Timestamp:      $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host ""

$swTotal = [System.Diagnostics.Stopwatch]::StartNew()
$stagesPassed = 0

# ------------------------------------------------------------------------------
# STAGE 1: ZERO-DEPENDENCY INVARIANT & DOCTOR AUDIT
# ------------------------------------------------------------------------------
Write-Host "[STAGE 1: ZERO-DEPENDENCY INVARIANTS & DOCTOR CHECK]" -ForegroundColor White
$cCount = (Get-ChildItem -Recurse -Filter *.c).Count
$ydvNative = (Get-ChildItem internal\qadav_native\*.ydv).Count
$ydvSelfhost = (Get-ChildItem selfhost\*.ydv).Count
$goCount = (Get-ChildItem -Recurse -Filter *.go).Count
$pyCount = (Get-ChildItem -Recurse -Filter *.py).Count
$patents = (Select-String -Path docs\PATENTS.md -Pattern '^## \d+\.').Count

if ($cCount -eq 0 -and $ydvNative -eq 62 -and $ydvSelfhost -eq 6 -and $goCount -eq 0 -and $pyCount -eq 0 -and $patents -eq 25) {
    Write-Host "  [PASS] All 6 Core Invariants Verified (C=0, Go=0, Py=0, Native=62, Selfhost=6, Patents=25)" -ForegroundColor Green
    $stagesPassed++
} else {
    throw "Stage 1 Failed: Invariants check mismatch"
}

# ------------------------------------------------------------------------------
# STAGE 2: AUTOMATED AI PROJECT SCAFFOLDING & VERIFICATION
# ------------------------------------------------------------------------------
Write-Host ""
Write-Host "[STAGE 2: PROJECT SCAFFOLDING & DEVELOPER WORKFLOW]" -ForegroundColor White
$demoProjectDir = Join-Path $rootDir "TMP\projects\commercial_test_agent"
if (Test-Path $demoProjectDir) { Remove-Item -Recurse -Force $demoProjectDir }

& .\scripts\new_ai_project.ps1 -ProjectName "commercial_test_agent" -Destination "TMP\projects" -RunVerification | Out-Null
$testExe = Join-Path $demoProjectDir "bin\commercial_test_agent.exe"
if (Test-Path $testExe) {
    Write-Host "  [PASS] Automated Scaffolding & One-Step Build Verification Succeeded" -ForegroundColor Green
    $stagesPassed++
} else {
    throw "Stage 2 Failed: Scaffolded binary was not generated"
}

# ------------------------------------------------------------------------------
# STAGE 3: MULTI-TARGET COMPILATION (PE32+, WASM32, LLVM IR)
# ------------------------------------------------------------------------------
Write-Host ""
Write-Host "[STAGE 3: MULTI-TARGET BACKEND COMPILATION (PE32+ / WASM32 / LLVM IR)]" -ForegroundColor White

# 3.1 Typecheck
& .\bin\qyn.exe check examples\calc.ydv | Out-Null
if ($LASTEXITCODE -ne 0) { throw "qyn check failed" }
Write-Host "  [PASS] Typecheck Engine: examples\calc.ydv Verified (Exit 0)" -ForegroundColor Green

# 3.2 Native Standalone PE32+
& .\bin\qyn.exe build examples\calc.ydv -o TMP\calc_standalone.exe | Out-Null
if ($LASTEXITCODE -ne 0 -or -not (Test-Path "TMP\calc_standalone.exe")) { throw "qyn build failed" }
Write-Host "  [PASS] Standalone PE32+ Emitter: TMP\calc_standalone.exe Generated" -ForegroundColor Green

# 3.3 WebAssembly Wasm32
& .\bin\qyn.exe emit-wasm examples\calc.ydv -o TMP\calc_target.wasm | Out-Null
if ($LASTEXITCODE -ne 0 -or -not (Test-Path "TMP\calc_target.wasm")) { throw "qyn emit-wasm failed" }
$wasmBytes = (Get-Item "TMP\calc_target.wasm").Length
Write-Host "  [PASS] WebAssembly Wasm32 Emitter: TMP\calc_target.wasm ($wasmBytes bytes)" -ForegroundColor Green

# 3.4 LLVM IR
$llvmText = & .\bin\qyn.exe emit-llvm examples\calc.ydv
if ($llvmText.Length -gt 100) {
    Write-Host "  [PASS] LLVM IR Target: Successfully Generated Clean IR Stream" -ForegroundColor Green
    $stagesPassed++
} else {
    throw "Stage 3 Failed: LLVM IR stream too short"
}

# ------------------------------------------------------------------------------
# STAGE 4: CONNECTED COGNITIVE & DELTA-RESONANCE ENGINE BENCHMARK
# ------------------------------------------------------------------------------
Write-Host ""
Write-Host "[STAGE 4: CONNECTED COGNITIVE & HIGH-THROUGHPUT ENGINE EXECUTION]" -ForegroundColor White

# 4.1 Unified Engine Execution
$swUnified = [System.Diagnostics.Stopwatch]::StartNew()
& .\TMP\unified_full.exe | Out-Null
$swUnified.Stop()
if ($LASTEXITCODE -ne 0) { throw "Unified engine failed" }
Write-Host "  [PASS] Unified QNMF + ECRA + Vishwanath Cognitive Pipeline Executed in $($swUnified.ElapsedMilliseconds) ms" -ForegroundColor Green

# 4.2 10,000-Record Delta-Resonance Stress Benchmark
$sw10k = [System.Diagnostics.Stopwatch]::StartNew()
& .\TMP\delta_resonance_full.exe | Out-Null
$sw10k.Stop()
if ($LASTEXITCODE -ne 0) { throw "10k Delta-Resonance benchmark failed" }
$throughput = [math]::Round((10000.0 / ($sw10k.ElapsedMilliseconds / 1000.0)), 2)
Write-Host "  [PASS] 10,000-Record Delta-Resonance Benchmark: $throughput records/sec ($($sw10k.ElapsedMilliseconds) ms)" -ForegroundColor Green
$stagesPassed++

# ------------------------------------------------------------------------------
# STAGE 5: MEMORY DISCIPLINE & HARDWARE FAIL-STOP SECURITY TRAP
# ------------------------------------------------------------------------------
Write-Host ""
Write-Host "[STAGE 5: MEMORY DISCIPLINE & HARDWARE FAIL-STOP SECURITY TRAP]" -ForegroundColor White

# Normal run
$pNormal = Start-Process -FilePath ".\TMP\delta_resonance_full.exe" -NoNewWindow -PassThru -Wait
if ($pNormal.ExitCode -ne 0) { throw "Normal execution returned non-zero code" }
Write-Host "  [PASS] Normal Certified Execution: Clean Exit Code 0" -ForegroundColor Green

# Tamper attack run
$pTamper = Start-Process -FilePath ".\TMP\product_ttpcr_tamper.exe" -NoNewWindow -PassThru -Wait
if ($pTamper.ExitCode -eq -559038737) {
    Write-Host "  [PASS] Hardware Fail-Stop Anti-Tamper: ExitProcess(0xDEADBEEF) / -559038737 Verified" -ForegroundColor Green
    $stagesPassed++
} else {
    throw "Stage 5 Failed: Tamper abort exit code mismatch: $($pTamper.ExitCode)"
}

$swTotal.Stop()
Write-Host ""
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  COMMERCIAL AUDIT COMPLETE: 5/5 STAGES PASSED ($($swTotal.ElapsedMilliseconds) ms TOTAL)" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
