package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/qadav-lang/qadav/internal/llvm"
	"github.com/qadav-lang/qadav/internal/runtime"
)

var (
	runtimeOnce sync.Once
	runtimeObjs []string
	runtimeErr  error
)

func buildRuntimeOnce() ([]string, error) {
	tc, err := llvm.Detect()
	if err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(os.TempDir(), "qadav_testutil_runtime_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	return runtime.Build(tc, cacheDir)
}

// RuntimeObjects returns cached runtime object paths.
// Compiles runtime ONCE per test binary, reused by all tests.
func RuntimeObjects(t *testing.T) ([]string, error) {
	runtimeOnce.Do(func() {
		runtimeObjs, runtimeErr = buildRuntimeOnce()
	})
	if runtimeErr != nil && t != nil {
		t.Skip("runtime build unavailable:", runtimeErr)
	}
	return runtimeObjs, runtimeErr
}
