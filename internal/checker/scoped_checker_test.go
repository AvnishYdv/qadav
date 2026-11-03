package checker

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func TestScopedChecker_ValidAndDenied(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "mp", Type: types.Text})
	c.globalScope.Define(&Symbol{Name: "agent2", Type: types.Text})
	c.globalScope.Define(&Symbol{Name: "tok", Type: types.Text})

	// 1. Declare recall token
	tokDecl := &ast.RecallTokenDecl{
		Pos_:      pos,
		Name:      "tok",
		Memory:    &ast.Ident{Pos_: pos, Name: "mp"},
		Recipient: &ast.Ident{Pos_: pos, Name: "agent2"},
		Keys: []ast.Expr{
			&ast.StringLit{Pos_: pos, Value: "allowed_key"},
		},
		Expiry: &ast.IntLit{Pos_: pos, Value: 3600},
	}
	c.checkRecallTokenDecl(tokDecl)

	// 2. Grant statement
	grantStmt := &ast.GrantStmt{
		Pos_:      pos,
		Token:     &ast.Ident{Pos_: pos, Name: "tok"},
		Recipient: &ast.Ident{Pos_: pos, Name: "agent2"},
	}
	c.checkGrantStmt(grantStmt)

	// 3. Allowed key recall
	validExpr := &ast.ScopedRecallExpr{
		Pos_:   pos,
		Memory: &ast.Ident{Pos_: pos, Name: "mp"},
		Key:    &ast.StringLit{Pos_: pos, Value: "allowed_key"},
		Token:  &ast.Ident{Pos_: pos, Name: "tok"},
	}
	ty := c.checkScopedRecallExpr(validExpr)
	if ty != types.Text {
		t.Fatalf("expected text, got %v", ty)
	}
	if len(c.Errors()) != 0 {
		t.Fatalf("expected 0 errors, got %v", c.Errors())
	}

	// 4. Denied key recall (must error!)
	deniedExpr := &ast.ScopedRecallExpr{
		Pos_:   pos,
		Memory: &ast.Ident{Pos_: pos, Name: "mp"},
		Key:    &ast.StringLit{Pos_: pos, Value: "denied_key"},
		Token:  &ast.Ident{Pos_: pos, Name: "tok"},
	}
	c.checkScopedRecallExpr(deniedExpr)
	if len(c.Errors()) == 0 {
		t.Fatalf("expected error for denied_key, got none")
	}

	// 5. Revoke statement
	revokeStmt := &ast.RevokeStmt{
		Pos_:  pos,
		Token: &ast.Ident{Pos_: pos, Name: "tok"},
	}
	c.checkRevokeStmt(revokeStmt)

	// 6. Unknown token error paths
	badGrant := &ast.GrantStmt{
		Pos_:  pos,
		Token: &ast.Ident{Pos_: pos, Name: "unknown_tok"},
	}
	c.checkGrantStmt(badGrant)

	badRevoke := &ast.RevokeStmt{
		Pos_:  pos,
		Token: &ast.Ident{Pos_: pos, Name: "unknown_tok"},
	}
	c.checkRevokeStmt(badRevoke)

	// Nil safety checks
	c.checkRecallTokenDecl(nil)
	c.checkGrantStmt(nil)
	c.checkRevokeStmt(nil)
	c.checkScopedRecallExpr(nil)
}
