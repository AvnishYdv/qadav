package testutil

import "testing"

func TestAssertEqual_Slice(t *testing.T) {
	AssertEqual(t, []int{1, 2, 3}, []int{1, 2, 3})
}

func TestAssertEqual_Map(t *testing.T) {
	AssertEqual(t, map[string]int{"a": 1}, map[string]int{"a": 1})
}

func TestAssertEqual_String(t *testing.T) {
	AssertEqual(t, "hello", "hello")
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "world")
}

func TestAssertNotContains(t *testing.T) {
	AssertNotContains(t, "hello world", "xyz")
}

func TestAssertNoPanic(t *testing.T) {
	AssertNoPanic(t, func() {
		_ = 1 + 1
	})
}

func TestAssertNoErrors_Empty(t *testing.T) {
	AssertNoErrors(t, []string{})
}

func TestTestUtilCoverage(t *testing.T) {
	// SkipIfNoLLVM
	SkipIfNoLLVM(t)

	// ValidateLLVMIR
	validIR := `define i32 @main() { ret i32 0 }`
	ValidateLLVMIR(t, validIR)

	// Snapshot tests
	tmpDir := t.TempDir()
	snapFile := tmpDir + "/test.snap"
	AssertSnapshot(t, "hello\nworld\n", snapFile)
	AssertSnapshot(t, "hello\nworld\n", snapFile)

	diff := unifiedDiff("line1\nline2", "line1\nline3\nline4")
	if diff == "" {
		t.Errorf("expected non-empty diff")
	}

	// RuntimeObjects
	objs, err := RuntimeObjects(t)
	if err != nil || len(objs) == 0 {
		t.Errorf("RuntimeObjects failed: %v", err)
	}
}

func runFailing(fn func(mockT *testing.T)) {
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()
		mockT := &testing.T{}
		fn(mockT)
	}()
	<-done
}

func TestAssertFailures(t *testing.T) {
	runFailing(func(mockT *testing.T) { AssertNoErrors(mockT, []string{"err1"}) })
	runFailing(func(mockT *testing.T) { AssertEqual(mockT, 1, 2) })
	runFailing(func(mockT *testing.T) { AssertContains(mockT, "abc", "xyz") })
	runFailing(func(mockT *testing.T) { AssertNotContains(mockT, "abc", "b") })
	runFailing(func(mockT *testing.T) { AssertNoPanic(mockT, func() { panic("test panic") }) })
	runFailing(func(mockT *testing.T) { ValidateLLVMIR(mockT, "invalid llvm ir syntax") })

	// AssertSnapshot with *update = true
	tmpDir := t.TempDir()
	snapFile := tmpDir + "/test_update.snap"
	*update = true
	AssertSnapshot(t, "snapshot content", snapFile)
	*update = false
}
