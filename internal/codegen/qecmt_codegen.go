package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genLicenseDecl(l *ast.LicenseDecl) {
	var tRef = "null"
	if l.Teacher != nil {
		tVal := g.genExpr(l.Teacher)
		tRef = tVal.Ref
	}

	var sRef = "null"
	if l.Student != nil {
		if id, ok := l.Student.(*ast.Ident); ok {
			sNet := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i8* @.get_network_%s()\n", sNet, id.Name)
			sRef = sNet
		} else {
			sVal := g.genExpr(l.Student)
			sRef = sVal.Ref
		}
	}

	maxEpochs := 100
	if l.MaxEpochs != nil {
		if lit, ok := l.MaxEpochs.(*ast.IntLit); ok {
			maxEpochs = int(lit.Value)
		}
	}

	maxRate := 0.5
	if l.MaxRate != nil {
		if lit, ok := l.MaxRate.(*ast.FloatLit); ok {
			maxRate = lit.Value
		}
	}

	expiry := 0
	if l.Expiry != nil {
		if lit, ok := l.Expiry.(*ast.IntLit); ok {
			expiry = int(lit.Value)
		}
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qecmt_issue_license(i8* null, i8* %s, i8* %s, i64 %d, double %f, i64 %d, i8* null)\n",
		resTemp, tRef, sRef, maxEpochs, maxRate, expiry)

	if l.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", l.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i32\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i32",
			qtype:    types.License,
		}
		g.vars[l.Name] = slot
		fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", resTemp, ptr)
	}
}

func (g *Generator) genTransferStmt(s *ast.TransferStmt) {
	licVal := g.genExpr(s.License)
	var tRef = "null"
	if s.Teacher != nil {
		tVal := g.genExpr(s.Teacher)
		tRef = tVal.Ref
	}
	var sRef = "null"
	if s.Student != nil {
		if id, ok := s.Student.(*ast.Ident); ok {
			sNet := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i8* @.get_network_%s()\n", sNet, id.Name)
			sRef = sNet
		} else {
			sVal := g.genExpr(s.Student)
			sRef = sVal.Ref
		}
	}

	numLayers := len(s.Layers)
	if numLayers == 0 {
		numLayers = 1
	}

	layersArr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [%d x i32]\n", layersArr, numLayers)
	for i, l := range s.Layers {
		lVal := g.genExpr(l)
		elemPtr := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x i32], [%d x i32]* %s, i64 0, i64 %d\n", elemPtr, numLayers, numLayers, layersArr, i)
		fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", lVal.Ref, elemPtr)
	}

	castPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast [%d x i32]* %s to i32*\n", castPtr, numLayers, layersArr)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qecmt_transfer_layer(i8* null, i32 %s, i8* %s, i8* %s, i32* %s, i32 %d, i8* null)\n",
		resTemp, licVal.Ref, tRef, sRef, castPtr, len(s.Layers))
}

func (g *Generator) genDistillStmt(s *ast.DistillStmt) {
	licVal := g.genExpr(s.License)
	var sRef = "null"
	if s.Student != nil {
		if id, ok := s.Student.(*ast.Ident); ok {
			sNet := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = call i8* @.get_network_%s()\n", sNet, id.Name)
			sRef = sNet
		} else {
			sVal := g.genExpr(s.Student)
			sRef = sVal.Ref
		}
	}

	var xRef = "null"
	if s.DatasetX != nil {
		xVal := g.genExpr(s.DatasetX)
		xRef = xVal.Ref
	}
	var yRef = "null"
	if s.DatasetY != nil {
		yVal := g.genExpr(s.DatasetY)
		yRef = yVal.Ref
	}

	rateVal := "0.100000"
	if s.Rate != nil {
		val := g.genExpr(s.Rate)
		rateVal = val.Ref
	}

	epochsVal := "10"
	if s.Epochs != nil {
		val := g.genExpr(s.Epochs)
		epochsVal = val.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qecmt_distill(i8* null, i32 %s, i8* %s, i8* %s, i8* %s, double %s, i32 %s)\n",
		resTemp, licVal.Ref, sRef, xRef, yRef, rateVal, epochsVal)
}

