package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genGPUMatmulStmt(s *ast.GPUMatmulStmt) {
	leftVal := g.genExpr(s.Left)
	rightVal := g.genExpr(s.Right)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_matmul(i8* %s, i8* %s)\n", resTemp, leftVal.Ref, rightVal.Ref)

	if s.Target != nil {
		if ident, ok := s.Target.(*ast.Ident); ok {
			slot, ok := g.vars[ident.Name]
			if !ok {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
				slot = &varSlot{
					ptr:      alloca,
					llvmType: "i8*",
				}
				g.vars[ident.Name] = slot
			}
			fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, slot.ptr)
		}
	}
}

func (g *Generator) genGPUSoftmaxStmt(s *ast.GPUSoftmaxStmt) {
	srcVal := g.genExpr(s.Source)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_softmax(i8* %s)\n", resTemp, srcVal.Ref)

	if s.Target != nil {
		if ident, ok := s.Target.(*ast.Ident); ok {
			slot, ok := g.vars[ident.Name]
			if !ok {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
				slot = &varSlot{
					ptr:      alloca,
					llvmType: "i8*",
				}
				g.vars[ident.Name] = slot
			}
			fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, slot.ptr)
		}
	}
}

func (g *Generator) genGPULayernormStmt(s *ast.GPULayernormStmt) {
	srcVal := g.genExpr(s.Source)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_layernorm(i8* %s, double 0.00001)\n", resTemp, srcVal.Ref)

	if s.Target != nil {
		if ident, ok := s.Target.(*ast.Ident); ok {
			slot, ok := g.vars[ident.Name]
			if !ok {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
				slot = &varSlot{
					ptr:      alloca,
					llvmType: "i8*",
				}
				g.vars[ident.Name] = slot
			}
			fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, slot.ptr)
		}
	}
}

func (g *Generator) genAGKEAttestStmt(s *ast.AGKEAttestStmt) {
	var cRef = "null"
	if s.Chain != nil {
		cVal := g.genExpr(s.Chain)
		cRef = cVal.Ref
	}

	fmt.Fprintf(&g.sb, "  call i32 @qadav_agke_attest(i8* %s, i8* null)\n", cRef)
}

func (g *Generator) genAGKEVerifyStmt(s *ast.AGKEVerifyStmt) {
	var cRef = "null"
	if s.Chain != nil {
		cVal := g.genExpr(s.Chain)
		cRef = cVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_agke_verify(i8* %s, i8* null)\n", resTemp, cRef)

	if s.Target != nil {
		if ident, ok := s.Target.(*ast.Ident); ok {
			slot, ok := g.vars[ident.Name]
			if !ok {
				alloca := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = alloca i1\n", alloca)
				slot = &varSlot{
					ptr:      alloca,
					llvmType: "i1",
				}
				g.vars[ident.Name] = slot
			}
			boolTemp := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = icmp eq i32 %s, 1\n", boolTemp, resTemp)
			fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", boolTemp, slot.ptr)
		}
	}
}

func (g *Generator) genCorruptGPUDigestStmt(s *ast.CorruptGPUDigestStmt) {
	recIdx := 0
	if s.Offset != nil {
		if lit, ok := s.Offset.(*ast.IntLit); ok {
			recIdx = int(lit.Value)
		}
	}
	fmt.Fprintf(&g.sb, "  call i32 @qadav_agke_corrupt_record(i32 %d, i32 1)\n", recIdx)
}

func (g *Generator) genGPUMatmulExpr(e *ast.GPUMatmulExpr) Value {
	leftVal := g.genExpr(e.Left)
	rightVal := g.genExpr(e.Right)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_matmul(i8* %s, i8* %s)\n", resTemp, leftVal.Ref, rightVal.Ref)
	return Value{
		Type: "i8*",
		Ref:  resTemp,
	}
}

func (g *Generator) genGPUSoftmaxExpr(e *ast.GPUSoftmaxExpr) Value {
	srcVal := g.genExpr(e.Source)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_softmax(i8* %s)\n", resTemp, srcVal.Ref)
	return Value{
		Type: "i8*",
		Ref:  resTemp,
	}
}

func (g *Generator) genGPULayernormExpr(e *ast.GPULayernormExpr) Value {
	srcVal := g.genExpr(e.Source)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_gpu_tensor_layernorm(i8* %s, double 0.00001)\n", resTemp, srcVal.Ref)
	return Value{
		Type: "i8*",
		Ref:  resTemp,
	}
}

func (g *Generator) genAGKEVerifyExpr(e *ast.AGKEVerifyExpr) Value {
	var cRef = "null"
	if e.Chain != nil {
		cVal := g.genExpr(e.Chain)
		cRef = cVal.Ref
	}
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_agke_verify(i8* %s, i8* null)\n", resTemp, cRef)
	boolTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = icmp eq i32 %s, 1\n", boolTemp, resTemp)
	return Value{
		Type: "i1",
		Ref:  boolTemp,
	}
}
