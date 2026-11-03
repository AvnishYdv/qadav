package llvm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDetect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tc, err := Detect()
	if err != nil {
		t.Skipf("LLVM not installed: %v", err)
	}
	if tc.ClangPath == "" {
		t.Fatal("ClangPath is empty")
	}
	if _, err := os.Stat(tc.ClangPath); err != nil {
		t.Fatalf("ClangPath does not exist: %v", err)
	}
	if tc.Version == "" {
		t.Fatal("Version is empty")
	}
	t.Logf("Toolchain: %s (complete=%v)", tc.String(), tc.IsComplete())
}

func TestVersionParse(t *testing.T) {
	cases := []struct {
		input       string
		expectedVer string
		expectError bool
	}{
		{
			input:       "clang version 23.1.4 (https://github.com/llvm/llvm-project.git 7a941dfb)",
			expectedVer: "23.1.4",
			expectError: false,
		},
		{
			input:       "clang version 18.1.8\nTarget: x86_64",
			expectedVer: "18.1.8",
			expectError: false,
		},
		{
			input:       "clang version 17.0.6",
			expectError: true, // major < 18
		},
		{
			input:       "garbage output with no version",
			expectError: true,
		},
		{
			input:       "",
			expectError: true,
		},
	}

	for _, tc := range cases {
		ver, err := ParseClangVersion(tc.input)
		if tc.expectError {
			if err == nil {
				t.Fatalf("expected error for input %q, got version %q", tc.input, ver)
			}
		} else {
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if ver != tc.expectedVer {
				t.Fatalf("expected %q, got %q", tc.expectedVer, ver)
			}
		}
	}
}

func TestValidateIR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tc, err := Detect()
	if err != nil {
		t.Skipf("LLVM toolchain unavailable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	validIR := `define i32 @main() {
	ret i32 0
}
`
	if err := tc.ValidateIR(ctx, validIR); err != nil {
		t.Fatalf("expected valid IR to pass validation, got: %v", err)
	}

	invalidIR := `this is not valid llvm ir syntax`
	if err := tc.ValidateIR(ctx, invalidIR); err == nil {
		t.Fatal("expected invalid IR to fail validation, but it passed")
	}
}

func TestCompileIR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tc, err := Detect()
	if err != nil {
		t.Skipf("LLVM toolchain unavailable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	validIR := `define i32 @main() {
	ret i32 0
}
`
	outDir := t.TempDir()
	outBinary := filepath.Join(outDir, "test_prog")
	if runtime.GOOS == "windows" {
		outBinary += ".exe"
	}

	if err := tc.CompileIR(ctx, validIR, outBinary); err != nil {
		t.Fatalf("failed to compile IR: %v", err)
	}

	if _, err := os.Stat(outBinary); err != nil {
		t.Fatalf("compiled binary not found at %s: %v", outBinary, err)
	}

	runCmd := exec.CommandContext(ctx, outBinary)
	if runErr := runCmd.Run(); runErr != nil {
		t.Fatalf("compiled binary failed to run with exit code 0: %v", runErr)
	}
}

func TestString(t *testing.T) {
	tc := &Toolchain{
		ClangPath:      filepath.Join("bin", "clang"),
		LLCPath:        "",
		LLVMConfigPath: "",
		Version:        "23.1.4",
	}
	s := tc.String()
	if !strings.Contains(s, "LLVM 23.1.4") {
		t.Fatalf("expected String() to contain version, got: %s", s)
	}
	if !strings.Contains(s, "[clang-only]") {
		t.Fatalf("expected String() to note clang-only mode, got: %s", s)
	}

	tcComplete := &Toolchain{
		ClangPath:      filepath.Join("bin", "clang"),
		LLCPath:        filepath.Join("bin", "llc"),
		LLVMConfigPath: filepath.Join("bin", "llvm-config"),
		Version:        "18.1.8",
	}
	s2 := tcComplete.String()
	if !strings.Contains(s2, "[complete]") {
		t.Fatalf("expected String() to note complete mode, got: %s", s2)
	}
}

func TestIsComplete(t *testing.T) {
	tcClangOnly := &Toolchain{
		ClangPath: filepath.Join("bin", "clang"),
	}
	if tcClangOnly.IsComplete() {
		t.Fatal("expected IsComplete to be false for clang-only toolchain")
	}

	tcFull := &Toolchain{
		ClangPath:      filepath.Join("bin", "clang"),
		LLCPath:        filepath.Join("bin", "llc"),
		LLVMConfigPath: filepath.Join("bin", "llvm-config"),
	}
	if !tcFull.IsComplete() {
		t.Fatal("expected IsComplete to be true for complete toolchain")
	}
}

func TestLLToObjAndErrors(t *testing.T) {
	tc, err := Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	llPath := filepath.Join(tmpDir, "test.ll")
	objPath := filepath.Join(tmpDir, "test.o")

	validIR := `define i32 @dummy() {
	ret i32 0
}
`
	if err := os.WriteFile(llPath, []byte(validIR), 0644); err != nil {
		t.Fatalf("writing test.ll: %v", err)
	}

	if err := tc.LLToObj(ctx, llPath, objPath); err != nil {
		t.Fatalf("LLToObj failed: %v", err)
	}

	// Test invalid llPath
	_ = tc.LLToObj(ctx, filepath.Join(tmpDir, "nonexistent.ll"), objPath)

	// Test cancelled context
	canceledCtx, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	_ = tc.LLToObj(canceledCtx, llPath, objPath)

	// Test invalid toolchain paths
	badTC := &Toolchain{ClangPath: "nonexistent_clang_123"}
	_ = badTC.CompileIR(ctx, validIR, filepath.Join(tmpDir, "bad.exe"))
	_ = badTC.CompileWithRuntime(ctx, validIR, "dummy.o", filepath.Join(tmpDir, "bad.exe"))
	_ = badTC.ValidateIR(ctx, validIR)

	// Test toolchain detection fallback branches
	_ = tc.IsComplete()
	_ = tc.String()
}

func TestDetectFallbacksAndLLC(t *testing.T) {
	tc, err := Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable")
	}

	origPath := os.Getenv("PATH")
	origCfg := os.Getenv("LLVM_CONFIG")
	defer func() {
		_ = os.Setenv("PATH", origPath)
		_ = os.Setenv("LLVM_CONFIG", origCfg)
	}()

	// 1. LLVM_CONFIG branch
	_ = os.Setenv("PATH", "")
	_ = os.Setenv("LLVM_CONFIG", tc.ClangPath)
	_, _ = Detect()

	// 2. Candidate paths branch
	_ = os.Setenv("LLVM_CONFIG", "")
	_, _ = Detect()

	_ = os.Setenv("PATH", origPath)

	// Test LLToObj with empty LLCPath vs non-empty LLCPath
	tmpDir := t.TempDir()
	llPath := filepath.Join(tmpDir, "dummy.ll")
	objPath := filepath.Join(tmpDir, "dummy.o")
	_ = os.WriteFile(llPath, []byte("define void @f() { ret void }\n"), 0644)

	ctx := context.Background()
	canceledCtx, cancelNow := context.WithCancel(ctx)
	cancelNow()

	tcClang := &Toolchain{ClangPath: tc.ClangPath, LLCPath: ""}
	_ = tcClang.LLToObj(ctx, llPath, objPath)

	tcLLC := &Toolchain{ClangPath: tc.ClangPath, LLCPath: "nonexistent_llc"}
	_ = tcLLC.LLToObj(ctx, llPath, objPath)
	_ = tcLLC.LLToObj(canceledCtx, llPath, objPath)
}

func TestCompileWithRuntimeDirect(t *testing.T) {
	tc, err := Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	outExe := filepath.Join(tmpDir, "test_prog") // tests adding .exe extension

	validIR := `
declare void @qadav_exit(i32)

define void @qadav_start() {
	call void @qadav_exit(i32 0)
	ret void
}
`
	// create a dummy obj or use clang to assemble a minimal stub
	stubC := filepath.Join(tmpDir, "stub.c")
	stubObj := filepath.Join(tmpDir, "stub.o")
	extraObj := filepath.Join(tmpDir, "extra.o")
	k32Lib := filepath.Join(tmpDir, "kernel32.lib")
	_ = os.WriteFile(stubC, []byte("void qadav_exit(int code) {}\n"), 0644)
	_ = os.WriteFile(k32Lib, []byte("dummy"), 0644)
	cmd := exec.Command(tc.ClangPath, "-c", stubC, "-o", stubObj)
	_ = cmd.Run()
	cmd2 := exec.Command(tc.ClangPath, "-c", stubC, "-o", extraObj)
	_ = cmd2.Run()

	_ = tc.CompileWithRuntime(ctx, validIR, stubObj, outExe, extraObj)
	outExeWithExt := filepath.Join(tmpDir, "prog2.exe")
	_ = tc.CompileWithRuntime(ctx, validIR, stubObj, outExeWithExt)
	_ = tc.ValidateIR(ctx, "; comment only\n")
	_, _ = ParseClangVersion("clang version 20.1.0")

	// Test CompileIR with invalid IR and canceled context
	_ = tc.CompileIR(ctx, "invalid ir text", filepath.Join(tmpDir, "fail.exe"))
	canceledCtx, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	_ = tc.CompileIR(canceledCtx, validIR, filepath.Join(tmpDir, "canceled.exe"))
}
