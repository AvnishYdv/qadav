# CB-SRCPP Specification: Cross-Boundary Semantic Recall with Cryptographic Privacy Preservation
## Invention #14 (Patent Pending — Qadav Phase 5)

---

### 1. Abstract & Executive Summary

In multi-agent cognitive architectures and distributed autonomous systems, sovereign agents operate under strict ethical boundaries, security policies, and confidentiality rules. Previous privacy mechanisms faced an intractable trade-off: either an agent disclosed its entire confidential ethical policy and boundary constraints to an external verifier, or external agents were completely barred from retrieving shared semantic memories across boundary domains.

**Cross-Boundary Semantic Recall with Cryptographic Privacy Preservation (CB-SRCPP — Invention #14)** solves this fundamental dilemma. CB-SRCPP introduces a cryptographic scoped access token protocol that enables cross-agent episodic and semantic memory recall from Quantum-Entangled Memory Provenance (QEMP) vaults without revealing the memory owner's confidential ethical boundary rules. By decoupling authorization proof (via Merkle commitment of permitted keys and recipient binding) from boundary constraint evaluation, CB-SRCPP achieves provably zero boundary disclosure while guaranteeing instant active decoherence if access scope, token validity, recipient identity, or memory integrity is violated.

---

### 2. Architectural Principles

1. **Zero Boundary Disclosure:** The issuing agent binds its ethical boundary hash commitment (SHA3-256 over canonical rule names and penalty parameters) into the token without transmitting or exposing the internal Hamiltonian rule structure.
2. **Cryptographic Scoped Membership:** Permitted keys are structured into a Merkle leaf tree:
   $$\text{leaf}_i = \text{SHA3-256}(\text{key}_i \parallel i)$$
   $$\text{keys\_root} = \text{MerkleRoot}(\{\text{leaf}_0, \dots, \text{leaf}_{n-1}\})$$
3. **Recipient Identity Binding:** Tokens are bound to the recipient agent's cryptographic identifier or public key hash via constant-time comparison ($\mathcal{O}(1)$ timing-leak free).
4. **Temporal Expiration & Instant Revocation:** Tokens enforce monotonic expiry timestamps and atomic revocation flags.
5. **Active Decoherence on Tamper:** Any attempt to spoof recipient identity, query unauthorized keys, or present a revoked token immediately triggers active decoherence across quantum and episodic memory registers.

---

### 3. Syntax & Grammar

In Qadav source files (`.ydv`), CB-SRCPP introduces five sovereign keywords and constructs:

#### 3.1 Token Declaration
```ydv
recall_token dr_token for secure_vault on research_agent ⟦
    keys ↦ ⟬ "patient_record" ⟭ ·
    expiry ↦ 3600 ·
⟧
```

#### 3.2 Grant Statement
```ydv
grant dr_token to research_agent ·
```

#### 3.3 Revoke Statement
```ydv
revoke dr_token ·
```

#### 3.4 Scoped Recall Expression
```ydv
text accessed_record ↦ recall_scoped secure_vault⟬"patient_record"⟭ with dr_token ·
```

---

### 4. Zero-Libc C Runtime API

Freestanding implementation in `internal/runtime/qadav_scoped.h` and `internal/runtime/qadav_scoped.c`:

```c
typedef struct qadav_scoped_token {
    char               name[64];
    unsigned char      token_id[32];
    unsigned char      boundary_hash[32];
    unsigned char      keys_root[32];
    unsigned char      recipient_hash[32];
    unsigned long long issued_at;
    unsigned long long expiry;
    int                revoked;
    int                num_keys;
    char               allowed_keys[16][64];
} qadav_scoped_token;

qadav_scoped_token* qadav_scoped_issue(const char* name, void* owner_boundary,
                                       const unsigned char recipient_hash[32],
                                       const char* const* allowed_keys, int num_keys,
                                       unsigned long long expiry_seconds);

int qadav_scoped_grant(qadav_scoped_token* token, const unsigned char recipient_hash[32]);
int qadav_scoped_revoke(qadav_scoped_token* token);
int qadav_scoped_verify(const qadav_scoped_token* token,
                        const unsigned char recipient_hash[32],
                        const char* requested_key,
                        unsigned long long current_time);

int qadav_scoped_recall(qadav_memory_provenance* mp,
                        const qadav_scoped_token* token,
                        const unsigned char recipient_hash[32],
                        const char* requested_key,
                        unsigned long long current_time,
                        void** out_value, unsigned long long* out_len);

void qadav_scoped_free(qadav_scoped_token* token);
```

---

### 5. Mathematical Proof of Privacy Preservation

Let $\mathcal{H}_A$ be the ethical Hamiltonian of Owner Agent $A$, defined by private rules $R_A = \{(r_1, p_1), \dots, (r_m, p_m)\}$.
Under standard QEEH evaluation:
$$H_A = \text{SHA3-256}\left(\sum_{j=1}^m r_j \parallel p_j\right)$$

In CB-SRCPP, Agent $A$ constructs token $T$ granting Agent $B$ access to subset $K_{sub} \subset K_A$.
The token contains:
$$T = \langle \text{token\_id}, H_A, \text{keys\_root}, H_B, t_{exp}, \text{status} \rangle$$

1. **One-Way Boundary Security:** By the preimage resistance of SHA3-256, Agent $B$ possessing $T$ learns $0$ bits of information regarding the rules $r_j$ and penalties $p_j$ in $R_A$.
2. **Strict Soundness:** When Agent $B$ executes `recall_scoped mp⟬k⟭ with T`, the QEMP engine verifies:
   $$\text{SHA3-256}(T.H_A) = \text{mp.auth\_boundary\_id}$$
   and validates that $k \in K_{sub}$.
3. **No Side Channels:** All string and digest comparisons use constant-time `qadav_sha3_ct_compare`.

---

### 6. Verification & Test Evidence

CB-SRCPP is verified by:
- Unit & Crypto Suite: `TestScopedIssueAndVerify`, `TestScopedRecallPrivacy`, `TestScopedUnauthorizedKey`
- Static Type Checker: `internal/checker/scoped_checker_test.go`
- Valid Execution: `examples/scoped_demo.ydv` (Passes with output `clinical_data_anonymized`)
- Negative Compile-Time Gating: `examples/scoped_denied.ydv` (`key "billing_id" not in token's allowed_keys`)
- Negative Runtime Decoherence: `examples/scoped_revoked.ydv` (Exits with `0xDEADBEEF`)

---

### 7. Hardened Edge Cases & Privacy Guarantees (Fix F19)

As part of Milestone M5.9/M5.10 hardening, three critical edge cases were formally tested and verified in `internal/runtime/qec_error_test.go`:
1. **Temporal Expiration Enforcement (`TestScopedExpiredToken`)**:
   Tokens issued with finite lifetimes (e.g., `expiry=1s`) reject recall requests with return code `-3` once wall-clock time passes expiration. Expired tokens cannot be used to retrieve secret shares.
2. **Recipient Identity Mismatch (`TestScopedRecipientMismatch`)**:
   Tokens bound to agent recipient hash $H_{rec}$ strictly reject presentation by unauthorized recipients ($H_{actual} \neq H_{rec}$), returning `-2` via constant-time comparison without disclosing memory contents.
3. **Boundary Non-Disclosure (`TestScopedBoundaryNotDisclosed`)**:
   Verification proves that `token.boundary_hash` is strictly an authentic 32-byte cryptographic SHA3-256 commitment of the issuer's Hamiltonian. The token neither contains nor leaks the raw rule strings, penalty floats, or internal boundary structures.

