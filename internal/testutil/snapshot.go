package testutil

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// update is the -update flag for regenerating golden files.
// Registered once at package init; safe across multiple test binaries.
var update = flag.Bool("update", false, "regenerate golden files")

func AssertSnapshot(t *testing.T, actual string, snapshotPath string) {
	t.Helper()
	if *update {
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
			t.Fatalf("failed to create directories for %s: %v", snapshotPath, err)
		}
		if err := os.WriteFile(snapshotPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("failed to write golden file %s: %v", snapshotPath, err)
		}
		return
	}
	data, err := os.ReadFile(snapshotPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755); err != nil {
			t.Fatalf("failed to create directories for %s: %v", snapshotPath, err)
		}
		if err := os.WriteFile(snapshotPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("failed to write golden file %s: %v", snapshotPath, err)
		}
		t.Logf("golden file created: %s", snapshotPath)
		return
	}
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", snapshotPath, err)
	}
	expected := string(data)
	if actual != expected {
		diff := unifiedDiff(expected, actual)
		t.Fatalf("snapshot mismatch for %s:\n%s", snapshotPath, diff)
	}
}

func unifiedDiff(a, b string) string {
	var sb strings.Builder
	sb.WriteString("--- expected\n+++ actual\n")
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")
	maxLen := len(aLines)
	if len(bLines) > maxLen {
		maxLen = len(bLines)
	}
	for i := 0; i < maxLen; i++ {
		var aLine, bLine string
		if i < len(aLines) {
			aLine = aLines[i]
		}
		if i < len(bLines) {
			bLine = bLines[i]
		}
		if aLine != bLine {
			fmt.Fprintf(&sb, "- %s\n+ %s\n", aLine, bLine)
		}
	}
	return sb.String()
}
