package ast

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/token"
)

func TestPositionString(t *testing.T) {
	p := Position{Line: 3, Col: 12, File: "test.ydv"}
	expected := "test.ydv:3:12"
	if p.String() != expected {
		t.Fatalf("Position.String() = %q, expected %q", p.String(), expected)
	}
}

func TestFromToken(t *testing.T) {
	tok := token.Token{
		Type:    token.INT,
		Literal: "42",
		Line:    5,
		Col:     10,
		File:    "x.ydv",
	}
	p := FromToken(tok)
	if p.Line != 5 || p.Col != 10 || p.File != "x.ydv" {
		t.Fatalf("FromToken mismatch: got %+v", p)
	}
}

func TestNodeInterfaces(t *testing.T) {
	pos := Position{Line: 1, Col: 1, File: "test.ydv"}

	// Decls
	var _ Decl = &ProcessDecl{Pos_: pos}
	var _ Decl = &NetworkDecl{Pos_: pos}
	var _ Decl = &LearnDecl{Pos_: pos}
	var _ Decl = &AccessDecl{Pos_: pos}
	var _ Decl = &ExposeDecl{Pos_: pos}

	// Statements
	var _ Stmt = &Block{Pos_: pos}
	var _ Stmt = &BindStmt{Pos_: pos}
	var _ Stmt = &AssignStmt{Pos_: pos}
	var _ Stmt = &EmitStmt{Pos_: pos}
	var _ Stmt = &ObserveStmt{Pos_: pos}
	var _ Stmt = &YieldStmt{Pos_: pos}
	var _ Stmt = &WhenStmt{Pos_: pos}
	var _ Stmt = &IterateStmt{Pos_: pos}
	var _ Stmt = &CycleStmt{Pos_: pos}
	var _ Stmt = &SuperposeStmt{Pos_: pos}
	var _ Stmt = &EntangleStmt{Pos_: pos}
	var _ Stmt = &CollapseStmt{Pos_: pos}
	var _ Stmt = &HaltStmt{Pos_: pos}
	var _ Stmt = &SkipStmt{Pos_: pos}
	var _ Stmt = &ExprStmt{Pos_: pos}

	// Expressions
	var _ Expr = &Ident{Pos_: pos}
	var _ Expr = &IntLit{Pos_: pos}
	var _ Expr = &FloatLit{Pos_: pos}
	var _ Expr = &StringLit{Pos_: pos}
	var _ Expr = &BinaryExpr{Pos_: pos}
	var _ Expr = &UnaryExpr{Pos_: pos}
	var _ Expr = &CallExpr{Pos_: pos}
	var _ Expr = &MemberExpr{Pos_: pos}
	var _ Expr = &IndexExpr{Pos_: pos}
	var _ Expr = &GroupExpr{Pos_: pos}
	var _ Expr = &RangeExpr{Pos_: pos}

	// Types
	var _ Type = &PrimitiveType{Pos_: pos}
	var _ Type = &TensorType{Pos_: pos}
	var _ Type = &NamedType{Pos_: pos}

	// Root
	var _ Node = &Program{Pos_: pos}
}

func TestAllNodesHavePos(t *testing.T) {
	pos := Position{Line: 42, Col: 7, File: "foo.ydv"}

	nodes := []Node{
		&Program{Pos_: pos},
		&ProcessDecl{Pos_: pos},
		&NetworkDecl{Pos_: pos},
		&LearnDecl{Pos_: pos},
		&AccessDecl{Pos_: pos},
		&ExposeDecl{Pos_: pos},
		&Param{Pos_: pos},
		&Block{Pos_: pos},
		&BindStmt{Pos_: pos},
		&AssignStmt{Pos_: pos},
		&EmitStmt{Pos_: pos},
		&ObserveStmt{Pos_: pos},
		&YieldStmt{Pos_: pos},
		&WhenStmt{Pos_: pos},
		&IterateStmt{Pos_: pos},
		&CycleStmt{Pos_: pos},
		&SuperposeStmt{Pos_: pos},
		&EntangleStmt{Pos_: pos},
		&CollapseStmt{Pos_: pos},
		&HaltStmt{Pos_: pos},
		&SkipStmt{Pos_: pos},
		&ExprStmt{Pos_: pos},
		&Ident{Pos_: pos},
		&IntLit{Pos_: pos},
		&FloatLit{Pos_: pos},
		&StringLit{Pos_: pos},
		&BinaryExpr{Pos_: pos},
		&UnaryExpr{Pos_: pos},
		&CallExpr{Pos_: pos},
		&MemberExpr{Pos_: pos},
		&IndexExpr{Pos_: pos},
		&GroupExpr{Pos_: pos},
		&RangeExpr{Pos_: pos},
		&PrimitiveType{Pos_: pos},
		&TensorType{Pos_: pos},
		&NamedType{Pos_: pos},
	}

	for i, n := range nodes {
		if n.Pos() != pos {
			t.Fatalf("[%d] Node %T returned pos %+v, expected %+v", i, n, n.Pos(), pos)
		}
	}
}

type countVisitor struct {
	visited int
}

func (c *countVisitor) Visit(node Node) {
	if node != nil {
		c.visited++
	}
}

func TestWalk(t *testing.T) {
	pos := Position{Line: 1, Col: 1, File: "main.ydv"}
	prog := &Program{
		Pos_: pos,
		Decls: []Decl{
			&ProcessDecl{
				Pos_: pos,
				Name: "main",
				Params: []*Param{
					{Pos_: pos, Name: "arg", Type: &PrimitiveType{Pos_: pos, Name: "int32"}},
				},
				Returns: &PrimitiveType{Pos_: pos, Name: "int32"},
				Body: &Block{
					Pos_: pos,
					Statements: []Stmt{
						&BindStmt{
							Pos_:  pos,
							Type:  &PrimitiveType{Pos_: pos, Name: "int32"},
							Name:  "x",
							Value: &IntLit{Pos_: pos, Value: 10},
						},
						&EmitStmt{
							Pos_:  pos,
							Value: &Ident{Pos_: pos, Name: "x"},
						},
						&YieldStmt{
							Pos_:  pos,
							Value: &IntLit{Pos_: pos, Value: 0},
						},
					},
				},
			},
		},
	}

	v := &countVisitor{}
	Walk(v, prog)

	if v.visited < 8 {
		t.Fatalf("expected at least 8 visited nodes, got %d", v.visited)
	}
}
