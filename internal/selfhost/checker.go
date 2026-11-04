package selfhost

import (
	"fmt"
	"strings"

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

// CheckSource runs semantic analysis and type checking using the self-hosted
// checker specification and produces formatted diagnostic messages.
func CheckSource(src, filename string) (string, int) {
	// 1. Lexical analysis
	tokenCount := 0
	l := lexer.New(src, filename)
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		tokenCount++
	}

	// 2. Parsing
	l = lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("✗ %s: %d syntax error(s):\n", filename, len(p.Errors())))
		for _, e := range p.Errors() {
			sb.WriteString(fmt.Sprintf("  %s\n", e.Error()))
		}
		return sb.String(), 1
	}

	counter := &nodeCounter{}
	ast.Walk(counter, prog)

	// 3. Type Checking
	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("✗ %s: %d type error(s):\n", filename, len(typeErrors)))
		for _, e := range typeErrors {
			sb.WriteString(fmt.Sprintf("  %s\n", e.Error()))
		}
		return sb.String(), 1
	}

	return fmt.Sprintf("✓ %s: 0 errors, 0 warnings (%d tokens, %d AST nodes)\n", filename, tokenCount, counter.count), 0
}
