package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func TestParseTensorTypeBare(t *testing.T) {
	src := `process f⦅tensor t⦆ ⟦ ⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	tt, ok := proc.Params[0].Type.(*ast.TensorType)
	if !ok || tt.Elem != nil || len(tt.Dims) != 0 {
		t.Fatalf("expected bare TensorType, got %+v", proc.Params[0].Type)
	}
}

func TestParseTensorTypeWithElem(t *testing.T) {
	src := `process f⦅tensor⟨float64⟩ t⦆ ⟦ ⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	tt, ok := proc.Params[0].Type.(*ast.TensorType)
	if !ok || tt.Elem == nil {
		t.Fatalf("expected TensorType with elem, got %+v", proc.Params[0].Type)
	}
	pt, ok := tt.Elem.(*ast.PrimitiveType)
	if !ok || pt.Name != "float64" {
		t.Fatalf("expected elem float64, got %+v", tt.Elem)
	}
}

func TestParseTensorTypeWithShape(t *testing.T) {
	src := `process f⦅tensor⟨float64⟩⟬3⸴ 4⟭ t⦆ ⟦ ⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	tt := proc.Params[0].Type.(*ast.TensorType)
	if len(tt.Dims) != 2 || tt.Dims[0] != 3 || tt.Dims[1] != 4 {
		t.Fatalf("expected dims [3, 4], got %v", tt.Dims)
	}
}

func TestParseTensorTypeWithShape1D(t *testing.T) {
	src := `process f⦅tensor⟨int64⟩⟬128⟭ t⦆ ⟦ ⟧`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	tt := proc.Params[0].Type.(*ast.TensorType)
	if len(tt.Dims) != 1 || tt.Dims[0] != 128 {
		t.Fatalf("expected dims [128], got %v", tt.Dims)
	}
}

func TestParseTensorLit1D(t *testing.T) {
	expr := parseExpr(t, `tensor⟬1.0⸴ 2.0⸴ 3.0⟭`)
	lit, ok := expr.(*ast.TensorLit)
	if !ok {
		t.Fatalf("expected *ast.TensorLit, got %T", expr)
	}
	if len(lit.Dims) != 1 || lit.Dims[0] != 3 {
		t.Fatalf("expected dims [3], got %v", lit.Dims)
	}
	if len(lit.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(lit.Items))
	}
}

func TestParseTensorLit2D(t *testing.T) {
	expr := parseExpr(t, `tensor⟬⟬1.0⸴ 2.0⟭⸴ ⟬3.0⸴ 4.0⟭⟭`)
	lit, ok := expr.(*ast.TensorLit)
	if !ok {
		t.Fatalf("expected *ast.TensorLit, got %T", expr)
	}
	if len(lit.Dims) != 2 || lit.Dims[0] != 2 || lit.Dims[1] != 2 {
		t.Fatalf("expected dims [2, 2], got %v", lit.Dims)
	}
	if len(lit.Items) != 2 {
		t.Fatalf("expected 2 outer items, got %d", len(lit.Items))
	}
}

func TestParseTensorLitMixedError(t *testing.T) {
	src := `tensor⟬1.0⸴ tensor⟬2.0⸴ 3.0⟭⟭`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	_ = p.parseExpression(LOWEST)
	if len(p.Errors()) == 0 {
		t.Fatalf("expected error for mixed scalar/tensor literal, got none")
	}
}

func TestParseTensorLitNonUniformError(t *testing.T) {
	src := `tensor⟬⟬1.0⟭⸴ ⟬2.0⸴ 3.0⟭⟭`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	_ = p.parseExpression(LOWEST)
	if len(p.Errors()) == 0 {
		t.Fatalf("expected error for non-uniform tensor literal, got none")
	}
}

func TestParseMatmulExpr(t *testing.T) {
	expr := parseExpr(t, "a ⊛ b")
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.MATMUL {
		t.Fatalf("expected BinaryExpr with MATMUL (⊛), got %+v", expr)
	}
}

func TestParseSumExpr(t *testing.T) {
	expr := parseExpr(t, "∑ t")
	un, ok := expr.(*ast.UnaryExpr)
	if !ok || un.Op != token.SUM {
		t.Fatalf("expected UnaryExpr with SUM (∑), got %+v", expr)
	}
}

func TestParseProductExpr(t *testing.T) {
	expr := parseExpr(t, "∏ t")
	un, ok := expr.(*ast.UnaryExpr)
	if !ok || un.Op != token.PRODUCT {
		t.Fatalf("expected UnaryExpr with PRODUCT (∏), got %+v", expr)
	}
}

func TestParseTransposeExpr(t *testing.T) {
	expr := parseExpr(t, "⋆ t")
	un, ok := expr.(*ast.UnaryExpr)
	if !ok || un.Op != token.ADJOINT {
		t.Fatalf("expected UnaryExpr with ADJOINT (⋆), got %+v", expr)
	}
}

func TestParseTensorBind(t *testing.T) {
	src := `tensor⟨float64⟩⟬3⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭·`
	stmt := parseStmt(t, src)
	bind, ok := stmt.(*ast.BindStmt)
	if !ok {
		t.Fatalf("expected *ast.BindStmt, got %T", stmt)
	}
	tt, ok := bind.Type.(*ast.TensorType)
	if !ok || len(tt.Dims) != 1 || tt.Dims[0] != 3 {
		t.Fatalf("unexpected bind type: %+v", bind.Type)
	}
}

func TestParseTensorProcess(t *testing.T) {
	src := `process dot⦅tensor⟨float64⟩⟬3⟭ a⸴ tensor⟨float64⟩⟬3⟭ b⦆ → float64 ⟦
    yield ∑ ⦅a ⊗ b⦆·
⟧`
	l := lexer.New(src, "dot.ydv")
	p := New(l, "dot.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}
}

func TestParseTensorPrecedence(t *testing.T) {
	expr := parseExpr(t, "a ⊛ b ⊕ c")
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.OSUM {
		t.Fatalf("expected outer OSUM (⊕), got %+v", expr)
	}
	leftBin, ok := bin.Left.(*ast.BinaryExpr)
	if !ok || leftBin.Op != token.MATMUL {
		t.Fatalf("expected inner Left MATMUL (⊛), got %+v", bin.Left)
	}
}

func TestParseTensorLiteralLaTeXFallback(t *testing.T) {
	src := `tensor\langle float64\rangle\bracket 3\ebracket v \bind tensor\bracket 1.0\sep 2.0\sep 3.0\ebracket\stmt`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	stmt := p.parseStatement()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors with LaTeX fallbacks: %v", p.Errors())
	}
	bind, ok := stmt.(*ast.BindStmt)
	if !ok {
		t.Fatalf("expected BindStmt, got %T", stmt)
	}
	tt, ok := bind.Type.(*ast.TensorType)
	if !ok || len(tt.Dims) != 1 || tt.Dims[0] != 3 {
		t.Fatalf("expected TensorType with dims [3], got %+v", bind.Type)
	}
}
