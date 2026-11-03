package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func parseExpr(t *testing.T, input string) ast.Expr {
	t.Helper()
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	expr := p.parseExpression(LOWEST)
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors for %q: %v", input, p.Errors())
	}
	return expr
}

func parseStmt(t *testing.T, input string) ast.Stmt {
	t.Helper()
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	stmt := p.parseStatement()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors for %q: %v", input, p.Errors())
	}
	return stmt
}

func TestParseIntLit(t *testing.T) {
	cases := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"0xFF", 255},
		{"0b1010", 10},
		{"1_000_000", 1000000},
	}
	for _, c := range cases {
		expr := parseExpr(t, c.input)
		intLit, ok := expr.(*ast.IntLit)
		if !ok || intLit.Value != c.expected {
			t.Fatalf("expected IntLit %d for %q, got %+v", c.expected, c.input, expr)
		}
	}
}

func TestParseFloatLit(t *testing.T) {
	cases := []struct {
		input     string
		expected  float64
		imaginary bool
	}{
		{"3.14", 3.14, false},
		{"1e10", 1e10, false},
		{"3.14i", 3.14, true},
	}
	for _, c := range cases {
		expr := parseExpr(t, c.input)
		flt, ok := expr.(*ast.FloatLit)
		if !ok || flt.Value != c.expected || flt.Imaginary != c.imaginary {
			t.Fatalf("expected FloatLit (%v, %v) for %q, got %+v", c.expected, c.imaginary, c.input, expr)
		}
	}
}

func TestParseStringLit(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"multi\nline"`, `multi\nline`},
		{`"""raw\nlines"""`, `raw\nlines`},
	}
	for _, c := range cases {
		expr := parseExpr(t, c.input)
		strLit, ok := expr.(*ast.StringLit)
		if !ok || strLit.Value != c.expected {
			t.Fatalf("expected StringLit %q for %s, got %+v", c.expected, c.input, expr)
		}
	}
}

func TestParseIdent(t *testing.T) {
	cases := []string{"x", "foo", "my_var"}
	for _, c := range cases {
		expr := parseExpr(t, c)
		id, ok := expr.(*ast.Ident)
		if !ok || id.Name != c {
			t.Fatalf("expected Ident %q, got %+v", c, expr)
		}
	}
}

func TestParseBinaryExpr(t *testing.T) {
	expr := parseExpr(t, "1 ⊕ 2")
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.OSUM {
		t.Fatalf("expected BinaryExpr with OSUM, got %+v", expr)
	}

	// Precedence: 1 ⊕ 2 ⊗ 3 -> 1 ⊕ (2 ⊗ 3)
	exprPrec := parseExpr(t, "1 ⊕ 2 ⊗ 3")
	binPrec, ok := exprPrec.(*ast.BinaryExpr)
	if !ok || binPrec.Op != token.OSUM {
		t.Fatalf("expected outer op OSUM, got %+v", exprPrec)
	}
	rightBin, ok := binPrec.Right.(*ast.BinaryExpr)
	if !ok || rightBin.Op != token.OTIM {
		t.Fatalf("expected inner op OTIM, got %+v", binPrec.Right)
	}
}

func TestParsePrecedence(t *testing.T) {
	// 1 ⊕ 2 ⊗ 3 ⊕ 4 -> ((1 ⊕ (2 ⊗ 3)) ⊕ 4)
	expr := parseExpr(t, "1 ⊕ 2 ⊗ 3 ⊕ 4")
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.OSUM {
		t.Fatalf("expected top-level op OSUM, got %+v", expr)
	}
	rLit, ok := bin.Right.(*ast.IntLit)
	if !ok || rLit.Value != 4 {
		t.Fatalf("expected rightmost literal 4, got %+v", bin.Right)
	}
	lBin, ok := bin.Left.(*ast.BinaryExpr)
	if !ok || lBin.Op != token.OSUM {
		t.Fatalf("expected left op OSUM, got %+v", bin.Left)
	}
}

func TestParseUnaryExpr(t *testing.T) {
	cases := []struct {
		input string
		op    token.Type
	}{
		{"¬x", token.NEG},
		{"⊖5", token.ODIFF},
		{"⋆M", token.ADJOINT},
	}
	for _, c := range cases {
		expr := parseExpr(t, c.input)
		un, ok := expr.(*ast.UnaryExpr)
		if !ok || un.Op != c.op {
			t.Fatalf("expected UnaryExpr with op %v for %q, got %+v", c.op, c.input, expr)
		}
	}
}

func TestParseCallExpr(t *testing.T) {
	// f⦅x⸴y⦆
	expr := parseExpr(t, "f⦅x⸴ y⦆")
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		t.Fatalf("expected CallExpr with 2 args, got %+v", expr)
	}

	// f⦅⦆
	emptyCall := parseExpr(t, "f⦅⦆")
	call2, ok := emptyCall.(*ast.CallExpr)
	if !ok || len(call2.Args) != 0 {
		t.Fatalf("expected CallExpr with 0 args, got %+v", emptyCall)
	}

	// f⦅g⦅x⦆⦆
	nestCall := parseExpr(t, "f⦅g⦅x⦆⦆")
	call3, ok := nestCall.(*ast.CallExpr)
	if !ok || len(call3.Args) != 1 {
		t.Fatalf("expected CallExpr with 1 arg, got %+v", nestCall)
	}
	if _, ok := call3.Args[0].(*ast.CallExpr); !ok {
		t.Fatalf("expected nested CallExpr arg, got %+v", call3.Args[0])
	}
}

func TestParseIndexExpr(t *testing.T) {
	expr := parseExpr(t, "arr⟬0⟭")
	idx, ok := expr.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %+v", expr)
	}
	if lit, ok := idx.Index.(*ast.IntLit); !ok || lit.Value != 0 {
		t.Fatalf("expected index literal 0, got %+v", idx.Index)
	}

	expr2 := parseExpr(t, "arr⟬i ⊕ 1⟭")
	idx2, ok := expr2.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %+v", expr2)
	}
	if _, ok := idx2.Index.(*ast.BinaryExpr); !ok {
		t.Fatalf("expected BinaryExpr inside index, got %+v", idx2.Index)
	}
}

func TestParseMemberExpr(t *testing.T) {
	expr := parseExpr(t, "obj∸field")
	if fa, ok := expr.(*ast.FieldAccessExpr); ok {
		if fa.Field != "field" {
			t.Fatalf("expected FieldAccessExpr with field, got %+v", expr)
		}
	} else if mem, ok := expr.(*ast.MemberExpr); ok {
		if mem.Member != "field" {
			t.Fatalf("expected MemberExpr with field, got %+v", expr)
		}
	} else {
		t.Fatalf("expected FieldAccessExpr or MemberExpr with field, got %+v", expr)
	}

	// a∸b∸c (left-assoc)
	expr2 := parseExpr(t, "a∸b∸c")
	if fa2, ok := expr2.(*ast.FieldAccessExpr); ok {
		if fa2.Field != "c" {
			t.Fatalf("expected outer FieldAccessExpr member 'c', got %+v", expr2)
		}
		innerFa, ok := fa2.Object.(*ast.FieldAccessExpr)
		if !ok || innerFa.Field != "b" {
			t.Fatalf("expected inner FieldAccessExpr member 'b', got %+v", fa2.Object)
		}
	} else if mem2, ok := expr2.(*ast.MemberExpr); ok {
		if mem2.Member != "c" {
			t.Fatalf("expected outer MemberExpr member 'c', got %+v", expr2)
		}
		innerMem, ok := mem2.Object.(*ast.MemberExpr)
		if !ok || innerMem.Member != "b" {
			t.Fatalf("expected inner MemberExpr member 'b', got %+v", mem2.Object)
		}
	}
}

func TestParseGroupExpr(t *testing.T) {
	expr := parseExpr(t, "⦅1 ⊕ 2⦆ ⊗ 3")
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.OTIM {
		t.Fatalf("expected outer op OTIM, got %+v", expr)
	}
	grp, ok := bin.Left.(*ast.GroupExpr)
	if !ok {
		t.Fatalf("expected GroupExpr on left, got %+v", bin.Left)
	}
	if _, ok := grp.Inner.(*ast.BinaryExpr); !ok {
		t.Fatalf("expected inner BinaryExpr, got %+v", grp.Inner)
	}
}

func TestParseBindStmt(t *testing.T) {
	stmt := parseStmt(t, "int32 x ↦ 42·")
	bind, ok := stmt.(*ast.BindStmt)
	if !ok || bind.Name != "x" {
		t.Fatalf("expected BindStmt for 'x', got %+v", stmt)
	}
	if pType, ok := bind.Type.(*ast.PrimitiveType); !ok || pType.Name != "int32" {
		t.Fatalf("expected PrimitiveType int32, got %+v", bind.Type)
	}
	if lit, ok := bind.Value.(*ast.IntLit); !ok || lit.Value != 42 {
		t.Fatalf("expected value 42, got %+v", bind.Value)
	}
}

func TestParseBindStmt_Fallback(t *testing.T) {
	stmt := parseStmt(t, `int32 x \bind 42\stmt`)
	bind, ok := stmt.(*ast.BindStmt)
	if !ok || bind.Name != "x" {
		t.Fatalf("expected BindStmt for 'x' from fallback, got %+v", stmt)
	}
	if lit, ok := bind.Value.(*ast.IntLit); !ok || lit.Value != 42 {
		t.Fatalf("expected value 42, got %+v", bind.Value)
	}
}

func TestParseAssignStmt(t *testing.T) {
	stmt := parseStmt(t, "x ← 42·")
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok || assign.Op != token.ASSIGN {
		t.Fatalf("expected AssignStmt with ASSIGN, got %+v", stmt)
	}
	if id, ok := assign.LHS.(*ast.Ident); !ok || id.Name != "x" {
		t.Fatalf("expected LHS 'x', got %+v", assign.LHS)
	}
}

func TestParseEmitStmt(t *testing.T) {
	stmt := parseStmt(t, "emit x·")
	em, ok := stmt.(*ast.EmitStmt)
	if !ok {
		t.Fatalf("expected EmitStmt, got %+v", stmt)
	}
	if id, ok := em.Value.(*ast.Ident); !ok || id.Name != "x" {
		t.Fatalf("expected EmitStmt with 'x', got %+v", em.Value)
	}

	stmt2 := parseStmt(t, "emit 1 ⊕ 2·")
	em2, ok := stmt2.(*ast.EmitStmt)
	if !ok {
		t.Fatalf("expected EmitStmt, got %+v", stmt2)
	}
	if _, ok := em2.Value.(*ast.BinaryExpr); !ok {
		t.Fatalf("expected BinaryExpr in emit, got %+v", em2.Value)
	}
}

func TestParseYieldStmt(t *testing.T) {
	stmt := parseStmt(t, "yield 0·")
	yd, ok := stmt.(*ast.YieldStmt)
	if !ok || yd.Value == nil {
		t.Fatalf("expected YieldStmt with value, got %+v", stmt)
	}

	stmtBare := parseStmt(t, "yield·")
	ydBare, ok := stmtBare.(*ast.YieldStmt)
	if !ok || ydBare.Value != nil {
		t.Fatalf("expected bare YieldStmt, got %+v", stmtBare)
	}
}

func TestParseWhenStmt(t *testing.T) {
	input1 := "when x ≡ 5 ⟦ emit x· ⟧"
	stmt1 := parseStmt(t, input1)
	w1, ok := stmt1.(*ast.WhenStmt)
	if !ok || w1.Else != nil {
		t.Fatalf("expected WhenStmt without else, got %+v", stmt1)
	}

	input2 := "when x ≡ 5 ⟦ emit x· ⟧ otherwise ⟦ emit 0· ⟧"
	stmt2 := parseStmt(t, input2)
	w2, ok := stmt2.(*ast.WhenStmt)
	if !ok || w2.Else == nil {
		t.Fatalf("expected WhenStmt with else, got %+v", stmt2)
	}
}

func TestParseIterateStmt(t *testing.T) {
	stmt := parseStmt(t, "iterate x ∈ xs ⟦ emit x· ⟧")
	iter, ok := stmt.(*ast.IterateStmt)
	if !ok || iter.Var != "x" {
		t.Fatalf("expected IterateStmt for 'x', got %+v", stmt)
	}
	if id, ok := iter.Collection.(*ast.Ident); !ok || id.Name != "xs" {
		t.Fatalf("expected collection 'xs', got %+v", iter.Collection)
	}
}

func TestParseCycleStmt(t *testing.T) {
	stmt := parseStmt(t, "cycle x ≺ 10 ⟦ x ← x ⊕ 1· ⟧")
	cyc, ok := stmt.(*ast.CycleStmt)
	if !ok || cyc.Condition == nil {
		t.Fatalf("expected CycleStmt with condition, got %+v", stmt)
	}
}

func TestParseSuperposeStmt(t *testing.T) {
	stmt := parseStmt(t, "superpose q0·")
	sup, ok := stmt.(*ast.SuperposeStmt)
	if !ok {
		t.Fatalf("expected SuperposeStmt, got %+v", stmt)
	}
	if id, ok := sup.Qubit.(*ast.Ident); !ok || id.Name != "q0" {
		t.Fatalf("expected qubit 'q0', got %+v", sup.Qubit)
	}
}

func TestParseEntangleStmt(t *testing.T) {
	stmt := parseStmt(t, "entangle q0 q1·")
	ent, ok := stmt.(*ast.EntangleStmt)
	if !ok {
		t.Fatalf("expected EntangleStmt, got %+v", stmt)
	}
	if id1, ok := ent.Q1.(*ast.Ident); !ok || id1.Name != "q0" {
		t.Fatalf("expected Q1 'q0', got %+v", ent.Q1)
	}
	if id2, ok := ent.Q2.(*ast.Ident); !ok || id2.Name != "q1" {
		t.Fatalf("expected Q2 'q1', got %+v", ent.Q2)
	}
}

func TestParseCollapseStmt(t *testing.T) {
	stmt := parseStmt(t, "collapse q0 ↦ bit·")
	col, ok := stmt.(*ast.CollapseStmt)
	if !ok || col.Into == nil {
		t.Fatalf("expected CollapseStmt with Into target, got %+v", stmt)
	}
}

func TestParseProcessDecl(t *testing.T) {
	input := `process foo⦅⦆ → int32 ⟦ yield 0· ⟧`
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok || proc.Name != "foo" {
		t.Fatalf("expected ProcessDecl foo, got %+v", prog.Decls[0])
	}
}

func TestParseProcessDecl_WithParams(t *testing.T) {
	input := `process add⦅int32 a⸴ int32 b⦆ → int32 ⟦ yield a ⊕ b· ⟧`
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}
	if proc.Params[0].Name != "a" || proc.Params[1].Name != "b" {
		t.Fatalf("parameter names mismatch: %+v, %+v", proc.Params[0], proc.Params[1])
	}
}

func TestParseNetworkDecl(t *testing.T) {
	input := `network GPT ⟦ dense 128· ⟧`
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	net, ok := prog.Decls[0].(*ast.NetworkDecl)
	if !ok || net.Name != "GPT" {
		t.Fatalf("expected NetworkDecl GPT, got %+v", prog.Decls[0])
	}
	if len(net.Layers) == 0 {
		t.Fatal("expected non-empty Layers in NetworkDecl")
	}
}

func TestParseLearnDecl(t *testing.T) {
	input := `learn GPT on corpus ⟦ ⟧`
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	lrn, ok := prog.Decls[0].(*ast.LearnDecl)
	if !ok || lrn.Name != "GPT" || lrn.OnName != "corpus" {
		t.Fatalf("expected LearnDecl for GPT on corpus, got %+v", prog.Decls[0])
	}
}

func TestParseAccessDecl(t *testing.T) {
	input := `access std::io·`
	l := lexer.New(input, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	acc, ok := prog.Decls[0].(*ast.AccessDecl)
	if !ok || acc.Path != "std::io" {
		t.Fatalf("expected AccessDecl path 'std::io', got %+v", prog.Decls[0])
	}
}

func TestParseFullProgram(t *testing.T) {
	input := `process main⦅⦆ → int32 ⟦
    int32 x ↦ 1 ⊕ 2 ⊗ 3·
    emit x·
    yield 0·
⟧`
	l := lexer.New(input, "prog.ydv")
	p := New(l, "prog.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 process decl, got %d", len(prog.Decls))
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 3 {
		t.Fatalf("expected 3 statements in main, got %d", len(proc.Body.Statements))
	}
}

func TestParseFullProgram_Fallback(t *testing.T) {
	input := `process main\group\egroup \yields int32 \block
    int32 x \bind 1 \oplus 2 \otimes 3\stmt
    emit x\stmt
    yield 0\stmt
\eblock`
	l := lexer.New(input, "prog.ydv")
	p := New(l, "prog.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parse errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 process decl, got %d", len(prog.Decls))
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 3 {
		t.Fatalf("expected 3 statements in main fallback, got %d", len(proc.Body.Statements))
	}
}

func TestParseErrors_Recovery(t *testing.T) {
	// Program with syntax error: missing expression and invalid statement
	input := `process bad⦅⦆ → int32 ⟦
    int32 a ↦ ·
    emit·
    yield 0·
⟧`
	l := lexer.New(input, "bad.ydv")
	p := New(l, "bad.ydv")
	p.Parse()
	errs := p.Errors()
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 parse errors, got %d: %v", len(errs), errs)
	}
}

func TestParseErrors_BadOperator(t *testing.T) {
	input := `process test⦅⦆ ⟦ 1 + 2· ⟧`
	l := lexer.New(input, "bad_op.ydv")
	p := New(l, "bad_op.ydv")
	p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected parse error for legacy operator '+', but got none")
	}
}

func TestParseErrors_MissingStmtEnd(t *testing.T) {
	input := `process test⦅⦆ ⟦ int32 x ↦ 42 ⟧`
	l := lexer.New(input, "no_stmt.ydv")
	p := New(l, "no_stmt.ydv")
	p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected error for missing statement terminator '·', but got none")
	}
}

func TestParsePrecedenceTable(t *testing.T) {
	if precedenceOf(token.OTIM) <= precedenceOf(token.OSUM) {
		t.Fatal("expected OTIM (⊗) to bind tighter than OSUM (⊕)")
	}
	if precedenceOf(token.OPOW) <= precedenceOf(token.OTIM) {
		t.Fatal("expected OPOW (↑) to bind tighter than OTIM (⊗)")
	}
	if precedenceOf(token.MEMBER_ACCESS) <= precedenceOf(token.GROUP_OPEN) {
		t.Fatal("expected MEMBER_ACCESS to have higher precedence than GROUP_OPEN")
	}
}

func TestParseRangeExpr(t *testing.T) {
	expr := parseExpr(t, "0…10")
	rng, ok := expr.(*ast.RangeExpr)
	if !ok {
		t.Fatalf("expected *ast.RangeExpr, got %T", expr)
	}
	start, ok1 := rng.Start.(*ast.IntLit)
	end, ok2 := rng.End.(*ast.IntLit)
	if !ok1 || start.Value != 0 || !ok2 || end.Value != 10 {
		t.Fatalf("unexpected range bounds: %+v … %+v", rng.Start, rng.End)
	}
}

func TestParseNestedWhen(t *testing.T) {
	input := `when a ⟦ when b ⟦ yield 1· ⟧ ⟧`
	stmt := parseStmt(t, input)
	w, ok := stmt.(*ast.WhenStmt)
	if !ok {
		t.Fatalf("expected *ast.WhenStmt, got %T", stmt)
	}
	if len(w.Then.Statements) != 1 {
		t.Fatalf("expected 1 nested statement, got %d", len(w.Then.Statements))
	}
	innerWhen, ok := w.Then.Statements[0].(*ast.WhenStmt)
	if !ok {
		t.Fatalf("expected inner *ast.WhenStmt, got %T", w.Then.Statements[0])
	}
	if innerWhen.Condition == nil {
		t.Fatal("inner when condition is nil")
	}
}

func TestParseNestedCycle(t *testing.T) {
	input := `cycle x ⟦ cycle y ⟦ yield 1· ⟧ ⟧`
	stmt := parseStmt(t, input)
	c, ok := stmt.(*ast.CycleStmt)
	if !ok {
		t.Fatalf("expected *ast.CycleStmt, got %T", stmt)
	}
	if len(c.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement in outer cycle body, got %d", len(c.Body.Statements))
	}
	innerCycle, ok := c.Body.Statements[0].(*ast.CycleStmt)
	if !ok {
		t.Fatalf("expected inner *ast.CycleStmt, got %T", c.Body.Statements[0])
	}
	if innerCycle.Condition == nil {
		t.Fatal("inner cycle condition is nil")
	}
}

func TestParseMultipleStatements(t *testing.T) {
	input := `process f⦅⦆ ⟦
    int64 a ↦ 1·
    int64 b ↦ 2·
    int64 c ↦ 3·
    int64 d ↦ 4·
    int64 e ↦ 5·
    yield a ⊕ b ⊕ c ⊕ d ⊕ e·
⟧`
	l := lexer.New(input, "multi.ydv")
	p := New(l, "multi.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 6 {
		t.Fatalf("expected 6 statements, got %d", len(proc.Body.Statements))
	}
}

func TestParseEmptyBlock(t *testing.T) {
	input := `process f⦅⦆ ⟦ ⟧`
	l := lexer.New(input, "empty.ydv")
	p := New(l, "empty.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Body.Statements) != 0 {
		t.Fatalf("expected 0 statements in empty block, got %d", len(proc.Body.Statements))
	}
}

func TestParseProcessNoReturn(t *testing.T) {
	input := `process f⦅⦆ ⟦ yield 0· ⟧`
	l := lexer.New(input, "noreturn.ydv")
	p := New(l, "noreturn.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if proc.Returns != nil {
		t.Fatalf("expected Returns to be nil, got %+v", proc.Returns)
	}
}

func TestParseProcessMultiParams(t *testing.T) {
	input := `process f⦅int64 a⸴ int64 b⸴ int64 c⦆ ⟦ yield a ⊕ b ⊕ c· ⟧`
	l := lexer.New(input, "params.ydv")
	p := New(l, "params.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(proc.Params))
	}
	if proc.Params[0].Name != "a" || proc.Params[1].Name != "b" || proc.Params[2].Name != "c" {
		t.Fatalf("unexpected param names: %s, %s, %s", proc.Params[0].Name, proc.Params[1].Name, proc.Params[2].Name)
	}
}

func TestParseCallWithArgs(t *testing.T) {
	expr := parseExpr(t, "f⦅1⸴ 2⸴ 3⦆")
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected *ast.CallExpr, got %T", expr)
	}
	if len(call.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(call.Args))
	}
}

func TestParseGroupPrecedence(t *testing.T) {
	exprGrouped := parseExpr(t, "⦅1 ⊕ 2⦆ ⊗ 3")
	binGrouped, ok := exprGrouped.(*ast.BinaryExpr)
	if !ok || binGrouped.Op != token.OTIM {
		t.Fatalf("expected outer OTIM (⊗), got %+v", exprGrouped)
	}
	grp, ok := binGrouped.Left.(*ast.GroupExpr)
	if !ok {
		t.Fatalf("expected left operand to be GroupExpr, got %T", binGrouped.Left)
	}
	if _, ok := grp.Inner.(*ast.BinaryExpr); !ok {
		t.Fatalf("expected inner operand to be BinaryExpr (1 ⊕ 2)")
	}

	exprUngrouped := parseExpr(t, "1 ⊕ 2 ⊗ 3")
	binUngrouped, ok := exprUngrouped.(*ast.BinaryExpr)
	if !ok || binUngrouped.Op != token.OSUM {
		t.Fatalf("expected outer OSUM (⊕), got %+v", exprUngrouped)
	}
	if _, ok := binUngrouped.Right.(*ast.BinaryExpr); !ok {
		t.Fatalf("expected right operand to be inner BinaryExpr (2 ⊗ 3)")
	}
}

func TestParseAssignCompound(t *testing.T) {
	cases := []struct {
		input string
		op    token.Type
	}{
		{"x ↦⊕ 5·", token.BIND_ADD},
		{"x ↦⊖ 5·", token.BIND_SUB},
		{"x ↦⊗ 5·", token.BIND_MUL},
		{"x ↦⊘ 5·", token.BIND_DIV},
	}
	for _, c := range cases {
		stmt := parseStmt(t, c.input)
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			t.Fatalf("expected *ast.AssignStmt for %q, got %T", c.input, stmt)
		}
		if assign.Op != c.op {
			t.Fatalf("expected op %s, got %s", c.op, assign.Op)
		}
	}
}

func BenchmarkParser(b *testing.B) {
	prog := `process main⦅⦆ → int32 ⟦
    int32 x ↦ 1 ⊕ 2 ⊗ 3·
    int32 y ↦ x ⊘ 2·
    emit y·
    yield 0·
⟧`
	b.ReportAllocs()
	b.SetBytes(int64(len(prog)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := lexer.New(prog, "bench.ydv")
		p := New(l, "bench.ydv")
		_ = p.Parse()
	}
}

func TestParseCIRAndAGIStatements(t *testing.T) {
	src := `
expose my_mod ·

procedure add_one ⦅ int64 x ⦆ → int64 ⟦
	yield x ⊕ 1 ·
⟧

process main ⦅ ⦆ → int64 ⟦
	energy e1 ⟦
		yield 0 ·
	⟧

	qubit q1 ·
	observe q1 ·

	assert val ≽ 0.95 ·
	infer m on inp ↦ out ·

	memory mem ⟬ 50 ⟭ ·
	remember mem⟬"key"⟭ ↦ 123 ·
	int64 rec ↦ recall mem⟬"key"⟭ ·

	graph g ·
	node g ⟬ "a" ↦ 1 ⟭ ·
	link g ⟬ "a" ⸴ "b" ⸴ 0.5 ⟭ ·
	tensor⟨int64⟩ tr ↦ trace g ⟬ "a" ⸴ 2 ⟭ ·

	execute add_one ⦅ 10 ⦆ ·
	int64 exe ↦ execute add_one ⦅ 20 ⦆ ·

	confident⟨float64⟩ cf ↦ 0.9 ∵ 42.0 ·
	yield 0 ·
⟧
`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if len(prog.Decls) < 3 {
		t.Fatalf("expected at least 3 decls, got %d", len(prog.Decls))
	}

	// Test synchronize error recovery
	errSrc := `process main⦅⦆ ⟦ 123 !!! ??? ; yield 0 · ⟧`
	lErr := lexer.New(errSrc, "err.ydv")
	pErr := New(lErr, "err.ydv")
	_ = pErr.Parse()
}
