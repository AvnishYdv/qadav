# Phase 8 Roadmap: Self-Hosting & Autonomous Compiler Evolution

## Objective
Transition Qadav from a Go-hosted reference language into an autonomous, self-hosted, quantum-native compiler written entirely in Qadav (`.ydv`), verified by Recursive Self-Compilation Attestation (RSCA).

## Milestone Trajectory

### M8.1 + M8.2: Self-Hosting Prerequisites + Self-Hosted Lexer/Parser + Invention #23 (RSCA) [COMPLETED]
- Qadav-native data structures: `record`, `variant`, `match`, `record_init`, `variant_init`, field access (`.` / `∸`).
- Invention #23 (RSCA): 144-byte cryptographic preimage, 3-stage bootstrap, constant-time output comparison, decoherence abort (`0xDEADBEEF`).
- 43 zero-libc runtime objects.
- Self-hosted lexer (`selfhost/lexer.ydv`) and parser (`selfhost/parser.ydv`) bootstrap seeds.

### M8.3 + M8.4: Self-Hosted Checker + Codegen + Invention #24
- Complete self-hosted type checker and symbol resolver in Qadav.
- Self-hosted LLVM IR emission and backend interface in Qadav.
- Invention #24: Deterministic Compiler Irreversibility Attestation.

### M8.5 + M8.6: Stage 1 Compiler Bootstrap & Full Toolchain Cutover + Invention #25
- Stage 0 $\to$ Stage 1 full compilation cutover.
- Production of standalone self-hosted `qyn` binary with zero Go dependency.
- Fixed-point byte-for-byte RSCA verification across Stage 1 and Stage 2.
- Phase 8 Grand Finale & Autonomous Self-Hosting Sign-off.
