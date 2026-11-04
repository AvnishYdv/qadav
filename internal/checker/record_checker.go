package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkRecordDecl(s *ast.RecordDeclStmt) {
	if s == nil {
		return
	}
	if existing, exists := c.recordDecls[s.Name]; exists {
		if existing != s {
			c.errorf(s.Pos(), "duplicate record declaration %q", s.Name)
		}
		return
	}

	fieldNames := make(map[string]bool)
	for _, f := range s.Fields {
		if fieldNames[f.Name] {
			c.errorf(f.Pos(), "duplicate field %q in record %q", f.Name, s.Name)
		}
		fieldNames[f.Name] = true
		if c.resolveType(&ast.NamedType{Name: f.Type}) == nil {
			c.errorf(f.Pos(), "unknown type %q for field %q in record %q", f.Type, f.Name, s.Name)
		}
	}

	c.recordDecls[s.Name] = s
	c.curScope.Define(&Symbol{
		Name: s.Name,
		Type: &types.Type{
			Kind: types.KindRecord,
			Name: s.Name,
		},
	})
}

func (c *Checker) checkVariantDecl(s *ast.VariantDeclStmt) {
	if s == nil {
		return
	}
	if existing, exists := c.variantDecls[s.Name]; exists {
		if existing != s {
			c.errorf(s.Pos(), "duplicate variant declaration %q", s.Name)
		}
		return
	}

	c.variantDecls[s.Name] = s

	caseNames := make(map[string]bool)
	for _, cs := range s.Cases {
		if caseNames[cs.Name] {
			c.errorf(cs.Pos(), "duplicate case %q in variant %q", cs.Name, s.Name)
		}
		caseNames[cs.Name] = true
		for _, f := range cs.Fields {
			if c.resolveType(&ast.NamedType{Name: f.Type}) == nil {
				c.errorf(f.Pos(), "unknown type %q for field %q in variant %q case %q", f.Type, f.Name, s.Name, cs.Name)
			}
		}
	}

	c.curScope.Define(&Symbol{
		Name: s.Name,
		Type: &types.Type{
			Kind: types.KindVariant,
			Name: s.Name,
		},
	})
}

func (c *Checker) checkMatchStmt(s *ast.MatchStmt) {
	if s == nil {
		return
	}

	targetType := c.checkExpr(s.Subject)
	if targetType == nil || targetType.Kind != types.KindVariant {
		c.errorf(s.Pos(), "match subject must be a variant, got %s", targetType)
		return
	}
	vdecl := c.variantDecls[targetType.Name]

	coveredCases := make(map[string]bool)
	hasOtherwise := (s.Otherwise != nil)

	for _, arm := range s.Arms {
		if arm.CaseName == "otherwise" || arm.CaseName == "_" {
			hasOtherwise = true
		} else if vdecl != nil {
			var foundCase *ast.VariantCase
			for _, cs := range vdecl.Cases {
				if cs.Name == arm.CaseName {
					foundCase = cs
					break
				}
			}
			if foundCase == nil {
				c.errorf(arm.Pos(), "pattern %q is not a valid case of variant %q", arm.CaseName, vdecl.Name)
			} else {
				if coveredCases[arm.CaseName] {
					c.errorf(arm.Pos(), "duplicate match arm for case %q", arm.CaseName)
				}
				coveredCases[arm.CaseName] = true
				if len(arm.Bindings) != len(foundCase.Fields) {
					c.errorf(arm.Pos(), "case %q expects %d bindings, got %d", arm.CaseName, len(foundCase.Fields), len(arm.Bindings))
				}
			}
		}

		c.enterScope()
		for _, b := range arm.Bindings {
			c.curScope.Define(&Symbol{
				Name: b,
				Type: types.Int64,
			})
		}
		if arm.Body != nil {
			c.checkStmt(arm.Body)
		}
		c.leaveScope()
	}

	if vdecl != nil && !hasOtherwise {
		for _, cs := range vdecl.Cases {
			if !coveredCases[cs.Name] {
				c.errorf(s.Pos(), "non-exhaustive match on %q: missing case %q", vdecl.Name, cs.Name)
			}
		}
	}

	if s.Otherwise != nil {
		c.enterScope()
		c.checkStmt(s.Otherwise)
		c.leaveScope()
	}
}

func (c *Checker) checkRecordInitExpr(e *ast.RecordInitExpr) *types.Type {
	if e == nil {
		return types.Void
	}

	rdecl, ok := c.recordDecls[e.RecordName]
	if !ok {
		c.errorf(e.Pos(), "unknown record type %q", e.RecordName)
		return types.Void
	}

	for _, initF := range e.Fields {
		var foundField *ast.FieldDecl
		for _, rf := range rdecl.Fields {
			if rf.Name == initF.Field {
				foundField = rf
				break
			}
		}
		if foundField == nil {
			c.errorf(initF.Pos(), "field %q does not exist on record %q", initF.Field, e.RecordName)
		}
		c.checkExpr(initF.Value)
	}

	return &types.Type{
		Kind: types.KindRecord,
		Name: e.RecordName,
	}
}

func (c *Checker) checkVariantInitExpr(e *ast.VariantInitExpr) *types.Type {
	if e == nil {
		return types.Void
	}

	vdecl, ok := c.variantDecls[e.VariantName]
	if !ok {
		c.errorf(e.Pos(), "unknown variant type %q", e.VariantName)
		return types.Void
	}

	var foundCase *ast.VariantCase
	for _, cs := range vdecl.Cases {
		if cs.Name == e.CaseName {
			foundCase = cs
			break
		}
	}
	if foundCase == nil {
		c.errorf(e.Pos(), "unknown case %q on variant %q", e.CaseName, e.VariantName)
		return types.Void
	}

	args := e.Values
	if len(args) == 0 && e.Value != nil {
		args = []ast.Expr{e.Value}
	}

	if len(args) != len(foundCase.Fields) {
		c.errorf(e.Pos(), "case %q expects %d arguments, got %d", e.CaseName, len(foundCase.Fields), len(args))
	}

	for _, arg := range args {
		c.checkExpr(arg)
	}

	return &types.Type{
		Kind: types.KindVariant,
		Name: e.VariantName,
	}
}

func (c *Checker) checkFieldAccessExpr(e *ast.FieldAccessExpr) *types.Type {
	if e == nil {
		return types.Void
	}

	recvType := c.checkExpr(e.Object)
	if recvType == nil {
		return types.Void
	}

	if recvType.Kind != types.KindRecord {
		c.errorf(e.Pos(), "cannot access field %q on non-record type %s", e.Field, recvType)
		return types.Void
	}

	rdecl := c.recordDecls[recvType.Name]
	if rdecl != nil {
		for _, f := range rdecl.Fields {
			if f.Name == e.Field {
				switch f.Type {
				case "text":
					return types.Text
				case "logic":
					return types.Logic
				case "float64":
					return types.Float64
				default:
					return types.Int64
				}
			}
		}
		c.errorf(e.Pos(), "field %q not found on record %q", e.Field, recvType.Name)
	}

	return types.Int64
}
