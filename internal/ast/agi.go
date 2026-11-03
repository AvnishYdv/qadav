package ast

// MemoryDecl declares an episodic memory buffer: memory m ⟬ 100 ⟭ ·
type MemoryDecl struct {
	Pos_     Position
	Name     string
	Capacity Expr // max entries
}

func (s *MemoryDecl) Pos() Position { return s.Pos_ }
func (s *MemoryDecl) stmtNode()     {}

// RememberStmt stores an episodic memory entry: remember m⟬key⟭ ↦ val ·
type RememberStmt struct {
	Pos_   Position
	Memory Expr
	Key    Expr
	Value  Expr
}

func (s *RememberStmt) Pos() Position { return s.Pos_ }
func (s *RememberStmt) stmtNode()     {}

// RecallExpr retrieves an episodic memory entry: recall m⟬key⟭
type RecallExpr struct {
	Pos_   Position
	Memory Expr
	Key    Expr
}

func (e *RecallExpr) Pos() Position { return e.Pos_ }
func (e *RecallExpr) exprNode()     {}

// GraphDecl declares a semantic association graph: graph g ·
type GraphDecl struct {
	Pos_ Position
	Name string
}

func (s *GraphDecl) Pos() Position { return s.Pos_ }
func (s *GraphDecl) stmtNode()     {}

// NodeStmt declares a node in a semantic graph: node g ⟬ id ↦ data ⟭ ·
type NodeStmt struct {
	Pos_  Position
	Graph Expr
	ID    Expr
	Data  Expr // optional
}

func (s *NodeStmt) Pos() Position { return s.Pos_ }
func (s *NodeStmt) stmtNode()     {}

// LinkStmt creates an edge in a semantic graph: link g⟬from⸴ to⸴ weight⟭ ·
type LinkStmt struct {
	Pos_   Position
	Graph  Expr
	From   Expr
	To     Expr
	Weight Expr
}

func (s *LinkStmt) Pos() Position { return s.Pos_ }
func (s *LinkStmt) stmtNode()     {}

// TraceExpr performs BFS on a semantic graph: trace g⟬start⸴ depth⟭
type TraceExpr struct {
	Pos_  Position
	Graph Expr
	Start Expr
	Depth Expr
}

func (e *TraceExpr) Pos() Position { return e.Pos_ }
func (e *TraceExpr) exprNode()     {}

// ProcedureDecl declares a reusable executable procedure: procedure p⦅...⦆ → T ⟦ ... ⟧
type ProcedureDecl struct {
	Pos_    Position
	Name    string
	Params  []*Param
	RetType Type
	Body    *Block
}

func (s *ProcedureDecl) Pos() Position { return s.Pos_ }
func (s *ProcedureDecl) declNode()     {}
func (s *ProcedureDecl) stmtNode()     {}

// ExecuteExpr invokes an executable procedure: execute p⦅args⦆
type ExecuteExpr struct {
	Pos_      Position
	Procedure Expr
	Args      []Expr
}

func (e *ExecuteExpr) Pos() Position { return e.Pos_ }
func (e *ExecuteExpr) exprNode()     {}

// ExecuteStmt wraps procedure execution as a standalone statement
type ExecuteStmt struct {
	Pos_      Position
	Procedure Expr
	Args      []Expr
}

func (s *ExecuteStmt) Pos() Position { return s.Pos_ }
func (s *ExecuteStmt) stmtNode()     {}
