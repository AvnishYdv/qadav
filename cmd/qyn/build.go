package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/codegen"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/llvm"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/runtime"
	"github.com/qadav-lang/qadav/internal/selfcheck"
)

func runBuild(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qyn build <file.ydv> [-o <output>] [--emit-ir] [--emit-obj] [--no-cache]")
		return 1
	}

	var filename string
	var outFile string
	emitIR := false
	emitObj := false
	noCache := false
	selfCheck := false
	emitMctp := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-o":
			if i+1 < len(args) {
				outFile = args[i+1]
				i++
			}
		case "--emit-ir":
			emitIR = true
		case "--emit-obj":
			emitObj = true
		case "--no-cache":
			noCache = true
		case "--selfcheck":
			selfCheck = true
		case "--emit-mctp":
			emitMctp = true
		default:
			if !strings.HasPrefix(arg, "-") && filename == "" {
				filename = arg
			}
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "qyn build: error: no input file specified")
		return 1
	}

	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: error reading %s: %v\n", filename, err)
		return 1
	}
	src := string(srcBytes)

	// Step 1: Lexical analysis
	l := lexer.New(src, filename)

	// Step 2: Parsing
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "qyn build: syntax errors in %s:\n", filename)
		for _, parseErr := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", parseErr.Error())
		}
		return 1
	}

	// Step 3: Type checking
	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "qyn build: type errors in %s:\n", filename)
		for _, typeErr := range typeErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", typeErr.Error())
		}
		return 1
	}

	// Step 4: LLVM IR Generation
	g := codegen.New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: codegen error: %v\n", err)
		return 1
	}

	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	llPath := base + ".ll"
	if err := os.WriteFile(llPath, []byte(ir), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: error writing IR file %s: %v\n", llPath, err)
		return 1
	}

	if emitIR {
		fmt.Print(ir)
	}
	fmt.Printf("Generated LLVM IR: %s\n", llPath)

	tc, err := llvm.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: LLVM toolchain not available (%v)\n", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if emitObj {
		objPath := base + ".o"
		if outFile != "" {
			objPath = outFile
		}
		if err := tc.LLToObj(ctx, llPath, objPath); err != nil {
			fmt.Fprintf(os.Stderr, "qyn build: error creating object file: %v\n", err)
			return 1
		}
		fmt.Printf("Compiled object file: %s\n", objPath)
		return 0
	}

	cacheDir := ""
	if noCache {
		tmpDir, err := os.MkdirTemp("", "qadav_nocache_*")
		if err == nil {
			cacheDir = tmpDir
			defer func() { _ = os.RemoveAll(tmpDir) }()
		}
	}

	rtObjs, err := runtime.Build(tc, cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: error building native runtime: %v\n", err)
		return 1
	}

	if outFile == "" {
		outFile = base + ".exe"
	}

	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], outFile, rtObjs[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "qyn build: compilation error: %v\n", err)
		return 1
	}

	if selfCheck {
		if err := selfcheck.Embed(outFile, 4096); err != nil {
			fmt.Fprintf(os.Stderr, "qyn build: error generating selfcheck sidecar: %v\n", err)
			return 1
		}
		fmt.Printf("Generated self-check sidecar: %s\n", selfcheck.SidecarPath(outFile))
	}

	if emitMctp {
		mctpPath := base + ".mctp"
		if data, err := os.ReadFile("model.mctp"); err == nil {
			_ = os.WriteFile(mctpPath, data, 0644)
		}
		fmt.Printf("Generated training provenance certificate: %s\n", mctpPath)
	}

	fmt.Printf("Compiled binary: %s\n", outFile)
	return 0
}
