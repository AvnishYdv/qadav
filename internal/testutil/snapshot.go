package testutil

import (
    "bytes"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "testing"
)

var update = flag.Bool("update", false, "update golden files")

// AssertSnapshot compares actual output with a golden file located at snapshotPath.
// If the golden file does not exist, it is created (first run).
// On mismatch, the test fails and prints a unified diff.
func AssertSnapshot(t *testing.T, actual string, snapshotPath string) {
    t.Helper()
    if *update {
        // Regenerate golden file
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
        // First run – create golden file automatically
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
        // Produce a simple diff
        diff := unifiedDiff(expected, actual)
        t.Fatalf("snapshot mismatch for %s:\n%s", snapshotPath, diff)
    }
}

// unifiedDiff returns a minimal unified diff between two strings.
func unifiedDiff(a, b string) string {
    // For simplicity, use a naive line‑by‑line diff.
    // In real projects you might use a diff library.
    var buf bytes.Buffer
    buf.WriteString("--- expected\n+++ actual\n")
    aLines := bytes.Split([]byte(a), []byte{"\n"})
    bLines := bytes.Split([]byte(b), []byte{"\n"})
    max := len(aLines)
    if len(bLines) > max {
        max = len(bLines)
    }
    for i := 0; i < max; i++ {
        var aLine, bLine []byte
        if i < len(aLines) {
            aLine = aLines[i]
        }
        if i < len(bLines) {
            bLine = bLines[i]
        }
        if !bytes.Equal(aLine, bLine) {
            buf.WriteString(fmt.Sprintf("- %s\n+ %s\n", aLine, bLine))
        }
    }
    return buf.String()
}

