# generate_telemetry.ps1
param (
    [string]$OutFile = "$PSScriptRoot/telemetry_stream.txt",
    [int]$RecordCount = 5120
)

$lines = [System.Collections.Generic.List[string]]::new()
$lines.Add("# Industrial Diagnostics Agent Telemetry Stream (M13.14)")
$lines.Add("# $RecordCount Structured Sensor Telemetry Records across $([math]::Ceiling($RecordCount / 64)) Batches (Batch Size: 64)")
$lines.Add("# Format: StepIdx|BatchIdx|SensorID|VibrationHash|ThermalEntropy|Coherence|Confidence|SigHash|ExpectedStatus")

for ($i = 1; $i -le $RecordCount; $i++) {
    $batch = [math]::Floor(($i - 1) / 64) + 1
    $sensorId = "TB_" + $i.ToString("D6")
    $vibe = 7000 + ($i * 13 % 1000)
    $entropy = 120 + ($i * 7 % 200)
    $coherence = 920 + ($i % 40)
    $conf = 88 + ($i % 12)
    $sig = 5000000 + $i
    $status = "NOMINAL_MONITORED"
    $lines.Add("$i|$batch|$sensorId|$vibe|$entropy|$coherence|$conf|$sig|$status")
}

[System.IO.File]::WriteAllLines($OutFile, $lines)
Write-Host "Generated $RecordCount telemetry stream records at $OutFile" -ForegroundColor Green
