package ast

// QubitDecl declares one or more qubit variables: qubit q0 ⸴ q1 ·
type QubitDecl struct {
	Pos_  Position
	Names []string
}

func (d *QubitDecl) Pos() Position { return d.Pos_ }
func (d *QubitDecl) stmtNode()     {}
func (d *QubitDecl) declNode()     {}

// GateStmt applies a single-qubit gate: ℋ q0 ·  𝒳 q0 ·  𝒵 q0 ·  𝕊 q0 ·  𝕋 q0 ·
type GateStmt struct {
	Pos_  Position
	Gate  string // "H","X","Y","Z","S","T"
	Qubit Expr
}

func (s *GateStmt) Pos() Position { return s.Pos_ }
func (s *GateStmt) stmtNode()     {}

// CNOTStmt applies a two-qubit CNOT gate: ℂ q0 ⸴ q1 ·
type CNOTStmt struct {
	Pos_    Position
	Control Expr
	Target  Expr
}

func (s *CNOTStmt) Pos() Position { return s.Pos_ }
func (s *CNOTStmt) stmtNode()     {}

// MeasureExpr evaluates to int64 (0 or 1): measure q0
type MeasureExpr struct {
	Pos_  Position
	Qubit Expr
}

func (e *MeasureExpr) Pos() Position { return e.Pos_ }
func (e *MeasureExpr) exprNode()     {}

// CollapseExpr evaluates to int64 (0 or 1), expression form: collapse q0
type CollapseExpr struct {
	Pos_  Position
	Qubit Expr
}

func (e *CollapseExpr) Pos() Position { return e.Pos_ }
func (e *CollapseExpr) exprNode()     {}
