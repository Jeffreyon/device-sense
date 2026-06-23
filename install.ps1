# device-sense PowerShell installer
# Usage: irm https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.ps1 | iex

$REPO_RAW   = "https://raw.githubusercontent.com/jeffreyon/device-sense/main"
$INSTALL_DIR = "$HOME\.device-sense"
$SCRIPT      = "$INSTALL_DIR\device-sense.sh"

Write-Host ""
Write-Host "  Installing device-sense..." -ForegroundColor Cyan
Write-Host ""

# ── Check bash is available ────────────────────────────────────────────────────
if (-not (Get-Command bash -ErrorAction SilentlyContinue)) {
    Write-Host "  Error: bash not found. Install Git for Windows first:" -ForegroundColor Red
    Write-Host "  https://git-scm.com/download/win" -ForegroundColor Yellow
    exit 1
}

# ── Download script ────────────────────────────────────────────────────────────
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
}

Invoke-WebRequest -Uri "$REPO_RAW/device-sense.sh" -OutFile $SCRIPT -UseBasicParsing
Write-Host "  Downloaded to $SCRIPT" -ForegroundColor Green

# ── Add function to PowerShell profile ────────────────────────────────────────
$funcBlock = @"

# device-sense — hardware verifier
function device-sense { bash "$($SCRIPT -replace '\\','/')" }
"@

if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}

if (-not (Select-String -Path $PROFILE -Pattern "device-sense" -Quiet -ErrorAction SilentlyContinue)) {
    Add-Content -Path $PROFILE -Value $funcBlock
    Write-Host "  Added 'device-sense' function to $PROFILE" -ForegroundColor Green
} else {
    Write-Host "  Profile already has device-sense — skipping" -ForegroundColor DarkGray
}

# ── Done ──────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  Done! Reload your profile then run it:" -ForegroundColor White
Write-Host ""
Write-Host "    . `$PROFILE" -ForegroundColor Yellow
Write-Host "    device-sense" -ForegroundColor Yellow
Write-Host ""
