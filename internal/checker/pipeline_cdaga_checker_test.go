package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func parseAndCheckSource(src string, isTest bool) (*Checker, []Error) {
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	c := New()
	c.IsTest = isTest
	errs := c.Check(prog)
	return c, errs
}

func TestPipelineChecker_Valid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⟭ ·
    @multi_device pipeline_claim 2 stages micro_batch 4 into my_plan ·
    pipeline_run my_plan on a into out_tensor ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckSource(src, false)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestPipelineChecker_InvalidStagesAndMicroBatch(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    pipeline_claim 0 stages micro_batch 0 into bad_plan ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckSource(src, false)
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 errors for invalid stages and micro_batch, got %d: %v", len(errs), errs)
	}
}

func TestPipelineChecker_UndefinedPlan(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⟭ ·
    pipeline_run ghost_plan on a into out_tensor ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckSource(src, false)
	if len(errs) == 0 {
		t.Fatalf("expected error for undefined pipeline plan, got none")
	}
}

func TestCDAGAChecker_Valid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary auth ⟦ rule safe ↦ 1.0 · ⟧
    reasoning_chain my_chain with auth ⟦ capacity ↦ 20 · ⟧
    tensor⟨float64⟩⟬2⸴ 2⟭ g0 ↦ tensor⟬⟬0.01⸴ 0.02⟭⸴ ⟬0.03⸴ 0.04⟭⟭ ·
    tensor⟨float64⟩⟬2⸴ 2⟭ g1 ↦ tensor⟬⟬0.05⸴ 0.06⟭⸴ ⟬0.07⸴ 0.08⟭⟭ ·

    cdaga_record from ⟬g0⸴ g1⟭ into rec ·
    cdaga_attest rec to my_chain ·
    cdaga_verify rec into is_valid ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckSource(src, false)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestCDAGAChecker_UndefinedRecordAndChain(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    cdaga_attest no_rec to no_chain ·
    cdaga_verify no_rec into is_valid ·
    yield 0 ·
⟧`
	_, errs := parseAndCheckSource(src, false)
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors for undefined record/chain, got %d: %v", len(errs), errs)
	}
}

func TestCDAGAChecker_CorruptTestGating(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary auth ⟦ rule safe ↦ 1.0 · ⟧
    reasoning_chain my_chain with auth ⟦ capacity ↦ 20 · ⟧
    tensor⟨float64⟩⟬2⸴ 2⟭ g0 ↦ tensor⟬⟬0.01⸴ 0.02⟭⸴ ⟬0.03⸴ 0.04⟭⟭ ·
    cdaga_record from ⟬g0⟭ into rec ·
    corrupt_cdaga rec at ↦ 10 ·
    yield 0 ·
⟧`
	// In production mode (isTest = false), corrupt_cdaga must be rejected
	_, errs := parseAndCheckSource(src, false)
	if len(errs) == 0 {
		t.Fatalf("expected error for corrupt_cdaga in production mode, got none")
	}

	// In test mode (isTest = true), it must be accepted
	_, errsTest := parseAndCheckSource(src, true)
	if len(errsTest) > 0 {
		t.Fatalf("unexpected checker errors in test mode: %v", errsTest)
	}
}
