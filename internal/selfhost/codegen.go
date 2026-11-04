package selfhost

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/codegen"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

// CodegenSourceLegacy is the Go reference bridge (internal/codegen) for fallback
// and comparative testing against the native self-hosted codegen.
func CodegenSourceLegacy(src, filename string) (string, int) {
	l := lexer.New(src, filename)
	p := parser.New(l, filename)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Sprintf("syntax error in %s\n", filename), 1
	}

	c := checker.New()
	typeErrors := c.Check(prog)
	if len(typeErrors) > 0 {
		return fmt.Sprintf("type error in %s\n", filename), 1
	}

	g := codegen.New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		return fmt.Sprintf("codegen error: %v\n", err), 1
	}

	return ir, 0
}

// CodegenSourceNative executes the compiled native self-hosted codegen binary
// (selfhost/codegen.exe) to generate LLVM IR directly without invoking the Go backend.
func CodegenSourceNative(src, filename string) (string, int) {
	// Locate selfhost/codegen.exe
	candidates := []string{
		filepath.Join("selfhost", "codegen.exe"),
		filepath.Join("..", "selfhost", "codegen.exe"),
		filepath.Join("..", "..", "selfhost", "codegen.exe"),
		"codegen.exe",
	}

	var exePath string
	for _, cand := range candidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			exePath = cand
			break
		}
	}

	if exePath == "" {
		return "qyn selfhost-codegen: native binary selfhost/codegen.exe not found (run 'qyn build selfhost/codegen.ydv -o selfhost/codegen.exe' first)\n", 1
	}

	// Prepare target input file
	targetFile := filename
	var cleanupTmp bool
	if _, err := os.Stat(targetFile); err != nil {
		// In-memory source provided without physical file; create temp file
		tmp, err := os.CreateTemp("", "qadav_sh_*.ydv")
		if err != nil {
			return fmt.Sprintf("failed to create temporary file for codegen: %v\n", err), 1
		}
		targetFile = tmp.Name()
		cleanupTmp = true
		if _, err := tmp.WriteString(src); err != nil {
			tmp.Close()
			os.Remove(targetFile)
			return fmt.Sprintf("failed to write source to temporary file: %v\n", err), 1
		}
		tmp.Close()
	}

	if cleanupTmp {
		defer os.Remove(targetFile)
	}

	cmd := exec.Command(exePath, targetFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stderr.String(), exitErr.ExitCode()
		}
		return fmt.Sprintf("failed to execute %s: %v\n", exePath, err), 1
	}

	return stdout.String(), 0
}

// CodegenSource compiles source into LLVM IR using the native self-hosted codegen driver.
func CodegenSource(src, filename string) (string, int) {
	return CodegenSourceNative(src, filename)
}
