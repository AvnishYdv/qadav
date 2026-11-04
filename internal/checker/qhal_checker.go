package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkQHALBackendAnnotation(s *ast.QHALBackendAnnotation) {
	if s == nil {
		return
	}
	switch s.Kind {
	case "sim", "ibmq", "ionq", "rigetti":
		// Valid
	default:
		c.errorf(s.Pos(), "invalid QHAL backend: %q (expected sim, ibmq, ionq, or rigetti)", s.Kind)
	}
}

func (c *Checker) checkQASMExportStmt(s *ast.QASMExportStmt) {
	if s == nil {
		return
	}
	if s.BufLen != nil {
		c.checkExpr(s.BufLen)
	}
	if s.Into != "" {
		c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.Text,
		})
	}
}
