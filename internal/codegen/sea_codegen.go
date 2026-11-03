package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genEvolveStmt(s *ast.EvolveStmt) {
	histRef := "null"
	if slot, ok := g.vars["__sea_history"]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		histRef = loaded
	} else {
		histTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @sea_create()\n", histTemp)
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		slot := &varSlot{
			ptr:      alloca,
			llvmType: "i8*",
		}
		g.vars["__sea_history"] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", histTemp, slot.ptr)
		histRef = histTemp
	}

	fromHash := g.genStateHash(s.FromState, "state_a")
	toHash := g.genStateHash(s.ToState, "state_b")
	reasonHash := g.genStateHash(s.Reason, "reason_init")

	propAlloca := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", propAlloca)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @sea_propose(i8* %s, i8* %s, i8* %s, i8* %s, i8** %s)\n",
		resTemp, histRef, fromHash, toHash, reasonHash, propAlloca)

	propPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", propPtr, propAlloca)

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
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", propPtr, slot.ptr)
	}
}

func (g *Generator) genStateHash(e ast.Expr, fallbackStr string) string {
	bufTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 32)\n", bufTemp)

	var textPtr string
	if ident, ok := e.(*ast.Ident); ok {
		if slot, ok := g.vars[ident.Name]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			textPtr = loaded
		}
	}
	if textPtr == "" {
		textPtr = g.genStringConst(fallbackStr)
	}

	fmt.Fprintf(&g.sb, "  call void @sea_hash_text(i8* %s, i8* %s)\n", textPtr, bufTemp)
	return bufTemp
}

func (g *Generator) genEvolveVerifyStmt(s *ast.EvolveVerifyStmt) {
	histRef := "null"
	if slot, ok := g.vars["__sea_history"]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		histRef = loaded
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @sea_verify_history(i8* %s)\n", resTemp, histRef)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i1\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i1",
			}
			g.vars[s.Into] = slot
		}
		cmp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", cmp, resTemp)
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", cmp, slot.ptr)
	}
}

func (g *Generator) genEvolveHistoryStmt(s *ast.EvolveHistoryStmt) {
	histRef := "null"
	if slot, ok := g.vars["__sea_history"]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		histRef = loaded
	}

	chainRef := "null"
	if slot, ok := g.vars[s.Chain]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		chainRef = loaded
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @sea_attest(i8* %s, i8* %s)\n", resTemp, histRef, chainRef)
}

func (g *Generator) genCorruptEvolveStmt(s *ast.CorruptEvolveStmt) {
	propRef := "null"
	if slot, ok := g.vars[s.Proposal]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		propRef = loaded
	}

	offsetVal := 16
	if s.Offset != nil {
		if lit, ok := s.Offset.(*ast.IntLit); ok {
			offsetVal = int(lit.Value)
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @sea_corrupt(i8* %s, i32 %d)\n", resTemp, propRef, offsetVal)
}
