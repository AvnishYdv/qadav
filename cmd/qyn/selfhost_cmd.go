package main

import (
	"fmt"
	"os"

	"github.com/qadav-lang/qadav/internal/selfhost"
)

func runSelfhostLexer(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn selfhost-lexer <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-lexer: error reading %s: %v\n", filename, err)
		return 1
	}

	tokens, err := selfhost.LexSource(string(srcBytes), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-lexer: %v\n", err)
		return 1
	}

	fmt.Print(selfhost.FormatTokens(tokens))
	return 0
}

func runSelfhostParser(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn selfhost-parser <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-parser: error reading %s: %v\n", filename, err)
		return 1
	}

	sexpr, err := selfhost.ParseToSExpr(string(srcBytes), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-parser: %v\n", err)
		return 1
	}

	fmt.Println(sexpr)
	return 0
}
