package ast

// RSCAAttestStmt represents a cryptographic self-compilation attestation step:
// rsca_attest stage ↦ 1 from source ↦ source_bytes compiler ↦ stage0_bin output ↦ stage1_bin into rec1 ·
type RSCAAttestStmt struct {
	Pos_     Position
	Stage    int
	Source   string
	Compiler string
	Output   string
	Into     string
}

func (s *RSCAAttestStmt) Pos() Position { return s.Pos_ }
func (s *RSCAAttestStmt) stmtNode()     {}

// RSCAVerifyStmt represents verification of an RSCA stage record:
// rsca_verify rec1 into v1 ·
type RSCAVerifyStmt struct {
	Pos_   Position
	Record string
	Into   string
}

func (s *RSCAVerifyStmt) Pos() Position { return s.Pos_ }
func (s *RSCAVerifyStmt) stmtNode()     {}

// CorruptRSCAStmt represents a test-only mutation of an RSCA record:
// corrupt_rsca rec2 at ↦ 16 ·
type CorruptRSCAStmt struct {
	Pos_   Position
	Record string
	Offset Expr
}

func (s *CorruptRSCAStmt) Pos() Position { return s.Pos_ }
func (s *CorruptRSCAStmt) stmtNode()     {}
