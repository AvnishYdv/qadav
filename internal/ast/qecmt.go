package ast

// LicenseDecl represents:
// license l for teacher_vault on StudentNet ⟦ max_epochs ↦ 50 · max_rate ↦ 0.5 · expiry ↦ 1735689600 · ⟧
type LicenseDecl struct {
	Pos_      Position
	Name      string
	Teacher   Expr // QEMWV vault reference
	Student   Expr // Target network
	MaxEpochs Expr
	MaxRate   Expr // Max distillation rate (0.0-1.0)
	Expiry    Expr // Unix timestamp
}

func (s *LicenseDecl) Pos() Position { return s.Pos_ }
func (s *LicenseDecl) stmtNode()     {}
func (s *LicenseDecl) declNode()     {}

// TransferStmt represents:
// transfer l from teacher_vault to StudentNet layers ⟬0⸴ 1⟭ ·
type TransferStmt struct {
	Pos_    Position
	License Expr
	Teacher Expr
	Student Expr
	Layers  []Expr // specific layer indices
}

func (s *TransferStmt) Pos() Position { return s.Pos_ }
func (s *TransferStmt) stmtNode()     {}

// DistillStmt represents:
// distill l on StudentNet with x ⸴ y ⟦ rate ↦ 0.3 · epochs ↦ 20 · ⟧
type DistillStmt struct {
	Pos_     Position
	License  Expr
	Student  Expr
	DatasetX Expr
	DatasetY Expr
	Rate     Expr
	Epochs   Expr
}

func (s *DistillStmt) Pos() Position { return s.Pos_ }
func (s *DistillStmt) stmtNode()     {}
