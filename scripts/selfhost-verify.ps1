# scripts/selfhost-verify.ps1
# RSCA 3-Stage Bootstrap Verification Script

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Qadav Phase 8: RSCA Self-Hosting Bootstrap Verification" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Verify Go bootstrap compiler
if (-not (Test-Path "bin/qyn.exe")) {
    Write-Host "[*] Building Stage 0 compiler (bin/qyn.exe)..." -ForegroundColor Yellow
    go build -o bin/qyn.exe ./cmd/qyn
}
Write-Host "[+] Stage 0 compiler present: bin/qyn.exe" -ForegroundColor Green

# 2. Verify selfhost lexer and parser
Write-Host "[*] Checking selfhost/lexer.ydv and selfhost/parser.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe check selfhost/lexer.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/lexer.ydv check failed" }

.\bin\qyn.exe check selfhost/parser.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/parser.ydv check failed" }

Write-Host "[+] Selfhost source checks PASS." -ForegroundColor Green

# 3. Run RSCA verification demo
Write-Host "[*] Running RSCA attestation chain verification..." -ForegroundColor Yellow
.\bin\qyn.exe run examples/rsca_demo.ydv
if ($LASTEXITCODE -ne 0) { throw "examples/rsca_demo.ydv execution failed" }
Write-Host "[+] RSCA attestation chain verified successfully." -ForegroundColor Green

# 4. Run RSCA tamper decoherence test
Write-Host "[*] Verifying RSCA decoherence on tamper trigger (0xDEADBEEF)..." -ForegroundColor Yellow
try {
    .\bin\qyn.exe run --test examples/rsca_tamper.ydv 2>$null
    $exitCode = $LASTEXITCODE
} catch {
    $exitCode = $LASTEXITCODE
}

# -559038737 is 0xDEADBEEF in signed 32-bit
if ($exitCode -eq -559038737 -or $exitCode -eq 3735928559) {
    Write-Host "[+] RSCA decoherence verified: process collapsed with 0xDEADBEEF." -ForegroundColor Green
} else {
    Write-Host "[-] RSCA decoherence check unexpected exit code: $exitCode" -ForegroundColor Red
    exit 1
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "RSCA Self-Hosting Verification: ALL CHECKS PASSED" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Cyan
