package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestChecker_CTSAStatements(t *testing.T) {
	input := `
process main⦅⦆ → int64 ⟦
    ctsa_attest ast ↦ "ast_tree" env ↦ "type_env" rules ↦ "type_rules" into cert0 ·
    ctsa_verify cert0 into v0 ·
    corrupt_ctsa cert0 field ↦ 1 ·
    yield 0 ·
⟧
`
	l := lexer.New(input, "ctsa_test.ydv")
	p := parser.New(l, "ctsa_test.ydv")
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

func TestChecker_CTSACorruptNonTestMode(t *testing.T) {
	input := `
process main⦅⦆ → int64 ⟦
    ctsa_attest ast ↦ "ast_tree" env ↦ "type_env" rules ↦ "type_rules" into cert0 ·
    corrupt_ctsa cert0 field ↦ 1 ·
    yield 0 ·
⟧
`
	l := lexer.New(input, "ctsa_tamper.ydv")
	p := parser.New(l, "ctsa_tamper.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse failed: %v", p.Errors())
	}

	c := New()
	c.IsTest = false
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for corrupt_ctsa in non-test mode")
	}
	if !strings.Contains(errs[0].Error(), "corrupt_ctsa is only permitted in test mode") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}
