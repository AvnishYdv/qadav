package ast

// CTSAAttestStmt represents a compile-time type safety attestation step:
// ctsa_attest ast ↦ a env ↦ e rules ↦ r into cert ·
type CTSAAttestStmt struct {
	Pos_  Position
	Ast   string
	Env   string
	Rules string
	Into  string
}

func (s *CTSAAttestStmt) Pos() Position { return s.Pos_ }
func (s *CTSAAttestStmt) stmtNode()     {}

// CTSAVerifyStmt represents verification of a CTSA certificate:
// ctsa_verify cert into res ·
type CTSAVerifyStmt struct {
	Pos_ Position
	Cert string
	Into string
}

func (s *CTSAVerifyStmt) Pos() Position { return s.Pos_ }
func (s *CTSAVerifyStmt) stmtNode()     {}

// CorruptCTSAStmt represents a test-only corruption of a CTSA certificate field:
// corrupt_ctsa cert field ↦ 1 ·
type CorruptCTSAStmt struct {
	Pos_  Position
	Cert  string
	Field Expr
}

func (s *CorruptCTSAStmt) Pos() Position { return s.Pos_ }
func (s *CorruptCTSAStmt) stmtNode()     {}
