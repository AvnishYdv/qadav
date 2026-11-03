# Memory Consolidation (Episodic → Semantic) Specification
# मेमोरी समेकन (एपिसोडिक → सेमेंटिक) विनिर्देश

## 1. Architectural Overview / वास्तुकला अवलोकन

In human cognitive neuroscience, the consolidation process transitions transient, highly-detailed experiential records stored in the hippocampus (episodic memory) into durable, generalized topological knowledge representations distributed throughout the neocortex (semantic memory).

In Qadav, **Memory Consolidation** provides the formal language and runtime foundation for this cognitive transition. It bridges:
1. **Episodic Memory Provenance (`memory_prov`)**: Shamir-shared, timestamped, cryptographically-attested sequential interaction records.
2. **Semantic Knowledge Graphs (`graph`)**: Prime-symbolic associative knowledge networks supporting topological traversal, path tracing, and conceptual clustering.

The consolidation operation promotes episodic memories exceeding a defined epistemic confidence threshold into semantic graph nodes and associative links while preserving unbroken cryptographic provenance and ethical boundary enforcement.

---

## 2. Language Grammar & Type Semantics / व्याकरण एवं प्रकार अर्थमीमांसा

### Grammar:
```ebnf
ConsolidateStmt ::= "consolidate" Ident "into" Ident "with" "threshold" "↦" FloatLit "·"
```

### Type Checking Constraints:
1. Source expression must be of type `memory_prov` (or identifier referencing a declared `MemoryProvenance`).
2. Target expression must be of type `graph` (or identifier referencing a declared `Graph`).
3. Confidence threshold must be a floating-point scalar in the closed interval $[0.0, 1.0]$. Values outside this range are rejected at compile time.
4. If the source `memory_prov` is protected by an `ethical_boundary`, the consolidation routine verifies that the active process environment possesses valid authorization matching the boundary's content hash.

---

## 3. Consolidation Algorithmic Workflow / समेकन कलन विधि कार्यप्रवाह

The consolidation engine executes within the zero-libc runtime (`qadav_consolidate.c`):

```
       [ Episodic Memory Store (memory_prov) ]
                         │
      (1) Boundary Authorization Verification
                         ▼
      (2) Filter: confidence >= threshold && !is_locked
                         ▼
      (3) QEMP Cryptographic Recall (k-of-n Shamir Reconstruction)
                         ▼
      (4) Deterministic Prime-Symbolic Coordinate Projection
                         ▼
      (5) Graph Node Insertion: qadav_graph_node(g, key, prime_coord)
                         ▼
      (6) Associative Edge Creation: qadav_graph_link(g, prev_key, key, 1.0)
                         ▼
       [ Durable Semantic Knowledge Graph (graph) ]
```

### Step 3.1: Boundary Authorization
Before reading any memory entries, the consolidation routine computes the content hash of the executing boundary and compares it in constant time with `mp->auth_boundary_id`. If boundary violation is detected:
- QEEH decoherence is triggered (`qadav_qeeh_trigger_decoherence`).
- The process terminates immediately via `ExitProcess(0xDEADBEEF)`.

### Step 3.2: Recall & Prime Coordinate Projection
For each unlocked episodic entry:
1. Entry value bytes are reconstructed via $k=3$ Shamir polynomial interpolation over $GF(256)$.
2. Value bytes are deterministically mapped to a prime-symbolic coordinate using modular polynomial hashing:
   $$\text{coord} = \Big(\sum_{j=0}^{L-1} v_j \cdot 31^{L-1-j} + 101\Big) \pmod{7919}$$
   (ensuring coordinates map into the precomputed prime table space).
3. The entry key and computed prime coordinate are inserted into the target knowledge graph via `qadav_graph_node`.
4. Contiguous episodic transitions generate associative semantic links (`qadav_graph_link`) weighted by transition confidence.

---

## 4. Security & Data Integrity Invariants / सुरक्षा एवं डेटा अखंडता अपरिवर्तनीय नियम

1. **Zero Data Leakage**: Reconstructed cleartext buffers during Shamir recall are zeroized immediately after prime coordinate extraction.
2. **Locked Entry Immunity**: Entries marked as locked (`is_locked != 0`) are skipped, protecting sensitive or unverified episodic states from premature semantic promotion.
3. **Deterministic Graph Topology**: Identical episodic memory logs consolidated under the same threshold produce topologically isomorphic semantic knowledge graphs across all Windows runs.
4. **Zero-Libc Standalone Execution**: All Shamir reconstruction, string hashing, and graph manipulations operate exclusively via Win32 OS primitives.

