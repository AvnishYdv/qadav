package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genQASMExportStmt(s *ast.QASMExportStmt) {
	circTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qhal_create(i32 2, i32 0)\n", circTemp)

	hRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qhal_add_op(i8* %s, i8 0, i32 0, i32 0)\n", hRes, circTemp)

	cnotRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qhal_add_op(i8* %s, i8 4, i32 0, i32 1)\n", cnotRes, circTemp)

	bufLen := int64(1024)
	if s.BufLen != nil {
		if lit, ok := s.BufLen.(*ast.IntLit); ok {
			bufLen = lit.Value
		}
	}
	bufTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 %d)\n", bufTemp, bufLen)

	writtenAlloca := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i64\n", writtenAlloca)

	expRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qhal_export_qasm3(i8* %s, i8* %s, i64 %d, i64* %s)\n", expRes, circTemp, bufTemp, bufLen, writtenAlloca)

	fmt.Fprintf(&g.sb, "  call void @qhal_shutdown(i8* %s)\n", circTemp)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.Into] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", bufTemp, slot.ptr)
	}
}
