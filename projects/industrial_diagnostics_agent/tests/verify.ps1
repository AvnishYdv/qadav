# projects/industrial_diagnostics_agent/tests/verify.ps1
# Automated Multi-Tier Verification Suite for Industrial Diagnostics Agent
# Verifies: Exit 0 (Nominal), Exit 100 (Abstain), Exit 101 (Rollback), Exit -559038737 (Tamper Fail-Stop)

$ErrorActionPreference = "Stop"
$projRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$repoRoot = Resolve-Path (Join-Path $projRoot "..\..")
Set-Location $projRoot

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  INDUSTRIAL DIAGNOSTICS AGENT: AUTOMATED VERIFICATION SUITE" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan

$binDir = Join-Path $projRoot "bin"
if (-not (Test-Path $binDir)) { New-Item -ItemType Directory -Path $binDir -Force | Out-Null }
$tmpDir = Join-Path $repoRoot "TMP"

$qyn = Join-Path $repoRoot "bin\qyn.exe"
$rtObjs = "$tmpDir\runtime_objs\io.o $tmpDir\runtime_objs\rt.o"
$kernelLib = "C:\Users\avnis\.qadav\runtime\kernel32.lib"

function Compile-YdvTarget([string]$srcFile, [string]$outFile) {
    $srcPath = Join-Path $projRoot $srcFile
    $base = [System.IO.Path]::GetFileNameWithoutExtension($srcFile)
    $llPath = Join-Path $tmpDir "diag_$base.ll"
    $objPath = Join-Path $tmpDir "diag_$base.o"
    $exePath = Join-Path $projRoot $outFile

    cmd.exe /c "`"$qyn`" emit-llvm `"$srcPath`" > `"$llPath`" && clang -c -O2 `"$llPath`" -o `"$objPath`" && clang -O2 `"$objPath`" $rtObjs -nostdlib -fuse-ld=lld -Wl,-entry:qadav_start -Wl,-subsystem:console `"$kernelLib`" -o `"$exePath`""
    if ($LASTEXITCODE -ne 0) {
        throw "Compilation failed for $srcFile"
    }
}

# 1. Typecheck Core Sources
Write-Host "`n[PHASE 1: TYPECHECK & SYNTAX AUDIT]" -ForegroundColor Yellow
& "$qyn" check (Join-Path $projRoot "src\main.ydv") | Out-Null
if ($LASTEXITCODE -ne 0) { throw "Typecheck failed on src/main.ydv" }
Write-Host "  [PASS] Zero type errors detected across agent modules" -ForegroundColor Green

# 2. Build All 4 Targets
Write-Host "`n[PHASE 2: COMPILING STANDALONE PRODUCTION EXECUTABLES]" -ForegroundColor Yellow
Compile-YdvTarget "src\main.ydv" "bin\main.exe"
Compile-YdvTarget "src\agent_abstain.ydv" "bin\abstain.exe"
Compile-YdvTarget "src\agent_rollback.ydv" "bin\rollback.exe"
Compile-YdvTarget "src\agent_tamper.ydv" "bin\tamper.exe"
Write-Host "  [PASS] Successfully compiled 4 standalone PE32+ executables" -ForegroundColor Green

# 3. Test Tier 1: Nominal Execution (Exit 0)
Write-Host "`n[PHASE 3: TIER 1 - NOMINAL STREAM EXECUTION]" -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$mainOut = & (Join-Path $binDir "main.exe")
$mainEc = $LASTEXITCODE
$sw.Stop()

Write-Host ($mainOut -join "`n")
if ($mainEc -ne 0) {
    throw "Tier 1 Nominal execution failed with exit code $mainEc (Expected: 0)"
}
Write-Host "  [PASS] Tier 1 Nominal Execution Verified: Exit Code 0 in $($sw.ElapsedMilliseconds) ms" -ForegroundColor Green

# 4. Test Tier 2: Calibrated Epistemic Risk Abstention (Exit 100)
Write-Host "`n[PHASE 4: TIER 2 - CALIBRATED RISK ABSTENTION (ECRA)]" -ForegroundColor Yellow
$abstainOut = & (Join-Path $binDir "abstain.exe")
$abstainEc = $LASTEXITCODE
Write-Host ($abstainOut -join "`n")
if ($abstainEc -ne 100) {
    throw "Tier 2 Abstention check failed! Expected exit code 100, got $abstainEc"
}
Write-Host "  [PASS] Tier 2 Calibrated Abstention Verified: Exit Code 100 [ABSTAIN]" -ForegroundColor Green

# 5. Test Tier 3: Invariant Contradiction Rollback (Exit 101)
Write-Host "`n[PHASE 5: TIER 3 - INVARIANT CONTRADICTION ROLLBACK]" -ForegroundColor Yellow
$rollbackOut = & (Join-Path $binDir "rollback.exe")
$rollbackEc = $LASTEXITCODE
Write-Host ($rollbackOut -join "`n")
if ($rollbackEc -ne 101) {
    throw "Tier 3 Contradiction rollback check failed! Expected exit code 101, got $rollbackEc"
}
Write-Host "  [PASS] Tier 3 Contradiction Rollback Verified: Exit Code 101 [ROLLBACK]" -ForegroundColor Green

# 6. Test Tier 4: Hardware Fail-Stop Anti-Tamper Abort (Exit -559038737 / 0xDEADBEEF)
Write-Host "`n[PHASE 6: TIER 4 - HARDWARE FAIL-STOP ANTI-TAMPER]" -ForegroundColor Yellow
$tamperOut = & (Join-Path $binDir "tamper.exe")
$tamperEc = $LASTEXITCODE
Write-Host ($tamperOut -join "`n")
if ($tamperEc -ne -559038737) {
    throw "Tier 4 Hardware tamper trap failed! Expected -559038737 (0xDEADBEEF), got $tamperEc"
}
Write-Host "  [PASS] Tier 4 Hardware Fail-Stop Trap Verified: ExitProcess(0xDEADBEEF) / $tamperEc" -ForegroundColor Green

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  ALL 4 EXECUTION TIERS VERIFIED: [100% PASS - COMMERCIALLY READY]" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
exit 0
