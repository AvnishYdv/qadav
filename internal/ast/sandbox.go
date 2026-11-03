package ast

type SandboxCreateStmt struct {
	Pos_     Position
	Name     string
	Boundary string
	Capacity Expr
}

func (s *SandboxCreateStmt) Pos() Position { return s.Pos_ }
func (s *SandboxCreateStmt) stmtNode()     {}

type SandboxRunStmt struct {
	Pos_     Position
	Sandbox  string
	ToolName string
	Args     []Expr
}

func (s *SandboxRunStmt) Pos() Position { return s.Pos_ }
func (s *SandboxRunStmt) stmtNode()     {}

type SandboxVerifyStmt struct {
	Pos_    Position
	Sandbox string
	Into    string
}

func (s *SandboxVerifyStmt) Pos() Position { return s.Pos_ }
func (s *SandboxVerifyStmt) stmtNode()     {}

type SandboxVerifyExpr struct {
	Pos_    Position
	Sandbox string
}

func (s *SandboxVerifyExpr) Pos() Position { return s.Pos_ }
func (s *SandboxVerifyExpr) exprNode()     {}
