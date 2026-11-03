# Qadav Self-Validating Binary Specification (.qsc)
# कादव स्व-सत्यापित बाइनरी विनिर्देश

This document specifies the format, cryptographic construction, and runtime verification protocol for Qadav's Self-Validating Binary architecture (Patent-Free Invention #5).

---

## 1. Overview / अवलोकन

Qadav executables compiled with the `--selfcheck` flag feature post-quantum cryptographic integrity validation. The integrity payload is stored either in an external sidecar file (`<binary>.qsc`) or embedded directly within the PE binary. At process initialization (in `qadav_start` before user `main`), the freestanding runtime independently hashes the executable image using a zero-libc FIPS 202 SHA3-256 Merkle tree, verifies the Merkle root against a Lamport one-time signature, and halts execution with memory zeroization if any byte has been tampered with.

---

## 2. File Format Specification (.qsc) / फ़ाइल प्रारूप विनिर्देश

The `.qsc` (Qadav Self-Check) container is a fixed-header binary structure of exactly 24,632 bytes:

| Field Offset | Field Name | Type | Size (Bytes) | Description |
|---|---|---|---|---|
| `0x0000` | `Magic` | `uint32` | 4 | Magic bytes: `0x51534331` (ASCII `"QSC1"`) |
| `0x0004` | `Version` | `uint16` | 2 | Format version: `0x0001` |
| `0x0006` | `Flags` | `uint16` | 2 | Control flags (`0x0000` = default standalone) |
| `0x0008` | `BinarySize` | `uint64` | 8 | Size of verified executable in bytes |
| `0x0010` | `ChunkSize` | `uint32` | 4 | Leaf chunk size in bytes (default `4096`) |
| `0x0014` | `ChunkCount` | `uint32` | 4 | Number of chunks: $\lceil \text{BinarySize} / \text{ChunkSize} \rceil$ |
| `0x0018` | `MerkleRoot` | `uint8[32]` | 32 | SHA3-256 root hash of chunk Merkle tree |
| `0x0038` | `PublicKey` | `uint8[16384]` | 16,384 | Lamport OTS public key ($256 \times 2 \times 32$ bytes) |
| `0x4038` | `Signature` | `uint8[8192]` | 8,192 | Lamport OTS signature ($256 \times 32$ bytes) |
| **Total** | | | **24,632** | Fixed payload length |

---

## 3. Cryptographic Primitives / क्रिप्टोग्राफिक घटक

### 3.1 SHA3-256 (NIST FIPS 202)
- Zero-libc implementation of Keccak-$f[1600]$ permutation ($24$ rounds).
- Capacity: 512 bits, Rate: 1088 bits (136 bytes), Output length: 256 bits (32 bytes).
- Domain separator: `0x06` suffix appended to message bytes.
- Standard NIST test vectors verified:
  - Empty string: `a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a`
  - `"abc"`: `3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532`

### 3.2 Merkle Tree
- The binary is divided into contiguous chunks of `ChunkSize` (4,096 bytes).
- Each chunk is hashed with SHA3-256 to form leaf nodes $L_0, L_1, \dots, L_{n-1}$.
- Non-leaf parent nodes are computed as:
  $$\text{Parent} = \text{SHA3-256}(\text{LeftChild} \mathbin{\Vert} \text{RightChild})$$
- Odd nodes at any level are duplicated to provide strict binary balance.
- The root of the tree represents the cryptographic digest of the entire file.

### 3.3 Winternitz One-Time Signature (WOTS, w=256)
- **Security:** 128-bit post-quantum security against Grover quantum search attacks.
- **Key Generation & Parameters:**
  - Parameter $w = 256$, message length = 32 bytes, checksum length = 2 bytes (total chains $L = 34$).
  - Private key: 34 random 32-byte seeds.
  - Public key: Each seed hashed 255 times via SHA3-256 ($34 \times 32 = 1,088$ bytes).
- **Signing:**
  - Each message byte $v_i$ dictates $v_i$ iterative SHA3-256 hashes of private seed $i$.
  - Checksum bytes are similarly hashed.
  - Signature size: $34 \times 32 = 1,088$ bytes (< 2KB compact post-quantum signature).
- **Verification:**
  - For each chain $i \in [0, 33]$, hash signature element $255 - v_i$ times and compare against public key element in constant time.
  - Any single mismatch invalidates the entire signature.

---

## 4. Runtime Integrity Flow & Self-Nullification

```
[qadav_start]
      │
      ▼
Check <target>.qsc exists and matches size/chunk
      │
      ├────► Failed / Missing ──► [qadav_selfcheck_nullify]
      │                                     │
      ▼                                     ▼
Compute Merkle Root of binary image    VirtualProtect(&main, pe_main_len, PAGE_EXECUTE_READWRITE)
      │                                     │
      ▼                                     ▼
Verify WOTS of Merkle Root             memset((void*)&main, 0, pe_main_len)
      │                                     │
      ├────► Mismatch ─────────────────────► ExitProcess(0xDEADBEEF)
      │
      ▼ Valid
Proceed to user main() [Exit 0 on success]
```

When tampering is detected:
1. Runtime queries PE Exception Directory (`.pdata`) to determine exact function bounds (`pe_main_len`), eliminating hardcoded assumptions.
2. `VirtualProtect` alters memory protection of the user code segment containing `main()` to `PAGE_EXECUTE_READWRITE`.
3. `memset` completely zeroes the instruction bytes of the entry procedure.
4. The process terminates immediately via `ExitProcess(0xDEADBEEF)` (`-559038737`), rendering memory forensics and execution hijacking impossible.


---

## 5. Tooling Support (`qyn`)

- **Build with sidecar:**
  ```powershell
  qyn build program.ydv --selfcheck
  ```
- **Verify binary against sidecar:**
  ```powershell
  qyn selfcheck verify program.exe
  ```
- **Strip sidecar:**
  ```powershell
  qyn selfcheck strip program.exe
  ```

