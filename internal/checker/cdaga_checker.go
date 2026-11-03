package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkCDAGARecordStmt(s *ast.CDAGARecordStmt) {
	if s == nil {
		return
	}
	for _, grad := range s.GradList {
		if grad != nil {
			c.checkExpr(grad)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkCDAGAAttestStmt(s *ast.CDAGAAttestStmt) {
	if s == nil {
		return
	}
	if s.Record != "" {
		if _, ok := c.curScope.Lookup(s.Record); !ok {
			c.errorf(s.Pos(), "undefined cdaga record: %s", s.Record)
		}
	}
	if s.Chain != "" {
		if _, ok := c.curScope.Lookup(s.Chain); !ok {
			c.errorf(s.Pos(), "undefined chain: %s", s.Chain)
		}
	}
}

func (c *Checker) checkCDAGAVerifyStmt(s *ast.CDAGAVerifyStmt) {
	if s == nil {
		return
	}
	if s.Record != "" {
		if _, ok := c.curScope.Lookup(s.Record); !ok {
			c.errorf(s.Pos(), "undefined cdaga record: %s", s.Record)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptCDAGAStmt(s *ast.CorruptCDAGAStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_cdaga is only permitted in test mode (--test)")
	}
	if s.Record != "" {
		if _, ok := c.curScope.Lookup(s.Record); !ok {
			c.errorf(s.Pos(), "undefined cdaga record: %s", s.Record)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}
