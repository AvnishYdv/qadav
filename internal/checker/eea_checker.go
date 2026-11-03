package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkEvolvePolicyStmt(s *ast.EvolvePolicyStmt) {
	if s == nil {
		return
	}
	switch s.Kind {
	case "constant", "decaying", "adaptive":
		// Valid
	default:
		c.errorf(s.Pos(), "invalid evolution policy kind: %q (expected constant, decaying, or adaptive)", s.Kind)
	}

	if s.InitialRate != nil {
		c.checkExpr(s.InitialRate)
		if fLit, ok := s.InitialRate.(*ast.FloatLit); ok {
			if fLit.Value <= 0.0 || fLit.Value > 1.0 {
				c.errorf(s.Pos(), "initial_rate must be in (0.0, 1.0], got %f", fLit.Value)
			}
		}
	}

	if s.TotalBudget != nil {
		c.checkExpr(s.TotalBudget)
		if iLit, ok := s.TotalBudget.(*ast.IntLit); ok && iLit.Value < 1 {
			c.errorf(s.Pos(), "total_budget must be >= 1, got %d", iLit.Value)
		}
	}

	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkEvolveBoundedStmt(s *ast.EvolveBoundedStmt) {
	if s == nil {
		return
	}
	if s.Policy != "" {
		if _, ok := c.curScope.Lookup(s.Policy); !ok {
			c.errorf(s.Pos(), "undefined evolution policy: %s", s.Policy)
		}
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
	if s.MaxBitFlips != nil {
		c.checkExpr(s.MaxBitFlips)
		if iLit, ok := s.MaxBitFlips.(*ast.IntLit); ok && iLit.Value < 1 {
			c.errorf(s.Pos(), "max_flips must be >= 1, got %d", iLit.Value)
		}
	}
	if s.MaxNewSymbols != nil {
		c.checkExpr(s.MaxNewSymbols)
		if iLit, ok := s.MaxNewSymbols.(*ast.IntLit); ok && iLit.Value < 0 {
			c.errorf(s.Pos(), "max_new_symbols must be >= 0, got %d", iLit.Value)
		}
	}

	if s.IntoProposal != "" {
		c.curScope.Define(&Symbol{
			Name: s.IntoProposal,
			Type: types.Int64,
		})
	}
	if s.IntoEnvelope != "" {
		c.curScope.Define(&Symbol{
			Name: s.IntoEnvelope,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkEEAVerifyStmt(s *ast.EEAVerifyStmt) {
	if s == nil {
		return
	}
	if s.Envelope != "" {
		if _, ok := c.curScope.Lookup(s.Envelope); !ok {
			c.errorf(s.Pos(), "undefined evolution envelope: %s", s.Envelope)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkCorruptEEAStmt(s *ast.CorruptEEAStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_eea is only permitted in test mode (--test)")
	}
	if s.Envelope != "" {
		if _, ok := c.curScope.Lookup(s.Envelope); !ok {
			c.errorf(s.Pos(), "undefined evolution envelope: %s", s.Envelope)
		}
	}
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}
