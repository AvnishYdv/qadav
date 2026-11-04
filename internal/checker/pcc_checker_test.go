package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestChecker_PCCStatements(t *testing.T) {
	input := `
process main⦅⦆ → int64 ⟦
    pcc_attest ast ↦ "ast_hello" codegen ↦ "codegen_path" ir ↦ "llvm_ir" into cert0 ·
    pcc_verify cert0 into v0 ·
    corrupt_pcc cert0 field ↦ 1 ·
    yield 0 ·
⟧
`
	l := lexer.New(input, "pcc_test.ydv")
	p := parser.New(l, "pcc_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse failed: %v", p.Errors())
	}

	c := New()
	c.IsTest = true
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("Check failed: %v", errs)
	}
}

func TestChecker_PCCCorruptNonTestMode(t *testing.T) {
	input := `
process main⦅⦆ → int64 ⟦
    pcc_attest ast ↦ "ast_hello" codegen ↦ "codegen_path" ir ↦ "llvm_ir" into cert0 ·
    corrupt_pcc cert0 field ↦ 1 ·
    yield 0 ·
⟧
`
	l := lexer.New(input, "pcc_tamper.ydv")
	p := parser.New(l, "pcc_tamper.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse failed: %v", p.Errors())
	}

	c := New()
	c.IsTest = false
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for corrupt_pcc in non-test mode")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "test mode") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error containing 'test mode', got %v", errs)
	}
}

func TestChecker_PCCNilEdgeCases(t *testing.T) {
	c := New()
	c.checkPCCAttestStmt(nil)
	c.checkPCCVerifyStmt(nil)
	c.checkCorruptPCCStmt(nil)
}
