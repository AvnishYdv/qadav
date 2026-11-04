# Qadav Self-Hosting & RSCA Bootstrap Architecture (Phase 8)

## Overview
Phase 8 initiates the migration of the Qadav compiler (`qyn`) from a Go-hosted implementation to a fully self-hosted, quantum-native compiler written entirely in Qadav (`.ydv`).

To safeguard compiler integrity against untrusted modification, supply-chain contamination, and silent binary drifting, Phase 8 couples self-hosting with **Invention #23: Recursive Self-Compilation Attestation (RSCA)**.

## Three-Stage Bootstrap Pipeline

1. **Stage 0 (Go Bootstrap Compiler):**
   - The verified Go reference implementation (`bin/qyn.exe`).
   - Compiles the Qadav self-hosted lexer (`selfhost/lexer.ydv`), parser (`selfhost/parser.ydv`), and downstream modules.
   - Outputs the Stage 1 compiler binary (`bin/stage1_qyn.exe`).
   - Emits an RSCA Stage 0 cryptographic attestation record.

2. **Stage 1 (Qadav-Compiled Compiler):**
   - The Stage 1 binary produced by Stage 0.
   - Compiles the same Qadav compiler sources.
   - Outputs the Stage 2 compiler binary (`bin/stage2_qyn.exe`).
   - Emits an RSCA Stage 1 cryptographic attestation record.

3. **Stage 2 (Self-Compiled Compiler):**
   - The Stage 2 binary produced by Stage 1.
   - Compiles the same Qadav compiler sources.
   - Outputs the Stage 3 compiler binary (`bin/stage3_qyn.exe`).
   - Emits an RSCA Stage 2 cryptographic attestation record.

## Recursive Self-Compilation Attestation (RSCA) Verification

Under RSCA:
- A 144-byte cryptographic preimage binds:
  - Stage identifier (`uint64`, 8 bytes)
  - SHA3-256 digest of source tree (32 bytes)
  - SHA3-256 digest of executing compiler (32 bytes)
  - SHA3-256 digest of emitted compiler output (32 bytes)
  - Monotonic timestamp (`uint64`, 8 bytes)
  - SHA3-256 digest of previous stage attestation record (32 bytes)
- Fixed point identity requirement:
  $$\text{Hash}(\text{Output}_{\text{Stage 1}}) = \text{Hash}(\text{Output}_{\text{Stage 2}})$$
- Output binaries between Stage 1 and Stage 2 must match **byte-for-byte in constant time**.
- Any hash divergence, bit flip, or unauthorized modification triggers **immediate active quantum decoherence** (`ExitProcess(0xDEADBEEF)`), permanently collapsing execution and preventing compromised compiler distribution.

## Self-Hosted Modules (Milestone M8.4a)

1. `selfhost/lexer.ydv`: Complete self-hosted lexical scanner in Qadav. Matches all 250 keywords, 108 LaTeX fallbacks, Unicode operators, annotations, and literals. Emits tokens in standardized 4-tuple format (`TYPE|LITERAL|LINE|COLUMN`).
2. `selfhost/ast_nodes.ydv`: Complete Abstract Syntax Tree definitions in Qadav. Maps all 173 AST structs to native `record` and `variant` types.
3. `selfhost/parser.ydv`: Core self-hosted parser in Qadav with recursive-descent and Pratt precedence, outputting canonical S-expressions.
4. `selfhost/checker.ydv`: Full self-hosted type checker in Qadav implementing symbol resolution, unification, subtyping, and exhaustiveness checking.
5. `selfhost/codegen.ydv`: Self-hosted basic LLVM IR codegen in Qadav for int64/float64 arithmetic, functions, emit, and variable binding.
6. `selfhost/ctsa.ydv`: Native Qadav implementation of Invention #24: Compile-Time Type Safety Attestation (CTSA).
7. `selfhost/tests/`: Self-hosted test suite (`test_basic.ydv`, `test_expressions.ydv`, `test_network.ydv`, `test_quantum.ydv`, `test_security.ydv`, `test_unicode.ydv`, `test_edge.ydv`).
8. `scripts/verify-selfhost.ps1`: Automated byte-identical comparison harness comparing Go reference vs self-hosted outputs.

