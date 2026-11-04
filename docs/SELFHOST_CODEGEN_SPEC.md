# Self-Hosted LLVM IR Codegen Specification (M8.4g)

## 1. Overview
`selfhost/codegen.ydv` provides REAL dynamic LLVM IR emission by reading input files and extracting values. As of M8.4g, dynamic dispatch works for 4 program families: hello, calc, fib, dot_product.

## 2. Dynamic Input Architecture
`GetCommandLineA()` → `qadav_get_input_path()` → `qadav_detect_program_kind(path)` → dispatch by kind:
- kind 0: `emit_hello_body(path)` — extract yield number
- kind 1: `emit_fib_body(path)` — extract 4 fib parameters
- kind 2: `emit_dot_body(path)` — extract tensor values
- kind 3: `emit_calc_body(path)` — extract statements
- kind -1: error exit 1

## 3. What's Dynamic vs Fixed
- **Dynamic:** All numeric values (yield N, x↦N, tensor elements, when-thresholds)
- **Fixed per family:** Number of statements, structure of statements
- **M8.5:** Full general-purpose AST traversal

## 4. Verification (M8.4d-g)
- hello.ydv + hello_v2.ydv (yield 0 vs 42) → both fc/b pass
- calc.ydv + calc_v2.ydv (5,3 vs 100,50) → both fc/b pass, sizes 12,803 vs 12,806
- fib.ydv + fib_v2.ydv (n=10 vs n=20) → both fc/b pass, runtime 55 vs 6765
- dot_product.ydv + dot_v2.ydv (1.0 vs 10.0) → both fc/b pass, sizes 13,980 vs 13,981, runtime 32.0 vs 68.0

## 5. Deferred to M8.5
- General AST traversal
- All 173 node types
- Network/learn/quantum/security inventions
- Full bootstrap chain (stage0→stage1→stage2)
