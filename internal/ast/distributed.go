package ast

// AllReduceStmt represents:
// all_reduce from ⟬grad_a⸴ grad_b⟭ into result ·
// OR: all_reduce grad_buf ·
type AllReduceStmt struct {
	Pos_     Position
	GradList []Expr
	Target   Expr
	Into     string
}

func (s *AllReduceStmt) Pos() Position { return s.Pos_ }
func (s *AllReduceStmt) stmtNode()     {}

// PserverInitStmt represents:
// pserver_init into ps with 4 workers ·
type PserverInitStmt struct {
	Pos_    Position
	Into    string
	Workers Expr
}

func (s *PserverInitStmt) Pos() Position { return s.Pos_ }
func (s *PserverInitStmt) stmtNode()     {}

// PserverAggregateStmt represents:
// pserver_aggregate ps ·
type PserverAggregateStmt struct {
	Pos_    Position
	Pserver string
}

func (s *PserverAggregateStmt) Pos() Position { return s.Pos_ }
func (s *PserverAggregateStmt) stmtNode()     {}
