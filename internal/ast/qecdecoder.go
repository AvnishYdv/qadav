package ast

// QECDDecodeStmt represents decoding a syndrome via greedy min-weight QEC decoder:
// qecd_decode distance ↦ 3 from syn0 into decoded ·
type QECDDecodeStmt struct {
	Pos_       Position
	Distance   Expr
	SyndromeIn string
	Into       string
}

func (s *QECDDecodeStmt) Pos() Position { return s.Pos_ }
func (s *QECDDecodeStmt) stmtNode()     {}

// QECDVerifyStmt represents verifying the SHA3 decode attestation of a QEC decode result:
// qecd_verify decoded into is_valid ·
type QECDVerifyStmt struct {
	Pos_   Position
	Result string
	Into   string
}

func (s *QECDVerifyStmt) Pos() Position { return s.Pos_ }
func (s *QECDVerifyStmt) stmtNode()     {}

// QECDApplyStmt represents applying decoded Pauli corrections to a quantum state vector:
// qecd_apply decoded on state_vec ·
type QECDApplyStmt struct {
	Pos_     Position
	Result   string
	StateVec string
}

func (s *QECDApplyStmt) Pos() Position { return s.Pos_ }
func (s *QECDApplyStmt) stmtNode()     {}
