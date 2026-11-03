package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkModalityDecl(d *ast.ModalityDecl) {
	if d.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: d.Name,
			Type: types.Modality,
			Pos:  d.Pos(),
		})
	}

	if d.Source != nil {
		srcType := c.checkExpr(d.Source)
		switch d.Kind {
		case token.TENSOR_INPUT:
			if srcType != nil && srcType.Kind != types.KindTensor {
				// Warn or check if not tensor
			}
		case token.TEXT_INPUT:
			if srcType != nil && srcType.Kind != types.KindText {
				// Warn or check if not text
			}
		case token.CONCEPT_INPUT:
			if srcType != nil && srcType.Kind != types.KindConcept && !srcType.IsInteger() {
				// Warn or check if not concept
			}
		}
	}
}

func (c *Checker) checkCrossModalReasonStmt(s *ast.CrossModalReasonStmt) {
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

	switch s.Operation {
	case token.DEDUCE, token.INDUCE, token.ABDUCE, token.VERIFY:
		// Valid
	default:
		c.errorf(s.Pos(), "invalid reasoning operation: %v (must be deduce, induce, abduce, or verify)", s.Operation)
	}

	for _, m := range s.Modalities {
		mType := c.checkExpr(m)
		if mType != nil && mType.Kind != types.KindModality {
			c.errorf(m.Pos(), "cross_modal_reason input must be a modality, got %s", mType)
		}
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

