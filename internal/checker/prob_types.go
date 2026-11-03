package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

// ConfidenceType represents Confident<T> — a value with confidence score
// Part of CIR (Cognitive Intermediate Representation) — Invention #4
type ConfidenceType struct {
	Inner   *types.Type
	MinConf float64 // minimum confidence (0.0-1.0)
}

// ExploreType represents Explore<T> — a value in exploration mode
// Part of CIR (Cognitive Intermediate Representation) — Invention #4
type ExploreType struct {
	Inner *types.Type
}

// CheckProbabilisticType validates confidence score and constructs a Confident<T> type.
func (c *Checker) CheckProbabilisticType(pos ast.Position, baseType *types.Type, conf float64) *types.Type {
	if conf < 0.0 || conf > 1.0 {
		c.errorf(pos, "confidence score must be between 0.0 and 1.0, got %f", conf)
		return baseType
	}
	return types.ConfidentOf(baseType, conf)
}
