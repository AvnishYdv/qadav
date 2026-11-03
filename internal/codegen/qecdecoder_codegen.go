package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genQECDDecodeStmt(s *ast.QECDDecodeStmt) {
	dist := 3
	if s.Distance != nil {
		if iLit, ok := s.Distance.(*ast.IntLit); ok {
			dist = int(iLit.Value)
		}
	}
	nQubits := dist * dist

	synRef := "null"
	synLen := 0
	if s.SyndromeIn != "" {
		if slot, ok := g.vars[s.SyndromeIn]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			synRef = loaded
			synLen = 8
		} else {
			// Allocate a small dummy syndrome buffer (with bit 0 set for demonstration)
			dummyBuf := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 8)\n", dummyBuf)
			fmt.Fprintf(&g.sb, "  store i8 1, i8* %s\n", dummyBuf)
			synRef = dummyBuf
			synLen = 1
		}
	} else {
		dummyBuf := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 8)\n", dummyBuf)
		fmt.Fprintf(&g.sb, "  store i8 1, i8* %s\n", dummyBuf)
		synRef = dummyBuf
		synLen = 1
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qecd_decode(i32 %d, i32 %d, i8* %s, i32 %d)\n",
		resTemp, dist, nQubits, synRef, synLen)

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

func (g *Generator) genQECDVerifyStmt(s *ast.QECDVerifyStmt) {
	resRef := "null"
	if s.Result != "" {
		if slot, ok := g.vars[s.Result]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			resRef = loaded
		}
	}

	vTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qecd_verify_decode_attestation(i8* %s)\n", vTemp, resRef)

	boolTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", boolTemp, vTemp)

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
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", boolTemp, slot.ptr)
	}
}

func (g *Generator) genQECDApplyStmt(s *ast.QECDApplyStmt) {
	resRef := "null"
	if s.Result != "" {
		if slot, ok := g.vars[s.Result]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			resRef = loaded
		}
	}

	stateRef := "null"
	if s.StateVec != "" {
		if slot, ok := g.vars[s.StateVec]; ok {
			loaded := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
			stateRef = loaded
		}
	}

	callRes := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qecd_apply_corrections(i8* %s, i8* %s, i32 64)\n",
		callRes, resRef, stateRef)
}
