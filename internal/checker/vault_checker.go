package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkVaultDecl(v *ast.VaultDecl) {
	if v.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: v.Name,
			Type: types.Vault,
			Pos:  v.Pos(),
		})
	}

	if v.Boundary != nil {
		bType := c.checkExpr(v.Boundary)
		if bType != nil && bType.Kind != types.KindEthicalBoundary {
			c.errorf(v.Boundary.Pos(), "vault boundary must be an ethical_boundary, got %s", bType)
		}
	}

	if v.NetworkRef != nil {
		isNet := false
		if id, ok := v.NetworkRef.(*ast.Ident); ok {
			if _, exists := c.networks[id.Name]; exists {
				isNet = true
			}
		}
		netType := c.checkExpr(v.NetworkRef)
		if !isNet && netType != nil && netType.Kind != types.KindNetwork {
			c.errorf(v.NetworkRef.Pos(), "vault network target must be a network, got %s", netType)
		}
	}

	if v.Threshold != nil {
		tType := c.checkExpr(v.Threshold)
		if lit, ok := v.Threshold.(*ast.IntLit); ok {
			if lit.Value < 2 || lit.Value > 5 {
				c.errorf(v.Threshold.Pos(), "vault threshold must be between 2 and 5, got %d", lit.Value)
			}
		} else if tType != nil && tType.Kind != types.KindInt64 && tType.Kind != types.KindInt32 {
			c.errorf(v.Threshold.Pos(), "vault threshold must be an integer, got %s", tType)
		}
	}
}

func (c *Checker) checkSealStmt(s *ast.SealStmt) {
	vType := c.checkExpr(s.Vault)
	if vType != nil && vType.Kind != types.KindVault {
		c.errorf(s.Vault.Pos(), "seal target must be a vault, got %s", vType)
	}
	if s.Network != nil {
		_ = c.checkExpr(s.Network)
	}
}

func (c *Checker) checkUnsealStmt(s *ast.UnsealStmt) {
	vType := c.checkExpr(s.Vault)
	if vType != nil && vType.Kind != types.KindVault {
		c.errorf(s.Vault.Pos(), "unseal target must be a vault, got %s", vType)
	}
	if s.Boundary != nil {
		bType := c.checkExpr(s.Boundary)
		if bType != nil && bType.Kind != types.KindEthicalBoundary {
			c.errorf(s.Boundary.Pos(), "unseal boundary must be an ethical_boundary, got %s", bType)
		}
	}
}

func (c *Checker) checkWipeStmt(s *ast.WipeStmt) {
	vType := c.checkExpr(s.Vault)
	if vType != nil && vType.Kind != types.KindVault {
		c.errorf(s.Vault.Pos(), "wipe target must be a vault, got %s", vType)
	}
}
