# Qadav Toolchain Setup Guide

This document outlines the setup instructions to configure the prerequisites and compile Qadav across supported platforms.

---

## 1. Prerequisites

- **Go**: Version 1.22 or higher
- **Git**: Modern release
- **LLVM / Clang**: Version 18.0.0 or higher (`clang` is mandatory)
- **Make**: GNU Make (for running build tasks)

---

## 2. Windows Setup (Native)

### Recommended: Scoop Package Manager

Scoop installs user-space binaries without requiring administrative prompts:

```powershell
# Set execution policy
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Install Scoop
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Install dependencies
scoop install go llvm make git
```

### Alternative: Windows Package Manager (winget)

```powershell
winget install --id GoLang.Go -e
winget install --id LLVM.LLVM --version 18.1.8 -e
```

### Alternative: Chocolatey

```powershell
choco install golang llvm make -y
```

### Verification & Windows Adaptation Notes

Check that `clang` is reachable:

```powershell
clang --version
```

> **Important Architecture Note:**  
> The official Windows prebuilt distribution of LLVM supplies `clang.exe`, `lld.exe`, and related development utilities, but does not bundle `llc.exe` or `llvm-config.exe`.  
> Qadav is natively adapted to operate in **clang-only mode** on Windows. `clang -x ir` and `clang -c` handle all intermediate validation and object generation without requiring `llc`.

---

## 3. macOS Setup

Install LLVM via Homebrew:

```bash
brew install llvm@18
```

Add the LLVM toolchain binaries to your shell profile (`~/.zshrc` or `~/.bash_profile`):

```bash
export PATH="/opt/homebrew/opt/llvm@18/bin:$PATH"
source ~/.zshrc
```

Verify installation:

```bash
clang --version
```

---

## 4. Linux Setup (Ubuntu / Debian)

Install Clang 18 and LLVM packages:

```bash
sudo apt update
sudo apt install -y llvm-18 clang-18 llvm-18-dev make
```

Configure alternatives if necessary:

```bash
sudo update-alternatives --install /usr/bin/clang clang /usr/bin/clang-18 100
```

Verify installation:

```bash
clang --version
```

---

## 5. Environment Verification

Once prerequisites are installed, compile and run the diagnostic check:

```bash
go run ./cmd/qyn doctor
```

Or using the compiled binary:

```powershell
.\bin\qyn.exe doctor
```

All critical indicators should report success.

---

## 6. Troubleshooting

### "clang not found"
Ensure LLVM is installed and that the directory containing `clang` (or `clang.exe`) is included in your system PATH.

### "LLVM 18+ required, found X.Y.Z"
Your installed LLVM toolchain is older than the required minimum. Upgrade to LLVM 18.1.8 or higher.

### "make not found on Windows"
If executing via PowerShell, install GNU Make through Scoop:

```powershell
scoop install make
```

---

## 7. Testing & Compiler Flags

### Tamper Testing Flag: `--test`

Qadav provides native cryptographic tamper simulation statements (such as `corrupt_memory_share`, `corrupt_reasoning_step`, `corrupt_syndrome`) for automated test harnesses. To prevent accidental inclusion in production builds, these statements require the `--test` flag:

```powershell
# Compile and run with test-only statements enabled
.\bin\qyn.exe run --test examples\qemp_tamper.ydv
.\bin\qyn.exe run --test examples\qera_tamper.ydv
.\bin\qyn.exe run --test examples\qecma_tamper.ydv
.\bin\qyn.exe run --test examples\qec_tamper.ydv

# Build binary with test-only statements enabled
.\bin\qyn.exe build --test examples\qera_tamper.ydv
.\bin\qyn.exe build --test examples\qec_tamper.ydv
```

In standard builds without `--test`, any test-only statement produces a compile-time type-checker error.

### Runtime Modules
Qadav features 24 modular zero-libc runtime objects compiled on demand: `rt`, `tensor`, `nn`, `quantum`, `prime`, `io`, `dataset`, `quasinary`, `cir`, `agi`, `sha3`, `merkle`, `lamport`, `selfcheck`, `qeeh`, `qemwv`, `mctp`, `qemp`, `qecmt`, `qera`, `qecma`, `qec` (integrating QEC-QEMP and memory consolidation), `scoped` (CB-SRCPP privacy-preserving access), and `tools` (safe tool orchestration).

### Standard Library Management:
```powershell
# List registered standard library modules
.\bin\qyn.exe stdlib list

# Parse and type-check all standard library modules
.\bin\qyn.exe stdlib build
```



