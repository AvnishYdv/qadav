package ast

// GPUMatmulStmt represents:
// [@gpu] gpu_matmul a ⸴ b into c [audit chain] ·
type GPUMatmulStmt struct {
	Pos_        Position
	IsAnnotated bool
	Left        Expr
	Right       Expr
	Target      Expr
	Audit       Expr // optional QERA reasoning chain
}

func (s *GPUMatmulStmt) Pos() Position { return s.Pos_ }
func (s *GPUMatmulStmt) stmtNode()     {}

// GPUSoftmaxStmt represents:
// [@gpu] gpu_softmax x into y [audit chain] ·
type GPUSoftmaxStmt struct {
	Pos_        Position
	IsAnnotated bool
	Source      Expr
	Target      Expr
	Audit       Expr // optional QERA reasoning chain
}

func (s *GPUSoftmaxStmt) Pos() Position { return s.Pos_ }
func (s *GPUSoftmaxStmt) stmtNode()     {}

// GPULayernormStmt represents:
// [@gpu] gpu_layernorm x into y [audit chain] ·
type GPULayernormStmt struct {
	Pos_        Position
	IsAnnotated bool
	Source      Expr
	Target      Expr
	Audit       Expr // optional QERA reasoning chain
}

func (s *GPULayernormStmt) Pos() Position { return s.Pos_ }
func (s *GPULayernormStmt) stmtNode()     {}

// GPUMatmulExpr represents:
// gpu_matmul ⦅ left ⸴ right ⦆
type GPUMatmulExpr struct {
	Pos_  Position
	Left  Expr
	Right Expr
}

func (e *GPUMatmulExpr) Pos() Position { return e.Pos_ }
func (e *GPUMatmulExpr) exprNode()     {}

// GPUSoftmaxExpr represents:
// gpu_softmax ⦅ source ⦆
type GPUSoftmaxExpr struct {
	Pos_   Position
	Source Expr
}

func (e *GPUSoftmaxExpr) Pos() Position { return e.Pos_ }
func (e *GPUSoftmaxExpr) exprNode()     {}

// GPULayernormExpr represents:
// gpu_layernorm ⦅ source ⦆
type GPULayernormExpr struct {
	Pos_   Position
	Source Expr
}

func (e *GPULayernormExpr) Pos() Position { return e.Pos_ }
func (e *GPULayernormExpr) exprNode()     {}

// AGKEAttestStmt represents:
// agke_attest chain ⦅ kernel ⸴ in ⸴ out ⦆ ·
// or: agke_attest chain kernel ↦ "name" input ↦ in output ↦ out ·
type AGKEAttestStmt struct {
	Pos_         Position
	Chain        Expr
	KernelName   Expr
	InputTensor  Expr
	OutputTensor Expr
}

func (s *AGKEAttestStmt) Pos() Position { return s.Pos_ }
func (s *AGKEAttestStmt) stmtNode()     {}

// AGKEVerifyStmt represents:
// agke_verify chain into res ·
type AGKEVerifyStmt struct {
	Pos_   Position
	Chain  Expr
	Target Expr
}

func (s *AGKEVerifyStmt) Pos() Position { return s.Pos_ }
func (s *AGKEVerifyStmt) stmtNode()     {}

// AGKEVerifyExpr represents:
// agke_verify chain
type AGKEVerifyExpr struct {
	Pos_  Position
	Chain Expr
}

func (e *AGKEVerifyExpr) Pos() Position { return e.Pos_ }
func (e *AGKEVerifyExpr) exprNode()     {}

// CorruptGPUDigestStmt represents:
// corrupt_gpu_digest chain at ↦ index · (test-only statement, requires --test)
type CorruptGPUDigestStmt struct {
	Pos_   Position
	Chain  Expr
	Offset Expr
}

func (s *CorruptGPUDigestStmt) Pos() Position { return s.Pos_ }
func (s *CorruptGPUDigestStmt) stmtNode()     {}
