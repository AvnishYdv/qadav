package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAndVerifyTrainQDS(t *testing.T) {
	// Create examples/data directory if it doesn't exist
	dir := filepath.Join("..", "..", "examples", "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create examples/data: %v", err)
	}

	qdsPath := filepath.Join(dir, "train.qds")

	// 4 samples x 2 features float64 + 4 samples int64 labels
	tensors := []TensorSpec{
		{
			Name:     "features",
			Shape:    []int{4, 2},
			DType:    "float64",
			Float64s: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
		},
		{
			Name:   "labels",
			Shape:  []int{4},
			DType:  "int64",
			Int64s: []int64{0, 1, 0, 1},
		},
	}

	if err := WriteQDS(qdsPath, tensors); err != nil {
		t.Fatalf("failed to write train.qds: %v", err)
	}

	fi, err := os.Stat(qdsPath)
	if err != nil {
		t.Fatalf("failed to stat train.qds: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("train.qds is empty")
	}
}
