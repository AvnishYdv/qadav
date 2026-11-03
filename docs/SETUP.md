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
Qadav features 36 modular zero-libc runtime objects compiled on demand (expanded from 30 in M6.3+M6.4, 32 in M6.5+M6.6, and 34 in M7.1+M7.2):
1. `qadav_alloc` (Zero-libc arena & page allocator)
2. `qadav_sha3` (Keccak-f1600 / SHA3-256 cryptographic sponge)
3. `qadav_merkle` (Merkle tree builder & root authentication)
4. `qadav_lamport` (WOTS post-quantum signature engine)
5. `qadav_selfcheck` (PE self-validation & code nullification)
6. `qadav_runtime` (Core runtime entrypoint & syscall layer)
7. `qadav_tensor` (Tensor storage, strided slicing, allocations)
8. `qadav_quantum` (Simulated statevector & qubit entanglement)
9. `qadav_quasinary` (Balanced-ternary & energy minimization)
10. `qadav_prime` (Prime-symbolic knowledge ontology)
11. `qadav_dataset` (Batching, streaming, dataset pipelines)
12. `qadav_cir` (Cognitive Intermediate Representation)
13. `qadav_agi` (Native AGI types & causality graphs)
14. `qadav_qeeh` (Quantum-Entangled Ethical Hamiltonian)
15. `qadav_qemwv` (Quantum-Entangled Model Weight Vault)
16. `qadav_mctp` (Merkle-Certified Training Provenance)
17. `qadav_qemp` (Quantum-Entangled Memory Provenance)
18. `qadav_qecmt` (Quantum-Entangled Cross-Model Transfer)
19. `qadav_qera` (Quantum-Entangled Reasoning Attestation)
20. `qadav_qecma` (Quantum-Entangled Cross-Modal Attestation)
21. `qadav_qec` (Quantum Error-Corrected Memory Provenance)
22. `qadav_consolidate` (Episodic-to-semantic memory consolidation)
23. `qadav_scoped` (CB-SRCPP privacy-preserving scoped access)
24. `qadav_tools` (Safe tool orchestration & pre/post conditions)
25. `qadav_certificate` (TTPCR type-tagged reasoning certificates)
26. `qadav_abstain` (Calibrated epistemic uncertainty & abstention)
27. `qadav_sandbox` (QESI quantum-entangled sandbox isolation)
28. `qadav_gpu` (DirectCompute GPU device detection & CPU fallback)
29. `qadav_agke` (Attested GPU Kernel Execution - Invention #17)
30. `qadav_gpu_matmul` (Zero-libc fallback matrix kernels)
31. `qadav_multigpu` (Multi-GPU DXGI pipeline parallelism, added in M6.3)
32. `qadav_cdaga` (Cross-Device Attested Gradient Aggregation - Invention #18, added in M6.4)
33. `qadav_distributed` (In-process Ring AllReduce & Parameter Server, added in M6.5)
34. `qadav_barat` (Byzantine-Attested Robust Aggregation with Transparent selection - Invention #19, added in M6.6)
35. `qadav_qhal` (Quantum Hardware Abstraction Layer & OpenQASM 3.0 export, added in M7.1)
36. `qadav_sea` (Self-Evolution Attestation & lineage chain - Invention #20, added in M7.2)

*Transition Notes:*
- *30 → 32 (M6.3 + M6.4):* Added `qadav_multigpu.c` (DXGI device topology & pipeline execution) and `qadav_cdaga.c` (canonical device-sorted gradient aggregation attestation).
- *32 → 34 (M6.5 + M6.6):* Added `qadav_distributed.c` (in-process zero-socket ring allreduce and parameter server) and `qadav_barat.c` (Byzantine-attested robust trimmed-mean aggregation with transparency certificate and catastrophic decoherence).
- *34 → 36 (M7.1 + M7.2):* Added `qadav_qhal.c` (OpenQASM 3.0 export & backend stub abstraction) and `qadav_sea.c` (Self-Evolution Attestation proposal verification, Merkle lineage DAG, and QERA coupling).

### Standard Library Management:
```powershell
# List registered standard library modules
.\bin\qyn.exe stdlib list

# Parse and type-check all standard library modules
.\bin\qyn.exe stdlib build
```



