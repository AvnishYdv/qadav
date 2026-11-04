package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParsePCCStatements(t *testing.T) {
	input := `
process main⦅⦆ → int64 ⟦
    pcc_attest ast ↦ "ast_hello" codegen ↦ "codegen_path" ir ↦ "llvm_ir" into cert0 ·
    pcc_verify cert0 into v0 ·
    corrupt_pcc cert0 field ↦ 1 ·
    yield 0 ·
⟧
`
	l := lexer.New(input, "pcc_parse_test.ydv")
	p := New(l, "pcc_parse_test.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}

	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected ProcessDecl, got %T", prog.Decls[0])
	}

	if len(proc.Body.Statements) != 4 {
		t.Fatalf("expected 4 statements in body, got %d", len(proc.Body.Statements))
	}

	attest, ok := proc.Body.Statements[0].(*ast.PCCAttestStmt)
	if !ok || attest.Ast != "ast_hello" || attest.Codegen != "codegen_path" || attest.IR != "llvm_ir" || attest.Into != "cert0" {
		t.Fatalf("unexpected PCCAttestStmt: %+v", attest)
	}

	verify, ok := proc.Body.Statements[1].(*ast.PCCVerifyStmt)
	if !ok || verify.Cert != "cert0" || verify.Into != "v0" {
		t.Fatalf("unexpected PCCVerifyStmt: %+v", verify)
	}

	corrupt, ok := proc.Body.Statements[2].(*ast.CorruptPCCStmt)
	if !ok || corrupt.Cert != "cert0" {
		t.Fatalf("unexpected CorruptPCCStmt: %+v", corrupt)
	}
}
