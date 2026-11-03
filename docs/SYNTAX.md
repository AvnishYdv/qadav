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

VaultDecl           ::= "vault" Ident "⟬" Ident "⟭" "threshold" "↦" IntLit "·"
SealStmt            ::= "seal" Ident "·"
UnsealStmt          ::= "unseal" Ident "with" Ident "·"
WipeStmt            ::= "wipe" Ident "·"

ProvenanceDecl      ::= "provenance" Ident "⟬" Ident "⟭" "·"
CertifyStmt         ::= "certify" Ident "·"
VerifyTrainingExpr  ::= "verify_training" "⟬" Expr "⟭"
```


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
