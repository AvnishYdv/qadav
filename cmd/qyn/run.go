package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/codegen"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/llvm"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/runtime"
)

func runRun(args []string) int {
	var filename string
	isTest := false
	for _, arg := range args {
		if arg == "--test" {
			isTest = true
		} else if !strings.HasPrefix(arg, "-") && filename == "" {
			filename = arg
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "Usage: qyn run <file.ydv> [--test] [arguments]")
		return 1
	}
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn run: error reading %s: %v\n", filename, err)
		return 1
	}
	src := string(srcBytes)

	l := lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "qyn run: syntax errors in %s:\n", filename)
		for _, parseErr := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", parseErr.Error())
		}
		return 1
	}

	c := checker.New()
	c.IsTest = isTest
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "qyn run: type errors in %s:\n", filename)
		for _, typeErr := range typeErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", typeErr.Error())
		}
		return 1
	}

	g := codegen.New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn run: codegen error: %v\n", err)
		return 1
	}

	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	llPath := base + ".ll"
	_ = os.WriteFile(llPath, []byte(ir), 0644)

	tc, err := llvm.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn run: LLVM toolchain not available: %v\n", err)
		return 1
	}

	rtObjs, err := runtime.Build(tc, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn run: runtime build error: %v\n", err)
		return 1
	}

	tmpExe := filepath.Join(os.TempDir(), fmt.Sprintf("qadav_run_%d.exe", time.Now().UnixNano()))
	defer func() { _ = os.Remove(tmpExe) }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], tmpExe, rtObjs[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "qyn run: compilation error: %v\n", err)
		return 1
	}

	cmd := exec.CommandContext(ctx, tmpExe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "qyn run: execution error: %v\n", err)
		return 1
	}

	return 0
}
