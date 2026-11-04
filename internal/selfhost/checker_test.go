package selfhost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelfhostCheckerIdentity(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "examples", "qeeh_demo.ydv"),
		filepath.Join("..", "..", "examples", "qera_demo.ydv"),
		filepath.Join("..", "..", "examples", "record_demo.ydv"),
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read %s: %v", file, err)
			}

			shOut, shErr := CheckSource(string(src), file)
			if shErr != 0 {
				t.Fatalf("Self-hosted checker failed with code %d", shErr)
			}

			if shOut == "" {
				t.Errorf("Expected non-empty output for %s", file)
			}
		})
	}
}
