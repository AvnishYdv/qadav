package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseRecordDecl(t *testing.T) {
	src := `record Token ⟦ kind ↦ text · value ↦ text · line ↦ int64 · col ↦ int64 · ⟧`
	l := lexer.New(src, "test_record.ydv")
	p := New(l, "test_record.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}

	rec, ok := prog.Decls[0].(*ast.RecordDeclStmt)
	if !ok {
		t.Fatalf("expected *ast.RecordDeclStmt, got %T", prog.Decls[0])
	}
	if rec.Name != "Token" {
		t.Fatalf("expected record name Token, got %s", rec.Name)
	}
	if len(rec.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(rec.Fields))
	}
	if rec.Fields[0].Name != "kind" || rec.Fields[0].Type != "text" {
		t.Fatalf("expected field kind:text, got %s:%s", rec.Fields[0].Name, rec.Fields[0].Type)
	}
}

func TestParseVariantDecl(t *testing.T) {
	src := `variant Expr ⟦ | Num(val ↦ int64) · | Identifier(name ↦ text) · ⟧`
	l := lexer.New(src, "test_variant.ydv")
	p := New(l, "test_variant.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}

	vdecl, ok := prog.Decls[0].(*ast.VariantDeclStmt)
	if !ok {
		t.Fatalf("expected *ast.VariantDeclStmt, got %T", prog.Decls[0])
	}
	if vdecl.Name != "Expr" {
		t.Fatalf("expected variant name Expr, got %s", vdecl.Name)
	}
	if len(vdecl.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(vdecl.Cases))
	}
	if vdecl.Cases[0].Name != "Num" || len(vdecl.Cases[0].Fields) != 1 {
		t.Fatalf("expected case Num with 1 field, got %s with %d fields", vdecl.Cases[0].Name, len(vdecl.Cases[0].Fields))
	}
}

func TestParseMatchStmt(t *testing.T) {
	src := `process test⦅⦆ → int64 ⟦
    match e ⟦
        | Num(x) ⟹ yield x ·
        | Identifier(s) ⟹ yield 0 ·
        otherwise ⟹ yield 99 ·
    ⟧
⟧`
	l := lexer.New(src, "test_match.ydv")
	p := New(l, "test_match.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	fn := prog.Decls[0].(*ast.ProcessDecl)
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement in body, got %d", len(fn.Body.Statements))
	}
	m, ok := fn.Body.Statements[0].(*ast.MatchStmt)
	if !ok {
		t.Fatalf("expected *ast.MatchStmt, got %T", fn.Body.Statements[0])
	}
	if len(m.Arms) != 2 {
		t.Fatalf("expected 2 arms, got %d", len(m.Arms))
	}
	if m.Otherwise == nil {
		t.Fatalf("expected otherwise block, got nil")
	}
}

func TestParseRecordInitAndFieldAccess(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    Token tok ↦ record_init Token ⟦ kind ↦ "IDENT" ⸴ value ↦ "hello" ⸴ line ↦ 42 ⸴ col ↦ 1 · ⟧ ·
    text k ↦ tok.kind ·
    int64 l ↦ tok.line ·
    yield l ·
⟧`
	l := lexer.New(src, "test_init.ydv")
	p := New(l, "test_init.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	fn := prog.Decls[0].(*ast.ProcessDecl)
	if len(fn.Body.Statements) != 4 {
		t.Fatalf("expected 4 statements in body, got %d", len(fn.Body.Statements))
	}
}

func TestParseRSCAStatements(t *testing.T) {
	src := `process test_rsca⦅⦆ → int64 ⟦
    rsca_attest stage ↦ 0 from source ↦ "src" compiler ↦ "go" output ↦ "stage1" into rec0 ·
    rsca_verify rec0 into v0 ·
    corrupt_rsca rec0 at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test_rsca.ydv")
	p := New(l, "test_rsca.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	fn := prog.Decls[0].(*ast.ProcessDecl)
	if len(fn.Body.Statements) != 4 {
		t.Fatalf("expected 4 statements in body, got %d", len(fn.Body.Statements))
	}
}

func TestParseVariantInit(t *testing.T) {
	src := `process test_vinit⦅⦆ → int64 ⟦
    Expr e1 ↦ variant_init Expr ⦅Num ↦ 42⦆ ·
    Expr e2 ↦ variant_init Expr ⦅Identifier ↦ "x"⦆ ·
    yield 0 ·
⟧`
	l := lexer.New(src, "test_vinit.ydv")
	p := New(l, "test_vinit.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	fn := prog.Decls[0].(*ast.ProcessDecl)
	if len(fn.Body.Statements) != 3 {
		t.Fatalf("expected 3 statements in body, got %d", len(fn.Body.Statements))
	}
	bind1 := fn.Body.Statements[0].(*ast.BindStmt)
	vinit, ok := bind1.Value.(*ast.VariantInitExpr)
	if !ok {
		t.Fatalf("expected VariantInitExpr, got %T", bind1.Value)
	}
	if vinit.VariantName != "Expr" || vinit.CaseName != "Num" {
		t.Fatalf("unexpected VariantInitExpr: %+v", vinit)
	}
}

func TestParseDistributedAnnotation(t *testing.T) {
	src := `process test_dist⦅⦆ → int64 ⟦
    @distributed yield 42 ·
⟧`
	l := lexer.New(src, "test_dist.ydv")
	p := New(l, "test_dist.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}
