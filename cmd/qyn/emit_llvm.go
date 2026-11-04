package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/codegen"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func runEmitLLVM(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn emit-llvm <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn emit-llvm: error reading %s: %v\n", filename, err)
		return 1
	}

	l := lexer.New(string(srcBytes), filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "qyn emit-llvm: %d syntax error(s):\n", len(p.Errors()))
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return 1
	}

	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "qyn emit-llvm: %d type error(s):\n", len(typeErrors))
		for _, typeErr := range typeErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", typeErr.Error())
		}
		return 1
	}

	g := codegen.New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn emit-llvm: codegen error: %v\n", err)
		return 1
	}

	fmt.Print(ir)
	return 0
}
