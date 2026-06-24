#!/bin/bash
# device-sense universal installer
# Detects OS, downloads the right binary, adds it to PATH.
#
# macOS / Linux:
#   curl -fsSL https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.sh | bash
#
# Windows (Git Bash or WSL):
#   curl -fsSL https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.sh | bash
#
# Windows (PowerShell — only if you have no bash at all):
#   irm https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.ps1 | iex

REPO="jeffreyon/device-sense"
INSTALL_DIR="$HOME/bin"

echo ""
echo "  device-sense installer"
echo ""

# ── Detect OS ─────────────────────────────────────────────────────────────────
OS_RAW="$(uname -s)"
case "$OS_RAW" in
    Darwin)              OS="macos";   ASSET="device-sense-macos"      ; EXT=""     ;;
    Linux)               OS="linux";   ASSET="device-sense-linux"      ; EXT=""     ;;
    MINGW*|MSYS*|CYGWIN*) OS="windows"; ASSET="device-sense-windows.exe"; EXT=".exe" ;;
    *)
        echo "  Unsupported OS: $OS_RAW"
        echo "  Visit https://github.com/$REPO/releases to download manually."
        exit 1
        ;;
esac

echo "  Detected: $OS"

# ── Fetch latest release URL for the right binary ─────────────────────────────
API="https://api.github.com/repos/$REPO/releases/latest"

if command -v curl &>/dev/null; then
    DOWNLOAD_URL=$(curl -fsSL "$API" | grep "browser_download_url" | grep "$ASSET" | cut -d '"' -f 4)
elif command -v wget &>/dev/null; then
    DOWNLOAD_URL=$(wget -qO- "$API" | grep "browser_download_url" | grep "$ASSET" | cut -d '"' -f 4)
else
    echo "  Error: curl or wget is required."
    exit 1
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo "  Error: no release found for $OS."
    echo "  Make sure a release exists at https://github.com/$REPO/releases"
    exit 1
fi

# ── Download binary ───────────────────────────────────────────────────────────
mkdir -p "$INSTALL_DIR"
TARGET="$INSTALL_DIR/device-sense$EXT"

echo "  Downloading $ASSET..."
if command -v curl &>/dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TARGET"
else
    wget -qO "$TARGET" "$DOWNLOAD_URL"
fi

chmod +x "$TARGET"
echo "  Saved to $TARGET"

# ── On Windows (Git Bash), also create a device-sense.cmd wrapper so it works
# ── in plain cmd.exe terminals opened after installation ──────────────────────
if [ "$OS" = "windows" ]; then
    CMD_WRAPPER="$INSTALL_DIR/device-sense.cmd"
    printf '@echo off\n"%~dp0device-sense.exe" %%*\n' > "$CMD_WRAPPER"
fi

# ── Add ~/bin to PATH ─────────────────────────────────────────────────────────
add_to_rc() {
    local rc="$1"
    [ -f "$rc" ] && grep -q 'device-sense' "$rc" 2>/dev/null && return
    printf '\n# added by device-sense installer\nexport PATH="$HOME/bin:$PATH"\n' >> "$rc"
}

if [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
    add_to_rc "$HOME/.bashrc"
    add_to_rc "$HOME/.bash_profile"
    add_to_rc "$HOME/.zshrc"
    export PATH="$HOME/bin:$PATH"
    echo "  Added ~/bin to PATH"
fi

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "  Done! Run:"
echo ""
echo "    device-sense"
echo ""
if [ "$OS" = "windows" ]; then
    echo "  On Windows: works in Git Bash, WSL, and cmd.exe (new terminal)"
fi
echo "  (Open a new terminal if the command isn't found yet)"
echo ""
