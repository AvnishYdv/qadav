package selfhost

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/codegen"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

// CodegenSource compiles source into LLVM IR using the self-hosted codegen driver.
func CodegenSource(src, filename string) (string, int) {
	l := lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Sprintf("syntax error in %s\n", filename), 1
	}

	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		return fmt.Sprintf("type error in %s\n", filename), 1
	}

	g := codegen.New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		return fmt.Sprintf("codegen error: %v\n", err), 1
	}

	return ir, 0
}
