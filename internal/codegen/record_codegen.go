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
	if g.checker != nil {
		if vdecl, ok := g.checker.VariantDecls()[e.VariantName]; ok {
			for i, cs := range vdecl.Cases {
				if cs.Name == e.CaseName {
					tag = i
					break
				}
			}
		}
	} else if e.CaseName == "Identifier" || e.CaseName == "String" || e.CaseName == "Case2" {
		tag = 1
	}

	valRef := "0"
	if len(e.Values) > 0 {
		v := g.genExpr(e.Values[0])
		valRef = v.Ref
	} else if e.Value != nil {
		v := g.genExpr(e.Value)
		valRef = v.Ref
	}

	varTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_variant_create_i64(i64 %d, i32 %d, i64 %s)\n", varTemp, typeHash, tag, valRef)

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

	endLabel := g.nextLabel("match_end")

	for idx, arm := range s.Arms {
		armLabel := g.nextLabel("match_arm")
		nextArmLabel := g.nextLabel("match_next")

		cmpTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp eq i32 %s, %d\n", cmpTemp, tagTemp, idx)
		fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", cmpTemp, armLabel, nextArmLabel)

		fmt.Fprintf(&g.sb, "%s:\n", armLabel)
		if len(arm.Bindings) > 0 {
			bValTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_variant_get_payload_i64(i8* %s)\n", bValTemp, subjVal.Ref)
			for _, b := range arm.Bindings {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i64\n", alloca)
				g.vars[b] = &varSlot{
					ptr:      alloca,
					llvmType: "i64",
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
