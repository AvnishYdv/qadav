package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkEthicalBoundaryDecl(b *ast.EthicalBoundaryDecl) {
	if b.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: b.Name,
			Type: types.EthicalBoundary,
			Pos:  b.Pos(),
		})
	}
	for _, rule := range b.Rules {
		if rule.Condition != nil {
			_ = c.checkExpr(rule.Condition)
		}
		if rule.Penalty < 0.0 || rule.Penalty > 1.0 {
			c.errorf(rule.Pos(), "ethical rule penalty must be between 0.0 and 1.0, got %f", rule.Penalty)
		}
	}
}

func (c *Checker) checkEnforceStmt(s *ast.EnforceStmt) {
	bType := c.checkExpr(s.Boundary)
	if bType != nil && bType.Kind != types.KindEthicalBoundary {
		c.errorf(s.Boundary.Pos(), "enforce target must be ethical_boundary, got %s", bType)
	}
	_ = c.checkExpr(s.Action)
}

func (c *Checker) checkDecoherenceStmt(s *ast.DecoherenceStmt) {
	if s.Target != nil {
		_ = c.checkExpr(s.Target)
	}
}
