package ast

// ProvenanceDecl declares a training provenance tracker for a network:
// provenance p ⟬ ProvenanceModel ⟭ ·
type ProvenanceDecl struct {
	Pos_       Position
	Name       string
	NetworkRef Expr
}

func (s *ProvenanceDecl) Pos() Position { return s.Pos_ }
func (s *ProvenanceDecl) stmtNode()     {}
func (s *ProvenanceDecl) declNode()     {}

// CertifyStmt finalizes and cryptographically certifies training provenance:
// certify p ·
type CertifyStmt struct {
	Pos_       Position
	Provenance Expr
}

func (s *CertifyStmt) Pos() Position { return s.Pos_ }
func (s *CertifyStmt) stmtNode()     {}

// VerifyTrainingExpr represents verification of a training certificate:
// verify_training ⟬ cert ⟭
type VerifyTrainingExpr struct {
	Pos_        Position
	Certificate Expr
	Model       Expr
}

func (e *VerifyTrainingExpr) Pos() Position { return e.Pos_ }
func (e *VerifyTrainingExpr) exprNode()     {}
