package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkCTSAAttestStmt(s *ast.CTSAAttestStmt) {
	if s == nil {
		return
	}
	if s.Ast == "" {
		c.errorf(s.Pos(), "ctsa_attest requires ast")
	}
	if s.Env == "" {
		c.errorf(s.Pos(), "ctsa_attest requires env")
	}
	if s.Rules == "" {
		c.errorf(s.Pos(), "ctsa_attest requires rules")
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkCTSAVerifyStmt(s *ast.CTSAVerifyStmt) {
	if s == nil {
		return
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined CTSA certificate: %s", s.Cert)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptCTSAStmt(s *ast.CorruptCTSAStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_ctsa is only permitted in test mode (--test)")
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined CTSA certificate: %s", s.Cert)
		}
	}
	if s.Field != nil {
		c.checkExpr(s.Field)
	}
}
