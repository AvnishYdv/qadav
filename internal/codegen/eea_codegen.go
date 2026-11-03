package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genEvolvePolicyStmt(s *ast.EvolvePolicyStmt) {
	kindCode := 0
	switch s.Kind {
	case "constant":
		kindCode = 0
	case "decaying":
		kindCode = 1
	case "adaptive":
		kindCode = 2
	}

	rateStr := "0.100000"
	if s.InitialRate != nil {
		if fLit, ok := s.InitialRate.(*ast.FloatLit); ok {
			rateStr = fmt.Sprintf("%f", fLit.Value)
		}
	}

	budget := int64(100)
	if s.TotalBudget != nil {
		if iLit, ok := s.TotalBudget.(*ast.IntLit); ok {
			budget = iLit.Value
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @eea_create_policy(i32 %d, double %s, i64 %d)\n",
		resTemp, kindCode, rateStr, budget)

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
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, slot.ptr)
	}
}

func (g *Generator) genEvolveBoundedStmt(s *ast.EvolveBoundedStmt) {
	policyRef := "null"
	if s.Policy != "" {
		if slot, ok := g.vars[s.Policy]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			policyRef = loaded
		}
	}

	// SEA History & Proposal
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

	proposeRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @sea_propose(i8* %s, i8* %s, i8* %s, i8* %s, i8** %s)\n",
		proposeRes, histRef, fromHash, toHash, reasonHash, propAlloca)

	propPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", propPtr, propAlloca)

	if s.IntoProposal != "" {
		slot, ok := g.vars[s.IntoProposal]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.IntoProposal] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", propPtr, slot.ptr)
	}

	// Resolve from and to text pointers
	fromTextPtr := g.resolveTextPtr(s.FromState, "state_a")
	toTextPtr := g.resolveTextPtr(s.ToState, "state_b")

	maxFlips := int64(8)
	if s.MaxBitFlips != nil {
		if lit, ok := s.MaxBitFlips.(*ast.IntLit); ok {
			maxFlips = lit.Value
		}
	}

	maxNewSyms := int64(2)
	if s.MaxNewSymbols != nil {
		if lit, ok := s.MaxNewSymbols.(*ast.IntLit); ok {
			maxNewSyms = lit.Value
		}
	}

	envPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @eea_create_envelope()\n", envPtr)

	buildRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @eea_build_envelope_text(i8* %s, i8* %s, i8* %s, i8* %s, i64 %d, i64 %d, i8* %s)\n",
		buildRes, policyRef, propPtr, fromTextPtr, toTextPtr, maxFlips, maxNewSyms, envPtr)

	if s.IntoEnvelope != "" {
		slot, ok := g.vars[s.IntoEnvelope]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.IntoEnvelope] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", envPtr, slot.ptr)
	}
	// Also track __eea_envelope
	if slot, ok := g.vars["__eea_envelope"]; ok {
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", envPtr, slot.ptr)
	} else {
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		slot := &varSlot{ptr: alloca, llvmType: "i8*"}
		g.vars["__eea_envelope"] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", envPtr, slot.ptr)
	}
}

func (g *Generator) resolveTextPtr(e ast.Expr, fallback string) string {
	if ident, ok := e.(*ast.Ident); ok {
		if slot, ok := g.vars[ident.Name]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			return loaded
		}
	}
	return g.genStringConst(fallback)
}

func (g *Generator) genEEAVerifyStmt(s *ast.EEAVerifyStmt) {
	envRef := "null"
	if s.Envelope != "" {
		if slot, ok := g.vars[s.Envelope]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			envRef = loaded
		}
	} else if slot, ok := g.vars["__eea_envelope"]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		envRef = loaded
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @eea_verify_envelope(i8* %s)\n", resTemp, envRef)

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
		cmpTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", cmpTemp, resTemp)
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", cmpTemp, slot.ptr)
	}
}

func (g *Generator) genCorruptEEAStmt(s *ast.CorruptEEAStmt) {
	envRef := "null"
	if s.Envelope != "" {
		if slot, ok := g.vars[s.Envelope]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			envRef = loaded
		}
	}

	offsetVal := "16"
	if s.Offset != nil {
		if lit, ok := s.Offset.(*ast.IntLit); ok {
			offsetVal = fmt.Sprintf("%d", lit.Value)
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @eea_corrupt(i8* %s, i32 %s)\n", resTemp, envRef, offsetVal)
}
