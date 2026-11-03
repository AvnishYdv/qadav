package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genPipelineClaimStmt(s *ast.PipelineClaimStmt) {
	ctxTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_multigpu_init()\n", ctxTemp)
	retTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_multigpu_build_pipeline(i8* %s, i32 8, i32 %d, i32 %d)\n", retTemp, ctxTemp, s.Stages, s.MicroBatch)

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
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", ctxTemp, slot.ptr)
	}
}

func (g *Generator) genPipelineRunStmt(s *ast.PipelineRunStmt) {
	var ctxRef = "null"
	if slot, ok := g.vars[s.Plan]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		ctxRef = loaded
	}

	var inRef = "null"
	if s.Input != nil {
		inVal := g.genExpr(s.Input)
		inRef = inVal.Ref
	}

	outTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 256)\n", outTemp)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_multigpu_run_pipeline(i8* %s, i8* %s, i8* %s)\n", resTemp, ctxRef, inRef, outTemp)

	if s.Output != nil {
		if ident, ok := s.Output.(*ast.Ident); ok {
			slot, ok := g.vars[ident.Name]
			if !ok {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
				slot = &varSlot{
					ptr:      alloca,
					llvmType: "i8*",
				}
				g.vars[ident.Name] = slot
			}
			fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", outTemp, slot.ptr)
		}
	}
}
