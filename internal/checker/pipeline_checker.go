package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkPipelineClaimStmt(s *ast.PipelineClaimStmt) {
	if s == nil {
		return
	}
	if s.Stages < 1 || s.Stages > 16 {
		c.errorf(s.Pos(), "pipeline stages must be between 1 and 16, got %d", s.Stages)
	}
	if s.MicroBatch < 1 {
		c.errorf(s.Pos(), "pipeline micro_batch must be at least 1, got %d", s.MicroBatch)
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Int64,
		})
	}
}

func (c *Checker) checkPipelineRunStmt(s *ast.PipelineRunStmt) {
	if s == nil {
		return
	}
	if s.Plan != "" {
		if _, ok := c.curScope.Lookup(s.Plan); !ok {
			c.errorf(s.Pos(), "undefined pipeline plan: %s", s.Plan)
		}
	}
	if s.Input != nil {
		c.checkExpr(s.Input)
	}
	if s.Output != nil {
		if id, ok := s.Output.(*ast.Ident); ok {
			if _, ok := c.curScope.Lookup(id.Name); !ok {
				c.curScope.Define(&Symbol{Name: id.Name, Type: types.TensorOfShape(types.Float64, []int{2, 2})})
			}
		} else {
			c.checkExpr(s.Output)
		}
	}
}
