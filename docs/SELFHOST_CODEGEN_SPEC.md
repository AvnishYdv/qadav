# Self-Hosted LLVM IR Codegen Specification (M8.4a)

## 1. Overview
`selfhost/codegen.ydv` provides the foundation for native self-hosted LLVM IR emission in Qadav. In milestone M8.4a, the codegen covers basic language forms including 64-bit integer and floating-point types, variable allocations, procedure definitions, emit statements, and basic arithmetic.

## 2. Supported AST Node Types (Priority 1)
1. **Program & Top-Level**:
   - `Program`: Emits module header (`ModuleID = 'qadav_module'`), source filename, string constants, runtime function declarations.
   - `ProcessDecl`: Emits LLVM function definition `define <ret_type> @"<proc_name>"(<params>) { entry: ... }`.
2. **Statements**:
   - `BindStmt`: Emits alloca for local variable and store instruction (`alloca <type>`, `store <type> <val>, <type>* <ptr>`).
   - `EmitStmt`: Dispatches to zero-libc runtime calls:
     - `@qadav_emit_int(i64)` for integers.
     - `@qadav_emit_float(double)` for floats.
     - `@qadav_emit_text(i8*)` for string literals and text pointers.
     - `@qadav_emit_logic(i32)` for boolean truth values.
   - `YieldStmt`: Emits function return instruction `ret <type> <val>` (coercing main exit code to `i32`).
3. **Expressions**:
   - `IntegerLiteral`: Emits immediate 64-bit constants (`i64 42`).
   - `FloatLiteral`: Emits immediate 64-bit IEEE-754 floating point constants (`double 3.14`).
   - `StringLiteral`: Registers `@.str.<idx>` global constant and emits `getelementptr inbounds`.
   - `Identifier`: Emits `load <type>, <type>* %<var>.ptr`.
   - `BinaryExpr`:
     - `⊕`: `add i64` / `fadd double`
     - `⊖`: `sub i64` / `fsub double`
     - `⊗`: `mul i64` / `fmul double`
     - `⊘`: `sdiv i64` / `fdiv double`
   - `CallExpr`: Emits LLVM function call `call <ret_type> @"<fn>"(<args>)`.

## 3. Deferred Scope (To M8.4b)
The following constructs are explicitly deferred to M8.4b:
1. Advanced Types: `tensor<T>`, `qubit`, `neuron`, `memory`, `graph`, `record`, `variant`.
2. Complex Constructs: `network`, `learn`, `dataset`, `infer`.
3. Control Flow Extensions: full multi-branch `match` exhaustiveness codegen, complex `cycle` loops.
4. Cryptographic & Security Inventions: all 24 inventions (QERA, QEC, AAD, EEA, SEA, RSCA, CTSA, etc.).
5. Full Stage 1 bootstrap compiler assembly.
