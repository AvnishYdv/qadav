#!/usr/bin/env bash
set -e

# Detect OS
OS="$(uname)"

install_llvm_macos() {
    echo "Installing LLVM 18+ via Homebrew..."
    brew install llvm@18
    # Ensure llvm-config is in PATH
    if ! command -v llvm-config >/dev/null 2>&1; then
        echo "Adding llvm-config to PATH..."
        export PATH="$(brew --prefix llvm@18)/bin:$PATH"
    fi
    llvm-config --version
}

install_llvm_linux() {
    echo "Installing LLVM 18+ via apt..."
    sudo apt-get update
    sudo apt-get install -y llvm-18 llvm-18-dev clang-18
    # Verify installation
    llvm-config-18 --version || llvm-config --version
}

case "$OS" in
    "Darwin")
        install_llvm_macos
        ;;
    "Linux")
        install_llvm_linux
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Confirm llvm-config is reachable
if command -v llvm-config >/dev/null 2>&1; then
    echo "llvm-config found: $(which llvm-config)"
else
    echo "llvm-config not found in PATH. Please add it manually."
    exit 1
fi

