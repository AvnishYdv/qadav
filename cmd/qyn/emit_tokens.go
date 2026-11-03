package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func runEmitTokens(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn emit-tokens <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn emit-tokens: error reading %s: %v\n", filename, err)
		return 1
	}

	l := lexer.New(string(srcBytes), filename)
	for {
		tok := l.NextToken()
		fmt.Printf("%s|%s|%d|%d\n", tok.Type, tok.Literal, tok.Line, tok.Col)
		if tok.Type == token.EOF {
			break
		}
	}

	return 0
}
