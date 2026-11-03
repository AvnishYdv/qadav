package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genQuasiLit(q *ast.QuasiLit) Value {
	return Value{
		Type:  "i64",
		Ref:   fmt.Sprintf("%d", q.Value),
		QType: types.Quasi,
	}
}

func (g *Generator) genMinimizeStmt(m *ast.MinimizeStmt) {
	gridVal := g.genExpr(m.Grid)
	iterVal := g.genExpr(m.Iterations)
	iterTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", iterTemp, iterVal.Ref)
	fmt.Fprintf(&g.sb, "  call void @qadav_quasi_minimize(i8* %s, i32 %s)\n", gridVal.Ref, iterTemp)
}
