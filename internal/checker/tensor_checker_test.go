package checker

import (
	"fmt"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/types"
)

func parseProgAndCheck(src string) (*Checker, *ast.Program, []Error) {
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	c := New()
	errs := c.Check(prog)
	return c, prog, errs
}

func TestCheckTensorLit1D(t *testing.T) {
	src := `process main⦅⦆ ⟦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭· ⟧`
	c, prog, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	stmt := proc.Body.Statements[0].(*ast.ExprStmt)
	typ := c.ExprType(stmt.Expr)
	if !typ.IsTensor() || typ.Rank() != 1 || typ.Shape[0] != 3 {
		t.Fatalf("expected 1D tensor of shape [3], got %v", typ)
	}
	if typ.TensorElem() != types.Float64 {
		t.Fatalf("expected elem float64, got %v", typ.TensorElem())
	}
}

func TestCheckTensorLit2D(t *testing.T) {
	src := `process main⦅⦆ ⟦ tensor⟬⟬1.0⸴ 2.0⸴ 3.0⟭⸴ ⟬4.0⸴ 5.0⸴ 6.0⟭⟭· ⟧`
	c, prog, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	stmt := proc.Body.Statements[0].(*ast.ExprStmt)
	typ := c.ExprType(stmt.Expr)
	if !typ.IsTensor() || typ.Rank() != 2 || typ.Shape[0] != 2 || typ.Shape[1] != 3 {
		t.Fatalf("expected 2D tensor of shape [2, 3], got %v", typ)
	}
}

func TestCheckTensorBind(t *testing.T) {
	src := `process main⦅⦆ ⟦ tensor⟨float64⟩⟬3⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭· ⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckTensorBindShapeMismatch(t *testing.T) {
	src := `process main⦅⦆ ⟦ tensor⟨float64⟩⟬4⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭· ⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for shape mismatch [4] vs [3], got none")
	}
}

func TestCheckTensorElemMismatch(t *testing.T) {
	src := `process main⦅⦆ ⟦ tensor⟨int64⟩⟬3⟭ v ↦ tensor⟬1.5⸴ 2.5⸴ 3.5⟭· ⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for binding float tensor to int tensor, got none")
	}
}

func TestCheckTensorAddSameShape(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⟭ a⸴ tensor⟨float64⟩⟬3⟭ b⦆ ⟦
    tensor⟨float64⟩⟬3⟭ c ↦ a ⊕ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckTensorAddBroadcast(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⟭ a⦆ ⟦
    tensor⟨float64⟩⟬3⟭ c ↦ a ⊕ 1.0·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckTensorAddShapeMismatch(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ ⟦
    a ⊕ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for shape mismatch [3] vs [4], got none")
	}
}

func TestCheckTensorPromoteElem(t *testing.T) {
	src := `process main⦅tensor⟨int64⟩⟬3⟭ a⦆ ⟦
    tensor⟨float64⟩⟬3⟭ c ↦ a ⊗ 2.5·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckMatmul2D2D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⸴ 4⟭ a⸴ tensor⟨float64⟩⟬4⸴ 5⟭ b⦆ ⟦
    tensor⟨float64⟩⟬3⸴ 5⟭ c ↦ a ⊛ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckMatmul2D2DMismatch(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⸴ 4⟭ a⸴ tensor⟨float64⟩⟬5⸴ 6⟭ b⦆ ⟦
    a ⊛ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for matmul inner dim mismatch 4 vs 5, got none")
	}
}

func TestCheckMatmul1D2D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⸴ 5⟭ b⦆ ⟦
    tensor⟨float64⟩⟬5⟭ c ↦ a ⊛ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckMatmul2D1D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⸴ 4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ ⟦
    tensor⟨float64⟩⟬3⟭ c ↦ a ⊛ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckMatmul1D1D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬4⟭ a⸴ tensor⟨float64⟩⟬4⟭ b⦆ ⟦
    float64 c ↦ a ⊛ b·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckSumReduce(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⟭ a⦆ ⟦
    float64 s ↦ ∑ a·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckProductReduce(t *testing.T) {
	src := `process main⦅tensor⟨int64⟩⟬3⟭ a⦆ ⟦
    int64 p ↦ ∏ a·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckSumNonTensor(t *testing.T) {
	src := `process main⦅⦆ ⟦ ∑ 5· ⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for ∑ on non-tensor, got none")
	}
}

func TestCheckTranspose2D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⸴ 4⟭ a⦆ ⟦
    tensor⟨float64⟩⟬4⸴ 3⟭ b ↦ ⋆ a·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckTranspose1D(t *testing.T) {
	src := `process main⦅tensor⟨float64⟩⟬3⟭ a⦆ ⟦
    ⋆ a·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error for ⋆ on 1D tensor, got none")
	}
}

func TestCheckTensorProcessParams(t *testing.T) {
	src := `process dot⦅tensor⟨float64⟩⟬3⟭ a⸴ tensor⟨float64⟩⟬3⟭ b⦆ → float64 ⟦
    yield ∑ ⦅a ⊗ b⦆·
⟧

process main⦅⦆ → int64 ⟦
    tensor⟨float64⟩⟬3⟭ x ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭·
    tensor⟨float64⟩⟬3⟭ y ↦ tensor⟬4.0⸴ 5.0⸴ 6.0⟭·
    float64 d ↦ dot⦅x⸴ y⦆·
    emit d·
    yield 0·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckEmitTensor(t *testing.T) {
	src := `process main⦅⦆ ⟦
    tensor⟨float64⟩⟬3⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭·
    emit v·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestCheckFullTensorProgram(t *testing.T) {
	src := `process main⦅⦆ ⟦
    tensor⟨int64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1⸴ 2⟭⸴ ⟬3⸴ 4⟭⟭·
    tensor⟨int64⟩⟬2⸴ 2⟭ b ↦ tensor⟬⟬5⸴ 6⟭⸴ ⟬7⸴ 8⟭⟭·
    tensor⟨int64⟩⟬2⸴ 2⟭ c ↦ a ⊛ b·
    tensor⟨int64⟩⟬2⸴ 2⟭ d ↦ ⋆ c·
    int64 s ↦ ∑ d·
    emit s·
⟧`
	_, _, errs := parseProgAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func BenchmarkTensorCheck(b *testing.B) {
	src := `process bench⦅⦆ ⟦
`
	for i := 0; i < 20; i++ {
		src += fmt.Sprintf("    tensor⟨float64⟩⟬2⸴ 2⟭ t%d ↦ tensor⟬⟬1.0⸴ 2.0⟭⸴ ⟬3.0⸴ 4.0⟭⟭·\n", i)
	}
	src += "⟧"

	l := lexer.New(src, "bench.ydv")
	p := parser.New(l, "bench.ydv")
	prog := p.Parse()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := New()
		_ = c.Check(prog)
	}
}
