package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkQECMemoryDecl(d *ast.QECMemoryDecl) {
	if d.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: d.Name,
			Type: types.QECMemoryT(),
			Pos:  d.Pos(),
		})
	}

	if d.CodeType != token.SURFACE_CODE && d.CodeType != token.QLDPC_CODE {
		c.errorf(d.Pos(), "invalid QEC code type: %v (must be surface_code or qldpc_code)", d.CodeType)
	}

	if d.Distance != nil {
		distType := c.checkExpr(d.Distance)
		if lit, ok := d.Distance.(*ast.IntLit); ok {
			if lit.Value < 3 || lit.Value > 11 || lit.Value%2 == 0 {
				c.errorf(d.Distance.Pos(), "distance must be an odd integer between 3 and 11, got %d", lit.Value)
			}
		} else if distType != nil && !distType.IsInteger() {
			c.errorf(d.Distance.Pos(), "distance must be an integer, got %s", distType)
		}
	}

	if d.Boundary != nil {
		bType := c.checkExpr(d.Boundary)
		if bType != nil && bType.Kind != types.KindEthicalBoundary {
			c.errorf(d.Boundary.Pos(), "qec_memory boundary must be an ethical_boundary, got %s", bType)
		}
	}
}

func (c *Checker) checkSyndromeStmt(s *ast.SyndromeStmt) {
	if s.QECMem != nil {
		memType := c.checkExpr(s.QECMem)
		if memType != nil && memType.Kind != types.KindQECMemory {
			c.errorf(s.QECMem.Pos(), "syndrome target must be a qec_memory, got %s", memType)
		}
	}

	if s.Into != "" {
		_ = c.curScope.Define(&Symbol{
			Name: s.Into,
			Type: types.SyndromeT(),
			Pos:  s.Pos(),
		})
		c.verifiedSyndromes[s.Into] = false
	}
}

func (c *Checker) checkVerifySyndromeStmt(s *ast.VerifySyndromeStmt) {
	if s.QECMem != nil {
		memType := c.checkExpr(s.QECMem)
		if memType != nil && memType.Kind != types.KindQECMemory {
			c.errorf(s.QECMem.Pos(), "verify_syndrome target must be a qec_memory, got %s", memType)
		}
	}

	if s.Syndrome != nil {
		synType := c.checkExpr(s.Syndrome)
		if synType != nil && synType.Kind != types.KindSyndrome {
			c.errorf(s.Syndrome.Pos(), "verify_syndrome requires syndrome handle, got %s", synType)
		}
		if ident, ok := s.Syndrome.(*ast.Ident); ok {
			c.verifiedSyndromes[ident.Name] = true
		}
	}
}

func (c *Checker) checkCorrectStmt(s *ast.CorrectStmt) {
	if s.QECMem != nil {
		memType := c.checkExpr(s.QECMem)
		if memType != nil && memType.Kind != types.KindQECMemory {
			c.errorf(s.QECMem.Pos(), "correct target must be a qec_memory, got %s", memType)
		}
	}

	if s.Syndrome != nil {
		synType := c.checkExpr(s.Syndrome)
		if synType != nil && synType.Kind != types.KindSyndrome {
			c.errorf(s.Syndrome.Pos(), "correct requires syndrome handle, got %s", synType)
		}
		if ident, ok := s.Syndrome.(*ast.Ident); ok {
			if !c.verifiedSyndromes[ident.Name] {
				c.errorf(s.Pos(), "correct requires verified syndrome, call verify_syndrome first")
			}
		}
	}
}

func (c *Checker) checkCorruptSyndromeStmt(s *ast.CorruptSyndromeStmt) {
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_syndrome is test-only, use --test flag")
		return
	}

	if s.QECMem != nil {
		memType := c.checkExpr(s.QECMem)
		if memType != nil && memType.Kind != types.KindQECMemory {
			c.errorf(s.QECMem.Pos(), "corrupt_syndrome target must be a qec_memory, got %s", memType)
		}
	}

	if s.Syndrome != nil {
		_ = c.checkExpr(s.Syndrome)
	}

	if s.AtVal != nil {
		atType := c.checkExpr(s.AtVal)
		if atType != nil && !atType.IsInteger() {
			c.errorf(s.AtVal.Pos(), "offset must be an integer, got %s", atType)
		}
	}
}
