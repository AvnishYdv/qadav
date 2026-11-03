package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestCheckerQHAL(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @qhal_backend sim
    qasm_export into qasm_str with size 1024 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_qhal.ydv")
	p := parser.New(l, "chk_qhal.ydv")
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

func TestCheckerQHALInvalidBackend(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @qhal_backend invalid_quantum_backend
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_qhal_inv.ydv")
	p := parser.New(l, "chk_qhal_inv.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for invalid backend, got none")
	}
	if !strings.Contains(errs[0].Error(), "invalid QHAL backend") {
		t.Fatalf("unexpected error message: %v", errs)
	}
}

func TestCheckerSEA(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text s1 ↦ "v1" ·
    text s2 ↦ "v2" ·
    text r ↦ "reason" ·
    evolve from s1 to s2 with reason r into prop ·
    evolve_verify prop into is_valid ·
    corrupt_evolve prop at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_sea.ydv")
	p := parser.New(l, "chk_sea.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// 1. Non-test mode should reject corrupt_evolve
	c := New()
	c.IsTest = false
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for corrupt_evolve in non-test mode")
	}
	if !strings.Contains(errs[0].Error(), "corrupt_evolve is only permitted in test mode") {
		t.Fatalf("unexpected error message: %v", errs)
	}

	// 2. Test mode should pass
	cTest := New()
	cTest.IsTest = true
	errsTest := cTest.Check(prog)
	if len(errsTest) > 0 {
		t.Fatalf("unexpected errors in test mode: %v", errsTest)
	}
}

func TestCheckerSEAUndefinedSymbols(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_verify undef_prop into is_valid ·
    evolve_history to undef_chain ·
    corrupt_evolve undef_prop2 at ↦ 0 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_sea_undef.ydv")
	p := parser.New(l, "chk_sea_undef.ydv")
	prog := p.Parse()
	c := New()
	c.IsTest = true
	errs := c.Check(prog)
	if len(errs) != 3 {
		t.Fatalf("expected 3 undefined symbol errors, got %d: %v", len(errs), errs)
	}
}
