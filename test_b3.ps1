$targets = @('lexer', 'parser', 'checker', 'codegen')
$all_pass = $true
foreach ($t in $targets) {
    $f = "selfhost\$t.ydv"
    Write-Host "Comparing $f..."
    cmd /c "bin\qyn.exe emit-llvm $f > TMP\go_$t.ll"
    cmd /c "selfhost\codegen.exe $f > TMP\sh_$t.ll"
    $res = cmd /c "fc /b TMP\go_$t.ll TMP\sh_$t.ll"
    if ($res -match "no differences encountered") {
        Write-Host "[PASS] $t : Byte-identical" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] $t : Differences found" -ForegroundColor Red
        Write-Host $res
        $all_pass = $false
    }
}
Write-Host "Step B3 All Pass: $all_pass"

