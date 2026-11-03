package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkCertifiedReasonStmt(s *ast.CertifiedReasonStmt) {
	if s == nil {
		return
	}
	if s.Chain == nil {
		if c.activeReasoningChain != nil {
			s.Chain = c.activeReasoningChain
		} else {
			c.errorf(s.Pos(), "no active reasoning_chain declared in scope")
		}
	} else {
		chType := c.checkExpr(s.Chain)
		if chType != nil && chType.Kind != types.KindReasoningChain {
			c.errorf(s.Chain.Pos(), "target must be a reasoning_chain, got %s", chType)
		}
	}

	if s.Kind != token.DEDUCE && s.Kind != token.INDUCE && s.Kind != token.ABDUCE && s.Kind != token.VERIFY {
		c.errorf(s.Pos(), "invalid certified_reason kind %v: must be deduce, induce, abduce, or verify", s.Kind)
	}

	if s.InputExpr != nil {
		c.checkExpr(s.InputExpr)
	} else {
		c.errorf(s.Pos(), "certified_reason requires input premise")
	}

	if s.OutputExpr != nil {
		c.checkExpr(s.OutputExpr)
	} else {
		c.errorf(s.Pos(), "certified_reason requires output conclusion")
	}

	if s.Witness == nil {
		c.errorf(s.Pos(), "certified_reason requires non-empty witness")
	} else {
		wType := c.checkExpr(s.Witness)
		if slit, ok := s.Witness.(*ast.StringLit); ok {
			if len(slit.Value) == 0 {
				c.errorf(s.Witness.Pos(), "certified_reason witness cannot be empty")
			} else if len(slit.Value) > 512 {
				c.errorf(s.Witness.Pos(), "certified_reason witness exceeds maximum 512 bytes (%d)", len(slit.Value))
			}
		} else if wType != nil && wType.Kind != types.KindText && wType.Kind != types.KindConcept {
			c.errorf(s.Witness.Pos(), "certified_reason witness must be a string or byte literal, got %s", wType)
		}
	}

	if s.Confidence != nil {
		confType := c.checkExpr(s.Confidence)
		if lit, ok := s.Confidence.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(s.Confidence.Pos(), "confidence must be in range [0.0, 1.0], got %f", lit.Value)
			}
		} else if confType != nil && !confType.IsFloat() {
			c.errorf(s.Confidence.Pos(), "confidence must be a float, got %s", confType)
		}
	}
}

func (c *Checker) checkAbstainThresholdDecl(d *ast.AbstainThresholdDecl) {
	if d == nil {
		return
	}
	if d.Chain == nil {
		if c.activeReasoningChain != nil {
			d.Chain = c.activeReasoningChain
		} else {
			c.errorf(d.Pos(), "no active reasoning_chain declared in scope")
		}
	} else {
		chType := c.checkExpr(d.Chain)
		if chType != nil && chType.Kind != types.KindReasoningChain {
			c.errorf(d.Chain.Pos(), "target must be a reasoning_chain, got %s", chType)
		}
	}

	if d.Threshold != nil {
		tType := c.checkExpr(d.Threshold)
		if lit, ok := d.Threshold.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(d.Threshold.Pos(), "abstain_threshold must be in range [0.0, 1.0], got %f", lit.Value)
			}
		} else if tType != nil && !tType.IsFloat() {
			c.errorf(d.Threshold.Pos(), "abstain_threshold must be a float, got %s", tType)
		}
	}

	if d.Mode != token.RETRY && d.Mode != token.ESCALATE && d.Mode != token.REFUSE {
		c.errorf(d.Pos(), "invalid abstain mode %v: must be retry, escalate, or refuse", d.Mode)
	}

	if d.MaxRetries != nil {
		mrType := c.checkExpr(d.MaxRetries)
		if lit, ok := d.MaxRetries.(*ast.IntLit); ok {
			if lit.Value < 1 {
				c.errorf(d.MaxRetries.Pos(), "max_retries must be positive, got %d", lit.Value)
			}
		} else if mrType != nil && !mrType.IsInteger() {
			c.errorf(d.MaxRetries.Pos(), "max_retries must be an integer, got %s", mrType)
		}
	}
}

func (c *Checker) checkAbstainCheckStmt(s *ast.AbstainCheckStmt) {
	if s == nil {
		return
	}
	if s.Chain == nil {
		if c.activeReasoningChain != nil {
			s.Chain = c.activeReasoningChain
		} else {
			c.errorf(s.Pos(), "no active reasoning_chain declared in scope")
		}
	} else {
		chType := c.checkExpr(s.Chain)
		if chType != nil && chType.Kind != types.KindReasoningChain {
			c.errorf(s.Chain.Pos(), "target must be a reasoning_chain, got %s", chType)
		}
	}

	if s.Into != "" && c.curScope != nil {
		_ = c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
			Pos:  s.Pos(),
		})
	}
}

func (c *Checker) checkAbstainCheckExpr(e *ast.AbstainCheckExpr) *types.Type {
	if e == nil {
		return types.Int64
	}
	if e.Chain == nil {
		if c.activeReasoningChain != nil {
			e.Chain = c.activeReasoningChain
		} else {
			c.errorf(e.Pos(), "no active reasoning_chain declared in scope")
		}
	} else {
		chType := c.checkExpr(e.Chain)
		if chType != nil && chType.Kind != types.KindReasoningChain {
			c.errorf(e.Chain.Pos(), "target must be a reasoning_chain, got %s", chType)
		}
	}
	return types.Int64
}
