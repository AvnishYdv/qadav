package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseSandboxFullProgram(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		sandbox_create untrusted_sb with host_auth ⟦ capacity ↦ 15 · ⟧ ·
		sandbox_run untrusted_sb ⦅ tool_eval⦅"arg1"⸴ "arg2"⦆ ⦆ ·
		sandbox_verify untrusted_sb into res ·
		text att ↦ sandbox_verify untrusted_sb ·
		yield 0 ·
	⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected ProcessDecl, got %T", prog.Decls[0])
	}
	body := proc.Body.Statements
	if len(body) < 5 {
		t.Fatalf("expected at least 5 statements, got %d", len(body))
	}

	create, ok := body[0].(*ast.SandboxCreateStmt)
	if !ok || create.Name != "untrusted_sb" || create.Boundary != "host_auth" {
		t.Fatalf("unexpected create statement: %v", body[0])
	}

	run, ok := body[1].(*ast.SandboxRunStmt)
	if !ok || run.Sandbox != "untrusted_sb" || run.ToolName != "tool_eval" || len(run.Args) != 2 {
		t.Fatalf("unexpected run statement: %v", body[1])
	}

	vstmt, ok := body[2].(*ast.SandboxVerifyStmt)
	if !ok || vstmt.Sandbox != "untrusted_sb" || vstmt.Into != "res" {
		t.Fatalf("unexpected verify stmt: %v", body[2])
	}

	bind, ok := body[3].(*ast.BindStmt)
	if !ok {
		t.Fatalf("unexpected bind stmt: %v", body[3])
	}
	vexpr, ok := bind.Value.(*ast.SandboxVerifyExpr)
	if !ok || vexpr.Sandbox != "untrusted_sb" {
		t.Fatalf("unexpected verify expr: %v", bind.Value)
	}
}

func TestParseSandboxCreateVariants(t *testing.T) {
	// Without block
	src1 := `sandbox_create sb1 with b1 ·`
	l1 := lexer.New(src1, "test.ydv")
	p1 := New(l1, "test.ydv")
	s1 := p1.parseSandboxCreateStmt()
	if s1 == nil || s1.Name != "sb1" || s1.Boundary != "b1" {
		t.Fatalf("failed parse s1: %v", s1)
	}

	// With empty args in run
	src2 := `sandbox_run sb1 ⦅ tool_no_args⦅⦆ ⦆ ·`
	l2 := lexer.New(src2, "test.ydv")
	p2 := New(l2, "test.ydv")
	s2 := p2.parseSandboxRunStmt()
	if s2 == nil || len(s2.Args) != 0 {
		t.Fatalf("failed parse s2: %v", s2)
	}

	// Verify stmt without into
	src3 := `sandbox_verify sb1 ·`
	l3 := lexer.New(src3, "test.ydv")
	p3 := New(l3, "test.ydv")
	s3 := p3.parseSandboxVerifyStmt()
	if s3 == nil || s3.Into != "" {
		t.Fatalf("failed parse s3: %v", s3)
	}
}
