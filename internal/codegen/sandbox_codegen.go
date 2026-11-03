package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genSandboxCreateStmt(s *ast.SandboxCreateStmt) {
	nameStr := g.genStringConst(s.Name)

	var bRef = "null"
	if slot, ok := g.vars[s.Boundary]; ok {
		t := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
		bRef = t
	}

	capVal := 10
	if s.Capacity != nil {
		if lit, ok := s.Capacity.(*ast.IntLit); ok {
			capVal = int(lit.Value)
		}
	}

	sbTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_sandbox_create(i8* %s, i8* %s, i32 %d)\n", sbTemp, nameStr, bRef, capVal)

	slot, ok := g.vars[s.Name]
	if !ok {
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		slot = &varSlot{
			ptr:      alloca,
			llvmType: "i8*",
		}
		g.vars[s.Name] = slot
	}
	fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", sbTemp, slot.ptr)
}

func (g *Generator) genSandboxRunStmt(s *ast.SandboxRunStmt) {
	var sbRef = "null"
	if slot, ok := g.vars[s.Sandbox]; ok {
		t := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
		sbRef = t
	} else {
		sbStr := g.genStringConst(s.Sandbox)
		t := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_sandbox_find(i8* %s)\n", t, sbStr)
		sbRef = t
	}

	toolStr := g.genStringConst(s.ToolName)

	var argRef = "null"
	if len(s.Args) > 0 {
		argVal := g.genExpr(s.Args[0])
		if argVal.Type == "i8*" {
			argRef = argVal.Ref
		}
	} else {
		argRef = g.genStringConst("")
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_sandbox_run(i8* %s, i8* %s, i8* %s)\n", resTemp, sbRef, toolStr, argRef)
}

func (g *Generator) genSandboxVerifyStmt(s *ast.SandboxVerifyStmt) {
	sbStr := g.genStringConst(s.Sandbox)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_sandbox_verify_text(i8* %s)\n", resTemp, sbStr)

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

func (g *Generator) genSandboxVerifyExpr(e *ast.SandboxVerifyExpr) Value {
	sbStr := g.genStringConst(e.Sandbox)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_sandbox_verify_text(i8* %s)\n", resTemp, sbStr)
	return Value{
		Type:  "i8*",
		Ref:   resTemp,
		QType: types.Text,
	}
}
