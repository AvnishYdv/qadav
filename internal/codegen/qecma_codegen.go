package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
)

type modalityInfo struct {
	tag     byte
	dataRef string
	lenVal  int
}

func (g *Generator) genModalityDecl(d *ast.ModalityDecl) {
	tag := byte(0x58) // 'X' text default
	switch d.Kind {
	case token.TENSOR_INPUT:
		tag = 0x54 // 'T'
	case token.TEXT_INPUT:
		tag = 0x58 // 'X'
	case token.CONCEPT_INPUT:
		tag = 0x43 // 'C'
	}

	dataRef := "null"
	lenVal := 0

	if d.Source != nil {
		srcVal := g.genExpr(d.Source)
		switch srcVal.Type {
		case "i8*":
			dataRef = srcVal.Ref
			if lit, ok := d.Source.(*ast.StringLit); ok {
				lenVal = len(lit.Value)
			}
		case "i64", "i32":
			// Integer or concept scalar - store in temporary stack alloca
			tmpPtr := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i64\n", tmpPtr)
			if srcVal.Type == "i32" {
				ext := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", ext, srcVal.Ref)
				fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", ext, tmpPtr)
			} else {
				fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", srcVal.Ref, tmpPtr)
			}
			castPtr := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = bitcast i64* %s to i8*\n", castPtr, tmpPtr)
			dataRef = castPtr
			lenVal = 8
		default:
			// Pointer or struct
			castPtr := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = bitcast %s %s to i8*\n", castPtr, srcVal.Type, srcVal.Ref)
			dataRef = castPtr
			lenVal = 32 // nominal struct length
		}
	}

	if g.modalities == nil {
		g.modalities = make(map[string]modalityInfo)
	}
	if d.Name != "" {
		g.modalities[d.Name] = modalityInfo{
			tag:     tag,
			dataRef: dataRef,
			lenVal:  lenVal,
		}
	}
}

func (g *Generator) genCrossModalReasonStmt(s *ast.CrossModalReasonStmt) {
	// 1. Create bundle
	bundleTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qecma_create()\n", bundleTemp)

	// 2. Add modalities
	for _, m := range s.Modalities {
		mName := ""
		if ident, ok := m.(*ast.Ident); ok {
			mName = ident.Name
		}
		if info, ok := g.modalities[mName]; ok {
			addTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qecma_add(i8* %s, i8 %d, i8* %s, i64 %d)\n",
				addTemp, bundleTemp, info.tag, info.dataRef, info.lenVal)
		} else {
			// Evaluate expression dynamically
			mVal := g.genExpr(m)
			addTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qecma_add(i8* %s, i8 88, i8* %s, i64 0)\n",
				addTemp, bundleTemp, mVal.Ref)
		}
	}

	// 3. Resolve reasoning chain
	var chainRef = "null"
	if s.Chain != nil {
		chainVal := g.genExpr(s.Chain)
		chainRef = chainVal.Ref
	} else if g.activeReasoningChain != "" {
		if slot, ok := g.vars[g.activeReasoningChain]; ok {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			chainRef = t
		}
	}

	// 4. Resolve operation
	opCode := 1
	switch s.Operation {
	case token.DEDUCE:
		opCode = 1
	case token.INDUCE:
		opCode = 2
	case token.ABDUCE:
		opCode = 3
	case token.VERIFY:
		opCode = 4
	}

	// 5. Resolve output
	var outRef = "null"
	outLen := 0
	if s.OutputExpr != nil {
		outVal := g.genExpr(s.OutputExpr)
		if outVal.Type == "i8*" {
			outRef = outVal.Ref
		}
		if lit, ok := s.OutputExpr.(*ast.StringLit); ok {
			outLen = len(lit.Value)
		}
	}

	// 6. Resolve confidence
	confRef := "1.0"
	if s.Confidence != nil {
		confVal := g.genExpr(s.Confidence)
		if confVal.Type == "i64" || confVal.Type == "i32" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = sitofp %s %s to double\n", t, confVal.Type, confVal.Ref)
			confRef = t
		} else {
			confRef = confVal.Ref
		}
	}

	// 7. Call qadav_qera_reason_cross_modal
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qera_reason_cross_modal(i8* %s, i32 %d, i8* %s, i8* %s, i64 %d, double %s, i8* null)\n",
		resTemp, chainRef, opCode, bundleTemp, outRef, outLen, confRef)

	// 8. Cleanup bundle
	fmt.Fprintf(&g.sb, "  call void @qadav_qecma_free(i8* %s)\n", bundleTemp)
}

