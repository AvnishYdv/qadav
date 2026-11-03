package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParsePipelineClaimAndRun(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    @multi_device pipeline_claim 4 stages micro_batch 8 into my_plan ·
    pipeline_run my_plan on in_tensor into out_tensor ·
    pipeline_claim 2 stages micro_batch 1 into default_plan ·
    pipeline_run default_plan on in_tensor into out_tensor ·
    yield 0 ·
⟧`
	l := lexer.New(src, "pipeline_test.ydv")
	p := New(l, "pipeline_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.PipelineClaimStmt)
	if !ok0 || !s0.IsAnnotated || s0.Stages != 4 || s0.MicroBatch != 8 || s0.Into != "my_plan" {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.PipelineRunStmt)
	if !ok1 || s1.Plan != "my_plan" || s1.Input == nil || s1.Output == nil {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
}

func TestParseCDAGARecordAttestVerify(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    cdaga_record from ⟬g0⸴ g1⸴ g2⟭ into my_rec ·
    cdaga_attest my_rec to train_chain ·
    cdaga_verify my_rec into is_ok ·
    corrupt_cdaga my_rec at ↦ 32 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "cdaga_test.ydv")
	p := New(l, "cdaga_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	s0, ok0 := proc.Body.Statements[0].(*ast.CDAGARecordStmt)
	if !ok0 || len(s0.GradList) != 3 || s0.Into != "my_rec" {
		t.Fatalf("unexpected stmt 0: %+v", proc.Body.Statements[0])
	}
	s1, ok1 := proc.Body.Statements[1].(*ast.CDAGAAttestStmt)
	if !ok1 || s1.Record != "my_rec" || s1.Chain != "train_chain" {
		t.Fatalf("unexpected stmt 1: %+v", proc.Body.Statements[1])
	}
	s2, ok2 := proc.Body.Statements[2].(*ast.CDAGAVerifyStmt)
	if !ok2 || s2.Record != "my_rec" || s2.Into != "is_ok" {
		t.Fatalf("unexpected stmt 2: %+v", proc.Body.Statements[2])
	}
	s3, ok3 := proc.Body.Statements[3].(*ast.CorruptCDAGAStmt)
	if !ok3 || s3.Record != "my_rec" || s3.Offset == nil {
		t.Fatalf("unexpected stmt 3: %+v", proc.Body.Statements[3])
	}
}

func TestParsePipelineAndCDAGALaTeX(t *testing.T) {
	src := `process main\group\egroup \yields int64 \block
    \multi_device \pipeline_claim 2 stages micro_batch 4 into p1 \stmt
    \pipeline_run p1 on in into out \stmt
    \cdaga_record from \bracket g0 \sep g1 \ebracket into rec1 \stmt
    \cdaga_attest rec1 to my_chain \stmt
    \cdaga_verify rec1 into ok1 \stmt
    \corrupt_cdaga rec1 at \bind 16 \stmt
    yield 0 \stmt
\eblock`
	l := lexer.New(src, "latex_test.ydv")
	p := New(l, "latex_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("latex parser errors: %v", p.Errors())
	}
	if len(prog.Decls) == 0 {
		t.Fatalf("expected decls, got 0")
	}
}
