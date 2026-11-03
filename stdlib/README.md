# Qadav Standard Library (stdlib)

The Qadav Standard Library provides pure Qadav modules written entirely using native Qadav syntax (`.ydv`), free of legacy keywords, ASCII operators, and libc dependencies.

## Modules

### 1. `math` (`stdlib/math.ydv`)
Foundational mathematical constants and functions:
- `pi⦅⦆ → float64`: Mathematical constant $\pi \approx 3.141592653589793$
- `e⦅⦆ → float64`: Euler's number $e \approx 2.718281828459045$
- `sqrt2⦅⦆ → float64`: Square root of 2 $\approx 1.414213562373095$
- `abs_f64⦅float64 x⦆ → float64`: Absolute value of a floating-point number
- `sign_f64⦅float64 x⦆ → float64`: Signum function ($\pm 1.0$ or $0.0$)

### 2. `io` (`stdlib/io.ydv`)
Diagnostic and stream output primitives:
- `emit_text_val⦅text msg⦆ → int64`: Emits textual diagnostic message
- `emit_int_val⦅int64 n⦆ → int64`: Emits integer value

### 3. `linear_algebra` (`stdlib/linear_algebra.ydv`)
Concrete 4-element vector operations:
- `dot_product_4⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ → float64`: Vector dot product
- `vec_add_4⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ → tensor⟨float64⟩⟬4⟭`: Vector element-wise addition
- `vec_scale_4⦅tensor⟨float64⟩⟬4⟭ a⸴ float64 s⦆ → tensor⟨float64⟩⟬4⟭`: Vector scalar multiplication

### 4. `statistics` (`stdlib/statistics.ydv`)
Concrete statistical calculations:
- `mean_4⦅tensor⟨float64⟩⟬4⟭ x⦆ → float64`: Vector arithmetic mean
- `variance_4⦅tensor⟨float64⟩⟬4⟭ x⦆ → float64`: Vector sample variance

### 5. `string` (`stdlib/string.ydv`)
String utilities:
- `text_concat⦅text a⸴ text b⦆ → text`: Text concatenation
- `text_equals⦅text a⸴ text b⦆ → logic`: Text equivalence comparison
- `is_empty_text⦅text s⦆ → logic`: Text empty check

## Tooling Integration
The `qyn` compiler CLI provides dedicated stdlib commands:
```bash
qyn stdlib list
qyn stdlib build
```
