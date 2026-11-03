package parser

import (
	"os"
	"testing"

	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseQECDemo(t *testing.T) {
	src, err := os.ReadFile("../../examples/qec_demo.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "qec_demo.ydv")
	p := New(l, "qec_demo.ydv")
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

func TestParseQECTamper(t *testing.T) {
	src, err := os.ReadFile("../../examples/qec_tamper.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "qec_tamper.ydv")
	p := New(l, "qec_tamper.ydv")
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

func TestParseQECUnverifiedCorrect(t *testing.T) {
	src, err := os.ReadFile("../../examples/qec_unverified_correct.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "qec_unverified_correct.ydv")
	p := New(l, "qec_unverified_correct.ydv")
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

func TestParseConsolidationDemo(t *testing.T) {
	src, err := os.ReadFile("../../examples/consolidation_demo.ydv")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	l := lexer.New(string(src), "consolidation_demo.ydv")
	p := New(l, "consolidation_demo.ydv")
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

func TestParseQECVariations(t *testing.T) {
	src := `process test_qec_var⦅⦆ → int64 ⟦
    qec_memory qm1 as surface_code distance ↦ 5 ⟦ ⟧ ·
    qec_memory qm2 as qldpc_code distance ↦ 7 with auth ⟦ ⟧ ·
    syndrome qm1 into s1 ·
    verify_syndrome qm1 with s1 ·
    correct qm1 with s1 ·
    corrupt_syndrome qm1⟬s1⟭ at ↦ 4 ·
    consolidate ep into sem with threshold ↦ 0.85 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "qec_var.ydv")
	p := New(l, "qec_var.ydv")
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

func TestParseQECErrors(t *testing.T) {
	cases := []string{
		`process p⦅⦆ → int64 ⟦ qec_memory · ⟧`,
		`process p⦅⦆ → int64 ⟦ qec_memory qm · ⟧`,
		`process p⦅⦆ → int64 ⟦ qec_memory qm as · ⟧`,
		`process p⦅⦆ → int64 ⟦ syndrome · ⟧`,
		`process p⦅⦆ → int64 ⟦ syndrome qm into · ⟧`,
		`process p⦅⦆ → int64 ⟦ verify_syndrome · ⟧`,
		`process p⦅⦆ → int64 ⟦ verify_syndrome qm · ⟧`,
		`process p⦅⦆ → int64 ⟦ correct · ⟧`,
		`process p⦅⦆ → int64 ⟦ correct qm · ⟧`,
		`process p⦅⦆ → int64 ⟦ corrupt_syndrome · ⟧`,
		`process p⦅⦆ → int64 ⟦ consolidate · ⟧`,
		`process p⦅⦆ → int64 ⟦ consolidate ep · ⟧`,
		`process p⦅⦆ → int64 ⟦ consolidate ep into · ⟧`,
	}
	for _, src := range cases {
		l := lexer.New(src, "test_err.ydv")
		p := New(l, "test_err.ydv")
		_ = p.Parse()
		if len(p.Errors()) == 0 {
			t.Errorf("expected errors for %q, got none", src)
		}
	}
}
