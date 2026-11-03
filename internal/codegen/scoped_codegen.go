package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genRecallTokenDecl(d *ast.RecallTokenDecl) {
	nameStr := g.genStringConst(d.Name)

	bRef := "null"
	for _, slot := range g.vars {
		if slot.llvmType == "i8*" && slot.qtype != nil && slot.qtype.String() == "ethical_boundary" {
			t := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", t, slot.ptr)
			bRef = t
			break
		}
	}

	rRef := "null"
	if d.Recipient != nil {
		if id, ok := d.Recipient.(*ast.Ident); ok {
			rRef = g.genStringConst(id.Name)
		} else {
			rVal := g.genExpr(d.Recipient)
			rRef = rVal.Ref
		}
	}

	expirySec := int64(3600)
	if d.Expiry != nil {
		if lit, ok := d.Expiry.(*ast.IntLit); ok {
			expirySec = lit.Value
		}
	}

	numKeys := len(d.Keys)
	keysArrayPtr := "null"
	if numKeys > 0 {
		arrTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca [%d x i8*]\n", arrTemp, numKeys)
		for i, k := range d.Keys {
			kVal := g.genExpr(k)
			elemPtr := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = getelementptr [%d x i8*], [%d x i8*]* %s, i32 0, i32 %d\n", elemPtr, numKeys, numKeys, arrTemp, i)
			fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", kVal.Ref, elemPtr)
		}
		castTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = bitcast [%d x i8*]* %s to i8**\n", castTemp, numKeys, arrTemp)
		keysArrayPtr = castTemp
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_scoped_issue(i8* %s, i8* %s, i8* %s, i8** %s, i32 %d, i64 %d)\n",
		resTemp, nameStr, bRef, rRef, keysArrayPtr, numKeys, expirySec)

	if d.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", d.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.Text,
		}
		g.vars[d.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genGrantStmt(s *ast.GrantStmt) {
	tokVal := g.genExpr(s.Token)
	rRef := "null"
	if s.Recipient != nil {
		if id, ok := s.Recipient.(*ast.Ident); ok {
			rRef = g.genStringConst(id.Name)
		} else {
			rVal := g.genExpr(s.Recipient)
			rRef = rVal.Ref
		}
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_scoped_grant(i8* %s, i8* %s)\n", resTemp, tokVal.Ref, rRef)
}

func (g *Generator) genRevokeStmt(s *ast.RevokeStmt) {
	tokVal := g.genExpr(s.Token)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_scoped_revoke(i8* %s)\n", resTemp, tokVal.Ref)
}

func (g *Generator) genScopedRecallExpr(e *ast.ScopedRecallExpr) Value {
	mVal := g.genExpr(e.Memory)
	kVal := g.genExpr(e.Key)
	tokVal := g.genExpr(e.Token)

	outValPtr := g.nextTemp()
	outLenPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", outValPtr)
	fmt.Fprintf(&g.sb, "  %s = alloca i64\n", outLenPtr)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_scoped_recall(i8* %s, i8* %s, i8* null, i8* %s, i64 1000, i8** %s, i64* %s)\n",
		resTemp, mVal.Ref, tokVal.Ref, kVal.Ref, outValPtr, outLenPtr)

	loadTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loadTemp, outValPtr)
	return Value{Type: "i8*", Ref: loadTemp, QType: types.Text}
}
