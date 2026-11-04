package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genCTSAAttestStmt(s *ast.CTSAAttestStmt) {
	if s == nil {
		return
	}

	dummy := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 32)\n", dummy)

	idxTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_ctsa_attest(i8* %s, i8* %s, i8* %s, i64 10, i64 5)\n",
		idxTemp, dummy, dummy, dummy)

	if s.Into != "" {
		strPtr := g.genStringConst("ctsa_cert_verified")
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

func (g *Generator) genCTSAVerifyStmt(s *ast.CTSAVerifyStmt) {
	if s == nil {
		return
	}

	vTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_ctsa_verify()\n", vTemp)

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

func (g *Generator) genCorruptCTSAStmt(s *ast.CorruptCTSAStmt) {
	if s == nil {
		return
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_ctsa_corrupt(i32 1)\n", resTemp)
}
