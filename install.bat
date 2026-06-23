@echo off
:: device-sense cmd installer
:: Usage: paste this URL into your browser to download, then double-click,
::        OR run:  curl -fsSL https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.bat -o install.bat && install.bat

setlocal enabledelayedexpansion

set INSTALL_DIR=%USERPROFILE%\.device-sense
set SCRIPT=%INSTALL_DIR%\device-sense.sh
set WRAPPER=%INSTALL_DIR%\device-sense.bat

echo.
echo   Installing device-sense...
echo.

:: ── Check bash is available ───────────────────────────────────────────────────
where bash >nul 2>&1
if %errorlevel% neq 0 (
    echo   Error: bash not found.
    echo   Install Git for Windows first: https://git-scm.com/download/win
    echo   Then re-run this installer.
    pause
    exit /b 1
)

:: ── Create install directory ──────────────────────────────────────────────────
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: ── Download the script ───────────────────────────────────────────────────────
curl -fsSL "https://raw.githubusercontent.com/jeffreyon/device-sense/main/device-sense.sh" -o "%SCRIPT%"
if %errorlevel% neq 0 (
    echo   Error: download failed. Check your internet connection.
    pause
    exit /b 1
)

echo   Downloaded to %SCRIPT%

:: ── Write a .bat wrapper that calls bash on the script ────────────────────────
(
    echo @echo off
    echo bash "%SCRIPT%"
) > "%WRAPPER%"

:: ── Add install dir to user PATH (permanent) via setx ────────────────────────
echo %PATH% | find /i "%INSTALL_DIR%" >nul 2>&1
if %errorlevel% neq 0 (
    setx PATH "%USERPROFILE%\.device-sense;%PATH%" >nul
    echo   Added %INSTALL_DIR% to PATH
)

:: ── Done ──────────────────────────────────────────────────────────────────────
echo.
echo   Done! Open a NEW Command Prompt window and run:
echo.
echo     device-sense
echo.
pause
