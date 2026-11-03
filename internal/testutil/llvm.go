package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/qadav-lang/qadav/internal/llvm"
)

// ValidateLLVMIR writes the provided LLVM IR to a temporary file and attempts
// to compile it with clang. The test fails if clang returns an error.
func ValidateLLVMIR(t *testing.T, ir string) {
	t.Helper()
	dir := t.TempDir()
	irPath := filepath.Join(dir, "tmp.ll")
	if err := os.WriteFile(irPath, []byte(ir), 0o600); err != nil {
		t.Fatalf("failed to write temporary LLVM IR file: %v", err)
	}
	clang := "clang"
	if tc, err := llvm.Detect(); err == nil && tc.ClangPath != "" {
		clang = tc.ClangPath
	}
	cmd := exec.Command(clang, "-x", "ir", "-c", "-o", os.DevNull, irPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clang failed to validate LLVM IR: %v\noutput: %s", err, string(out))
	}
}

// SkipIfNoLLVM skips the test if LLVM (clang) is not found in PATH.
func SkipIfNoLLVM(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		if _, errWin := exec.LookPath("clang.exe"); errWin != nil {
			t.Skip("clang not found in PATH – skipping LLVM‑dependent test")
		}
	}
}
