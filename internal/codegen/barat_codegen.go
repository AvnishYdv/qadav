package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
)

func (g *Generator) genBaratSubmitStmt(s *ast.BaratSubmitStmt) {
	var grpRef = "null"
	if slot, ok := g.vars[s.Group]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		grpRef = loaded
	} else {
		grpTemp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_barat_create_group(i32 1)\n", grpTemp)
		alloca := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
		slot := &varSlot{
			ptr:      alloca,
			llvmType: "i8*",
		}
		g.vars[s.Group] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", grpTemp, slot.ptr)
		grpRef = grpTemp
	}

	workerID := 0
	if s.Worker != nil {
		if lit, ok := s.Worker.(*ast.IntLit); ok {
			workerID = int(lit.Value)
		}
	}

	gradPtr := "null"
	if s.Gradient != nil {
		if ident, ok := s.Gradient.(*ast.Ident); ok {
			if slot, ok := g.vars[ident.Name]; ok {
				loaded := g.nextTemp()
				fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
				gradPtr = loaded
			}
		}
	}
	if gradPtr == "null" {
		dummy := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 32)\n", dummy)
		gradPtr = dummy
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_barat_submit(i8* %s, i32 %d, i8* %s, i64 32)\n", resTemp, grpRef, workerID, gradPtr)
}

func (g *Generator) genBaratAggregateStmt(s *ast.BaratAggregateStmt) {
	var grpRef = "null"
	if slot, ok := g.vars[s.Group]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		grpRef = loaded
	}

	outTensor := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 512)\n", outTensor)

	certTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 1024)\n", certTemp)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_barat_aggregate(i8* %s, i8* %s, i64 32, i8* %s)\n", resTemp, grpRef, outTensor, certTemp)

	if s.IntoGrad != "" {
		slot, ok := g.vars[s.IntoGrad]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.IntoGrad] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", outTensor, slot.ptr)
	}

	if s.IntoCert != "" {
		slot, ok := g.vars[s.IntoCert]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i8*",
			}
			g.vars[s.IntoCert] = slot
		}
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", certTemp, slot.ptr)
	}
}

func (g *Generator) genBaratVerifyStmt(s *ast.BaratVerifyStmt) {
	var certRef = "null"
	if slot, ok := g.vars[s.Cert]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		certRef = loaded
	}

	dummyTensor := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_alloc(i64 32)\n", dummyTensor)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_barat_verify(i8* %s, i8* %s, i64 32)\n", resTemp, certRef, dummyTensor)

	if s.Into != "" {
		slot, ok := g.vars[s.Into]
		if !ok {
			alloca := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = alloca i1\n", alloca)
			slot = &varSlot{
				ptr:      alloca,
				llvmType: "i1",
			}
			g.vars[s.Into] = slot
		}
		cmp := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", cmp, resTemp)
		fmt.Fprintf(&g.sb, "  store i1 %s, i1* %s\n", cmp, slot.ptr)
	}
}

func (g *Generator) genCorruptBaratStmt(s *ast.CorruptBaratStmt) {
	var certRef = "null"
	if slot, ok := g.vars[s.Cert]; ok {
		loaded := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = load i8*, i8** %s\n", loaded, slot.ptr)
		certRef = loaded
	}

	offsetVal := 0
	if s.Offset != nil {
		if lit, ok := s.Offset.(*ast.IntLit); ok {
			offsetVal = int(lit.Value)
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_barat_corrupt(i8* %s, i32 %d)\n", resTemp, certRef, offsetVal)
}
