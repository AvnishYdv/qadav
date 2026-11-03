package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkRSCAAttestStmt(s *ast.RSCAAttestStmt) {
	if s == nil {
		return
	}
	if s.Stage < 0 || s.Stage > 2 {
		c.errorf(s.Pos(), "invalid RSCA stage %d: must be 0, 1, or 2", s.Stage)
	}
	if s.Source == "" {
		c.errorf(s.Pos(), "rsca_attest requires source")
	}
	if s.Compiler == "" {
		c.errorf(s.Pos(), "rsca_attest requires compiler")
	}
	if s.Output == "" {
		c.errorf(s.Pos(), "rsca_attest requires output")
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkRSCAVerifyStmt(s *ast.RSCAVerifyStmt) {
	if s == nil {
		return
	}
	if s.Record != "" {
		if _, ok := c.curScope.Lookup(s.Record); !ok {
			c.errorf(s.Pos(), "undefined RSCA record: %s", s.Record)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptRSCAStmt(s *ast.CorruptRSCAStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_rsca is only permitted in test mode (--test)")
	}
	if s.Record != "" {
		if _, ok := c.curScope.Lookup(s.Record); !ok {
			c.errorf(s.Pos(), "undefined RSCA record: %s", s.Record)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}
