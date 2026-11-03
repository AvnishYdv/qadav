package llvm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Toolchain represents the LLVM tools available on this system.
// On Windows, LLCPath and LLVMConfigPath may be empty (they are optional).
type Toolchain struct {
	ClangPath      string // REQUIRED — path to clang/clang.exe
	LLCPath        string // OPTIONAL — empty on Windows prebuilt
	LLVMConfigPath string // OPTIONAL — empty on Windows prebuilt
	Version        string // Parsed from clang --version, format "23.1.4"
}

// IsComplete returns true if both llc and llvm-config are available.
// Some advanced features may require a complete toolchain.
func (t *Toolchain) IsComplete() bool {
	return t.LLCPath != "" && t.LLVMConfigPath != ""
}

// String returns a human-readable representation of the toolchain.
func (t *Toolchain) String() string {
	mode := "complete"
	if !t.IsComplete() {
		mode = "clang-only"
	}
	if t.LLCPath != "" {
		return fmt.Sprintf("LLVM %s (clang: %s, llc: %s) [%s]", t.Version, t.ClangPath, t.LLCPath, mode)
	}
	return fmt.Sprintf("LLVM %s (clang: %s) [%s]", t.Version, t.ClangPath, mode)
}

// ParseClangVersion extracts semantic version string from clang --version output.
func ParseClangVersion(output string) (string, error) {
	re := regexp.MustCompile(`clang version (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse clang version from output: %s", strings.TrimSpace(output))
	}
	verStr := matches[1]
	parts := strings.Split(verStr, ".")
	if len(parts) > 0 {
		major, err := strconv.Atoi(parts[0])
		if err == nil && major < 18 {
			return "", fmt.Errorf("LLVM 18+ required, found %s", verStr)
		}
	}
	return verStr, nil
}

// Detect locates clang and optional LLVM tools on the host.
func Detect() (*Toolchain, error) {
	clangFound := ""

	// Check order 1: LookPath in environment PATH
	if p, err := exec.LookPath("clang"); err == nil {
		clangFound = p
	} else if runtime.GOOS == "windows" {
		if pWin, errWin := exec.LookPath("clang.exe"); errWin == nil {
			clangFound = pWin
		}
	}

	// Check order 2: LLVM_CONFIG environment
	if clangFound == "" {
		if cfg := os.Getenv("LLVM_CONFIG"); cfg != "" {
			candidate := filepath.Join(filepath.Dir(cfg), "clang")
			if runtime.GOOS == "windows" {
				candidate += ".exe"
			}
			if _, err := os.Stat(candidate); err == nil {
				clangFound = candidate
			}
		}
	}

	// Check order 3, 4, 5: OS-specific candidates
	if clangFound == "" {
		candidates := []string{}
		home, _ := os.UserHomeDir()
		switch runtime.GOOS {
		case "darwin":
			candidates = append(candidates,
				"/opt/homebrew/opt/llvm@18/bin/clang",
				"/usr/local/opt/llvm@18/bin/clang",
			)
		case "windows":
			if home != "" {
				candidates = append(candidates,
					filepath.Join(home, "scoop", "apps", "llvm", "current", "bin", "clang.exe"),
				)
			}
			candidates = append(candidates,
				`C:\Users\avnis\scoop\apps\llvm\current\bin\clang.exe`,
				`C:\Program Files\LLVM\bin\clang.exe`,
			)
		default:
			candidates = append(candidates,
				"/usr/lib/llvm-18/bin/clang",
				"/usr/bin/clang",
			)
		}

		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				clangFound = c
				break
			}
		}
	}

	if clangFound == "" {
		return nil, fmt.Errorf("clang not found. Install LLVM 18+:\n  Windows: scoop install llvm\n  macOS:   brew install llvm@18\n  Ubuntu:  sudo apt install clang-18")
	}

	// Verify version using clang --version
	cmd := exec.Command(clangFound, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run %s --version: %w", clangFound, err)
	}

	ver, err := ParseClangVersion(string(out))
	if err != nil {
		return nil, err
	}

	binDir := filepath.Dir(clangFound)
	llcFound := ""
	cfgFound := ""

	llcCandidates := []string{filepath.Join(binDir, "llc")}
	cfgCandidates := []string{filepath.Join(binDir, "llvm-config")}
	if runtime.GOOS == "windows" {
		llcCandidates = []string{filepath.Join(binDir, "llc.exe")}
		cfgCandidates = []string{filepath.Join(binDir, "llvm-config.exe")}
	}

	for _, c := range llcCandidates {
		if _, statErr := os.Stat(c); statErr == nil {
			llcFound = c
			break
		}
	}
	for _, c := range cfgCandidates {
		if _, statErr := os.Stat(c); statErr == nil {
			cfgFound = c
			break
		}
	}

	tc := &Toolchain{
		ClangPath:      clangFound,
		LLCPath:        llcFound,
		LLVMConfigPath: cfgFound,
		Version:        ver,
	}
	return tc, nil
}

// CompileIR compiles an LLVM IR string directly to a native executable binary.
func (t *Toolchain) CompileIR(ctx context.Context, ir string, outputPath string) error {
	tmpFile, err := os.CreateTemp("", "qadav-*.ll")
	if err != nil {
		return fmt.Errorf("failed to create temporary IR file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.WriteString(ir); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temporary IR file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary IR file: %w", err)
	}

	targetPath := outputPath
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(targetPath), ".exe") {
		targetPath += ".exe"
	}

	outDir := filepath.Dir(targetPath)
	if outDir != "" && outDir != "." {
		_ = os.MkdirAll(outDir, 0o750)
	}

	cmd := exec.CommandContext(ctx, t.ClangPath, tmpPath, "-o", targetPath)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr

	if runErr := cmd.Run(); runErr != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return fmt.Errorf("compilation cancelled: %w", ctx.Err())
		}
		// On Windows without full MSVC CRT, retry with standalone entry
		if runtime.GOOS == "windows" && strings.Contains(bufErr.String(), "libcmt.lib") {
			fbCmd := exec.CommandContext(ctx, t.ClangPath, tmpPath, "-o", targetPath, "-nostdlib", "-Wl,-entry:main", "-fuse-ld=lld")
			fbBuf := &bytes.Buffer{}
			fbCmd.Stderr = fbBuf
			if fbErr := fbCmd.Run(); fbErr == nil {
				return nil
			}
		}
		return fmt.Errorf("clang failed to compile IR: %w\nstderr: %s", runErr, bufErr.String())
	}
	return nil
}

// CompileWithRuntime compiles the Qadav IR to an executable, linking
// against the Qadav minimal runtime (no libc).
func (t *Toolchain) CompileWithRuntime(ctx context.Context, ir string, runtimeObj string, outputPath string, extraObjs ...string) error {
	tmpFile, err := os.CreateTemp("", "qadav-*.ll")
	if err != nil {
		return fmt.Errorf("failed to create temporary IR file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.WriteString(ir); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temporary IR file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary IR file: %w", err)
	}

	targetPath := outputPath
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(targetPath), ".exe") {
		targetPath += ".exe"
	}

	outDir := filepath.Dir(targetPath)
	if outDir != "" && outDir != "." {
		_ = os.MkdirAll(outDir, 0o750)
	}

	// Single-step assemble & link: pass tmpPath (.ll) directly to clang with runtime objects
	var linkArgs []string
	if runtime.GOOS == "windows" {
		rtDir := filepath.Dir(runtimeObj)
		k32Lib := filepath.Join(rtDir, "kernel32.lib")
		linkArgs = []string{tmpPath, runtimeObj}
		linkArgs = append(linkArgs, extraObjs...)
		linkArgs = append(linkArgs,
			"-o", targetPath,
			"-nostdlib",
			"-fuse-ld=lld",
			"-Wl,-entry:qadav_start",
			"-Wl,-subsystem:console",
		)
		if _, err := os.Stat(k32Lib); err == nil {
			linkArgs = append(linkArgs, k32Lib)
		} else {
			linkArgs = append(linkArgs, "-lkernel32")
		}
	} else {
		// Linux
		linkArgs = []string{tmpPath, runtimeObj}
		linkArgs = append(linkArgs, extraObjs...)
		linkArgs = append(linkArgs,
			"-o", targetPath,
			"-nostdlib",
			"-Wl,-entry:_start",
			"-static",
		)
	}

	linkCmd := exec.CommandContext(ctx, t.ClangPath, linkArgs...)
	var linkErr bytes.Buffer
	linkCmd.Stderr = &linkErr
	if err := linkCmd.Run(); err != nil {
		return fmt.Errorf("clang linker failed: %w\nstderr: %s", err, linkErr.String())
	}

	return nil
}

// LLToObj compiles an LLVM IR source file (.ll) to an object file (.o/.obj).
func (t *Toolchain) LLToObj(ctx context.Context, llPath, objPath string) error {
	bufErr := &bytes.Buffer{}
	if t.LLCPath != "" {
		cmd := exec.CommandContext(ctx, t.LLCPath, llPath, "-filetype=obj", "-o", objPath)
		cmd.Stderr = bufErr
		if err := cmd.Run(); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return fmt.Errorf("compilation cancelled: %w", ctx.Err())
			}
			return fmt.Errorf("llc failed to produce object: %w\nstderr: %s", err, bufErr.String())
		}
		return nil
	}

	cmd := exec.CommandContext(ctx, t.ClangPath, "-c", llPath, "-o", objPath)
	cmd.Stderr = bufErr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return fmt.Errorf("compilation cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("clang failed to produce object: %w\nstderr: %s", err, bufErr.String())
	}
	return nil
}

// ValidateIR checks the validity of LLVM IR without generating an output binary.
func (t *Toolchain) ValidateIR(ctx context.Context, ir string) error {
	tmpFile, err := os.CreateTemp("", "qadav-val-*.ll")
	if err != nil {
		return fmt.Errorf("failed to create temporary IR file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmpFile.WriteString(ir); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temporary IR file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary IR file: %w", err)
	}

	devNull := "/dev/null"
	if runtime.GOOS == "windows" {
		devNull = "NUL"
	}

	cmd := exec.CommandContext(ctx, t.ClangPath, "-x", "ir", "-c", "-o", devNull, tmpPath)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr

	if runErr := cmd.Run(); runErr != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return fmt.Errorf("validation cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("clang failed to validate IR: %w\nstderr: %s", runErr, bufErr.String())
	}
	return nil
}
