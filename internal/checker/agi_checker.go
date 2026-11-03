package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkMemoryDecl(m *ast.MemoryDecl) {
	if m.Capacity != nil {
		capType := c.checkExpr(m.Capacity)
		if capType != nil && !capType.IsInteger() {
			c.errorf(m.Capacity.Pos(), "memory capacity must be integer, got %s", capType)
		}
	}
	_ = c.curScope.Define(&Symbol{
		Name: m.Name,
		Type: types.Memory,
		Pos:  m.Pos(),
	})
}

func (c *Checker) checkRememberStmt(s *ast.RememberStmt) {
	memType := c.checkExpr(s.Memory)
	if memType != nil && memType.Kind != types.KindMemory && memType.Kind != types.KindMemoryProv {
		c.errorf(s.Memory.Pos(), "remember target must be memory, got %s", memType)
	}
	_ = c.checkExpr(s.Key)
	_ = c.checkExpr(s.Value)
}

func (c *Checker) checkRecallExpr(e *ast.RecallExpr) *types.Type {
	memType := c.checkExpr(e.Memory)
	if memType != nil && memType.Kind != types.KindMemory && memType.Kind != types.KindMemoryProv {
		c.errorf(e.Memory.Pos(), "recall target must be memory, got %s", memType)
	}
	_ = c.checkExpr(e.Key)
	if memType != nil && memType.Kind == types.KindMemoryProv {
		return types.Text
	}
	return types.Int64
}

func (c *Checker) checkGraphDecl(g *ast.GraphDecl) {
	_ = c.curScope.Define(&Symbol{
		Name: g.Name,
		Type: types.Graph,
		Pos:  g.Pos(),
	})
}

func (c *Checker) checkNodeStmt(s *ast.NodeStmt) {
	gType := c.checkExpr(s.Graph)
	if gType != nil && gType.Kind != types.KindGraph {
		c.errorf(s.Graph.Pos(), "node target must be graph, got %s", gType)
	}
	_ = c.checkExpr(s.ID)
	if s.Data != nil {
		_ = c.checkExpr(s.Data)
	}
}

func (c *Checker) checkLinkStmt(s *ast.LinkStmt) {
	gType := c.checkExpr(s.Graph)
	if gType != nil && gType.Kind != types.KindGraph {
		c.errorf(s.Graph.Pos(), "link target must be graph, got %s", gType)
	}
	_ = c.checkExpr(s.From)
	_ = c.checkExpr(s.To)
	if s.Weight != nil {
		_ = c.checkExpr(s.Weight)
	}
}

func (c *Checker) checkTraceExpr(e *ast.TraceExpr) *types.Type {
	gType := c.checkExpr(e.Graph)
	if gType != nil && gType.Kind != types.KindGraph {
		c.errorf(e.Graph.Pos(), "trace target must be graph, got %s", gType)
	}
	_ = c.checkExpr(e.Start)
	if e.Depth != nil {
		_ = c.checkExpr(e.Depth)
	}
	return types.TensorOf(types.Int64)
}

func (c *Checker) checkProcedureDecl(p *ast.ProcedureDecl) {
	var paramTypes []*types.Type
	for _, param := range p.Params {
		pt := c.resolveType(param.Type)
		if pt == nil {
			pt = types.Int64
		}
		paramTypes = append(paramTypes, pt)
	}

	retType := c.resolveType(p.RetType)
	if retType == nil {
		retType = types.Int64
	}

	procType := &types.Type{
		Kind:    types.KindProcedure,
		Name:    p.Name,
		Params:  paramTypes,
		Results: []*types.Type{retType},
	}

	_ = c.globalScope.Define(&Symbol{
		Name: p.Name,
		Type: procType,
		Pos:  p.Pos(),
	})

	c.enterScope()
	defer c.leaveScope()

	oldRet := c.curRetType
	c.curRetType = retType
	defer func() { c.curRetType = oldRet }()

	for i, param := range p.Params {
		pSym := &Symbol{
			Name:    param.Name,
			Type:    paramTypes[i],
			Pos:     param.Pos(),
			IsParam: true,
		}
		_ = c.curScope.Define(pSym)
	}

	if p.Body != nil {
		for _, stmt := range p.Body.Statements {
			c.checkStmt(stmt)
		}
	}
}

func (c *Checker) checkExecuteExpr(e *ast.ExecuteExpr) *types.Type {
	pType := c.checkExpr(e.Procedure)
	for _, arg := range e.Args {
		c.checkExpr(arg)
	}
	if pType != nil && len(pType.Results) > 0 {
		return pType.Results[0]
	}
	return types.Int64
}

func (c *Checker) checkExecuteStmt(s *ast.ExecuteStmt) {
	_ = c.checkExpr(s.Procedure)
	for _, arg := range s.Args {
		c.checkExpr(arg)
	}
}
