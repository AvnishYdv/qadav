package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/testutil"
)

func TestDatasetPipelineExecution(t *testing.T) {
	t.Parallel()
	// Create temporary QDS file
	tmpDir := t.TempDir()
	qdsPath := filepath.Join(tmpDir, "pipeline.qds")
	tensors := []testutil.TensorSpec{
		{
			Name:     "x",
			Shape:    []int{4, 2},
			DType:    "float64",
			Float64s: []float64{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0},
		},
		{
			Name:   "y",
			Shape:  []int{4},
			DType:  "int64",
			Int64s: []int64{0, 1, 0, 1},
		},
	}
	if err := testutil.WriteQDS(qdsPath, tensors); err != nil {
		t.Fatalf("failed to write qds: %v", err)
	}

	qdsPathSlash := filepath.ToSlash(qdsPath)

	src := `process main⦅⦆ → int64 ⟦
  dataset d ↦ load ⟬ "` + qdsPathSlash + `" ⟭ ·
  normalize d minmax ·
  shuffle d ·
  batch d 2 ⟦
    emit d ·
  ⟧
  yield 0 ·
⟧`

	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "dataset⟨n_tensors=2, n_samples=4⟩:") {
		t.Fatalf("unexpected dataset output: %s", out)
	}
	if !strings.Contains(out, "[x]") || !strings.Contains(out, "[y]") {
		t.Fatalf("dataset output missing tensor names: %s", out)
	}
}

func TestQuasinaryArithmeticExecution(t *testing.T) {
	t.Parallel()
	src := `process main⦅⦆ → int64 ⟦
  quasi a ↦ ► ·
  quasi b ↦ ◄ ·
  quasi c ↦ a ⊕ b ·
  quasi d ↦ a ⊕ a ·
  quasi e ↦ b ⊕ b ·
  quasi f ↦ b ⊗ b ·
  emit c ·
  emit d ·
  emit e ·
  emit f ·
  yield 0 ·
⟧`

	out := compileAndRunSource(t, src)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected 4 lines of output, got %d: %q", len(lines), out)
	}
	// c = ► ⊕ ◄ = 〇
	if strings.TrimSpace(lines[0]) != "〇" {
		t.Errorf("expected 〇 for ► ⊕ ◄, got %q", lines[0])
	}
	// d = ► ⊕ ► = ► (saturated)
	if strings.TrimSpace(lines[1]) != "►" {
		t.Errorf("expected ► for ► ⊕ ►, got %q", lines[1])
	}
	// e = ◄ ⊕ ◄ = ◄ (saturated)
	if strings.TrimSpace(lines[2]) != "◄" {
		t.Errorf("expected ◄ for ◄ ⊕ ◄, got %q", lines[2])
	}
	// f = ◄ ⊗ ◄ = ►
	if strings.TrimSpace(lines[3]) != "►" {
		t.Errorf("expected ► for ◄ ⊗ ◄, got %q", lines[3])
	}
}

func TestQuasiGridFerromagnetic(t *testing.T) {
	t.Parallel()
	src := `process main⦅⦆ → int64 ⟦
  tensor⟨quasi⟩ grid ↦ tensor⟬
    ⟬ ► ⸴ ◄ ⸴ ► ⟭ ⸴
    ⟬ ◄ ⸴ 〇 ⸴ ◄ ⟭ ⸴
    ⟬ ► ⸴ ◄ ⸴ ► ⟭
  ⟭ ·
  minimize grid energy ising iterations ↦ 1000 ·
  emit grid ·
  yield 0 ·
⟧`

	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "tensor⟨quasi⟩⟬3 ⸴ 3⟭:") {
		t.Fatalf("unexpected grid output: %s", out)
	}
	cNeg := strings.Count(out, "◄")
	cPos := strings.Count(out, "►")
	if cNeg != 9 && cPos != 9 {
		t.Fatalf("expected ferromagnetic all-same grid (9 of either ◄ or ►), got ◄=%d, ►=%d:\n%s", cNeg, cPos, out)
	}
}
