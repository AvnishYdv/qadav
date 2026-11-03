package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseTranspileToStmt(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text buf ↦ "" ·
    transpile_to "qiskit" into buf with size 2048 ·
    transpile_to "cirq" into buf with size 1024 options ⟦ optimize · reduce_depth · ⟧ ·
    transpile_to "ionq" into buf with size 512 ·
    transpile_to "qasm3" into buf with size 2048 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "transpile_test.ydv")
	p := New(l, "transpile_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	body := proc.Body.Statements
	if len(body) < 5 {
		t.Fatalf("expected at least 5 statements, got %d", len(body))
	}
	tr1, ok := body[1].(*ast.TranspileToStmt)
	if !ok || tr1.Target != "qiskit" || tr1.Into != "buf" {
		t.Fatalf("expected transpile_to qiskit, got %#v", body[1])
	}
	tr2, ok := body[2].(*ast.TranspileToStmt)
	if !ok || tr2.Target != "cirq" || tr2.Options == nil || !tr2.Options.Optimize || !tr2.Options.ReduceDepth {
		t.Fatalf("expected transpile_to cirq with options, got %#v", body[2])
	}
}

func TestParseEEAStmts(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve_policy constant into policy with rate 0.10 budget 100 ·
    evolve_policy decaying into policy2 with rate 0.20 budget 50 ·
    evolve_policy adaptive into policy3 with rate 0.30 budget 200 ·
    evolve_bounded policy from "a" to "b" with reason "tweak" max_flips 8 max_new_symbols 2 into prop env ·
    eea_verify env into is_valid ·
    corrupt_eea env at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "eea_test.ydv")
	p := New(l, "eea_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	body := proc.Body.Statements
	if len(body) < 6 {
		t.Fatalf("expected at least 6 statements, got %d", len(body))
	}
	p1, ok := body[0].(*ast.EvolvePolicyStmt)
	if !ok || p1.Kind != "constant" || p1.Into != "policy" {
		t.Fatalf("expected evolve_policy constant, got %#v", body[0])
	}
	p2, ok := body[1].(*ast.EvolvePolicyStmt)
	if !ok || p2.Kind != "decaying" || p2.Into != "policy2" {
		t.Fatalf("expected evolve_policy decaying, got %#v", body[1])
	}
	p3, ok := body[2].(*ast.EvolvePolicyStmt)
	if !ok || p3.Kind != "adaptive" || p3.Into != "policy3" {
		t.Fatalf("expected evolve_policy adaptive, got %#v", body[2])
	}
	eb, ok := body[3].(*ast.EvolveBoundedStmt)
	if !ok || eb.Policy != "policy" || eb.IntoProposal != "prop" || eb.IntoEnvelope != "env" {
		t.Fatalf("expected evolve_bounded, got %#v", body[3])
	}
	ev, ok := body[4].(*ast.EEAVerifyStmt)
	if !ok || ev.Envelope != "env" || ev.Into != "is_valid" {
		t.Fatalf("expected eea_verify, got %#v", body[4])
	}
	ce, ok := body[5].(*ast.CorruptEEAStmt)
	if !ok || ce.Envelope != "env" {
		t.Fatalf("expected corrupt_eea, got %#v", body[5])
	}
}
