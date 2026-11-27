# Qadav (.ydv) — Quantum-Native AGI Programming Language

Qadav is a sovereign, high-performance, quantum-native AGI systems programming language built from scratch with zero external runtime dependencies (**`C = 0`**, **`Go = 0`**, **`Py = 0`**).

The toolchain is completely self-hosted (**6/6 `.ydv` compiler stages**, **62 native `.ydv` modules**, **25 patents**), featuring direct standalone binary emission (`PE32+`, `ELF64`, `Mach-O`, `Wasm32`), hardware-resident AVX2 SIMD acceleration, and built-in cognitive architectures:
- **QNMF (Quantum-Native Knowledge Manifold):** Commutative bilateral relational ontology solving the LLM reversal curse with zero reverse training.
- **ECRA (Epistemic Calibrated Risk Abstention):** Calibrated confidence gating preventing AI hallucinations (Exit Code `100`).
- **Vishwanath Dynamic State Engine:** Continuous multi-constraint invariant guards enforcing physical consistency and thermal bounds; triggers lossless state rollback on contradiction (Exit Code `101`).
- **Hardware Fail-Stop Anti-Tamper:** CPU-level fail-stop enforcing `ExitProcess(0xDEADBEEF)` (`-559038737`) upon cryptographic signature forgery.

---

## ⚡ Quickstart in 3 Simple Steps (Windows)

### Step 1: Download / Clone
```powershell
git clone https://github.com/AvnishYdv/qadav.git
cd qadav
```

### Step 2: One-Click Installation
Double-click `install_qadav.bat` or run:
```bat
install_qadav.bat
```
*(Or via PowerShell: `powershell -ExecutionPolicy Bypass -File scripts\install_qadav.ps1`)*

**What this automatically does:**
- Copies compiler `qyn.exe`, package manager `qpk.exe`, stdlib, and runtime into `%LOCALAPPDATA%\Qadav\bin`.
- Safely and idempotently adds `%LOCALAPPDATA%\Qadav\bin` to your User `PATH`.
- Registers `.ydv` file type metadata in Windows Registry (`HKCU\Software\Classes\.ydv`).
- Auto-detects and installs the `qadav-lang` extension into VS Code (`%USERPROFILE%\.vscode\extensions`) and Cursor (`%USERPROFILE%\.cursor\extensions`).

### Step 3: Scaffold & Run in IDE
Open any terminal in any directory and run:
```powershell
qadav-new my_agent
code my_agent
```
Press **`Ctrl+Shift+B`** in VS Code or Cursor to build and run all 4 cognitive & security verification tiers immediately!

---

## 🛠️ CLI Quick Reference

```powershell
# Typecheck a Qadav source file
qyn check src/main.ydv

# Compile to standalone Windows PE32+ executable
qyn build src/main.ydv -o bin/main.exe

# Compile to standalone Linux x86_64 ELF64 executable (12,288 B)
qyn build src/main.ydv -o bin/app_linux linux

# Compile to standalone macOS x86_64 Mach-O executable (4,096 B)
qyn build src/main.ydv -o bin/app_macos macos

# Run system and compiler diagnostics
qyn doctor
```

---

## 🔒 Workspace Invariants (Strictly Enforced)
- `C Files (*.c)`: **0**
- `Go Files (*.go)`: **0**
- `Python Files (*.py)`: **0**
- `Native Qadav Modules (internal/qadav_native/*.ydv)`: **62**
- `Self-Hosted Stages (selfhost/*.ydv)`: **6**
- `Patents (docs/PATENTS.md)`: **25**
