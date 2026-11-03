package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkGPUMatmulStmt(s *ast.GPUMatmulStmt) {
	if s == nil {
		return
	}
	leftTy := c.checkExpr(s.Left)
	rightTy := c.checkExpr(s.Right)

	if leftTy != nil && leftTy.Kind != types.KindTensor {
		c.errorf(s.Left.Pos(), "gpu_matmul left operand must be tensor, got %s", leftTy)
	}
	if rightTy != nil && rightTy.Kind != types.KindTensor {
		c.errorf(s.Right.Pos(), "gpu_matmul right operand must be tensor, got %s", rightTy)
	}

	if s.Audit != nil {
		c.checkExpr(s.Audit)
	}

	if id, ok := s.Target.(*ast.Ident); ok {
		c.curScope.Define(&Symbol{Name: id.Name, Type: types.TensorOfShape(types.Float64, []int{2, 2})})
	} else if s.Target != nil {
		c.checkExpr(s.Target)
	}
}

func (c *Checker) checkGPUSoftmaxStmt(s *ast.GPUSoftmaxStmt) {
	if s == nil {
		return
	}
	srcTy := c.checkExpr(s.Source)
	if srcTy != nil && srcTy.Kind != types.KindTensor {
		c.errorf(s.Source.Pos(), "gpu_softmax source operand must be tensor, got %s", srcTy)
	}

	if s.Audit != nil {
		c.checkExpr(s.Audit)
	}

	if id, ok := s.Target.(*ast.Ident); ok {
		c.curScope.Define(&Symbol{Name: id.Name, Type: types.TensorOfShape(types.Float64, []int{2, 2})})
	} else if s.Target != nil {
		c.checkExpr(s.Target)
	}
}

func (c *Checker) checkGPULayernormStmt(s *ast.GPULayernormStmt) {
	if s == nil {
		return
	}
	srcTy := c.checkExpr(s.Source)
	if srcTy != nil && srcTy.Kind != types.KindTensor {
		c.errorf(s.Source.Pos(), "gpu_layernorm source operand must be tensor, got %s", srcTy)
	}

	if s.Audit != nil {
		c.checkExpr(s.Audit)
	}

	if id, ok := s.Target.(*ast.Ident); ok {
		c.curScope.Define(&Symbol{Name: id.Name, Type: types.TensorOfShape(types.Float64, []int{2, 2})})
	} else if s.Target != nil {
		c.checkExpr(s.Target)
	}
}

func (c *Checker) checkGPUMatmulExpr(e *ast.GPUMatmulExpr) *types.Type {
	if e == nil {
		return types.Void
	}
	leftTy := c.checkExpr(e.Left)
	rightTy := c.checkExpr(e.Right)

	if leftTy != nil && leftTy.Kind != types.KindTensor {
		c.errorf(e.Left.Pos(), "gpu_matmul left operand must be tensor, got %s", leftTy)
	}
	if rightTy != nil && rightTy.Kind != types.KindTensor {
		c.errorf(e.Right.Pos(), "gpu_matmul right operand must be tensor, got %s", rightTy)
	}
	return types.TensorOfShape(types.Float64, []int{2, 2})
}

func (c *Checker) checkGPUSoftmaxExpr(e *ast.GPUSoftmaxExpr) *types.Type {
	if e == nil {
		return types.Void
	}
	srcTy := c.checkExpr(e.Source)
	if srcTy != nil && srcTy.Kind != types.KindTensor {
		c.errorf(e.Source.Pos(), "gpu_softmax source operand must be tensor, got %s", srcTy)
	}
	return types.TensorOfShape(types.Float64, []int{2, 2})
}

func (c *Checker) checkGPULayernormExpr(e *ast.GPULayernormExpr) *types.Type {
	if e == nil {
		return types.Void
	}
	srcTy := c.checkExpr(e.Source)
	if srcTy != nil && srcTy.Kind != types.KindTensor {
		c.errorf(e.Source.Pos(), "gpu_layernorm source operand must be tensor, got %s", srcTy)
	}
	return types.TensorOfShape(types.Float64, []int{2, 2})
}

func (c *Checker) checkAGKEAttestStmt(s *ast.AGKEAttestStmt) {
	if s == nil {
		return
	}
	c.checkExpr(s.Chain)
	if s.KernelName != nil {
		c.checkExpr(s.KernelName)
	}
	if s.InputTensor != nil {
		c.checkExpr(s.InputTensor)
	}
	if s.OutputTensor != nil {
		c.checkExpr(s.OutputTensor)
	}
}

func (c *Checker) checkAGKEVerifyStmt(s *ast.AGKEVerifyStmt) {
	if s == nil {
		return
	}
	c.checkExpr(s.Chain)
	if id, ok := s.Target.(*ast.Ident); ok {
		c.curScope.Define(&Symbol{Name: id.Name, Type: types.Int64})
	} else if s.Target != nil {
		c.checkExpr(s.Target)
	}
}

func (c *Checker) checkAGKEVerifyExpr(e *ast.AGKEVerifyExpr) *types.Type {
	if e == nil {
		return types.Int64
	}
	c.checkExpr(e.Chain)
	return types.Int64
}

func (c *Checker) checkCorruptGPUDigestStmt(s *ast.CorruptGPUDigestStmt) {
	if s == nil {
		return
	}
	if !c.IsTest {
		c.errorf(s.Pos(), "corrupt_gpu_digest is only permitted in test mode (--test)")
	}
	c.checkExpr(s.Chain)
	if s.Offset != nil {
		c.checkExpr(s.Offset)
	}
}
