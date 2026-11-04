# Phase 8 Roadmap: Self-Hosting & Autonomous Compiler Evolution

## Objective
Transition Qadav from a Go-hosted reference language into an autonomous, self-hosted, quantum-native compiler written entirely in Qadav (`.ydv`), verified by Recursive Self-Compilation Attestation (RSCA).

## Milestone Trajectory

### M8.1 + M8.2: Self-Hosting Prerequisites + Self-Hosted Lexer/Parser Seeds + Invention #23 (RSCA) [COMPLETED]
- Qadav-native data structures: `record`, `variant`, `match`, `record_init`, `variant_init`, field access (`.` / `∸`).
- Invention #23 (RSCA): 144-byte cryptographic preimage, 3-stage bootstrap, constant-time output comparison, decoherence abort (`0xDEADBEEF`).
- 43 zero-libc runtime objects.
- Self-hosted lexer (`selfhost/lexer.ydv`) and parser (`selfhost/parser.ydv`) bootstrap seeds.

### M8.3a: Self-Hosted Lexer Completion + AST Definitions + Core Parser (Byte-Identical Proof) [COMPLETED]
- Complete self-hosted lexer in Qadav (`selfhost/lexer.ydv`): 250 keywords, 108 fallbacks, Unicode operators, annotations, literals, and comments.
- Complete AST definitions in Qadav (`selfhost/ast_nodes.ydv`): 173 AST structs mapped to records and variants, compiled cleanly to native binary.
- Core self-hosted parser in Qadav (`selfhost/parser.ydv`): recursive-descent and Pratt precedence, emitting canonical AST S-expressions.
- Automated verification (`scripts/verify-selfhost.ps1`): 100% byte-identical against Go reference across tests and all 69 examples.

### M8.3b: M8.3a Errors Fix + Self-Hosted Checker + Invention #24 (CTSA) [COMPLETED]
- Phase 0 verification gate resolved: F29 fake diff replaced with actual `fc.exe /b` outputs; F30 3 required examples verified; F31 AST count reconciled; F32 69/69 examples verified for parser.
- Self-hosted type checker in Qadav (`selfhost/checker.ydv`): symbol environments, unification, subtyping, and exhaustiveness checking producing byte-identical output to Go reference via `qyn selfhost-check`.
- Invention #24 (CTSA): Compile-Time Type Safety Attestation certificate with 7 fields, 120-byte preimage, rolling rule chain, runtime verification, and fail-stop decoherence (`0xDEADBEEF`).
- 44 zero-libc runtime objects (`qadav_ctsa.c`).
- 24 patents in portfolio (18 pending + 6 defensive).

### M8.4a: Self-Hosted Codegen Basic + F33 Checker Coverage Fix [COMPLETED]
- Phase 0: F33 checker coverage gap fixed to ≥84.8% (84.8% verified).
- Phase 1: Self-hosted codegen basic in Qadav (`selfhost/codegen.ydv`): int64/float64 arithmetic, functions, emit, variable bindings.
- CLI subcommands: `qyn emit-llvm <file.ydv>` and `qyn selfhost-codegen <file.ydv>`.
- Phase 2: Byte-identical verification on 5 required examples (`hello`, `fib`, `dot_product`, `linear_regression`, `classifier`) with 0-byte difference via `fc.exe /b`.
- 44 zero-libc runtime objects, 24 patents preserved.

### M8.4b: Full Self-Hosted Codegen + Bootstrap Assembly + Invention #25
- Full LLVM IR codegen in Qadav: tensors, neural networks, quantum circuits, control flow (`when`/`cycle`), and all 24 security inventions.
- Invention #25 (Patent Pending).
- Stage 1 compiler frontend & backend assembly.

### M8.5 + M8.6: Stage 1 Compiler Bootstrap & Full Toolchain Cutover + Invention #25
- Stage 0 $\to$ Stage 1 full compilation cutover.
- Production of standalone self-hosted `qyn` binary with zero Go dependency.
- Fixed-point byte-for-byte RSCA verification across Stage 1 and Stage 2.
- Phase 8 Grand Finale & Autonomous Self-Hosting Sign-off.
