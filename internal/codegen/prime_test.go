package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
)

func TestPrimeSymbolicCheckerFunctions(t *testing.T) {
	if !checker.IsPrime(2) || !checker.IsPrime(3) || !checker.IsPrime(613) {
		t.Errorf("expected 2, 3, 613 to be prime")
	}
	if checker.IsPrime(1) || checker.IsPrime(4) || checker.IsPrime(30) {
		t.Errorf("expected 1, 4, 30 not to be prime")
	}

	factors := checker.Factorize(30)
	if len(factors) != 3 || factors[0] != 2 || factors[1] != 3 || factors[2] != 5 {
		t.Errorf("expected factors [2, 3, 5], got %v", factors)
	}

	enc := checker.EncodeConcept([]int64{2, 3, 5})
	if enc != 30 {
		t.Errorf("expected 30, got %d", enc)
	}

	dec := checker.DecodeConcept(613)
	if len(dec) != 1 || dec[0] != 613 {
		t.Errorf("expected [613], got %v", dec)
	}
}

func TestConceptIRGeneration(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    concept consciousness ↦ 2 ·
    concept genome ↦ 3 ·
    concept complex ↦ consciousness ⊗ genome ·
    emit complex ·
    yield 0 ·
⟧`
	l := lexer.New(src, "concept.ydv")
	p := parser.New(l, "concept.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	chk := checker.New()
	chkErrs := chk.Check(prog)
	if len(chkErrs) > 0 {
		t.Fatalf("checker errors: %v", chkErrs)
	}

	g := New(chk)
	ir, err := g.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "@qadav_print_concept") {
		t.Errorf("expected IR to call @qadav_print_concept")
	}
}

func TestConceptEncodingExecution(t *testing.T) {
	srcBytes, err := os.ReadFile(filepath.Join("..", "..", "examples", "concept_encoding.ydv"))
	if err != nil {
		t.Fatalf("reading concept_encoding.ydv: %v", err)
	}
	out := compileAndRunSource(t, string(srcBytes))
	if !strings.Contains(out, "30") {
		t.Fatalf("expected output to contain 30, got %q", out)
	}
}

func TestConceptDecomposeExecution(t *testing.T) {
	srcBytes, err := os.ReadFile(filepath.Join("..", "..", "examples", "concept_decompose.ydv"))
	if err != nil {
		t.Fatalf("reading concept_decompose.ydv: %v", err)
	}
	out := compileAndRunSource(t, string(srcBytes))
	if !strings.Contains(out, "613") {
		t.Fatalf("expected output to contain 613, got %q", out)
	}
}
