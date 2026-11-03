package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genAssertStmt(s *ast.AssertStmt) {
	val := g.genExpr(s.Value)
	threshVal := g.genExpr(s.Threshold)

	threshRef := threshVal.Ref
	if threshVal.Type == "i64" {
		conv := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = sitofp i64 %s to double\n", conv, threshVal.Ref)
		threshRef = conv
	}

	valRef := val.Ref
	if val.QType != nil && val.QType.Kind != types.KindConfident {
		// wrap into confidence object if numeric
		scoreTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_confidence_create(double %s, double 1.0, i8 0)\n", scoreTemp, valRef)
		valRef = scoreTemp
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_confidence_assert(i8* %s, double %s)\n", resTemp, valRef, threshRef)
}

func (g *Generator) genInferStmt(s *ast.InferStmt) {
	modelVal := g.genExpr(s.Model)
	inputVal := g.genExpr(s.Input)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_network_forward(i8* %s, i8* %s)\n", resTemp, modelVal.Ref, inputVal.Ref)

	ptr := fmt.Sprintf("%%%s.ptr", s.Into)
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
	slot := &varSlot{
		ptr:      ptr,
		llvmType: "i8*",
		qtype:    types.TensorOf(types.Float64),
	}
	g.vars[s.Into] = slot
	fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
}

func (g *Generator) genPropagateStmt(s *ast.PropagateStmt) {
	_ = g.genExpr(s.Value)
	_ = g.genExpr(s.Source)
}

func (g *Generator) genConfidentLit(e *ast.ConfidentLit) Value {
	val := g.genExpr(e.Value)
	valRef := val.Ref
	if val.Type == "i64" {
		conv := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = sitofp i64 %s to double\n", conv, val.Ref)
		valRef = conv
	}

	state := int8(0)
	switch e.State {
	case types.UnknownState:
		state = 1
	case types.ExploreState:
		state = 2
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_confidence_create(double %s, double %f, i8 %d)\n", resTemp, valRef, e.Score, state)
	return Value{
		Type:  "i8*",
		Ref:   resTemp,
		QType: types.ConfidentOf(types.Float64, e.Score),
	}
}
