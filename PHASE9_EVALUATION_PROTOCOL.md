# QADAV PHASE 9 EVALUATION PROTOCOL
## Step-by-Step Reproduction, Verification, and Adversarial Testing Guide for Independent Evaluators

**Target Audience:** Independent Evaluators, Security Auditors, Technical Due Diligence Teams  
**Target System:** Windows 10/11 x86_64  
**Scope:** Qadav Phase 9 Native AI Inventions (M9.1–M9.8)  

---

## 1. Prerequisites & Environment Setup

Qadav is designed to run completely standalone without requiring heavy developer toolchains (no Python, no Node.js, no Go, no Rust, and no Visual Studio C++ runtime). The only external tool required is the LLVM toolchain for native object linking.

### Required Software:
1. **Operating System:** Windows 10 or 11 (64-bit x86_64).
2. **LLVM / Clang:** `clang` and `lld-link` (version 15+ recommended, available in system PATH).
3. **PowerShell:** PowerShell 5.1+ or PowerShell Core 7+.
4. **Platform SDK Library:** `kernel32.lib` (standard Windows API import library located at `C:\Users\avnis\.qadav\runtime\kernel32.lib` or standard Windows SDK path).

---

## 2. Invariant Verification Checklist

Before running any benchmarks, the evaluator must verify that the repository adheres to the locked structural invariants. Run the following PowerShell commands in the repository root:

```powershell
# 1. Verify Patent Count (Must be exactly 25)
$patents = (Select-String -Path docs\PATENTS.md -Pattern '^## \d+\.').Count
Write-Host "Patents count: $patents (Expected: 25)"
if ($patents -ne 25) { throw "Patent count mismatch!" }

# 2. Verify C Runtime File Count (Must be exactly 46 total, 45 Windows-linked)
$totalC = (Get-ChildItem internal\runtime\*.c).Count
$winC = (Get-ChildItem internal\runtime\*.c | Where-Object { $_.Name -ne 'qadav_rt_linux.c' }).Count
Write-Host "Total C runtime files: $totalC (Expected: 46)"
Write-Host "Windows C runtime files: $winC (Expected: 45)"
if ($totalC -ne 46 -or $winC -ne 45) { throw "Runtime file count mismatch!" }

# 3. Verify Zero-Dependency Invariant (No .go or .py files in repository)
$goFiles = (Get-ChildItem -Recurse -Filter *.go -ErrorAction SilentlyContinue).Count
$pyFiles = (Get-ChildItem -Recurse -Filter *.py -ErrorAction SilentlyContinue).Count
Write-Host "Go files: $goFiles (Expected: 0)"
Write-Host "Python files: $pyFiles (Expected: 0)"
if ($goFiles -ne 0 -or $pyFiles -ne 0) { throw "Forbidden language files detected!" }

# 4. Verify Self-Hosted Compiler Binary
if (-not (Test-Path "bin\qyn.exe")) { throw "bin\qyn.exe missing!" }
Write-Host "Compiler verified: bin\qyn.exe present."
```

---

## 3. Phase 8 Core Regression Protocol (10/10 Byte-for-Byte Check)

To ensure the self-hosted compiler has not drifted or corrupted its LLVM IR emission, run the standard Phase 8 regression test. Every emitted file must match its reference counterpart byte-for-byte.

```powershell
# Run the automated Phase 8 regression verification
powershell -ExecutionPolicy Bypass -File TMP\check_reg.ps1
```

### Manual Individual Step Verification:
For each of the 10 test programs, generate the IR and execute a binary byte comparison:
```powershell
$tests = @('hello', 'calc', 'if_else', 'fib', 'dot_product', 'composite', 'classify_operator', 'classify_annotation', 'classify_keyword', 'match_n_arms_proof')

foreach ($t in $tests) {
    cmd.exe /c ".\bin\qyn.exe emit-llvm examples\$t.ydv > TMP\reg_$t.ll"
    fc.exe /b "TMP\ref_$t.ll" "TMP\reg_$t.ll"
}
```
**Pass Criterion:** All 10 tests must output `FC: no differences encountered`.

---

## 4. Phase 9 Product Demos Compilation & Execution Protocol

The following sections detail how an evaluator can compile and execute all 8 AI product demonstrations from source `.ydv` files.

### 4.1 Common Compilation Pattern
Each demo is compiled from Qadav source (`.ydv`) to LLVM IR via `bin/qyn.exe`, then linked against the zero-libc runtime objects via `clang`:

```cmd
.\bin\qyn.exe emit-llvm examples\<DEMO>.ydv > TMP\<DEMO>.ll
clang -O0 TMP\<DEMO>.ll TMP\runtime_objs\*.o -nostdlib -fuse-ld=lld -Wl,-entry:qadav_start -Wl,-subsystem:console C:\Users\avnis\.qadav\runtime\kernel32.lib -o TMP\<DEMO>.exe
```

---

### 4.2 Comprehensive Reproduction Matrix

Run the automated reproduction runner:
```powershell
$demos = @(
    @{ Exe = "TMP\product_qemwv.exe"; Expected = 0; Label = "QEMWV #7 Happy" },
    @{ Exe = "TMP\product_qemwv_tamper.exe"; Expected = -559038737; Label = "QEMWV #7 Tamper" },
    @{ Exe = "TMP\product_qera.exe"; Expected = 0; Label = "QERA #11 Happy" },
    @{ Exe = "TMP\product_qera_tamper.exe"; Expected = -559038737; Label = "QERA #11 Tamper" },
    @{ Exe = "TMP\product_mctp.exe"; Expected = 0; Label = "MCTP #8 Happy" },
    @{ Exe = "TMP\product_mctp_tamper.exe"; Expected = 0; Label = "MCTP #8 Tamper Test" },
    @{ Exe = "TMP\product_ttpcr.exe"; Expected = 0; Label = "TTPCR #15 Happy" },
    @{ Exe = "TMP\product_ttpcr_tamper.exe"; Expected = -559038737; Label = "TTPCR #15 Tamper" },
    @{ Exe = "TMP\product_ttpcr_abstain.exe"; Expected = 0; Label = "TTPCR #15 Abstain" },
    @{ Exe = "TMP\product_barat.exe"; Expected = 0; Label = "BARAT #19 Happy" },
    @{ Exe = "TMP\product_barat_tamper.exe"; Expected = -559038737; Label = "BARAT #19 Tamper" },
    @{ Exe = "TMP\product_cdaga.exe"; Expected = 0; Label = "CDAGA #18 Happy" },
    @{ Exe = "TMP\product_cdaga_tamper.exe"; Expected = -559038737; Label = "CDAGA #18 Tamper" },
    @{ Exe = "TMP\product_qecma.exe"; Expected = 0; Label = "QECMA #12 Happy" },
    @{ Exe = "TMP\product_qecma_tamper.exe"; Expected = -559038737; Label = "QECMA #12 Tamper" },
    @{ Exe = "TMP\product_qecma_swap.exe"; Expected = -559038737; Label = "QECMA #12 Swap" },
    @{ Exe = "TMP\product_qesi.exe"; Expected = 0; Label = "QESI #16 Contained Sandbox" },
    @{ Exe = "TMP\product_qesi_tamper.exe"; Expected = -559038737; Label = "QESI #16 Tamper" },
    @{ Exe = "TMP\product_qesi_isolation.exe"; Expected = 0; Label = "QESI #16 Host Isolation" },
    @{ Exe = "TMP\product_integrated.exe"; Expected = 0; Label = "Integrated Happy" },
    @{ Exe = "TMP\product_integrated_tamper.exe"; Expected = -559038737; Label = "Integrated Tamper" }
)

Write-Host "Running 21-point evaluation suite..."
$allPass = $true
foreach ($d in $demos) {
    & $d.Exe | Out-Null
    $code = $LASTEXITCODE
    $status = if ($code -eq $d.Expected) { "PASS" } else { "FAIL"; $allPass = $false }
    Write-Host "[$status] $($d.Label): Exit $code (Expected: $($d.Expected))"
}
if ($allPass) { Write-Host "`nALL 21 EVALUATION RUNS PASSED!" -ForegroundColor Green }
```

---

## 5. Adversarial Testing & Fault Injection Procedures

Independent evaluators may perform active adversarial testing to confirm the robustness of the Qadav security mechanisms:

### Test Case A: Model Weight Memory Manipulation
1. Run `TMP\product_qemwv_tamper.exe`.
2. **Mechanism:** The test alters one byte in the AES-GCM ciphertext prior to decryption.
3. **Verification:** Confirm that the process does NOT output plaintext weights, does NOT throw a generic exception, and immediately terminates with exit code `3735928559` (`0xDEADBEEF` / `-559038737`).

### Test Case B: Reasoning Witness Forgery
1. Run `TMP\product_ttpcr_tamper.exe`.
2. **Mechanism:** Step 1's proof witness (`pathology_protocol_wbc`) is substituted with an arbitrary unauthorized string while retaining the outer claim.
3. **Verification:** Confirm that the Merkle proof verification fails and triggers immediate decoherence (`ExitProcess(0xDEADBEEF)`).

### Test Case C: Multimodal Domain Hijacking
1. Run `TMP\product_qecma_swap.exe`.
2. **Mechanism:** Modality tag `0x58` (Text) is swapped into `0x54` (Tensor) to simulate embedding hijacking.
3. **Verification:** Confirm that the runtime detects cross-modal evidence domain separation violation and halts via decoherence (`-559038737`).

### Test Case D: Sandbox Containment & Host Resilience
1. Run `TMP\product_qesi.exe`.
2. **Mechanism:** An untrusted agent executes inside a sealed sandbox and attempts an unauthorized capability invocation.
3. **Verification:** Confirm that:
   - The sandbox execution halts immediately.
   - The host runtime survives cleanly (`HOST_SURVIVED`).
   - The final exit code is `0`.
   - Modifying the sandbox's forensic attestation (`TMP\product_qesi_tamper.exe`) triggers host decoherence (`-559038737`).
