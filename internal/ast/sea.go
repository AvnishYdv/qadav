package ast

// EvolveStmt represents:
// evolve from state_v1 to state_v2 with reason reason into proposal ·
type EvolveStmt struct {
	Pos_      Position
	FromState Expr
	ToState   Expr
	Reason    Expr
	Into      string
}

func (s *EvolveStmt) Pos() Position { return s.Pos_ }
func (s *EvolveStmt) stmtNode()     {}

// EvolveVerifyStmt represents:
// evolve_verify proposal into is_valid ·
type EvolveVerifyStmt struct {
	Pos_      Position
	Proposals string
	Into      string
}

func (s *EvolveVerifyStmt) Pos() Position { return s.Pos_ }
func (s *EvolveVerifyStmt) stmtNode()     {}

// EvolveHistoryStmt represents:
// evolve_history to evo_log ·
type EvolveHistoryStmt struct {
	Pos_  Position
	Chain string
}

func (s *EvolveHistoryStmt) Pos() Position { return s.Pos_ }
func (s *EvolveHistoryStmt) stmtNode()     {}

// CorruptEvolveStmt represents:
// corrupt_evolve p at ↦ 16 ·
type CorruptEvolveStmt struct {
	Pos_     Position
	Proposal string
	Offset   Expr
}

func (s *CorruptEvolveStmt) Pos() Position { return s.Pos_ }
func (s *CorruptEvolveStmt) stmtNode()     {}
