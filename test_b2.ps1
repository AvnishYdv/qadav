$tests = @('hello', 'calc', 'if_else', 'fib', 'dot_product', 'composite', 'classify_operator', 'classify_annotation', 'classify_keyword', 'match_n_arms_proof')
$all_pass = $true
foreach ($t in $tests) {
    $f = "examples\$t.ydv"
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
Write-Host "Step B2 All Pass: $all_pass"

