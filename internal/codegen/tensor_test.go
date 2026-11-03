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

func compileTensorSourceToIR(src string) (string, error) {
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

func TestGenTensorLit1D(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_create") {
		t.Errorf("expected @qadav_tensor_create call in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_set_f64") {
		t.Errorf("expected @qadav_tensor_set_f64 call in IR")
	}
	if !strings.Contains(ir, "@.shape.") {
		t.Errorf("expected @.shape. constant in IR")
	}
}

func TestGenTensorLit2D(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬2⸴ 2⟭ m ↦ tensor⟬⟬1⸴ 2⟭⸴ ⟬3⸴ 4⟭⟭ ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_create") {
		t.Errorf("expected @qadav_tensor_create call in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_set_i64") {
		t.Errorf("expected @qadav_tensor_set_i64 call in IR")
	}
}

func TestGenTensorAdd(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⟭ a ↦ tensor⟬1.0⸴ 2.0⟭ ·
		tensor⟨float64⟩⟬2⟭ b ↦ tensor⟬3.0⸴ 4.0⟭ ·
		tensor⟨float64⟩⟬2⟭ c ↦ a ⊕ b ·
		tensor⟨float64⟩⟬2⟭ d ↦ a ⊕ 5.0 ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_add(") {
		t.Errorf("expected @qadav_tensor_add call in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_add_scalar_f64(") {
		t.Errorf("expected @qadav_tensor_add_scalar_f64 call in IR")
	}
}

func TestGenTensorMul(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⟭ a ↦ tensor⟬1.0⸴ 2.0⟭ ·
		tensor⟨float64⟩⟬2⟭ b ↦ a ⊗ 2.0 ·
		tensor⟨float64⟩⟬2⟭ c ↦ a ⊗ a ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_mul(") {
		t.Errorf("expected @qadav_tensor_mul call in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_mul_scalar_f64(") {
		t.Errorf("expected @qadav_tensor_mul_scalar_f64 call in IR")
	}
}

func TestGenMatmul(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1⸴ 2⟭⸴ ⟬3⸴ 4⟭⟭ ·
		tensor⟨int64⟩⟬2⸴ 2⟭ b ↦ tensor⟬⟬5⸴ 6⟭⸴ ⟬7⸴ 8⟭⟭ ·
		tensor⟨int64⟩⟬2⸴ 2⟭ c ↦ a ⊛ b ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_matmul(") {
		t.Errorf("expected @qadav_tensor_matmul call in IR")
	}
}

func TestGenSum(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ a ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		float64 s ↦ ∑ a ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_reduce_sum_f64(") {
		t.Errorf("expected @qadav_tensor_reduce_sum_f64 call in IR")
	}
}

func TestGenProduct(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬3⟭ a ↦ tensor⟬2⸴ 3⸴ 4⟭ ·
		int64 p ↦ ∏ a ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_reduce_prod_i64(") {
		t.Errorf("expected @qadav_tensor_reduce_prod_i64 call in IR")
	}
}

func TestGenTranspose(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⸴ 3⟭ a ↦ tensor⟬⟬1.0⸴ 2.0⸴ 3.0⟭⸴ ⟬4.0⸴ 5.0⸴ 6.0⟭⟭ ·
		tensor⟨float64⟩⟬3⸴ 2⟭ b ↦ ⋆ a ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_transpose(") {
		t.Errorf("expected @qadav_tensor_transpose call in IR")
	}
}

func TestGenEmitTensor(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⟭ f ↦ tensor⟬1.0⸴ 2.0⟭ ·
		tensor⟨int64⟩⟬2⟭ i ↦ tensor⟬3⸴ 4⟭ ·
		emit f ·
		emit i ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_print_f64(") {
		t.Errorf("expected @qadav_tensor_print_f64 call in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_print_i64(") {
		t.Errorf("expected @qadav_tensor_print_i64 call in IR")
	}
}

func TestGenTensorBind(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		t ↦ tensor⟬10⸴ 20⟭ ·
		emit t ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_create") {
		t.Errorf("expected tensor creation for inferred bind")
	}
}

func TestValidateIR_TensorPrograms(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ a ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		tensor⟨float64⟩⟬3⟭ b ↦ tensor⟬4.0⸴ 5.0⸴ 6.0⟭ ·
		tensor⟨float64⟩⟬3⟭ c ↦ a ⊕ b ·
		float64 s ↦ ∑ c ·
		emit s ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := tc.ValidateIR(ctx, ir); err != nil {
		t.Fatalf("ValidateIR failed: %v\nIR:\n%s", err, ir)
	}
}

func TestCompileAndRun_TensorCreate(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ v ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		emit v ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}
	exePath := filepath.Join(tmpDir, "tensor_create_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
	out, err := exec.CommandContext(ctx, exePath).CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v, output: %s", err, string(out))
	}
	got := string(out)
	if !strings.Contains(got, "tensor⟨float64⟩⟬3⟭:") {
		t.Errorf("expected tensor header in output, got: %q", got)
	}
	if !strings.Contains(got, "1.000000") || !strings.Contains(got, "2.000000") || !strings.Contains(got, "3.000000") {
		t.Errorf("expected elements in output, got: %q", got)
	}
}

func TestCompileAndRun_DotProduct(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process dot⦅tensor⟨float64⟩⟬3⟭ a⸴ tensor⟨float64⟩⟬3⟭ b⦆ → float64 ⟦
		yield ∑ ⦅a ⊗ b⦆ ·
	⟧

	process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ x ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		tensor⟨float64⟩⟬3⟭ y ↦ tensor⟬4.0⸴ 5.0⸴ 6.0⟭ ·
		float64 result ↦ dot⦅x⸴ y⦆ ·
		emit result ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}
	exePath := filepath.Join(tmpDir, "dot_prod_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
	out, err := exec.CommandContext(ctx, exePath).CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v, output: %s", err, string(out))
	}
	got := strings.TrimSpace(string(out))
	if !strings.Contains(got, "32.000000") {
		t.Errorf("expected dot product output to contain 32.000000, got: %q", got)
	}
}

func TestCompileAndRun_Matmul(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬2⸴ 3⟭ a ↦ tensor⟬⟬1⸴ 2⸴ 3⟭⸴ ⟬4⸴ 5⸴ 6⟭⟭ ·
		tensor⟨int64⟩⟬3⸴ 2⟭ b ↦ tensor⟬⟬7⸴ 8⟭⸴ ⟬9⸴ 10⟭⸴ ⟬11⸴ 12⟭⟭ ·
		tensor⟨int64⟩⟬2⸴ 2⟭ c ↦ a ⊛ b ·
		emit c ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}
	exePath := filepath.Join(tmpDir, "matmul_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
	out, err := exec.CommandContext(ctx, exePath).CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v, output: %s", err, string(out))
	}
	got := string(out)
	if !strings.Contains(got, "tensor⟨int64⟩⟬2⸴ 2⟭:") {
		t.Errorf("expected tensor header in output, got: %q", got)
	}
	if !strings.Contains(got, "58") || !strings.Contains(got, "64") || !strings.Contains(got, "139") || !strings.Contains(got, "154") {
		t.Errorf("expected 58, 64, 139, 154 in output, got: %q", got)
	}
}

func TestCompileAndRun_TensorAdd(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⟭ a ↦ tensor⟬1.5⸴ 2.5⟭ ·
		tensor⟨float64⟩⟬2⟭ b ↦ tensor⟬0.5⸴ 1.5⟭ ·
		tensor⟨float64⟩⟬2⟭ c ↦ a ⊕ b ·
		emit c ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}
	exePath := filepath.Join(tmpDir, "tadd_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
	out, err := exec.CommandContext(ctx, exePath).CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v, output: %s", err, string(out))
	}
	got := string(out)
	if !strings.Contains(got, "2.000000") || !strings.Contains(got, "4.000000") {
		t.Errorf("expected 2.000000 and 4.000000 in output, got: %q", got)
	}
}

func TestCompileAndRun_Transpose(t *testing.T) {
	t.Parallel()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬2⸴ 3⟭ a ↦ tensor⟬⟬1⸴ 2⸴ 3⟭⸴ ⟬4⸴ 5⸴ 6⟭⟭ ·
		tensor⟨int64⟩⟬3⸴ 2⟭ b ↦ ⋆ a ·
		emit b ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	tmpDir := t.TempDir()
	rtObjs, err := testutil.RuntimeObjects(t)
	if err != nil {
		t.Fatalf("runtime build failed: %v", err)
	}
	exePath := filepath.Join(tmpDir, "transpose_test.exe")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := tc.CompileWithRuntime(ctx, ir, rtObjs[0], exePath, rtObjs[1:]...); err != nil {
		t.Fatalf("compilation failed: %v", err)
	}
	out, err := exec.CommandContext(ctx, exePath).CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v, output: %s", err, string(out))
	}
	got := string(out)
	if !strings.Contains(got, "tensor⟨int64⟩⟬3⸴ 2⟭:") {
		t.Errorf("expected tensor⟨int64⟩⟬3⸴ 2⟭: in output, got: %q", got)
	}
}

func TestGenTensorFullCoverage(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨int64⟩⟬2⟭ a ↦ tensor⟬10⸴ 20⟭ ·
		tensor⟨int64⟩⟬2⟭ b ↦ tensor⟬2⸴ 5⟭ ·
		tensor⟨int64⟩⟬2⟭ c1 ↦ a ⊕ 5 ·
		tensor⟨int64⟩⟬2⟭ c2 ↦ a ⊖ 5 ·
		tensor⟨int64⟩⟬2⟭ c3 ↦ a ⊗ 5 ·
		tensor⟨int64⟩⟬2⟭ c4 ↦ a ⊘ 5 ·
		tensor⟨int64⟩⟬2⟭ c5 ↦ 5 ⊕ a ·
		tensor⟨int64⟩⟬2⟭ c6 ↦ 5 ⊗ a ·
		tensor⟨float64⟩⟬2⟭ f ↦ tensor⟬10.0⸴ 20.0⟭ ·
		tensor⟨float64⟩⟬2⟭ f1 ↦ f ⊖ 2.0 ·
		tensor⟨float64⟩⟬2⟭ f2 ↦ f ⊘ 2.0 ·
		tensor⟨float64⟩⟬2⟭ f3 ↦ 2.0 ⊕ f ·
		tensor⟨float64⟩⟬2⟭ f4 ↦ 2.0 ⊗ f ·
		int64 dot_int ↦ a ⊛ b ·
		int64 sum_int ↦ ∑ a ·
		float64 prod_f ↦ ∏ f ·
		tensor⟨int64⟩⟬2⟭ neg_int ↦ ⊖ a ·
		tensor⟨float64⟩⟬2⟭ neg_f ↦ ⊖ f ·
		emit dot_int ·
		emit sum_int ·
		emit prod_f ·
		yield 0 ·
	⟧`
	ir, err := compileTensorSourceToIR(src)
	if err != nil {
		t.Fatalf("codegen failed: %v", err)
	}
	if !strings.Contains(ir, "@qadav_tensor_add_scalar_i64") {
		t.Errorf("expected @qadav_tensor_add_scalar_i64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_sub_scalar_i64") {
		t.Errorf("expected @qadav_tensor_sub_scalar_i64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_mul_scalar_i64") {
		t.Errorf("expected @qadav_tensor_mul_scalar_i64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_div_scalar_i64") {
		t.Errorf("expected @qadav_tensor_div_scalar_i64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_sub_scalar_f64") {
		t.Errorf("expected @qadav_tensor_sub_scalar_f64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_div_scalar_f64") {
		t.Errorf("expected @qadav_tensor_div_scalar_f64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_reduce_sum_i64") {
		t.Errorf("expected @qadav_tensor_reduce_sum_i64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_reduce_prod_f64") {
		t.Errorf("expected @qadav_tensor_reduce_prod_f64 in IR")
	}
	if !strings.Contains(ir, "@qadav_tensor_get_i64") {
		t.Errorf("expected @qadav_tensor_get_i64 for dot product in IR")
	}
}

func TestGenStringEscapes(t *testing.T) {
	g := New(nil)
	name, length := g.addStringConstant("hello\n\r\t\"\\world\x01")
	if name == "" || length == 0 {
		t.Errorf("expected valid string constant")
	}
	if len(g.stringConstants) != 1 {
		t.Fatalf("expected 1 string constant")
	}
	sc := g.stringConstants[0]
	if !strings.Contains(sc.escaped, `\0A`) || !strings.Contains(sc.escaped, `\0D`) || !strings.Contains(sc.escaped, `\09`) || !strings.Contains(sc.escaped, `\22`) || !strings.Contains(sc.escaped, `\5C`) {
		t.Errorf("expected escape sequences in constant: %q", sc.escaped)
	}
}

func BenchmarkTensorCodegen(b *testing.B) {
	src := `process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬2⸴ 2⟭ a ↦ tensor⟬⟬1.0⸴ 2.0⟭⸴ ⟬3.0⸴ 4.0⟭⟭ ·
		tensor⟨float64⟩⟬2⸴ 2⟭ b ↦ tensor⟬⟬5.0⸴ 6.0⟭⸴ ⟬7.0⸴ 8.0⟭⟭ ·
		tensor⟨float64⟩⟬2⸴ 2⟭ c ↦ a ⊛ b ·
		emit c ·
		yield 0 ·
	⟧`
	l := lexer.New(src, "bench.ydv")
	p := parser.New(l, "bench.ydv")
	prog := p.Parse()
	c := checker.New()
	_ = c.Check(prog)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := New(c)
		_, _ = g.Generate(prog)
	}
}
