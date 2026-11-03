package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

type SandboxInfo struct {
	Name     string
	Boundary string
	Capacity int
}

func (c *Checker) checkSandboxCreateStmt(s *ast.SandboxCreateStmt) {
	if s.Name == "" {
		c.errorf(s.Pos(), "sandbox name cannot be empty")
		return
	}

	// Verify boundary exists
	if _, ok := c.curScope.Lookup(s.Boundary); !ok {
		c.errorf(s.Pos(), "undefined ethical boundary %q for sandbox %q", s.Boundary, s.Name)
	}

	// Verify capacity
	capVal := 10
	if s.Capacity != nil {
		capType := c.checkExpr(s.Capacity)
		if capType != nil && capType.Kind != types.KindInt64 && capType.Kind != types.KindInt32 {
			c.errorf(s.Capacity.Pos(), "sandbox capacity must be an integer, got %s", capType)
		}
		if lit, ok := s.Capacity.(*ast.IntLit); ok {
			if lit.Value <= 0 {
				c.errorf(s.Capacity.Pos(), "sandbox capacity must be positive, got %d", lit.Value)
			}
			capVal = int(lit.Value)
		}
	}

	c.sandboxes[s.Name] = &SandboxInfo{
		Name:     s.Name,
		Boundary: s.Boundary,
		Capacity: capVal,
	}
	c.curScope.Define(&Symbol{Name: s.Name, Type: types.Concept})
}

func (c *Checker) checkSandboxRunStmt(s *ast.SandboxRunStmt) {
	if s.Sandbox == "" {
		c.errorf(s.Pos(), "sandbox name cannot be empty")
		return
	}

	if _, ok := c.sandboxes[s.Sandbox]; !ok {
		if _, inScope := c.curScope.Lookup(s.Sandbox); !inScope {
			c.errorf(s.Pos(), "undefined sandbox %q", s.Sandbox)
		}
	}

	if s.ToolName == "" {
		c.errorf(s.Pos(), "tool name cannot be empty")
		return
	}

	if _, ok := c.tools[s.ToolName]; !ok {
		if _, inScope := c.curScope.Lookup(s.ToolName); !inScope {
			c.errorf(s.Pos(), "tool %q is not registered", s.ToolName)
		}
	}

	for _, arg := range s.Args {
		c.checkExpr(arg)
	}
}

func (c *Checker) checkSandboxVerifyStmt(s *ast.SandboxVerifyStmt) {
	if s.Sandbox == "" {
		c.errorf(s.Pos(), "sandbox name cannot be empty")
		return
	}

	if _, ok := c.sandboxes[s.Sandbox]; !ok {
		if _, inScope := c.curScope.Lookup(s.Sandbox); !inScope {
			c.errorf(s.Pos(), "undefined sandbox %q", s.Sandbox)
		}
	}

	if s.Into != "" {
		c.curScope.Define(&Symbol{Name: s.Into, Type: types.Text})
	}
}

func (c *Checker) checkSandboxVerifyExpr(e *ast.SandboxVerifyExpr) *types.Type {
	if e.Sandbox == "" {
		c.errorf(e.Pos(), "sandbox name cannot be empty")
		return types.Text
	}

	if _, ok := c.sandboxes[e.Sandbox]; !ok {
		if _, inScope := c.curScope.Lookup(e.Sandbox); !inScope {
			c.errorf(e.Pos(), "undefined sandbox %q", e.Sandbox)
		}
	}

	return types.Text
}
