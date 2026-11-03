package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genVaultDecl(v *ast.VaultDecl) {
	threshold := 3
	if v.Threshold != nil {
		if lit, ok := v.Threshold.(*ast.IntLit); ok {
			threshold = int(lit.Value)
		}
	}
	nShares := 5

	var bRef = "null"
	if v.Boundary != nil {
		bVal := g.genExpr(v.Boundary)
		bRef = bVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qemwv_create(i32 %d, i32 %d, i8* %s)\n", resTemp, nShares, threshold, bRef)

	if v.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", v.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.Vault,
		}
		g.vars[v.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genSealStmt(s *ast.SealStmt) {
	vVal := g.genExpr(s.Vault)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qemwv_seal(i8* %s, i8** null, i32 0)\n", resTemp, vVal.Ref)
}

func (g *Generator) genUnsealStmt(s *ast.UnsealStmt) {
	vVal := g.genExpr(s.Vault)
	var bRef = "null"
	if s.Boundary != nil {
		bVal := g.genExpr(s.Boundary)
		bRef = bVal.Ref
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8** @qadav_qemwv_unseal(i8* %s, i8* %s)\n", resTemp, vVal.Ref, bRef)
}

func (g *Generator) genWipeStmt(s *ast.WipeStmt) {
	vVal := g.genExpr(s.Vault)
	fmt.Fprintf(&g.sb, "  call void @qadav_qemwv_wipe(i8* %s)\n", vVal.Ref)
}
