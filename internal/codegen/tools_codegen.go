package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genToolRegisterDecl(d *ast.ToolRegisterDecl) {
	nameStr := g.genStringConst(d.Name)

	bRef := "null"
	if d.Boundary != nil {
		bVal := g.genExpr(d.Boundary)
		bRef = bVal.Ref
	} else {
		for _, slot := range g.vars {
			if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
				t := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
				bRef = t
				break
			}
		}
	}

	hasPre := 0
	if d.PreCondition != nil {
		hasPre = 1
	}
	hasPost := 0
	if d.PostCondition != nil {
		hasPost = 1
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_tool_register(i8* %s, i8* %s, i32 %d, i32 %d)\n",
		resTemp, nameStr, bRef, hasPre, hasPost)
}

func (g *Generator) genToolCallStmt(s *ast.ToolCallStmt) {
	nameStr := g.genStringConst(s.ToolName)
	argStr := "null"
	if len(s.Args) > 0 {
		argVal := g.genExpr(s.Args[0])
		argStr = argVal.Ref
	}

	auditRef := "null"
	if s.Audit != nil {
		auditVal := g.genExpr(s.Audit)
		auditRef = auditVal.Ref
	}

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
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_tool_call(i8* %s, i8* %s, i8* %s, i8* %s)\n",
		resTemp, nameStr, bRef, argStr, auditRef)

	if s.Target != nil {
		if id, ok := s.Target.(*ast.Ident); ok {
			ptr := fmt.Sprintf("%%%s.ptr", id.Name)
			fmt.Fprintf(&g.sb, "  %s = alloca i32\n", ptr)
			slot := &varSlot{
				ptr:      ptr,
				llvmType: "i32",
				qtype:    types.Int64,
			}
			g.vars[id.Name] = slot
			fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", resTemp, ptr)
		}
	}
}

func (g *Generator) genToolCallExpr(e *ast.ToolCallExpr) Value {
	nameStr := g.genStringConst(e.ToolName)
	argStr := "null"
	if len(e.Args) > 0 {
		argVal := g.genExpr(e.Args[0])
		argStr = argVal.Ref
	}

	auditRef := "null"
	if e.Audit != nil {
		auditVal := g.genExpr(e.Audit)
		auditRef = auditVal.Ref
	}

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
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_tool_call(i8* %s, i8* %s, i8* %s, i8* %s)\n",
		resTemp, nameStr, bRef, argStr, auditRef)

	convTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", convTemp, resTemp)
	return Value{Type: "i64", Ref: convTemp, QType: types.Int64}
}

