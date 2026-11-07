$go = Get-Content TMP\go_codegen.ll
$sh = Get-Content TMP\sh_codegen.ll
$min = [Math]::Min($go.Length, $sh.Length)
for ($i = 0; $i -lt $min; $i++) {
    if ($go[$i] -ne $sh[$i]) {
        Write-Host "First diff at line $($i + 1):"
        Write-Host "GO: $($go[$i])" -ForegroundColor Yellow
        Write-Host "SH: $($sh[$i])" -ForegroundColor Cyan
        for ($j = [Math]::Max(0, $i - 3); $j -lt [Math]::Min($min, $i + 6); $j++) {
            Write-Host "[$($j+1)] GO: $($go[$j])"
            Write-Host "[$($j+1)] SH: $($sh[$j])"
        }
        break
    }
}

