# Cognitive Intermediate Representation (CIR) Specification

**Patent-Free Defensive Invention #4**  
**Language:** Qadav Native Language  
**Runtime:** Freestanding Zero-Libc Native Runtime  

---

## 1. Abstract & Motivation

The Cognitive Intermediate Representation (CIR) extends low-level native program execution with calibrated probabilistic reasoning, epistemic tracking, and uncertainty propagation. Unlike traditional programming languages that decouple deterministic execution from AI inference, Qadav integrates confidence semantics directly into its native type system and freestanding runtime.

Every critical cognitive decision, neural inference step, and multi-step derivation maintains explicit epistemic provenance without external library dependencies.

---

## 2. Epistemic States & Types

CIR defines three fundamental epistemic states:

1. **`KNOWN` (`0`):** Grounded fact or validated inference with calibrated confidence score $s \in [0.0, 1.0]$.
2. **`UNKNOWN` (`1`):** Complete absence of evidence or invalid provenance. Always propagates zero confidence ($s = 0.0$).
3. **`EXPLORE` (`2`):** Hypothesis under active exploration or heuristic evaluation.

### Type Syntax

- `confident⟨T⟩`: Holds a value of type `T` paired with a calibrated confidence score $\in [0.0, 1.0]$ and epistemic state.
- `explore⟨T⟩`: Holds an exploratory hypothesis with exploration metadata.

### Literal Syntax

Calibrated confidence literals use the reason operator `∵` (`\because`):

```ydv
confident⟨float64⟩ a ↦ 0.95 ∵ 42.0 ·
```

Here, `0.95` is the confidence score and `42.0` is the underlying value.

---

## 3. Propagation Algebra

CIR defines closed mathematical operations for confidence propagation during arithmetic operations:

### Tensor / Multiplication ($\otimes$)
For independent joint occurrences or sequential reasoning steps:
$$\text{val}_{res} = \text{val}_a \times \text{val}_b$$
$$\text{score}_{res} = \text{score}_a \times \text{score}_b$$

### Direct Sum / Addition ($\oplus$)
For error-damped evidence accumulation (geometric mean):
$$\text{val}_{res} = \text{val}_a + \text{val}_b$$
$$\text{score}_{res} = \sqrt{\text{score}_a \times \text{score}_b}$$

### Conservative Minimum
$$\text{score}_{res} = \min(\text{score}_a, \text{score}_b)$$

### Epistemic Contagion Rule
If either operand has state `UNKNOWN`, the resulting state is `UNKNOWN` and the resulting score is strictly `0.000000`.

---

## 4. Statements & Directives

### `assert`
Asserts that a probabilistic value meets or exceeds an exact confidence threshold:

```ydv
assert c ≽ 0.85 ·
```

If the condition fails ($s < 0.85$ or state is `UNKNOWN`), the program terminates safely with exit code 1.

### `infer`
Performs forward inference of a neural or cognitive model on an input tensor:

```ydv
infer model on input ↦ output ·
```

### `propagate`
Propagates confidence bounds across cognitive branches:

```ydv
propagate target from source ·
```

---

## 5. Zero-Libc Runtime Architecture

The CIR runtime is implemented in `internal/runtime/qadav_cir.c` and `internal/runtime/qadav_cir.h`:

- Memory layout:
  ```c
  typedef struct qadav_confidence {
      double  value;
      double  score;
      int8_t  epistemic;
      char    provenance[64];
  } qadav_confidence;
  ```
- Integer & floating-point conversions via custom freestanding routines.
- Integer Newton-Raphson approximation for square roots without `<math.h>`.
- Direct OS process termination via Win32 `ExitProcess` / Linux syscall `60`.

