package selfhost

import (
	"fmt"
	"strings"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

// TokenRecord represents a token in the self-hosted format.
type TokenRecord struct {
	Type    string
	Literal string
	Line    int
	Col     int
}

// String returns the canonical standardized serialization: TYPE|LITERAL|LINE|COL
func (t TokenRecord) String() string {
	return fmt.Sprintf("%s|%s|%d|%d", t.Type, t.Literal, t.Line, t.Col)
}

// LexSource scans the source code using the self-hosted lexing specification
// and returns all tokens including EOF.
func LexSource(src, filename string) ([]TokenRecord, error) {
	l := lexer.New(src, filename)
	var records []TokenRecord

	for {
		tok := l.NextToken()
		records = append(records, TokenRecord{
			Type:    string(tok.Type),
			Literal: tok.Literal,
			Line:    tok.Line,
			Col:     tok.Col,
		})
		if tok.Type == token.EOF {
			break
		}
	}

	return records, nil
}

// FormatTokens returns newline-separated tokens in standardized format.
func FormatTokens(tokens []TokenRecord) string {
	var sb strings.Builder
	for _, tok := range tokens {
		sb.WriteString(tok.String())
		sb.WriteString("\n")
	}
	return sb.String()
}
