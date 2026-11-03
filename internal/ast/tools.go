package ast

// ToolRegisterDecl represents:
// tool_register name with boundary pre ⟦ ... ⟧ post ⟦ ... ⟧ ·
type ToolRegisterDecl struct {
	Pos_          Position
	Name          string
	Boundary      Expr
	PreCondition  Expr
	PostCondition Expr
}

func (s *ToolRegisterDecl) Pos() Position { return s.Pos_ }
func (s *ToolRegisterDecl) stmtNode()     {}
func (s *ToolRegisterDecl) declNode()     {}

// ToolCallExpr represents:
// tool_call name ⦅ args ⦆ audit chain
type ToolCallExpr struct {
	Pos_     Position
	ToolName string
	Args     []Expr
	Audit    Expr // QERA reasoning chain
}

func (s *ToolCallExpr) Pos() Position { return s.Pos_ }
func (s *ToolCallExpr) exprNode()     {}

// ToolCallStmt represents:
// tool_call name ⦅ args ⦆ audit chain ·
// or: res ↦ tool_call name ⦅ args ⦆ audit chain ·
type ToolCallStmt struct {
	Pos_     Position
	Target   Expr // optional target identifier
	ToolName string
	Args     []Expr
	Audit    Expr // QERA reasoning chain
}

func (s *ToolCallStmt) Pos() Position { return s.Pos_ }
func (s *ToolCallStmt) stmtNode()     {}
