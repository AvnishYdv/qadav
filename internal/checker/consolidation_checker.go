package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkConsolidateStmt(s *ast.ConsolidateStmt) {
	if s.EpisodicMem != nil {
		epType := c.checkExpr(s.EpisodicMem)
		if epType != nil && epType.Kind != types.KindMemoryProv {
			c.errorf(s.EpisodicMem.Pos(), "consolidate source must be a memory_prov, got %s", epType)
		}
	}

	if s.SemanticMem != nil {
		semType := c.checkExpr(s.SemanticMem)
		if semType != nil && semType.Kind != types.KindGraph {
			c.errorf(s.SemanticMem.Pos(), "consolidate target must be a graph, got %s", semType)
		}
	}

	if s.MinConfidence != nil {
		thType := c.checkExpr(s.MinConfidence)
		if lit, ok := s.MinConfidence.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(s.MinConfidence.Pos(), "consolidation threshold must be between 0.0 and 1.0, got %f", lit.Value)
			}
		} else if thType != nil && !thType.IsFloat() {
			c.errorf(s.MinConfidence.Pos(), "consolidation threshold must be a float, got %s", thType)
		}
	}
}

