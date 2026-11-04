package checker

import (
	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkLicenseDecl(l *ast.LicenseDecl) {
	if l.Name != "" {
		_ = c.curScope.Define(&Symbol{
			Name: l.Name,
			Type: types.License,
			Pos:  l.Pos(),
		})
	}

	if l.Teacher != nil {
		tType := c.checkExpr(l.Teacher)
		if tType != nil && tType.Kind != types.KindVault {
			c.errorf(l.Teacher.Pos(), "license teacher must be a vault, got %s", tType)
		}
	}

	if l.Student != nil {
		_ = c.checkExpr(l.Student)
	}

	if l.MaxRate != nil {
		rType := c.checkExpr(l.MaxRate)
		if lit, ok := l.MaxRate.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(l.MaxRate.Pos(), "max_rate must be between 0.0 and 1.0, got %f", lit.Value)
			}
		} else if rType != nil && !rType.IsFloat() {
			c.errorf(l.MaxRate.Pos(), "max_rate must be float, got %s", rType)
		}
	}

	if l.MaxEpochs != nil {
		eType := c.checkExpr(l.MaxEpochs)
		if lit, ok := l.MaxEpochs.(*ast.IntLit); ok {
			if lit.Value <= 0 {
				c.errorf(l.MaxEpochs.Pos(), "max_epochs must be positive, got %d", lit.Value)
			}
		} else if eType != nil && !eType.IsInteger() {
			c.errorf(l.MaxEpochs.Pos(), "max_epochs must be integer, got %s", eType)
		}
	}

	if l.Expiry != nil {
		_ = c.checkExpr(l.Expiry)
	}
}

func (c *Checker) checkTransferStmt(s *ast.TransferStmt) {
	if s.License != nil {
		lType := c.checkExpr(s.License)
		if lType != nil && lType.Kind != types.KindLicense {
			c.errorf(s.License.Pos(), "transfer target must be a license, got %s", lType)
		}
	}

	if s.Teacher != nil {
		_ = c.checkExpr(s.Teacher)
	}
	if s.Student != nil {
		_ = c.checkExpr(s.Student)
	}

	for _, layer := range s.Layers {
		lType := c.checkExpr(layer)
		if lit, ok := layer.(*ast.IntLit); ok {
			if lit.Value < 0 {
				c.errorf(layer.Pos(), "layer index cannot be negative, got %d", lit.Value)
			}
		} else if lType != nil && !lType.IsInteger() {
			c.errorf(layer.Pos(), "layer index must be integer, got %s", lType)
		}
	}
}

func (c *Checker) checkDistillStmt(s *ast.DistillStmt) {
	if s.License != nil {
		lType := c.checkExpr(s.License)
		if lType != nil && lType.Kind != types.KindLicense {
			c.errorf(s.License.Pos(), "distill target must be a license, got %s", lType)
		}
	}

	if s.Student != nil {
		_ = c.checkExpr(s.Student)
	}

	if s.DatasetX != nil {
		_ = c.checkExpr(s.DatasetX)
	}
	if s.DatasetY != nil {
		_ = c.checkExpr(s.DatasetY)
	}

	if s.Rate != nil {
		rType := c.checkExpr(s.Rate)
		if lit, ok := s.Rate.(*ast.FloatLit); ok {
			if lit.Value < 0.0 || lit.Value > 1.0 {
				c.errorf(s.Rate.Pos(), "distillation rate must be between 0.0 and 1.0, got %f", lit.Value)
			}
		} else if rType != nil && !rType.IsFloat() {
			c.errorf(s.Rate.Pos(), "distillation rate must be float, got %s", rType)
		}
	}

	if s.Epochs != nil {
		eType := c.checkExpr(s.Epochs)
		if lit, ok := s.Epochs.(*ast.IntLit); ok {
			if lit.Value <= 0 {
				c.errorf(s.Epochs.Pos(), "distill epochs must be positive, got %d", lit.Value)
			}
		} else if eType != nil && !eType.IsInteger() {
			c.errorf(s.Epochs.Pos(), "distill epochs must be integer, got %s", eType)
		}
	}
}
