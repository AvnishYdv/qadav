package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func TestToolsChecker_RegisterAndCall(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "sec_boundary", Type: types.EthicalBoundary})
	c.globalScope.Define(&Symbol{Name: "res", Type: types.Int64})
	c.globalScope.Define(&Symbol{Name: "chain1", Type: types.ReasoningChainT()})

	// 1. Tool register declaration
	reg := &ast.ToolRegisterDecl{
		Pos_:          pos,
		Name:          "file_reader",
		Boundary:      &ast.Ident{Pos_: pos, Name: "sec_boundary"},
		PreCondition:  &ast.IntLit{Pos_: pos, Value: 1},
		PostCondition: &ast.IntLit{Pos_: pos, Value: 1},
	}
	c.checkToolRegisterDecl(reg)

	// 2. Tool call statement
	callStmt := &ast.ToolCallStmt{
		Pos_:     pos,
		Target:   &ast.Ident{Pos_: pos, Name: "res"},
		ToolName: "file_reader",
		Args:     []ast.Expr{&ast.StringLit{Pos_: pos, Value: "/data"}},
		Audit:    &ast.Ident{Pos_: pos, Name: "chain1"},
	}
	c.checkToolCallStmt(callStmt)

	// 3. Tool call expression
	callExpr := &ast.ToolCallExpr{
		Pos_:     pos,
		ToolName: "file_reader",
		Args:     []ast.Expr{&ast.StringLit{Pos_: pos, Value: "/data"}},
		Audit:    &ast.Ident{Pos_: pos, Name: "chain1"},
	}
	ty := c.checkToolCallExpr(callExpr)
	if ty != types.Int64 {
		t.Fatalf("expected int64, got %v", ty)
	}

	// Nil safety
	c.checkToolRegisterDecl(nil)
	c.checkToolCallStmt(nil)
	c.checkToolCallExpr(nil)
}
