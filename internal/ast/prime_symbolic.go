package ast

// ConceptLit represents a concept literal value: concept c ↦ 613 ·
// The Value is the integer encoding (product of primes).
type ConceptLit struct {
	Pos_  Position
	Value int64
}

func (e *ConceptLit) Pos() Position { return e.Pos_ }
func (e *ConceptLit) exprNode()     {}

// DecomposeExpr factorizes a concept into its prime components: decompose⦅expr⦆
type DecomposeExpr struct {
	Pos_  Position
	Input Expr
}

func (e *DecomposeExpr) Pos() Position { return e.Pos_ }
func (e *DecomposeExpr) exprNode()     {}
