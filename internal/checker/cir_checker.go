package checker

import (
	"math"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkAssertStmt(s *ast.AssertStmt) {
	valType := c.checkExpr(s.Value)
	if valType != nil && valType.Kind != types.KindConfident && !valType.IsNumeric() {
		c.errorf(s.Value.Pos(), "assert target must be confident or numeric, got %s", valType)
	}

	threshType := c.checkExpr(s.Threshold)
	if threshType != nil && !threshType.IsNumeric() {
		c.errorf(s.Threshold.Pos(), "assert threshold must be numeric, got %s", threshType)
	}

	if fLit, ok := s.Threshold.(*ast.FloatLit); ok {
		if fLit.Value < 0.0 || fLit.Value > 1.0 {
			c.errorf(s.Threshold.Pos(), "assert threshold must be between 0.0 and 1.0, got %f", fLit.Value)
		}
	}
}

func (c *Checker) checkInferStmt(s *ast.InferStmt) {
	mType := c.checkExpr(s.Model)
	if mType != nil && mType.Kind != types.KindNetwork && mType.Kind != types.KindTensor {
		c.errorf(s.Pos_, "infer model must be a network or tensor, got %s", mType)
	}

	_ = c.checkExpr(s.Input)

	if s.Confidence != nil {
		c.checkExpr(s.Confidence)
	}

	_ = c.curScope.Define(&Symbol{
		Name: s.Into,
		Type: types.TensorOf(types.Float64),
		Pos:  s.Pos(),
	})
}

func (c *Checker) checkPropagateStmt(s *ast.PropagateStmt) {
	_ = c.checkExpr(s.Value)
	_ = c.checkExpr(s.Source)
}

func (c *Checker) checkConfidentLit(e *ast.ConfidentLit) *types.Type {
	valType := c.checkExpr(e.Value)
	if valType == nil {
		valType = types.Float64
	}
	if e.Score < 0.0 || e.Score > 1.0 {
		c.errorf(e.Pos(), "confidence score must be between 0.0 and 1.0, got %f", e.Score)
	}
	return types.ConfidentOf(valType, e.Score)
}

// CheckConfidentBinary computes the propagation type for confidence arithmetic.
func CheckConfidentBinary(a, b *types.Type, isAdd bool) *types.Type {
	scoreA := 1.0
	scoreB := 1.0
	elemA := a
	elemB := b

	if a.Kind == types.KindConfident {
		scoreA = a.Confidence
		if a.Elem != nil {
			elemA = a.Elem
		}
	}
	if b.Kind == types.KindConfident {
		scoreB = b.Confidence
		if b.Elem != nil {
			elemB = b.Elem
		}
	}

	var resScore float64
	if isAdd {
		resScore = math.Sqrt(scoreA * scoreB)
	} else {
		resScore = scoreA * scoreB
	}

	resElem := types.Promote(elemA, elemB)
	if resElem == nil {
		resElem = types.Float64
	}
	return types.ConfidentOf(resElem, resScore)
}
