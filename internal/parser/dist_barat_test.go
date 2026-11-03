package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseDistributedStatements(t *testing.T) {
	src := `@distributed
process main⦅⦆ → int64 ⟦
    pserver_init into ps with 4 workers ·
    all_reduce from ⟬g0⸴ g1⟭ into g_agg ·
    all_reduce g_buf ·
    pserver_aggregate ps ·
    yield 0 ·
⟧`
	l := lexer.New(src, "dist_test.ydv")
	p := New(l, "dist_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.PserverInitStmt)
	if !ok0 || s0.Into != "ps" || s0.Workers == nil {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.AllReduceStmt)
	if !ok1 || len(s1.GradList) != 2 || s1.Into != "g_agg" {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
	s2, ok2 := proc.Body.Statements[2].(*ast.AllReduceStmt)
	if !ok2 || s2.Target == nil {
		t.Fatalf("unexpected stmt 2: %+v", proc.Body.Statements[2])
	}
	s3, ok3 := proc.Body.Statements[3].(*ast.PserverAggregateStmt)
	if !ok3 || s3.Pserver != "ps" {
		t.Fatalf("unexpected stmt 3: %+v", proc.Body.Statements[3])
	}
}

func TestParseBaratStatements(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    barat_submit 0 into b_grp with grad0 ·
    barat_submit from grad1 into b_grp ·
    barat_aggregate b_grp into agg_g cert into b_cert ·
    barat_verify b_cert into is_ok ·
    corrupt_barat b_cert at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "barat_parse_test.ydv")
	p := New(l, "barat_parse_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.BaratSubmitStmt)
	if !ok0 || s0.Group != "b_grp" || s0.Worker == nil || s0.Gradient == nil {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.BaratSubmitStmt)
	if !ok1 || s1.Group != "b_grp" || s1.Gradient == nil {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
	s2, ok2 := proc.Body.Statements[2].(*ast.BaratAggregateStmt)
	if !ok2 || s2.Group != "b_grp" || s2.IntoGrad != "agg_g" || s2.IntoCert != "b_cert" {
		t.Fatalf("unexpected stmt 2: %+v", proc.Body.Statements[2])
	}
	s3, ok3 := proc.Body.Statements[3].(*ast.BaratVerifyStmt)
	if !ok3 || s3.Cert != "b_cert" || s3.Into != "is_ok" {
		t.Fatalf("unexpected stmt 3: %+v", proc.Body.Statements[3])
	}
	s4, ok4 := proc.Body.Statements[4].(*ast.CorruptBaratStmt)
	if !ok4 || s4.Cert != "b_cert" || s4.Offset == nil {
		t.Fatalf("unexpected stmt 4: %+v", proc.Body.Statements[4])
	}
}

