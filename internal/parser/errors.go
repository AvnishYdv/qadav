package parser

import (
	"fmt"
	"strings"

	"github.com/qadav-lang/qadav/internal/ast"
)

// ParseError records an error encountered during parsing.
type ParseError struct {
	Pos     ast.Position
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// ErrorList accumulates multiple syntax errors across parsing without terminating early.
type ErrorList []ParseError

func (l ErrorList) Error() string {
	if len(l) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, e := range l {
		sb.WriteString(e.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}
