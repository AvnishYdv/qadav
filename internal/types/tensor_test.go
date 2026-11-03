package types

import (
	"testing"
)

func TestTensorOf(t *testing.T) {
	typ := TensorOf(Float64)
	if !typ.IsTensor() {
		t.Fatalf("expected IsTensor() == true")
	}
	if typ.TensorElem() != Float64 {
		t.Fatalf("expected elem Float64, got %v", typ.TensorElem())
	}
	if typ.Rank() != 0 {
		t.Fatalf("expected rank 0 for unsized tensor, got %d", typ.Rank())
	}
}

func TestTensorOfShape(t *testing.T) {
	typ := TensorOfShape(Float64, []int{3, 4})
	if !typ.IsTensor() {
		t.Fatalf("expected IsTensor() == true")
	}
	if typ.Rank() != 2 {
		t.Fatalf("expected rank 2, got %d", typ.Rank())
	}
	if len(typ.Shape) != 2 || typ.Shape[0] != 3 || typ.Shape[1] != 4 {
		t.Fatalf("unexpected shape %v", typ.Shape)
	}
}

func TestTensorRank(t *testing.T) {
	s0 := TensorOf(Int64)
	if s0.Rank() != 0 {
		t.Errorf("rank should be 0, got %d", s0.Rank())
	}
	s1 := TensorOfShape(Int64, []int{10})
	if s1.Rank() != 1 {
		t.Errorf("rank should be 1, got %d", s1.Rank())
	}
	s2 := TensorOfShape(Float64, []int{2, 3})
	if s2.Rank() != 2 {
		t.Errorf("rank should be 2, got %d", s2.Rank())
	}
	if Int64.Rank() != 0 {
		t.Errorf("non-tensor rank should be 0")
	}
}

func TestTensorElem(t *testing.T) {
	t1 := TensorOf(Float64)
	if t1.TensorElem() != Float64 {
		t.Errorf("expected Float64 elem")
	}
	if Int64.TensorElem() != nil {
		t.Errorf("expected nil elem for non-tensor")
	}
}

func TestTensorString(t *testing.T) {
	tBare := TensorOf(nil)
	if tBare.String() != "tensor" {
		t.Errorf("expected 'tensor', got %q", tBare.String())
	}
	tElem := TensorOf(Float64)
	if tElem.String() != "tensor⟨float64⟩" {
		t.Errorf("expected 'tensor⟨float64⟩', got %q", tElem.String())
	}
	tShape := TensorOfShape(Float64, []int{3, 4})
	if tShape.String() != "tensor⟨float64⟩⟬3⸴ 4⟭" {
		t.Errorf("expected 'tensor⟨float64⟩⟬3⸴ 4⟭', got %q", tShape.String())
	}
}

func TestTensorEqual(t *testing.T) {
	a := TensorOfShape(Float64, []int{3, 4})
	b := TensorOfShape(Float64, []int{3, 4})
	c := TensorOfShape(Int64, []int{3, 4})
	d := TensorOfShape(Float64, []int{4, 3})
	e := TensorOf(Float64)

	if !a.Equal(b) {
		t.Errorf("expected a == b")
	}
	if a.Equal(c) {
		t.Errorf("expected a != c (different elem)")
	}
	if a.Equal(d) {
		t.Errorf("expected a != d (different shape)")
	}
	if a.Equal(e) {
		t.Errorf("expected a != e (different rank/shape)")
	}
}
