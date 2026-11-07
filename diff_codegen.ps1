$c = Compare-Object (Get-Content TMP\go_codegen.ll) (Get-Content TMP\sh_codegen.ll) -SyncWindow 5
Write-Host "Total diff lines: $($c.Count)"
$c | Select-Object -First 40 | Format-Table -AutoSize

