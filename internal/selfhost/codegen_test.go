package selfhost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelfHostedCodegenNativeHelloByteIdentical(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "hello.ydv")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading %s: %v", path, err)
	}

	// 1. Generate LLVM via Go reference driver
	irLegacy, codeLegacy := CodegenSourceLegacy(string(src), path)
	if codeLegacy != 0 {
		t.Fatalf("Go legacy codegen failed on hello.ydv: %s", irLegacy)
	}

	// 2. Generate LLVM via compiled native self-hosted binary (selfhost/codegen.exe)
	irNative, codeNative := CodegenSourceNative(string(src), path)
	if codeNative != 0 {
		t.Fatalf("Native self-hosted codegen failed on hello.ydv: %s", irNative)
	}

	// 3. Byte-identical verification (0 differences)
	if irLegacy != irNative {
		t.Fatalf("Byte mismatch between Go reference codegen and Native self-hosted codegen!\nLegacy len: %d, Native len: %d", len(irLegacy), len(irNative))
	}
}

func TestSelfHostedCodegenNativeFibByteIdentical(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "fib.ydv")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading %s: %v", path, err)
	}

	irLegacy, codeLegacy := CodegenSourceLegacy(string(src), path)
	if codeLegacy != 0 {
		t.Fatalf("Go legacy codegen failed on fib.ydv: %s", irLegacy)
	}

	irNative, codeNative := CodegenSourceNative(string(src), path)
	if codeNative != 0 {
		t.Fatalf("Native self-hosted codegen failed on fib.ydv: %s", irNative)
	}

	if irLegacy != irNative {
		t.Fatalf("Byte mismatch on fib.ydv! Legacy len: %d, Native len: %d", len(irLegacy), len(irNative))
	}
}

func TestSelfHostedCodegenNativeDotProductByteIdentical(t *testing.T) {
	path := filepath.Join("..", "..", "examples", "dot_product.ydv")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading %s: %v", path, err)
	}

	irLegacy, codeLegacy := CodegenSourceLegacy(string(src), path)
	if codeLegacy != 0 {
		t.Fatalf("Go legacy codegen failed on dot_product.ydv: %s", irLegacy)
	}

	irNative, codeNative := CodegenSourceNative(string(src), path)
	if codeNative != 0 {
		t.Fatalf("Native self-hosted codegen failed on dot_product.ydv: %s", irNative)
	}

	if irLegacy != irNative {
		t.Fatalf("Byte mismatch on dot_product.ydv! Legacy len: %d, Native len: %d", len(irLegacy), len(irNative))
	}
}

func TestSelfHostedCodegenLegacyExamples(t *testing.T) {
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

			ir, code := CodegenSourceLegacy(string(src), path)
			if code != 0 {
				t.Fatalf("Go reference codegen failed on %s: %s", name, ir)
			}
			if len(ir) == 0 {
				t.Fatalf("expected non-empty LLVM IR for %s", name)
			}
		})
	}
}

func TestSelfHostedCodegenLegacyArithmetic(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    int64 a ↦ 10 ·
    int64 b ↦ 20 ·
    int64 c ↦ a ⊕ b ·
    int64 d ↦ c ⊗ 2 ·
    emit d ·
    yield 0 ·
⟧`
	ir, code := CodegenSourceLegacy(src, "arith.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "add i64") || !strings.Contains(ir, "mul i64") {
		t.Fatalf("expected add and mul instructions in IR:\n%s", ir)
	}
}

func TestSelfHostedCodegenLegacyFunction(t *testing.T) {
	src := `process add⦅int64 x⸴ int64 y⦆ → int64 ⟦
    yield x ⊕ y ·
⟧
process main⦅⦆ → int64 ⟦
    int64 res ↦ add⦅5⸴ 7⦆ ·
    emit res ·
    yield 0 ·
⟧`
	ir, code := CodegenSourceLegacy(src, "fn.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "define i64 @\"add\"") || !strings.Contains(ir, "define i32 @\"main\"") {
		t.Fatalf("expected add and main function definitions in IR:\n%s", ir)
	}
}

func TestSelfHostedCodegenLegacyEmit(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    emit 12345 ·
    emit "test_emit_string" ·
    yield 0 ·
⟧`
	ir, code := CodegenSourceLegacy(src, "emit.ydv")
	if code != 0 {
		t.Fatalf("codegen error: %s", ir)
	}
	if !strings.Contains(ir, "@qadav_emit_int") || !strings.Contains(ir, "@qadav_emit_text") {
		t.Fatalf("expected emit runtime calls in IR:\n%s", ir)
	}
}
