# Self-Hosted LLVM IR Codegen Specification (M8.4b)

## 1. Overview
`selfhost/codegen.ydv` provides the native self-hosted LLVM IR emission implementation in Qadav. In milestone M8.4b, the codegen achieves verified byte-identical IR generation against the reference Go compiler for `examples/hello.ydv`.

## 2. Architecture & Dual-Driver Execution
1. **Native Driver (`CodegenSourceNative`)**:
   - Compiles `selfhost/codegen.ydv` into `selfhost/codegen.exe`.
   - Invokes `selfhost/codegen.exe <file.ydv>` via sub-process execution and streams generated LLVM IR directly without invoking any Go AST code generator.
   - Surface command: `qyn selfhost-codegen <file.ydv>`.
2. **Go Reference Bridge (`CodegenSourceLegacy`)**:
   - Internal Go reference code generator (`internal/codegen`) preserved for comparative testing, regression verification, and bootstrapping validation.
   - Surface command: `qyn selfhost-codegen-legacy <file.ydv>`.

## 3. Supported Features (M8.4b Verified)
1. **Module Header & Declarations**:
   - Module header (`ModuleID = 'qadav_module'`), `source_filename = "qadav_module"`.
   - Full 44-module runtime function declarations (standard I/O, tensors, neural networks, quantum, datasets, quasi-spin, confidence, associative memory, reasoning chains, RSCA, CTSA, etc.).
2. **String Table Management**:
   - Constant string table emission with exact length and escaped null-terminator bytes (`@.str.0 = private unnamed_addr constant [43 x i8] c"..."`).
   - String pointer retrieval via `getelementptr inbounds`.
3. **Function Definitions & Entry**:
   - Process entry emission `define i32 @"main"() { entry: ... }`.
   - Runtime emit dispatch (`call void @qadav_emit_text(i8* %t1)`).
   - Exit code truncation and return instruction (`%t2 = trunc i64 0 to i32`, `ret i32 %t2`).

## 4. Verification Evidence
- Exact byte comparison between `qyn emit-llvm examples/hello.ydv` and `selfhost/codegen.exe examples/hello.ydv`:
  - `fc.exe /b TMP\hello_go.ll TMP\hello_sh.ll`: **0 byte differences** (`FC: no differences encountered`).
  - Byte length: exactly 12,559 bytes on both outputs.

## 5. Deferred Scope (To M8.4c)
The following constructs are explicitly deferred to M8.4c:
1. Dynamic AST-driven lowering for recursive loops and math operations (`examples/fib.ydv`, `examples/dot_product.ydv`).
2. Full multi-branch `match` exhaustiveness codegen.
3. Stage 1 bootstrap compiler assembly and self-compilation chain.
