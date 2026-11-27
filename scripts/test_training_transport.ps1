# scripts/test_training_transport.ps1
# Qadav Large-Batch Training & Transport Verification Suite (Milestone M13.11)
# Pure PowerShell, Zero C, Zero Go, Zero Python

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV TRAINING & TRANSPORT SUITE (M13.11)" -ForegroundColor Cyan
Write-Host "================================================================================" -ForegroundColor Cyan

# STAGE 1: CORE INVARIANTS
Write-Host "`n[STAGE 1: CORE INVARIANTS AUDIT]" -ForegroundColor Yellow
$c = (Get-ChildItem -Recurse -Filter *.c).Count
$native = (Get-ChildItem internal\qadav_native\*.ydv).Count
$selfhost = (Get-ChildItem selfhost\*.ydv).Count
$go = (Get-ChildItem -Recurse -Filter *.go).Count
$py = (Get-ChildItem -Recurse -Filter *.py).Count
$patents = (Select-String -Path docs\PATENTS.md -Pattern '^## \d+\.').Count

if ($c -ne 0 -or $native -ne 62 -or $selfhost -ne 6 -or $go -ne 0 -or $py -ne 0 -or $patents -ne 25) {
    Write-Host "Invariant violation: C=$c, Native=$native, Selfhost=$selfhost, Go=$go, Py=$py, Patents=$patents" -ForegroundColor Red
    exit 1
}
Write-Host "  [PASS] All 6 Core Invariants Intact (C=0, Go=0, Py=0, Native=62, Selfhost=6, Patents=25)" -ForegroundColor Green

# STAGE 2: BUILD TRAINING & TRANSPORT ENGINES
Write-Host "`n[STAGE 2: COMPILATION & STANDALONE EMISSION]" -ForegroundColor Yellow
cmd.exe /c "bin\qyn.exe emit-llvm examples\training_transport_engine.ydv > TMP\training_transport_raw.ll && clang -c -O2 TMP\training_transport_raw.ll -o TMP\training_transport.o && clang -O2 TMP\training_transport.o TMP\runtime_objs\io.o TMP\runtime_objs\rt.o -nostdlib -fuse-ld=lld -Wl,-entry:qadav_start -Wl,-subsystem:console C:\Users\avnis\.qadav\runtime\kernel32.lib -o TMP\training_transport_full.exe"
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to compile training_transport_full.exe" -ForegroundColor Red; exit 1 }

cmd.exe /c "bin\qyn.exe emit-llvm examples\training_transport_tamper.ydv > TMP\training_transport_tamper_raw.ll && clang -c -O2 TMP\training_transport_tamper_raw.ll -o TMP\training_transport_tamper.o && clang -O2 TMP\training_transport_tamper.o TMP\runtime_objs\io.o TMP\runtime_objs\rt.o -nostdlib -fuse-ld=lld -Wl,-entry:qadav_start -Wl,-subsystem:console C:\Users\avnis\.qadav\runtime\kernel32.lib -o TMP\training_transport_tamper.exe"
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to compile training_transport_tamper.exe" -ForegroundColor Red; exit 1 }
Write-Host "  [PASS] Compiled Standalone PE32+ Binaries (Normal + Tamper)" -ForegroundColor Green

# STAGE 3: 10,240-RECORD BATCH TRAINING EXECUTION
Write-Host "`n[STAGE 3: 10,240-RECORD BATCH BENCHMARK & SIMD SYNC]" -ForegroundColor Yellow
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$output = & .\TMP\training_transport_full.exe
$ec = $LASTEXITCODE
$sw.Stop()

if ($ec -ne 0) { Write-Host "Training engine failed with exit code $ec" -ForegroundColor Red; exit 1 }
$ws = (Get-Process -Id $PID).WorkingSet64 / 1MB
Write-Host ($output -join "`n")
Write-Host "  [METRICS] Wall Time: $($sw.ElapsedMilliseconds) ms | Host RSS: $([math]::Round($ws, 2)) MB | Exit: $ec" -ForegroundColor Green

# STAGE 4: HARDWARE FAIL-STOP SECURITY TEST
Write-Host "`n[STAGE 4: HARDWARE FAIL-STOP SECURITY TRAP]" -ForegroundColor Yellow
& .\TMP\training_transport_tamper.exe | Out-Null
$tamperEc = $LASTEXITCODE
if ($tamperEc -ne -559038737) {
    Write-Host "Tamper check failed! Expected -559038737 (0xDEADBEEF), got $tamperEc" -ForegroundColor Red
    exit 1
}
Write-Host "  [PASS] Hardware Fail-Stop Interception Trapped: ExitProcess(0xDEADBEEF) / $tamperEc" -ForegroundColor Green

# STAGE 5: REGRESSION & PRODUCT DEMOS
Write-Host "`n[STAGE 5: REGRESSION & PRODUCT DEMOS GATE]" -ForegroundColor Yellow
$fcPass = $true
$fcTests = @('hello', 'calc', 'if_else', 'fib', 'dot_product', 'composite', 'classify_operator', 'classify_annotation', 'classify_keyword', 'match_n_arms_proof')
foreach ($t in $fcTests) {
    cmd.exe /c "bin\qyn.exe emit-llvm examples\$t.ydv > TMP\reg_$t.ll"
    $ref = "TMP\ref_$t.ll"
    $reg = "TMP\reg_$t.ll"
    $out = cmd.exe /c "fc.exe /b $ref $reg"
    if (($out -join ' ') -notmatch 'no differences encountered') { $fcPass = $false }
}
if (-not $fcPass) { Write-Host "Bitwise regression failed!" -ForegroundColor Red; exit 1 }
Write-Host "  [PASS] 10/10 Phase 8 fc.exe /b Bitwise Regressions Passed (0 Differences)" -ForegroundColor Green

$demos = @('barat', 'cdaga', 'mctp', 'qecma', 'qemwv', 'qera', 'qesi', 'ttpcr')
foreach ($d in $demos) {
    & "TMP\product_$d.exe" | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Host "Demo product_$d.exe failed" -ForegroundColor Red; exit 1 }
}
Write-Host "  [PASS] 8/8 Product Demonstrators Passed (Exit Code 0)" -ForegroundColor Green

Write-Host "`n================================================================================" -ForegroundColor Cyan
Write-Host "  M13.11 VERIFICATION COMPLETE: ALL 5 STAGES PASSED SUCCESSFULLY" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
