package ast

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/token"
)

// Position represents a location in source code (1-based line/col).
type Position struct {
	Line int
	Col  int
	File string
}

// String returns "file:line:col" format.
func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
}

// FromToken converts a token.Token to a Position.
func FromToken(tok token.Token) Position {
	return Position{Line: tok.Line, Col: tok.Col, File: tok.File}
}
