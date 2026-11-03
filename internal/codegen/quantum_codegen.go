package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

// ── Quantum state management ──────────────────────────────────────────
//
// We use a single process-global quantum state initialised lazily.
// The global is @.qstate (i8* pointer to a qadav_qstate handle).
// Qubit IDs are stored in local alloca slots as i32.

// generateQuantumGlobal returns the global quantum state declaration for the module header.
func (g *Generator) generateQuantumGlobal() string {
	return ""
}

// genQStateLoad emits code to retrieve the process-global quantum state pointer.
// Returns the temp holding the i8* pointer.
func (g *Generator) genQStateLoad() string {
	qsTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_get_qstate()\n", qsTemp)
	return qsTemp
}

// genQubitDecl compiles a QubitDecl: allocates qubit IDs in the global state.
func (g *Generator) genQubitDecl(d *ast.QubitDecl) {
	qsPtr := g.genQStateLoad()
	for _, name := range d.Names {
		qid := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qubit_alloc(i8* %s)\n", qid, qsPtr)
		ptr := fmt.Sprintf("%%%s.ptr", name)
		fmt.Fprintf(&g.sb, "  %s = alloca i32\n", ptr)
		fmt.Fprintf(&g.sb, "  store i32 %s, i32* %s\n", qid, ptr)
		g.vars[name] = &varSlot{ptr: ptr, llvmType: "i32", qtype: types.Qubit}
	}
}

// genSuperpose compiles superpose q0 ·  → qadav_gate_h(qs, qid)
func (g *Generator) genSuperpose(s *ast.SuperposeStmt) {
	qsPtr := g.genQStateLoad()
	qidVal := g.genExpr(s.Qubit)
	qid := qidVal.Ref
	if qidVal.Type != "i32" {
		c := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", c, qid)
		qid = c
	}
	fmt.Fprintf(&g.sb, "  call void @qadav_gate_h(i8* %s, i32 %s)\n", qsPtr, qid)
}

// genEntangle compiles entangle q0 ⸴ q1 ·  → H(q0) + CNOT(q0, q1)
func (g *Generator) genEntangle(s *ast.EntangleStmt) {
	qsPtr := g.genQStateLoad()
	q1Val := g.genExpr(s.Q1)
	q2Val := g.genExpr(s.Q2)
	q1 := q1Val.Ref
	if q1Val.Type != "i32" {
		c := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", c, q1)
		q1 = c
	}
	q2 := q2Val.Ref
	if q2Val.Type != "i32" {
		c := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", c, q2)
		q2 = c
	}
	// entangle applies CNOT(control=q1, target=q2)
	fmt.Fprintf(&g.sb, "  call void @qadav_gate_cnot(i8* %s, i32 %s, i32 %s)\n", qsPtr, q1, q2)
}

// genCollapseStmt compiles collapse q0 ↦ r ·
func (g *Generator) genCollapseStmt(s *ast.CollapseStmt) {
	qsPtr := g.genQStateLoad()
	qidVal := g.genExpr(s.Qubit)
	qid := qidVal.Ref
	if qidVal.Type != "i32" {
		c := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", c, qid)
		qid = c
	}
	measTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_measure(i8* %s, i32 %s)\n", measTemp, qsPtr, qid)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", resTemp, measTemp)

	// Store into binding if Into is provided
	if s.Into != nil {
		if ident, ok := s.Into.(*ast.Ident); ok {
			if slot, exists := g.vars[ident.Name]; exists {
				fmt.Fprintf(&g.sb, "  store i64 %s, i64* %s\n", resTemp, slot.ptr)
			}
		}
	}
}

// genCollapseExpr compiles collapse q0 as an expression → i64 (0 or 1)
func (g *Generator) genCollapseExpr(e *ast.CollapseExpr) Value {
	qsPtr := g.genQStateLoad()
	qidVal := g.genExpr(e.Qubit)
	qid := qidVal.Ref
	if qidVal.Type != "i32" {
		c := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = trunc i64 %s to i32\n", c, qid)
		qid = c
	}
	measTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_measure(i8* %s, i32 %s)\n", measTemp, qsPtr, qid)
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = sext i32 %s to i64\n", resTemp, measTemp)
	return Value{Type: "i64", Ref: resTemp, QType: types.Int64}
}
