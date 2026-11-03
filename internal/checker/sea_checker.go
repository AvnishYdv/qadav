package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkEvolveStmt(s *ast.EvolveStmt) {
	if s == nil {
		return
	}
	if s.FromState != nil {
		c.checkExpr(s.FromState)
	}
	if s.ToState != nil {
		c.checkExpr(s.ToState)
	}
	if s.Reason != nil {
		c.checkExpr(s.Reason)
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkEvolveVerifyStmt(s *ast.EvolveVerifyStmt) {
	if s == nil {
		return
	}
	if s.Proposals != "" {
		if _, ok := c.curScope.Lookup(s.Proposals); !ok {
			c.errorf(s.Pos(), "undefined proposal: %s", s.Proposals)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkEvolveHistoryStmt(s *ast.EvolveHistoryStmt) {
	if s == nil {
		return
	}
	if s.Chain != "" {
		if _, ok := c.curScope.Lookup(s.Chain); !ok {
			c.errorf(s.Pos(), "undefined reasoning chain: %s", s.Chain)
		}
	}
}

func (c *Checker) checkCorruptEvolveStmt(s *ast.CorruptEvolveStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_evolve is only permitted in test mode (--test)")
	}
	if s.Proposal != "" {
		if _, ok := c.curScope.Lookup(s.Proposal); !ok {
			c.errorf(s.Pos(), "undefined proposal: %s", s.Proposal)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}

