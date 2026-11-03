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

func TestQuantumIRGeneration(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    qubit q0 ·
    superpose q0 ·
    int64 r ↦ collapse q0 ·
    emit r ·
    yield 0 ·
⟧`
	l := lexer.New(src, "quantum.ydv")
	p := parser.New(l, "quantum.ydv")
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

	if !strings.Contains(ir, "@qadav_get_qstate") {
		t.Errorf("expected IR to declare/call @qadav_get_qstate")
	}
	if !strings.Contains(ir, "@qadav_gate_h") {
		t.Errorf("expected IR to call @qadav_gate_h")
	}
	if !strings.Contains(ir, "@qadav_measure") {
		t.Errorf("expected IR to call @qadav_measure")
	}
}

func TestQuantumHelloExecution(t *testing.T) {
	t.Parallel()
	srcBytes, err := os.ReadFile(filepath.Join("..", "..", "examples", "quantum_hello.ydv"))
	if err != nil {
		t.Fatalf("reading quantum_hello.ydv: %v", err)
	}
	out := compileAndRunSource(t, string(srcBytes))
	if out != "0" && out != "1" {
		t.Fatalf("expected 0 or 1, got %q", out)
	}
}

func TestBellStateExecution(t *testing.T) {
	t.Parallel()
	srcBytes, err := os.ReadFile(filepath.Join("..", "..", "examples", "bell_state.ydv"))
	if err != nil {
		t.Fatalf("reading bell_state.ydv: %v", err)
	}
	exePath := compileSourceToExe(t, string(srcBytes))
	for i := 0; i < 5; i++ {
		out := runExePath(t, exePath)
		if out != "0" {
			t.Fatalf("expected 0 for entangled bell state, got %q", out)
		}
	}
}
