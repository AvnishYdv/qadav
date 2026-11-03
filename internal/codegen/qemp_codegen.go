package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genMemoryProvenanceDecl(m *ast.MemoryProvenanceDecl) {
	capacity := 1024
	if m.Capacity != nil {
		if lit, ok := m.Capacity.(*ast.IntLit); ok {
			capacity = int(lit.Value)
		}
	}

	var bRef = "null"
	if m.Boundary != nil {
		bVal := g.genExpr(m.Boundary)
		bRef = bVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qemp_create(i8* %s, i32 %d)\n", resTemp, bRef, capacity)

	if m.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", m.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.MemoryProv,
		}
		g.vars[m.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genVerifyMemoryStmt(s *ast.VerifyMemoryStmt) {
	mVal := g.genExpr(s.Mem)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qemp_verify_all(i8* %s)\n", resTemp, mVal.Ref)
}

func (g *Generator) genAuditMemoryStmt(s *ast.AuditMemoryStmt) {
	mVal := g.genExpr(s.Mem)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qemp_audit(i8* %s)\n", resTemp, mVal.Ref)
}

func (g *Generator) genCorruptShareStmt(s *ast.CorruptShareStmt) {
	mVal := g.genExpr(s.Mem)
	kVal := g.genExpr(s.Key)
	idx := 0
	if s.ShareIndex != nil {
		if lit, ok := s.ShareIndex.(*ast.IntLit); ok {
			idx = int(lit.Value)
		}
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qemp_corrupt_share(i8* %s, i8* %s, i32 %d)\n", resTemp, mVal.Ref, kVal.Ref, idx)
}
