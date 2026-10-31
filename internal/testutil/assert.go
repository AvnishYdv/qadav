package testutil

import (
    "strings"
    "testing"
)

// AssertNoErrors fails the test if the error slice is non‑empty.
func AssertNoErrors(t *testing.T, errs []string) {
    t.Helper()
    if len(errs) != 0 {
        t.Fatalf("unexpected errors: %v", errs)
    }
}

// AssertEqual checks deep equality using the standard == operator where applicable.
func AssertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if expected != actual {
        t.Fatalf("expected %v, got %v", expected, actual)
    }
}

// AssertContains verifies that substr is present in s.
func AssertContains(t *testing.T, s, substr string) {
    t.Helper()
    if !strings.Contains(s, substr) {
        t.Fatalf("expected %q to contain %q", s, substr)
    }
}

// AssertNotContains verifies that substr is absent from s.
func AssertNotContains(t *testing.T, s, substr string) {
    t.Helper()
    if strings.Contains(s, substr) {
        t.Fatalf("expected %q to NOT contain %q", s, substr)
    }
}

// AssertNoPanic ensures that fn does not panic.
func AssertNoPanic(t *testing.T, fn func()) {
    t.Helper()
    defer func() {
        if r := recover(); r != nil {
            t.Fatalf("unexpected panic: %v", r)
        }
    }()
    fn()
}
