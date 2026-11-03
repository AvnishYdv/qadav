package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genTranspileToStmt(s *ast.TranspileToStmt) {
	circTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qhal_create(i32 2, i32 0)\n", circTemp)

	hRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qhal_add_op(i8* %s, i8 0, i32 0, i32 0)\n", hRes, circTemp)

	cnotRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qhal_add_op(i8* %s, i8 4, i32 0, i32 1)\n", cnotRes, circTemp)

	targetCode := 0
	switch s.Target {
	case "qiskit":
		targetCode = 0
	case "cirq":
		targetCode = 1
	case "ionq":
		targetCode = 2
	case "qasm3":
		targetCode = 3
	}

	bufLen := int64(2048)
	if s.BufLen != nil {
		if lit, ok := s.BufLen.(*ast.IntLit); ok {
			bufLen = lit.Value
		}
	}
	bufTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 %d)\n", bufTemp, bufLen)

	writtenAlloca := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i64\n", writtenAlloca)

	opCountAlloca := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i32\n", opCountAlloca)

	optsPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 16)\n", optsPtr)
	if s.Options != nil && s.Options.Optimize {
		optsI32 := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = bitcast i8* %s to i32*\n", optsI32, optsPtr)
		fmt.Fprintf(&g.sb, "  store i32 1, i32* %s\n", optsI32)
	}

	trRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @transpile_circuit(i8* %s, i32 %d, i8* %s, i8* %s, i64 %d, i64* %s, i32* %s)\n",
		trRes, circTemp, targetCode, optsPtr, bufTemp, bufLen, writtenAlloca, opCountAlloca)

	fmt.Fprintf(&g.sb, "  call void @qhal_shutdown(i8* %s)\n", circTemp)
	fmt.Fprintf(&g.sb, "  call void @qadav_free(i8* %s)\n", optsPtr)

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
