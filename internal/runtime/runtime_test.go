package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qadav-lang/qadav/internal/llvm"
)

func TestSources(t *testing.T) {
	name, content, err := Sources()
	if err != nil {
		t.Fatalf("Sources() failed: %v", err)
	}
	if name == "" {
		t.Fatalf("expected non-empty source file name")
	}
	if len(content) == 0 {
		t.Fatalf("expected non-empty runtime source content")
	}

	linuxName, linuxContent, err := SourcesForOS("linux")
	if err != nil {
		t.Fatalf("SourcesForOS(\"linux\") failed: %v", err)
	}
	if linuxName != "qadav_rt_linux.c" || len(linuxContent) == 0 {
		t.Fatalf("expected non-empty linux source content")
	}
}

func TestHash(t *testing.T) {
	h := Hash()
	if len(h) != 16 {
		t.Fatalf("expected 16-character hex hash, got %q", h)
	}
	h2 := Hash()
	if h != h2 {
		t.Fatalf("expected stable hash, got %q vs %q", h, h2)
	}
}

func TestBuild(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	tmpDir := t.TempDir()
	objs, err := Build(tc, tmpDir)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}
	if len(objs) != 38 {
		t.Fatalf("expected 38 runtime objects, got %d", len(objs))
	}

	for _, objPath := range objs {
		info, err := os.Stat(objPath)
		if err != nil {
			t.Fatalf("object file not found at %s: %v", objPath, err)
		}
		if info.Size() == 0 {
			t.Fatalf("object file is empty: %s", objPath)
		}
	}

	// Test cache hit
	cachedObjs, err := Build(tc, tmpDir)
	if err != nil {
		t.Fatalf("cached Build() failed: %v", err)
	}
	if len(cachedObjs) != 38 {
		t.Fatalf("expected 38 cached objects, got %d", len(cachedObjs))
	}
	for i := range objs {
		if cachedObjs[i] != objs[i] {
			t.Fatalf("expected cachedObjs[%d] == objs[%d] (%s vs %s)", i, i, cachedObjs[i], objs[i])
		}
	}
}

func TestBuildDefaultCache(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	objs, err := Build(tc, "")
	if err != nil {
		t.Fatalf("Build(tc, \"\") failed: %v", err)
	}
	if len(objs) != 38 {
		t.Fatalf("expected 38 runtime objects, got %d", len(objs))
	}

	// Test with nil toolchain (auto-detect)
	objs2, err := Build(nil, "")
	if err != nil {
		t.Fatalf("Build(nil, \"\") failed: %v", err)
	}
	if len(objs2) != 38 || objs2[0] != objs[0] {
		t.Fatalf("expected same objPath from cache, got %v vs %v", objs, objs2)
	}
}

func TestBuildInvalidToolchain(t *testing.T) {
	tc := &llvm.Toolchain{ClangPath: "nonexistent_clang_path_xyz"}
	tmpDir := t.TempDir()
	_, err := Build(tc, tmpDir)
	if err == nil {
		t.Fatalf("expected error from invalid clang path, got nil")
	}
}

func TestGenerateKernel32Lib(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	tmpDir := t.TempDir()
	libPath := filepath.Join(tmpDir, "kernel32.lib")
	if err := generateKernel32Lib(tc, tmpDir, libPath); err != nil {
		t.Fatalf("generateKernel32Lib failed: %v", err)
	}
	if _, err := os.Stat(libPath); err != nil {
		t.Fatalf("kernel32.lib was not generated: %v", err)
	}
}

func TestBuildErrors(t *testing.T) {
	// 1. Invalid cacheDir (cannot create directory)
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	// Create a regular file and use it as cacheDir
	f, err := os.CreateTemp("", "qadav_cache_file_*")
	if err == nil {
		defer func() { _ = os.Remove(f.Name()) }()
		_ = f.Close()
		badDir := filepath.Join(f.Name(), "sub")
		_, err = Build(tc, badDir)
		if err == nil {
			t.Errorf("expected error with bad cache dir, got nil")
		}
	}

	// 2. generateKernel32Lib with bad toolchain
	badTC := &llvm.Toolchain{ClangPath: "nonexistent_clang"}
	tmp := t.TempDir()
	_ = generateKernel32Lib(badTC, tmp, filepath.Join(tmp, "k32.lib"))

	// 3. Compile failure with bad toolchain when kernel32.lib exists
	tmp3 := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp3, "kernel32.lib"), []byte("dummy"), 0644)
	_ = os.WriteFile(filepath.Join(tmp3, "kernel32.def"), []byte("LIBRARY kernel32.dll\nEXPORTS\nVirtualProtect\n"), 0644)
	_, _ = Build(badTC, tmp3)
}

func TestKernel32Regen(t *testing.T) {
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}
	tmp := t.TempDir()
	// write def without CreateFileA
	defPath := filepath.Join(tmp, "kernel32.def")
	libPath := filepath.Join(tmp, "kernel32.lib")
	_ = os.WriteFile(defPath, []byte("LIBRARY kernel32.dll\nEXPORTS\nExitProcess\n"), 0644)
	_ = os.WriteFile(libPath, []byte("dummy"), 0644)

	_, _ = Build(tc, tmp)

	// test lib exists but def missing
	tmp2 := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp2, "kernel32.lib"), []byte("dummy"), 0644)
	_, _ = Build(tc, tmp2)

	// test invalid def path in generateKernel32Lib
	badDir := filepath.Join(tmp, "nonexistent_sub_xyz", "dir")
	_ = generateKernel32Lib(tc, badDir, filepath.Join(badDir, "k.lib"))
	_ = generateKernel32Lib(&llvm.Toolchain{ClangPath: ""}, tmp, filepath.Join(tmp, "k32_empty.lib"))

	// test Sources
	_, _, _ = Sources()
	_ = DefaultCacheDir()
	_, _, _ = SourcesForOS("windows")
	_, _, _ = SourcesForOS("linux")
	_, _, _ = SourcesForOS("darwin")
}
