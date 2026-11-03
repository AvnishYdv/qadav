package checker

import (
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (c *Checker) checkDatasetDecl(d *ast.DatasetDecl) {
	pathType := c.checkExpr(d.Path)
	if pathType != nil && pathType.Kind != types.KindText {
		c.errorf(d.Path.Pos(), "dataset load path must be a string literal, got %s", pathType)
	}
	if strLit, ok := d.Path.(*ast.StringLit); ok {
		if strings.Contains(strLit.Value, "..") {
			c.errorf(d.Path.Pos(), "dataset load path contains illegal path traversal '..': %q", strLit.Value)
		}
		if !strings.HasSuffix(strLit.Value, ".qds") {
			c.errorf(d.Path.Pos(), "dataset file must have .qds extension, got %q", strLit.Value)
		}
	}
	_ = c.curScope.Define(&Symbol{
		Name: d.Name,
		Type: types.Dataset,
		Pos:  d.Pos(),
	})
}

func (c *Checker) checkNormalizeStmt(n *ast.NormalizeStmt) {
	dType := c.checkExpr(n.Dataset)
	if dType != nil && dType.Kind != types.KindDataset && dType.Kind != types.KindTensor {
		c.errorf(n.Dataset.Pos(), "normalize target must be dataset or tensor, got %s", dType)
	}
	if n.Mode != token.MINMAX && n.Mode != token.ZSCORE {
		c.errorf(n.Pos(), "normalize mode must be minmax or zscore")
	}
}

func (c *Checker) checkShuffleStmt(s *ast.ShuffleStmt) {
	dType := c.checkExpr(s.Dataset)
	if dType != nil && dType.Kind != types.KindDataset {
		c.errorf(s.Dataset.Pos(), "shuffle target must be dataset, got %s", dType)
	}
}

func (c *Checker) checkBatchStmt(b *ast.BatchStmt) {
	dType := c.checkExpr(b.Dataset)
	if dType != nil && dType.Kind != types.KindDataset {
		c.errorf(b.Dataset.Pos(), "batch target must be dataset, got %s", dType)
	}
	sizeType := c.checkExpr(b.Size)
	if sizeType != nil && !sizeType.IsInteger() {
		c.errorf(b.Size.Pos(), "batch size must be an integer, got %s", sizeType)
	}
	if intLit, ok := b.Size.(*ast.IntLit); ok {
		if intLit.Value <= 0 {
			c.errorf(b.Size.Pos(), "batch size must be positive, got %d", intLit.Value)
		} else if intLit.Value > 10000 {
			c.errorf(b.Size.Pos(), "batch size exceeds maximum of 10000, got %d", intLit.Value)
		}
	}
	if b.Body != nil {
		c.enterScope()
		for _, stmt := range b.Body.Statements {
			c.checkStmt(stmt)
		}
		c.leaveScope()
	}
}
