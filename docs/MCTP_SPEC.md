# Merkle-Certified Training Provenance (MCTP) Specification
# मर्कल-प्रमाणित प्रशिक्षण साक्ष्य विनिर्देश
## Invention #8 — Patent Pending

---

## 1. Overview & Purpose / सारांश एवं उद्देश्य

The **Merkle-Certified Training Provenance (MCTP)** specification establishes an immutable, post-quantum cryptographic certificate format and runtime verification system for AI model training pipelines in Qadav.

Modern machine learning systems suffer from training opacity, data poisoning risks, and unprovable model origin. MCTP solves these challenges by:
1. **Per-Epoch Merkle Auditing**: Hashing model weights, dataset partitions, loss metrics, and hyperparameters into structured Merkle tree leaves at every training epoch.
2. **Post-Quantum Root Authentication**: Signing the resulting Merkle tree root with a 1,088-byte Winternitz One-Time Signature (WOTS, $w=256$), providing 128-bit security against quantum Grover search.
3. **Zero-Libc Standalone Format (`.mctp`)**: Implementing a portable binary container that can be written, inspected, and verified independently of external libraries or runtimes.

---

## 2. Binary Certificate File Format (`.mctp`) / बाइनरी प्रमाणपत्र फ़ाइल प्रारूप

An `.mctp` certificate file has the following binary structure:

| Field | Offset | Size | Type | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Magic Header** | `0x00` | 4 bytes | ASCII | `MCTP` (`0x4D 0x43 0x54 0x50`) |
| **Format Version** | `0x04` | 2 bytes | uint16 | `0x0001` (Version 1.0) |
| **Epoch Count ($N$)** | `0x06` | 4 bytes | uint32 | Total training epochs recorded |
| **Epoch Leaves** | `0x0A` | $N \times 108$ bytes | Array | Serialized epoch audit leaves |
| **Merkle Root Digest**| Variable | 32 bytes | Byte array | FIPS 202 SHA3-256 root digest |
| **WOTS Signature** | Variable | 1,088 bytes | Byte array | 34 hash chains ($w=256$) |
| **WOTS Public Key** | Variable | 1,088 bytes | Byte array | 34 chain terminal digests |

Total certificate size for $N$ epochs:
$$\text{Size}(N) = 10 + (108 \times N) + 32 + 1088 + 1088 = 2218 + 108N \text{ bytes}$$

---

## 3. Merkle Leaf Structure (108 Bytes) / मर्कल लीफ संरचना

Each leaf records the exact state of training at epoch $e$:

```
+---------------+----------------------+----------------------+-----------+-------------------------+
| Epoch (4B)    | Weights Hash (32B)   | Dataset Hash (32B)   | Loss (8B) | Hyperparams Hash (32B)  |
+---------------+----------------------+----------------------+-----------+-------------------------+
0               4                      36                     68          76                        108 bytes
```

1. **Epoch Index** (4 bytes): 32-bit unsigned integer, little-endian.
2. **Weights Hash** (32 bytes): SHA3-256 digest of all trainable weights at completion of epoch.
3. **Dataset Hash** (32 bytes): SHA3-256 digest of training dataset partition consumed during epoch.
4. **Loss Metric** (8 bytes): IEEE 754 double-precision float64 recorded training loss.
5. **Hyperparams Hash** (32 bytes): SHA3-256 digest of learning rate, batch size, and optimizer configuration.

---

## 4. Cryptographic Tree Construction & Signing / क्रिप्टोग्राफ़िक ट्री निर्माण एवं हस्ताक्षर

### 4.1 Leaf Hashing
For leaf $i$ containing 108 contiguous bytes $B_i$:
$$H_{\text{leaf}, i} = \text{SHA3-256}(B_i)$$

### 4.2 Merkle Root Evaluation
Leaves are organized pairwise into a binary hash tree:
$$H_{\text{parent}} = \text{SHA3-256}(H_{\text{left}} \parallel H_{\text{right}})$$
If an odd number of nodes exists at any level, the last node is duplicated:
$$H_{\text{parent, odd}} = \text{SHA3-256}(H_{\text{last}} \parallel H_{\text{last}})$$

### 4.3 Post-Quantum Authentication
The Merkle root digest $R$ (32 bytes) is signed using Winternitz OTS with parameter $w=256$:
- $R$ is split into 32 byte values: $m_0, m_1, \dots, m_{31} \in [0, 255]$.
- Checksum $C = \sum_{i=0}^{31} (255 - m_i)$ is computed and encoded as two additional bytes $m_{32}, m_{33}$.
- For each chain $i \in \{0, \dots, 33\}$, private key seed $sk_i$ is hashed $m_i$ times:
  $$\sigma_i = \text{SHA3-256}^{m_i}(sk_i)$$
- The resulting 34 values form the 1,088-byte WOTS signature.

---

## 5. Verification Protocol / सत्यापन प्रोटोकॉल

A client or auditor verifies an `.mctp` certificate via the following deterministic pipeline:

```
[ Read .mctp File ]
       |
[ Parse Header: Magic == "MCTP", Version == 1 ]
       |
[ Hash Each 108B Leaf with SHA3-256 ]
       |
[ Reconstruct Merkle Tree -> Compute Root R' ]
       |
[ Verify R' == Stored Merkle Root R ]
       |
[ Verify WOTS Signature over R using Stored PK ]
       |
[ (Optional) Match Final Epoch Weights Hash with Live Model ]
       |
   [ PASS / FAIL ]
```

Verification fails (exit status 1) if:
- Magic header or version is invalid.
- File length does not match $10 + 108N + 2208$.
- Recomputed Merkle root $R'$ does not match stored root $R$.
- WOTS signature verification fails.
- Model weights differ from the final certified epoch hash.

---

## 6. Language Syntax Integration / भाषा सिंटैक्स एकीकरण

```qadav
network Classifier ⟦
    input ↦ 4 ·
    hidden ↦ 8 ·
    output ↦ 2 ·
⟧ ·

provenance train_cert ⟦
    target ↦ Classifier ·
    cert ↦ "classifier.mctp" ·
⟧ ·

learn Classifier ⟦
    epochs ↦ 10 ·
    rate ↦ 0.01 ·
⟧ ·

certify train_cert ·

# Verify provenance in code
checked ≔ verify_training ⦅ "classifier.mctp" ⦆ ·
```

---

## 7. Command Line Interface / कमांड लाइन इंटरफ़ेस

```bash
# Verify standalone provenance certificate
qyn verify-training classifier.mctp

# Verify provenance certificate against model binary
qyn verify-training classifier.mctp classifier.exe

# Build with automatic training provenance certificate emission
qyn build model_pipeline.ydv --emit-mctp
```

