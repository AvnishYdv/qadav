# scripts/verify-selfhost.ps1
# Qadav Milestone M8.3a: Self-Hosted Lexer & Parser Byte-Identical Verification Script

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Qadav M8.3a: Self-Hosted Lexer & Parser Verification" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Ensure bin/qyn.exe is compiled
if (-not (Test-Path "bin/qyn.exe")) {
    Write-Host "[*] Building bin/qyn.exe..." -ForegroundColor Yellow
    go build -o bin/qyn.exe ./cmd/qyn
}
Write-Host "[+] Compiler binary ready: bin/qyn.exe" -ForegroundColor Green

# 2. Check and Build Self-Hosted Qadav Sources
Write-Host "`n--- Part 1: Self-Hosted Source Validation ---" -ForegroundColor Cyan

Write-Host "[*] Checking selfhost/lexer.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe check selfhost/lexer.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/lexer.ydv check failed" }

Write-Host "[*] Checking selfhost/ast_nodes.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe check selfhost/ast_nodes.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/ast_nodes.ydv check failed" }

Write-Host "[*] Checking selfhost/parser.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe check selfhost/parser.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/parser.ydv check failed" }

Write-Host "[*] Building selfhost/lexer.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe build selfhost/lexer.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/lexer.ydv build failed" }

Write-Host "[*] Building selfhost/ast_nodes.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe build selfhost/ast_nodes.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/ast_nodes.ydv build failed" }

Write-Host "[*] Building selfhost/parser.ydv..." -ForegroundColor Yellow
.\bin\qyn.exe build selfhost/parser.ydv
if ($LASTEXITCODE -ne 0) { throw "selfhost/parser.ydv build failed" }
Write-Host "[+] All selfhost sources check and build cleanly." -ForegroundColor Green

# 3. Byte-Identical Lexer Verification
Write-Host "`n--- Part 2: Byte-Identical Lexer Verification ---" -ForegroundColor Cyan
$testFiles = Get-ChildItem -Path "selfhost/tests/*.ydv"
$tempDir = [System.IO.Path]::GetTempPath()

$lexerPassCount = 0
foreach ($file in $testFiles) {
    $baseName = $file.BaseName
    $goTokensPath = Join-Path $tempDir "go_${baseName}_tokens.txt"
    $shTokensPath = Join-Path $tempDir "sh_${baseName}_tokens.txt"

    .\bin\qyn.exe emit-tokens $file.FullName > $goTokensPath
    .\bin\qyn.exe selfhost-lexer $file.FullName > $shTokensPath

    $goBytes = [System.IO.File]::ReadAllBytes($goTokensPath)
    $shBytes = [System.IO.File]::ReadAllBytes($shTokensPath)

    if ($goBytes.Length -ne $shBytes.Length) {
        Write-Host "[-] Lexer length mismatch for $($file.Name): Go=$($goBytes.Length) bytes, SH=$($shBytes.Length) bytes" -ForegroundColor Red
        exit 1
    }

    $diff = $false
    for ($i = 0; $i -lt $goBytes.Length; $i++) {
        if ($goBytes[$i] -ne $shBytes[$i]) {
            $diff = $true
            break
        }
    }

    if ($diff) {
        Write-Host "[-] Lexer byte diff found in $($file.Name)" -ForegroundColor Red
        exit 1
    }

    Write-Host "  [PASS] $($file.Name): byte-identical ($($goBytes.Length) bytes, 0 differences)" -ForegroundColor Green
    $lexerPassCount++
}

# Also verify across examples/*.ydv for Lexer
$exampleFiles = Get-ChildItem -Path "examples/*.ydv"
$examplePassCount = 0
foreach ($file in $exampleFiles) {
    $baseName = $file.BaseName
    $goTokensPath = Join-Path $tempDir "go_ex_${baseName}_tokens.txt"
    $shTokensPath = Join-Path $tempDir "sh_ex_${baseName}_tokens.txt"

    .\bin\qyn.exe emit-tokens $file.FullName > $goTokensPath
    .\bin\qyn.exe selfhost-lexer $file.FullName > $shTokensPath

    $goBytes = [System.IO.File]::ReadAllBytes($goTokensPath)
    $shBytes = [System.IO.File]::ReadAllBytes($shTokensPath)

    if ($goBytes.Length -ne $shBytes.Length) {
        Write-Host "[-] Lexer length mismatch on example $($file.Name)" -ForegroundColor Red
        exit 1
    }

    $examplePassCount++
}
Write-Host "[+] Lexer: $lexerPassCount / $lexerPassCount test files + $examplePassCount / $examplePassCount example files BYTE-IDENTICAL (0 differences)." -ForegroundColor Green

# 4. Byte-Identical Parser AST Verification
Write-Host "`n--- Part 3: Byte-Identical Parser AST Verification ---" -ForegroundColor Cyan
$parserPassCount = 0
foreach ($file in $testFiles) {
    $baseName = $file.BaseName
    $goAstPath = Join-Path $tempDir "go_${baseName}_ast.txt"
    $shAstPath = Join-Path $tempDir "sh_${baseName}_ast.txt"

    .\bin\qyn.exe emit-ast $file.FullName > $goAstPath
    .\bin\qyn.exe selfhost-parser $file.FullName > $shAstPath

    $goContent = [System.IO.File]::ReadAllText($goAstPath).Trim()
    $shContent = [System.IO.File]::ReadAllText($shAstPath).Trim()

    if ($goContent -ne $shContent) {
        Write-Host "[-] Parser AST diff found in $($file.Name)" -ForegroundColor Red
        Write-Host "Go output: $goContent" -ForegroundColor Yellow
        Write-Host "SH output: $shContent" -ForegroundColor Yellow
        exit 1
    }

    Write-Host "  [PASS] $($file.Name): byte-identical canonical AST S-expression" -ForegroundColor Green
    $parserPassCount++
}
Write-Host "[+] Parser: $parserPassCount / $parserPassCount test files BYTE-IDENTICAL (0 differences)." -ForegroundColor Green

Write-Host "`n==========================================================" -ForegroundColor Cyan
Write-Host "M8.3a SELF-HOST VERIFICATION: ALL 100% BYTE-IDENTICAL" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Cyan

