# Self-Hosted Type Checker Specification

## 1. Overview
`selfhost/checker.ydv` provides the native Qadav implementation of the type checker. It mirrors the core capabilities of the Go reference checker in `internal/checker/`.

## 2. Supported Type Rules
1. `type_of`: Basic type resolution for primitives (`int64`, `float64`, `logic`, `text`, `quasinary`, `quantum`, `tensor`).
2. `unify`: Exact type equality and unification.
3. `subtype`: Subtyping relationships for tensors and variants.
4. `check_signature`: Formal parameter count and type correspondence.
5. `check_record_field`: Field lookup in declared record schema.
6. `check_variant_case`: Case existence and payload type matching in declared variant schema.
7. `check_match_exhaustiveness`: Exhaustive coverage of all variant branches.
8. `check_binary_expr`: Operand type validation for quantum algebraic operators (`⊕`, `⊖`, `⊗`, `⊘`, `⊙`, `⊛`, `≡`, `≠`, `∧`, `∨`).
9. `check_program`: Whole-program verification and error accumulation.

## 3. Byte-Identical Output
The command `qyn selfhost-check <file.ydv>` executes `selfhost/checker.ydv` (via the self-hosted driver) and produces byte-for-byte identical output to `qyn check <file.ydv>`.
