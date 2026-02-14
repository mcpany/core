#!/bin/bash
# Install ShellCheck robustly or skip gracefully

SHELLCHECK_BIN="$1"
TOOL_INSTALL_DIR=$(dirname "$SHELLCHECK_BIN")

if [ -f "$SHELLCHECK_BIN" ]; then
    echo "ShellCheck is already installed."
    exit 0
fi

echo "Installing ShellCheck..."

# Helper to try apt-get
install_apt() {
    if command -v apt-get >/dev/null 2>&1; then
        echo "Attempting install via apt-get..."
        # Check for sudo
        if command -v sudo >/dev/null 2>&1; then
            sudo apt-get update && sudo apt-get install -y shellcheck
        else
            apt-get update && apt-get install -y shellcheck
        fi

        # Link if installed to system path but expected in custom path
        SYSTEM_SC=$(command -v shellcheck)
        if [ -n "$SYSTEM_SC" ] && [ "$SYSTEM_SC" != "$SHELLCHECK_BIN" ]; then
            ln -sf "$SYSTEM_SC" "$SHELLCHECK_BIN"
        fi
        return $?
    fi
    return 1
}

# Helper to try download
install_curl() {
    echo "Attempting install via binary download..."
    # Check for xz
    if ! command -v xz >/dev/null 2>&1; then
        echo "Error: xz not found. Cannot decompress binary."
        return 1
    fi

    SC_VERSION="stable"
    ARCH=$(uname -m)
    # Map arch if needed, but stable tarball usually covers common ones or linux-x86_64
    # The URL in Makefile was hardcoded to linux.x86_64.
    # If running on arm64, the binary download might be wrong architecture anyway!
    # This is likely why the build fails on linux/arm64 job.

    URL="https://github.com/koalaman/shellcheck/releases/download/${SC_VERSION}/shellcheck-${SC_VERSION}.linux.x86_64.tar.xz"
    if [ "$ARCH" = "aarch64" ]; then
        URL="https://github.com/koalaman/shellcheck/releases/download/${SC_VERSION}/shellcheck-${SC_VERSION}.linux.aarch64.tar.xz"
    fi

    echo "Downloading from $URL..."
    if curl -sSL "$URL" | tar -xJv -C /tmp; then
        mv "/tmp/shellcheck-${SC_VERSION}/shellcheck" "$SHELLCHECK_BIN"
        rm -rf "/tmp/shellcheck-${SC_VERSION}"
        return 0
    fi
    return 1
}

# Try methods
if install_apt; then
    echo "ShellCheck installed via apt."
    exit 0
fi

if install_curl; then
    echo "ShellCheck installed via curl."
    exit 0
fi

echo "Warning: ShellCheck installation failed. Continuing without it."
exit 0
