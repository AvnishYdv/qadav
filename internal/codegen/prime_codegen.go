package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

// genConceptLit compiles a concept literal → i64 constant.
func (g *Generator) genConceptLit(e *ast.ConceptLit) Value {
	return Value{
		Type:  "i64",
		Ref:   fmt.Sprintf("%d", e.Value),
		QType: types.Concept,
	}
}

// genDecomposeExpr compiles decompose⦅concept⦆ → tensor⟨int64⟩ of prime factors.
// Calls @qadav_factorize, builds a tensor from the results.
func (g *Generator) genDecomposeExpr(e *ast.DecomposeExpr) Value {
	inputVal := g.genExpr(e.Input)
	nVal := inputVal.Ref
	if inputVal.Type != "i64" {
		ext := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", ext, inputVal.Type, nVal)
		nVal = ext
	}

	// Allocate a stack buffer for up to 128 factors
	bufPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [128 x i64]\n", bufPtr)
	rawPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [128 x i64], [128 x i64]* %s, i32 0, i32 0\n", rawPtr, bufPtr)

	// Call factorize
	countTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_factorize(i64 %s, i64* %s, i32 128)\n", countTemp, nVal, rawPtr)

	// Build a tensor from the factors with shape [countTemp]
	shapePtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [1 x i32]\n", shapePtr)
	shapeElemPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [1 x i32], [1 x i32]* %s, i32 0, i32 0\n", shapeElemPtr, shapePtr)
	fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", countTemp, shapeElemPtr)
	tensorPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_tensor_create(i32 4, i32 1, i32* %s)\n", tensorPtr, shapeElemPtr)

	// Fill tensor with factors
	loopIdx := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca i32\n", loopIdx)
	fmt.Fprintf(&g.sb, "  store i32 0, i32* %s\n", loopIdx)

	loopStart := g.nextLabel("decomp_loop")
	loopBody := g.nextLabel("decomp_body")
	loopEnd := g.nextLabel("decomp_end")

	fmt.Fprintf(&g.sb, "  br label %%%s\n", loopStart)
	fmt.Fprintf(&g.sb, "%s:\n", loopStart)
	g.isTerminated = false
	curIdx := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i32, i32* %s\n", curIdx, loopIdx)
	cond := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = icmp slt i32 %s, %s\n", cond, curIdx, countTemp)
	fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", cond, loopBody, loopEnd)

	fmt.Fprintf(&g.sb, "%s:\n", loopBody)
	g.isTerminated = false
	factorPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds i64, i64* %s, i32 %s\n", factorPtr, rawPtr, curIdx)
	factorVal := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load i64, i64* %s\n", factorVal, factorPtr)
	fmt.Fprintf(&g.sb, "  call void @qadav_tensor_set_i64(i8* %s, i32 %s, i64 %s)\n", tensorPtr, curIdx, factorVal)
	nextIdx := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = add i32 %s, 1\n", nextIdx, curIdx)
	fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", nextIdx, loopIdx)
	fmt.Fprintf(&g.sb, "  br label %%%s\n", loopStart)

	fmt.Fprintf(&g.sb, "%s:\n", loopEnd)
	g.isTerminated = false

	return Value{
		Type:  "i8*",
		Ref:   tensorPtr,
		QType: types.TensorOf(types.Int64),
	}
}

// genEmitConcept emits a concept value using qadav_print_concept.
func (g *Generator) genEmitConcept(val Value) {
	nRef := val.Ref
	if val.Type != "i64" {
		ext := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = sext %s %s to i64\n", ext, val.Type, nRef)
		nRef = ext
	}
	fmt.Fprintf(&g.sb, "  call void @qadav_print_concept(i64 %s)\n", nRef)
}
