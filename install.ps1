# device-sense PowerShell installer
# Downloads the pre-built .exe from GitHub Releases — no Go or bash required.
# Usage: irm https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.ps1 | iex

$REPO    = "jeffreyon/device-sense"
$EXE     = "device-sense-windows.exe"
$INSTALL = "$env:USERPROFILE\.device-sense"
$TARGET  = "$INSTALL\device-sense.exe"

Write-Host ""
Write-Host "  Installing device-sense..." -ForegroundColor Cyan
Write-Host ""

# ── Fetch latest release download URL ─────────────────────────────────────────
$api = "https://api.github.com/repos/$REPO/releases/latest"
try {
    $release = Invoke-RestMethod -Uri $api -UseBasicParsing
} catch {
    Write-Host "  Error: could not reach GitHub API. Check your connection." -ForegroundColor Red
    exit 1
}

$asset = $release.assets | Where-Object { $_.name -eq $EXE }
if (-not $asset) {
    Write-Host "  Error: no Windows binary found in the latest release." -ForegroundColor Red
    Write-Host "  Make sure a release has been published at https://github.com/$REPO/releases" -ForegroundColor Yellow
    exit 1
}

# ── Download ───────────────────────────────────────────────────────────────────
if (-not (Test-Path $INSTALL)) {
    New-Item -ItemType Directory -Path $INSTALL | Out-Null
}

Write-Host "  Downloading $($asset.name) ..." -ForegroundColor DarkGray
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $TARGET -UseBasicParsing
Write-Host "  Saved to $TARGET" -ForegroundColor Green

# ── Add to user PATH (permanent) ──────────────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$INSTALL*") {
    [Environment]::SetEnvironmentVariable("PATH", "$INSTALL;$userPath", "User")
    $env:PATH = "$INSTALL;$env:PATH"
    Write-Host "  Added $INSTALL to your PATH" -ForegroundColor Green
}

# ── Done ──────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  Done! Run it:" -ForegroundColor White
Write-Host ""
Write-Host "    device-sense" -ForegroundColor Yellow
Write-Host ""
Write-Host "  (Open a new terminal if the command isn't found yet)" -ForegroundColor DarkGray
Write-Host ""
