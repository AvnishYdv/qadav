package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genQECMemoryDecl(d *ast.QECMemoryDecl) {
	nameConst, nameLen := g.addStringConstant(d.Name)
	namePtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i32 0, i32 0\n", namePtr, nameLen, nameLen, nameConst)

	codeType := 1 // QADAV_QEC_SURFACE default
	if d.CodeType == token.QLDPC_CODE {
		codeType = 2
	}

	distVal := 3
	if d.Distance != nil {
		if lit, ok := d.Distance.(*ast.IntLit); ok {
			distVal = int(lit.Value)
		}
	}

	bRef := "null"
	if d.Boundary != nil {
		bVal := g.genExpr(d.Boundary)
		bRef = bVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qec_create(i8* %s, i32 %d, i32 %d, i8* %s)\n", resTemp, namePtr, codeType, distVal, bRef)

	if d.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", d.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.QECMemoryT(),
		}
		g.vars[d.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genSyndromeStmt(s *ast.SyndromeStmt) {
	memVal := g.genExpr(s.QECMem)

	// Non-zero test raw syndrome payload (8 bytes: 0x5a, 0x3c, ...)
	rawConst, rawLen := g.addStringConstant("\x5a\x3c\x12\x7f\x89\xab\xcd\xef")
	rawPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i32 0, i32 0\n", rawPtr, rawLen, rawLen, rawConst)

	idAlloca := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i64\n", idAlloca)

	bRef := "null"
	for _, slot := range g.vars {
		if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			bRef = t
			break
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qec_measure(i8* %s, i8* %s, i64 8, i64* %s, i8* %s)\n", resTemp, memVal.Ref, rawPtr, idAlloca, bRef)

	if s.Into != "" {
		loadedId := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i64, i64* %s\n", loadedId, idAlloca)
		ptr := fmt.Sprintf("%%%s.ptr", s.Into)
		fmt.Fprintf(&g.sb, "  %s = alloca i64\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i64",
			qtype:    types.SyndromeT(),
		}
		g.vars[s.Into] = slot
		fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", loadedId, ptr)
	}
}

func (g *Generator) genVerifySyndromeStmt(s *ast.VerifySyndromeStmt) {
	memVal := g.genExpr(s.QECMem)
	synVal := g.genExpr(s.Syndrome)

	bRef := "null"
	for _, slot := range g.vars {
		if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			bRef = t
			break
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qec_verify(i8* %s, i64 %s, i8* %s)\n", resTemp, memVal.Ref, synVal.Ref, bRef)
}

func (g *Generator) genCorrectStmt(s *ast.CorrectStmt) {
	memVal := g.genExpr(s.QECMem)
	synVal := g.genExpr(s.Syndrome)

	bRef := "null"
	for _, slot := range g.vars {
		if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			bRef = t
			break
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qec_correct(i8* %s, i64 %s, i8* %s)\n", resTemp, memVal.Ref, synVal.Ref, bRef)
}

func (g *Generator) genCorruptSyndromeStmt(s *ast.CorruptSyndromeStmt) {
	memVal := g.genExpr(s.QECMem)
	synVal := g.genExpr(s.Syndrome)
	offset := 0
	if s.AtVal != nil {
		if lit, ok := s.AtVal.(*ast.IntLit); ok {
			offset = int(lit.Value)
		}
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qec_corrupt_syndrome(i8* %s, i64 %s, i32 %d)\n", resTemp, memVal.Ref, synVal.Ref, offset)
}
