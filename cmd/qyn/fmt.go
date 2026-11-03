package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/token"
)

func runFmt(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn fmt <file.ydv> [--check] [-w]")
		return 1
	}

	checkOnly := false
	writeInPlace := false
	var filename string

	for _, arg := range args {
		switch arg {
		case "--check":
			checkOnly = true
		case "-w":
			writeInPlace = true
		default:
			if !strings.HasPrefix(arg, "-") && filename == "" {
				filename = arg
			}
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "qyn fmt: error: no input file specified")
		return 1
	}

	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn fmt: error reading %s: %v\n", filename, err)
		return 1
	}
	original := string(srcBytes)

	l := lexer.New(original, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "qyn fmt: syntax errors in %s:\n", filename)
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return 1
	}

	formatted := formatProgram(prog)

	// Normalize CRLF to LF for comparison
	cleanOrig := strings.TrimSpace(strings.ReplaceAll(original, "\r\n", "\n"))
	cleanFmt := strings.TrimSpace(strings.ReplaceAll(formatted, "\r\n", "\n"))

	if checkOnly {
		if cleanOrig != cleanFmt {
			fmt.Printf("✗ %s is not formatted\n", filename)
			return 1
		}
		fmt.Printf("✓ %s is formatted\n", filename)
		return 0
	}

	if writeInPlace {
		if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "qyn fmt: error writing %s: %v\n", filename, err)
			return 1
		}
		return 0
	}

	fmt.Print(formatted)
	return 0
}

func formatProgram(prog *ast.Program) string {
	var sb strings.Builder
	for i, decl := range prog.Decls {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(formatDecl(decl))
	}
	sb.WriteString("\n")
	return sb.String()
}

func formatDecl(decl ast.Decl) string {
	switch d := decl.(type) {
	case *ast.ProcessDecl:
		var sb strings.Builder
		sb.WriteString("process ")
		sb.WriteString(d.Name)
		sb.WriteString("⦅")
		for i, param := range d.Params {
			if i > 0 {
				sb.WriteString("⸴ ")
			}
			if param.Type != nil {
				sb.WriteString(formatType(param.Type))
				sb.WriteString(" ")
			}
			sb.WriteString(param.Name)
		}
		sb.WriteString("⦆")
		if d.Returns != nil {
			sb.WriteString(" → ")
			sb.WriteString(formatType(d.Returns))
		}
		sb.WriteString(" ")
		sb.WriteString(formatBlock(d.Body, 0))
		return sb.String()
	default:
		return ""
	}
}

func formatType(t ast.Type) string {
	switch typ := t.(type) {
	case *ast.PrimitiveType:
		return typ.Name
	case *ast.TensorType:
		return "tensor"
	case *ast.NamedType:
		return typ.Name
	default:
		return "void"
	}
}

func formatBlock(b *ast.Block, indent int) string {
	if b == nil {
		return "⟦ ⟧"
	}
	var sb strings.Builder
	sb.WriteString("⟦\n")
	indStr := strings.Repeat("\t", indent+1)
	for _, stmt := range b.Statements {
		sb.WriteString(indStr)
		sb.WriteString(formatStmt(stmt, indent+1))
		if !isBlockStmt(stmt) {
			sb.WriteString(" ·")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("\t", indent))
	sb.WriteString("⟧")
	return sb.String()
}

func isBlockStmt(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.WhenStmt, *ast.CycleStmt, *ast.IterateStmt, *ast.Block:
		return true
	default:
		return false
	}
}

func formatStmt(stmt ast.Stmt, indent int) string {
	switch s := stmt.(type) {
	case *ast.BindStmt:
		if s.Type != nil {
			return fmt.Sprintf("%s %s ↦ %s", formatType(s.Type), s.Name, formatExpr(s.Value))
		}
		return fmt.Sprintf("%s ↦ %s", s.Name, formatExpr(s.Value))
	case *ast.AssignStmt:
		return fmt.Sprintf("%s %s %s", formatExpr(s.LHS), opToString(s.Op), formatExpr(s.RHS))
	case *ast.EmitStmt:
		return fmt.Sprintf("emit %s", formatExpr(s.Value))
	case *ast.ObserveStmt:
		return fmt.Sprintf("observe %s", formatExpr(s.Value))
	case *ast.YieldStmt:
		if s.Value != nil {
			return fmt.Sprintf("yield %s", formatExpr(s.Value))
		}
		return "yield"
	case *ast.WhenStmt:
		res := fmt.Sprintf("when %s %s", formatExpr(s.Condition), formatBlock(s.Then, indent))
		if s.Else != nil {
			res += fmt.Sprintf(" otherwise %s", formatBlock(s.Else, indent))
		}
		return res
	case *ast.CycleStmt:
		if s.Condition != nil {
			return fmt.Sprintf("cycle %s %s", formatExpr(s.Condition), formatBlock(s.Body, indent))
		}
		return fmt.Sprintf("cycle %s", formatBlock(s.Body, indent))
	case *ast.SuperposeStmt:
		return fmt.Sprintf("superpose %s", formatExpr(s.Qubit))
	case *ast.EntangleStmt:
		return fmt.Sprintf("entangle %s ⸴ %s", formatExpr(s.Q1), formatExpr(s.Q2))
	case *ast.CollapseStmt:
		if s.Into != nil {
			return fmt.Sprintf("collapse %s ↦ %s", formatExpr(s.Qubit), formatExpr(s.Into))
		}
		return fmt.Sprintf("collapse %s", formatExpr(s.Qubit))
	case *ast.HaltStmt:
		return "halt"
	case *ast.SkipStmt:
		return "skip"
	case *ast.ExprStmt:
		return formatExpr(s.Expr)
	default:
		return ""
	}
}

func formatExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", e.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("%f", e.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", e.Value)
	case *ast.Ident:
		return e.Name
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", formatExpr(e.Left), opToString(e.Op), formatExpr(e.Right))
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", opToString(e.Op), formatExpr(e.Right))
	case *ast.GroupExpr:
		return fmt.Sprintf("⦅%s⦆", formatExpr(e.Inner))
	case *ast.CallExpr:
		var args []string
		for _, a := range e.Args {
			args = append(args, formatExpr(a))
		}
		return fmt.Sprintf("%s⦅%s⦆", formatExpr(e.Func), strings.Join(args, " ⸴ "))
	default:
		return ""
	}
}

func opToString(t token.Type) string {
	switch t {
	case token.OSUM:
		return "⊕"
	case token.ODIFF:
		return "⊖"
	case token.OTIM:
		return "⊗"
	case token.ODIV:
		return "⊘"
	case token.OMOD:
		return "⊙"
	case token.EQUIV:
		return "≡"
	case token.NEQ:
		return "≠"
	case token.PREC:
		return "≺"
	case token.SUCC:
		return "≻"
	case token.PREQ:
		return "≼"
	case token.SUEQ:
		return "≽"
	case token.WEDGE:
		return "∧"
	case token.VEE:
		return "∨"
	case token.NEG:
		return "¬"
	case token.ASSIGN:
		return "←"
	case token.BIND:
		return "↦"
	case token.BIND_ADD:
		return "↦⊕"
	case token.BIND_SUB:
		return "↦⊖"
	default:
		return string(t)
	}
}
