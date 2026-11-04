package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genConsolidateStmt(s *ast.ConsolidateStmt) {
	mpVal := g.genExpr(s.EpisodicMem)
	graphVal := g.genExpr(s.SemanticMem)

	threshold := 0.5
	if s.MinConfidence != nil {
		if lit, ok := s.MinConfidence.(*ast.FloatLit); ok {
			threshold = lit.Value
		}
	}

	bRef := "null"
	// Find if there is an active ethical boundary
	for _, slot := range g.vars {
		if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			bRef = t
			break
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_consolidate_memory(i8* %s, i8* %s, double %f, i8* %s)\n", resTemp, mpVal.Ref, graphVal.Ref, threshold, bRef)
}
