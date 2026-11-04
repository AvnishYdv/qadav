package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseCTSAStatements(t *testing.T) {
	input := `process main⦅⦆ → int64 ⟦
    ctsa_attest ast ↦ "ast_tree" env ↦ "type_env" rules ↦ "type_rules" into cert0 ·
    ctsa_verify cert0 into v0 ·
    corrupt_ctsa cert0 field ↦ 1 ·
    yield 0 ·
⟧`
	l := lexer.New(input, "ctsa_test.ydv")
	p := New(l, "ctsa_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse failed: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	fn := prog.Decls[0].(*ast.ProcessDecl)
	if len(fn.Body.Statements) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(fn.Body.Statements))
	}
}
