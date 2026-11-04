package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "doctor":
		os.Exit(runDoctor())
	case "version", "--version", "-v":
		fmt.Printf("qyn v%s (Phase 0 M0.3)\n", version)
		os.Exit(0)
	case "build":
		os.Exit(runBuild(os.Args[2:]))
	case "run":
		os.Exit(runRun(os.Args[2:]))
	case "check":
		os.Exit(runCheck(os.Args[2:]))
	case "fmt":
		os.Exit(runFmt(os.Args[2:]))
	case "dataset":
		os.Exit(runDataset(os.Args[2:]))
	case "selfcheck":
		os.Exit(runSelfCheck(os.Args[2:]))
	case "verify-training":
		os.Exit(runVerifyTraining(os.Args[2:]))
	case "stdlib":
		os.Exit(runStdlib(os.Args[2:]))
	case "emit-tokens":
		os.Exit(runEmitTokens(os.Args[2:]))
	case "emit-ast":
		os.Exit(runEmitAst(os.Args[2:]))
	case "selfhost-lexer":
		os.Exit(runSelfhostLexer(os.Args[2:]))
	case "selfhost-parser":
		os.Exit(runSelfhostParser(os.Args[2:]))
	case "selfhost-check":
		os.Exit(runSelfhostCheck(os.Args[2:]))
	case "help", "--help", "-h":
		showUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "qyn: unknown command %q\n\n", os.Args[1])
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Fprint(os.Stderr, `Qadav Compiler (qyn) v`+version+`

Usage:
  qyn <command> [arguments]

Commands:
  build <file.ydv>    Compile a Qadav source file to native executable
  run <file.ydv>      Compile and run a Qadav source file
  check <file.ydv>    Type-check without compiling
  fmt <file.ydv>      Format a Qadav source file
  dataset inspect <file.qds> Inspect a QDS binary dataset
  selfcheck <build|verify|strip> Binary integrity operations
  stdlib <list|build> Manage Qadav standard library modules
  emit-tokens <file>  Emit standardized lexical tokens
  emit-ast <file>     Emit canonical S-expression AST
  selfhost-lexer <file> Run self-hosted lexical tokenizer
  selfhost-parser <file> Run self-hosted parser AST generator
  selfhost-check <file>  Run self-hosted type checker
  doctor              Check environment health (LLVM, Go, etc.)
  version             Display version
  help                Show this message

Examples:
  qyn doctor
  qyn build hello.ydv
  qyn run hello.ydv
`)
}
