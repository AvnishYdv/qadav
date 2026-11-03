package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genRSCAAttestStmt(s *ast.RSCAAttestStmt) {
	if s == nil {
		return
	}

	srcPtr := "null"
	cmpPtr := "null"
	outPtr := "null"

	dummy := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 32)\n", dummy)
	srcPtr = dummy
	cmpPtr = dummy
	outPtr = dummy

	idxTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_rsca_record_stage(i64 %d, i8* %s, i8* %s, i8* %s)\n",
		idxTemp, s.Stage, srcPtr, cmpPtr, outPtr)

	if s.Into != "" {
		strPtr := g.genStringConst("rsca_record_verified")
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		g.vars[s.Into] = &varSlot{
			ptr:      alloca,
			llvmType: "i8*",
			qtype:    types.Text,
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", strPtr, alloca)
	}
}

func (g *Generator) genRSCAVerifyStmt(s *ast.RSCAVerifyStmt) {
	if s == nil {
		return
	}

	vTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_rsca_attest_chain()\n", vTemp)

	if s.Into != "" {
		boolTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", boolTemp, vTemp)
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i1\n", alloca)
		g.vars[s.Into] = &varSlot{
			ptr:      alloca,
			llvmType: "i1",
			qtype:    types.Logic,
		}
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", boolTemp, alloca)
	}
}

func (g *Generator) genCorruptRSCAStmt(s *ast.CorruptRSCAStmt) {
	if s == nil {
		return
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_rsca_corrupt_record(i32 0, i32 0)\n", resTemp)
}
