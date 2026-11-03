package testutil

import (
	"path/filepath"
	"testing"
)

func TestSnapshotBasic(t *testing.T) {
	// Use a temporary directory for golden file
	dir := t.TempDir()
	snapshotPath := filepath.Join(dir, "example.golden")
	actual := "hello world"
	// First run creates the golden file
	AssertSnapshot(t, actual, snapshotPath)
	// Second run should compare and pass
	AssertSnapshot(t, actual, snapshotPath)
}
