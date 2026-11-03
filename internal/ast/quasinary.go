package ast

// QuasiLit represents a balanced ternary literal: ◄ (-1), 〇 (0), or ► (+1).
type QuasiLit struct {
	Pos_  Position
	Value int8 // -1, 0, +1
}

func (q *QuasiLit) Pos() Position { return q.Pos_ }
func (q *QuasiLit) exprNode()     {}

// EnergyDecl represents: energy Name ⟦ ... ⟧
type EnergyDecl struct {
	Pos_ Position
	Name string
	Fn   *Block
}

func (e *EnergyDecl) Pos() Position { return e.Pos_ }
func (e *EnergyDecl) stmtNode()     {}

// MinimizeStmt represents: minimize grid energy ising iterations ↦ 1000 ·
type MinimizeStmt struct {
	Pos_       Position
	Grid       Expr
	Energy     Expr // e.g. ising
	Iterations Expr
}

func (m *MinimizeStmt) Pos() Position { return m.Pos_ }
func (m *MinimizeStmt) stmtNode()     {}

// TileStmt represents: tile grid pattern ·
type TileStmt struct {
	Pos_    Position
	Grid    Expr
	Pattern Expr
}

func (t *TileStmt) Pos() Position { return t.Pos_ }
func (t *TileStmt) stmtNode()     {}
