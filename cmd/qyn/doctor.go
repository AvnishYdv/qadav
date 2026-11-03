package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/qadav-lang/qadav/internal/llvm"
)

type doctorColors struct {
	green  string
	yellow string
	red    string
	bold   string
	reset  string
}

func initColors() doctorColors {
	if os.Getenv("NO_COLOR") != "" {
		return doctorColors{}
	}
	return doctorColors{
		green:  "\033[32m",
		yellow: "\033[33m",
		red:    "\033[31m",
		bold:   "\033[1m",
		reset:  "\033[0m",
	}
}

func checkGoVersion(c doctorColors) (bool, string) {
	ver := runtime.Version() // e.g. "go1.27.3"
	clean := strings.TrimPrefix(ver, "go")
	parts := strings.Split(clean, ".")
	if len(parts) >= 2 {
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			if major > 1 || (major == 1 && minor >= 22) {
				return true, fmt.Sprintf("%s%s ✅%s", c.green, ver, c.reset)
			}
		}
	}
	return false, fmt.Sprintf("%s%s ❌ (Go 1.22+ required)%s", c.red, ver, c.reset)
}

func checkCacheDir(c doctorColors) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("unavailable (%v) ⚠️", err)
	}
	qDir := filepath.Join(home, ".qadav")
	if info, statErr := os.Stat(qDir); statErr == nil && info.IsDir() {
		return fmt.Sprintf("%s (exists %s✅%s)", qDir, c.green, c.reset)
	}
	if mkErr := os.MkdirAll(qDir, 0o750); mkErr == nil {
		return fmt.Sprintf("%s (created %s✅%s)", qDir, c.green, c.reset)
	}
	return fmt.Sprintf("%s (failed to create) ⚠️", qDir)
}

func runDoctor() int {
	c := initColors()

	fmt.Println("Qadav Doctor v0.1.0")
	fmt.Println("────────────────────────────────────────────────────────────────")

	// 1. OS & Architecture
	fmt.Printf("OS:          %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// 2. Go runtime check
	goOk, goMsg := checkGoVersion(c)
	fmt.Printf("Go:          %s\n", goMsg)

	// 3. LLVM detection
	tc, llvmErr := llvm.Detect()
	allCriticalPassed := goOk

	if llvmErr != nil {
		allCriticalPassed = false
		fmt.Printf("LLVM:        %s❌ not found%s\n", c.red, c.reset)
		fmt.Printf("   Error: %v\n", llvmErr)
	} else {
		modeNote := "complete"
		if !tc.IsComplete() {
			modeNote = "clang-only mode"
		}
		fmt.Printf("LLVM:        clang %s %s✅%s (%s)\n", tc.Version, c.green, c.reset, modeNote)
		fmt.Printf("  clang:     %s %s✅%s\n", tc.ClangPath, c.green, c.reset)

		if tc.LLCPath != "" {
			fmt.Printf("  llc:       %s %s✅%s\n", tc.LLCPath, c.green, c.reset)
		} else {
			optNote := "optional"
			if runtime.GOOS == "windows" {
				optNote = "optional on Windows"
			}
			fmt.Printf("  llc:       not found (%s) %s⚠️%s\n", optNote, c.yellow, c.reset)
		}

		if tc.LLVMConfigPath != "" {
			fmt.Printf("  llvm-config: %s %s✅%s\n", tc.LLVMConfigPath, c.green, c.reset)
		} else {
			optNote := "optional"
			if runtime.GOOS == "windows" {
				optNote = "optional on Windows"
			}
			fmt.Printf("  llvm-config: not found (%s) %s⚠️%s\n", optNote, c.yellow, c.reset)
		}
	}

	// 4. CPU Cores
	fmt.Printf("CPU cores:   %d\n", runtime.NumCPU())

	// 5. Cache Directory
	fmt.Printf("Cache dir:   %s\n", checkCacheDir(c))

	// 6. LookPath diagnostics for verification
	goInPath := "found ✅"
	if _, err := exec.LookPath("go"); err != nil {
		goInPath = "not in PATH ⚠️"
	}
	clangInPath := "found ✅"
	if _, err := exec.LookPath("clang"); err != nil && runtime.GOOS != "windows" {
		clangInPath = "not in PATH ⚠️"
	} else if runtime.GOOS == "windows" {
		if _, errWin := exec.LookPath("clang.exe"); errWin != nil {
			clangInPath = "not in PATH (resolved via scoop/local) ⚠️"
		}
	}
	fmt.Printf("PATH tools:  go: %s | clang: %s\n", goInPath, clangInPath)

	fmt.Println()

	if allCriticalPassed {
		fmt.Printf("%s✅ All critical checks passed. You're ready to build with Qadav! 🚀%s\n\n", c.green, c.reset)
		if tc != nil && !tc.IsComplete() {
			fmt.Println("⚠️  Note: llc/llvm-config not available — Qadav will use clang-only mode.")
			fmt.Println("   This is normal on Windows and fully supported.")
		}
		return 0
	}

	fmt.Printf("%s❌ Critical environment requirements not met.%s\n", c.red, c.reset)
	fmt.Println("   Install instructions:")
	fmt.Println("   Windows: scoop install llvm")
	fmt.Println("   macOS:   brew install llvm@18")
	fmt.Println("   Ubuntu:  sudo apt install clang-18")
	return 1
}
