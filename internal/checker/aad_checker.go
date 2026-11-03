package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkAADCertStmt(s *ast.AADCertStmt) {
	if s == nil {
		return
	}
	if s.Model != nil {
		c.checkExpr(s.Model)
	}
	if s.TrainSet != nil {
		c.checkExpr(s.TrainSet)
	}
	if s.MaxSyndrome != nil {
		c.checkExpr(s.MaxSyndrome)
		if iLit, ok := s.MaxSyndrome.(*ast.IntLit); ok {
			if iLit.Value < 1 || iLit.Value > 128 {
				c.errorf(s.Pos(), "max_syndrome must be in [1, 128], got %d", iLit.Value)
			}
		}
	}
	if s.MaxCorrDist != nil {
		c.checkExpr(s.MaxCorrDist)
		if iLit, ok := s.MaxCorrDist.(*ast.IntLit); ok {
			if iLit.Value < 1 || iLit.Value > 16 {
				c.errorf(s.Pos(), "max_corr_dist must be in [1, 16], got %d", iLit.Value)
			}
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkAADAttestStmt(s *ast.AADAttestStmt) {
	if s == nil {
		return
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined AAD training certificate: %s", s.Cert)
		}
	}
	if s.Decode != "" {
		if _, ok := c.curScope.Lookup(s.Decode); !ok {
			c.errorf(s.Pos(), "undefined QEC decode result: %s", s.Decode)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkMetaFitStmt(s *ast.MetaFitStmt) {
	if s == nil {
		return
	}
	if s.Cert != "" {
		if _, ok := c.curScope.Lookup(s.Cert); !ok {
			c.errorf(s.Pos(), "undefined AAD training certificate: %s", s.Cert)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkMetaAdaptStmt(s *ast.MetaAdaptStmt) {
	if s == nil {
		return
	}
	if s.State != "" {
		if _, ok := c.curScope.Lookup(s.State); !ok {
			c.errorf(s.Pos(), "undefined meta state: %s", s.State)
		}
	}
	if s.Envelope != "" {
		if _, ok := c.curScope.Lookup(s.Envelope); !ok {
			c.errorf(s.Pos(), "undefined AAD decode envelope: %s", s.Envelope)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkMetaVerifyStmt(s *ast.MetaVerifyStmt) {
	if s == nil {
		return
	}
	if s.State != "" {
		if _, ok := c.curScope.Lookup(s.State); !ok {
			c.errorf(s.Pos(), "undefined meta state: %s", s.State)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptAADEnvStmt(s *ast.CorruptAADEnvStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_aad_env is only permitted in test mode (--test)")
	}
	if s.Env != "" {
		if _, ok := c.curScope.Lookup(s.Env); !ok {
			c.errorf(s.Pos(), "undefined AAD decode envelope: %s", s.Env)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}
