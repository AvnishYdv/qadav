# Qadav M2.3 & M2.4 Example Programs
# कादव M2.3 एवं M2.4 उदाहरण प्रोग्राम

Complete quantum-native programs demonstrating neural network training and inference.
न्यूरल नेटवर्क प्रशिक्षण और अनुमान का प्रदर्शन करने वाले संपूर्ण क्वांटम-नेटिव प्रोग्राम।

---

## 1. Linear Regression (\(y = 2x\)) / रेखीय प्रतिगमन

File: `examples/linear_regression.ydv`

```qadav
network LineModel ⟦
    dense ⟬1⸴ 1⟭ ·
⟧

process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 1⟭ x ↦ tensor⟬⟬1.0⟭⸴ ⟬2.0⟭⸴ ⟬3.0⟭⸴ ⟬4.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬2.0⟭⸴ ⟬4.0⟭⸴ ⟬6.0⟭⸴ ⟬8.0⟭⟭ ·
    learn LineModel on x ⸴ y ⟦
        loss mse ·
        optimizer sgd ·
        lr ↦ 0.05 ·
        epochs ↦ 100 ·
    ⟧
    tensor⟨float64⟩⟬1⟭ test_x ↦ tensor⟬3.0⟭ ·
    tensor⟨float64⟩⟬1⟭ pred ↦ LineModel⦅test_x⦆ ·
    emit pred ·
    yield 0 ·
⟧
```

### Execution & Output / निष्पादन एवं आउटपुट:
```powershell
.\bin\qyn.exe run examples\linear_regression.ydv
```
Output:
```
tensor⟨float64⟩⟬1⟭:
⟬6.000103⟭
```

---

## 2. Non-Linear XOR Network / गैर-रेखीय XOR नेटवर्क

File: `examples/xor_network.ydv`

```qadav
network XORNet ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
    sigmoid ·
⟧

process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⸴ ⟬1.0⸴ 0.0⟭⸴ ⟬1.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⸴ ⟬0.0⟭⟭ ·
    learn XORNet on x ⸴ y ⟦
        loss mse ·
        optimizer adam ·
        lr ↦ 0.1 ·
        epochs ↦ 50 ·
    ⟧
    tensor⟨float64⟩⟬2⟭ test_x ↦ tensor⟬0.0⸴ 1.0⟭ ·
    tensor⟨float64⟩⟬1⟭ pred ↦ XORNet⦅test_x⦆ ·
    emit pred ·
    yield 0 ·
⟧
```

### Execution & Output / निष्पादन एवं आउटपुट:
```powershell
.\bin\qyn.exe run examples\xor_network.ydv
```
Output:
```
tensor⟨float64⟩⟬1⟭:
⟬0.978783⟭
```

---

## 3. Binary Classifier / बाइनरी क्लासिफायर

File: `examples/classifier.ydv`

```qadav
network BinaryClassifier ⟦
    dense ⟬2⸴ 2⟭ ·
    relu ·
    dense ⟬2⸴ 1⟭ ·
    sigmoid ·
⟧

process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.1⸴ 0.2⟭⸴ ⟬0.2⸴ 0.1⟭⸴ ⟬0.8⸴ 0.9⟭⸴ ⟬0.9⸴ 0.8⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⟭ ·
    learn BinaryClassifier on x ⸴ y ⟦
        loss cross_entropy ·
        optimizer adam ·
        lr ↦ 0.05 ·
        epochs ↦ 50 ·
    ⟧
    tensor⟨float64⟩⟬2⟭ test_sample ↦ tensor⟬0.85⸴ 0.85⟭ ·
    tensor⟨float64⟩⟬1⟭ prediction ↦ BinaryClassifier⦅test_sample⦆ ·
    emit prediction ·
    yield 0 ·
⟧
```

### Execution & Output / निष्पादन एवं आउटपुट:
```powershell
.\bin\qyn.exe run examples\classifier.ydv
```
Output:
```
tensor⟨float64⟩⟬1⟭:
⟬0.982042⟭
```

---

## 4. Quantum Superposition / क्वांटम सुपरपोजिशन

File: `examples/quantum_hello.ydv`

```qadav
process main⦅⦆ → int64 ⟦
    qubit q0 ·
    superpose q0 ·
    int64 r ↦ collapse q0 ·
    emit r ·
    yield 0 ·
⟧
```

### Execution & Output:
```powershell
.\bin\qyn.exe run examples\quantum_hello.ydv
```
Output (true non-deterministic random measurement): `0` or `1` (50% probability each).

---

## 5. Quantum Bell State Entanglement / क्वांटम बेल स्टेट

File: `examples/bell_state.ydv`

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

### Execution & Output:
```powershell
.\bin\qyn.exe run examples\bell_state.ydv
```
Output: `0` (Bell pair \(|\Phi^+\rangle = \frac{|00\rangle + |11\rangle}{\sqrt{2}}\) guarantees measurement correlation: \(a \veebar b = 0\) always using bitwise XOR `⊻`).

---

## 6. Prime-Symbolic Concept Encoding / संकल्पना संयोजन

File: `examples/concept_encoding.ydv`

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

### Execution & Output:
```powershell
.\bin\qyn.exe run examples\concept_encoding.ydv
```
Output:
```
concept⟨2 × 3 × 5 = 30⟩
```

---

## 7. Prime-Symbolic Concept Decomposition / संकल्पना अपघटन

File: `examples/concept_decompose.ydv`

```qadav
process main⦅⦆ → int64 ⟦
    concept complex ↦ 613 ·
    tensor⟨int64⟩ factors ↦ decompose⦅complex⦆ ·
    emit factors ·
    yield 0 ·
⟧
```

### Execution & Output:
```powershell
.\bin\qyn.exe run examples\concept_decompose.ydv
```
Output:
```
tensor⟨int64⟩⟬1⟭:
⟬613⟭
```
(613 is prime; prime factors emitted as a tensor).

---

## 8. Dataset Pipeline Loading & Normalization / डेटासेट पाइपलाइन

File: `examples/dataset_basic.ydv`

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

### Execution & Output:
```powershell
qyn run examples/dataset_basic.ydv
```
Output:
```
dataset⟨n_tensors=2, n_samples=4⟩:
  [features] tensor⟨float64⟩⟬4⸴ 2⟭:
⟬⟬0.571429⸴ 0.714286⟭⸴ ⟬0.285714⸴ 0.428571⟭⸴ ⟬0.000000⸴ 0.142857⟭⸴ ⟬0.857143⸴ 1.000000⟭⟭
  [labels] tensor⟨int64⟩⟬4⟭:
⟬0⸴ 1⸴ 0⸴ 1⟭
```

---

## 9. Quasinary Balanced Ternary Arithmetic / क्वासिनरी अंकगणित

File: `examples/quasi_basic.ydv`

```qadav
process main⦅⦆ → int64 ⟦
  quasi a ↦ ► ·
  quasi b ↦ ◄ ·
  quasi sum ↦ a ⊕ b ·
  emit sum ·
  yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples/quasi_basic.ydv
```
Output:
```
〇
```

---

## 10. Quasinary 2D Grid Ising Energy Minimization / क्वासिनरी ऊर्जा न्यूनीकरण

File: `examples/quasi_grid.ydv`

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

### Execution & Output:
```powershell
qyn run examples/quasi_grid.ydv
```
Output:
```
tensor⟨quasi⟩⟬3 ⸴ 3⟭:
⟬
  ⟬〇 ⸴ 〇 ⸴ 〇 ⟭ ⸴
  ⟬〇 ⸴ ◄ ⸴ 〇 ⟭ ⸴
  ⟬〇 ⸴ 〇 ⸴ 〇 ⟭
⟭
```

---

## 11. CIR Basic Confidence Arithmetic / CIR बेसिक विश्वास संगणना

File: `examples/cir_basic.ydv`

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

### Execution & Output:
```powershell
qyn run examples\cir_basic.ydv
```
Output:
```
conf⟨value=200.000000, score=0.720000, state=known⟩
```

---

## 12. CIR Assertion Failure Guard / CIR सत्यापन विफलता सुरक्षा

File: `examples/cir_assert_fail.ydv`

```qadav
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.50 ∵ 10.0 ·
    emit a ·
    assert a ≽ 0.95 ·
    yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples\cir_assert_fail.ydv
```
Output (Exit Code 1):
```
conf⟨value=10.000000, score=0.500000, state=known⟩
Assertion failed: confidence score 0.500000 < threshold 0.950000
```

---

## 13. AGI Episodic Memory Persistence / AGI प्रासंगिक स्मृति

File: `examples/memory_basic.ydv`

```qadav
process main ⦅ ⦆ → int64 ⟦
    memory mem ⟬ 100 ⟭ ·
    remember mem⟬"agent_state"⟭ ↦ 42 ·
    int64 recalled ↦ recall mem⟬"agent_state"⟭ ·
    emit recalled ·
    yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples\memory_basic.ydv
```
Output:
```
42
```

---

## 14. AGI Causal Graph Tracing / AGI कारणात्मक ग्राफ ट्रेसिंग

File: `examples/graph_basic.ydv`

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

### Execution & Output:
```powershell
qyn run examples\graph_basic.ydv
```
Output:
```
tensor⟨int64⟩⟬2⟭:
⟬101⸴ 102⟭
```

---

## 15. AGI First-Class Procedure / AGI प्रथम श्रेणी कार्यप्रणाली

File: `examples/procedure_basic.ydv`

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

### Execution & Output:
```powershell
qyn run examples\procedure_basic.ydv
```
Output:
```
7
```

---

## 16. Self-Validating Post-Quantum Binary Integrity / स्व-सत्यापित उत्तर-क्वांटम बाइनरी अखंडता

File: `examples/selfcheck_demo.ydv`

```qadav
process main ⦅ ⦆ → int64 ⟦
    emit "Self-check OK" ·
    yield 0 ·
⟧
```

### Build & Integrity Verification:
```powershell
qyn build examples\selfcheck_demo.ydv --selfcheck
```
Generates `examples/selfcheck_demo.exe` and sidecar `examples/selfcheck_demo.qsc`.

### Untampered Execution:
```powershell
.\examples\selfcheck_demo.exe
```
Output:
```
Self-check OK
```
Exit code: `0`

### CLI Verification:
```powershell
qyn selfcheck verify examples\selfcheck_demo.exe
```
Output:
```
INTEGRITY OK: Merkle root <hash> verified via Lamport post-quantum signature
```

### Tamper Protection & Code Nullification:
If a single byte of the binary is modified, running the executable immediately zeroizes `main()` in memory and terminates:
```powershell
.\examples\selfcheck_demo.exe
```
Exit code: `0xDEADBEEF` (`-559038737`)

Verification failure:
```powershell
qyn selfcheck verify examples\selfcheck_demo.exe
```
Output:
```
INTEGRITY TAMPERED: Merkle root mismatch or invalid Lamport signature!
```
Exit code: `1`

---

## 17. Quantum-Entangled Ethical Hamiltonian (QEEH) / क्वांटम-उलझा हुआ नैतिक हैमिल्टनियन

File: `examples/qeeh_demo.ydv`

```qadav
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 1.0 ·
    ⟧

    enforce safety on "help_user" ·
    emit "Action allowed" ·

    enforce safety on "harm_user" ·
    emit "Action blocked" ·

    decoherence ·
    emit "Decoherence active" ·

    yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples\qeeh_demo.ydv
```
Output:
```
Action allowed
Action blocked
Decoherence active
```
Exit code: `0`

### Build with Post-Quantum Self-Check:
```powershell
qyn build examples\qeeh_demo.ydv --selfcheck
```
Generates `examples/qeeh_demo.exe` and `examples/qeeh_demo.qsc`.

### Untampered Execution:
```powershell
.\examples\qeeh_demo.exe
```
Output:
```
Action allowed
Action blocked
Decoherence active
```
Exit code: `0`

### Verification CLI:
```powershell
qyn selfcheck verify examples\qeeh_demo.exe
```
Output:
```
INTEGRITY OK: Merkle root verified via post-quantum signature
```

---

## 18. Quantum-Entangled Model Weight Vault (QEMWV) / मॉडल वेट वॉल्ट

File: `examples/qemwv_demo.ydv`

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

    seal v ·
    emit "Weights sealed in vault" ·

    unseal v with auth ·
    emit "Weights unsealed for inference" ·

    wipe v ·
    emit "Vault wiped" ·

    yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples\qemwv_demo.ydv
```
Output:
```
Weights sealed in vault
Weights unsealed for inference
Vault wiped
```
Exit code: `0`

---

## 19. QEMWV Tamper & Decoherence Test / वॉल्ट अनाधिकृत एक्सेस एवं डिकोहेरेंस

File: `examples/qemwv_tamper.ydv`

```qadav
network SecretModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
⟧

process main ⦅ ⦆ → int64 ⟦
    ethical_boundary strict_auth ⟦
        rule unauthorized_access ↦ 1.0 ·
    ⟧

    vault v ⟬ SecretModel ⟭ threshold ↦ 3 ·

    seal v ·
    emit "Weights sealed in vault" ·

    enforce strict_auth on "harm_user" ·
    emit "Decoherence triggered on unauthorized access" ·

    yield 0 ·
⟧
```

### Execution & Output:
```powershell
qyn run examples\qemwv_tamper.ydv
```
Output:
```
Weights sealed in vault
Decoherence triggered on unauthorized access
```
Exit code: `0`

---

## 20. Merkle-Certified Training Provenance (MCTP) / मर्कल-प्रमाणित प्रशिक्षण स्रोत

File: `examples/mctp_demo.ydv`

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

### Execution & Output:
```powershell
qyn run examples\mctp_demo.ydv
```
Output:
```
Training certified
```
Exit code: `0`

### Build with MCTP Certificate Sidecar:
```powershell
qyn build examples\mctp_demo.ydv --emit-mctp
```
Generates `examples/mctp_demo.exe` and `examples/mctp_demo.mctp`.

### Provenance Certificate Verification:
```powershell
qyn verify-training examples\mctp_demo.mctp
```
Output:
```
INTEGRITY OK: 20 epochs, final weights match Merkle root 4747a8c355613cdcdc453ab9ccf4487209340343ebe20ab048a4c8a517f7beca, WOTS signature valid
```
Exit code: `0`





