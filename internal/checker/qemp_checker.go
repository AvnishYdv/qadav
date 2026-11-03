package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkMemoryProvenanceDecl(m *ast.MemoryProvenanceDecl) {
	if m.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: m.Name,
			Type: types.MemoryProv,
			Pos:  m.Pos(),
		})
	}

	if m.Boundary != nil {
		bType := c.checkExpr(m.Boundary)
		if bType != nil && bType.Kind != types.KindEthicalBoundary {
			c.errorf(m.Boundary.Pos(), "memory_prov boundary must be an ethical_boundary, got %s", bType)
		}
	}

	if m.Capacity != nil {
		capType := c.checkExpr(m.Capacity)
		if lit, ok := m.Capacity.(*ast.IntLit); ok {
			if lit.Value < 1 || lit.Value > 1024 {
				c.errorf(m.Capacity.Pos(), "memory_prov capacity must be between 1 and 1024, got %d", lit.Value)
			}
		} else if capType != nil && !capType.IsInteger() {
			c.errorf(m.Capacity.Pos(), "memory_prov capacity must be an integer, got %s", capType)
		}
	}
}

func (c *Checker) checkVerifyMemoryStmt(s *ast.VerifyMemoryStmt) {
	mType := c.checkExpr(s.Mem)
	if mType != nil && mType.Kind != types.KindMemoryProv {
		c.errorf(s.Mem.Pos(), "verify_memory target must be memory_prov, got %s", mType)
	}
}

func (c *Checker) checkAuditMemoryStmt(s *ast.AuditMemoryStmt) {
	mType := c.checkExpr(s.Mem)
	if mType != nil && mType.Kind != types.KindMemoryProv {
		c.errorf(s.Mem.Pos(), "audit_memory target must be memory_prov, got %s", mType)
	}
}

func (c *Checker) checkCorruptShareStmt(s *ast.CorruptShareStmt) {
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_memory_share is test-only, use --test flag")
		return
	}
	mType := c.checkExpr(s.Mem)
	if mType != nil && mType.Kind != types.KindMemoryProv {
		c.errorf(s.Mem.Pos(), "corrupt_memory_share target must be memory_prov, got %s", mType)
	}
	_ = c.checkExpr(s.Key)
	if s.ShareIndex != nil {
		idxType := c.checkExpr(s.ShareIndex)
		if lit, ok := s.ShareIndex.(*ast.IntLit); ok {
			if lit.Value < 0 || lit.Value > 4 {
				c.errorf(s.ShareIndex.Pos(), "share index must be between 0 and 4, got %d", lit.Value)
			}
		} else if idxType != nil && !idxType.IsInteger() {
			c.errorf(s.ShareIndex.Pos(), "share index must be integer, got %s", idxType)
		}
	}
}
