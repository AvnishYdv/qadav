package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkBaratSubmitStmt(s *ast.BaratSubmitStmt) {
	if s == nil {
		return
	}
	if s.Worker != nil {
		c.checkExpr(s.Worker)
	}
	if s.Gradient != nil {
		c.checkExpr(s.Gradient)
	}
	if s.Group != "" {
		if _, ok := c.curScope.Lookup(s.Group); !ok {
			c.curScope.Define(&Symbol{
				Name: s.Group,
				Type: types.Int64,
			})
		}
	}
}

func (c *Checker) checkBaratAggregateStmt(s *ast.BaratAggregateStmt) {
	if s == nil {
		return
	}
	if s.Group != "" {
		if _, ok := c.curScope.Lookup(s.Group); !ok {
			c.errorf(s.Pos(), "undefined barat group: %s", s.Group)
		}
	}
	if s.IntoGrad != "" {
		c.curScope.Define(&Symbol{
			Name: s.IntoGrad,
			Type: types.Int64,
		})
	}
	if s.IntoCert != "" {
		c.curScope.Define(&Symbol{
			Name: s.IntoCert,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkBaratVerifyStmt(s *ast.BaratVerifyStmt) {
	if s == nil {
		return
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined barat cert: %s", s.Cert)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptBaratStmt(s *ast.CorruptBaratStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_barat is only permitted in test mode (--test)")
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined barat cert: %s", s.Cert)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}

