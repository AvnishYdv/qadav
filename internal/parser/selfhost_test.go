package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/selfhost"
)

func TestSelfHostedParserByteIdentical(t *testing.T) {
	testDir := filepath.Join("..", "..", "selfhost", "tests")
	files, err := filepath.Glob(filepath.Join(testDir, "*.ydv"))
	if err != nil || len(files) == 0 {
		t.Fatalf("failed to find selfhost test files: %v", err)
	}

	for _, file := range files {
		srcBytes, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		src := string(srcBytes)

		// 1. Reference Go Parser S-Expr
		l := lexer.New(src, file)
		p := parser.New(l, file)
		prog := p.Parse()
		if len(p.Errors()) > 0 {
			t.Fatalf("reference parser failed for %s: %v", file, p.Errors()[0])
		}
		goSExpr := ast.ToSExpr(prog)

		// 2. Self-Hosted Parser S-Expr
		shSExpr, err := selfhost.ParseToSExpr(src, file)
		if err != nil {
			t.Fatalf("selfhost parser failed for %s: %v", file, err)
		}

		if goSExpr != shSExpr {
			t.Fatalf("byte mismatch in AST for %s:\nGo:  %s\nSH:  %s", file, goSExpr, shSExpr)
		}
	}
}

func TestParserErrorRecovery(t *testing.T) {
	invalidInputs := []string{
		"process ⟦ ⟧", // missing name and signature
		"process broken⦅⦆ → int64 ⟦ emit · ⟧", // missing expression in emit
		"rsca_attest stage", // incomplete RSCA statement
	}

	for _, inp := range invalidInputs {
		l := lexer.New(inp, "invalid.ydv")
		p := parser.New(l, "invalid.ydv")
		prog := p.Parse()
		_ = prog
		if len(p.Errors()) == 0 {
			t.Fatalf("expected parser errors for %q, got 0", inp)
		}
	}
}
