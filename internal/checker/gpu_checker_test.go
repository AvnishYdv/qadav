package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func parseAndCheckGPUSource(src string, isTest bool) (*Checker, []Error) {
	l := lexer.New(src, "gpu_test.ydv")
	p := parser.New(l, "gpu_test.ydv")
	prog := p.Parse()
	c := New()
	c.IsTest = isTest
	errs := c.Check(prog)
	return c, errs
}

func TestGPUChecker_ValidStatements(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ b ↦ tensor⟬⟬2.0⸴ 3.0⟭⸴ ⟬4.0⸴ 5.0⟭⟭ ·

    ethical_boundary auth ⟦ rule safe ↦ 1.0 · ⟧
    reasoning_chain my_chain with auth ⟦ capacity ↦ 20 · ⟧

    @gpu gpu_matmul a ⸴ b into c audit my_chain ·
    gpu_softmax c into s audit my_chain ·
    @gpu gpu_layernorm s into ln audit my_chain ·

    agke_attest my_chain ⦅ "matmul_f64" ⸴ a ⸴ c ⦆ ·
    agke_verify my_chain into v_res ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckGPUSource(src, false)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestGPUChecker_InvalidOperands(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    int64 not_tensor ↦ 42 ·
    tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⟭ ·
    gpu_matmul not_tensor ⸴ a into c ·
    gpu_softmax not_tensor into s ·
    gpu_layernorm not_tensor into ln ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckGPUSource(src, false)
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors for non-tensor operands, got %d: %v", len(errs), errs)
	}
}

func TestGPUChecker_CorruptDigestTestGating(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary auth ⟦ rule safe ↦ 1.0 · ⟧
    reasoning_chain my_chain with auth ⟦ capacity ↦ 10 · ⟧
    corrupt_gpu_digest my_chain at ↦ 0 ·
    yield 0 ·
⟧`
	// Without --test -> error
	_, errsNoTest := parseAndCheckGPUSource(src, false)
	if len(errsNoTest) == 0 {
		t.Fatalf("expected error when corrupt_gpu_digest is used without --test")
	}

	// With --test -> success
	_, errsWithTest := parseAndCheckGPUSource(src, true)
	if len(errsWithTest) > 0 {
		t.Fatalf("unexpected error when corrupt_gpu_digest is used with --test: %v", errsWithTest)
	}
}

func TestGPUChecker_Expressions(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ b ↦ tensor⟬⟬2.0⸴ 3.0⟭⸴ ⟬4.0⸴ 5.0⟭⟭ ·
    ethical_boundary auth ⟦ rule safe ↦ 1.0 · ⟧
    reasoning_chain my_chain with auth ⟦ capacity ↦ 10 · ⟧

    tensor⟨float64⟩⟬2⸴ 2⟭ c ↦ gpu_matmul⦅a ⸴ b⦆ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ s ↦ gpu_softmax⦅c⦆ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ ln ↦ gpu_layernorm⦅s⦆ ·
    int64 v ↦ agke_verify my_chain ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckGPUSource(src, false)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors for GPU expressions: %v", errs)
	}
}

func TestGPUChecker_NilBranchesCoverage(t *testing.T) {
	c := New()
	c.checkGPUMatmulStmt(nil)
	c.checkGPUSoftmaxStmt(nil)
	c.checkGPULayernormStmt(nil)
	c.checkAGKEAttestStmt(nil)
	c.checkAGKEVerifyStmt(nil)
	c.checkCorruptGPUDigestStmt(nil)
	c.checkGPUMatmulExpr(nil)
	c.checkGPUSoftmaxExpr(nil)
	c.checkGPULayernormExpr(nil)
	c.checkAGKEVerifyExpr(nil)
}

