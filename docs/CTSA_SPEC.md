# CTSA: Compile-Time Type Safety Attestation (Invention #24)
## Specification and Technical Architecture

### 1. Abstract
Compile-Time Type Safety Attestation (CTSA) establishes a cryptographic guarantee that program type safety invariants validated by the Qadav type checker are preserved end-to-end through code generation, native LLVM compilation, and execution. CTSA produces a 120-byte preimage attestation certificate containing SHA3-256 hashes of the verified AST structure, type environment, and a rolling rule application chain.

### 2. Threat Model
CTSA defends against:
1. Compiler back-end corruption or code generator tampering.
2. Binary patching altering type dispatch or layout offsets.
3. Pre-main type-tag substitution attacks.

### 3. Certificate Data Structure
The CTSA certificate consists of 7 primary fields:
- `ast_hash` (32 bytes): SHA3-256 digest of canonical AST.
- `type_env_hash` (32 bytes): SHA3-256 digest of global and procedure type environments.
- `rule_chain_hash` (32 bytes): SHA3-256 rolling digest of applied inference/subtyping rules.
- `num_signatures` (8 bytes LE): Count of checked function/procedure signatures.
- `num_fields` (8 bytes LE): Count of checked record fields.
- `timestamp` (8 bytes LE): Compile-time RDTSC hardware clock tick.
- `cert_hash` (32 bytes): SHA3-256 digest across the 120-byte certificate preimage.

### 4. Rolling Rule Chain
Every type judgment applied by the checker updates a cumulative rolling hash:
`H_{i+1} = SHA3-256(H_i || RuleID || TypeID)`
This produces a deterministic, tamper-evident proof trace of type checking steps.

### 5. Runtime Verification
Prior to executing the procedure body or upon invocation of `ctsa_verify`, the runtime recomputes the expected SHA3-256 certificate digest from the embedded fields.

### 6. Decoherence Triggering
If any byte of the CTSA certificate is modified or fails verification, the runtime immediately invokes:
- Windows: `ExitProcess(0xDEADBEEF)`
- Linux/POSIX: `__builtin_trap()` / `SIGILL`
No memory is leaked, and no tampered code executes.

### 7. Language Syntax
```qadav
process main⦅⦆ → int64 ⟦
    ctsa_attest ast ↦ "ast_tree" env ↦ "type_env" rules ↦ "type_rules" into cert0 ·
    ctsa_verify cert0 into v0 ·
    emit "CTSA certificate verified" ·
    yield 0 ·
⟧
```

### 8. Test Mode Mutation
In `--test` mode, the diagnostic keyword `corrupt_ctsa` is permitted:
```qadav
corrupt_ctsa cert0 field ↦ 1 ·
```
Outside `--test` mode, usage of `corrupt_ctsa` causes a compilation error.

### 9. Zero-libc Runtime Integration
Implemented in zero-libc C within `internal/runtime/qadav_ctsa.c` and `internal/runtime/qadav_ctsa.h`, compiled into `qadav_ctsa.o` as runtime module #44.

### 10. Self-Hosted Qadav Representation
Defined as a native Qadav record in `selfhost/ctsa.ydv`, verifying that self-hosted stages can emit and inspect CTSA certificates without host dependencies.

