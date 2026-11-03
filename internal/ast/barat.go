package ast

// BaratSubmitStmt represents:
// barat_submit worker into b_grp with grad ·
// OR: barat_submit from grad into b_grp ·
type BaratSubmitStmt struct {
	Pos_     Position
	Worker   Expr
	Group    string
	Gradient Expr
}

func (s *BaratSubmitStmt) Pos() Position { return s.Pos_ }
func (s *BaratSubmitStmt) stmtNode()     {}

// BaratAggregateStmt represents:
// barat_aggregate b_grp into agg_grad cert into b_cert ·
type BaratAggregateStmt struct {
	Pos_     Position
	Group    string
	IntoGrad string
	IntoCert string
}

func (s *BaratAggregateStmt) Pos() Position { return s.Pos_ }
func (s *BaratAggregateStmt) stmtNode()     {}

// BaratVerifyStmt represents:
// barat_verify b_cert into is_valid ·
type BaratVerifyStmt struct {
	Pos_ Position
	Cert string
	Into string
}

func (s *BaratVerifyStmt) Pos() Position { return s.Pos_ }
func (s *BaratVerifyStmt) stmtNode()     {}

// CorruptBaratStmt represents:
// corrupt_barat b_cert at ↦ offset ·
type CorruptBaratStmt struct {
	Pos_   Position
	Cert   string
	Offset Expr
}

func (s *CorruptBaratStmt) Pos() Position { return s.Pos_ }
func (s *CorruptBaratStmt) stmtNode()     {}
