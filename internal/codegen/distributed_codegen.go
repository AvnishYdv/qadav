package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genAllReduceStmt(s *ast.AllReduceStmt) {
	distTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_dist_init(i32 4)\n", distTemp)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_dist_ring_allreduce(i8* %s)\n", resTemp, distTemp)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i64\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i64",
			}
			g.vars[s.Into] = slot
		}
		ext := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", ext, resTemp)
		fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", ext, slot.ptr)
	}

	fmt.Fprintf(&g.sb, "  call void @qadav_dist_shutdown(i8* %s)\n", distTemp)
}

func (g *Generator) genPserverInitStmt(s *ast.PserverInitStmt) {
	workerCount := 4
	if s.Workers != nil {
		if lit, ok := s.Workers.(*ast.IntLit); ok {
			workerCount = int(lit.Value)
		}
	}

	distTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_dist_init(i32 %d)\n", distTemp, workerCount)

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
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", distTemp, slot.ptr)
	}
}

func (g *Generator) genPserverAggregateStmt(s *ast.PserverAggregateStmt) {
	psRef := "null"
	if slot, ok := g.vars[s.Pserver]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		psRef = loaded
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_dist_pserver_aggregate(i8* %s)\n", resTemp, psRef)
}
