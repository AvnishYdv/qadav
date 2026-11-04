package codegen

import (
	"fmt"
	"hash/fnv"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func hashName64(name string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(name))
	return h.Sum64()
}

func (g *Generator) genRecordInitExpr(e *ast.RecordInitExpr) Value {
	if e == nil {
		return Value{Type: "i8*", Ref: "null"}
	}

	typeHash := hashName64(e.RecordName)
	nFields := len(e.Fields)

	recTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_record_alloc(i64 %d, i32 %d, i32* null)\n", recTemp, typeHash, nFields)

	for idx, f := range e.Fields {
		val := g.genExpr(f.Value)
		if val.Type == "i8*" {
			fmt.Fprintf(&g.sb, "  call void @qadav_record_set_field_ptr(i8* %s, i32 %d, i8* %s)\n", recTemp, idx, val.Ref)
		} else {
			i64Ref := val.Ref
			if val.Type != "i64" {
				conv := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = zext %s %s to i64\n", conv, val.Type, val.Ref)
				i64Ref = conv
			}
			fmt.Fprintf(&g.sb, "  call void @qadav_record_set_field_i64(i8* %s, i32 %d, i64 %s)\n", recTemp, idx, i64Ref)
		}
	}

	return Value{
		Type:  "i8*",
		Ref:   recTemp,
		QType: &types.Type{Kind: types.KindRecord, Name: e.RecordName},
	}
}

func (g *Generator) genFieldAccessExpr(e *ast.FieldAccessExpr) Value {
	if e == nil {
		return Value{Type: "i64", Ref: "0"}
	}

	objVal := g.genExpr(e.Object)

	idx := 0
	isText := false
	found := false

	if objVal.QType != nil && g.checker != nil {
		if rdecl, ok := g.checker.RecordDecls()[objVal.QType.Name]; ok {
			for i, f := range rdecl.Fields {
				if f.Name == e.Field {
					idx = i
					found = true
					if f.Type == "text" {
						isText = true
					}
					break
				}
			}
		}
	}

	if !found {
		if e.Field == "value" || e.Field == "val" {
			idx = 1
		} else if e.Field == "line" {
			idx = 2
		} else if e.Field == "col" {
			idx = 3
		}

		if e.Field == "kind" || e.Field == "value" || e.Field == "name" || e.Field == "val_str" {
			isText = true
		}
	}

	if isText {
		resTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_record_get_field_ptr(i8* %s, i32 %d)\n", resTemp, objVal.Ref, idx)
		return Value{
			Type:  "i8*",
			Ref:   resTemp,
			QType: types.Text,
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_record_get_field_i64(i8* %s, i32 %d)\n", resTemp, objVal.Ref, idx)
	return Value{
		Type:  "i64",
		Ref:   resTemp,
		QType: types.Int64,
	}
}

func (g *Generator) genVariantInitExpr(e *ast.VariantInitExpr) Value {
	if e == nil {
		return Value{Type: "i8*", Ref: "null"}
	}

	typeHash := hashName64(e.VariantName)
	tag := 0
	var foundCase *ast.VariantCase
	if g.checker != nil && g.checker.VariantDecls() != nil {
		if vdecl, ok := g.checker.VariantDecls()[e.VariantName]; ok {
			for i, cs := range vdecl.Cases {
				if cs.Name == e.CaseName {
					tag = i
					foundCase = cs
					break
				}
			}
		}
	} else if e.CaseName == "Identifier" || e.CaseName == "String" || e.CaseName == "Case2" {
		tag = 1
	}

	args := e.Values
	if len(args) == 0 && e.Value != nil {
		args = []ast.Expr{e.Value}
	}

	if len(args) == 0 {
		varTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create(i64 %d, i32 %d, i32 0, i8* null)\n", varTemp, typeHash, tag)
		return Value{
			Type:  "i8*",
			Ref:   varTemp,
			QType: &types.Type{Kind: types.KindVariant, Name: e.VariantName},
		}
	}

	if len(args) == 1 {
		v := g.genExpr(args[0])
		isPtr := (v.Type == "i8*" || v.Type == "ptr")
		if !isPtr && foundCase != nil && len(foundCase.Fields) > 0 {
			fType := foundCase.Fields[0].Type
			if fType != "int64" && fType != "logic" && fType != "float64" {
				isPtr = true
			}
		}

		varTemp := g.nextTemp()
		if isPtr {
			fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create_ptr(i64 %d, i32 %d, i8* %s)\n", varTemp, typeHash, tag, v.Ref)
		} else {
			valRef := v.Ref
			if v.Type != "i64" {
				extTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = zext %s %s to i64\n", extTemp, v.Type, v.Ref)
				valRef = extTemp
			}
			fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create_i64(i64 %d, i32 %d, i64 %s)\n", varTemp, typeHash, tag, valRef)
		}
		return Value{
			Type:  "i8*",
			Ref:   varTemp,
			QType: &types.Type{Kind: types.KindVariant, Name: e.VariantName},
		}
	}

	if len(args) == 2 {
		v0 := g.genExpr(args[0])
		v1 := g.genExpr(args[1])
		v0Ref := v0.Ref
		if v0.Type == "i8*" || v0.Type == "ptr" {
			conv0 := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = ptrtoint %s %s to i64\n", conv0, v0.Type, v0.Ref)
			v0Ref = conv0
		} else if v0.Type != "i64" {
			ext0 := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = zext %s %s to i64\n", ext0, v0.Type, v0.Ref)
			v0Ref = ext0
		}
		v1Ref := v1.Ref
		if v1.Type == "i8*" || v1.Type == "ptr" {
			conv1 := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = ptrtoint %s %s to i64\n", conv1, v1.Type, v1.Ref)
			v1Ref = conv1
		} else if v1.Type != "i64" {
			ext1 := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = zext %s %s to i64\n", ext1, v1.Type, v1.Ref)
			v1Ref = ext1
		}
		varTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create_2(i64 %d, i32 %d, i64 %s, i64 %s)\n", varTemp, typeHash, tag, v0Ref, v1Ref)
		return Value{
			Type:  "i8*",
			Ref:   varTemp,
			QType: &types.Type{Kind: types.KindVariant, Name: e.VariantName},
		}
	}

	bufTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [%d x i64]\n", bufTemp, len(args))
	for i, arg := range args {
		val := g.genExpr(arg)
		vRef := val.Ref
		if val.Type == "i8*" || val.Type == "ptr" {
			conv := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = ptrtoint %s %s to i64\n", conv, val.Type, val.Ref)
			vRef = conv
		} else if val.Type != "i64" {
			ext := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = zext %s %s to i64\n", ext, val.Type, val.Ref)
			vRef = ext
		}
		gepTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i64], [%d x i64]* %s, i64 0, i64 %d\n", gepTemp, len(args), len(args), bufTemp, i)
		fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", vRef, gepTemp)
	}
	rawBuf := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast [%d x i64]* %s to i8*\n", rawBuf, len(args), bufTemp)
	varTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create(i64 %d, i32 %d, i32 %d, i8* %s)\n", varTemp, typeHash, tag, len(args)*8, rawBuf)
	return Value{
		Type:  "i8*",
		Ref:   varTemp,
		QType: &types.Type{Kind: types.KindVariant, Name: e.VariantName},
	}
}

func (g *Generator) genMatchStmt(s *ast.MatchStmt) {
	if s == nil {
		return
	}

	subjVal := g.genExpr(s.Subject)
	tagTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_variant_get_tag(i8* %s)\n", tagTemp, subjVal.Ref)

	// Resolve variant declaration for type resolution of bindings
	var vdecl *ast.VariantDeclStmt
	var variantName string
	if subjVal.QType != nil && subjVal.QType.Kind == types.KindVariant {
		variantName = subjVal.QType.Name
	} else if t, ok := g.exprTypes[s.Subject]; ok && t != nil && t.Kind == types.KindVariant {
		variantName = t.Name
	} else if ident, ok := s.Subject.(*ast.Ident); ok {
		if slot, ok := g.vars[ident.Name]; ok && slot.qtype != nil && slot.qtype.Kind == types.KindVariant {
			variantName = slot.qtype.Name
		}
	}
	if variantName != "" && g.checker != nil && g.checker.VariantDecls() != nil {
		vdecl = g.checker.VariantDecls()[variantName]
	}
	if vdecl == nil && g.checker != nil && g.checker.VariantDecls() != nil && len(s.Arms) > 0 {
		for _, vd := range g.checker.VariantDecls() {
			for _, cs := range vd.Cases {
				if cs.Name == s.Arms[0].CaseName {
					vdecl = vd
					variantName = vd.Name
					break
				}
			}
			if vdecl != nil {
				break
			}
		}
	}

	endLabel := g.nextLabel("match_end")

	for idx, arm := range s.Arms {
		armLabel := g.nextLabel("match_arm")
		nextArmLabel := g.nextLabel("match_next")

		caseTag := idx
		var foundCase *ast.VariantCase
		if vdecl != nil {
			for cIdx, cs := range vdecl.Cases {
				if cs.Name == arm.CaseName {
					caseTag = cIdx
					foundCase = cs
					break
				}
			}
		}

		cmpTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp eq i32 %s, %d\n", cmpTemp, tagTemp, caseTag)
		fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", cmpTemp, armLabel, nextArmLabel)

		fmt.Fprintf(&g.sb, "%s:\n", armLabel)
		for bIdx, b := range arm.Bindings {
			isPtr := false
			var bQType *types.Type
			if foundCase != nil && bIdx < len(foundCase.Fields) {
				fType := foundCase.Fields[bIdx].Type
				if fType != "int64" && fType != "logic" && fType != "float64" {
					isPtr = true
				}
				if g.checker != nil {
					bQType = g.resolveType(&ast.NamedType{Name: fType})
				}
			} else if arm.CaseName == "String" || arm.CaseName == "Identifier" {
				isPtr = true
			}

			if isPtr {
				bValTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_get_field_ptr(i8* %s, i32 %d)\n", bValTemp, subjVal.Ref, bIdx)
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
				g.vars[b] = &varSlot{
					ptr:      alloca,
					llvmType: "i8*",
					qtype:    bQType,
				}
				fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", bValTemp, alloca)
			} else {
				bValTemp := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_variant_get_field_i64(i8* %s, i32 %d)\n", bValTemp, subjVal.Ref, bIdx)
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i64\n", alloca)
				g.vars[b] = &varSlot{
					ptr:      alloca,
					llvmType: "i64",
					qtype:    types.Int64,
				}
				fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", bValTemp, alloca)
			}
		}

		if arm.Body != nil {
			for _, stmt := range arm.Body.Statements {
				g.genStmt(stmt)
			}
		}
		if !g.isTerminated {
			fmt.Fprintf(&g.sb, "  br label %%%s\n", endLabel)
		}
		g.isTerminated = false

		fmt.Fprintf(&g.sb, "%s:\n", nextArmLabel)
	}

	if s.Otherwise != nil {
		for _, stmt := range s.Otherwise.Statements {
			g.genStmt(stmt)
		}
		if !g.isTerminated {
			fmt.Fprintf(&g.sb, "  br label %%%s\n", endLabel)
		}
		g.isTerminated = false
	} else {
		fmt.Fprintf(&g.sb, "  br label %%%s\n", endLabel)
	}

	fmt.Fprintf(&g.sb, "%s:\n", endLabel)
}
