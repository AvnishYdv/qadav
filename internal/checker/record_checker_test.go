package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestChecker_RecordDeclSuccess(t *testing.T) {
	src := `record Token ⟦ kind ↦ text · value ↦ text · line ↦ int64 · col ↦ int64 · ⟧`
	l := lexer.New(src, "chk_rec.ydv")
	p := parser.New(l, "chk_rec.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}
}

func TestChecker_RecordDuplicateField(t *testing.T) {
	src := `record Token ⟦ kind ↦ text · kind ↦ text · ⟧`
	l := lexer.New(src, "chk_rec_dup.ydv")
	p := parser.New(l, "chk_rec_dup.ydv")
	prog := p.Parse()

	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for duplicate field, got none")
	}
	if !strings.Contains(errs[0].Error(), "duplicate field") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}
}

func TestChecker_VariantDeclSuccess(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Identifier(name ↦ text) · ⟧`
	l := lexer.New(src, "chk_var.ydv")
	p := parser.New(l, "chk_var.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}
}

func TestChecker_VariantDuplicateCase(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Num(val ↦ int64) · ⟧`
	l := lexer.New(src, "chk_var_dup.ydv")
	p := parser.New(l, "chk_var_dup.ydv")
	prog := p.Parse()

	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for duplicate case, got none")
	}
	if !strings.Contains(errs[0].Error(), "duplicate case") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}
}

func TestChecker_MatchExhaustiveSuccess(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Identifier(name ↦ text) · ⟧
process classify⦅Expr e⦆ → int64 ⟦
    match e ⟦
        | Num(x) ⟹ yield x ·
        | Identifier(s) ⟹ yield 0 ·
    ⟧
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_match.ydv")
	p := parser.New(l, "chk_match.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}
}

func TestChecker_MatchNonExhaustiveFailure(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Identifier(name ↦ text) · ⟧
process classify⦅Expr e⦆ → int64 ⟦
    match e ⟦
        | Num(x) ⟹ yield x ·
    ⟧
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_match_non_exh.ydv")
	p := parser.New(l, "chk_match_non_exh.ydv")
	prog := p.Parse()

	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for non-exhaustive match, got none")
	}
	if !strings.Contains(errs[0].Error(), "non-exhaustive match") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}
}

func TestChecker_RecordInitAndAccessSuccess(t *testing.T) {
	src := `record Token ⟦ kind ↦ text · value ↦ text · line ↦ int64 · col ↦ int64 · ⟧
process main⦅⦆ → int64 ⟦
    Token tok ↦ record_init Token ⟦ kind ↦ "IDENT" ⸴ value ↦ "hello" ⸴ line ↦ 42 ⸴ col ↦ 1 · ⟧ ·
    text k ↦ tok.kind ·
    int64 l ↦ tok.line ·
    yield l ·
⟧`
	l := lexer.New(src, "chk_init.ydv")
	p := parser.New(l, "chk_init.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}
}

func TestChecker_RecordInitUnknownField(t *testing.T) {
	src := `record Token ⟦ kind ↦ text · ⟧
process main⦅⦆ → int64 ⟦
    Token tok ↦ record_init Token ⟦ unknown ↦ "x" · ⟧ ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_init_bad.ydv")
	p := parser.New(l, "chk_init_bad.ydv")
	prog := p.Parse()

	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for unknown field, got none")
	}
	if !strings.Contains(errs[0].Error(), "does not exist on record") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}
}

func TestChecker_CorruptRSCAGating(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text r ↦ "rec0" ·
    corrupt_rsca r at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_rsca.ydv")
	p := parser.New(l, "chk_rsca.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// 1. Without test flag: error
	c := New()
	c.IsTest = false
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected corrupt_rsca error without test flag, got none")
	}
	if !strings.Contains(errs[0].Error(), "corrupt_rsca is only permitted in test mode") {
		t.Fatalf("unexpected error message: %v", errs[0])
	}

	// 2. With test flag: passes
	cTest := New()
	cTest.IsTest = true
	errsTest := cTest.Check(prog)
	if len(errsTest) > 0 {
		t.Fatalf("corrupt_rsca should pass in test mode, got: %v", errsTest)
	}
}

func TestChecker_VariantInitSuccessAndErrors(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Identifier(name ↦ text) · ⟧
process main⦅⦆ → int64 ⟦
    Expr e1 ↦ variant_init Expr ⦅Num ↦ 42⦆ ·
    Expr e2 ↦ variant_init Expr ⦅Identifier ↦ "x"⦆ ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_vinit.ydv")
	p := parser.New(l, "chk_vinit.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	if len(c.RecordDecls()) != 0 {
		t.Fatalf("expected 0 record decls, got %d", len(c.RecordDecls()))
	}
	if len(c.VariantDecls()) != 1 {
		t.Fatalf("expected 1 variant decl, got %d", len(c.VariantDecls()))
	}

	// Unknown variant
	src2 := `process main⦅⦆ → int64 ⟦
    int64 x ↦ variant_init Unknown ⦅Case1 ↦ 1⦆ ·
    yield 0 ·
⟧`
	prog2 := parser.New(lexer.New(src2, "bad1.ydv"), "bad1.ydv").Parse()
	c2 := New()
	if len(c2.Check(prog2)) == 0 {
		t.Fatalf("expected error for unknown variant")
	}

	// Unknown case
	src3 := `variant Expr ⟦ | Num(val ↦ int64) · ⟧
process main⦅⦆ → int64 ⟦
    Expr e ↦ variant_init Expr ⦅BadCase ↦ 1⦆ ·
    yield 0 ·
⟧`
	prog3 := parser.New(lexer.New(src3, "bad2.ydv"), "bad2.ydv").Parse()
	c3 := New()
	if len(c3.Check(prog3)) == 0 {
		t.Fatalf("expected error for unknown variant case")
	}
}

func TestChecker_RSCAAttestAndVerify(t *testing.T) {
	src := `process main⦅⦆ → logic ⟦
    text source_code ↦ "source_code" ·
    text go_bootstrap ↦ "go_bootstrap" ·
    text stage1_bin ↦ "stage1_bin" ·
    rsca_attest stage ↦ 0 from source ↦ source_code compiler ↦ go_bootstrap output ↦ stage1_bin into rsca0 ·
    rsca_verify rsca0 into verified ·
    yield verified ·
⟧`
	l := lexer.New(src, "chk_rsca_pass.ydv")
	p := parser.New(l, "chk_rsca_pass.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	// Invalid stage
	srcBad := `process main⦅⦆ → int64 ⟦
    text s ↦ "s" · text c ↦ "c" · text o ↦ "o" ·
    rsca_attest stage ↦ 5 from source ↦ s compiler ↦ c output ↦ o into rsca0 ·
    yield 0 ·
⟧`
	progBad := parser.New(lexer.New(srcBad, "bad_stage.ydv"), "bad_stage.ydv").Parse()
	cBad := New()
	if len(cBad.Check(progBad)) == 0 {
		t.Fatalf("expected error for invalid RSCA stage")
	}
}

func TestChecker_FieldAccessErrors(t *testing.T) {
	src := `record Point ⟦ x ↦ int64 · y ↦ int64 · ⟧
process main⦅⦆ → int64 ⟦
    Point pt ↦ record_init Point ⟦ x ↦ 10 ⸴ y ↦ 20 · ⟧ ·
    int64 z ↦ pt.z ·
    yield z ·
⟧`
	prog := parser.New(lexer.New(src, "bad_field.ydv"), "bad_field.ydv").Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error accessing unknown field .z")
	}

	// Access on non-record
	src2 := `process main⦅⦆ → int64 ⟦
    int64 a ↦ 42 ·
    int64 b ↦ a.foo ·
    yield b ·
⟧`
	prog2 := parser.New(lexer.New(src2, "bad_acc.ydv"), "bad_acc.ydv").Parse()
	c2 := New()
	if len(c2.Check(prog2)) == 0 {
		t.Fatalf("expected error accessing field on non-record")
	}
}

func TestChecker_MatchAndDeclMoreErrors(t *testing.T) {
	// 1. Match subject not variant
	src1 := `process main⦅⦆ → int64 ⟦
    int64 x ↦ 10 ·
    match x ⟦
        | Case1(a) ⟹ yield 1 ·
    ⟧
    yield 0 ·
⟧`
	c1 := New()
	if len(c1.Check(parser.New(lexer.New(src1, "m1.ydv"), "m1.ydv").Parse())) == 0 {
		t.Fatalf("expected error matching on non-variant")
	}

	// 2. Arm wrong binding count
	src2 := `variant Shape ⟦ | Circle(radius ↦ int64) · ⟧
process main⦅Shape s⦆ → int64 ⟦
    match s ⟦
        | Circle(r ⸴ extra) ⟹ yield r ·
    ⟧
    yield 0 ·
⟧`
	c2 := New()
	if len(c2.Check(parser.New(lexer.New(src2, "m2.ydv"), "m2.ydv").Parse())) == 0 {
		t.Fatalf("expected error for binding count mismatch")
	}

	// 3. Duplicate arm case
	src3 := `variant Shape ⟦ | Circle(radius ↦ int64) · ⟧
process main⦅Shape s⦆ → int64 ⟦
    match s ⟦
        | Circle(r) ⟹ yield r ·
        | Circle(r2) ⟹ yield r2 ·
    ⟧
    yield 0 ·
⟧`
	c3 := New()
	if len(c3.Check(parser.New(lexer.New(src3, "m3.ydv"), "m3.ydv").Parse())) == 0 {
		t.Fatalf("expected error for duplicate arm case")
	}

	// 4. Invalid field types in record and variant
	src4 := `record BadR ⟦ x ↦ nonexistent_type · ⟧`
	c4 := New()
	if len(c4.Check(parser.New(lexer.New(src4, "m4.ydv"), "m4.ydv").Parse())) == 0 {
		t.Fatalf("expected error for invalid record field type")
	}

	src5 := `variant BadV ⟦ | CaseA(x ↦ nonexistent_type) · ⟧`
	c5 := New()
	if len(c5.Check(parser.New(lexer.New(src5, "m5.ydv"), "m5.ydv").Parse())) == 0 {
		t.Fatalf("expected error for invalid variant field type")
	}

	// 5. CorruptRSCA & VerifyRSCA errors
	src6 := `process main⦅⦆ → int64 ⟦
    corrupt_rsca missing_rec at ↦ "bad_offset" ·
    rsca_verify another_missing into v ·
    yield 0 ·
⟧`
	c6 := New()
	c6.IsTest = true
	errs6 := c6.Check(parser.New(lexer.New(src6, "m6.ydv"), "m6.ydv").Parse())
	if len(errs6) == 0 {
		t.Fatalf("expected errors for missing records and bad offset, got none")
	}

	// 6. Record init unknown record
	src7 := `process main⦅⦆ → int64 ⟦
    int64 x ↦ record_init NonExistent ⟦ f ↦ 1 · ⟧ ·
    yield x ·
⟧`
	c7 := New()
	if len(c7.Check(parser.New(lexer.New(src7, "m7.ydv"), "m7.ydv").Parse())) == 0 {
		t.Fatalf("expected error for unknown record in record_init")
	}

	// 7. Field not found on known record
	src8 := `record Person ⟦ name ↦ text · age ↦ int64 · ⟧
process main⦅⦆ → int64 ⟦
    Person p ↦ record_init Person ⟦ name ↦ "Alice" ⸴ age ↦ 30 · ⟧ ·
    text s ↦ p.unknown_field ·
    yield 0 ·
⟧`
	c8 := New()
	if len(c8.Check(parser.New(lexer.New(src8, "m8.ydv"), "m8.ydv").Parse())) == 0 {
		t.Fatalf("expected error for missing field on known record")
	}
}
