package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func runEmitAst(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn emit-ast <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn emit-ast: error reading %s: %v\n", filename, err)
		return 1
	}

	l := lexer.New(string(srcBytes), filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "qyn emit-ast: %d syntax error(s):\n", len(p.Errors()))
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return 1
	}

	fmt.Println(ast.ToSExpr(prog))
	return 0
}
