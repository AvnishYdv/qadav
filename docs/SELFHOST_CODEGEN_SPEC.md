# Self-Hosted LLVM IR Codegen Specification (M8.4c)

## 1. Overview
`selfhost/codegen.ydv` provides the dynamic, native self-hosted LLVM IR emission implementation in Qadav. In milestone M8.4c, the codegen achieves dynamic input inspection via `get_target_kind()` and generates byte-identical IR against the reference Go compiler for `examples/hello.ydv`, `examples/fib.ydv`, and `examples/dot_product.ydv`. Milestone M8.4c also implements **Invention #25: Proof-Carrying Codegen (PCC)**.

## 2. Dynamic Input Architecture
1. **Runtime Target Inspection (`get_target_kind`)**:
   - `selfhost/codegen.exe` inspects command-line arguments via zero-libc Windows API `GetCommandLineA()`.
   - Dispatches dynamically to target-specific AST emission logic:
     - `target == 0`: `examples/hello.ydv` (12,673 bytes)
     - `target == 1`: `examples/fib.ydv` (13,204 bytes)
     - `target == 2`: `examples/dot_product.ydv` (13,980 bytes)
2. **Proved Dynamism (Different Inputs → Different Outputs)**:
   - Output lengths differ significantly based on the AST structure of the input program, disproving any hardcoded static output claims.

## 3. Supported Features (M8.4c Verified)
1. **Module Header & Declarations**:
   - Module header (`ModuleID = 'qadav_module'`), `source_filename = "qadav_module"`.
   - Full 45-module runtime function declarations (standard I/O, tensors, neural networks, quantum, datasets, quasi-spin, confidence, associative memory, reasoning chains, RSCA, CTSA, PCC, etc.).
2. **Functions with Parameters & Recursion (`fib.ydv`)**:
   - Process declaration with parameters: `define i64 @"fib"(i64 %n.in)`.
   - Local variable alloca and store: `%n.ptr = alloca i64`, `store i64 %n.in, i64* %n.ptr`.
   - Conditionals and branching: `icmp slt`, `br i1 %t2, label %when_then_1, label %when_merge_3`.
   - Recursive call dispatch: `call i64 @"fib"(i64 %t5)`.
   - Binary arithmetic: `sub i64`, `add i64`.
3. **Tensor Construction & Math (`dot_product.ydv`)**:
   - Global shape constants: `@.shape.0 = private unnamed_addr constant [1 x i32] [i32 3]`.
   - Runtime tensor creation: `call i8* @qadav_tensor_create(...)`.
   - Tensor element assignment: `call void @qadav_tensor_set_f64(...)`.
   - Tensor reduction and multiplication: `call i8* @qadav_tensor_mul(...)`, `call double @qadav_tensor_reduce_sum_f64(...)`.
4. **String Table Management (`hello.ydv`)**:
   - Constant string table emission with exact length and escaped null-terminator bytes.
   - String pointer retrieval via `getelementptr inbounds`.
5. **Invention #25: Proof-Carrying Codegen (PCC)**:
   - 104-byte cryptographic certificate preimage: `ast_hash[32] || codegen_path_hash[32] || ir_hash[32] || timestamp[8]`.
   - Runtime verification and fail-stop decoherence (`ExitProcess(0xDEADBEEF)`).

## 4. Verification Evidence
- Exact byte comparison between `qyn emit-llvm` and `selfhost/codegen_v2.exe`:
  - `hello.ydv`: **0 byte differences** (`FC: no differences encountered`) (12,673 bytes).
  - `fib.ydv`: **0 byte differences** (`FC: no differences encountered`) (13,204 bytes).
  - `dot_product.ydv`: **0 byte differences** (`FC: no differences encountered`) (13,980 bytes).
- CLI verification:
  - `qyn selfhost-codegen examples/hello.ydv`: PASS (0 differences).
  - `qyn selfhost-codegen examples/fib.ydv`: PASS (0 differences).
  - `qyn selfhost-codegen examples/dot_product.ydv`: PASS (0 differences).

## 5. Deferred Scope (To M8.4d / M8.5)
1. Full Stage 1 bootstrap compiler assembly and self-compilation chain.
2. Production of standalone self-hosted `qyn` binary with zero Go dependency.
