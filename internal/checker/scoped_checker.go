package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

type RecallTokenInfo struct {
	Name        string
	Memory      string
	Recipient   string
	AllowedKeys map[string]bool
	Expiry      int64
}

func (c *Checker) checkRecallTokenDecl(decl *ast.RecallTokenDecl) {
	if decl == nil {
		return
	}
	if c.recallTokens == nil {
		c.recallTokens = make(map[string]*RecallTokenInfo)
	}

	if decl.Name != "" && c.curScope != nil {
		_ = c.curScope.Define(&Symbol{
			Name: decl.Name,
			Type: types.Text,
			Pos:  decl.Pos(),
		})
	}

	info := &RecallTokenInfo{
		Name:        decl.Name,
		AllowedKeys: make(map[string]bool),
	}

	// Validate memory
	if decl.Memory != nil {
		if id, ok := decl.Memory.(*ast.Ident); ok {
			info.Memory = id.Name
		}
		c.checkExpr(decl.Memory)
	}

	// Validate recipient
	if decl.Recipient != nil {
		if id, ok := decl.Recipient.(*ast.Ident); ok {
			info.Recipient = id.Name
			if c.curScope != nil {
				if _, exists := c.curScope.Lookup(id.Name); !exists {
					_ = c.curScope.Define(&Symbol{
						Name: id.Name,
						Type: types.Text,
						Pos:  id.Pos(),
					})
				}
			}
		}
		c.checkExpr(decl.Recipient)
	}

	// Validate allowed keys
	for _, k := range decl.Keys {
		if str, ok := k.(*ast.StringLit); ok {
			info.AllowedKeys[str.Value] = true
		} else if id, ok := k.(*ast.Ident); ok {
			info.AllowedKeys[id.Name] = true
		}
		c.checkExpr(k)
	}

	if decl.Expiry != nil {
		c.checkExpr(decl.Expiry)
	}

	c.recallTokens[decl.Name] = info
}

func (c *Checker) checkGrantStmt(stmt *ast.GrantStmt) {
	if stmt == nil {
		return
	}
	if stmt.Token != nil {
		c.checkExpr(stmt.Token)
		if id, ok := stmt.Token.(*ast.Ident); ok {
			if c.recallTokens != nil {
				if _, exists := c.recallTokens[id.Name]; !exists {
					c.errorf(stmt.Pos(), "grant: unknown recall_token %q", id.Name)
				}
			}
		}
	}
	if stmt.Recipient != nil {
		c.checkExpr(stmt.Recipient)
	}
}

func (c *Checker) checkRevokeStmt(stmt *ast.RevokeStmt) {
	if stmt == nil {
		return
	}
	if stmt.Token != nil {
		c.checkExpr(stmt.Token)
		if id, ok := stmt.Token.(*ast.Ident); ok {
			if c.recallTokens != nil {
				if _, exists := c.recallTokens[id.Name]; !exists {
					c.errorf(stmt.Pos(), "revoke: unknown recall_token %q", id.Name)
				}
			}
		}
	}
}

func (c *Checker) checkScopedRecallExpr(expr *ast.ScopedRecallExpr) *types.Type {
	if expr == nil {
		return types.Text
	}
	if expr.Memory != nil {
		c.checkExpr(expr.Memory)
	}

	var keyStr string
	if expr.Key != nil {
		c.checkExpr(expr.Key)
		if str, ok := expr.Key.(*ast.StringLit); ok {
			keyStr = str.Value
		} else if id, ok := expr.Key.(*ast.Ident); ok {
			keyStr = id.Name
		}
	}

	if expr.Token != nil {
		c.checkExpr(expr.Token)
		if id, ok := expr.Token.(*ast.Ident); ok {
			if c.recallTokens != nil {
				tokInfo, exists := c.recallTokens[id.Name]
				if exists {
					if keyStr != "" && !tokInfo.AllowedKeys[keyStr] {
						c.errorf(expr.Pos(), "key %q not in token's allowed_keys", keyStr)
					}
				}
			}
		}
	}

	resType := types.Text
	c.exprTypes[expr] = resType
	return resType
}
