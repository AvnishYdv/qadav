package ast

// RecallTokenDecl represents:
// recall_token tok for mem on recipient ⟦ keys ↦ ⟬ "k1" ⸴ "k2" ⟭ · expiry ↦ 3600 · ⟧
type RecallTokenDecl struct {
	Pos_      Position
	Name      string
	Memory    Expr
	Recipient Expr
	Keys      []Expr
	Expiry    Expr
}

func (s *RecallTokenDecl) Pos() Position { return s.Pos_ }
func (s *RecallTokenDecl) stmtNode()     {}
func (s *RecallTokenDecl) declNode()     {}

// GrantStmt represents:
// grant tok to recipient ·
type GrantStmt struct {
	Pos_      Position
	Token     Expr
	Recipient Expr
}

func (s *GrantStmt) Pos() Position { return s.Pos_ }
func (s *GrantStmt) stmtNode()     {}

// RevokeStmt represents:
// revoke tok ·
type RevokeStmt struct {
	Pos_  Position
	Token Expr
}

func (s *RevokeStmt) Pos() Position { return s.Pos_ }
func (s *RevokeStmt) stmtNode()     {}

// ScopedRecallExpr represents:
// recall_scoped mem⟬"key"⟭ with tok
type ScopedRecallExpr struct {
	Pos_   Position
	Memory Expr
	Key    Expr
	Token  Expr
}

func (s *ScopedRecallExpr) Pos() Position { return s.Pos_ }
func (s *ScopedRecallExpr) exprNode()     {}
