package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genReasoningChainDecl(d *ast.ReasoningChainDecl) {
	capacity := 1024
	if d.Capacity != nil {
		if lit, ok := d.Capacity.(*ast.IntLit); ok {
			capacity = int(lit.Value)
		}
	}

	nameConst, nameLen := g.addStringConstant(d.Name)
	namePtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i8], [%d x i8]* %s, i32 0, i32 0\n", namePtr, nameLen, nameLen, nameConst)

	var bRef = "null"
	if d.Boundary != nil {
		bVal := g.genExpr(d.Boundary)
		bRef = bVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qera_create(i8* %s, i32 %d, i8* %s)\n", resTemp, namePtr, capacity, bRef)

	if d.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", d.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.ReasoningChainT(),
		}
		g.vars[d.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
		g.activeReasoningChain = d.Name
	}
}

func (g *Generator) genReasonStepStmt(s *ast.ReasonStepStmt) {
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

	var inRef = "null"
	inLen := 0
	if s.InputExpr != nil {
		inVal := g.genExpr(s.InputExpr)
		if inVal.Type == "i8*" {
			inRef = inVal.Ref
		}
		if lit, ok := s.InputExpr.(*ast.StringLit); ok {
			inLen = len(lit.Value)
		}
	}

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

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qera_reason(i8* %s, i32 %d, i8* %s, i64 %d, i8* %s, i64 %d, i8* null, i64 0, double %s, i8* null)\n",
		resTemp, chainRef, opCode, inRef, inLen, outRef, outLen, confRef)
}

func (g *Generator) genVerifyChainStmt(s *ast.VerifyChainStmt) {
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

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_certificate_verify_chain(i8* %s)\n", resTemp, chainRef)
}

func (g *Generator) genAuditReasoningStmt(s *ast.AuditReasoningStmt) {
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

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qera_audit(i8* %s)\n", resTemp, chainRef)
}

func (g *Generator) genCorruptReasoningStepStmt(s *ast.CorruptReasoningStepStmt) {
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

	stepVal := g.genExpr(s.StepIndex)
	stepRef := stepVal.Ref
	if stepVal.Type == "i64" {
		t := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", t, stepVal.Ref)
		stepRef = t
	}

	offsetRef := "0"
	if s.AtVal != nil {
		offVal := g.genExpr(s.AtVal)
		offsetRef = offVal.Ref
		if offVal.Type == "i64" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", t, offVal.Ref)
			offsetRef = t
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qera_corrupt_step(i8* %s, i32 %s, i32 %s)\n", resTemp, chainRef, stepRef, offsetRef)
}
