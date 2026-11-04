package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genCDAGARecordStmt(s *ast.CDAGARecordStmt) {
	recTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 512)\n", recTemp)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_cdaga_record_aggregation(i8* null, i64 1, i8** null, i64* null, i32 %d, i8* %s)\n", resTemp, len(s.GradList), recTemp)

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
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", recTemp, slot.ptr)
	}
}

func (g *Generator) genCDAGAAttestStmt(s *ast.CDAGAAttestStmt) {
	var recRef = "null"
	if slot, ok := g.vars[s.Record]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		recRef = loaded
	}
	var chainRef = "null"
	if slot, ok := g.vars[s.Chain]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		chainRef = loaded
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_cdaga_attest(i8* %s, i8* %s, i8* null)\n", resTemp, chainRef, recRef)
}

func (g *Generator) genCDAGAVerifyStmt(s *ast.CDAGAVerifyStmt) {
	var recRef = "null"
	if slot, ok := g.vars[s.Record]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		recRef = loaded
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_cdaga_verify(i8* %s, i8* null)\n", resTemp, recRef)

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
		boolTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp eq i32 %s, 1\n", boolTemp, resTemp)
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", boolTemp, slot.ptr)
	}
}

func (g *Generator) genCorruptCDAGAStmt(s *ast.CorruptCDAGAStmt) {
	var recRef = "null"
	if slot, ok := g.vars[s.Record]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		recRef = loaded
	}
	offset := 100
	if s.Offset != nil {
		if lit, ok := s.Offset.(*ast.IntLit); ok {
			offset = int(lit.Value)
		}
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_cdaga_corrupt(i8* %s, i32 %d)\n", resTemp, recRef, offset)
}
