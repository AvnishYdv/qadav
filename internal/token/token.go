package token

import "fmt"

// Type represents the type of a lexical token.
type Type string

// Special tokens
const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	IDENT   Type = "IDENT"
	INT     Type = "INT"
	FLOAT   Type = "FLOAT"
	STRING  Type = "STRING"
)

// Arithmetic Operators (Quantum Algebra)
const (
	OSUM   Type = "OSUM"   // ⊕ \oplus
	ODIFF  Type = "ODIFF"  // ⊖ \ominus
	OTIM   Type = "OTIM"   // ⊗ \otimes
	ODIV   Type = "ODIV"   // ⊘ \oslash
	OMOD   Type = "OMOD"   // ⊙ \odot
	OPOW   Type = "OPOW"   // ↑ \uparrow
	MATMUL Type = "MATMUL" // ⊛ \matmul
)

// Comparison Operators
const (
	EQUIV     Type = "EQUIV"     // ≡ \equiv
	NEQ       Type = "NEQ"       // ≠ \neq
	PREC      Type = "PREC"      // ≺ \prec
	SUCC      Type = "SUCC"      // ≻ \succ
	PREQ      Type = "PREQ"      // ≼ \preceq
	SUEQ      Type = "SUEQ"      // ≽ \succeq
	APPROX    Type = "APPROX"    // ≈ \approx
	MEMBER    Type = "MEMBER"    // ∈ \in
	NOTMEMBER Type = "NOTMEMBER" // ∉ \notin
)

// Logical Operators
const (
	WEDGE Type = "WEDGE" // ∧ \wedge
	VEE   Type = "VEE"   // ∨ \vee
	NEG   Type = "NEG"   // ¬ \neg
	XVEE  Type = "XVEE"  // ⊻ \xor
)

// Binding & Flow
const (
	BIND     Type = "BIND"     // ↦ \bind
	ASSIGN   Type = "ASSIGN"   // ← \assign
	YIELDS   Type = "YIELDS"   // → \yields
	IMPLIES  Type = "IMPLIES"  // ⇒ \implies
	PIPELINE Type = "PIPELINE" // ↝ \pipe
)

// Compound Assignment
const (
	BIND_ADD Type = "BIND_ADD" // ↦⊕ \bindplus
	BIND_SUB Type = "BIND_SUB" // ↦⊖ \bindminus
	BIND_MUL Type = "BIND_MUL" // ↦⊗ \bindtim
	BIND_DIV Type = "BIND_DIV" // ↦⊘ \binddiv
)

// Structural Punctuation
const (
	GROUP_OPEN    Type = "GROUP_OPEN"    // ⦅ \group
	GROUP_CLOSE   Type = "GROUP_CLOSE"   // ⦆ \egroup
	BLOCK_OPEN    Type = "BLOCK_OPEN"    // ⟦ \block
	BLOCK_CLOSE   Type = "BLOCK_CLOSE"   // ⟧ \eblock
	INDEX_OPEN    Type = "INDEX_OPEN"    // ⟬ \bracket
	INDEX_CLOSE   Type = "INDEX_CLOSE"   // ⟭ \ebracket
	STMT_END      Type = "STMT_END"      // · \stmt
	SEP           Type = "SEP"           // ⸴ \sep
	ANNOT         Type = "ANNOT"         // ∵ \because
	MEMBER_ACCESS Type = "MEMBER_ACCESS" // ∸ \dot
	RANGE         Type = "RANGE"         // … \range
	LANGLE        Type = "LANGLE"        // ⟨ \langle
	RANGLE        Type = "RANGLE"        // ⟩ \rangle
)

// Quantum-Native Symbols
const (
	HADAMARD    Type = "HADAMARD"    // ℋ \had
	PAULI_X     Type = "PAULI_X"     // 𝒳 \pauliX
	PAULI_Y     Type = "PAULI_Y"     // 𝒴 \pauliY
	PAULI_Z     Type = "PAULI_Z"     // 𝒵 \pauliZ
	CNOT        Type = "CNOT"        // ℂ \cnot
	SGATE       Type = "SGATE"       // 𝕊 \sgate
	TGATE       Type = "TGATE"       // 𝕋 \tgate
	UNITARY     Type = "UNITARY"     // 𝕌 \unitary
	PROBABILITY Type = "PROBABILITY" // ℙ \prob
	ADJOINT     Type = "ADJOINT"     // ⋆ \adjoint
	SUM         Type = "SUM"         // ∑ \sum
	PRODUCT     Type = "PRODUCT"     // ∏ \prod
	GRADIENT    Type = "GRADIENT"    // ∂ \grad
	DIVERGENCE  Type = "DIVERGENCE"  // ∇ \div
	SQRT        Type = "SQRT"        // √ \sqrt
	INFINITY    Type = "INFINITY"    // ∞ \inf
	EMPTY       Type = "EMPTY"       // ∅ \empty
)

// Mathematical Operators
const (
	FORALL    Type = "FORALL"    // ∀ \forall
	EXISTS    Type = "EXISTS"    // ∃ \exists
	UNION     Type = "UNION"     // ∪ \union
	INTERSECT Type = "INTERSECT" // ∩ \inter
	SUBSET    Type = "SUBSET"    // ⊂ \subset
	SUBSET_EQ Type = "SUBSET_EQ" // ⊆ \subseteq
	TOP       Type = "TOP"       // ⊤ \top
	BOTTOM    Type = "BOTTOM"    // ⊥ \bottom
)

// Primitive Numeric & Core Type Keywords
const (
	INT8       Type = "INT8"
	INT16      Type = "INT16"
	INT32      Type = "INT32"
	INT64      Type = "INT64"
	UINT8      Type = "UINT8"
	UINT16     Type = "UINT16"
	UINT32     Type = "UINT32"
	UINT64     Type = "UINT64"
	FLOAT32    Type = "FLOAT32"
	FLOAT64    Type = "FLOAT64"
	COMPLEX64  Type = "COMPLEX64"
	COMPLEX128 Type = "COMPLEX128"
	LOGIC      Type = "LOGIC"
	TEXT       Type = "TEXT"
	TENSOR     Type = "TENSOR"
	QUBIT      Type = "QUBIT"
	NEURON     Type = "NEURON"
)

// Process & Flow Keywords
const (
	PROCESS    Type = "PROCESS"
	YIELD_KW   Type = "YIELD_KW"
	FLUX       Type = "FLUX"
	CONCURRENT Type = "CONCURRENT"
	CONVERGE   Type = "CONVERGE"
)

// Control Keywords
const (
	WHEN          Type = "WHEN"
	OTHERWISE     Type = "OTHERWISE"
	ALTERNATIVELY Type = "ALTERNATIVELY"
	ITERATE       Type = "ITERATE"
	CYCLE         Type = "CYCLE"
	TILL          Type = "TILL"
	HALT          Type = "HALT"
	SKIP          Type = "SKIP"
	BRANCH        Type = "BRANCH"
	PATH_KW       Type = "PATH_KW"
	ALIGN         Type = "ALIGN"
)

// Types & Data Keywords
const (
	FORM       Type = "FORM"
	SCHEMA     Type = "SCHEMA"
	PROTOCOL   Type = "PROTOCOL"
	VARIANT    Type = "VARIANT"
	KIND       Type = "KIND"
	FIXED      Type = "FIXED"
	ETERNAL    Type = "ETERNAL"
	STREAM     Type = "STREAM"
	MAPPING    Type = "MAPPING"
	SEQUENCE   Type = "SEQUENCE"
	COLLECTION Type = "COLLECTION"
	VOID_KW    Type = "VOID_KW"
	AFFIRM     Type = "AFFIRM"
	DENY       Type = "DENY"
)

// Module Keywords
const (
	ACCESS    Type = "ACCESS"
	EXPOSE    Type = "EXPOSE"
	NAMESPACE Type = "NAMESPACE"
	BUNDLE    Type = "BUNDLE"
)

// AGI & ML Keywords
const (
	NETWORK       Type = "NETWORK"
	LEARN         Type = "LEARN"
	TRAIN         Type = "TRAIN"
	INFER         Type = "INFER"
	EMIT          Type = "EMIT"
	OBSERVE       Type = "OBSERVE"
	ON            Type = "ON"
	WITH          Type = "WITH"
	DENSE         Type = "DENSE"
	CONV2D        Type = "CONV2D"
	TRANSFORMER   Type = "TRANSFORMER"
	ATTENTION     Type = "ATTENTION"
	LAYERNORM     Type = "LAYERNORM"
	RMSNORM       Type = "RMSNORM"
	GELU          Type = "GELU"
	SILU          Type = "SILU"
	SOFTMAX       Type = "SOFTMAX"
	DROPOUT       Type = "DROPOUT"
	EMBEDDING     Type = "EMBEDDING"
	MATMUL_KW     Type = "MATMUL_KW"
	RELU          Type = "RELU"
	SIGMOID       Type = "SIGMOID"
	CONFIDENT     Type = "CONFIDENT"
	EXPLORE       Type = "EXPLORE"
	LOSS          Type = "LOSS"
	OPTIMIZER     Type = "OPTIMIZER"
	EPOCHS        Type = "EPOCHS"
	LR            Type = "LR"
	BATCH_SIZE    Type = "BATCH_SIZE"
	MSE           Type = "MSE"
	CROSS_ENTROPY Type = "CROSS_ENTROPY"
	SGD           Type = "SGD"
	ADAM          Type = "ADAM"
)

// Quantum Keywords
const (
	SUPERPOSE Type = "SUPERPOSE"
	ENTANGLE  Type = "ENTANGLE"
	COLLAPSE  Type = "COLLAPSE"
	MEASURE   Type = "MEASURE"
	PREPARE   Type = "PREPARE"
	EVOLVE    Type = "EVOLVE"
	HERMITIAN Type = "HERMITIAN"
	KET_KW    Type = "KET_KW"
	BRA_KW    Type = "BRA_KW"
	INNER     Type = "INNER"
	OUTER     Type = "OUTER"
	MOD       Type = "MOD"
)

// Prime-Symbolic Keywords
const (
	CONCEPT   Type = "CONCEPT"
	DECOMPOSE Type = "DECOMPOSE"
)

// Data Pipeline Keywords
const (
	DATASET   Type = "DATASET"
	LOAD      Type = "LOAD"
	NORMALIZE Type = "NORMALIZE"
	MINMAX    Type = "MINMAX"
	ZSCORE    Type = "ZSCORE"
	SHUFFLE   Type = "SHUFFLE"
	BATCH     Type = "BATCH"
)

// CIR (Cognitive Intermediate Representation) Keywords
const (
	ASSERT    Type = "ASSERT"
	PROPAGATE Type = "PROPAGATE"
	EPISTEMIC Type = "EPISTEMIC"
	UNKNOWN   Type = "UNKNOWN"
	KNOWN     Type = "KNOWN"
)

// AGI Native Types Keywords
const (
	MEMORY    Type = "MEMORY"
	GRAPH     Type = "GRAPH"
	PROCEDURE Type = "PROCEDURE"
	RECALL    Type = "RECALL"
	REMEMBER  Type = "REMEMBER"
	LINK      Type = "LINK"
	NODE      Type = "NODE"
	TRACE     Type = "TRACE"
	EXECUTE   Type = "EXECUTE"
)

// QEEH (Quantum-Entangled Ethical Hamiltonian) Keywords
const (
	ETHICAL_BOUNDARY Type = "ETHICAL_BOUNDARY"
	ENFORCE          Type = "ENFORCE"
	DECOHERENCE      Type = "DECOHERENCE"
	HAMILTONIAN      Type = "HAMILTONIAN"
	RULE             Type = "RULE"
)

// QEMWV (Quantum-Entangled Model Weight Vault) Keywords
const (
	VAULT Type = "VAULT"
	WIPE  Type = "WIPE"
)

// MCTP (Merkle-Certified Training Provenance) Keywords
const (
	PROVENANCE   Type = "PROVENANCE"
	CERTIFY      Type = "CERTIFY"
	VERIFY_TRAIN Type = "VERIFY_TRAIN"
)

// QEMP (Quantum-Entangled Memory Provenance) Keywords
const (
	MEMORY_PROV   Type = "MEMORY_PROV"
	VERIFY_MEMORY Type = "VERIFY_MEMORY"
	AUDIT_MEMORY  Type = "AUDIT_MEMORY"
	CORRUPT_SHARE Type = "CORRUPT_SHARE"
)

// QECMT (Quantum-Entangled Cross-Model Transfer) Keywords
const (
	LICENSE  Type = "LICENSE"
	TRANSFER Type = "TRANSFER"
	DISTILL  Type = "DISTILL"
	TEACHER  Type = "TEACHER"
	STUDENT  Type = "STUDENT"
)

// QERA (Quantum-Entangled Reasoning Attestation) Keywords
const (
	REASONING_CHAIN        Type = "REASONING_CHAIN"
	REASON                 Type = "REASON"
	DEDUCE                 Type = "DEDUCE"
	INDUCE                 Type = "INDUCE"
	ABDUCE                 Type = "ABDUCE"
	VERIFY_CHAIN           Type = "VERIFY_CHAIN"
	AUDIT_REASONING        Type = "AUDIT_REASONING"
	CORRUPT_REASONING_STEP Type = "CORRUPT_REASONING_STEP"
)

// QECMA (Quantum-Entangled Cross-Modal Attestation) Keywords
const (
	MODALITY           Type = "MODALITY"
	CROSS_MODAL_REASON Type = "CROSS_MODAL_REASON"
	TENSOR_INPUT       Type = "TENSOR_INPUT"
	TEXT_INPUT         Type = "TEXT_INPUT"
	CONCEPT_INPUT      Type = "CONCEPT_INPUT"
)

// QEC-QEMP (Quantum Error-Corrected Memory Provenance — Invention #13) Keywords
const (
	QEC_MEMORY       Type = "QEC_MEMORY"
	SYNDROME         Type = "SYNDROME"
	CORRECT          Type = "CORRECT"
	VERIFY_SYNDROME  Type = "VERIFY_SYNDROME"
	QEC_CODE         Type = "QEC_CODE"
	SURFACE_CODE     Type = "SURFACE_CODE"
	QLDPC_CODE       Type = "QLDPC_CODE"
	CORRUPT_SYNDROME Type = "CORRUPT_SYNDROME"
)

// Memory Consolidation Keywords (M5.6)
const (
	CONSOLIDATE Type = "CONSOLIDATE"
	PROMOTE_TO  Type = "PROMOTE_TO"
	EPISODIC    Type = "EPISODIC"
	SEMANTIC    Type = "SEMANTIC"
)

// CB-SRCPP (Invention #14) Keywords
const (
	RECALL_TOKEN  Type = "RECALL_TOKEN"
	GRANT         Type = "GRANT"
	REVOKE        Type = "REVOKE"
	SCOPED_ACCESS Type = "SCOPED_ACCESS"
	RECALL_SCOPED Type = "RECALL_SCOPED"
)

// Safe Tool Orchestration (M5.8) Keywords
const (
	TOOL_REGISTER  Type = "TOOL_REGISTER"
	TOOL_CALL      Type = "TOOL_CALL"
	TOOL_AUDIT     Type = "TOOL_AUDIT"
	PRE_CONDITION  Type = "PRE_CONDITION"
	POST_CONDITION Type = "POST_CONDITION"
)

// TTPCR (Invention #15) & Calibrated Abstention Keywords
const (
	CERTIFIED_REASON  Type = "CERTIFIED_REASON"
	CERT_KIND         Type = "CERT_KIND"
	WITNESS           Type = "WITNESS"
	ABSTAIN           Type = "ABSTAIN"
	RETRY             Type = "RETRY"
	ESCALATE          Type = "ESCALATE"
	REFUSE            Type = "REFUSE"
	ABSTAIN_THRESHOLD Type = "ABSTAIN_THRESHOLD"
	ABSTAIN_CHECK     Type = "ABSTAIN_CHECK"
)

// QESI (Invention #16: Quantum-Entangled Sandbox Isolation) Keywords
const (
	SANDBOX        Type = "SANDBOX"
	SANDBOX_CREATE Type = "SANDBOX_CREATE"
	SANDBOX_RUN    Type = "SANDBOX_RUN"
	SANDBOX_VERIFY Type = "SANDBOX_VERIFY"
	QUARANTINE     Type = "QUARANTINE"
)

// Phase 6: DirectCompute GPU & AGKE (Invention #17) Keywords
const (
	AT_GPU             Type = "AT_GPU"
	GPU_MATMUL         Type = "GPU_MATMUL"
	GPU_SOFTMAX        Type = "GPU_SOFTMAX"
	GPU_LAYERNORM      Type = "GPU_LAYERNORM"
	AGKE_ATTEST        Type = "AGKE_ATTEST"
	AGKE_VERIFY        Type = "AGKE_VERIFY"
	CORRUPT_GPU_DIGEST Type = "CORRUPT_GPU_DIGEST"

	// M6.3: Multi-GPU Pipeline Parallelism Keywords
	PIPELINE_CLAIM Type = "PIPELINE_CLAIM"
	PIPELINE_RUN   Type = "PIPELINE_RUN"
	MULTI_DEVICE   Type = "MULTI_DEVICE"

	// M6.4: CDAGA (Invention #18: Cross-Device Attested Gradient Aggregation) Keywords
	CDAGA_RECORD  Type = "CDAGA_RECORD"
	CDAGA_ATTEST  Type = "CDAGA_ATTEST"
	CDAGA_VERIFY  Type = "CDAGA_VERIFY"
	CORRUPT_CDAGA Type = "CORRUPT_CDAGA"

	// M6.5: Distributed Training Keywords
	ALL_REDUCE   Type = "ALL_REDUCE"
	PSERVER_INIT Type = "PSERVER_INIT"
	PSERVER_AGG  Type = "PSERVER_AGG"
	DIST_INIT    Type = "DIST_INIT"

	// M6.6: BARAT (Invention #19: Byzantine-Attested Robust Aggregation) Keywords
	BARAT_SUBMIT    Type = "BARAT_SUBMIT"
	BARAT_AGGREGATE Type = "BARAT_AGGREGATE"
	BARAT_VERIFY    Type = "BARAT_VERIFY"
	CORRUPT_BARAT   Type = "CORRUPT_BARAT"

	// M7.1: Quantum Hardware Abstraction Layer (QHAL) Keywords
	QASM_EXPORT  Type = "QASM_EXPORT"
	QHAL_BACKEND Type = "QHAL_BACKEND"

	// M7.2: SEA (Invention #20: Self-Evolution Attestation) Keywords
	EVOLVE_VERIFY  Type = "EVOLVE_VERIFY"
	EVOLVE_HISTORY Type = "EVOLVE_HISTORY"
	CORRUPT_EVOLVE Type = "CORRUPT_EVOLVE"

	// M7.3: Hardware Target Transpilation Keywords
	TRANSPILE_TO   Type = "TRANSPILE_TO"
	TRANSPILE_OPTS Type = "TRANSPILE_OPTS"

	// M7.4: Evolution Operators & EEA (Invention #21) Keywords
	EVOLVE_POLICY  Type = "EVOLVE_POLICY"
	EVOLVE_BOUNDED Type = "EVOLVE_BOUNDED"
	EEA_VERIFY     Type = "EEA_VERIFY"
	CORRUPT_EEA    Type = "CORRUPT_EEA"

	// M7.5: Real QEC Decoder Keywords
	QECD_DECODE Type = "QECD_DECODE"
	QECD_VERIFY Type = "QECD_VERIFY"
	QECD_APPLY  Type = "QECD_APPLY"

	// M7.6: Meta-Learning & AAD (Invention #22) Keywords
	META_FIT        Type = "META_FIT"
	META_ADAPT      Type = "META_ADAPT"
	META_VERIFY     Type = "META_VERIFY"
	AAD_CERT        Type = "AAD_CERT"
	AAD_ATTEST      Type = "AAD_ATTEST"
	CORRUPT_AAD_ENV Type = "CORRUPT_AAD_ENV"
)

// Quasinary Keywords & Literals
const (
	QUASI      Type = "QUASI"
	QUASI_NEG  Type = "QUASI_NEG"  // ◄ \qneg
	QUASI_ZERO Type = "QUASI_ZERO" // 〇 \qzero
	QUASI_POS  Type = "QUASI_POS"  // ► \qpos
	ENERGY     Type = "ENERGY"
	MINIMIZE   Type = "MINIMIZE"
	TILE       Type = "TILE"
)

// Cryptography Keywords
const (
	SIGN    Type = "SIGN"
	VERIFY  Type = "VERIFY"
	SEAL    Type = "SEAL"
	UNSEAL  Type = "UNSEAL"
	KEYGEN  Type = "KEYGEN"
	PQC_KEY Type = "PQC_KEY"
	ENTROPY Type = "ENTROPY"
)

// Concurrency Keywords
const (
	PARALLEL Type = "PARALLEL"
	ATOMIC   Type = "ATOMIC"
	SPAWN    Type = "SPAWN"
	FORK     Type = "FORK"
	JOIN     Type = "JOIN"
)

// Token represents a lexical token in Qadav source code.
type Token struct {
	Type    Type
	Literal string // Original Unicode or fallback, as typed
	Line    int    // 1-based line number
	Col     int    // 1-based column number (runes)
	File    string // Source file
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("%s(%s) @ %s:%d:%d", t.Type, t.Literal, t.File, t.Line, t.Col)
}

// OpInfo stores information for each operator.
type OpInfo struct {
	Type     Type
	Unicode  string
	Fallback string
	Meaning  string
}

// AllOperators lists every quantum-native operator supported in Qadav.
func AllOperators() []OpInfo {
	return []OpInfo{
		// Arithmetic
		{OSUM, "⊕", `\oplus`, "Direct sum / addition"},
		{ODIFF, "⊖", `\ominus`, "Direct difference"},
		{OTIM, "⊗", `\otimes`, "Tensor product / multiplication"},
		{ODIV, "⊘", `\oslash`, "Division"},
		{OMOD, "⊙", `\odot`, "Modulo (circle-dot)"},
		{OPOW, "↑", `\uparrow`, "Power / exponent"},
		{MATMUL, "⊛", `\matmul`, "Matrix multiplication"},

		// Comparison
		{EQUIV, "≡", `\equiv`, "Equal"},
		{NEQ, "≠", `\neq`, "Not equal"},
		{PREC, "≺", `\prec`, "Less than"},
		{SUCC, "≻", `\succ`, "Greater than"},
		{PREQ, "≼", `\preceq`, "Less or equal"},
		{SUEQ, "≽", `\succeq`, "Greater or equal"},
		{APPROX, "≈", `\approx`, "Approx equal"},
		{MEMBER, "∈", `\in`, "Belongs to"},
		{NOTMEMBER, "∉", `\notin`, "Not belongs to"},

		// Logical
		{WEDGE, "∧", `\wedge`, "Logical AND"},
		{VEE, "∨", `\vee`, "Logical OR"},
		{NEG, "¬", `\neg`, "Logical NOT"},
		{XVEE, "⊻", `\xor`, "XOR"},

		// Binding & Flow
		{BIND, "↦", `\bind`, "Immutable binding"},
		{ASSIGN, "←", `\assign`, "Mutable assignment"},
		{YIELDS, "→", `\yields`, "Yield / return value"},
		{IMPLIES, "⇒", `\implies`, "Implies (for lambdas)"},
		{PIPELINE, "↝", `\pipe`, "Pipeline operator"},

		// Compound Assignment
		{BIND_ADD, "↦⊕", `\bindplus`, "bind-add"},
		{BIND_SUB, "↦⊖", `\bindminus`, "bind-sub"},
		{BIND_MUL, "↦⊗", `\bindtim`, "bind-mul"},
		{BIND_DIV, "↦⊘", `\binddiv`, "bind-div"},

		// Structural
		{GROUP_OPEN, "⦅", `\group`, "Open group"},
		{GROUP_CLOSE, "⦆", `\egroup`, "Close group"},
		{BLOCK_OPEN, "⟦", `\block`, "Open block"},
		{BLOCK_CLOSE, "⟧", `\eblock`, "Close block"},
		{INDEX_OPEN, "⟬", `\bracket`, "Open index"},
		{INDEX_CLOSE, "⟭", `\ebracket`, "Close index"},
		{STMT_END, "·", `\stmt`, "Statement end"},
		{SEP, "⸴", `\sep`, "Item separator"},
		{ANNOT, "∵", `\because`, "Type/annotation"},
		{MEMBER_ACCESS, "∸", `\dot`, "Member access"},
		{RANGE, "…", `\range`, "Range"},
		{LANGLE, "⟨", `\langle`, "Left angle bracket (type param)"},
		{RANGLE, "⟩", `\rangle`, "Right angle bracket (type param)"},

		// Quantum-Native
		{HADAMARD, "ℋ", `\had`, "Hadamard gate"},
		{PAULI_X, "𝒳", `\pauliX`, "Pauli X gate"},
		{PAULI_Y, "𝒴", `\pauliY`, "Pauli Y gate"},
		{PAULI_Z, "𝒵", `\pauliZ`, "Pauli Z gate"},
		{CNOT, "ℂ", `\cnot`, "CNOT gate"},
		{SGATE, "𝕊", `\sgate`, "S phase gate"},
		{TGATE, "𝕋", `\tgate`, "T phase gate"},
		{UNITARY, "𝕌", `\unitary`, "Unitary operator"},
		{PROBABILITY, "ℙ", `\prob`, "Probability"},
		{ADJOINT, "⋆", `\adjoint`, "Adjoint (†)"},
		{SUM, "∑", `\sum`, "Sum reduction"},
		{PRODUCT, "∏", `\prod`, "Product reduction"},
		{GRADIENT, "∂", `\grad`, "Gradient"},
		{DIVERGENCE, "∇", `\div`, "Divergence"},
		{SQRT, "√", `\sqrt`, "Square root"},
		{INFINITY, "∞", `\inf`, "Infinity"},
		{EMPTY, "∅", `\empty`, "Empty/null"},

		// Mathematical
		{FORALL, "∀", `\forall`, "For all"},
		{EXISTS, "∃", `\exists`, "There exists"},
		{UNION, "∪", `\union`, "Set union"},
		{INTERSECT, "∩", `\inter`, "Set intersection"},
		{SUBSET, "⊂", `\subset`, "Proper subset"},
		{SUBSET_EQ, "⊆", `\subseteq`, "Subset or equal"},
		{TOP, "⊤", `\top`, "Top type / truth"},
		{BOTTOM, "⊥", `\bottom`, "Bottom type / false"},

		// Quasinary Literals
		{QUASI_NEG, "◄", `\qneg`, "Quasi negative (-1)"},
		{QUASI_ZERO, "〇", `\qzero`, "Quasi zero (0)"},
		{QUASI_POS, "►", `\qpos`, "Quasi positive (+1)"},
	}
}

// Keyword definitions
var keywords = map[string]Type{
	// Types
	"int8":       INT8,
	"int16":      INT16,
	"int32":      INT32,
	"int64":      INT64,
	"uint8":      UINT8,
	"uint16":     UINT16,
	"uint32":     UINT32,
	"uint64":     UINT64,
	"float32":    FLOAT32,
	"float64":    FLOAT64,
	"complex64":  COMPLEX64,
	"complex128": COMPLEX128,
	"logic":      LOGIC,
	"text":       TEXT,
	"tensor":     TENSOR,
	"qubit":      QUBIT,
	"neuron":     NEURON,

	// Process & Flow
	"process":    PROCESS,
	"yield":      YIELD_KW,
	"flux":       FLUX,
	"concurrent": CONCURRENT,
	"converge":   CONVERGE,

	// Control
	"when":          WHEN,
	"otherwise":     OTHERWISE,
	"alternatively": ALTERNATIVELY,
	"iterate":       ITERATE,
	"cycle":         CYCLE,
	"till":          TILL,
	"halt":          HALT,
	"skip":          SKIP,
	"branch":        BRANCH,
	"path":          PATH_KW,
	"align":         ALIGN,

	// Types & Data
	"form":       FORM,
	"schema":     SCHEMA,
	"protocol":   PROTOCOL,
	"variant":    VARIANT,
	"kind":       KIND,
	"fixed":      FIXED,
	"eternal":    ETERNAL,
	"stream":     STREAM,
	"mapping":    MAPPING,
	"sequence":   SEQUENCE,
	"collection": COLLECTION,
	"void":       VOID_KW,
	"affirm":     AFFIRM,
	"deny":       DENY,

	// Module
	"access":    ACCESS,
	"expose":    EXPOSE,
	"namespace": NAMESPACE,
	"bundle":    BUNDLE,

	// AGI / ML
	"network":       NETWORK,
	"learn":         LEARN,
	"train":         TRAIN,
	"infer":         INFER,
	"emit":          EMIT,
	"observe":       OBSERVE,
	"on":            ON,
	"with":          WITH,
	"dense":         DENSE,
	"conv2d":        CONV2D,
	"transformer":   TRANSFORMER,
	"attention":     ATTENTION,
	"layernorm":     LAYERNORM,
	"rmsnorm":       RMSNORM,
	"gelu":          GELU,
	"silu":          SILU,
	"softmax":       SOFTMAX,
	"dropout":       DROPOUT,
	"embedding":     EMBEDDING,
	"matmul":        MATMUL_KW,
	"relu":          RELU,
	"sigmoid":       SIGMOID,
	"confident":     CONFIDENT,
	"explore":       EXPLORE,
	"loss":          LOSS,
	"optimizer":     OPTIMIZER,
	"epochs":        EPOCHS,
	"lr":            LR,
	"batch_size":    BATCH_SIZE,
	"mse":           MSE,
	"cross_entropy": CROSS_ENTROPY,
	"sgd":           SGD,
	"adam":          ADAM,

	// Quantum
	"superpose": SUPERPOSE,
	"entangle":  ENTANGLE,
	"collapse":  COLLAPSE,
	"measure":   MEASURE,
	"prepare":   PREPARE,
	"evolve":    EVOLVE,
	"hermitian": HERMITIAN,
	"ket":       KET_KW,
	"bra":       BRA_KW,
	"inner":     INNER,
	"outer":     OUTER,
	"mod":       MOD,

	// Prime-Symbolic
	"concept":   CONCEPT,
	"decompose": DECOMPOSE,

	// Data Pipeline
	"dataset":   DATASET,
	"load":      LOAD,
	"normalize": NORMALIZE,
	"minmax":    MINMAX,
	"zscore":    ZSCORE,
	"shuffle":   SHUFFLE,
	"batch":     BATCH,

	// Quasinary Computing
	"quasi":    QUASI,
	"energy":   ENERGY,
	"minimize": MINIMIZE,
	"tile":     TILE,

	// CIR (Cognitive Intermediate Representation)
	"assert":    ASSERT,
	"propagate": PROPAGATE,
	"epistemic": EPISTEMIC,
	"unknown":   UNKNOWN,
	"known":     KNOWN,

	// AGI Native Types
	"memory":    MEMORY,
	"graph":     GRAPH,
	"procedure": PROCEDURE,
	"recall":    RECALL,
	"remember":  REMEMBER,
	"link":      LINK,
	"node":      NODE,
	"trace":     TRACE,
	"execute":   EXECUTE,

	// QEEH (Quantum-Entangled Ethical Hamiltonian)
	"ethical_boundary": ETHICAL_BOUNDARY,
	"enforce":          ENFORCE,
	"decoherence":      DECOHERENCE,
	"hamiltonian":      HAMILTONIAN,
	"rule":             RULE,

	// QEMWV (Quantum-Entangled Model Weight Vault)
	"vault": VAULT,
	"wipe":  WIPE,

	// MCTP (Merkle-Certified Training Provenance)
	"provenance":      PROVENANCE,
	"certify":         CERTIFY,
	"verify_training": VERIFY_TRAIN,

	// QEMP (Quantum-Entangled Memory Provenance)
	"memory_prov":          MEMORY_PROV,
	"verify_memory":        VERIFY_MEMORY,
	"audit_memory":         AUDIT_MEMORY,
	"corrupt_memory_share": CORRUPT_SHARE,

	// QECMT (Quantum-Entangled Cross-Model Transfer)
	"license":  LICENSE,
	"transfer": TRANSFER,
	"distill":  DISTILL,
	"teacher":  TEACHER,
	"student":  STUDENT,

	// QERA (Quantum-Entangled Reasoning Attestation)
	"reasoning_chain":        REASONING_CHAIN,
	"reason":                 REASON,
	"deduce":                 DEDUCE,
	"induce":                 INDUCE,
	"abduce":                 ABDUCE,
	"verify_chain":           VERIFY_CHAIN,
	"audit_reasoning":        AUDIT_REASONING,
	"corrupt_reasoning_step": CORRUPT_REASONING_STEP,

	// QECMA
	"modality":           MODALITY,
	"cross_modal_reason": CROSS_MODAL_REASON,
	"tensor_input":       TENSOR_INPUT,
	"text_input":         TEXT_INPUT,
	"concept_input":      CONCEPT_INPUT,

	// QEC-QEMP (Invention #13)
	"qec_memory":       QEC_MEMORY,
	"syndrome":         SYNDROME,
	"correct":          CORRECT,
	"verify_syndrome":  VERIFY_SYNDROME,
	"qec_code":         QEC_CODE,
	"surface_code":     SURFACE_CODE,
	"qldpc_code":       QLDPC_CODE,
	"corrupt_syndrome": CORRUPT_SYNDROME,

	// Memory Consolidation (M5.6)
	"consolidate": CONSOLIDATE,
	"promote_to":  PROMOTE_TO,
	"episodic":    EPISODIC,
	"semantic":    SEMANTIC,

	// CB-SRCPP (Invention #14)
	"recall_token":  RECALL_TOKEN,
	"grant":         GRANT,
	"revoke":        REVOKE,
	"scoped_access": SCOPED_ACCESS,
	"recall_scoped": RECALL_SCOPED,

	// Safe Tool Orchestration (M5.8)
	"tool_register":  TOOL_REGISTER,
	"tool_call":      TOOL_CALL,
	"tool_audit":     TOOL_AUDIT,
	"pre_condition":  PRE_CONDITION,
	"post_condition": POST_CONDITION,

	// TTPCR (Invention #15) & Calibrated Abstention
	"certified_reason":  CERTIFIED_REASON,
	"cert_kind":         CERT_KIND,
	"witness":           WITNESS,
	"abstain":           ABSTAIN,
	"retry":             RETRY,
	"escalate":          ESCALATE,
	"refuse":            REFUSE,
	"abstain_threshold": ABSTAIN_THRESHOLD,
	"abstain_check":     ABSTAIN_CHECK,

	// QESI (Invention #16: Quantum-Entangled Sandbox Isolation)
	"sandbox":        SANDBOX,
	"sandbox_create": SANDBOX_CREATE,
	"sandbox_run":    SANDBOX_RUN,
	"sandbox_verify": SANDBOX_VERIFY,
	"quarantine":     QUARANTINE,

	// Phase 6: DirectCompute GPU & AGKE (Invention #17)
	"@gpu":               AT_GPU,
	"gpu_matmul":         GPU_MATMUL,
	"gpu_softmax":        GPU_SOFTMAX,
	"gpu_layernorm":      GPU_LAYERNORM,
	"agke_attest":        AGKE_ATTEST,
	"agke_verify":        AGKE_VERIFY,
	"corrupt_gpu_digest": CORRUPT_GPU_DIGEST,

	// M6.3: Multi-GPU Pipeline Parallelism
	"@multi_device":  MULTI_DEVICE,
	"pipeline_claim": PIPELINE_CLAIM,
	"pipeline_run":   PIPELINE_RUN,

	// M6.4: CDAGA (Invention #18)
	"cdaga_record":  CDAGA_RECORD,
	"cdaga_attest":  CDAGA_ATTEST,
	"cdaga_verify":  CDAGA_VERIFY,
	"corrupt_cdaga": CORRUPT_CDAGA,

	// M6.5: Distributed Training
	"all_reduce":        ALL_REDUCE,
	"pserver_init":      PSERVER_INIT,
	"pserver_aggregate": PSERVER_AGG,
	"@distributed":      DIST_INIT,

	// M6.6: BARAT (Invention #19)
	"barat_submit":    BARAT_SUBMIT,
	"barat_aggregate": BARAT_AGGREGATE,
	"barat_verify":    BARAT_VERIFY,
	"corrupt_barat":   CORRUPT_BARAT,

	// M7.1: Quantum Hardware Abstraction Layer (QHAL)
	"qasm_export":   QASM_EXPORT,
	"@qhal_backend": QHAL_BACKEND,

	// M7.2: SEA (Invention #20)
	"evolve_verify":  EVOLVE_VERIFY,
	"evolve_history": EVOLVE_HISTORY,
	"corrupt_evolve": CORRUPT_EVOLVE,

	// M7.3: Hardware Target Transpilation
	"transpile_to":   TRANSPILE_TO,
	"transpile_opts": TRANSPILE_OPTS,

	// M7.4: Evolution Operators & EEA (Invention #21)
	"evolve_policy":  EVOLVE_POLICY,
	"evolve_bounded": EVOLVE_BOUNDED,
	"eea_verify":     EEA_VERIFY,
	"corrupt_eea":    CORRUPT_EEA,

	// M7.5: Real QEC Decoder
	"qecd_decode": QECD_DECODE,
	"qecd_verify": QECD_VERIFY,
	"qecd_apply":  QECD_APPLY,

	// M7.6: Meta-Learning & AAD (Invention #22)
	"meta_fit":        META_FIT,
	"meta_adapt":      META_ADAPT,
	"meta_verify":     META_VERIFY,
	"aad_cert":        AAD_CERT,
	"aad_attest":      AAD_ATTEST,
	"corrupt_aad_env": CORRUPT_AAD_ENV,

	// Cryptography
	"sign":    SIGN,
	"verify":  VERIFY,
	"seal":    SEAL,
	"unseal":  UNSEAL,
	"keygen":  KEYGEN,
	"pqc_key": PQC_KEY,
	"entropy": ENTROPY,

	// Concurrency
	"parallel": PARALLEL,
	"atomic":   ATOMIC,
	"spawn":    SPAWN,
	"fork":     FORK,
	"join":     JOIN,
}

// Fallback map mapping LaTeX-style strings to token types
var fallbacks = func() map[string]Type {
	m := make(map[string]Type)
	for _, op := range AllOperators() {
		m[op.Fallback] = op.Type
	}
	kwFallbacks := map[string]Type{
		`\dataset`:                DATASET,
		`\load`:                   LOAD,
		`\normalize`:              NORMALIZE,
		`\minmax`:                 MINMAX,
		`\zscore`:                 ZSCORE,
		`\shuffle`:                SHUFFLE,
		`\batch`:                  BATCH,
		`\quasi`:                  QUASI,
		`\qneg`:                   QUASI_NEG,
		`\qzero`:                  QUASI_ZERO,
		`\qpos`:                   QUASI_POS,
		`\energy`:                 ENERGY,
		`\minimize`:               MINIMIZE,
		`\tile`:                   TILE,
		`\infer`:                  INFER,
		`\assert`:                 ASSERT,
		`\propagate`:              PROPAGATE,
		`\epistemic`:              EPISTEMIC,
		`\unknown`:                UNKNOWN,
		`\known`:                  KNOWN,
		`\memory`:                 MEMORY,
		`\graph`:                  GRAPH,
		`\procedure`:              PROCEDURE,
		`\recall`:                 RECALL,
		`\remember`:               REMEMBER,
		`\link`:                   LINK,
		`\node`:                   NODE,
		`\trace`:                  TRACE,
		`\execute`:                EXECUTE,
		`\ethical`:                ETHICAL_BOUNDARY,
		`\enforce`:                ENFORCE,
		`\decoherence`:            DECOHERENCE,
		`\hamiltonian`:            HAMILTONIAN,
		`\rule`:                   RULE,
		`\vault`:                  VAULT,
		`\seal`:                   SEAL,
		`\unseal`:                 UNSEAL,
		`\wipe`:                   WIPE,
		`\provenance`:             PROVENANCE,
		`\certify`:                CERTIFY,
		`\verify_training`:        VERIFY_TRAIN,
		`\memory_prov`:            MEMORY_PROV,
		`\verify_memory`:          VERIFY_MEMORY,
		`\audit_memory`:           AUDIT_MEMORY,
		`\corrupt_share`:          CORRUPT_SHARE,
		`\corrupt_memory_share`:   CORRUPT_SHARE,
		`\license`:                LICENSE,
		`\transfer`:               TRANSFER,
		`\distill`:                DISTILL,
		`\teacher`:                TEACHER,
		`\student`:                STUDENT,
		`\reasoning_chain`:        REASONING_CHAIN,
		`\reason`:                 REASON,
		`\deduce`:                 DEDUCE,
		`\induce`:                 INDUCE,
		`\abduce`:                 ABDUCE,
		`\verify_chain`:           VERIFY_CHAIN,
		`\audit_reasoning`:        AUDIT_REASONING,
		`\corrupt_reasoning_step`: CORRUPT_REASONING_STEP,
		`\modality`:               MODALITY,
		`\cross_modal_reason`:     CROSS_MODAL_REASON,
		`\tensor_input`:           TENSOR_INPUT,
		`\text_input`:             TEXT_INPUT,
		`\concept_input`:          CONCEPT_INPUT,
		`\qec_memory`:             QEC_MEMORY,
		`\syndrome`:               SYNDROME,
		`\correct`:                CORRECT,
		`\verify_syndrome`:        VERIFY_SYNDROME,
		`\qec_code`:               QEC_CODE,
		`\surface_code`:           SURFACE_CODE,
		`\qldpc_code`:             QLDPC_CODE,
		`\corrupt_syndrome`:       CORRUPT_SYNDROME,
		`\consolidate`:            CONSOLIDATE,
		`\promote_to`:             PROMOTE_TO,
		`\episodic`:               EPISODIC,
		`\semantic`:               SEMANTIC,
		`\recall_token`:           RECALL_TOKEN,
		`\grant`:                  GRANT,
		`\revoke`:                 REVOKE,
		`\scoped_access`:          SCOPED_ACCESS,
		`\recall_scoped`:          RECALL_SCOPED,
		`\tool_register`:          TOOL_REGISTER,
		`\tool_call`:              TOOL_CALL,
		`\tool_audit`:             TOOL_AUDIT,
		`\pre_condition`:          PRE_CONDITION,
		`\post_condition`:         POST_CONDITION,
		`\certified_reason`:       CERTIFIED_REASON,
		`\cert_kind`:              CERT_KIND,
		`\witness`:                WITNESS,
		`\abstain`:                ABSTAIN,
		`\retry`:                  RETRY,
		`\escalate`:               ESCALATE,
		`\refuse`:                 REFUSE,
		`\abstain_threshold`:      ABSTAIN_THRESHOLD,
		`\abstain_check`:          ABSTAIN_CHECK,
		`\sandbox`:                SANDBOX,
		`\sandbox_create`:         SANDBOX_CREATE,
		`\sandbox_run`:            SANDBOX_RUN,
		`\sandbox_verify`:         SANDBOX_VERIFY,
		`\quarantine`:             QUARANTINE,
		`\gpu`:                    AT_GPU,
		`\at_gpu`:                 AT_GPU,
		`\gpu_matmul`:             GPU_MATMUL,
		`\gpu_softmax`:            GPU_SOFTMAX,
		`\gpu_layernorm`:          GPU_LAYERNORM,
		`\agke_attest`:            AGKE_ATTEST,
		`\agke_verify`:            AGKE_VERIFY,
		`\corrupt_gpu_digest`:     CORRUPT_GPU_DIGEST,
		`\pipeline_claim`:         PIPELINE_CLAIM,
		`\pipeline_run`:           PIPELINE_RUN,
		`\multi_device`:           MULTI_DEVICE,
		`\cdaga_record`:           CDAGA_RECORD,
		`\cdaga_attest`:           CDAGA_ATTEST,
		`\cdaga_verify`:           CDAGA_VERIFY,
		`\corrupt_cdaga`:          CORRUPT_CDAGA,
		`\all_reduce`:             ALL_REDUCE,
		`\pserver_init`:           PSERVER_INIT,
		`\pserver_aggregate`:      PSERVER_AGG,
		`\distributed`:            DIST_INIT,
		`\barat_submit`:           BARAT_SUBMIT,
		`\barat_aggregate`:        BARAT_AGGREGATE,
		`\barat_verify`:           BARAT_VERIFY,
		`\corrupt_barat`:          CORRUPT_BARAT,
		`\qasm_export`:            QASM_EXPORT,
		`\qhal_backend`:           QHAL_BACKEND,
		`\evolve`:                 EVOLVE,
		`\evolve_verify`:          EVOLVE_VERIFY,
		`\evolve_history`:         EVOLVE_HISTORY,
		`\corrupt_evolve`:         CORRUPT_EVOLVE,
		`\transpile_to`:           TRANSPILE_TO,
		`\transpile_opts`:         TRANSPILE_OPTS,
		`\evolve_policy`:          EVOLVE_POLICY,
		`\evolve_bounded`:         EVOLVE_BOUNDED,
		`\eea_verify`:             EEA_VERIFY,
		`\corrupt_eea`:            CORRUPT_EEA,
		`\qecd_decode`:            QECD_DECODE,
		`\qecd_verify`:            QECD_VERIFY,
		`\qecd_apply`:             QECD_APPLY,
		`\meta_fit`:               META_FIT,
		`\meta_adapt`:             META_ADAPT,
		`\meta_verify`:            META_VERIFY,
		`\aad_cert`:               AAD_CERT,
		`\aad_attest`:             AAD_ATTEST,
		`\corrupt_aad_env`:        CORRUPT_AAD_ENV,
	}
	for k, v := range kwFallbacks {
		m[k] = v
	}
	return m
}()

// Unicode operator map
var unicodeOps = func() map[string]Type {
	m := make(map[string]Type)
	for _, op := range AllOperators() {
		m[op.Unicode] = op.Type
	}
	return m
}()

// LookupIdent inspects an identifier string and returns its token Type.
// If it matches a recognized keyword, the corresponding keyword Type is returned;
// otherwise IDENT is returned.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// LookupFallback maps a LaTeX-style fallback sequence (e.g. "\oplus") to its token Type.
func LookupFallback(name string) (Type, bool) {
	tok, ok := fallbacks[name]
	return tok, ok
}

// LookupUnicode maps a Unicode operator string (e.g. "⊕") to its token Type.
func LookupUnicode(sym string) (Type, bool) {
	tok, ok := unicodeOps[sym]
	return tok, ok
}

// AllKeywords returns a copy of the keyword map for testing and verification.
func AllKeywords() map[string]Type {
	res := make(map[string]Type, len(keywords))
	for k, v := range keywords {
		res[k] = v
	}
	return res
}
