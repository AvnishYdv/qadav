package ast

// CDAGARecordStmt represents:
// cdaga_record from ⟬grad_a⸴ grad_b⟭ into record ·
type CDAGARecordStmt struct {
	Pos_     Position
	GradList []Expr
	Into     string
}

func (s *CDAGARecordStmt) Pos() Position { return s.Pos_ }
func (s *CDAGARecordStmt) stmtNode()     {}

// CDAGAAttestStmt represents:
// cdaga_attest record to chain ·
type CDAGAAttestStmt struct {
	Pos_   Position
	Record string
	Chain  string
}

func (s *CDAGAAttestStmt) Pos() Position { return s.Pos_ }
func (s *CDAGAAttestStmt) stmtNode()     {}

// CDAGAVerifyStmt represents:
// cdaga_verify record into is_valid ·
type CDAGAVerifyStmt struct {
	Pos_   Position
	Record string
	Into   string
}

func (s *CDAGAVerifyStmt) Pos() Position { return s.Pos_ }
func (s *CDAGAVerifyStmt) stmtNode()     {}

// CorruptCDAGAStmt represents:
// corrupt_cdaga record at ↦ offset ·
type CorruptCDAGAStmt struct {
	Pos_   Position
	Record string
	Offset Expr
}

func (s *CorruptCDAGAStmt) Pos() Position { return s.Pos_ }
func (s *CorruptCDAGAStmt) stmtNode()     {}
