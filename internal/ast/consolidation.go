package ast

// ConsolidateStmt promotes high-confidence episodic memories into semantic knowledge graphs.
// Grammar: consolidate <episodic> into <semantic> with threshold ↦ <min_confidence> ·
type ConsolidateStmt struct {
	Pos_          Position
	EpisodicMem   Expr // memory_prov source
	SemanticMem   Expr // graph target
	MinConfidence Expr // threshold 0.0 - 1.0
}

func (s *ConsolidateStmt) Pos() Position { return s.Pos_ }
func (s *ConsolidateStmt) stmtNode()     {}
