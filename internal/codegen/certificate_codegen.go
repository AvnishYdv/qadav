package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genCertifiedReasonStmt(s *ast.CertifiedReasonStmt) {
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
	switch s.Kind {
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

	// 1. Call QERA reason step
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qera_reason(i8* %s, i32 %d, i8* %s, i64 %d, i8* %s, i64 %d, i8* null, i64 0, double %s, i8* null)\n",
		resTemp, chainRef, opCode, inRef, inLen, outRef, outLen, confRef)

	// 2. Locate step index: chain->n_steps - 1 (offset 64 in qadav_reasoning_chain)
	stepsPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds i8, i8* %s, i64 64\n", stepsPtr, chainRef)
	nstepsPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast i8* %s to i32*\n", nstepsPtr, stepsPtr)
	curNsteps := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i32, i32* %s\n", curNsteps, nstepsPtr)
	stepIdx32 := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = sub i32 %s, 1\n", stepIdx32, curNsteps)
	stepIdx64 := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = zext i32 %s to i64\n", stepIdx64, stepIdx32)

	// 3. Prepare witness
	var witnessRef = "null"
	witnessLen := 0
	if s.Witness != nil {
		wVal := g.genExpr(s.Witness)
		if wVal.Type == "i8*" {
			witnessRef = wVal.Ref
		}
		if lit, ok := s.Witness.(*ast.StringLit); ok {
			witnessLen = len(lit.Value)
		}
	}

	// 4. Attach certificate
	certTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_certificate_attach(i8* %s, i64 %s, i8 %d, i8* %s, i32 %d, double %s)\n",
		certTemp, chainRef, stepIdx64, opCode, witnessRef, witnessLen, confRef)
}

func (g *Generator) genAbstainThresholdDecl(d *ast.AbstainThresholdDecl) {
	chainName := ""
	if id, ok := d.Chain.(*ast.Ident); ok {
		chainName = id.Name
	} else if g.activeReasoningChain != "" {
		chainName = g.activeReasoningChain
	}

	modeVal := 3 // QADAV_ABSTAIN_REFUSE default
	switch d.Mode {
	case token.RETRY:
		modeVal = 1
	case token.ESCALATE:
		modeVal = 2
	case token.REFUSE:
		modeVal = 3
	}

	threshRef := "0.8"
	if d.Threshold != nil {
		tVal := g.genExpr(d.Threshold)
		threshRef = tVal.Ref
	}

	retriesVal := 0
	if d.MaxRetries != nil {
		if lit, ok := d.MaxRetries.(*ast.IntLit); ok {
			retriesVal = int(lit.Value)
		}
	}

	// Allocate policy struct on stack (24 bytes: double threshold (8) + i32 mode (4) + i32 max_retries (4))
	rawPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [24 x i8]\n", rawPtr)
	policyPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast [24 x i8]* %s to i8*\n", policyPtr, rawPtr)

	tPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast i8* %s to double*\n", tPtr, policyPtr)
	fmt.Fprintf(&g.sb, "  store double %s, double* %s\n", threshRef, tPtr)

	modeOffset := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds i8, i8* %s, i64 8\n", modeOffset, policyPtr)
	modePtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast i8* %s to i32*\n", modePtr, modeOffset)
	fmt.Fprintf(&g.sb, "  store i32 %d, i32* %s\n", modeVal, modePtr)

	retriesOffset := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds i8, i8* %s, i64 12\n", retriesOffset, policyPtr)
	retriesPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast i8* %s to i32*\n", retriesPtr, retriesOffset)
	fmt.Fprintf(&g.sb, "  store i32 %d, i32* %s\n", retriesVal, retriesPtr)

	if g.abstainPolicies == nil {
		g.abstainPolicies = make(map[string]string)
	}
	if chainName != "" {
		g.abstainPolicies[chainName] = policyPtr
	}
}

func (g *Generator) genAbstainCheckExpr(e *ast.AbstainCheckExpr) Value {
	var chainRef = "null"
	chainName := ""
	if id, ok := e.Chain.(*ast.Ident); ok {
		chainName = id.Name
	}

	if e.Chain != nil {
		chainVal := g.genExpr(e.Chain)
		chainRef = chainVal.Ref
	} else if g.activeReasoningChain != "" {
		chainName = g.activeReasoningChain
		if slot, ok := g.vars[g.activeReasoningChain]; ok {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			chainRef = t
		}
	}

	policyRef := "null"
	if chainName != "" && g.abstainPolicies != nil {
		if p, ok := g.abstainPolicies[chainName]; ok {
			policyRef = p
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_abstain_check(i8* %s, i8* %s, i32 0)\n",
		resTemp, chainRef, policyRef)

	extTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = zext i32 %s to i64\n", extTemp, resTemp)

	return Value{
		Type: "i64",
		Ref:  extTemp,
	}
}

func (g *Generator) genAbstainCheckStmt(s *ast.AbstainCheckStmt) {
	val := g.genAbstainCheckExpr(&ast.AbstainCheckExpr{
		Pos_:  s.Pos(),
		Chain: s.Chain,
	})

	if s.Into != "" {
		ptr := fmt.Sprintf("%%%s.ptr", s.Into)
		fmt.Fprintf(&g.sb, "  %s = alloca i64\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i64",
			qtype:    types.Int64,
		}
		g.vars[s.Into] = slot
		fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", val.Ref, ptr)
	}
}
