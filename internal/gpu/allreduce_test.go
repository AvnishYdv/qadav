package gpu

import (
	"math"
	"testing"
)

func TestRingAllReduce2Workers(t *testing.T) {
	rar, err := NewRingAllReduce(2, 32)
	if err != nil {
		t.Fatalf("NewRingAllReduce failed: %v", err)
	}

	// Initialize worker 0 with [1.0, 2.0, 3.0, 4.0]
	// Initialize worker 1 with [3.0, 4.0, 5.0, 6.0]
	w0Vals := []float64{1.0, 2.0, 3.0, 4.0}
	w1Vals := []float64{3.0, 4.0, 5.0, 6.0}
	for i := 0; i < 4; i++ {
		uint64ToByteSlice(math.Float64bits(w0Vals[i]), rar.Workers[0].GradientBuf[i*8:(i+1)*8])
		uint64ToByteSlice(math.Float64bits(w1Vals[i]), rar.Workers[1].GradientBuf[i*8:(i+1)*8])
	}

	if err := rar.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Expect average: [(1+3)/2, (2+4)/2, (3+5)/2, (4+6)/2] = [2.0, 3.0, 4.0, 5.0]
	expected := []float64{2.0, 3.0, 4.0, 5.0}
	for w := 0; w < 2; w++ {
		for i := 0; i < 4; i++ {
			actual := math.Float64frombits(byteSliceToUint64(rar.Workers[w].GradientBuf[i*8 : (i+1)*8]))
			if math.Abs(actual-expected[i]) > 1e-6 {
				t.Errorf("worker %d elem %d: got %f, want %f", w, i, actual, expected[i])
			}
		}
	}
}

func TestRingAllReduce4Workers(t *testing.T) {
	rar, err := NewRingAllReduce(4, 32)
	if err != nil {
		t.Fatalf("NewRingAllReduce failed: %v", err)
	}

	for w := 0; w < 4; w++ {
		for i := 0; i < 4; i++ {
			val := float64((w + 1) * 10)
			uint64ToByteSlice(math.Float64bits(val), rar.Workers[w].GradientBuf[i*8:(i+1)*8])
		}
	}

	if err := rar.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Average of 10, 20, 30, 40 is 25.0
	for w := 0; w < 4; w++ {
		for i := 0; i < 4; i++ {
			actual := math.Float64frombits(byteSliceToUint64(rar.Workers[w].GradientBuf[i*8 : (i+1)*8]))
			if math.Abs(actual-25.0) > 1e-6 {
				t.Errorf("worker %d elem %d: got %f, want 25.0", w, i, actual)
			}
		}
	}
}

func TestParameterServerPushPullAggregate(t *testing.T) {
	ps, err := NewParameterServer(3, 32)
	if err != nil {
		t.Fatalf("NewParameterServer failed: %v", err)
	}

	if ps.GetVersion() != 1 {
		t.Fatalf("expected version 1, got %d", ps.GetVersion())
	}

	// Workers push gradients
	for w := 0; w < 3; w++ {
		worker := ps.Workers[w]
		for i := 0; i < 4; i++ {
			val := float64((w + 1) * 3) // 3.0, 6.0, 9.0 -> avg = 6.0
			uint64ToByteSlice(math.Float64bits(val), worker.GradientBuf[i*8:(i+1)*8])
		}
		if err := ps.Push(worker); err != nil {
			t.Fatalf("Push worker %d failed: %v", w, err)
		}
	}

	if err := ps.Aggregate(); err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}

	if ps.GetVersion() != 2 {
		t.Fatalf("expected version 2 after aggregate, got %d", ps.GetVersion())
	}

	// Check global model: initial 0.0 + 6.0 = 6.0
	for i := 0; i < 4; i++ {
		actual := math.Float64frombits(byteSliceToUint64(ps.GlobalModel[i*8 : (i+1)*8]))
		if math.Abs(actual-6.0) > 1e-6 {
			t.Errorf("elem %d: got %f, want 6.0", i, actual)
		}
	}

	// Pull back to worker 0
	if err := ps.Pull(ps.Workers[0]); err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	for i := 0; i < 4; i++ {
		actual := math.Float64frombits(byteSliceToUint64(ps.Workers[0].GradientBuf[i*8 : (i+1)*8]))
		if math.Abs(actual-6.0) > 1e-6 {
			t.Errorf("pulled elem %d: got %f, want 6.0", i, actual)
		}
	}
}
