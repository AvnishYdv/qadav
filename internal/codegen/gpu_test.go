package codegen

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
)

func TestGPUCodegen(t *testing.T) {
	g := New(nil)

	stmt := &ast.GPUMatmulStmt{
		Left:   &ast.Ident{Name: "a"},
		Right:  &ast.Ident{Name: "b"},
		Target: &ast.Ident{Name: "c"},
	}
	g.genGPUMatmulStmt(stmt)

	sstmt := &ast.GPUSoftmaxStmt{
		Source: &ast.Ident{Name: "a"},
		Target: &ast.Ident{Name: "s"},
	}
	g.genGPUSoftmaxStmt(sstmt)

	lstmt := &ast.GPULayernormStmt{
		Source: &ast.Ident{Name: "a"},
		Target: &ast.Ident{Name: "l"},
	}
	g.genGPULayernormStmt(lstmt)

	attest := &ast.AGKEAttestStmt{
		Chain: &ast.Ident{Name: "ch"},
	}
	g.genAGKEAttestStmt(attest)

	vstmt := &ast.AGKEVerifyStmt{
		Chain:  &ast.Ident{Name: "ch"},
		Target: &ast.Ident{Name: "ok"},
	}
	g.genAGKEVerifyStmt(vstmt)

	corrupt := &ast.CorruptGPUDigestStmt{
		Offset: &ast.IntLit{Value: 0},
	}
	g.genCorruptGPUDigestStmt(corrupt)

	mExpr := g.genGPUMatmulExpr(&ast.GPUMatmulExpr{
		Left:  &ast.Ident{Name: "a"},
		Right: &ast.Ident{Name: "b"},
	})
	if mExpr.Type != "i8*" {
		t.Fatalf("expected i8*, got %s", mExpr.Type)
	}

	smExpr := g.genGPUSoftmaxExpr(&ast.GPUSoftmaxExpr{
		Source: &ast.Ident{Name: "a"},
	})
	if smExpr.Type != "i8*" {
		t.Fatalf("expected i8*, got %s", smExpr.Type)
	}

	lnExpr := g.genGPULayernormExpr(&ast.GPULayernormExpr{
		Source: &ast.Ident{Name: "a"},
	})
	if lnExpr.Type != "i8*" {
		t.Fatalf("expected i8*, got %s", lnExpr.Type)
	}

	vExpr := g.genAGKEVerifyExpr(&ast.AGKEVerifyExpr{
		Chain: &ast.Ident{Name: "ch"},
	})
	if vExpr.Type != "i1" {
		t.Fatalf("expected i1, got %s", vExpr.Type)
	}

	ir := g.sb.String()
	if !strings.Contains(ir, "@qadav_gpu_tensor_matmul") {
		t.Errorf("expected @qadav_gpu_tensor_matmul in IR")
	}
	if !strings.Contains(ir, "@qadav_gpu_tensor_softmax") {
		t.Errorf("expected @qadav_gpu_tensor_softmax in IR")
	}
	if !strings.Contains(ir, "@qadav_gpu_tensor_layernorm") {
		t.Errorf("expected @qadav_gpu_tensor_layernorm in IR")
	}
	if !strings.Contains(ir, "@qadav_agke_attest") {
		t.Errorf("expected @qadav_agke_attest in IR")
	}
	if !strings.Contains(ir, "@qadav_agke_verify") {
		t.Errorf("expected @qadav_agke_verify in IR")
	}
	if !strings.Contains(ir, "@qadav_agke_corrupt_record") {
		t.Errorf("expected @qadav_agke_corrupt_record in IR")
	}
}
