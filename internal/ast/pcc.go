package ast

// PCCAttestStmt represents a proof-carrying codegen attestation step:
// pcc_attest ast ↦ a codegen ↦ c ir ↦ i into cert ·
type PCCAttestStmt struct {
	Pos_    Position
	Ast     string
	Codegen string
	IR      string
	Into    string
}

func (s *PCCAttestStmt) Pos() Position { return s.Pos_ }
func (s *PCCAttestStmt) stmtNode()     {}

// PCCVerifyStmt represents verification of a PCC certificate:
// pcc_verify cert into res ·
type PCCVerifyStmt struct {
	Pos_ Position
	Cert string
	Into string
}

func (s *PCCVerifyStmt) Pos() Position { return s.Pos_ }
func (s *PCCVerifyStmt) stmtNode()     {}

// CorruptPCCStmt represents a test-only corruption of a PCC certificate:
// corrupt_pcc cert field ↦ 1 ·
type CorruptPCCStmt struct {
	Pos_  Position
	Cert  string
	Field Expr
}

func (s *CorruptPCCStmt) Pos() Position { return s.Pos_ }
func (s *CorruptPCCStmt) stmtNode()     {}
