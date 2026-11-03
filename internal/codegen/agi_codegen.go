package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genMemoryDecl(m *ast.MemoryDecl) {
	capRef := "64"
	if m.Capacity != nil {
		capVal := g.genExpr(m.Capacity)
		capRef = capVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_memory_create(i64 %s)\n", resTemp, capRef)

	ptr := fmt.Sprintf("%%%s.ptr", m.Name)
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
	slot := &varSlot{
		ptr:      ptr,
		llvmType: "i8*",
		qtype:    types.Memory,
	}
	g.vars[m.Name] = slot
	fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
}

func (g *Generator) genRememberStmt(s *ast.RememberStmt) {
	memVal := g.genExpr(s.Memory)
	keyVal := g.genExpr(s.Key)
	valVal := g.genExpr(s.Value)

	valRef := valVal.Ref
	if valVal.Type == "double" {
		conv := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = fptosi double %s to i64\n", conv, valVal.Ref)
		valRef = conv
	}

	fmt.Fprintf(&g.sb, "  call void @qadav_memory_remember(i8* %s, i8* %s, i64 %s)\n", memVal.Ref, keyVal.Ref, valRef)
}

func (g *Generator) genRecallExpr(e *ast.RecallExpr) Value {
	memVal := g.genExpr(e.Memory)
	keyVal := g.genExpr(e.Key)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i64 @qadav_memory_recall(i8* %s, i8* %s)\n", resTemp, memVal.Ref, keyVal.Ref)
	return Value{
		Type:  "i64",
		Ref:   resTemp,
		QType: types.Int64,
	}
}

func (g *Generator) genGraphDecl(d *ast.GraphDecl) {
	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_graph_create()\n", resTemp)

	ptr := fmt.Sprintf("%%%s.ptr", d.Name)
	fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
	slot := &varSlot{
		ptr:      ptr,
		llvmType: "i8*",
		qtype:    types.Graph,
	}
	g.vars[d.Name] = slot
	fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
}

func (g *Generator) genNodeStmt(s *ast.NodeStmt) {
	graphVal := g.genExpr(s.Graph)
	idVal := g.genExpr(s.ID)
	dataRef := "0"
	if s.Data != nil {
		dataVal := g.genExpr(s.Data)
		dataRef = dataVal.Ref
		if dataVal.Type == "double" {
			conv := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = fptosi double %s to i64\n", conv, dataVal.Ref)
			dataRef = conv
		}
	}

	fmt.Fprintf(&g.sb, "  call void @qadav_graph_node(i8* %s, i8* %s, i64 %s)\n", graphVal.Ref, idVal.Ref, dataRef)
}

func (g *Generator) genLinkStmt(s *ast.LinkStmt) {
	graphVal := g.genExpr(s.Graph)
	fromVal := g.genExpr(s.From)
	toVal := g.genExpr(s.To)
	weightRef := "1.000000"
	if s.Weight != nil {
		wVal := g.genExpr(s.Weight)
		if wVal.Type == "i64" {
			conv := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = sitofp i64 %s to double\n", conv, wVal.Ref)
			weightRef = conv
		} else {
			weightRef = wVal.Ref
		}
	}

	fmt.Fprintf(&g.sb, "  call void @qadav_graph_link(i8* %s, i8* %s, i8* %s, double %s)\n", graphVal.Ref, fromVal.Ref, toVal.Ref, weightRef)
}

func (g *Generator) genTraceExpr(e *ast.TraceExpr) Value {
	graphVal := g.genExpr(e.Graph)
	startVal := g.genExpr(e.Start)
	depthRef := "2"
	if e.Depth != nil {
		dVal := g.genExpr(e.Depth)
		depthRef = dVal.Ref
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_graph_trace(i8* %s, i8* %s, i64 %s)\n", resTemp, graphVal.Ref, startVal.Ref, depthRef)
	return Value{
		Type:  "i8*",
		Ref:   resTemp,
		QType: types.TensorOf(types.Int64),
	}
}

func (g *Generator) genProcedureDecl(p *ast.ProcedureDecl) {
	// Procedures are emitted as functions: @proc_<name>
	fnName := fmt.Sprintf("@proc_%s", p.Name)
	var params []string
	oldVars := g.vars
	g.vars = make(map[string]*varSlot)

	for _, param := range p.Params {
		params = append(params, fmt.Sprintf("i64 %%%s", param.Name))
	}

	fmt.Fprintf(&g.sb, "\ndefine i64 %s(%s) {\nentry:\n", fnName, stringsJoin(params, ", "))
	for _, param := range p.Params {
		ptr := fmt.Sprintf("%%%s.ptr", param.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i64\n", ptr)
		fmt.Fprintf(&g.sb, "  store i64 %%%s, i64* %s\n", param.Name, ptr)
		g.vars[param.Name] = &varSlot{
			ptr:      ptr,
			llvmType: "i64",
			qtype:    types.Int64,
		}
	}

	if p.Body != nil {
		for _, s := range p.Body.Statements {
			g.genStmt(s)
		}
	}

	if !g.isTerminated {
		fmt.Fprintf(&g.sb, "  ret i64 0\n")
	}
	fmt.Fprintf(&g.sb, "}\n")

	g.vars = oldVars
	g.isTerminated = false
}

func (g *Generator) genExecuteExpr(e *ast.ExecuteExpr) Value {
	procName := ""
	if id, ok := e.Procedure.(*ast.Ident); ok {
		procName = id.Name
	}

	var argRefs []string
	for _, arg := range e.Args {
		val := g.genExpr(arg)
		argRef := val.Ref
		if val.Type == "double" {
			conv := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = fptosi double %s to i64\n", conv, val.Ref)
			argRef = conv
		}
		argRefs = append(argRefs, fmt.Sprintf("i64 %s", argRef))
	}

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i64 @proc_%s(%s)\n", resTemp, procName, stringsJoin(argRefs, ", "))
	return Value{
		Type:  "i64",
		Ref:   resTemp,
		QType: types.Int64,
	}
}

func (g *Generator) genExecuteStmt(s *ast.ExecuteStmt) {
	procName := ""
	if id, ok := s.Procedure.(*ast.Ident); ok {
		procName = id.Name
	}

	var argRefs []string
	for _, arg := range s.Args {
		val := g.genExpr(arg)
		argRef := val.Ref
		if val.Type == "double" {
			conv := g.nextTemp()
			fmt.Fprintf(&g.sb, "  %s = fptosi double %s to i64\n", conv, val.Ref)
			argRef = conv
		}
		argRefs = append(argRefs, fmt.Sprintf("i64 %s", argRef))
	}

	fmt.Fprintf(&g.sb, "  call i64 @proc_%s(%s)\n", procName, stringsJoin(argRefs, ", "))
}

func stringsJoin(items []string, sep string) string {
	res := ""
	for i, it := range items {
		if i > 0 {
			res += sep
		}
		res += it
	}
	return res
}
