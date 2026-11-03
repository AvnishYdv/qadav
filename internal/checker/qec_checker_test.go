package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/token"
)

func parseAndCheckTestMode(src string, isTest bool) (*Checker, []Error) {
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
	c.IsTest = isTest
	errs := c.Check(prog)
	return c, errs
}

func TestQECChecker_DistanceValidation(t *testing.T) {
	validDistances := []int{3, 5, 7, 9, 11}
	for _, d := range validDistances {
		src := strings.Replace(`
		process main⦅⦆ → int64 ⟦
			qec_memory qm as surface_code distance ↦ DIST ⟦ ⟧ ·
			yield 0 ·
		⟧`, "DIST", string(rune('0'+d)), 1)
		if d >= 10 {
			src = strings.Replace(`
			process main⦅⦆ → int64 ⟦
				qec_memory qm as surface_code distance ↦ 11 ⟦ ⟧ ·
				yield 0 ·
			⟧`, "11", "11", 1)
		}
		_, errs := parseAndCheck(src)
		if len(errs) > 0 {
			t.Errorf("expected distance %d to be valid, got errors: %v", d, errs)
		}
	}

	invalidDistances := []string{"1", "2", "4", "6", "8", "10", "12", "13"}
	for _, dStr := range invalidDistances {
		src := `
		process main⦅⦆ → int64 ⟦
			qec_memory qm as surface_code distance ↦ ` + dStr + ` ⟦ ⟧ ·
			yield 0 ·
		⟧`
		_, errs := parseAndCheck(src)
		if len(errs) == 0 {
			t.Errorf("expected distance %s to be rejected, got none", dStr)
		}
	}

	// Non-integer distance expression
	c := New()
	decl := &ast.QECMemoryDecl{
		Name:     "qm_bad_dist",
		CodeType: token.SURFACE_CODE,
		Distance: &ast.FloatLit{Value: 3.5},
	}
	prog := &ast.Program{
		Decls: []ast.Decl{
			&ast.ProcessDecl{
				Name: "p",
				Body: &ast.Block{Statements: []ast.Stmt{decl}},
			},
		},
	}
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Errorf("expected error for non-integer distance float")
	}
}

func TestQECChecker_VerifyBeforeCorrect(t *testing.T) {
	// Calling correct without verify_syndrome must fail
	unverifiedSrc := `
	process main⦅⦆ → int64 ⟦
		qec_memory qm as surface_code distance ↦ 3 ⟦ ⟧ ·
		syndrome qm into syn0 ·
		correct qm with syn0 ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(unverifiedSrc)
	if len(errs) == 0 {
		t.Fatalf("expected compile error when correct is called before verify_syndrome")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Msg, "correct requires verified syndrome") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error 'correct requires verified syndrome', got: %v", errs)
	}

	// Calling verify_syndrome first then correct must succeed
	verifiedSrc := `
	process main⦅⦆ → int64 ⟦
		qec_memory qm as surface_code distance ↦ 3 ⟦ ⟧ ·
		syndrome qm into syn0 ·
		verify_syndrome qm with syn0 ·
		correct qm with syn0 ·
		yield 0 ·
	⟧`
	_, errs2 := parseAndCheck(verifiedSrc)
	if len(errs2) > 0 {
		t.Fatalf("expected verified syndrome to succeed, got errors: %v", errs2)
	}
}

func TestQECChecker_CorruptSyndromeGating(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		qec_memory qm as surface_code distance ↦ 3 ⟦ ⟧ ·
		syndrome qm into syn0 ·
		corrupt_syndrome qm⟬syn0⟭ at ↦ 0 ·
		verify_syndrome qm with syn0 ·
		correct qm with syn0 ·
		yield 0 ·
	⟧`

	// Without test mode: must produce error
	_, errs := parseAndCheckTestMode(src, false)
	if len(errs) == 0 {
		t.Fatalf("expected corrupt_syndrome to be rejected without --test flag")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Msg, "corrupt_syndrome is test-only") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error 'corrupt_syndrome is test-only', got: %v", errs)
	}

	// With test mode: must succeed
	_, errsTest := parseAndCheckTestMode(src, true)
	if len(errsTest) > 0 {
		t.Fatalf("expected corrupt_syndrome to pass with IsTest=true, got errors: %v", errsTest)
	}

	// Invalid offset type in corrupt_syndrome
	badOffsetSrc := `
	process main⦅⦆ → int64 ⟦
		qec_memory qm as surface_code distance ↦ 3 ⟦ ⟧ ·
		syndrome qm into syn0 ·
		corrupt_syndrome qm⟬syn0⟭ at ↦ "invalid_offset" ·
		yield 0 ·
	⟧`
	_, errsBadOffset := parseAndCheckTestMode(badOffsetSrc, true)
	if len(errsBadOffset) == 0 {
		t.Fatalf("expected error for non-integer offset in corrupt_syndrome")
	}
}

func TestQECChecker_BoundaryValidation(t *testing.T) {
	validBoundarySrc := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		qec_memory qm as surface_code distance ↦ 3 with b ⟦ ⟧ ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(validBoundarySrc)
	if len(errs) > 0 {
		t.Fatalf("expected valid ethical_boundary to pass, got: %v", errs)
	}

	invalidBoundarySrc := `
	process main⦅⦆ → int64 ⟦
		int64 not_a_boundary ↦ 42 ·
		qec_memory qm as surface_code distance ↦ 3 with not_a_boundary ⟦ ⟧ ·
		yield 0 ·
	⟧`
	_, errsInvalid := parseAndCheck(invalidBoundarySrc)
	if len(errsInvalid) == 0 {
		t.Fatalf("expected error when boundary is not an ethical_boundary")
	}
}

func TestQECChecker_InvalidCodeType(t *testing.T) {
	c := New()
	decl := &ast.QECMemoryDecl{
		Name:     "qm_bad_type",
		CodeType: token.IDENT, // not surface_code or qldpc_code
		Distance: &ast.IntLit{Value: 3},
	}
	prog := &ast.Program{
		Decls: []ast.Decl{
			&ast.ProcessDecl{
				Name: "p",
				Body: &ast.Block{Statements: []ast.Stmt{decl}},
			},
		},
	}
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for invalid code type")
	}
}

func TestQECChecker_TargetValidation(t *testing.T) {
	// syndrome on non-qec_memory
	badSynSrc := `
	process main⦅⦆ → int64 ⟦
		int64 not_qec ↦ 10 ·
		syndrome not_qec into s ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(badSynSrc)
	if len(errs) == 0 {
		t.Fatalf("expected error for syndrome on non-qec_memory")
	}

	// verify_syndrome on non-qec_memory or non-syndrome
	badVerifySrc := `
	process main⦅⦆ → int64 ⟦
		int64 not_qec ↦ 10 ·
		int64 not_syn ↦ 20 ·
		verify_syndrome not_qec with not_syn ·
		yield 0 ·
	⟧`
	_, errsVerify := parseAndCheck(badVerifySrc)
	if len(errsVerify) == 0 {
		t.Fatalf("expected errors for verify_syndrome with invalid types")
	}

	// correct on non-qec_memory or non-syndrome
	badCorrectSrc := `
	process main⦅⦆ → int64 ⟦
		int64 not_qec ↦ 10 ·
		int64 not_syn ↦ 20 ·
		correct not_qec with not_syn ·
		yield 0 ·
	⟧`
	_, errsCorrect := parseAndCheck(badCorrectSrc)
	if len(errsCorrect) == 0 {
		t.Fatalf("expected errors for correct with invalid types")
	}

	// corrupt_syndrome on non-qec_memory
	badCorruptSrc := `
	process main⦅⦆ → int64 ⟦
		int64 not_qec ↦ 10 ·
		int64 not_syn ↦ 20 ·
		corrupt_syndrome not_qec⟬not_syn⟭ at ↦ 0 ·
		yield 0 ·
	⟧`
	_, errsCorrupt := parseAndCheckTestMode(badCorruptSrc, true)
	if len(errsCorrupt) == 0 {
		t.Fatalf("expected errors for corrupt_syndrome on non-qec_memory")
	}
}
