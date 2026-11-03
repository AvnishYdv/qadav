package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkTranspileToStmt(s *ast.TranspileToStmt) {
	if s == nil {
		return
	}
	switch s.Target {
	case "qiskit", "cirq", "ionq", "qasm3":
		// Valid target
	default:
		c.errorf(s.Pos(), "invalid transpilation target: %q (expected qiskit, cirq, ionq, or qasm3)", s.Target)
	}

	if s.BufLen != nil {
		c.checkExpr(s.BufLen)
		if intLit, ok := s.BufLen.(*ast.IntLit); ok && intLit.Value < 256 {
			c.errorf(s.Pos(), "transpile buffer size %d is below minimum required 256 bytes", intLit.Value)
		}
	}

	if s.Into != "" {
		if sym, ok := c.curScope.Lookup(s.Into); ok {
			if sym.Type != types.Text {
				c.errorf(s.Pos(), "variable %s must have type text for transpilation output, got %s", s.Into, sym.Type)
			}
		} else {
			c.curScope.Define(&Symbol{
				Name: s.Into,
				Type: types.Text,
			})
		}
	}
}
