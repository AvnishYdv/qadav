package selfhost

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

// ParseSource parses source code into an AST node using the self-hosted parser specification.
func ParseSource(src, filename string) (ast.Node, []error) {
	l := lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		var errs []error
		for i := range p.Errors() {
			errs = append(errs, &p.Errors()[i])
		}
		return nil, errs
	}
	return prog, nil
}

// ParseToSExpr parses the source code and returns the canonical S-expression string.
func ParseToSExpr(src, filename string) (string, error) {
	node, errs := ParseSource(src, filename)
	if len(errs) > 0 {
		return "", fmt.Errorf("%d parser error(s): %v", len(errs), errs[0])
	}
	return ast.ToSExpr(node), nil
}
