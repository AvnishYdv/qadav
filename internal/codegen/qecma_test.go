package codegen

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestQECMABundleFreed(t *testing.T) {
	src := `process test_qecma_free⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦
        capacity ↦ 50 ·
    ⟧
    text msg ↦ "sensor_text" ·
    modality m1 as text_input from msg ·
    cross_modal_reason rc deduce from ⟬m1⟭ to "decision" with 0.95 ·
    verify_chain rc ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test_qecma_free.ydv")
	p := parser.New(l, "test_qecma_free.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	chk := checker.New()
	errs := chk.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	gen := New(chk)
	llvmIR, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	createCalls := strings.Count(llvmIR, "call i8* @qadav_qecma_create()")
	freeCalls := strings.Count(llvmIR, "call void @qadav_qecma_free(")
	reasonCalls := strings.Count(llvmIR, "call i32 @qadav_qera_reason_cross_modal(")

	if createCalls == 0 {
		t.Fatalf("expected at least one qadav_qecma_create call")
	}
	if freeCalls != createCalls {
		t.Fatalf("bundle leak detected: create calls = %d, free calls = %d", createCalls, freeCalls)
	}
	if reasonCalls != createCalls {
		t.Fatalf("reasoning mismatch: reason calls = %d, create calls = %d", reasonCalls, createCalls)
	}
}

func TestQECMABundleFreedOnError(t *testing.T) {
	// Tests multiple consecutive cross-modal reasoning steps to verify each has its own cleanup
	src := `process test_multi_reason⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦
        capacity ↦ 50 ·
    ⟧
    text msg1 ↦ "data1" ·
    text msg2 ↦ "data2" ·
    modality m1 as text_input from msg1 ·
    modality m2 as text_input from msg2 ·
    cross_modal_reason rc deduce from ⟬m1⟭ to "out1" with 0.90 ·
    cross_modal_reason rc deduce from ⟬m2⟭ to "out2" with 0.95 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test_multi_reason.ydv")
	p := parser.New(l, "test_multi_reason.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	chk := checker.New()
	errs := chk.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	gen := New(chk)
	llvmIR, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	createCalls := strings.Count(llvmIR, "call i8* @qadav_qecma_create()")
	freeCalls := strings.Count(llvmIR, "call void @qadav_qecma_free(")

	if createCalls != 2 || freeCalls != 2 {
		t.Fatalf("expected 2 create and 2 free calls, got %d creates and %d frees", createCalls, freeCalls)
	}
}

