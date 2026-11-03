package checker

import (
	"fmt"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

// NetworkInfo stores checked type and architectural metadata for a neural network.
type NetworkInfo struct {
	Name       string
	InputSize  int
	OutputSize int
	Layers     []*LayerInfo
	Decl       *ast.NetworkDecl
	Type       *types.Type
}

// LayerInfo stores input and output dimensions for a layer.
type LayerInfo struct {
	Kind    token.Type
	InSize  int
	OutSize int
}

func (c *Checker) registerNetworkSignature(net *ast.NetworkDecl) {
	if net == nil {
		return
	}
	info, err := c.analyzeNetwork(net)
	if err != nil {
		c.errorf(net.Pos(), "%s", err.Error())
		return
	}

	c.networks[net.Name] = info

	// Define callable symbol for network forward evaluation: Net⦅x⦆
	procType := &types.Type{
		Kind:    types.KindProcess,
		Name:    net.Name,
		Params:  []*types.Type{types.TensorOfShape(types.Float64, []int{info.InputSize})},
		Results: []*types.Type{types.TensorOfShape(types.Float64, []int{info.OutputSize})},
	}
	info.Type = procType

	sym := &Symbol{
		Name: net.Name,
		Type: procType,
		Pos:  net.Pos(),
	}
	if err := c.globalScope.Define(sym); err != nil {
		c.errorf(net.Pos(), "%s", err.Error())
	}
}

func (c *Checker) analyzeNetwork(net *ast.NetworkDecl) (*NetworkInfo, error) {
	info := &NetworkInfo{
		Name:   net.Name,
		Decl:   net,
		Layers: []*LayerInfo{},
	}

	if len(net.Layers) == 0 {
		return info, nil
	}

	currentDim := 0
	isFirst := true

	for idx, layer := range net.Layers {
		var inDim, outDim int

		switch layer.Kind {
		case token.DENSE:
			if len(layer.Args) >= 2 {
				inVal := c.evalIntArg(layer.Args[0])
				outVal := c.evalIntArg(layer.Args[1])
				inDim = inVal
				outDim = outVal
			} else if len(layer.Args) == 1 {
				outVal := c.evalIntArg(layer.Args[0])
				if isFirst {
					inDim = outVal
					outDim = outVal
				} else {
					inDim = currentDim
					outDim = outVal
				}
			} else {
				return nil, fmt.Errorf("layer %d (dense) in network %q requires dimension argument(s)", idx+1, net.Name)
			}

			if isFirst {
				info.InputSize = inDim
				currentDim = inDim
				isFirst = false
			}

			if inDim != currentDim {
				return nil, fmt.Errorf("layer %d in network %q shape mismatch: expected input %d, got %d", idx+1, net.Name, currentDim, inDim)
			}
			currentDim = outDim

		case token.RELU, token.SIGMOID, token.SOFTMAX, token.LAYERNORM:
			if isFirst {
				return nil, fmt.Errorf("first layer of network %q cannot be an activation layer without input dimension", net.Name)
			}
			inDim = currentDim
			outDim = currentDim

		default:
			inDim = currentDim
			outDim = currentDim
		}

		layer.InSize = inDim
		layer.OutSize = outDim
		info.Layers = append(info.Layers, &LayerInfo{
			Kind:    layer.Kind,
			InSize:  inDim,
			OutSize: outDim,
		})
	}

	info.OutputSize = currentDim
	return info, nil
}

func (c *Checker) evalIntArg(expr ast.Expr) int {
	if expr == nil {
		return 1
	}
	if intLit, ok := expr.(*ast.IntLit); ok {
		return int(intLit.Value)
	}
	return 1
}

func (c *Checker) checkNetworkDecl(net *ast.NetworkDecl) {
	// Already registered and validated in first pass
}

func (c *Checker) checkLearnDecl(learn *ast.LearnDecl) {
	if learn == nil {
		return
	}

	netName := learn.NetworkName
	if netName == "" {
		netName = learn.Name
	}

	if _, exists := c.networks[netName]; !exists {
		c.errorf(learn.Pos(), "unknown network %q in learn block", netName)
	}

	loss := strings.ToLower(learn.Loss)
	if loss != "" && loss != "mse" && loss != "cross_entropy" {
		c.errorf(learn.Pos(), "invalid loss function %q, expected 'mse' or 'cross_entropy'", learn.Loss)
	}

	opt := strings.ToLower(learn.Optimizer)
	if opt != "" && opt != "sgd" && opt != "adam" {
		c.errorf(learn.Pos(), "invalid optimizer %q, expected 'sgd' or 'adam'", learn.Optimizer)
	}

	if learn.LR != nil {
		lrType := c.checkExpr(learn.LR)
		if lrType != nil && !lrType.IsNumeric() {
			c.errorf(learn.LR.Pos(), "learning rate must be numeric, got %s", lrType)
		}
	}

	if learn.Epochs != nil {
		epochsType := c.checkExpr(learn.Epochs)
		if epochsType != nil && !epochsType.IsInteger() && !epochsType.IsNumeric() {
			c.errorf(learn.Epochs.Pos(), "epochs count must be integer, got %s", epochsType)
		}
	}

	if learn.BatchSize != nil {
		c.checkExpr(learn.BatchSize)
	}
}
