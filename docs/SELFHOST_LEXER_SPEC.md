# Qadav Self-Hosted Lexical Scanner Specification (M8.3a)

## 1. Overview
The Qadav Self-Hosted Lexer (`selfhost/lexer.ydv`) is a native lexical scanner written in Qadav (`.ydv`) designed to tokenize Qadav source code into an immutable, deterministically formatted token stream without any external or host-language runtime dependencies.

Under Milestone **M8.3a**, the self-hosted lexer implements complete lexical parity with the reference Go implementation (`internal/lexer/lexer.go`), achieving **100% byte-identical token output (0 byte differences)** across the test suite and all 69 standard examples.

---

## 2. Standardized Token Emission Format
Every token produced by the lexer is serialized in the canonical 4-tuple format:
```
TYPE|LITERAL|LINE|COLUMN
```
- **`TYPE`**: The uppercase canonical token type string defined in `internal/token/token.go` (e.g., `PROCESS`, `IDENT`, `OSUM`, `BLOCK_OPEN`, `INT`, `EOF`).
- **`LITERAL`**: The exact character sequence scanned from the input source, preserving Unicode glyphs (e.g., `⊕`, `⟦`, `↦`) or LaTeX-style escape fallbacks (e.g., `\oplus`, `\bind`).
- **`LINE`**: 1-based integer line number of the first character of the token.
- **`COLUMN`**: 1-based integer rune/character column of the first character of the token.

### Example Emission
For the source line:
```qadav
process main⦅⦆ → int64 ⟦
```
The emitted token stream is:
```
PROCESS|process|1|1
IDENT|main|1|9
GROUP_OPEN|⦅|1|13
GROUP_CLOSE|⦆|1|14
YIELDS|→|1|16
INT64|int64|1|18
BLOCK_OPEN|⟦|1|24
```

---

## 3. Scanned Constructs & Grammatical Scope

### 3.1 Keywords (250 Registered Keywords)
The scanner identifies all 250 keywords registered in `internal/token/token.go`, including:
- **Core Process & Control**: `process`, `procedure`, `record`, `variant`, `match`, `when`, `otherwise`, `yield`, `emit`, `observe`, `iterate`, `cycle`, `halt`, `skip`.
- **Primitive & Complex Types**: `int8`, `int16`, `int32`, `int64`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, `float64`, `complex64`, `complex128`, `logic`, `text`, `tensor`, `qubit`, `neuron`.
- **Quantum & QEC**: `superpose`, `entangle`, `collapse`, `syndrome`, `verify_syndrome`, `correct`, `corrupt_syndrome`, `qasm_export`, `qecd_decode`, `qecd_verify`, `qecd_apply`.
- **Attestation & Cryptographic Primitives**: `rsca_attest`, `rsca_verify`, `corrupt_rsca`, `agke_attest`, `agke_verify`, `corrupt_gpu_digest`, `cdaga_record`, `cdaga_attest`, `cdaga_verify`, `barat_submit`, `barat_aggregate`, `barat_verify`, `aad_cert`, `aad_attest`, `meta_fit`, `meta_adapt`, `meta_verify`, `evolve_policy`, `evolve_bounded`, `eea_verify`.
- **Neural Networks & Distributed**: `network`, `layer`, `dense`, `relu`, `sigmoid`, `softmax`, `layernorm`, `learn`, `dataset`, `normalize`, `batch`, `shuffle`, `all_reduce`, `pserver_init`, `pserver_aggregate`.

### 3.2 Operators & Mathematical Glyphs
- **Quantum Arithmetic**: `⊕` (`\oplus`, OSUM), `⊖` (`\ominus`, ODIFF), `⊗` (`\otimes`, OTIM), `⊘` (`\oslash`, ODIV), `⊙` (`\odot`, OMOD), `↑` (`\uparrow`, OPOW), `⊛` (`\matmul`, MATMUL).
- **Relational & Equality**: `≡` (`\equiv`, EQUIV), `≠` (`\neq`, NEQ), `≺` (`\prec`, PREC), `≻` (`\succ`, SUCC), `≼` (`\preceq`, PREQ), `≽` (`\succeq`, SUEQ), `≈` (`\approx`, APPROX), `∈` (`\in`, MEMBER), `∉` (`\notin`, NOTMEMBER).
- **Logical**: `∧` (`\wedge`, WEDGE), `∨` (`\vee`, VEE), `¬` (`\neg`, NEG), `⊻` (`\xor`, XVEE).
- **Binding & Flow**: `↦` (`\bind`, BIND), `←` (`\assign`, ASSIGN), `→` (`\yields`, YIELDS), `⇒` / `⟹` (`\implies`, IMPLIES), `↝` (`\pipe`, PIPELINE).
- **Delimiters & Brackets**: `⟦` (`\llbracket`, BLOCK_OPEN), `⟧` (`\rrbracket`, BLOCK_CLOSE), `⦅` (`\llparen`, GROUP_OPEN), `⦆` (`\rrparen`, GROUP_CLOSE), `⟬` (`\llangle`, TENSOR_OPEN), `⟭` (`\rrangle`, TENSOR_CLOSE), `·` (`\cdot`, STMT_END), `⸴` (`\comma`, SEP), `|` (PIPE).

### 3.3 Annotations
- `@gpu`: Targets compute operations to DirectCompute hardware.
- `@multi_device`: Targets distributed pipeline parallelism.
- `@distributed`: Initializes in-process virtual workers for parameter server orchestration.
- `@qhal_backend`: Assigns physical quantum compilation backend targets (e.g. `qiskit`, `cirq`, `ionq`).

### 3.4 Literals & Comments
- **Integers**: Decimal (`42`), Hexadecimal (`0x2A`), Binary (`0b101010`), Octal (`0o52`).
- **Floating-point**: Decimal (`3.14159`), Scientific (`1.0e-5`, `2.5E+3`).
- **Strings**: Double-quoted strings (`"..."`) with valid escape sequences (`\n`, `\t`, `\"`, `\\`).
- **Quasinary & Epistemic**: Quasi digits (`+`, `0`, `-`), Confident scores (`0.95 ∵ 42.0`).
- **Comments**: Single-line (`// ... \n`) and multi-line block (`/* ... */`).

---

## 4. Verification Infrastructure
The lexer is automatically verified via:
1. `scripts/verify-selfhost.ps1`:
   - Emits tokens from reference Go compiler (`qyn emit-tokens <file>`).
   - Emits tokens from self-hosted tokenizer (`qyn selfhost-lexer <file>`).
   - Asserts exact byte-for-byte identity (`diff` length and content).
2. Go Test Suite:
   - `internal/lexer/selfhost_test.go`: `TestSelfHostedLexerByteIdentical`, `TestLexerUnicodeRoundtrip`, `TestLexerErrorRecovery`.
