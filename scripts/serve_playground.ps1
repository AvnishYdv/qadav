# scripts/serve_playground.ps1 - Pure PowerShell HTTP Server with COOP/COEP for WebAssembly Threads
param(
    [int]$Port = 8080,
    [switch]$OneShot = $false
)

$PSScriptRootLocal = Split-Path -Parent $MyInvocation.MyCommand.Definition
$RepoRoot = Split-Path -Parent $PSScriptRootLocal
$PlaygroundDir = Join-Path $RepoRoot "playground"

if (-not (Test-Path $PlaygroundDir)) {
    Write-Error "Playground directory not found at $PlaygroundDir"
    exit 1
}

$listener = New-Object System.Net.HttpListener
$prefix = "http://localhost:$Port/"
$listener.Prefixes.Add($prefix)

try {
    $listener.Start()
} catch {
    Write-Error "Failed to start HttpListener on $prefix : $_"
    exit 1
}

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  QADAV WEBASSEMBLY PLAYGROUND HTTP SERVER (Pure PowerShell)" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  URL:         http://localhost:$Port/" -ForegroundColor Yellow
Write-Host "  Directory:   $PlaygroundDir" -ForegroundColor White
Write-Host "  Headers:     Cross-Origin-Opener-Policy: same-origin" -ForegroundColor Gray
Write-Host "               Cross-Origin-Embedder-Policy: require-corp" -ForegroundColor Gray
Write-Host "  Status:      Ready for SharedArrayBuffer & WebAssembly Threads (Atomics)" -ForegroundColor Green
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "Press Ctrl+C to stop the server." -ForegroundColor White

$mimeTypes = @{
    ".html" = "text/html; charset=utf-8"
    ".js"   = "application/javascript; charset=utf-8"
    ".css"  = "text/css; charset=utf-8"
    ".wasm" = "application/wasm"
    ".json" = "application/json; charset=utf-8"
    ".png"  = "image/png"
    ".ico"  = "image/x-icon"
}

try {
    while ($listener.IsListening) {
        $context = $listener.GetContext()
        $request = $context.Request
        $response = $context.Response

        # Add mandatory headers for SharedArrayBuffer / multi-threaded Web Workers
        $response.Headers.Add("Cross-Origin-Opener-Policy", "same-origin")
        $response.Headers.Add("Cross-Origin-Embedder-Policy", "require-corp")
        $response.Headers.Add("Access-Control-Allow-Origin", "*")

        $urlPath = $request.Url.LocalPath
        if ($urlPath -eq "/" -or [string]::IsNullOrEmpty($urlPath)) {
            $urlPath = "/index.html"
        }

        $localFilePath = Join-Path $PlaygroundDir ($urlPath.TrimStart("/").Replace("/", [System.IO.Path]::DirectorySeparatorChar))

        if (Test-Path $localFilePath -PathType Leaf) {
            $ext = [System.IO.Path]::GetExtension($localFilePath).ToLower()
            $contentType = if ($mimeTypes.ContainsKey($ext)) { $mimeTypes[$ext] } else { "application/octet-stream" }
            $response.ContentType = $contentType

            $bytes = [System.IO.File]::ReadAllBytes($localFilePath)
            $response.ContentLength64 = $bytes.Length
            if ($request.HttpMethod -ne "HEAD") {
                $response.OutputStream.Write($bytes, 0, $bytes.Length)
            }
            $response.StatusCode = 200
            Write-Host "[200 OK] $urlPath ($($bytes.Length) bytes) [$contentType]" -ForegroundColor Green
        } else {
            $response.StatusCode = 404
            $errBytes = [System.Text.Encoding]::UTF8.GetBytes("404 Not Found: $urlPath")
            $response.ContentLength64 = $errBytes.Length
            if ($request.HttpMethod -ne "HEAD") {
                $response.OutputStream.Write($errBytes, 0, $errBytes.Length)
            }
            Write-Host "[404 Not Found] $urlPath" -ForegroundColor Red
        }

        $response.OutputStream.Close()

        if ($OneShot) {
            break
        }
    }
} finally {
    $listener.Stop()
    $listener.Close()
    Write-Host "`nServer stopped cleanly." -ForegroundColor Yellow
}
