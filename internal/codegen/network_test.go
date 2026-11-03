package codegen

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/qadav-lang/qadav/internal/checker"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/llvm"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/testutil"
)

func compileSourceToExe(t *testing.T, src string) string {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain not available:", err)
	}

	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := checker.New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	g := New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}

	exePath := filepath.Join(tmpDir, "test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compile failed: %v\nIR:\n%s", err, ir)
	}
	return exePath
}

func runExePath(t *testing.T, exePath string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exePath)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\noutput: %s", err, string(outBytes))
	}
	return strings.TrimSpace(string(outBytes))
}

func compileAndRunSource(t *testing.T, src string) string {
	exePath := compileSourceToExe(t, src)
	return runExePath(t, exePath)
}

func TestGenNetworkIR(t *testing.T) {
	src := `
network Regressor ⟦
    dense ⟬1⸴ 1⟭ ·
⟧

process main⦅⦆ ⟦
    tensor⟨float64⟩⟬1⟭ x ↦ tensor⟬2.0⟭ ·
    tensor⟨float64⟩⟬1⟭ pred ↦ Regressor⦅x⦆ ·
    emit pred ·
⟧
`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := checker.New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	g := New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "@.net.Regressor") {
		t.Errorf("expected IR to contain @.net.Regressor, got:\n%s", ir)
	}
	if !strings.Contains(ir, "@qadav_network_create") {
		t.Errorf("expected IR to contain @qadav_network_create")
	}
	if !strings.Contains(ir, "@qadav_network_forward") {
		t.Errorf("expected IR to contain @qadav_network_forward")
	}
}

func TestGenLearnIR(t *testing.T) {
	src := `
network Model ⟦
    dense ⟬1⸴ 1⟭ ·
⟧

process main⦅⦆ ⟦
    tensor⟨float64⟩⟬2⸴ 1⟭ x ↦ tensor⟬⟬1.0⟭⸴ ⟬2.0⟭⟭ ·
    tensor⟨float64⟩⟬2⸴ 1⟭ y ↦ tensor⟬⟬2.0⟭⸴ ⟬4.0⟭⟭ ·
    learn Model on x ⸴ y ⟦
        loss mse ·
        optimizer sgd ·
        lr ↦ 0.01 ·
        epochs ↦ 10 ·
    ⟧
⟧
`
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := checker.New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		t.Fatalf("checker errors: %v", errs)
	}

	g := New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "@qadav_train_epochs") {
		t.Errorf("expected IR to contain @qadav_train_epochs, got:\n%s", ir)
	}
}

func TestCompileAndRun_NetworkForward(t *testing.T) {
	t.Parallel()
	src := `
network Net ⟦
    dense ⟬2⸴ 2⟭ ·
    relu ·
⟧

process main⦅⦆ ⟦
    tensor⟨float64⟩⟬2⟭ x ↦ tensor⟬1.0⸴ 2.0⟭ ·
    tensor⟨float64⟩⟬2⟭ out ↦ Net⦅x⦆ ·
    emit out ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "tensor⟨float64⟩") {
		t.Errorf("expected tensor output, got: %s", out)
	}
}

func TestCompileAndRun_LinearRegression(t *testing.T) {
	t.Parallel()
	src := `
network LineFitter ⟦
    dense ⟬1⸴ 1⟭ ·
⟧

process main⦅⦆ ⟦
    tensor⟨float64⟩⟬4⸴ 1⟭ x ↦ tensor⟬⟬1.0⟭⸴ ⟬2.0⟭⸴ ⟬3.0⟭⸴ ⟬4.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬2.0⟭⸴ ⟬4.0⟭⸴ ⟬6.0⟭⸴ ⟬8.0⟭⟭ ·
    learn LineFitter on x ⸴ y ⟦
        loss mse ·
        optimizer sgd ·
        lr ↦ 0.05 ·
        epochs ↦ 30 ·
    ⟧
    tensor⟨float64⟩⟬1⟭ test_x ↦ tensor⟬3.0⟭ ·
    tensor⟨float64⟩⟬1⟭ pred ↦ LineFitter⦅test_x⦆ ·
    emit pred ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "tensor⟨float64⟩") {
		t.Fatalf("expected tensor output, got: %s", out)
	}
	if !strings.Contains(out, "⟬") || !strings.Contains(out, "⟭") {
		t.Errorf("expected brackets in tensor output, got: %s", out)
	}
}

func TestCompileAndRun_XORNetwork(t *testing.T) {
	t.Parallel()
	src := `
network XORNet ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
    sigmoid ·
⟧

process main⦅⦆ ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⸴ ⟬1.0⸴ 0.0⟭⸴ ⟬1.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⸴ ⟬0.0⟭⟭ ·
    learn XORNet on x ⸴ y ⟦
        loss mse ·
        optimizer adam ·
        lr ↦ 0.1 ·
        epochs ↦ 20 ·
    ⟧
    tensor⟨float64⟩⟬2⟭ test_x ↦ tensor⟬0.0⸴ 1.0⟭ ·
    tensor⟨float64⟩⟬1⟭ pred ↦ XORNet⦅test_x⦆ ·
    emit pred ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "tensor⟨float64⟩") {
		t.Errorf("expected tensor output for XOR network, got: %s", out)
	}
}

func TestCompileAndRun_ProbabilisticTypes(t *testing.T) {
	t.Parallel()
	src := `
process main⦅⦆ ⟦
    confident⟨float64⟩ conf ↦ 0.98 ·
    explore⟨float64⟩ exp ↦ 0.45 ·
    emit conf ·
    emit exp ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "0.98") || !strings.Contains(out, "0.45") {
		t.Errorf("expected 0.98 and 0.45 in output, got: %s", out)
	}
}
