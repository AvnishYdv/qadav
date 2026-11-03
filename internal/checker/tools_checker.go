package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

type ToolInfo struct {
	Name          string
	Boundary      ast.Expr
	PreCondition  ast.Expr
	PostCondition ast.Expr
}

func (c *Checker) checkToolRegisterDecl(decl *ast.ToolRegisterDecl) {
	if decl == nil {
		return
	}
	if c.tools == nil {
		c.tools = make(map[string]*ToolInfo)
	}

	if decl.Name != "" && c.curScope != nil {
		_ = c.curScope.Define(&Symbol{
			Name: decl.Name,
			Type: types.Int64,
			Pos:  decl.Pos(),
		})
	}

	if decl.Boundary != nil {
		c.checkExpr(decl.Boundary)
	}
	if decl.PreCondition != nil {
		c.checkExpr(decl.PreCondition)
	}
	if decl.PostCondition != nil {
		c.checkExpr(decl.PostCondition)
	}

	c.tools[decl.Name] = &ToolInfo{
		Name:          decl.Name,
		Boundary:      decl.Boundary,
		PreCondition:  decl.PreCondition,
		PostCondition: decl.PostCondition,
	}
}

func (c *Checker) checkToolCallStmt(stmt *ast.ToolCallStmt) {
	if stmt == nil {
		return
	}
	if stmt.Target != nil {
		c.checkExpr(stmt.Target)
	}
	for _, arg := range stmt.Args {
		c.checkExpr(arg)
	}
	if stmt.Audit != nil {
		c.checkExpr(stmt.Audit)
	}
}

func (c *Checker) checkToolCallExpr(expr *ast.ToolCallExpr) *types.Type {
	if expr == nil {
		return types.Int64
	}
	for _, arg := range expr.Args {
		c.checkExpr(arg)
	}
	if expr.Audit != nil {
		c.checkExpr(expr.Audit)
	}
	resType := types.Int64
	c.exprTypes[expr] = resType
	return resType
}
