package codegen

import (
	"fmt"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

// Layer kinds matching qadav_nn.h
const (
	layerKindDense     = 1
	layerKindReLU      = 2
	layerKindSigmoid   = 3
	layerKindSoftmax   = 4
	layerKindLayerNorm = 5
)

// Loss kinds matching qadav_nn.h
const (
	lossKindMSE          = 1
	lossKindCrossEntropy = 2
)

// Optimizer kinds matching qadav_nn.h
const (
	optKindSGD  = 1
	optKindAdam = 2
)

func layerKindCode(kind token.Type) int {
	switch kind {
	case token.DENSE:
		return layerKindDense
	case token.RELU:
		return layerKindReLU
	case token.SIGMOID:
		return layerKindSigmoid
	case token.SOFTMAX:
		return layerKindSoftmax
	case token.LAYERNORM:
		return layerKindLayerNorm
	default:
		return layerKindDense
	}
}

func lossKindCode(loss string) int {
	switch strings.ToLower(loss) {
	case "cross_entropy":
		return lossKindCrossEntropy
	default:
		return lossKindMSE
	}
}

func optKindCode(opt string) int {
	switch strings.ToLower(opt) {
	case "adam":
		return optKindAdam
	default:
		return optKindSGD
	}
}

// genNetworkGlobals emits the global descriptor arrays and getter function for a network.
func (g *Generator) genNetworkGlobals(net *ast.NetworkDecl) string {
	var sb strings.Builder
	name := net.Name
	numLayers := len(net.Layers)
	if numLayers == 0 {
		return ""
	}

	kinds := make([]string, numLayers)
	inSizes := make([]string, numLayers)
	outSizes := make([]string, numLayers)

	for i, l := range net.Layers {
		kinds[i] = fmt.Sprintf("i32 %d", layerKindCode(l.Kind))
		inSizes[i] = fmt.Sprintf("i32 %d", l.InSize)
		outSizes[i] = fmt.Sprintf("i32 %d", l.OutSize)
	}

	// 1. Global pointer storage
	fmt.Fprintf(&sb, "@.net.%s = internal global i8* null\n", name)
	fmt.Fprintf(&sb, "@.net.%s.kinds = private unnamed_addr constant [%d x i32] [%s]\n", name, numLayers, strings.Join(kinds, ", "))
	fmt.Fprintf(&sb, "@.net.%s.in_sizes = private unnamed_addr constant [%d x i32] [%s]\n", name, numLayers, strings.Join(inSizes, ", "))
	fmt.Fprintf(&sb, "@.net.%s.out_sizes = private unnamed_addr constant [%d x i32] [%s]\n\n", name, numLayers, strings.Join(outSizes, ", "))

	// 2. Getter function
	fmt.Fprintf(&sb, "define internal i8* @.get_network_%s() {\n", name)
	fmt.Fprintf(&sb, "entry:\n")
	fmt.Fprintf(&sb, "  %%0 = load i8*, i8** @.net.%s\n", name)
	fmt.Fprintf(&sb, "  %%1 = icmp eq i8* %%0, null\n")
	fmt.Fprintf(&sb, "  br i1 %%1, label %%init, label %%done\n\n")
	fmt.Fprintf(&sb, "init:\n")
	fmt.Fprintf(&sb, "  %%kp = getelementptr inbounds [%d x i32], [%d x i32]* @.net.%s.kinds, i64 0, i64 0\n", numLayers, numLayers, name)
	fmt.Fprintf(&sb, "  %%ip = getelementptr inbounds [%d x i32], [%d x i32]* @.net.%s.in_sizes, i64 0, i64 0\n", numLayers, numLayers, name)
	fmt.Fprintf(&sb, "  %%op = getelementptr inbounds [%d x i32], [%d x i32]* @.net.%s.out_sizes, i64 0, i64 0\n", numLayers, numLayers, name)
	fmt.Fprintf(&sb, "  %%created = call i8* @qadav_network_create(i32 %d, i32* %%kp, i32* %%ip, i32* %%op)\n", numLayers)
	fmt.Fprintf(&sb, "  store i8* %%created, i8** @.net.%s\n", name)
	fmt.Fprintf(&sb, "  br label %%done\n\n")
	fmt.Fprintf(&sb, "done:\n")
	fmt.Fprintf(&sb, "  %%res = phi i8* [ %%0, %%entry ], [ %%created, %%init ]\n")
	fmt.Fprintf(&sb, "  ret i8* %%res\n")
	fmt.Fprintf(&sb, "}\n\n")

	return sb.String()
}

// genNetworkForward emits an evaluation of Model(input).
func (g *Generator) genNetworkForward(netInfo *checker.NetworkInfo, netName string, arg ast.Expr) Value {
	inputVal := g.genExpr(arg)

	netPtrTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @.get_network_%s()\n", netPtrTemp, netName)

	predTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_network_forward(i8* %s, i8* %s)\n", predTemp, netPtrTemp, inputVal.Ref)

	outDim := 1
	if netInfo != nil {
		outDim = netInfo.OutputSize
	}
	retType := types.TensorOfShape(types.Float64, []int{outDim})

	return Value{
		Type:  "i8*",
		Ref:   predTemp,
		QType: retType,
	}
}

// genLearnStmt compiles a learn statement block into a call to @qadav_train_epochs.
func (g *Generator) genLearnStmt(s *ast.LearnDecl) {
	netName := s.NetworkName
	if netName == "" {
		netName = s.Name
	}

	netPtrTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @.get_network_%s()\n", netPtrTemp, netName)

	// Resolve dataset X and Y pointers
	var xRef, yRef string

	if s.OnYName != "" {
		if slot, ok := g.vars[s.OnName]; ok {
			xRef = g.loadVarRef(slot)
		}
		if slot, ok := g.vars[s.OnYName]; ok {
			yRef = g.loadVarRef(slot)
		}
	} else {
		// Check standard patterns: OnName_x and OnName_y, or OnName and y, or x and y
		if slot, ok := g.vars[s.OnName+"_x"]; ok {
			xRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars[s.OnName]; ok {
			xRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["x"]; ok {
			xRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["dataset_x"]; ok {
			xRef = g.loadVarRef(slot)
		}

		if slot, ok := g.vars[s.OnName+"_y"]; ok {
			yRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["y"]; ok {
			yRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["dataset_y"]; ok {
			yRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["labels"]; ok {
			yRef = g.loadVarRef(slot)
		} else if slot, ok := g.vars["target"]; ok {
			yRef = g.loadVarRef(slot)
		}
	}

	if xRef == "" {
		// Fallback: pass null if not found
		xRef = "null"
	}
	if yRef == "" {
		// If y not found separately, use xRef
		yRef = xRef
	}

	lossCode := lossKindCode(s.Loss)
	optCode := optKindCode(s.Optimizer)

	// Learning rate (default 0.01)
	lrVal := "0.010000"
	if s.LR != nil {
		val := g.genExpr(s.LR)
		switch val.Type {
		case "double":
			lrVal = val.Ref
		case "float":
			ext := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = fpext float %s to double\n", ext, val.Ref)
			lrVal = ext
		default:
			ext := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = sitofp %s %s to double\n", ext, val.Type, val.Ref)
			lrVal = ext
		}
	}

	// Epochs (default 100)
	epochsVal := "100"
	if s.Epochs != nil {
		val := g.genExpr(s.Epochs)
		switch val.Type {
		case "i32":
			epochsVal = val.Ref
		case "i64":
			trunc := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", trunc, val.Ref)
			epochsVal = trunc
		default:
			epochsVal = val.Ref
		}
	}

	fmt.Fprintf(&g.sb, "  call void @qadav_train_epochs(i8* %s, i8* %s, i8* %s, i32 %d, i32 %d, double %s, i32 %s)\n",
		netPtrTemp, xRef, yRef, lossCode, optCode, lrVal, epochsVal)

	for _, slot := range g.vars {
		if slot.qtype != nil && slot.qtype.Kind == types.KindProvenance {
			provRef := g.loadVarRef(slot)
			fmt.Fprintf(&g.sb, "  call void @qadav_mctp_record_epochs(i8* %s, i32 %s, i8* %s, i64 128, i8* %s, i64 64, double 0.5)\n",
				provRef, epochsVal, netPtrTemp, xRef)
		}
	}
}

func (g *Generator) loadVarRef(slot *varSlot) string {
	loadTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = load %s, %s* %s\n", loadTemp, slot.llvmType, slot.llvmType, slot.ptr)
	return loadTemp
}
