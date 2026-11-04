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

func runSelfhostCheck(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn selfhost-check <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-check: error reading %s: %v\n", filename, err)
		return 1
	}

	msg, code := selfhost.CheckSource(string(srcBytes), filename)
	if code != 0 {
		fmt.Fprint(os.Stderr, msg)
		return code
	}
	fmt.Print(msg)
	return 0
}

func runSelfhostCodegen(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn selfhost-codegen <file.ydv>")
		return 1
	}

	filename := args[0]
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn selfhost-codegen: error reading %s: %v\n", filename, err)
		return 1
	}

	ir, code := selfhost.CodegenSource(string(srcBytes), filename)
	if code != 0 {
		fmt.Fprint(os.Stderr, ir)
		return code
	}
	fmt.Print(ir)
	return 0
}
