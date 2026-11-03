# Qadav Standard Library Specification (stdlib)

## 1. Overview
The Qadav Standard Library (`stdlib/`) provides essential mathematical, diagnostic, linear algebra, statistical, and text utilities implemented entirely in pure Qadav syntax (`.ydv`). It contains zero external dependencies, zero libc headers, and strictly adheres to Qadav native grammar.

## 2. Modules

### 2.1 `math` (`stdlib/math.ydv`)
Foundational mathematical constants and scalar functions.

```qadav
process pi⦅⦆ → float64 ⟦
    yield 3.141592653589793 ·
⟧

process e⦅⦆ → float64 ⟦
    yield 2.718281828459045 ·
⟧

process sqrt2⦅⦆ → float64 ⟦
    yield 1.414213562373095 ·
⟧

process abs_f64⦅float64 x⦆ → float64 ⟦
    when x ≺ 0.0 ⟦
        yield ⊖x ·
    ⟧
    yield x ·
⟧

process sign_f64⦅float64 x⦆ → float64 ⟦
    when x ≺ 0.0 ⟦
        yield ⊖1.0 ·
    ⟧
    when x ≻ 0.0 ⟦
        yield 1.0 ·
    ⟧
    yield 0.0 ·
⟧
```

### 2.2 `io` (`stdlib/io.ydv`)
Diagnostic and stream output helpers wrapping native `emit`.

```qadav
process emit_text_val⦅text msg⦆ → int64 ⟦
    emit msg ·
    yield 0 ·
⟧

process emit_int_val⦅int64 n⦆ → int64 ⟦
    emit n ·
    yield 0 ·
⟧
```

### 2.3 `linear_algebra` (`stdlib/linear_algebra.ydv`)
Concrete 4-element vector operations utilizing native tensor arithmetic operators.

```qadav
process dot_product_4⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ → float64 ⟦
    yield ∑ ⦅a ⊗ b⦆ ·
⟧

process vec_add_4⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ → tensor⟨float64⟩⟬4⟭ ⟦
    yield a ⊕ b ·
⟧

process vec_scale_4⦅tensor⟨float64⟩⟬4⟭ a⸴ float64 s⦆ → tensor⟨float64⟩⟬4⟭ ⟦
    yield a ⊗ s ·
⟧
```

### 2.4 `statistics` (`stdlib/statistics.ydv`)
Concrete statistical calculations (mean and sample variance) for 4-element vectors.

```qadav
process mean_4⦅tensor⟨float64⟩⟬4⟭ x⦆ → float64 ⟦
    float64 s ↦ ∑ x ·
    yield s ⊘ 4.0 ·
⟧

process variance_4⦅tensor⟨float64⟩⟬4⟭ x⦆ → float64 ⟦
    float64 m ↦ mean_4⦅x⦆ ·
    tensor⟨float64⟩⟬4⟭ dev ↦ x ⊖ m ·
    tensor⟨float64⟩⟬4⟭ sq ↦ dev ⊗ dev ·
    float64 sum_sq ↦ ∑ sq ·
    yield sum_sq ⊘ 4.0 ·
⟧
```

### 2.5 `string` (`stdlib/string.ydv`)
String concatenation, equality comparison, and empty-check utilities.

```qadav
process text_concat⦅text a⸴ text b⦆ → text ⟦
    yield a ⊕ b ·
⟧

process text_equals⦅text a⸴ text b⦆ → logic ⟦
    when a ≡ b ⟦
        yield ⊤ ·
    ⟧
    yield ⊥ ·
⟧

process is_empty_text⦅text s⦆ → logic ⟦
    when s ≡ "" ⟦
        yield ⊤ ·
    ⟧
    yield ⊥ ·
⟧
```

## 3. Tooling and Verification
The compiler CLI provides standard library commands:
- `qyn stdlib list`: Enumerates registered standard library modules.
- `qyn stdlib build`: Parses and type-checks all standard library modules.

## 4. Cognitive Memory Consolidation & QEC Integration
Standard library modules integrate with Qadav native cognitive primitives:
- Prime-symbolic mappings used in `stdlib/math.ydv` align with semantic coordinates generated during `consolidate episodic into semantic` memory promotion.
- Diagnostic I/O routines in `stdlib/io.ydv` interface with verified quantum syndrome measurement pipelines (`syndrome`, `verify_syndrome`, `correct`).

## 5. Scoped Access & Safe Tool Integration
Standard library routines operate securely across privacy and capability boundaries:
- `recall_scoped` operations protect confidential agent memory during standard library analytics.
- `tool_register` and `tool_call` allow standard library functions to be safely invoked with formal QERA attestation.

## 6. Type-Tagged Proofs & Calibrated Abstention Integration
Standard library algorithms support formal attestation and epistemic bounds:
- `certified_reason` statements can attach formal mathematical witnesses to computations executed by standard library modules (e.g. `stdlib/math.ydv` or `stdlib/linear_algebra.ydv`).
- `abstain_check` policies protect critical analytics against low-confidence reasoning results, cleanly executing `retry`, `escalate`, or `refuse` without crashing or triggering quantum decoherence.


