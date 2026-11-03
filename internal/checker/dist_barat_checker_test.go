package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestCheckerDistributed(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    pserver_init into ps with 4 workers ·
    tensor⟨float64⟩⟬2⸴ 2⟭ g0 ↦ tensor⟬⟬0.1⸴ 0.2⟭⸴ ⟬0.3⸴ 0.4⟭⟭ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ g1 ↦ tensor⟬⟬0.5⸴ 0.6⟭⸴ ⟬0.7⸴ 0.8⟭⟭ ·
    all_reduce from ⟬g0⸴ g1⟭ into g_agg ·
    pserver_aggregate ps ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_dist.ydv")
	p := parser.New(l, "chk_dist.ydv")
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

func TestCheckerBarat(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ g0 ↦ tensor⟬⟬0.1⸴ 0.2⟭⸴ ⟬0.3⸴ 0.4⟭⟭ ·
    barat_submit 0 into b_grp with g0 ·
    barat_aggregate b_grp into agg_g cert into b_cert ·
    barat_verify b_cert into is_ok ·
    corrupt_barat b_cert at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_barat.ydv")
	p := parser.New(l, "chk_barat.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// 1. In non-test mode, corrupt_barat should fail
	c := New()
	c.IsTest = false
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for corrupt_barat in non-test mode, got none")
	}
	hasTestModeErr := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "corrupt_barat is only permitted in test mode") {
			hasTestModeErr = true
			break
		}
	}
	if !hasTestModeErr {
		t.Fatalf("expected test mode error message, got: %v", errs)
	}

	// 2. In test mode, corrupt_barat should pass
	cTest := New()
	cTest.IsTest = true
	errsTest := cTest.Check(prog)
	if len(errsTest) > 0 {
		t.Fatalf("unexpected checker errors in test mode: %v", errsTest)
	}
}

func TestCheckerDistBaratUndefinedSymbols(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    pserver_aggregate undef_ps ·
    barat_aggregate undef_grp into agg cert into c ·
    barat_verify undef_cert into is_ok ·
    corrupt_barat undef_cert2 at ↦ 0 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "chk_undef.ydv")
	p := parser.New(l, "chk_undef.ydv")
	prog := p.Parse()
	c := New()
	c.IsTest = true
	errs := c.Check(prog)
	if len(errs) != 4 {
		t.Fatalf("expected 4 undefined symbol errors, got %d: %v", len(errs), errs)
	}
}
