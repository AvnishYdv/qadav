package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkReasoningChainDecl(d *ast.ReasoningChainDecl) {
	if d.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: d.Name,
			Type: types.ReasoningChain,
			Pos:  d.Pos(),
		})
		c.activeReasoningChain = &ast.Ident{Pos_: d.Pos(), Name: d.Name}
	}

	if d.Boundary != nil {
		bType := c.checkExpr(d.Boundary)
		if bType != nil && bType.Kind != types.KindEthicalBoundary {
			c.errorf(d.Boundary.Pos(), "reasoning_chain boundary must be an ethical_boundary, got %s", bType)
		}
	}

	if d.Capacity != nil {
		capType := c.checkExpr(d.Capacity)
		if lit, ok := d.Capacity.(*ast.IntLit); ok {
			if lit.Value < 1 || lit.Value > 10000 {
				c.errorf(d.Capacity.Pos(), "reasoning_chain capacity must be between 1 and 10000, got %d", lit.Value)
			}
		} else if capType != nil && !capType.IsInteger() {
			c.errorf(d.Capacity.Pos(), "reasoning_chain capacity must be an integer, got %s", capType)
		}
	}
}

func (c *Checker) checkReasonStepStmt(s *ast.ReasonStepStmt) {
	// If chain not explicitly provided in statement, use active reasoning chain
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

	// Validate operation
	switch s.Operation {
	case token.DEDUCE, token.INDUCE, token.ABDUCE, token.VERIFY:
		// Valid
	default:
		c.errorf(s.Pos(), "invalid reasoning operation: %v (must be deduce, induce, abduce, or verify)", s.Operation)
	}

	if s.InputExpr != nil {
		_ = c.checkExpr(s.InputExpr)
	}
	if s.OutputExpr != nil {
		_ = c.checkExpr(s.OutputExpr)
	}

	if s.Confidence != nil {
		confType := c.checkExpr(s.Confidence)
		if lit, ok := s.Confidence.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(s.Confidence.Pos(), "confidence must be between 0.0 and 1.0, got %f", lit.Value)
			}
		} else if confType != nil && !confType.IsFloat() {
			c.errorf(s.Confidence.Pos(), "confidence must be a float, got %s", confType)
		}
	}
}

func (c *Checker) checkVerifyChainStmt(s *ast.VerifyChainStmt) {
	chType := c.checkExpr(s.Chain)
	if chType != nil && chType.Kind != types.KindReasoningChain {
		c.errorf(s.Chain.Pos(), "verify_chain target must be reasoning_chain, got %s", chType)
	}
}

func (c *Checker) checkAuditReasoningStmt(s *ast.AuditReasoningStmt) {
	chType := c.checkExpr(s.Chain)
	if chType != nil && chType.Kind != types.KindReasoningChain {
		c.errorf(s.Chain.Pos(), "audit_reasoning target must be reasoning_chain, got %s", chType)
	}
}

func (c *Checker) checkCorruptReasoningStepStmt(s *ast.CorruptReasoningStepStmt) {
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_reasoning_step is test-only, use --test flag")
		return
	}
	chType := c.checkExpr(s.Chain)
	if chType != nil && chType.Kind != types.KindReasoningChain {
		c.errorf(s.Chain.Pos(), "corrupt_reasoning_step target must be reasoning_chain, got %s", chType)
	}
	if s.StepIndex != nil {
		idxType := c.checkExpr(s.StepIndex)
		if idxType != nil && !idxType.IsInteger() {
			c.errorf(s.StepIndex.Pos(), "step index must be an integer, got %s", idxType)
		}
	}
	if s.AtVal != nil {
		_ = c.checkExpr(s.AtVal)
	}
}

