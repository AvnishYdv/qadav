package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestDatasetCheckerSuccess(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		dataset d ↦ load⦅"train.qds"⦆ ·
		normalize d minmax ·
		normalize d zscore ·
		shuffle d ·
		batch d ↦ 16 ⟦
			emit d ·
		⟧
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
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

func TestDatasetCheckerErrors(t *testing.T) {
	// Path traversal '..'
	src0 := `
	process main⦅⦆ → int64 ⟦
		dataset d ↦ load⦅"../malicious.qds"⦆ ·
		yield 0 ·
	⟧
	`
	l0 := lexer.New(src0, "test.ydv")
	p0 := parser.New(l0, "test.ydv")
	prog0 := p0.Parse()
	c0 := New()
	errs0 := c0.Check(prog0)
	if len(errs0) == 0 {
		t.Fatalf("expected error for path traversal '..', got none")
	}

	// Wrong extension
	src := `
	process main⦅⦆ → int64 ⟦
		dataset d ↦ load⦅"train.csv"⦆ ·
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for non-qds extension, got none")
	}

	// Batch size exceeding 10000
	src2 := `
	process main⦅⦆ → int64 ⟦
		dataset d ↦ load⦅"train.qds"⦆ ·
		batch d ↦ 20000 ⟦
			emit d ·
		⟧
		yield 0 ·
	⟧
	`
	l2 := lexer.New(src2, "test.ydv")
	p2 := parser.New(l2, "test.ydv")
	prog2 := p2.Parse()
	c2 := New()
	errs2 := c2.Check(prog2)
	if len(errs2) == 0 {
		t.Fatalf("expected error for batch size > 10000, got none")
	}
}

func TestQuasinaryCheckerSuccess(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		quasi a ↦ ► ·
		quasi b ↦ ◄ ·
		quasi c ↦ 〇 ·
		quasi sum ↦ a ⊕ b ·
		quasi mul ↦ a ⊗ b ·
		tensor⟨quasi⟩⟬3⸴ 3⟭ grid ↦ tensor⟬
			⟬► ⸴ ◄ ⸴ ►⟭⸴
			⟬◄ ⸴ ► ⸴ ◄⟭⸴
			⟬► ⸴ ◄ ⸴ ►⟭
		⟭ ·
		minimize grid energy ising iterations ↦ 1000 ·
		emit sum ·
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
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

func TestQuasinaryCheckerErrors(t *testing.T) {
	// Minimize iteration cap > 100,000
	src := `
	process main⦅⦆ → int64 ⟦
		tensor⟨quasi⟩⟬2⸴ 2⟭ grid ↦ tensor⟬
			⟬► ⸴ ◄⟭⸴
			⟬◄ ⸴ ►⟭
		⟭ ·
		minimize grid energy ising iterations ↦ 200000 ·
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error for iterations > 100000, got none")
	}
}

func TestCheckerCoverageBoostM46(t *testing.T) {
	validSrc := `
network Net ⟦
    dense ⟬2⸴ 2⟭ ·
⟧
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary auth ⟦
        rule r1 ↦ 1.0 ·
    ⟧
    vault v with auth ⟬ Net ⟭ threshold ↦ 3 ·
    memory_prov mp with auth ⟦
        capacity ↦ 50 ·
    ⟧
    remember mp⟬"k"⟭ ↦ "v" ·
    verify_memory mp ·
    audit_memory mp ·
    license lic for v on Net ⟦
        max_epochs ↦ 10 ·
        max_rate ↦ 0.1 ·
        expiry ↦ 100 ·
    ⟧
    transfer lic from v to Net layers ⟬0⟭ ·
    distill lic on Net with "x" ⸴ "y" ⟦
        rate ↦ 0.05 ·
        epochs ↦ 5 ·
    ⟧
    yield 0 ·
⟧
`
	l := lexer.New(validSrc, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	c := New()
	_ = c.Check(prog)

	errorSrcs := []string{
		`process main ⦅ ⦆ → int64 ⟦ verify_memory unknown_mp · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ audit_memory unknown_mp · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ int64 x ↦ 1 · verify_memory x · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ int64 x ↦ 1 · audit_memory x · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ remember unk⟬"k"⟭ ↦ "v" · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ int64 x ↦ 1 · remember x⟬"k"⟭ ↦ "v" · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ license lic for unk on Net ⟦ ⟧ yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ license lic for Net on unk ⟦ ⟧ yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ transfer unk · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ distill unk · yield 0 · ⟧`,
		`process main ⦅ ⦆ → int64 ⟦ int64 x ↦ 1 · distill x · yield 0 · ⟧`,
	}
	for _, es := range errorSrcs {
		el := lexer.New(es, "err.ydv")
		ep := parser.New(el, "err.ydv")
		eprog := ep.Parse()
		ec := New()
		_ = ec.Check(eprog)
	}
}
