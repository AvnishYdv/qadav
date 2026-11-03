# Qadav Project Changelog
# कादव परियोजना परिवर्तन लॉग

All notable updates, architectural enhancements, and completed milestones for the Qadav programming language.

---

## [M4.1 + M4.2] — Combined Milestone: Quantum-Entangled Model Weight Vault (QEMWV — Invention #7) + Merkle-Certified Training Provenance (MCTP — Invention #8)
## [M4.1 + M4.2] — संयुक्त माइलस्टोन: क्वांटम-उलझा हुआ मॉडल भार वॉल्ट (QEMWV — आविष्कार #7) एवं मर्कल-प्रमाणित प्रशिक्षण साक्ष्य (MCTP — आविष्कार #8)

### Fixed & Technical Debt Resolutions (4 Mandatory Fixes from M3.5+M3.6: I1–I4) / तकनीकी ऋण सुधार:
- **Fix I1 (SHA3-256 Cycle Count Consistency)**: Standardized SHA3-256 benchmark figure across `docs/CHANGELOG.md` to 1,640 cycles (~0.55 $\mu\text{s}$) at 3 GHz, completely removing legacy 2,475 cycle references.
- **Fix I2 (Explicit Decoherence Statement Execution)**: Updated `examples/qeeh_demo.ydv` to explicitly invoke the `decoherence ·` statement and verified runtime output (`"Decoherence active"`).
- **Fix I3 (Typed Action & SHA3 Hashing in QEEH)**: Upgraded QEEH semantic action evaluation in `qadav_qeeh.h` / `qadav_qeeh.c` from simple substring matching to a typed `qadav_action_kind` enum and discrete SHA3-256 action message digest hashing (`qadav_qeeh_evaluate_action_hash`).
- **Fix I4 (Tamper Mechanism Separation)**: Explicitly delineated cognitive QEEH wavefunction decoherence (zeroing quantum state amplitudes) from executable binary PE self-check code nullification (`ExitProcess(0xDEADBEEF)`) in tests and documentation.

### Added / नया जोड़ा गया:
- **Quantum-Entangled Model Weight Vault (QEMWV — Invention #7, Patent Pending)**:
  - Language tokens and syntax: `vault <name> ⟦ threshold ↦ <k> · shares ↦ <n> · target ↦ <network> · ⟧`, `seal <vault> ·`, `unseal <vault> with <boundary> ·`, and `wipe <vault> ·`.
  - Information-Theoretic Security: Shamir's Secret Sharing over Galois Field $GF(256)$ with field polynomial $x^8 + x^4 + x^3 + x + 1$ ($0x11b$), zero-libc implementation in `qadav_qemwv.c` / `qadav_qemwv.h`.
  - Cryptographic Verification: Per-tensor SHA3-256 commitment digests verified upon polynomial interpolation.
  - Active Tamper Decoherence: Any unseal attempt lacking ethical boundary clearance or with tampered shares triggers immediate zeroization of quantum amplitudes (`qadav_qeeh_trigger_decoherence`).
- **Merkle-Certified Training Provenance (MCTP — Invention #8, Patent Pending)**:
  - Language tokens and syntax: `provenance <name> ⟦ target ↦ <network> · cert ↦ <filepath> · ⟧`, `certify <provenance> ·`, and `verify_training ⦅ <filepath> ⦆`.
  - Cryptographic Verification Architecture: Per-epoch 108-byte Merkle leaves combining epoch counter, SHA3 weights digest, SHA3 dataset digest, loss metric, and hyperparameter digest.
  - Post-Quantum Signature: Winternitz One-Time Signature (WOTS, $w=256$, 1,088 bytes) signing the Merkle root over the entire training lifecycle.
  - Standalone Binary Sidecar: Zero-libc file format `.mctp` written and read directly via Win32 API (`CreateFileW`, `WriteFile`, `ReadFile`).
- **Tooling & CLI**:
  - `qyn verify-training <cert.mctp> [model]`: Post-quantum training provenance cert validator CLI command.
  - `qyn build <file.ydv> --emit-mctp`: Compiler flag to automatically bind and emit training provenance certificates.
- **Runtime Expansion**:
  - Embedded zero-libc runtime expanded to 17 compiled C modules (added `qadav_qemwv.o` and `qadav_mctp.o`), maintaining 100% zero-libc status.
- **Coverage & Testing**:
  - All 11 internal packages maintained at $\ge 80\%$ test statement coverage.
  - Dedicated tests added for $GF(256)$ Shamir secret sharing, unauthorized decoherence, MCTP epoch integrity, full certificate verification, and compiler integration.
- **Documentation**:
  - Added Section 14 (QEMWV) and Section 15 (MCTP) to `docs/GUIDE.md`.
  - Added Inventions #7 and #8 to `docs/PATENTS.md`.
  - Added Section 11 to `docs/SYNTAX.md`.
  - Added Sections 18, 19, 20 to `docs/EXAMPLES.md`.
  - Added dedicated specification files `docs/QEMWV_SPEC.md` and `docs/MCTP_SPEC.md`.

---

## [M3.5 + M3.6] — Combined Milestone: QEEH Implementation & Final Integration & Hardening (Invention #6)
## [M3.5 + M3.6] — संयुक्त माइलस्टोन: QEEH कार्यान्वयन एवं अंतिम एकीकरण व सुदृढ़ीकरण (आविष्कार #6)

### Fixed & Technical Debt Resolutions (4 Mandatory Fixes from M3.3+M3.4) / तकनीकी ऋण सुधार:
- **Fix 1 (`docs/GUIDE.md` Cleanup)**: Cleaned duplicate Section 12 headings, established clean sequence of Sections 1 through 13.
- **Fix 2 (Precise Self-Nullification)**: Updated `qadav_selfcheck.c` to parse PE Exception Directory (`.pdata`) and Section Headers to obtain exact byte size of `main()`, completely eliminating arbitrary 4096-byte constants.
- **Fix 3 (SHA3-256 Performance Optimization)**: Optimized Keccak-$f[1600]$ in `qadav_sha3.c` with unrolled Theta, constant rotations, and fast Chi. Benchmarked at 1,640 CPU cycles (~0.55 $\mu\text{s}$), achieving >3000x speedup well below the 1ms target.
- **Fix 4 (Lamport OTS Size Reduction to Winternitz OTS)**: Transitioned OTS scheme to Winternitz OTS ($w=256$, 34 chains = 1,088 bytes) in `qadav_lamport.c` and `internal/selfcheck`, reducing signature size to 1.06 KB (< 2KB target) while preserving 128-bit Grover resistance.

### Added / नया जोड़ा गया:
- **Quantum-Entangled Ethical Hamiltonian (QEEH — Invention #6)**:
  - Language tokens and syntax: `ethical_boundary <name> ⟦ rule <r> ↦ <penalty> · ⟧`, `enforce <boundary> on <action> ·`, and `decoherence ·`.
  - Type-checking rules: ethical boundaries and rule penalty ranges ($[0.0, 1.0]$) validated at compile time.
  - Runtime implementation: `internal/runtime/qadav_qeeh.h` and `qadav_qeeh.c` evaluating energy $E = \sum w_i (1.0 - c_i)$ and triggering wavefunction zeroization.
  - LLVM IR generation: compiled down to `@qadav_qeeh_create`, `@qadav_qeeh_add_rule`, `@qadav_qeeh_evaluate_action`, `@qadav_qeeh_trigger_decoherence`.
- **Test Coverage Across All 11 Internal Packages**:
  - `internal/token`: 100.0%
  - `internal/lexer`: 96.1%
  - `internal/selfcheck`: 95.6%
  - `internal/runtime`: 91.9%
  - `internal/ast`: 88.8%
  - `internal/types`: 86.4%
  - `internal/llvm`: 83.0%
  - `internal/codegen`: 82.2%
  - `internal/parser`: 80.9%
  - `internal/checker`: 80.7%
  - `internal/testutil`: 80.2%
  - Every single package $\ge 80\%$ statement test coverage.
- **Examples & Tooling**:
  - `examples/qeeh_demo.ydv`: Demonstrates allowed actions and ethical boundary decoherence blocking.
- **Documentation**:
  - Added Section 13 to `docs/GUIDE.md`.
  - Published Invention #6 in `docs/PATENTS.md`.
  - Updated `docs/SECURITY.md`, `docs/SELFCHECK_SPEC.md`, and `docs/EXAMPLES.md`.

---

## [M3.3 + M3.4] — Combined Milestone: Merkle Infrastructure & Post-Quantum Signature (Invention #5)
## [M3.3 + M3.4] — संयुक्त माइलस्टोन: मर्कल इंफ्रास्ट्रक्चर एवं उत्तर-क्वांटम हस्ताक्षर (आविष्कार #5)

### Fixed & Technical Debt Resolutions (5 Documentation Drifts Resolved) / तकनीकी ऋण सुधार:
- **Drift 1 (CIR Basic Example)**: Fixed `docs/EXAMPLES.md` Section 11 from `0.90 ∵ 100.0` / `0.80 ∵ 100.0` to `0.90 ∵ 10.0` / `0.80 ∵ 20.0` matching `examples/cir_basic.ydv`.
- **Drift 2 (Episodic Memory Example)**: Fixed `docs/EXAMPLES.md`, `docs/GUIDE.md`, and `docs/SYNTAX.md` to `memory mem ⟬ 100 ⟭ ·`, `remember mem⟬"agent_state"⟭ ↦ 42 ·`, `recall mem⟬"agent_state"⟭` matching `examples/memory_basic.ydv`.
- **Drift 3 (Semantic Graph Example)**: Fixed `docs/EXAMPLES.md`, `docs/GUIDE.md`, and `docs/SYNTAX.md` to `node g ⟬ "conceptA" ↦ 101 ⟭ ·`, `link g ⟬ "conceptA" ⸴ "conceptB" ⸴ 1.0 ⟭ ·`, `trace g ⟬ "conceptA" ⸴ 2 ⟭` matching `examples/graph_basic.ydv`.
- **Drift 4 (Tensor Multiplication Confidence Semantics)**: Clarified in `docs/SYNTAX.md` and `docs/SECURITY.md` that $\otimes$ strictly follows the joint probability product rule $s_a \times s_b$ per `docs/CIR_SPEC.md`.
- **Drift 5 (Codegen Test Timing)**: Standardized to single consistent reference timing: warm 8.25s (target <12s), cold 18.78s (first run).

### Added / नया जोड़ा गया:
- **Self-Validating Post-Quantum Binary Integrity (Invention #5)**:
  - Zero-libc FIPS 202 SHA3-256 implementation (`internal/runtime/qadav_sha3.c`, `qadav_sha3.h`) implementing Keccak-$f[1600]$ without `<openssl/*>`, `<sodium.h>`, or CRT dependencies.
  - Zero-libc Merkle tree constructor (`internal/runtime/qadav_merkle.c`, `qadav_merkle.h`) computing hierarchical binary digests over 4KB chunks with duplicate balancing.
  - Zero-libc Lamport One-Time Signature (OTS) scheme (`internal/runtime/qadav_lamport.c`, `qadav_lamport.h`) providing 128-bit post-quantum security against Grover search.
  - Pre-main integrity check (`internal/runtime/qadav_selfcheck.c`, `qadav_selfcheck.h`) hooked into `qadav_start`.
  - Self-nullification routine: upon binary tampering or sidecar mismatch, unlocks code page memory via `VirtualProtect`, completely overwrites `main()` instructions with zeroes, and terminates with exit code `0xDEADBEEF` (`-559038737`).
  - Runtime object compilation: added `qadav_sha3.c`, `qadav_merkle.c`, `qadav_lamport.c`, `qadav_selfcheck.c` to `internal/runtime/runtime.go` (14 runtime objects total). Exported Win32 `VirtualProtect`, `GetModuleHandleA`, `GetModuleFileNameA`, `GetFileSize` in freestanding `kernel32.lib`.
- **Pure Go Self-Check Package (`internal/selfcheck`)**:
  - Pure Go stdlib implementation of FIPS 202 SHA3-256, Merkle tree construction, Lamport OTS key generation, signing, and verification.
  - Sidecar container support (`.qsc`) with `Embed`, `Verify`, and `Strip` routines.
  - 93.6% statement test coverage with complete unit test suite (`selfcheck_test.go`).
- **CLI Commands (`qyn selfcheck`)**:
  - `qyn build <file.ydv> --selfcheck`: Compiles binary and generates cryptographic sidecar `.qsc`.
  - `qyn selfcheck build <target.exe>`: Generates `.qsc` sidecar for an existing executable.
  - `qyn selfcheck verify <target.exe>`: Verifies executable integrity against its `.qsc` sidecar.
  - `qyn selfcheck strip <target.exe>`: Removes the `.qsc` sidecar.
- **Examples**:
  - `examples/selfcheck_demo.ydv`: Demonstrates untampered execution (exiting 0) and tampered self-nullification (exiting `0xDEADBEEF`).
- **Documentation**:
  - Created `docs/SELFCHECK_SPEC.md` specifying `.qsc` format, Merkle tree layout, and Lamport signature scheme.
  - Defensively published Invention #5 in `docs/PATENTS.md`.
## [M3.1 + M3.2] — संयुक्त माइलस्टोन: पूर्ण कॉग्निटिव इंटरमीडिएट रिप्रेजेंटेशन (CIR) एवं AGI मूल प्रकार

### Fixed & Technical Debt Resolutions (Carry-Overs Resolved) / तकनीकी ऋण सुधार:
- **Fix A1 (Test Coverage $\ge 80\%$)**: All 10 internal packages now achieve $\ge 80\%$ test statement coverage:
  - `internal/ast`: 91.1%
  - `internal/checker`: 80.5%
  - `internal/codegen`: 81.8%
  - `internal/lexer`: 96.1%
  - `internal/llvm`: 83.0%
  - `internal/parser`: 86.2%
  - `internal/runtime`: 80.0%
  - `internal/testutil`: 80.2%
  - `internal/token`: 100.0%
  - `internal/types`: 86.4%
- **Fix A2 (Quasi Grid Codegen)**: `quasi_grid.ll` now generates 9 calls to `@qadav_tensor_set_i64` and strictly 0 calls to `@qadav_tensor_set_f64`.
- **Fix A3 (Ferromagnetic Ground State)**: `quasi_grid.exe` converges to all-same spins (`◄` ground state) under zero-libc simulated annealing. Added unit regression test `TestQuasiGridFerromagnetic`.
- **Fix A4 (Codegen Test Execution Speed)**: Single-step LLVM compilation directly links `.ll` with freestanding runtime objects, reducing `go test ./internal/codegen/...` run time to **warm 8.25s (target <12s), cold 18.78s (first run)**.

### Added / नया जोड़ा गया:
- **Full Cognitive Intermediate Representation (CIR — Invention #4)**:
  - Probabilistic types: `confident⟨T⟩` and `explore⟨T⟩`.
  - Epistemic states: `KNOWN` (0), `UNKNOWN` (1), `EXPLORE` (2).
  - Calibrated confidence literals: `score ∵ value` with compile-time bound validation ($0.0 \le score \le 1.0$).
  - Closed propagation algebra: tensor product $\otimes$ ($s_a \times s_b$), independent error geometric mean $\oplus$ ($\sqrt{s_a \times s_b}$), conservative lower bound $\min(s_a, s_b)$, and epistemic contagion.
  - Native verification statements: `assert value ≽ threshold ·` (terminating on violation with exit code 1), `infer model on input ↦ out ·`, and `propagate value from source ·`.
  - Zero-libc runtime implementation: `internal/runtime/qadav_cir.c` and `qadav_cir.h`.
- **AGI Native Types (Invention #4 Extension)**:
  - `memory`: Associative episodic memory buffer with 64-bit FNV-1a hash table and bucket chaining. Keywords: `memory`, `remember`, `recall`.
  - `graph`: Semantic association graph with adjacency list and native Breadth-First Search topological `trace` returning `tensor⟨int64⟩`. Keywords: `graph`, `node`, `link`, `trace`.
  - `procedure`: First-class executable cognitive tokens with direct zero-libc dispatching. Keywords: `procedure`, `execute`.
  - Zero-libc runtime implementation: `internal/runtime/qadav_agi.c` and `qadav_agi.h`.
- **Runtime Compilation Expansion**:
  - `runtime.Build` now compiles 10 modular freestanding native objects (`qadav_rt`, `qadav_tensor`, `qadav_nn`, `qadav_quantum`, `qadav_prime`, `qadav_io`, `qadav_dataset`, `qadav_quasinary`, `qadav_cir`, `qadav_agi`).
- **Examples**:
  - `examples/cir_basic.ydv`: Confident arithmetic and assert pass.
  - `examples/cir_assert_fail.ydv`: Assert failure exit code 1.
  - `examples/memory_basic.ydv`: Episodic memory remember and recall.
  - `examples/graph_basic.ydv`: Semantic graph node/link/trace.
  - `examples/procedure_basic.ydv`: Procedure declaration and execute.
- **Specifications & Documentation**:
  - Created `docs/CIR_SPEC.md` and `docs/AGI_TYPES.md`.
  - Defensively published Invention #4 in `docs/PATENTS.md`.

---

## [M2.7 + M2.8] — Combined Milestone: Dataset Pipeline & Quasinary Computing (Invention #3)
## [M2.7 + M2.8] — संयुक्त माइलस्टोन: डेटासेट पाइपलाइन एवं क्वासिनरी कम्प्यूटिंग

### Fixed & Technical Debt Resolutions / सुधार एवं तकनीकी ऋण समाधान:
- **⊕ vs ⊻ Dual Semantics Eliminated**: `⊕` (`OSUM`) strictly and unconditionally emits arithmetic addition (`add i64`, `fadd double`). `⊻` (`XVEE`) strictly and unconditionally emits bitwise/logical XOR (`xor i64`, `xor i1`). Removed all dual `IsBit` tracking.
- **Codegen Test Runtime Caching**: Implemented `internal/testutil/runtime_cache.go` (`testutil.RuntimeObjects(t)`) caching precompiled native runtime objects using `sync.Once`. Codegen test execution time reduced from 69.7s to **15.3s** (a 4.5x speedup).

### Added / नया जोड़ा गया:
- **Binary QDS Dataset Format**: Custom zero-libc dataset format with 32-byte header, tensor records, and hard dataset bomb protections ($\le 1$ GB total, $\le 100$ tensors, $\le 500$ MB per tensor).
- **Zero-Libc File I/O Engine**: Standalone Win32 I/O (`CreateFileA`, `ReadFile`, `CloseHandle`) in `internal/runtime/qadav_io.c` with path traversal defense (`..` rejection).
- **Dataset Operations**: Native keywords `dataset`, `load`, `normalize`, `minmax`, `zscore`, `shuffle`, `batch`.
- **CLI Dataset Inspection**: Command `qyn dataset inspect <path.qds>` verifying container integrity and printing tensor metadata.
- **Quasinary Computing (Invention #3)**: Balanced ternary primitive `quasi` taking values `◄` (-1), `〇` (0), `►` (+1) with saturated ternary arithmetic (`► ⊕ ► ↦ ►`, `◄ ⊕ ◄ ↦ ◄`, `► ⊕ ◄ ↦ 〇`).
- **2D Topological Grid & Energy Minimization**: 2D toroidal lattice mapping with periodic boundary conditions, Ising energy evaluation ($E = -J \sum s_i s_j$), and zero-libc simulated annealing solver (`minimize grid energy ising iterations ↦ N ·`).
- **Examples**:
  - `examples/dataset_basic.ydv`: Ingestion, normalization, shuffling, batching.
  - `examples/quasi_basic.ydv`: Saturated ternary arithmetic demonstration.
  - `examples/quasi_grid.ydv`: 3x3 topological grid Ising spin minimization.
- **Defensive Publication**: Invention #3 (Quasinary Computing) defensively published in `docs/PATENTS.md`.

---

## [M2.5 + M2.6] — Combined Milestone: Quantum Runtime Simulation & Prime-Symbolic Knowledge Engine
## [M2.5 + M2.6] — संयुक्त माइलस्टोन: क्वांटम रनटाइम सिमुलेशन एवं प्राइम-प्रतीकात्मक ज्ञान इंजन

### Added / नया जोड़ा गया:
- **Quantum Simulation Engine**: Zero-libc state vector simulator for up to 20 qubits (`internal/runtime/qadav_quantum.c`, `qadav_quantum.h`).
- **Quantum Operations**: Native keywords `qubit`, `superpose`, `entangle`, `collapse`, `measure` and quantum gates (\(\mathcal{H}\), \(\mathcal{X}\), \(\mathcal{Y}\), \(\mathcal{Z}\), \(\mathcal{S}\), \(\mathcal{T}\), \(\mathcal{C}\)).
- **Projective Measurement & Renormalization**: Wavefunction collapse with probability calculation and norm preservation.
- **Hardware-Entropy Seeded PRNG**: xorshift64 PRNG seeded via CPU timestamp cycle counter (`__builtin_readcyclecounter`) ensuring statistical measurement randomness.
- **Prime-Symbolic Knowledge Engine**: Patent-Free Invention #2 implementation with static 10,000 prime table (`internal/runtime/qadav_prime.c`, `qadav_prime.h`).
- **Semantic Concept Composition**: Concept type `concept`, semantic composition operator `⊗`, and prime factorization operator `decompose⦅...⦆`.
- **Runtime Expansion**: `runtime.Build` compiles and caches 5 modular runtime objects (`qadav_rt`, `qadav_tensor`, `qadav_nn`, `qadav_quantum`, `qadav_prime`).
- **Demonstration Examples**:
  - `examples/quantum_hello.ydv`: Demonstrates quantum superposition and projective measurement.
  - `examples/bell_state.ydv`: Demonstrates quantum entanglement (\(|\Phi^+\rangle\)) and perfect correlation.
  - `examples/concept_encoding.ydv`: Demonstrates concept composition (`concept⟨2 × 3 × 5 = 30⟩`).
  - `examples/concept_decompose.ydv`: Demonstrates concept decomposition to prime factors tensor.
- **Defensive Publication**: Invention #2 added to `docs/PATENTS.md`.

---

## [M2.3 + M2.4] — Combined Milestone: Network Declarations, Backpropagation & Probabilistic Types
## [M2.3 + M2.4] — संयुक्त माइलस्टोन: नेटवर्क घोषणाएँ, बैकप्रॉपैगैशन एवं संभाव्य प्रकार

### Added / नया जोड़ा गया:
- **`network` Declaration**: First-class language keyword and AST representation for deep neural networks (`internal/token`, `internal/ast/network.go`).
- **Standard Layers**: Dense linear transformation with Xavier initialization, ReLU, Sigmoid, Softmax, and LayerNorm (`internal/runtime/qadav_nn.c`).
- **`learn` Training Block**: Language syntax supporting dataset specification, loss functions (`mse`, `cross_entropy`), and optimizers (`sgd`, `adam`) (`internal/ast/learn.go`).
- **Zero-Libc Math Engine**: Newton-Raphson `qadav_sqrt`, Taylor polynomial with range reduction `qadav_exp` and `qadav_log` (`internal/runtime/qadav_nn.c`).
- **Automatic Differentiation**: Reverse-mode autodiff engine propagating error deltas backwards through layers with clean gradient accumulation and reset.
- **Cognitive Intermediate Representation (CIR)**: Base foundation for probabilistic types `confident⟨T⟩` and `explore⟨T⟩` (`internal/checker/prob_types.go`, `internal/types/types.go`).
- **LLVM Code Generation**: Direct translation of `NetworkDecl` to global network instances and getters, and `LearnDecl` to `@qadav_train_epochs` (`internal/codegen/network_codegen.go`).
- **Multi-Object Native Runtime**: Extended `runtime.Build` to embed and compile `qadav_nn.c` alongside `qadav_rt.c` and `qadav_tensor.c`.
- **Working Native Examples**: `examples/linear_regression.ydv`, `examples/xor_network.ydv`, and `examples/classifier.ydv`.

### Security Hardening / सुरक्षा सुदृढ़ीकरण:
- Hardened zero-libc runtime against memory overflow and invalid DLL exports.
- Numerical clamp guards in `qadav_exp` and `qadav_log` preventing NaN poisoning and divergence.
- Static layer shape validation in semantic checker preventing dimension mismatch crashes.

---

## [M2.1 + M2.2] — Tensor Type System & Operations
- Tensor types `tensor⟨T⟩⟬dims⟭`, broadcasting, matmul, transpose, reduction operators (`∑`, `∏`).

---

## [M1.1 - M1.7.1] — Phase 1 Complete
- Quantum-native Pratt parser, AST v2, zero-libc runtime, Windows/Linux native binaries.

