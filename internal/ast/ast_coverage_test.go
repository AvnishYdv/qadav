package ast

import (
	"testing"
)

func TestASTAllNodesCoverage(t *testing.T) {
	pos := Position{File: "test.ydv", Line: 1, Col: 1}

	// Expressions
	exprs := []Expr{
		&IntLit{Pos_: pos, Value: 1},
		&FloatLit{Pos_: pos, Value: 1.0},
		&StringLit{Pos_: pos, Value: "s"},
		&Ident{Pos_: pos, Name: "x"},
		&BinaryExpr{Pos_: pos},
		&UnaryExpr{Pos_: pos},
		&CallExpr{Pos_: pos},
		&IndexExpr{Pos_: pos},
		&MemberExpr{Pos_: pos},
		&RangeExpr{Pos_: pos},
		&GroupExpr{Pos_: pos},
		&TensorLit{Pos_: pos},
		&QuasiLit{Pos_: pos},
		&ConceptLit{Pos_: pos},
		&DecomposeExpr{Pos_: pos},
		&CollapseExpr{Pos_: pos},
		&ConfidentLit{Pos_: pos},
		&RecallExpr{Pos_: pos},
		&TraceExpr{Pos_: pos},
		&ExecuteExpr{Pos_: pos},
	}
	for _, e := range exprs {
		if e.Pos() != pos {
			t.Errorf("expected pos %v, got %v", pos, e.Pos())
		}
		e.exprNode()
	}

	// Statements
	stmts := []Stmt{
		&BindStmt{Pos_: pos},
		&AssignStmt{Pos_: pos},
		&EmitStmt{Pos_: pos},
		&ObserveStmt{Pos_: pos},
		&YieldStmt{Pos_: pos},
		&WhenStmt{Pos_: pos},
		&CycleStmt{Pos_: pos},
		&IterateStmt{Pos_: pos},
		&HaltStmt{Pos_: pos},
		&SkipStmt{Pos_: pos},
		&ExprStmt{Pos_: pos},
		&SuperposeStmt{Pos_: pos},
		&EntangleStmt{Pos_: pos},
		&CollapseStmt{Pos_: pos},
		&QubitDecl{Pos_: pos},
		&DatasetDecl{Pos_: pos},
		&NormalizeStmt{Pos_: pos},
		&ShuffleStmt{Pos_: pos},
		&BatchStmt{Pos_: pos},
		&EnergyDecl{Pos_: pos},
		&MinimizeStmt{Pos_: pos},
		&TileStmt{Pos_: pos},
		&InferStmt{Pos_: pos},
		&AssertStmt{Pos_: pos},
		&PropagateStmt{Pos_: pos},
		&MemoryDecl{Pos_: pos},
		&RememberStmt{Pos_: pos},
		&GraphDecl{Pos_: pos},
		&NodeStmt{Pos_: pos},
		&LinkStmt{Pos_: pos},
		&ExecuteStmt{Pos_: pos},
		&ProcedureDecl{Pos_: pos},
		&Block{Pos_: pos},
	}
	for _, s := range stmts {
		if s.Pos() != pos {
			t.Errorf("expected pos %v, got %v", pos, s.Pos())
		}
		s.stmtNode()
	}

	// Decls
	decls := []Decl{
		&ProcessDecl{Pos_: pos},
		&NetworkDecl{Pos_: pos},
		&LearnDecl{Pos_: pos},
		&ProcedureDecl{Pos_: pos},
		&AccessDecl{Pos_: pos},
		&ExposeDecl{Pos_: pos},
	}
	for _, d := range decls {
		if d.Pos() != pos {
			t.Errorf("expected pos %v, got %v", pos, d.Pos())
		}
		d.declNode()
	}

	// Types
	typesList := []Type{
		&PrimitiveType{Pos_: pos},
		&TensorType{Pos_: pos},
		&NamedType{Pos_: pos},
	}
	for _, ty := range typesList {
		if ty.Pos() != pos {
			t.Errorf("expected pos %v, got %v", pos, ty.Pos())
		}
		ty.typeNode()
	}

	// Position String
	_ = pos.String()
	p := &Program{Pos_: pos}
	_ = p.Pos()
}

type testVisitor struct{}

func (v *testVisitor) Visit(node Node) {}

func TestASTWalkCoverage(t *testing.T) {
	v := &testVisitor{}
	pos := Position{File: "test.ydv", Line: 1, Col: 1}

	prog := &Program{
		Pos_: pos,
		Decls: []Decl{
			&ProcessDecl{
				Pos_: pos,
				Params: []*Param{
					{Pos_: pos, Type: &PrimitiveType{Pos_: pos}},
				},
				Returns: &PrimitiveType{Pos_: pos},
				Body: &Block{
					Pos_: pos,
					Statements: []Stmt{
						&BindStmt{Pos_: pos, Type: &PrimitiveType{Pos_: pos}, Value: &IntLit{Pos_: pos}},
						&AssignStmt{Pos_: pos, LHS: &Ident{Pos_: pos}, RHS: &FloatLit{Pos_: pos}},
						&EmitStmt{Pos_: pos, Value: &StringLit{Pos_: pos}},
						&ObserveStmt{Pos_: pos, Value: &Ident{Pos_: pos}},
						&YieldStmt{Pos_: pos, Value: &BinaryExpr{Pos_: pos, Left: &Ident{Pos_: pos}, Right: &IntLit{Pos_: pos}}},
						&WhenStmt{Pos_: pos, Condition: &Ident{Pos_: pos}, Then: &Block{Pos_: pos}, Else: &Block{Pos_: pos}},
						&IterateStmt{Pos_: pos, Collection: &RangeExpr{Pos_: pos, Start: &IntLit{Pos_: pos}, End: &IntLit{Pos_: pos}}, Body: &Block{Pos_: pos}},
						&CycleStmt{Pos_: pos, Condition: &Ident{Pos_: pos}, Body: &Block{Pos_: pos}},
						&SuperposeStmt{Pos_: pos, Qubit: &Ident{Pos_: pos}},
						&EntangleStmt{Pos_: pos, Q1: &Ident{Pos_: pos}, Q2: &Ident{Pos_: pos}},
						&CollapseStmt{Pos_: pos, Qubit: &Ident{Pos_: pos}, Into: &Ident{Pos_: pos}},
						&ExprStmt{Pos_: pos, Expr: &UnaryExpr{Pos_: pos, Right: &GroupExpr{Pos_: pos, Inner: &TensorLit{Pos_: pos, Items: []Expr{&IntLit{Pos_: pos}}}}}},
						&ExprStmt{Pos_: pos, Expr: &CallExpr{Pos_: pos, Func: &Ident{Pos_: pos}, Args: []Expr{&Ident{Pos_: pos}}}},
						&ExprStmt{Pos_: pos, Expr: &MemberExpr{Pos_: pos, Object: &Ident{Pos_: pos}}},
						&ExprStmt{Pos_: pos, Expr: &IndexExpr{Pos_: pos, Object: &Ident{Pos_: pos}, Index: &IntLit{Pos_: pos}}},
						&HaltStmt{Pos_: pos},
						&SkipStmt{Pos_: pos},
					},
				},
			},
			&AccessDecl{Pos_: pos},
			&ExposeDecl{Pos_: pos},
		},
	}

	Walk(v, prog)
	Walk(v, &TensorType{Pos_: pos, Elem: &PrimitiveType{Pos_: pos}})
	Walk(v, &NamedType{Pos_: pos})
	Walk(v, nil)
}
