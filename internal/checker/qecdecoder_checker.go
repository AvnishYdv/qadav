package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkQECDDecodeStmt(s *ast.QECDDecodeStmt) {
	if s == nil {
		return
	}
	if s.Distance != nil {
		c.checkExpr(s.Distance)
		if intLit, ok := s.Distance.(*ast.IntLit); ok {
			if intLit.Value != 3 && intLit.Value != 5 && intLit.Value != 7 {
				c.errorf(s.Pos(), "code distance must be 3, 5, or 7, got %d", intLit.Value)
			}
		}
	}
	if s.SyndromeIn != "" {
		if _, ok := c.curScope.Lookup(s.SyndromeIn); !ok {
			c.errorf(s.Pos(), "undefined syndrome variable: %s", s.SyndromeIn)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}

func (c *Checker) checkQECDVerifyStmt(s *ast.QECDVerifyStmt) {
	if s == nil {
		return
	}
	if s.Result != "" {
		if _, ok := c.curScope.Lookup(s.Result); !ok {
			c.errorf(s.Pos(), "undefined QEC decode result: %s", s.Result)
		}
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Logic,
		})
	}
}

func (c *Checker) checkQECDApplyStmt(s *ast.QECDApplyStmt) {
	if s == nil {
		return
	}
	if s.Result != "" {
		if _, ok := c.curScope.Lookup(s.Result); !ok {
			c.errorf(s.Pos(), "undefined QEC decode result: %s", s.Result)
		}
	}
	if s.StateVec != "" {
		if _, ok := c.curScope.Lookup(s.StateVec); !ok {
			c.errorf(s.Pos(), "undefined state vector: %s", s.StateVec)
		}
	}
}
