package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestChecker_TranspileValid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text buf ↦ "" ·
    transpile_to "qiskit" into buf with size 2048 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_tr_valid.ydv")
	p := parser.New(l, "chk_tr_valid.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestChecker_TranspileInvalidTarget(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text buf ↦ "" ·
    transpile_to "unknown_target" into buf with size 2048 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_tr_inv.ydv")
	p := parser.New(l, "chk_tr_inv.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for invalid transpilation target, got none")
	}
	if !strings.Contains(errs[0].Error(), "invalid transpilation target") {
		t.Fatalf("expected invalid transpilation target error, got: %v", errs[0])
	}
}

func TestChecker_TranspileSmallBuffer(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text buf ↦ "" ·
    transpile_to "qiskit" into buf with size 128 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_tr_small.ydv")
	p := parser.New(l, "chk_tr_small.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for buffer size < 256, got none")
	}
	if !strings.Contains(errs[0].Error(), "below minimum required 256 bytes") {
		t.Fatalf("expected small buffer error, got: %v", errs[0])
	}
}

func TestChecker_EEAValid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy constant into policy with rate 0.10 budget 100 ·
    evolve_bounded policy from "a" to "b" with reason "tweak" max_flips 8 max_new_symbols 2 into prop env ·
    eea_verify env into is_valid ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_valid.ydv")
	p := parser.New(l, "chk_eea_valid.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	c := New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestChecker_EEAInvalidPolicyKind(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy bad_kind into policy with rate 0.10 budget 100 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_inv_kind.ydv")
	p := parser.New(l, "chk_eea_inv_kind.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for invalid policy kind, got none")
	}
	if !strings.Contains(errs[0].Error(), "invalid evolution policy kind") {
		t.Fatalf("expected invalid policy kind error, got: %v", errs[0])
	}
}

func TestChecker_CorruptEEAGating(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy constant into policy with rate 0.10 budget 100 ·
    evolve_bounded policy from "a" to "b" with reason "tweak" max_flips 8 max_new_symbols 2 into prop env ·
    corrupt_eea env at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_corrupt.ydv")
	p := parser.New(l, "chk_eea_corrupt.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Without --test
	cProd := New()
	cProd.IsTest = false
	errsProd := cProd.Check(prog)
	if len(errsProd) == 0 {
		t.Fatalf("expected corrupt_eea to fail in production mode, got none")
	}
	if !strings.Contains(errsProd[0].Error(), "corrupt_eea is only permitted in test mode") {
		t.Fatalf("expected test mode error, got: %v", errsProd[0])
	}

	// With --test
	cTest := New()
	cTest.IsTest = true
	errsTest := cTest.Check(prog)
	if len(errsTest) > 0 {
		t.Fatalf("corrupt_eea should pass in test mode, got errors: %v", errsTest)
	}
}

func TestChecker_TranspileWrongIntoType(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    int64 buf ↦ 42 ·
    transpile_to "qiskit" into buf with size 2048 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_tr_bad_type.ydv")
	p := parser.New(l, "chk_tr_bad_type.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for non-text into variable, got none")
	}
	if !strings.Contains(errs[0].Error(), "must have type text") {
		t.Fatalf("expected must have type text error, got: %v", errs[0])
	}
}

func TestChecker_EEAInvalidRateAndBudget(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy constant into policy with rate 1.50 budget 0 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_bad_rate.ydv")
	p := parser.New(l, "chk_eea_bad_rate.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) < 2 {
		t.Fatalf("expected 2 errors for invalid rate and budget, got %d: %v", len(errs), errs)
	}
}

func TestChecker_EEAUndefinedPolicy(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_bounded missing_policy from "a" to "b" with reason "tweak" max_flips 8 max_new_symbols 2 into prop env ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_undef_pol.ydv")
	p := parser.New(l, "chk_eea_undef_pol.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for undefined policy, got none")
	}
	if !strings.Contains(errs[0].Error(), "undefined evolution policy") {
		t.Fatalf("expected undefined evolution policy error, got: %v", errs[0])
	}
}

func TestChecker_EEAInvalidBounds(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy constant into pol with rate 0.10 budget 100 ·
    evolve_bounded pol from "a" to "b" with reason "tweak" max_flips 0 max_new_symbols 2 into prop env ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_bad_flips.ydv")
	p := parser.New(l, "chk_eea_bad_flips.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for max_flips < 1, got none")
	}
	if !strings.Contains(errs[0].Error(), "max_flips must be >= 1") {
		t.Fatalf("expected max_flips error, got: %v", errs[0])
	}
}

func TestChecker_EEAUndefinedEnvelope(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    eea_verify missing_env into is_valid ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_eea_undef_env.ydv")
	p := parser.New(l, "chk_eea_undef_env.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for undefined envelope, got none")
	}
	if !strings.Contains(errs[0].Error(), "undefined evolution envelope") {
		t.Fatalf("expected undefined evolution envelope error, got: %v", errs[0])
	}
}
