# Qadav Language Formal Syntax Reference
# कादव भाषा औपचारिक व्याकरण संदर्भ

This document specifies the exact quantum-native grammar and vocabulary for Qadav M2.3 through M2.8.
यह दस्तावेज़ कादव M2.3 से M2.8 के संपूर्ण क्वांटम-नेटिव व्याकरण और शब्दावली को निर्दिष्ट करता है।

---

## 1. Network Declaration Grammar / नेटवर्क घोषणा व्याकरण

```ebnf
NetworkDecl ::= "network" Ident "⟦" LayerDecl* "⟧"
LayerDecl   ::= DenseLayer | ActLayer
DenseLayer  ::= "dense" "⟬" IntLit "⸴" IntLit "⟭" "·"
ActLayer    ::= ("relu" | "sigmoid" | "softmax" | "layernorm") "·"
```

### Constraints / प्रतिबंध:
- Every network must contain at least one layer.
- First layer must declare input dimension (e.g. `dense ⟬in⸴ out⟭`).
- Adjacent layers must possess matching output and input dimensions.

---

## 2. Learning & Training Grammar / प्रशिक्षण व्याकरण

```ebnf
LearnDecl   ::= "learn" Ident "on" Ident ("⸴" Ident)? "⟦" LearnField* "⟧"
LearnField  ::= LossField | OptField | LRField | EpochsField | BatchField
LossField   ::= "loss" ("mse" | "cross_entropy") "·"?
OptField    ::= "optimizer" ("sgd" | "adam") "·"?
LRField     ::= "lr" "↦"? Expr "·"?
EpochsField ::= "epochs" "↦"? Expr "·"?
BatchField  ::= "batch_size" "↦"? Expr "·"?
```

---

## 3. Probabilistic Types & Cognitive Intermediate Representation (CIR) / संभाव्य प्रकार एवं CIR

```ebnf
ProbType       ::= "confident" "⟨" Type "⟩" | "explore" "⟨" Type "⟩"
ConfidentLit   ::= Expr "∵" Expr
AssertStmt     ::= "assert" Expr "·"
InferStmt      ::= "infer" Expr "on" Expr ("with" Expr)? "↦" Ident "·"
PropagateStmt  ::= "propagate" Expr "into" Ident "·"
```

### Confidence Propagation Algebra:
- Tensor / Multiplication: `a ⊗ b ↦ a.score * b.score` (joint probability product).
- Addition: `a ⊕ b ↦ sqrt(a.score * b.score)` (independent noise geometric mean).
- Conservative Bound: `min(a.score, b.score)`.
- Assertion: `assert x ≽ threshold ·` validates `x.score >= threshold`. If failed, halts execution with code 1.

---

## 4. Quantum Computation Grammar / क्वांटम संगणना व्याकरण

```ebnf
QubitDecl      ::= "qubit" Ident ("⸴" Ident)* "·"
SuperposeStmt  ::= "superpose" Expr "·"
EntangleStmt   ::= "entangle" Expr "⸴"? Expr "·"
CollapseStmt   ::= "collapse" Expr ("↦" Expr)? "·"
CollapseExpr   ::= "collapse" Expr | "measure" Expr
GateSymbols    ::= "ℋ" | "𝒳" | "𝒴" | "𝒵" | "ℂ" | "𝕊" | "𝕋" | "𝕌" | "ℙ"
```

---

## 5. Prime-Symbolic Knowledge Grammar / प्राइम-प्रतीकात्मक व्याकरण

```ebnf
ConceptDecl    ::= "concept" Ident "↦" Expr "·"
ConceptLit     ::= "concept" IntLit
DecomposeExpr  ::= "decompose" "⦅" Expr "⦆"
ConceptComp    ::= Expr "⊗" Expr
```

---

## 6. Dataset Pipeline Grammar / डेटासेट पाइपलाइन व्याकरण

```ebnf
DatasetDecl    ::= "dataset" Ident "↦" "load" "⟬" StringLit "⟭" "·"
                 | "dataset" Ident "↦" "load" "⦅" StringLit "⦆" "·"
NormalizeStmt  ::= "normalize" Expr ("minmax" | "zscore") "·"
ShuffleStmt    ::= "shuffle" Expr "·"
BatchStmt      ::= "batch" Expr ("size" "↦")? Expr ("⟦" Stmt* "⟧" | "·")
```

---

## 7. Quasinary Computing Grammar / क्वासिनरी कम्प्यूटिंग व्याकरण

```ebnf
QuasiType      ::= "quasi"
QuasiLit       ::= "◄" | "〇" | "►"
EnergyDecl     ::= "energy" Ident "⟦" Stmt* "⟧"
MinimizeStmt   ::= "minimize" Expr "energy" Ident "iterations" "↦" Expr "·"
TileStmt       ::= "tile" Expr "·"
```

Saturated Ternary Arithmetic:
- `► ⊕ ► ↦ ►`
- `◄ ⊕ ◄ ↦ ◄`
- `► ⊕ ◄ ↦ 〇`
- `◄ ⊗ ◄ ↦ ►`

---

## 8. AGI Native Types Grammar / AGI नेटिव प्रकार व्याकरण

```ebnf
MemoryDecl     ::= "memory" Ident "⟬" IntLit "⟭" "·"
RememberStmt   ::= "remember" Ident "⟬" Expr "⟭" "↦" Expr "·"
RecallExpr     ::= "recall" Ident "⟬" Expr "⟭"

GraphDecl      ::= "graph" Ident "·"
NodeStmt       ::= "node" Ident "⟬" Expr "↦" Expr "⟭" "·"
LinkStmt       ::= "link" Ident "⟬" Expr "⸴" Expr "⸴" Expr "⟭" "·"
TraceExpr      ::= "trace" Ident "⟬" Expr "⸴" Expr "⟭"

ProcedureDecl  ::= "procedure" Ident "⦅" ParamList? "⦆" "→" Type "⟦" Stmt* "⟧"
ExecuteExpr    ::= "execute" Ident "⦅" ArgList? "⦆"
```

---

## 9. Mathematical Operator Semantics / ऑपरेटर अर्थमीमांसा

- `⊕` (`OSUM`): Always denotes **arithmetic addition** (`add i64`, `fadd double`).
- `⊻` (`XVEE`): Always denotes **bitwise and logical XOR** (`xor i64`, `xor i1`).

---

## 11. Security, Vault & Provenance Grammar / सुरक्षा, वॉल्ट एवं स्रोत व्याकरण

```ebnf
EthicalBoundaryDecl ::= "ethical_boundary" Ident "⟦" (EthicalRule)* "⟧"
EthicalRule         ::= "rule" Ident "↦" FloatLit "·"
EnforceStmt         ::= "enforce" Ident "on" Expr "·"
DecoherenceStmt     ::= "decoherence" "·"

VaultDecl           ::= "vault" Ident "with" Ident "⟬" Ident "⟭" "threshold" "↦" IntLit "·"
SealStmt            ::= "seal" Ident "·"
UnsealStmt          ::= "unseal" Ident "with" Ident "·"
WipeStmt            ::= "wipe" Ident "·"

ProvenanceDecl      ::= "provenance" Ident "⟬" Ident "⟭" "·"
CertifyStmt         ::= "certify" Ident "·"
VerifyTrainingExpr  ::= "verify_training" "⟬" Expr "⟭"

MemoryProvDecl      ::= "memory_prov" Ident "with" Ident "⟦" ("capacity" "↦" IntLit "·")? "⟧" "·"?
VerifyMemoryStmt    ::= "verify_memory" Ident "·"
AuditMemoryStmt     ::= "audit_memory" Ident "·"
CorruptShareStmt    ::= "corrupt_memory_share" Ident "⟬" Expr "⟭" "at" "↦" IntLit "·"

LicenseDecl         ::= "license" Ident "for" Ident "on" Ident "⟦" ("max_epochs" "↦" IntLit "·")? ("max_rate" "↦" FloatLit "·")? ("expiry" "↦" IntLit "·")? "⟧" "·"?
TransferStmt        ::= "transfer" Ident "from" Ident "to" Ident "layers" "⟬" (IntLit ("⸴" IntLit)*)? "⟭" "·"
DistillStmt         ::= "distill" Ident "on" Ident "with" Expr "⸴" Expr "⟦" ("rate" "↦" FloatLit "·")? ("epochs" "↦" IntLit "·")? "⟧" "·"?

ReasoningChainDecl  ::= "reasoning_chain" Ident ("with" Ident)? "⟦" ("capacity" "↦" IntLit "·")? "⟧" "·"?
ReasonStepStmt      ::= "reason" ("deduce" | "induce" | "abduce" | "verify") ("from" Expr)? "to" Expr "with" FloatLit "·"
VerifyChainStmt     ::= "verify_chain" Ident "·"
AuditReasoningStmt  ::= "audit_reasoning" Ident "·"
CorruptReasonStep   ::= "corrupt_reasoning_step" Ident "⟬" Expr "⟭" "at" "↦" IntLit "·"

ModalityDecl        ::= "modality" Ident "as" ("tensor_input" | "text_input" | "concept_input") "from" Expr "·"
CrossModalReason    ::= "cross_modal_reason" Ident ("deduce" | "induce" | "abduce" | "verify") "from" "⟬" (Expr ("⸴" Expr)*)? "⟭" "to" Expr "with" FloatLit "·"

QECMemoryDecl       ::= "qec_memory" Ident "as" ("surface_code" | "qldpc_code") "distance" "↦" IntLit ("with" Ident)? ("⟦" "⟧")? "·"
SyndromeStmt        ::= "syndrome" Ident ("into" Ident)? "·"
VerifySyndromeStmt  ::= "verify_syndrome" Ident "with" Ident "·"
CorrectStmt         ::= "correct" Ident "with" Ident "·"
CorruptSyndromeStmt ::= "corrupt_syndrome" Ident "⟬" Ident "⟭" "at" "↦" IntLit "·"

ConsolidateStmt     ::= "consolidate" Ident "into" Ident "with" "threshold" "↦" FloatLit "·"

RecallTokenDecl     ::= "recall_token" Ident "for" Ident ("on" | "to") Ident "⟦" ("keys" "↦" "⟬" (StringLit ("⸴" StringLit)*)? "⟭" "·")? ("expiry" "↦" IntLit "·")? "⟧" "·"?
GrantStmt           ::= "grant" Ident ("to" Ident)? "·"
RevokeStmt          ::= "revoke" Ident "·"
ScopedRecallExpr    ::= "recall_scoped" Ident "⟬" Expr "⟭" ("with" | "using") Ident

ToolRegisterDecl    ::= "tool_register" Ident ("with" Ident)? ("pre" "⟦" Expr "⟧")? ("post" "⟦" Expr "⟧")? "·"
ToolCallStmt        ::= "tool_call" Ident "⦅" (Expr ("⸴" Expr)*)? "⦆" ("audit" Ident)? "·"
ToolCallExpr        ::= "tool_call" Ident "⦅" (Expr ("⸴" Expr)*)? "⦆" ("audit" Ident)?

CertifiedReasonStmt ::= "certified_reason" Ident ("deduce" | "induce" | "abduce" | "verify") "from" Expr "to" Expr "witness" "⟬" Expr "⟭" "confidence" "↦" FloatLit "·"
AbstainThreshDecl   ::= "abstain_threshold" Ident "↦" FloatLit "mode" ("retry" | "escalate" | "refuse") ("max_retries" "↦" IntLit)? "·"
AbstainCheckStmt    ::= "abstain_check" Ident ("into" Ident)? "·"
AbstainCheckExpr    ::= "abstain_check" Ident

SandboxCreateStmt   ::= "sandbox_create" Ident "with" Ident "⟦" ("capacity" "↦" IntLit "·")? "⟧" "·"?
SandboxRunStmt      ::= "sandbox_run" Ident "⟦" Statement* "⟧" "into" Ident "·"
SandboxVerifyStmt   ::= "sandbox_verify" Ident ("into" Ident)? "·"
SandboxVerifyExpr   ::= "sandbox_verify" Ident
```

### 19. Calibrated Confidence Formulation
For Type-Tagged Proof-Carrying Reasoning (TTPCR) and Calibrated Abstention, cumulative chain confidence is evaluated using the variance-penalized geometric composite confidence:
$$C_{\text{chain}} = \left(\prod_{i=1}^n c_i\right)^{1/n} \times \left(1.0 - \frac{\sigma^2_c}{4}\right)$$
where $c_i \in [0.0, 1.0]$ are individual step confidences and $\sigma^2_c = \frac{1}{n} \sum_{i=1}^n (c_i - \bar{c})^2$ is step confidence variance.

### 20. Sandbox Grammar & Quantum Isolation
Quantum-Entangled Sandbox Isolation (QESI — Invention #16) provisions isolated quantum registers:
- `sandbox_create <id> with <boundary> ⟦ capacity ↦ <N> · ⟧`
- `sandbox_run <id> ⟦ <stmts> ⟧ into <status_var> ·`
- `sandbox_verify <id> into <attestation_var> ·` / `<attestation_var> ↦ sandbox_verify <id>`

Violations inside `sandbox_run` decohere only the isolated sandbox register and store a 32-byte SHA3 forensic attestation, returning `-2` to the host process without host termination.

### 21. GPU DirectCompute & AGKE Grammar
Phase 6 introduces native DirectCompute GPU operations and Invention #17 (AGKE):
```bnf
GPUMatmulStmt       ::= ("@gpu")? "gpu_matmul" Expr "⸴" Expr "into" Ident ("audit" Ident)? "·"
GPUSoftmaxStmt      ::= ("@gpu")? "gpu_softmax" Expr "into" Ident ("audit" Ident)? "·"
GPULayernormStmt    ::= ("@gpu")? "gpu_layernorm" Expr "into" Ident ("audit" Ident)? "·"
AGKEAttestStmt      ::= "agke_attest" Ident ("⦅" Expr "⸴" Expr "⸴" Expr "⦆")? "·"
AGKEVerifyStmt      ::= "agke_verify" Ident ("into" Ident)? "·"
AGKEVerifyExpr      ::= "agke_verify" Ident
CorruptGPUDigestStmt::= "corrupt_gpu_digest" Ident "at" "↦" IntLit "·"  // --test only
```

### 22. Multi-GPU Pipeline Parallelism Grammar (M6.3)
Multi-GPU pipeline parallelism grammar enables multi-accelerator stage partitioning:
```bnf
PipelineClaimStmt   ::= ("@multi_device")? "pipeline_claim" IntLit "stages" "micro_batch" IntLit "into" Ident "·"
PipelineRunStmt     ::= "pipeline_run" Ident "on" Expr "into" Expr "·"
```

### 23. Cross-Device Attested Gradient Aggregation (CDAGA) Grammar (M6.4)
Invention #18 provides cryptographic attestation for distributed gradient reduction:
```bnf
CDAGARecordStmt     ::= "cdaga_record" ("from")? "⟬" (Expr ("⸴" Expr)*)? "⟭" "into" Ident "·"
CDAGAAttestStmt     ::= "cdaga_attest" Ident "to" Ident "·"
CDAGAVerifyStmt     ::= "cdaga_verify" Ident ("into" Ident)? "·"
CorruptCDAGAStmt    ::= "corrupt_cdaga" Ident "at" "↦" Expr "·"  // --test only
```

### 24. Distributed Training Grammar (M6.5)
In-process distributed ring reduction and parameter server operations:
```bnf
AllReduceStmt       ::= "all_reduce" (("from" "⟬" (Expr ("⸴" Expr)*)? "⟭" "into" Ident) | Expr) ("into" Ident)? "·"
PserverInitStmt     ::= "pserver_init" "into" Ident "with" Expr ("workers")? "·"
PserverAggStmt      ::= "pserver_aggregate" Ident "·"
DistAnnotation      ::= "@distributed"
```

### 25. Byzantine-Attested Robust Aggregation (BARAT) Grammar (M6.6)
Invention #19 enables Byzantine-resilient trimmed-mean aggregation with transparency proofs:
```bnf
BaratSubmitStmt     ::= "barat_submit" ((Expr "into" Ident "with" Expr) | ("from" Expr "into" Ident)) "·"
BaratAggregateStmt  ::= "barat_aggregate" Ident "into" Ident ("cert" | "with") ("into")? Ident "·"
BaratVerifyStmt     ::= "barat_verify" Ident ("into" Ident)? "·"
CorruptBaratStmt    ::= "corrupt_barat" Ident "at" "↦" Expr "·"  // --test only
```

### 26. Quantum Hardware Abstraction Layer (QHAL) Grammar (M7.1)
OpenQASM 3.0 export and backend targeting:
```bnf
QASMExportStmt      ::= "qasm_export" "into" Ident ("with" "size" Expr)? "·"
QHALBackendAnnot    ::= "@qhal_backend" (Ident | StringLit)
```

### 27. Self-Evolution Attestation (SEA) Grammar (M7.2)
Invention #20 provides cryptographically-attested autonomous program mutation:
```bnf
EvolveStmt          ::= "evolve" "from" Expr "to" Expr "with" "reason" Expr "into" Ident "·"
EvolveVerifyStmt    ::= "evolve_verify" Ident ("into" Ident)? "·"
EvolveHistoryStmt   ::= "evolve_history" "to" Ident "·"
CorruptEvolveStmt   ::= "corrupt_evolve" Ident "at" "↦" Expr "·"  // --test only
```

### 28. Hardware Target Transpilation Grammar (M7.3)
QHAL quantum circuit transpilation to external framework representations:
```bnf
TranspileToStmt     ::= "transpile_to" (StringLit | Ident) "into" Ident ("with" "size" Expr)? ("options" "⟦" (Ident "·")* "⟧" "·")? "·"
```

### 29. Evolution Envelope Attestation (EEA) Grammar (M7.4)
Invention #21 establishes cryptographic bounded mutation envelopes for autonomous evolution:
```bnf
EvolvePolicyStmt    ::= "evolve_policy" Ident "into" Ident "with" "rate" Expr "budget" Expr "·"
EvolveBoundedStmt   ::= "evolve_bounded" Ident "from" Expr "to" Expr "with" "reason" Expr "max_flips" Expr "max_new_symbols" Expr "into" Ident Ident "·"
EEAVerifyStmt       ::= "eea_verify" Ident "into" Ident "·"
CorruptEEAStmt      ::= "corrupt_eea" Ident "at" "↦" Expr "·"  // --test only
```

### 30. Real QEC Decoder Grammar (M7.5)
Enables zero-libc greedy min-weight decoding and verification on surface code distances $d \in \{3, 5, 7\}$:
```bnf
QECDDecodeStmt  ::= "qecd_decode" "distance" "↦" Expr "from" Ident "into" Ident "·"
QECDVerifyStmt  ::= "qecd_verify" Ident "into" Ident "·"
QECDApplyStmt   ::= "qecd_apply" Ident "on" Ident "·"
```

### 31. Attested Adaptive Decoder (AAD) Grammar (M7.6 — Invention #22)
Invention #22 establishes cryptographic operating envelopes and meta-learning lineage:
```bnf
AADCertStmt         ::= "aad_cert" "with" "model" "↦" Expr "train" "↦" Expr "max_syndrome" "↦" Expr "max_corr_dist" "↦" Expr "into" Ident "·"
AADAttestStmt       ::= "aad_attest" Ident ("cert" "↦" Ident "decode" "↦" Ident)? "into" Ident "·"
MetaFitStmt         ::= "meta_fit" "into" Ident "cert" "↦" Ident "·"
MetaAdaptStmt       ::= "meta_adapt" Ident "with" Ident "into" Ident "·"
MetaVerifyStmt      ::= "meta_verify" Ident "into" Ident "·"
CorruptAADEnvStmt   ::= "corrupt_aad_env" Ident "at" "↦" Expr "·"  // --test only

## 32. Record Declaration & Initialization Grammar (M8.1)
RecordDeclStmt      ::= "record" Ident "⟦" (Ident "↦" Type "·")* "⟧"
RecordInitExpr      ::= "record_init" Ident "⟦" (Ident "↦" Expr "⸴"?)* "·" "⟧"
FieldAccessExpr     ::= Expr ("." | "∸") Ident

## 33. Variant Declaration & Initialization Grammar (M8.1)
VariantDeclStmt     ::= "variant" Ident "⟦" ("|" Ident "(" (Ident "↦" Type "⸴"?)* ")" "·")* "⟧"
VariantInitExpr     ::= "variant_init" Ident "⦅" Ident "↦" Expr "⦆"

## 34. Match Statement Grammar (M8.1)
MatchStmt           ::= "match" Expr "⟦" MatchArm* ("otherwise" "⟹" StmtList "·")? "⟧"
MatchArm            ::= "|" Ident "(" (Ident "⸴"?)* ")" "⟹" StmtList "·"

## 35. Recursive Self-Compilation Attestation (RSCA) Grammar (M8.2 — Invention #23)
RSCAAttestStmt      ::= "rsca_attest" "stage" "↦" Expr "from" "source" "↦" Expr "compiler" "↦" Expr "output" "↦" Expr "into" Ident "·"
RSCAVerifyStmt      ::= "rsca_verify" Ident "into" Ident "·"
CorruptRSCAStmt     ::= "corrupt_rsca" Ident "at" "↦" Expr "·"  // --test only

## 36. Self-Hosted Serialization Grammar (M8.3a)
TokenRecord         ::= TokenType "|" Literal "|" LineNumber "|" ColumnNumber
SExprProgram        ::= "(" "Program" SExprDecl* ")"
SExprProcess        ::= "(" "ProcessDecl" StringLit "(" SExprParam* ")" SExprType SExprBlock ")"
SExprBlock          ::= "(" "Block" SExprStmt* ")"

## 37. Self-Hosted Codegen Grammar (M8.4d–M8.4g)

The self-hosted codegen (`selfhost/codegen.ydv`) uses C runtime helpers exposed by `qadav_io.c`. These are callable from `.ydv` code but are not part of user-facing language grammar — they are internal codegen primitives.

```bnf
// File path extraction (Win32 GetCommandLineA)
qadav_get_input_path() → text

// Program kind detection (0=hello-family, 1=fib-family, 2=dot-family, 3=calc-family, -1=error)
qadav_detect_program_kind(path) → int64

// Hello-family: extract yield number
qadav_extract_yield_number(path) → int64

// Fib-family: extract 4 values (n0, when_threshold, rec_offset1, rec_offset2)
qadav_scan_fib_values(path) → int64
qadav_get_fib_value(idx) → int64

// Dot-family: extract tensors
qadav_scan_tensors(path) → int64
qadav_get_tensor_name(t_idx) → text
qadav_get_tensor_size(t_idx) → int64
qadav_get_tensor_int_val(t_idx, e_idx) → int64

// Calc-family: extract statements
qadav_scan_statements(path) → int64
qadav_stmt_kind(idx) → int64      // 0=var, 1=emit, 2=yield
qadav_stmt_name(idx) → text
qadav_stmt_value(idx) → int64
qadav_stmt_expr(idx) → text
```

Supported program kinds in M8.4g:
- Kind 0: hello-family (single string emit + single yield)
- Kind 1: fib-family (process with recursion + when)
- Kind 2: dot-family (2 tensors + mul + reduce)
- Kind 3: calc-family (2 var binds + 2 emits + 1 yield)
- Kind -1: unrecognized → clean error exit 1

*Deferred to M8.5:* Full general-purpose AST traversal supporting all 173 node types.
```

### Compiler Testing Flags:
- `--test`: Enables test-only AST statements (e.g. `corrupt_memory_share`, `corrupt_reasoning_step`, `corrupt_syndrome`) for cryptographic tamper resilience verification. In normal builds, test-only statements are rejected with a checker error.

### Strict Syntax Rules:


1. **NO Legacy Keywords**: `let`, `mut`, `var`, `print`, `if`, `else`, `fn`, `return`, `class`, `struct`, `enum`, `import`, `export`, `true`, `false`, `null`, `async`, `await` are completely prohibited.
2. **NO Legacy ASCII Operators**: `+ - * / % = == != < > <= >= && || ! & | ^ ~ @ -> =>` are forbidden.
3. **NO Legacy ASCII Punctuation**: `( ) { } [ ] ; , :` are banned in Qadav code.
4. **Quantum-Native Delimiters**:
   - Blocks: `⟦` and `⟧`
   - Groups / Argument Lists: `⦅` and `⦆`
   - Tensor / Dimension Brackets: `⟬` and `⟭`
   - Generic Type Parameters: `⟨` and `⟩`
   - Binding / Assignment: `↦` and `←`
   - Statement Terminator: `·`
   - List Separator: `⸴`
