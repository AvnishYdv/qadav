package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkAllReduceStmt(s *ast.AllReduceStmt) {
	if s == nil {
		return
	}
	for _, grad := range s.GradList {
		if grad != nil {
			c.checkExpr(grad)
		}
	}
	if s.Target != nil {
		c.checkExpr(s.Target)
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkPserverInitStmt(s *ast.PserverInitStmt) {
	if s == nil {
		return
	}
	if s.Workers != nil {
		c.checkExpr(s.Workers)
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkPserverAggregateStmt(s *ast.PserverAggregateStmt) {
	if s == nil {
		return
	}
	if s.Pserver != "" {
		if _, ok := c.curScope.Lookup(s.Pserver); !ok {
			c.errorf(s.Pos(), "undefined parameter server: %s", s.Pserver)
		}
	}
}

