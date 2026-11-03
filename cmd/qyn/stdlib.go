package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func runStdlib(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		fmt.Println("Qadav Standard Library Modules:")
		fmt.Println("  math           (stdlib/math.ydv)           - Core mathematical constants and primitives")
		fmt.Println("  io             (stdlib/io.ydv)             - Text and integer diagnostic emission")
		fmt.Println("  linear_algebra (stdlib/linear_algebra.ydv) - Vector operations (dot product, addition, scale)")
		fmt.Println("  statistics     (stdlib/statistics.ydv)     - Statistical calculations (mean, variance)")
		fmt.Println("  string         (stdlib/string.ydv)         - String primitives and utilities")
		return 0
	}

	if args[0] == "build" {
		modules := []string{
			"stdlib/math.ydv",
			"stdlib/io.ydv",
			"stdlib/linear_algebra.ydv",
			"stdlib/statistics.ydv",
			"stdlib/string.ydv",
		}
		for _, mod := range modules {
			src, err := os.ReadFile(mod)
			if err != nil {
				if _, err2 := os.Stat(filepath.Join(".", mod)); err2 != nil {
					fmt.Fprintf(os.Stderr, "error reading stdlib module %s: %v\n", mod, err)
					return 1
				}
			}

			l := lexer.New(string(src), mod)
			p := parser.New(l, mod)
			prog := p.Parse()
			if len(p.Errors()) > 0 {
				fmt.Fprintf(os.Stderr, "parser error in %s: %d error(s)\n", mod, len(p.Errors()))
				for _, e := range p.Errors() {
					fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
				}
				return 1
			}

			c := checker.New()
			errs := c.Check(prog)
			if len(errs) > 0 {
				fmt.Fprintf(os.Stderr, "type-checker error in %s: %d error(s)\n", mod, len(errs))
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "  %s\n", e.Error())
				}
				return 1
			}

			fmt.Printf("[OK] %s: parsed and verified cleanly\n", mod)
		}
		fmt.Printf("Qadav standard library built successfully (%d modules).\n", len(modules))
		return 0
	}

	fmt.Fprintf(os.Stderr, "unknown stdlib subcommand %q. Available: list, build\n", args[0])
	return 1
}
