package parser

import (
	"os"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseQERADemo(t *testing.T) {
	src, err := os.ReadFile("../../examples/qera_demo.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "qera_demo.ydv")
	p := New(l, "qera_demo.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %v", e)
		}
	}
	if prog == nil {
		t.Fatalf("prog is nil")
	}
}

func TestParseQERATamper(t *testing.T) {
	src, err := os.ReadFile("../../examples/qera_tamper.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "qera_tamper.ydv")
	p := New(l, "qera_tamper.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %v", e)
		}
	}
	if prog == nil {
		t.Fatalf("prog is nil")
	}
}

func TestParseQECMADeclAndReason(t *testing.T) {
	src := `process test_qecma⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain c with b ⟦
        capacity ↦ 50 ·
    ⟧
    modality m_tensor as tensor_input from t1 ·
    modality m_text as text_input from txt1 ·
    modality m_concept as concept_input from c1 ·
    cross_modal_reason c deduce from ⟬m_tensor⸴ m_text⸴ m_concept⟭ to out1 with 0.95 ·
    cross_modal_reason c induce from ⟬m_tensor⸴ m_text⟭ to out2 with 0.90 ·
    cross_modal_reason c abduce from ⟬m_concept⟭ to out3 with 0.85 ·
    cross_modal_reason c verify from ⟬m_tensor⟭ to out4 with 0.99 ·
    verify_chain c ·
    audit_reasoning c ·
    corrupt_reasoning_step c⟬0⟭ at ↦ 99 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "qecma_inline.ydv")
	p := New(l, "qecma_inline.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %v", e)
		}
	}
	if prog == nil {
		t.Fatalf("prog is nil")
	}
}

func TestParseQECMABranchCoverage(t *testing.T) {
	src := `process test_qecma_alt⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain c with b ⟦
        capacity ↦ 50 ·
    ⟧
    modality m1 as tensor_input from x ·
    modality m2 as text_input from y ·
    modality m3 as concept_input from z ·
    cross_modal_reason c deduce from ⟬m1⸴ m2⟭ to res with 0.9 ·
    corrupt_reasoning_step c⟬1⟭ at ↦ 42 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "qecma_alt.ydv")
	p := New(l, "qecma_alt.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			t.Errorf("parse error: %v", e)
		}
	}
	if prog == nil {
		t.Fatalf("prog is nil")
	}
}
