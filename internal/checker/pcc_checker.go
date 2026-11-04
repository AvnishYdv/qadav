package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkPCCAttestStmt(s *ast.PCCAttestStmt) {
	if s == nil {
		return
	}
	if s.Ast == "" {
		c.errorf(s.Pos(), "pcc_attest requires ast")
	}
	if s.Codegen == "" {
		c.errorf(s.Pos(), "pcc_attest requires codegen")
	}
	if s.IR == "" {
		c.errorf(s.Pos(), "pcc_attest requires ir")
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkPCCVerifyStmt(s *ast.PCCVerifyStmt) {
	if s == nil {
		return
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined PCC certificate: %s", s.Cert)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptPCCStmt(s *ast.CorruptPCCStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_pcc is only permitted in test mode (--test)")
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined PCC certificate: %s", s.Cert)
		}
	}
	if s.Field != nil {
		c.checkExpr(s.Field)
	}
}
