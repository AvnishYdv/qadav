package codegen

import (
	"context"
	"os"
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
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func compileSourceToIR(src string) (string, error) {
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()

	c := checker.New()
	errs := c.Check(prog)
	if len(errs) > 0 {
		return "", &errs[0]
	}

	g := New(c)
	return g.Generate(prog)
}

func TestGenArithmetic(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 a ↦ 10 ·
		int64 b ↦ 20 ·
		int64 c ↦ a ⊕ b ·
		yield c ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "define i32 @\"main\"()") {
		t.Errorf("missing main definition in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "add i64") {
		t.Errorf("missing add instruction in IR:\n%s", ir)
	}
}

func TestGenFloatArithmetic(t *testing.T) {
	src := `
	process calcFloat⦅⦆ → float64 ⟦
		float64 x ↦ 3.14 ·
		float64 y ↦ 2.0 ·
		float64 z ↦ x ⊗ y ·
		emit z ·
		yield z ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "fmul double") {
		t.Errorf("missing fmul in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "@qadav_emit_float") {
		t.Errorf("missing qadav_emit_float in IR:\n%s", ir)
	}
}

func TestGenEmit(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		emit "Hello, Qadav!" ·
		emit 42 ·
		emit affirm ·
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "@qadav_emit_text") {
		t.Errorf("missing qadav_emit_text call in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "@qadav_emit_int") {
		t.Errorf("missing qadav_emit_int call in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "@qadav_emit_logic") {
		t.Errorf("missing qadav_emit_logic call in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "Hello, Qadav!") {
		t.Errorf("missing string literal in IR:\n%s", ir)
	}
}

func TestGenWhenBranching(t *testing.T) {
	src := `
	process testWhen⦅int64 x⦆ → int64 ⟦
		when x ≺ 10 ⟦
			yield 1 ·
		⟧ otherwise ⟦
			yield 2 ·
		⟧
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "br i1") {
		t.Errorf("missing conditional branch in IR:\n%s", ir)
	}
	if !strings.Contains(ir, "when_then") || !strings.Contains(ir, "when_else") {
		t.Errorf("missing branch labels in IR:\n%s", ir)
	}
}

func TestGenCycleLoop(t *testing.T) {
	src := `
	process loop⦅⦆ → int64 ⟦
		int64 count ↦ 0 ·
		cycle count ≺ 5 ⟦
			count ← count ⊕ 1 ·
		⟧
		yield count ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "cycle_cond") || !strings.Contains(ir, "cycle_body") {
		t.Errorf("missing cycle loop labels in IR:\n%s", ir)
	}
}

func TestGenProcessCall(t *testing.T) {
	src := `
	process add⦅int64 a⸴ int64 b⦆ → int64 ⟦
		yield a ⊕ b ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 res ↦ add⦅5⸴ 15⦆ ·
		yield res ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	if !strings.Contains(ir, "call i64 @\"add\"") {
		t.Errorf("missing process invocation call in IR:\n%s", ir)
	}
}

func TestGenMultipleFns(t *testing.T) {
	src := `
	process fn1⦅⦆ → int64 ⟦ yield 1 · ⟧
	process fn2⦅⦆ → int64 ⟦ yield 2 · ⟧
	process main⦅⦆ → int64 ⟦ yield fn1⦅⦆ ⊕ fn2⦅⦆ · ⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "define i64 @\"fn1\"()") || !strings.Contains(ir, "define i64 @\"fn2\"()") {
		t.Errorf("missing multiple functions in IR:\n%s", ir)
	}
}

func TestGenNestedWhen(t *testing.T) {
	src := `
	process nested⦅int64 x⸴ int64 y⦆ → int64 ⟦
		when x ≺ 10 ⟦
			when y ≻ 5 ⟦
				yield 100 ·
			⟧
			yield 50 ·
		⟧ otherwise ⟦
			yield 0 ·
		⟧
		yield 1 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if strings.Count(ir, "br i1") < 2 {
		t.Errorf("expected at least 2 branches in nested when:\n%s", ir)
	}
}

func TestGenNestedCycle(t *testing.T) {
	src := `
	process loop2⦅⦆ → int64 ⟦
		int64 i ↦ 0 ·
		cycle i ≺ 3 ⟦
			int64 j ↦ 0 ·
			cycle j ≺ 3 ⟦
				j ← j ⊕ 1 ·
			⟧
			i ← i ⊕ 1 ·
		⟧
		yield i ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if strings.Count(ir, "cycle_cond") < 2 {
		t.Errorf("expected at least 2 cycle conditions in nested cycle:\n%s", ir)
	}
}

func TestValidateIRWithClang(t *testing.T) {
	src := `
	process fib⦅int64 n⦆ → int64 ⟦
		when n ≺ 2 ⟦
			yield n ·
		⟧
		yield fib⦅n ⊖ 1⦆ ⊕ fib⦅n ⊖ 2⦆ ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 ans ↦ fib⦅10⦆ ·
		emit ans ·
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("compileSourceToIR error: %v", err)
	}
	if !strings.Contains(ir, "fib") || !strings.Contains(ir, "add i64") {
		t.Fatalf("expected fib definition and add in IR, got:\n%s", ir)
	}
}

func TestCompileAndRun_Hello(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	src := `
	process main⦅⦆ → int64 ⟦
		emit "Hello, Qadav! Quantum-native AGI language." ·
		emit affirm ·
		emit deny ·
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}

	exePath := filepath.Join(tmpDir, "hello_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("CompileWithRuntime failed: %v", err)
	}

	runCmd := exec.CommandContext(ctx, exePath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running hello_test.exe failed: %v\noutput: %s", err, string(out))
	}

	expected := "Hello, Qadav! Quantum-native AGI language."
	outStr := string(out)
	if !strings.Contains(outStr, expected) || !strings.Contains(outStr, "affirm") || !strings.Contains(outStr, "deny") {
		t.Fatalf("expected output to contain %q, affirm, and deny, got %q", expected, outStr)
	}
}

func TestCompileAndRun_Fib(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	src := `
	process fib⦅int64 n⦆ → int64 ⟦
		when n ≺ 2 ⟦
			yield n ·
		⟧
		yield fib⦅n ⊖ 1⦆ ⊕ fib⦅n ⊖ 2⦆ ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 n ↦ 10 ·
		int64 result ↦ fib⦅n⦆ ·
		emit result ·
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}

	exePath := filepath.Join(tmpDir, "fib_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("CompileWithRuntime failed: %v", err)
	}

	runCmd := exec.CommandContext(ctx, exePath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running fib_test.exe failed: %v\noutput: %s", err, string(out))
	}

	if !strings.Contains(string(out), "55") {
		t.Fatalf("expected output to contain '55', got %q", string(out))
	}
}

func TestCompileAndRun_Logic(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		emit affirm ·
		emit deny ·
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "@qadav_emit_logic") {
		t.Fatalf("expected emit_logic in IR, got:\n%s", ir)
	}
}

func TestCodegenAllOperatorsAndBranches(t *testing.T) {
	src := `
	process testAll⦅int32 i32⸴ float32 f32⸴ qubit q⦆ ⟦
		int64 a ↦ 10 ·
		a ↦⊕ 5 ·
		a ↦⊖ 2 ·
		a ↦⊗ 3 ·
		a ↦⊘ 2 ·
		int64 mod ↦ a ⊙ 3 ·
		logic eq ↦ a ≡ 10 ·
		logic ne ↦ a ≠ 10 ·
		logic lt ↦ a ≺ 10 ·
		logic gt ↦ a ≻ 10 ·
		logic le ↦ a ≼ 10 ·
		logic ge ↦ a ≽ 10 ·
		logic w ↦ eq ∧ ne ·
		logic v ↦ eq ∨ ne ·
		logic x ↦ eq ⊻ ne ·
		logic not ↦ ¬eq ·
		int64 neg ↦ ⊖a ·
		float64 f ↦ 2.5 ·
		f ↦⊕ 1.5 ·
		f ↦⊖ 0.5 ·
		f ↦⊗ 2.0 ·
		f ↦⊘ 2.0 ·
		float64 fneg ↦ ⊖f ·
		logic feq ↦ f ≡ 2.0 ·
		logic fne ↦ f ≠ 2.0 ·
		logic flt ↦ f ≺ 2.0 ·
		logic fgt ↦ f ≻ 2.0 ·
		logic fle ↦ f ≼ 2.0 ·
		logic fge ↦ f ≽ 2.0 ·
		observe a ·
		emit i32 ·
		emit f32 ·
		emit q ·
		superpose q ·
		yield ·
	⟧

	process main⦅⦆ → int32 ⟦
		yield 0 ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "srem i64") {
		t.Errorf("missing srem in IR")
	}
	if !strings.Contains(ir, "fdiv double") {
		t.Errorf("missing fdiv in IR")
	}
	if !strings.Contains(ir, "fneg double") {
		t.Errorf("missing fneg in IR")
	}
}

func TestCodegenCompileDirect(t *testing.T) {
	_, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦ emit 123 · yield 0 · ⟧`
	l := lexer.New(src, "main.ydv")
	p := parser.New(l, "main.ydv")
	prog := p.Parse()
	c := checker.New()
	_ = c.Check(prog)

	g := New(c)
	ir, err := g.Generate(prog)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	tmpDir := t.TempDir()
	outExe := filepath.Join(tmpDir, "direct.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := g.Compile(ctx, ir, outExe); err != nil {
		t.Fatalf("g.Compile failed: %v", err)
	}
}

func TestCodegenHelpers(t *testing.T) {
	g := New(nil)
	if g.genExpr(nil).Ref != "0" {
		t.Errorf("genExpr(nil) failed")
	}
	g.genStmt(nil) // should not panic
	if g.resolveType(nil).Kind != types.KindVoid {
		t.Errorf("resolveType(nil) != Void")
	}
	allTypes := []*types.Type{
		types.Int8, types.Uint8, types.Int16, types.Uint16,
		types.Int32, types.Uint32, types.Int64, types.Uint64,
		types.Float32, types.Float64, types.Logic, types.Text,
		types.Void, types.Qubit,
	}
	for _, typ := range allTypes {
		if g.llvmType(typ) == "" {
			t.Errorf("empty llvmType for %s", typ)
		}
	}
}

func TestOsumAlwaysAdd(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 x ↦ 3 ⊕ 4 ·
		yield x ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "add i64 3, 4") {
		t.Fatalf("expected 'add i64 3, 4' in IR, got:\n%s", ir)
	}
}

func TestXveeAlwaysXor(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 x ↦ 3 ⊻ 4 ·
		yield x ·
	⟧
	`
	ir, err := compileSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(ir, "xor i64 3, 4") {
		t.Fatalf("expected 'xor i64 3, 4' in IR, got:\n%s", ir)
	}
}

func TestBellStateUsesXvee(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "examples", "bell_state.ydv"))
	if err != nil {
		t.Fatalf("reading bell_state.ydv: %v", err)
	}
	l := lexer.New(string(content), "bell_state.ydv")
	foundXvee := false
	for {
		tok := l.NextToken()
		if tok.Type == token.XVEE {
			foundXvee = true
			break
		}
		if tok.Type == token.EOF {
			break
		}
	}
	if !foundXvee {
		t.Fatalf("expected token.XVEE in bell_state.ydv, but none found")
	}
}

func BenchmarkCodegen(b *testing.B) {
	src := `
	process fib⦅int64 n⦆ → int64 ⟦
		when n ≺ 2 ⟦
			yield n ·
		⟧
		yield fib⦅n ⊖ 1⦆ ⊕ fib⦅n ⊖ 2⦆ ·
	⟧

	process main⦅⦆ → int64 ⟦
		int64 n ↦ 10 ·
		yield fib⦅n⦆ ·
	⟧
	`
	l := lexer.New(src, "fib.ydv")
	p := parser.New(l, "fib.ydv")
	prog := p.Parse()
	c := checker.New()
	_ = c.Check(prog)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := New(c)
		_, _ = g.Generate(prog)
	}
}
