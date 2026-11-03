package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkProvenanceDecl(p *ast.ProvenanceDecl) {
	if p.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: p.Name,
			Type: types.Provenance,
			Pos:  p.Pos(),
		})
	}

	if p.NetworkRef != nil {
		isNet := false
		if id, ok := p.NetworkRef.(*ast.Ident); ok {
			if _, exists := c.networks[id.Name]; exists {
				isNet = true
			}
		}
		netType := c.checkExpr(p.NetworkRef)
		if !isNet && netType != nil && netType.Kind != types.KindNetwork {
			c.errorf(p.NetworkRef.Pos(), "provenance network target must be a network, got %s", netType)
		}
	}
}

func (c *Checker) checkCertifyStmt(s *ast.CertifyStmt) {
	pType := c.checkExpr(s.Provenance)
	if pType != nil && pType.Kind != types.KindProvenance {
		c.errorf(s.Provenance.Pos(), "certify target must be a provenance tracker, got %s", pType)
	}
}

func (c *Checker) checkVerifyTrainingExpr(e *ast.VerifyTrainingExpr) *types.Type {
	if e.Certificate != nil {
		_ = c.checkExpr(e.Certificate)
	}
	if e.Model != nil {
		_ = c.checkExpr(e.Model)
	}
	return types.Logic
}
