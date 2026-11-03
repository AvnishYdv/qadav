package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/token"
)

type nodeCounter struct {
	count int
}

func (c *nodeCounter) Visit(node ast.Node) {
	if node != nil {
		c.count++
	}
}

func runCheck(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn check <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn check: error reading %s: %v\n", filename, err)
		return 1
	}
	src := string(srcBytes)

	// Step 1: Lexical analysis (count tokens)
	tokenCount := 0
	l := lexer.New(src, filename)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		tokenCount++
	}

	// Step 2: Parsing
	l = lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "✗ %s: %d syntax error(s):\n", filename, len(p.Errors()))
		for _, e := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return 1
	}

	// Step 3: Type Checking
	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "✗ %s: %d type error(s):\n", filename, len(typeErrors))
		for _, e := range typeErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
		}
		return 1
	}

	nc := &nodeCounter{}
	ast.Walk(nc, prog)

	fmt.Printf("✓ %s: 0 errors, 0 warnings (%d tokens, %d AST nodes)\n", filename, tokenCount, nc.count)
	return 0
}
