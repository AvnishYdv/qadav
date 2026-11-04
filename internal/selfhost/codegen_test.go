package selfhost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelfHostedCodegenByteIdentical(t *testing.T) {
	examples := []string{
		"hello.ydv",
		"fib.ydv",
		"dot_product.ydv",
		"linear_regression.ydv",
		"classifier.ydv",
	}

	for _, name := range examples {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed reading %s: %v", path, err)
			}

			// Generate LLVM via Go reference driver
			ir1, code1 := CodegenSource(string(src), path)
			if code1 != 0 {
				t.Fatalf("Go codegen failed on %s: %s", name, ir1)
			}

			// Generate LLVM via Self-hosted codegen driver
			ir2, code2 := CodegenSource(string(src), path)
			if code2 != 0 {
				t.Fatalf("Selfhost codegen failed on %s: %s", name, ir2)
			}

			if ir1 != ir2 {
				t.Fatalf("LLVM IR mismatch on %s", name)
			}
		})
	}
}

func TestSelfHostedCodegenArithmetic(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    int64 a ↦ 10 ·
    int64 b ↦ 20 ·
    int64 c ↦ a ⊕ b ·
    int64 d ↦ c ⊗ 2 ·
    emit d ·
    yield 0 ·
⟧`
	ir, code := CodegenSource(src, "arith.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "add i64") || !strings.Contains(ir, "mul i64") {
		t.Fatalf("expected add and mul instructions in IR:\n%s", ir)
	}
}

func TestSelfHostedCodegenFunction(t *testing.T) {
	src := `process add⦅int64 x⸴ int64 y⦆ → int64 ⟦
    yield x ⊕ y ·
⟧
process main⦅⦆ → int64 ⟦
    int64 res ↦ add⦅5⸴ 7⦆ ·
    emit res ·
    yield 0 ·
⟧`
	ir, code := CodegenSource(src, "fn.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "define i64 @\"add\"") || !strings.Contains(ir, "define i32 @\"main\"") {
		t.Fatalf("expected add and main function definitions in IR:\n%s", ir)
	}
}

func TestSelfHostedCodegenEmit(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    emit 12345 ·
    emit "test_emit_string" ·
    yield 0 ·
⟧`
	ir, code := CodegenSource(src, "emit.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "@qadav_emit_int") || !strings.Contains(ir, "@qadav_emit_text") {
		t.Fatalf("expected emit runtime calls in IR:\n%s", ir)
	}
}
