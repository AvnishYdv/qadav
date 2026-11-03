package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkEnergyDecl(e *ast.EnergyDecl) {
	_ = c.curScope.Define(&Symbol{
		Name: e.Name,
		Type: types.Float64,
		Pos:  e.Pos(),
	})
	if e.Fn != nil {
		c.enterScope()
		for _, stmt := range e.Fn.Statements {
			c.checkStmt(stmt)
		}
		c.leaveScope()
	}
}

func (c *Checker) checkMinimizeStmt(m *ast.MinimizeStmt) {
	gridType := c.checkExpr(m.Grid)
	if gridType != nil {
		if !gridType.IsTensor() || gridType.TensorElem() == nil || gridType.TensorElem().Kind != types.KindQuasi {
			c.errorf(m.Grid.Pos(), "minimize grid must be tensor⟨quasi⟩, got %s", gridType)
		}
	}
	if ident, ok := m.Energy.(*ast.Ident); ok {
		if ident.Name != "ising" && ident.Name != "majority" {
			c.checkExpr(m.Energy)
		}
	} else {
		c.checkExpr(m.Energy)
	}
	iterType := c.checkExpr(m.Iterations)
	if iterType != nil && !iterType.IsInteger() {
		c.errorf(m.Iterations.Pos(), "minimize iterations must be integer, got %s", iterType)
	}
	if intLit, ok := m.Iterations.(*ast.IntLit); ok {
		if intLit.Value <= 0 {
			c.errorf(m.Iterations.Pos(), "minimize iterations must be positive, got %d", intLit.Value)
		} else if intLit.Value > 100000 {
			c.errorf(m.Iterations.Pos(), "minimize iterations exceed maximum limit of 100,000, got %d", intLit.Value)
		}
	}
}

func (c *Checker) checkTileStmt(t *ast.TileStmt) {
	gridType := c.checkExpr(t.Grid)
	if gridType != nil && !gridType.IsTensor() {
		c.errorf(t.Grid.Pos(), "tile grid must be tensor, got %s", gridType)
	}
	c.checkExpr(t.Pattern)
}
