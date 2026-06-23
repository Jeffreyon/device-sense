#!/bin/bash
# device-sense installer
# Usage: curl -fsSL https://raw.githubusercontent.com/jeffreyon/device-sense/main/install.sh | bash

REPO_RAW="https://raw.githubusercontent.com/jeffreyon/device-sense/main"
INSTALL_DIR="$HOME/bin"
CMD_NAME="device-sense"

echo ""
echo "  Installing device-sense..."
echo ""

# ── Create ~/bin if it doesn't exist ──────────────────────────────────────────
mkdir -p "$INSTALL_DIR"

# ── Download the script ────────────────────────────────────────────────────────
if command -v curl &>/dev/null; then
    curl -fsSL "$REPO_RAW/device-sense.sh" -o "$INSTALL_DIR/$CMD_NAME"
elif command -v wget &>/dev/null; then
    wget -qO "$INSTALL_DIR/$CMD_NAME" "$REPO_RAW/device-sense.sh"
else
    echo "  Error: curl or wget is required. Install either and try again."
    exit 1
fi

chmod +x "$INSTALL_DIR/$CMD_NAME"

# ── Add ~/bin to PATH if it isn't already ─────────────────────────────────────
add_to_path() {
    local rc="$1"
    if [ -f "$rc" ] && grep -q 'device-sense' "$rc" 2>/dev/null; then
        return  # already added
    fi
    printf '\n# added by device-sense installer\nexport PATH="$HOME/bin:$PATH"\n' >> "$rc"
}

if [[ ":$PATH:" != *":$HOME/bin:"* ]]; then
    # cover both .bashrc and .bash_profile so it works in login + interactive shells
    add_to_path "$HOME/.bashrc"
    add_to_path "$HOME/.bash_profile"
    # make it available in the current session immediately
    export PATH="$HOME/bin:$PATH"
    echo "  Added ~/bin to PATH in ~/.bashrc and ~/.bash_profile"
fi

# ── Done ──────────────────────────────────────────────────────────────────────
echo "  device-sense installed to $INSTALL_DIR/$CMD_NAME"
echo ""
echo "  Run it now:"
echo "    device-sense"
echo ""
echo "  (If the command isn't found, restart your terminal or run:  source ~/.bashrc)"
echo ""
