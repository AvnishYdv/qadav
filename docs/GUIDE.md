# Qadav Neural Networks & Cognitive Architecture Guide
# कादव न्यूरल नेटवर्क एवं संज्ञानात्मक आर्किटेक्चर गाइड

Welcome to the definitive guide for designing, declaring, training, and predicting with neural networks and probabilistic types in Qadav.
कादव में न्यूरल नेटवर्क और संभाव्य प्रकारों (probabilistic types) के साथ काम करने की संपूर्ण गाइड में आपका स्वागत है।

---

## 1. Introduction / परिचय

Qadav introduces native, first-class neural network declarations and online training paradigms integrated directly into its quantum-native syntax. Unlike conventional programming languages that delegate tensor math and neural backpropagation to massive multi-gigabyte Python frameworks, Qadav compiles deep neural operations directly to LLVM IR backed by a zero-libc runtime.

कादव क्वांटम-नेटिव सिंटैक्स में न्यूरल नेटवर्क और ऑनलाइन प्रशिक्षण (online training) को फर्स्ट-क्लास नागरिक बनाता है। पारंपरिक भाषाओं की तरह पायथन और भारी-भरकम लाइब्रेरियों पर निर्भर रहने के बजाय, कादव सीधे LLVM IR में शून्य-लाइबसी (zero-libc) रनटाइम के साथ कंपाइल होता है।

---

## 2. Declaring a Network / नेटवर्क घोषणा

Networks are declared at the top level using the `network` keyword followed by the architecture definition in unicode blocks `⟦ ... ⟧`:

नेटवर्क को `network` कीवर्ड और यूनिकोड ब्लॉक `⟦ ... ⟧` के साथ शीर्ष स्तर पर घोषित किया जाता है:

```qadav
network MyModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
    sigmoid ·
⟧
```

### Supported Layers / समर्थित परतें:
1. `dense ⟬in_dim⸴ out_dim⟭ ·`: Fully-connected linear transformation with Xavier parameter initialization.
2. `relu ·`: Rectified Linear Unit activation function (\(f(x) = \max(0, x)\)).
3. `sigmoid ·`: Smooth logistic sigmoid activation (\(f(x) = \frac{1}{1 + e^{-x}}\)) evaluated via zero-libc Taylor/Newton math.
4. `softmax ·`: Multi-class probability distribution normalization with numerical stability shift.
5. `layernorm ·`: Mean and variance normalization layer across feature dimensions.

---

## 3. Training with the `learn` Block / `learn` ब्लॉक द्वारा प्रशिक्षण

Training is executed with the `learn` statement, either inside a process or at file scope:

प्रशिक्षण को प्रोसेस के भीतर या फ़ाइल स्तर पर `learn` स्टेटमेंट द्वारा निष्पादित किया जाता है:

```qadav
process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⸴ ⟬1.0⸴ 0.0⟭⸴ ⟬1.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⸴ ⟬0.0⟭⟭ ·

    learn MyModel on x ⸴ y ⟦
        loss mse ·
        optimizer adam ·
        lr ↦ 0.1 ·
        epochs ↦ 50 ·
    ⟧

    yield 0 ·
⟧
```




### Training Parameters / प्रशिक्षण पैरामीटर:
- **`loss`**: Loss objective function (`mse` for regression/classification, `cross_entropy` for probabilistic targets).
- **`optimizer`**: Optimization algorithm (`sgd` for stochastic gradient descent, `adam` for adaptive moment estimation with first and second momentum buffers).
- **`lr ↦ <rate>`**: Learning rate step multiplier.
- **`epochs ↦ <count>`**: Number of complete iterations over the dataset.

---

## 4. Probabilistic Types (CIR Foundation) / संभाव्य प्रकार (CIR नींव)

As part of Patent-Free Invention #4 (Cognitive Intermediate Representation), Qadav introduces probabilistic cognitive types:

पेटेंट-मुक्त आविष्कार #4 के अंतर्गत कादव संभाव्य संज्ञानात्मक प्रकार प्रदान करता है:

1. **`confident⟨T⟩`**: Represents a value of type `T` accompanied by a verified confidence threshold (\(\ge 0.95\)).
2. **`explore⟨T⟩`**: Represents an exploratory hypothesis or stochastic parameter subject to epistemic uncertainty.

Example / उदाहरण:
```qadav
process main⦅⦆ → int64 ⟦
    confident⟨float64⟩ alpha ↦ 0.98 ·
    explore⟨float64⟩ beta ↦ 0.45 ·
    emit alpha ·
    emit beta ·
    yield 0 ·
⟧
```




---

## 5. Quantum Computing Simulation / क्वांटम संगणना सिमुलेशन

Qadav features native, zero-libc state vector quantum simulation supporting up to 20 qubits:
कादव में 20 क्वबिट तक का नेटिव, शून्य-लाइबसी स्टेट वेक्टर क्वांटम सिमुलेशन अंतर्निर्मित है:

### Declaring and Superposing Qubits / क्वबिट घोषणा और सुपरपोजिशन:
```qadav
process main⦅⦆ → int64 ⟦
    qubit q0 ·
    superpose q0 ·
    int64 r ↦ collapse q0 ·
    emit r ·
    yield 0 ·
⟧
```




### Entanglement and Bell States / एंटैंगलमेंट और बेल स्टेट:
```qadav
process main⦅⦆ → int64 ⟦
    qubit q0 ⸴ q1 ·
    superpose q0 ·
    entangle q0 ⸴ q1 ·
    int64 a ↦ collapse q0 ·
    int64 b ↦ collapse q1 ·
    emit a ⊻ b ·
    yield 0 ·
⟧
```



The Bell state \(|\Phi^+\rangle = \frac{|00\rangle + |11\rangle}{\sqrt{2}}\) guarantees perfect correlation between measurements: \(a \veebar b = 0\) always (using the bitwise XOR operator `⊻`).

---

## 6. Prime-Symbolic Knowledge Engine / प्राइम-प्रतीकात्मक ज्ञान इंजन

Based on Patent-Free Invention #2, concepts are encoded as unique prime numbers and composed via algebraic multiplication:
पेटेंट-मुक्त आविष्कार #2 के आधार पर प्रत्येक संकल्पना को अभाज्य संख्या के रूप में एनकोड किया जाता है:

```qadav
process main⦅⦆ → int64 ⟦
    concept consciousness ↦ 2 ·
    concept genome ↦ 3 ·
    concept human_scale ↦ 5 ·
    concept complex ↦ consciousness ⊗ genome ⊗ human_scale ·
    emit complex ·
    yield 0 ·
⟧
```



Output: `concept⟨2 × 3 × 5 = 30⟩`

Decomposition via prime factorization:
```qadav
process main⦅⦆ → int64 ⟦
    concept complex ↦ 613 ·
    tensor⟨int64⟩ factors ↦ decompose⦅complex⦆ ·
    emit factors ·
    yield 0 ·
⟧
```



Output: `tensor⟨int64⟩⟬1⟭: ⟬613⟭` (613 is prime).

---

## 7. Dataset Pipeline (QDS Binary Format) / डेटासेट पाइपलाइन

The zero-libc dataset pipeline provides secure, high-speed binary dataset ingestion through the `dataset` primitive and the custom `.qds` format.

```qadav
process main⦅⦆ → int64 ⟦
    dataset d ↦ load ⟬ "examples/data/train.qds" ⟭ ·
    normalize d minmax ·
    shuffle d ·
    batch d 2 ⟦
        emit d ·
    ⟧
    yield 0 ·
⟧
```




Features:
- Standalone zero-libc I/O with path-traversal protection (`..` rejection).
- Hard defensive limits: $\le 1$ GB total payload, $\le 100$ tensors, $\le 500$ MB per tensor.
- MinMax and Z-Score normalization routines in native machine code.
- Fisher-Yates shuffle driven by CPU hardware cycle-counter (`rdtsc`) xorshift64 PRNG.

---

## 8. Quasinary Computing (Patent-Free Invention #3) / क्वासिनरी कम्प्यूटिंग

Quasinary computing introduces balanced ternary logic (`quasi`) with states:
- `◄` (-1)
- `〇` (0)
- `►` (+1)

Arithmetic is strictly saturated:
- `► ⊕ ► ↦ ►`
- `◄ ⊕ ◄ ↦ ◄`
- `► ⊕ ◄ ↦ 〇`
- `◄ ⊗ ◄ ↦ ►`

```qadav
process main⦅⦆ → int64 ⟦
    quasi a ↦ ► ·
    quasi b ↦ ◄ ·
    quasi sum ↦ a ⊕ b ·
    emit sum ·
    yield 0 ·
⟧
```



Output: `〇`

### 2D Topological Ising Energy Minimization:
```qadav
process main⦅⦆ → int64 ⟦
    tensor⟨quasi⟩ grid ↦ tensor⟬
        ⟬ ► ⸴ ◄ ⸴ ► ⟭ ⸴
        ⟬ ◄ ⸴ 〇 ⸴ ◄ ⟭ ⸴
        ⟬ ► ⸴ ◄ ⸴ ► ⟭
    ⟭ ·
    minimize grid energy ising iterations ↦ 1000 ·
    emit grid ·
    yield 0 ·
⟧
```




---

## 9. Operator Semantics: `⊕` vs `⊻` / ऑपरेटर अर्थमीमांसा

To guarantee formal mathematical clarity without dual semantics:
- `⊕` (`OSUM`): Always denotes **arithmetic addition** (`add i64`, `fadd double`).
- `⊻` (`XVEE`): Always denotes **bitwise and logical XOR** (`xor i64`, `xor i1`).

---

## 10. Cognitive Intermediate Representation (CIR) / संज्ञानात्मक मध्यवर्ती प्रतिनिधित्व

CIR (Patent-Free Invention #4) introduces runtime-aware confidence arithmetic, calibration, and epistemic state assertions:

### Calibrated Confidence Literal (`score ∵ value`):
```qadav
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.90 ∵ 10.0 ·
    confident⟨float64⟩ b ↦ 0.80 ∵ 20.0 ·
    confident⟨float64⟩ c ↦ a ⊗ b ·
    emit c ·
    assert c ≽ 0.70 ·
    yield 0 ·
⟧
```



Output:
```
conf⟨value=200.000000, score=0.720000, state=known⟩
```

### Cognitive Assertions (`assert`):
```qadav
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.50 ∵ 10.0 ·
    emit a ·
    assert a ≽ 0.95 ·
    yield 0 ·
⟧
```


If `a.score < 0.95`, the program prints a formal assertion failure and halts with exit code 1.

If `a.score < 0.95`, the program prints a formal assertion failure and halts with exit code 1.

---

## 11. AGI Native Types: Memory, Graph, Procedure / AGI नेटिव प्रकार

Qadav natively incorporates first-class episodic memory, semantic causal graphs, and procedural execution:

### Episodic Memory (`memory`, `remember`, `recall`):
Episodic key-value memory implemented with a zero-libc FNV-1a hash table and linked list chaining:
```qadav
process main ⦅ ⦆ → int64 ⟦
    memory mem ⟬ 100 ⟭ ·
    remember mem⟬"agent_state"⟭ ↦ 42 ·
    int64 recalled ↦ recall mem⟬"agent_state"⟭ ·
    emit recalled ·
    yield 0 ·
⟧
```




### Semantic Causal Graph (`graph`, `node`, `link`, `trace`):
Directed causal dependency graphs with zero-libc Breadth-First Search topological tracing returning `tensor⟨int64⟩`:
```qadav
process main ⦅ ⦆ → int64 ⟦
    graph g ·
    node g ⟬ "conceptA" ↦ 101 ⟭ ·
    node g ⟬ "conceptB" ↦ 102 ⟭ ·
    link g ⟬ "conceptA" ⸴ "conceptB" ⸴ 1.0 ⟭ ·
    tensor⟨int64⟩ visited ↦ trace g ⟬ "conceptA" ⸴ 2 ⟭ ·
    emit visited ·
    yield 0 ·
⟧
```




### First-Class Procedures (`procedure`, `execute`):
Reusable procedural units dispatched directly via the Qadav runtime:
```qadav
procedure add_two ⦅ int64 x ⦆ → int64 ⟦
    yield x ⊕ 2 ·
⟧

process main ⦅ ⦆ → int64 ⟦
    int64 res ↦ execute add_two ⦅ 5 ⦆ ·
    emit res ·
    yield 0 ·
⟧
```

---

## 12. Self-Validating Post-Quantum Binary Integrity (M3.3 + M3.4) / स्व-सत्यापित उत्तर-क्वांटम बाइनरी अखंडता

Qadav executables can be built with post-quantum integrity verification and active self-nullification:

### Building with `--selfcheck`:
```powershell
qyn build program.ydv --selfcheck
```
This produces `program.exe` alongside its cryptographic sidecar `program.qsc`.

### How it works:
1. **Pre-Main Verification (`qadav_start`)**:
   Before user code runs, the runtime computes a zero-libc SHA3-256 Merkle tree of the executable file in 4KB chunks and verifies it against the embedded 128-bit post-quantum Winternitz One-Time Signature (WOTS, w=256, 1,088 bytes).
2. **Untampered Execution**:
   If the Merkle root and Lamport/WOTS signature match, execution proceeds to `main` normally (exit code 0).
3. **Tamper Detection & Self-Nullification**:
   If any byte is modified or the sidecar signature fails, the runtime modifies the memory page permissions of `main` via `VirtualProtect` (using exact PE exception directory boundaries), completely zeroes out the code memory, and halts with exit code `0xDEADBEEF` (`-559038737`).

### CLI Commands:
- Verify binary integrity:
  ```powershell
  qyn selfcheck verify program.exe
  ```
- Strip self-check sidecar:
  ```powershell
  qyn selfcheck strip program.exe
  ```

---

## 13. Quantum-Entangled Ethical Hamiltonian (QEEH) (M3.5 + M3.6) / क्वांटम-उलझा हुआ नैतिक हैमिल्टनियन

Qadav introduces native language-level ethical boundaries governed by quantum mechanical principles:

```qadav
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 1.0 ·
    ⟧

    enforce safety on "help_user" ·
    emit "Action allowed" ·

    enforce safety on "harm_user" ·
    emit "Action blocked" ·

    yield 0 ·
⟧
```

### Key Principles:
1. **Ethical Energy Hamiltonian ($H$)**:
   Each rule defines an energy penalty. When an action is evaluated, $E = \sum \text{penalty}_i \times (1.0 - \text{condition}_i)$.
2. **Decoherence as Ethical Enforcement**:
   If energy threshold is breached ($E > 0.0$ on prohibited actions), `qadav_qeeh_trigger_decoherence` is invoked. This instantly zeroes out the quantum superposition state amplitudes across the entire register, preventing unauthorized cognitive or physical execution.
3. **Zero-Libc Native Runtime**:
   Fully integrated into Qadav's zero-libc C runtime (`qadav_qeeh.h` / `qadav_qeeh.c`).

---

## 14. Quantum-Entangled Model Weight Vault (QEMWV) (M4.1) / क्वांटम-उलझा हुआ मॉडल वेट वॉल्ट

Qadav introduces proprietary model weight protection using Shamir's secret sharing and QEEH-authorized reconstruction:

```qadav
network SecretModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
⟧

process main ⦅ ⦆ → int64 ⟦
    ethical_boundary auth ⟦
        rule user_authorized ↦ 1.0 ·
    ⟧

    vault v ⟬ SecretModel ⟭ threshold ↦ 3 ·

    // Seal weights into 5 Shamir shares
    seal v ·
    emit "Weights sealed in vault" ·

    // Unseal requires valid QEEH ethical boundary
    unseal v with auth ·
    emit "Weights unsealed for inference" ·

    // Emergency wipe zeroizes all shares in memory
    wipe v ·
    emit "Vault wiped" ·

    yield 0 ·
⟧
```

### Key Principles:
1. **Information-Theoretic Security**: Weights are divided into $n=5$ shares using a degree $t-1=2$ polynomial over $GF(256)$. Any $t-1=2$ shares reveal strictly zero information about the weights.
2. **QEEH Authorization**: Unsealing verifies that the ethical Hamiltonian boundary ground state is preserved ($E = 0.0$).
3. **Decoherence upon Tampering**: Any attempt to read raw memory shares directly triggers instant quantum decoherence and completely wipes the vault in memory.

---

## 15. Merkle-Certified Training Provenance (MCTP) (M4.2) / मर्कल-प्रमाणित प्रशिक्षण स्रोत

Qadav generates cryptographic proof of model training history without revealing proprietary training datasets or model weights:

```qadav
network ProvenanceModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
⟧

process main ⦅ ⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⸴ ⟬1.0⸴ 0.0⟭⸴ ⟬1.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⸴ ⟬0.0⟭⟭ ·

    provenance p ⟬ ProvenanceModel ⟭ ·

    learn ProvenanceModel on x ⸴ y ⟦
        loss mse ·
        optimizer adam ·
        lr ↦ 0.1 ·
        epochs ↦ 20 ·
    ⟧

    certify p ·
    emit "Training certified" ·

    yield 0 ·
⟧
```

### CLI Verification:
```powershell
qyn verify-training examples\mctp_demo.mctp
```
Output:
```
INTEGRITY OK: 20 epochs, final weights match Merkle root 4747a8c355613cdcdc453ab9ccf4487209340343ebe20ab048a4c8a517f7beca, WOTS signature valid
```


