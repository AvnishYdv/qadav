package checker

import (
	"fmt"
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/types"
)

func parseAndCheck(src string) (*Checker, []Error) {
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		var errs []Error
		for _, e := range p.Errors() {
			errs = append(errs, Error{Pos: e.Pos, Msg: e.Message})
		}
		return nil, errs
	}
	c := New()
	errs := c.Check(prog)
	return c, errs
}

func TestValidBindAndAssign(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 x ↦ 42 ·
		x ← 100 ·
		yield x ·
	⟧
	`
	c, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
	sym, ok := c.globalScope.Lookup("main")
	if !ok || sym.Type.Kind != types.KindProcess {
		t.Fatalf("main process not found in global scope")
	}
}

func TestInferredBind(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		x ↦ 10 ·
		yield x ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestUndefinedVariable(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		yield undefinedVar ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for undefined variable, got none")
	}
}

func TestRedeclarationError(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 a ↦ 1 ·
		int64 a ↦ 2 ·
		yield a ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected redeclaration error, got none")
	}
}

func TestTypeMismatchInBind(t *testing.T) {
	src := `
	process main⦅⦆ ⟦
		int64 x ↦ "text string" ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected type mismatch error, got none")
	}
}

func TestTypeMismatchInAssign(t *testing.T) {
	src := `
	process main⦅⦆ ⟦
		int64 x ↦ 10 ·
		x ← "hello" ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected type mismatch error, got none")
	}
}

func TestArithmeticAndPromotion(t *testing.T) {
	src := `
	process calc⦅⦆ → float64 ⟦
		int32 a ↦ 10 ·
		float64 b ↦ 3.14 ·
		float64 c ↦ a ⊕ b ·
		yield c ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestComparisonOperators(t *testing.T) {
	src := `
	process comp⦅⦆ → logic ⟦
		int64 a ↦ 10 ·
		int64 b ↦ 20 ·
		logic res1 ↦ a ≺ b ·
		logic res2 ↦ a ≡ b ·
		logic res3 ↦ a ≼ b ·
		yield res1 ∧ res2 ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestLogicalOperators(t *testing.T) {
	src := `
	process testLogic⦅⦆ → logic ⟦
		logic tru ↦ affirm ·
		logic fls ↦ deny ·
		logic r ↦ ⦅tru ∧ fls⦆ ∨ ⦅¬fls⦆ ·
		yield r ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestWhenStatement(t *testing.T) {
	src := `
	process testWhen⦅⦆ → int64 ⟦
		int64 x ↦ 5 ·
		when x ≺ 10 ⟦
			yield 1 ·
		⟧ otherwise ⟦
			yield 2 ·
		⟧
		yield 0 ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestWhenInvalidCondition(t *testing.T) {
	src := `
	process testWhenInvalid⦅⦆ ⟦
		int64 x ↦ 5 ·
		when x ⟦
			emit x ·
		⟧
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for non-logic when condition, got none")
	}
}

func TestCycleLoop(t *testing.T) {
	src := `
	process loop⦅⦆ → int64 ⟦
		int64 count ↦ 0 ·
		cycle count ≺ 10 ⟦
			count ← count ⊕ 1 ·
		⟧
		yield count ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestProcessCallAndYield(t *testing.T) {
	src := `
	process add⦅int64 a⸴ int64 b⦆ → int64 ⟦
		yield a ⊕ b ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 sum ↦ add⦅10⸴ 20⦆ ·
		yield sum ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestYieldMismatch(t *testing.T) {
	src := `
	process getInt⦅⦆ → int64 ⟦
		yield "not an integer" ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected yield type mismatch error, got none")
	}
}

func TestEmitStatement(t *testing.T) {
	src := `
	process logger⦅⦆ ⟦
		emit 42 ·
		emit "hello" ·
		emit affirm ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestQuantumOps(t *testing.T) {
	src := `
	process qop⦅qubit q⦆ ⟦
		superpose q ·
		collapse q ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestQuantumErrorNonQubit(t *testing.T) {
	src := `
	process qerr⦅⦆ ⟦
		int64 x ↦ 5 ·
		superpose x ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected quantum type error on non-qubit, got none")
	}
}

func TestCheckMultipleErrors(t *testing.T) {
	src := `
	process bad⦅⦆ → int64 ⟦
		int64 a ↦ "str" ·
		b ← 10 ·
		int64 c ↦ 5 ·
		int64 c ↦ 6 ·
		when 123 ⟦
			yield "not int" ·
		⟧
		yield "wrong" ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) < 5 {
		t.Fatalf("expected at least 5 errors, got %d: %v", len(errs), errs)
	}
}

func TestCheckShadowing(t *testing.T) {
	src := `
	process shadow⦅⦆ → int64 ⟦
		int64 x ↦ 10 ·
		when affirm ⟦
			text x ↦ "nested" ·
			emit x ·
		⟧
		yield x ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for shadowed variable: %v", errs)
	}
}

func TestCheckRecursiveProcess(t *testing.T) {
	src := `
	process fact⦅int64 n⦆ → int64 ⟦
		when n ≼ 1 ⟦
			yield 1 ·
		⟧
		yield n ⊗ fact⦅n ⊖ 1⦆ ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected error in recursion: %v", errs)
	}
}

func TestCheckMutualRecursion(t *testing.T) {
	src := `
	process isEven⦅int64 n⦆ → logic ⟦
		when n ≡ 0 ⟦
			yield affirm ·
		⟧
		yield isOdd⦅n ⊖ 1⦆ ·
	⟧

	process isOdd⦅int64 n⦆ → logic ⟦
		when n ≡ 0 ⟦
			yield deny ·
		⟧
		yield isEven⦅n ⊖ 1⦆ ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected error in mutual recursion: %v", errs)
	}
}

func TestCheckDuplicateProcess(t *testing.T) {
	src := `
	process foo⦅⦆ ⟦ ⟧
	process foo⦅⦆ ⟦ ⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error on duplicate process declaration, got none")
	}
}

func TestCheckAllNumericPromotions(t *testing.T) {
	cases := []struct {
		typ1, val1, typ2, val2, resTyp string
	}{
		{"int8", "1", "int32", "2", "int32"},
		{"int32", "1", "float64", "2.0", "float64"},
		{"uint32", "1", "int64", "2", "int64"},
		{"float32", "1.0", "float64", "2.0", "float64"},
	}
	for _, tc := range cases {
		src := fmt.Sprintf(`
		process testPromote⦅⦆ → %s ⟦
			%s a ↦ %s ·
			%s b ↦ %s ·
			yield a ⊕ b ·
		⟧`, tc.resTyp, tc.typ1, tc.val1, tc.typ2, tc.val2)
		_, errs := parseAndCheck(src)
		if len(errs) > 0 {
			t.Errorf("promotion %s ⊕ %s -> %s failed: %v", tc.typ1, tc.typ2, tc.resTyp, errs)
		}
	}
}

func TestCheckComplexLiteral(t *testing.T) {
	src := `
	process testComplex⦅⦆ → complex128 ⟦
		complex128 z ↦ 3.14i ·
		yield z ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected error for complex literal: %v", errs)
	}
}

func TestCheckVoidProcess(t *testing.T) {
	src := `
	process doNothing⦅⦆ ⟦
		emit "action" ·
		yield ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected error in void process: %v", errs)
	}
}

func TestCheckErrorMessages(t *testing.T) {
	src := `
	process errPos⦅⦆ ⟦
		int64 x ↦ "wrong" ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error")
	}
	msg := errs[0].Error()
	if !strings.Contains(msg, "test.ydv:") {
		t.Errorf("error message missing file coordinate: %s", msg)
	}
}

func TestCheckScopeLookup(t *testing.T) {
	global := NewScope(nil)
	child := NewScope(global)
	if child.Parent() != global {
		t.Errorf("child.Parent() != global")
	}
	if err := global.Define(nil); err == nil {
		t.Errorf("expected error on Define(nil)")
	}
	sym := &Symbol{Name: "val", Type: types.Int64}
	_ = global.Define(sym)
	if s, ok := child.Lookup("val"); !ok || s != sym {
		t.Errorf("child.Lookup failed")
	}
	if _, ok := child.LookupCurrent("val"); ok {
		t.Errorf("LookupCurrent should not find parent symbol")
	}
}

func TestCheckCompoundAssignAndUnaries(t *testing.T) {
	src := `
	process compound⦅int64 a⸴ int64 b⦆ → int64 ⟦
		int64 x ↦ 10 ·
		x ↦⊕ 5 ·
		x ↦⊖ 2 ·
		x ↦⊗ 3 ·
		x ↦⊘ 2 ·
		int64 neg ↦ ⊖x ·
		logic b1 ↦ affirm ·
		logic nb ↦ ¬b1 ·
		yield neg ·
	⟧
	`
	c, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors in compound/unaries: %v", errs)
	}
	if len(c.ExprTypes()) == 0 {
		t.Errorf("expected non-empty ExprTypes()")
	}
	if len(c.Errors()) != 0 {
		t.Errorf("expected 0 Errors()")
	}
}

func TestCheckEntangleAndCollapseErrors(t *testing.T) {
	src := `
	process badQuantum⦅⦆ ⟦
		int64 x ↦ 10 ·
		qubit q ↦ 0 ·
		entangle x ⸴ q ·
		entangle q ⸴ x ·
		collapse x ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 quantum errors, got %d: %v", len(errs), errs)
	}
}

func TestCheckDuplicateParamsAndProcessErrors(t *testing.T) {
	src := `
	process dupParams⦅int64 a⸴ int64 a⦆ ⟦ ⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected duplicate parameter error")
	}
}

func TestCheckAllStatementsAndBranches(t *testing.T) {
	src := `
	access std::io ·
	expose myProc ·

	network MyNet ⟦ ⟧
	learn MyNet on MyData ⟦ ⟧

	process myProc⦅int64 a⸴ float64 b⦆ → int64 ⟦
		observe a ·
		cycle affirm ⟦
			halt ·
			skip ·
		⟧
		when 1.0 ≺ 2.0 ⟦
			logic r ↦ affirm ⊻ deny ·
		⟧
		yield a ·
	⟧
	`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckVariousErrorBranches(t *testing.T) {
	// Void emit error
	s1 := `process p⦅⦆ ⟦ ⟧ process m⦅⦆ ⟦ emit p⦅⦆ · ⟧`
	_, e1 := parseAndCheck(s1)
	if len(e1) == 0 {
		t.Errorf("expected error on void emit")
	}

	// Void return value error
	s2 := `process v⦅⦆ ⟦ yield 10 · ⟧`
	_, e2 := parseAndCheck(s2)
	if len(e2) == 0 {
		t.Errorf("expected error on yield in void")
	}

	// Missing return value error
	s3 := `process ret⦅⦆ → int64 ⟦ yield · ⟧`
	_, e3 := parseAndCheck(s3)
	if len(e3) == 0 {
		t.Errorf("expected error on bare yield in non-void")
	}

	// Logical and Unary operand errors
	s4 := `process errs⦅⦆ ⟦
		logic l ↦ 1 ∧ 2 ·
		logic n ↦ ¬10 ·
		int64 neg ↦ ⊖"str" ·
		int64 badCall ↦ neg⦅⦆ ·
	⟧`
	_, e4 := parseAndCheck(s4)
	if len(e4) < 3 {
		t.Errorf("expected at least 3 errors, got %d", len(e4))
	}

	// Call argument count and type error
	s5 := `process target⦅int64 x⦆ → int64 ⟦ yield x · ⟧
	process caller⦅⦆ ⟦
		int64 r1 ↦ target⦅⦆ ·
		int64 r2 ↦ target⦅"wrong"⦆ ·
	⟧`
	_, e5 := parseAndCheck(s5)
	if len(e5) < 2 {
		t.Errorf("expected at least 2 call errors, got %d", len(e5))
	}

	// Incompatible arithmetic
	s6 := `process arithErr⦅⦆ ⟦
		int64 x ↦ 1 ⊕ "str" ·
		logic c ↦ 1 ≡ "str" ·
	⟧`
	_, e6 := parseAndCheck(s6)
	if len(e6) < 2 {
		t.Errorf("expected arithmetic and comparison errors, got %d", len(e6))
	}

	// Compound assignment error
	s7 := `process cmpErr⦅⦆ ⟦
		text s ↦ "hello" ·
		s ↦⊕ 5 ·
	⟧`
	_, e7 := parseAndCheck(s7)
	if len(e7) == 0 {
		t.Errorf("expected compound assignment error")
	}
}

func TestCheckDirectASTNodes(t *testing.T) {
	c := New()
	pos := ast.Position{File: "test.ydv", Line: 1, Col: 1}

	// Test resolveType
	if c.resolveType(&ast.TensorType{Pos_: pos}) == nil {
		t.Errorf("resolveType(TensorType) failed")
	}
	if c.resolveType(&ast.NamedType{Pos_: pos, Name: "int32"}) == nil {
		t.Errorf("resolveType(NamedType) failed")
	}
	if c.resolveType(&ast.NamedType{Pos_: pos, Name: "unknown_xyz"}) != nil {
		t.Errorf("expected nil on unknown named type")
	}

	// Test checkExpr on IndexExpr, MemberExpr, RangeExpr
	c.enterScope()
	_ = c.curScope.Define(&Symbol{Name: "arr", Type: &types.Type{Kind: types.KindTensor, Elem: types.Float64}, Pos: pos})

	idxExpr := &ast.IndexExpr{
		Pos_:   pos,
		Object: &ast.Ident{Pos_: pos, Name: "arr"},
		Index:  &ast.IntLit{Pos_: pos, Value: 0},
	}
	if typ := c.checkExpr(idxExpr); !typ.Equal(types.Float64) {
		t.Errorf("IndexExpr type = %s, expected Float64", typ)
	}

	memberExpr := &ast.MemberExpr{
		Pos_:   pos,
		Object: &ast.Ident{Pos_: pos, Name: "arr"},
		Member: "len",
	}
	if typ := c.checkExpr(memberExpr); typ == nil {
		t.Errorf("MemberExpr type was nil")
	}

	rangeExpr := &ast.RangeExpr{
		Pos_:  pos,
		Start: &ast.IntLit{Pos_: pos, Value: 1},
		End:   &ast.IntLit{Pos_: pos, Value: 10},
	}
	if typ := c.checkExpr(rangeExpr); !typ.Equal(types.Tensor) {
		t.Errorf("RangeExpr type = %s, expected Tensor", typ)
	}

	c.leaveScope()
}

func BenchmarkChecker(b *testing.B) {
	src := `
	process fib⦅int64 n⦆ → int64 ⟦
		when n ≺ 2 ⟦
			yield n ·
		⟧
		yield fib⦅n ⊖ 1⦆ ⊕ fib⦅n ⊖ 2⦆ ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 n ↦ 10 ·
		yield fib⦅n⦆ ·
	⟧
	`
	l := lexer.New(src, "fib.ydv")
	p := parser.New(l, "fib.ydv")
	prog := p.Parse()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := New()
		_ = c.Check(prog)
	}
}

func TestCheckerCoverageBoost(t *testing.T) {
	src := `
access foo ·
expose bar ·

process helper⦅int64 a⸴ float64 b⦆ → int64 ⟦
	yield a ·
⟧

process main⦅⦆ → int64 ⟦
	// Loop with halt and skip
	cycle int64 i ↦ 0 · i ≺ 10 · i ← i ⊕ 1 ⟦
		when i ⩵ 2 ⟦
			skip ·
		⟧
		when i ⩵ 8 ⟦
			halt ·
		⟧
	⟧

	// Infer and propagate
	infer m on inp ↦ out ·
	propagate a from b ·

	// Direct calls
	int64 r ↦ helper⦅1⸴ 2.5⦆ ·

	// Error triggers
	memory m1 ⟬ 10 ⟭ ·
	remember m1⟬"k"⟭ ↦ 100 ·
	int64 rc ↦ recall m1⟬"k"⟭ ·

	graph g1 ·
	node g1 ⟬ "a" ↦ 1 ⟭ ·
	link g1 ⟬ "a" ⸴ "b" ⸴ 0.5 ⟭ ·
	tensor⟨int64⟩ tr ↦ trace g1 ⟬ "a" ⸴ 1 ⟭ ·

	yield r ·
⟧
`
	c, _ := parseAndCheck(src)
	if c != nil {
		_ = c.ExprTypes()
		_ = c.Networks()
	}

	// Negative error checks for CIR and AGI
	errSrcs := []string{
		`process main⦅⦆ → int64 ⟦ assert 1 ≽ "bad" · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ assert "bad" ≽ 0.5 · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ memory m ⟬ 1.5 ⟭ · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ remember 123⟬"k"⟭ ↦ 1 · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ int64 x ↦ recall 123⟬"k"⟭ · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ node 123 ⟬ "a" ↦ 1 ⟭ · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ link 123 ⟬ "a" ⸴ "b" ⸴ 1.0 ⟭ · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ tensor⟨int64⟩ tr ↦ trace 123 ⟬ "a" ⸴ 1 ⟭ · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ int64 x ↦ 0.5 ∵ 10.0 · yield 0 · ⟧`,
		`process main⦅⦆ → int64 ⟦ int64 x ↦ 1.5 ∵ 10.0 · yield 0 · ⟧`,
	}
	for _, es := range errSrcs {
		_, errs := parseAndCheck(es)
		if len(errs) == 0 {
			t.Errorf("expected error for src %s, got none", es)
		}
	}
}

func TestDirectCoverageBoost(t *testing.T) {
	c := New()
	pos := ast.Position{File: "test.ydv", Line: 1, Col: 1}

	// Prime symbolic methods
	if !IsPrime(7) || IsPrime(8) {
		t.Errorf("IsPrime failed")
	}
	factors := Factorize(12)
	if len(factors) != 3 {
		t.Errorf("Factorize(12) expected 3 factors, got %d", len(factors))
	}
	encoded := EncodeConcept([]int64{2, 3, 5})
	if encoded != 30 {
		t.Errorf("EncodeConcept failed: %d", encoded)
	}
	decoded := DecodeConcept(encoded)
	if len(decoded) != 3 {
		t.Errorf("DecodeConcept failed: %v", decoded)
	}

	// ExecuteStmt
	c.checkExecuteStmt(&ast.ExecuteStmt{
		Pos_:      pos,
		Procedure: &ast.Ident{Pos_: pos, Name: "proc"},
		Args:      []ast.Expr{&ast.IntLit{Pos_: pos, Value: 1}},
	})

	// InferStmt & PropagateStmt
	c.checkInferStmt(&ast.InferStmt{
		Pos_:       pos,
		Model:      &ast.Ident{Pos_: pos, Name: "net"},
		Input:      &ast.Ident{Pos_: pos, Name: "inp"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.9},
		Into:       "out_tensor",
	})
	c.checkPropagateStmt(&ast.PropagateStmt{
		Pos_:   pos,
		Value:  &ast.Ident{Pos_: pos, Name: "v"},
		Source: &ast.Ident{Pos_: pos, Name: "src"},
	})

	// EnergyDecl & TileStmt
	c.checkEnergyDecl(&ast.EnergyDecl{
		Pos_: pos,
		Name: "e1",
		Fn:   &ast.Block{Pos_: pos},
	})
	c.checkTileStmt(&ast.TileStmt{
		Pos_:    pos,
		Grid:    &ast.Ident{Pos_: pos, Name: "g"},
		Pattern: &ast.Ident{Pos_: pos, Name: "p"},
	})
	c.checkNetworkDecl(&ast.NetworkDecl{
		Pos_: pos,
		Name: "net1",
	})
}
