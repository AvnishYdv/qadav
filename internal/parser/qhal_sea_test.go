package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseQHALExport(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @qhal_backend sim
    qasm_export into qasm_out with size 1024 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "qhal_test.ydv")
	p := New(l, "qhal_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.QHALBackendAnnotation)
	if !ok0 || s0.Kind != "sim" {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.QASMExportStmt)
	if !ok1 || s1.Into != "qasm_out" || s1.BufLen == nil {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
}

func TestParseSEAStatements(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    evolve from s1 to s2 with reason r into prop ·
    evolve_history to evo_log ·
    evolve_verify prop into is_valid ·
    corrupt_evolve prop at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "sea_test.ydv")
	p := New(l, "sea_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.EvolveStmt)
	if !ok0 || s0.Into != "prop" || s0.FromState == nil || s0.ToState == nil || s0.Reason == nil {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.EvolveHistoryStmt)
	if !ok1 || s1.Chain != "evo_log" {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
	s2, ok2 := proc.Body.Statements[2].(*ast.EvolveVerifyStmt)
	if !ok2 || s2.Proposals != "prop" || s2.Into != "is_valid" {
		t.Fatalf("unexpected stmt 2: %+v", proc.Body.Statements[2])
	}
	s3, ok3 := proc.Body.Statements[3].(*ast.CorruptEvolveStmt)
	if !ok3 || s3.Proposal != "prop" || s3.Offset == nil {
		t.Fatalf("unexpected stmt 3: %+v", proc.Body.Statements[3])
	}
}
