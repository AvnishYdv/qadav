# Qadav Self-Hosted Parser Specification (M8.3a)

## 1. Overview
The Qadav Self-Hosted Core Parser (`selfhost/parser.ydv`) is a native parser written in Qadav (`.ydv`) designed to convert lexical token streams into immutable, validated Abstract Syntax Tree (AST) nodes as defined in `selfhost/ast_nodes.ydv`.

Under Milestone **M8.3a**, the core self-hosted parser achieves **100% byte-identical AST S-expression output (0 byte differences)** against the reference Go parser across the selfhost verification suite.

---

## 2. Canonical S-Expression AST Serialization
To guarantee deterministic, whitespace-normalized, byte-for-byte equivalence, AST nodes are serialized into canonical S-expressions:
```scheme
(Program
  (ProcessDecl "main" () (PrimitiveType "int64")
    (Block
      (Yield (IntLit 0)))))
```

### Grammar Serialization Rules
- **Program**: `(Program <decl1> <decl2> ...)`
- **Process Declaration**: `(ProcessDecl "<name>" (<param1> ...) <return_type> <block>)`
- **Procedure Declaration**: `(ProcedureDecl "<name>" (<param1> ...) <return_type> <block>)`
- **Record Declaration**: `(RecordDecl "<name>" ((Field "<f1>" "<t1>") ...))`
- **Variant Declaration**: `(VariantDecl "<name>" ((Case "<c1>" ((Field "<f1>" "<t1>") ...)) ...))`
- **Network Declaration**: `(NetworkDecl "<name>" ((Layer "<kind>" <in_dim> <out_dim>) ...))`
- **Block**: `(Block <stmt1> <stmt2> ...)`
- **Bind Statement**: `(Bind "<name>" <type> <expr>)`
- **Assign Statement**: `(Assign <lhs> "<op>" <rhs>)`
- **Emit Statement**: `(Emit <expr>)`
- **Yield Statement**: `(Yield <expr>)`
- **When Statement**: `(When <condition> <then_block> [<else_block>])`
- **Match Statement**: `(Match <subject> ((Arm "<case>" ("<binding>") <block>) ...) [(Otherwise <block>)])`
- **Quantum Gates**: `(Gate "<op>" <target>)`, `(CNOT <control> <target>)`
- **Attestation Primitives**: `(RSCAAttest <stage> "<source>" "<compiler>" "<output>" "<into>")`, `(RSCAVerify "<record>" "<into>")`
- **Expressions**: `(BinaryExpr "<op>" <left> <right>)`, `(UnaryExpr "<op>" <expr>)`, `(CallExpr <callee> (<arg1> ...))`, `(RecordInit "<name>" ((FieldInit "<field>" <val>) ...))`, `(VariantInit "<variant>" "<case>" <val>)`, `(FieldAccess <object> "<field>")`.

---

## 3. Supported vs. Deferred Constructs (M8.3a Scope)

### 3.1 Fully Supported in M8.3a
- **Declarations**: `process`, `procedure`, `record`, `variant`, `network`, `qubit`.
- **Control Flow**: `when ... otherwise`, `iterate ... ∈`, `cycle`, `match ... ⟦ | Case ⟹ ... ⟧`.
- **Statements**: Immutable binding (`↦`), mutable assignment (`←`), `emit`, `observe`, `yield`, `halt`, `skip`.
- **Expressions**: Complete Pratt precedence hierarchy for binary operators (`OSUM`, `ODIFF`, `OTIM`, `ODIV`, `EQUIV`, `NEQ`), prefix unary operators (`NEG`, `OSUM`, `ODIFF`), function calls, record construction (`record_init`), variant construction (`variant_init`), field access (`.`).
- **Quantum Primitives**: Qubit declaration, `superpose`, `entangle`, `collapse`.
- **Security & Attestation**: RSCA self-compilation attestation (`rsca_attest`, `rsca_verify`, `corrupt_rsca`).

### 3.2 Deferred to Milestone M8.3b
- **Full Type-Checker in Qadav**: Bidirectional type checking, subtyping, and generic instantiation in `.ydv`.
- **Symbol Table & Scope Environment**: Persistent nested lexical scoping engine written natively in Qadav.
- **Invention #24 (Autonomous Self-Hosting Verification Engine)**: Full fixed-point binary verification harness.

---

## 4. Verification Methodology
AST equivalence is verified automatically via:
1. `scripts/verify-selfhost.ps1`:
   - Runs `qyn emit-ast <file.ydv>` (Go reference implementation).
   - Runs `qyn selfhost-parser <file.ydv>` (Self-hosted parser engine).
   - Asserts 0 byte differences between the two canonical S-expressions.
2. Go Test Suite:
   - `internal/parser/selfhost_test.go`: `TestSelfHostedParserByteIdentical`, `TestParserErrorRecovery`.

