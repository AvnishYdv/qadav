package testutil

import (
    "os"
    "os/exec"
    "testing"
    "path/filepath"
)

// ValidateLLVMIR writes the provided LLVM IR to a temporary file and attempts to compile it with clang.
// The test fails if clang returns an error.
func ValidateLLVMIR(t *testing.T, ir string) {
    t.Helper()
    // Create temporary directory
    dir := t.TempDir()
    // Write IR to a file with .ll extension
    irPath := filepath.Join(dir, "tmp.ll")
    if err := os.WriteFile(irPath, []byte(ir), 0o600); err != nil {
        t.Fatalf("failed to write temporary LLVM IR file: %v", err)
    }
    // Run clang to check syntax (compile to object without output)
    cmd := exec.Command("clang", "-x", "ir", "-c", "-o", os.DevNull, irPath)
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("clang failed to validate LLVM IR: %v\noutput: %s", err, string(out))
    }
}

// SkipIfNoLLVM skips the test if LLVM (clang or llc) is not found in PATH.
func SkipIfNoLLVM(t *testing.T) {
    t.Helper()
    if _, err := exec.LookPath("clang"); err != nil {
        t.Skip("clang not found in PATH – skipping LLVM‑dependent test")
    }
    if _, err := exec.LookPath("llc"); err != nil {
        t.Skip("llc not found in PATH – skipping LLVM‑dependent test")
    }
}

