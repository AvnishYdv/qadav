package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genDatasetDecl(d *ast.DatasetDecl) {
	pathVal := g.genExpr(d.Path)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_dataset_load(i8* %s)\n", resTemp, pathVal.Ref)

	ptr := fmt.Sprintf("%%%s.ptr", d.Name)
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
	slot := &varSlot{
		ptr:      ptr,
		llvmType: "i8*",
		qtype:    types.Dataset,
	}
	g.vars[d.Name] = slot
	fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
}

func (g *Generator) genNormalizeStmt(n *ast.NormalizeStmt) {
	val := g.genExpr(n.Dataset)
	if n.Mode == token.ZSCORE {
		fmt.Fprintf(&g.sb, "  call void @qadav_dataset_normalize_zscore(i8* %s)\n", val.Ref)
	} else {
		fmt.Fprintf(&g.sb, "  call void @qadav_dataset_normalize_minmax(i8* %s)\n", val.Ref)
	}
}

func (g *Generator) genShuffleStmt(s *ast.ShuffleStmt) {
	val := g.genExpr(s.Dataset)
	fmt.Fprintf(&g.sb, "  call void @qadav_dataset_shuffle(i8* %s)\n", val.Ref)
}

func (g *Generator) genBatchStmt(b *ast.BatchStmt) {
	if b.Body != nil {
		for _, s := range b.Body.Statements {
			g.genStmt(s)
		}
	}
}
