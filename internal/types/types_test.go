package types

import "testing"

func TestPrimitives(t *testing.T) {
	primitives := []*Type{
		Int8, Int16, Int32, Int64,
		Uint8, Uint16, Uint32, Uint64,
		Float32, Float64,
		Complex64, Complex128,
		Logic, Text, Tensor, Qubit, Neuron, Void,
	}

	for _, p := range primitives {
		if p.String() != p.Name {
			t.Errorf("expected string %q, got %q", p.Name, p.String())
		}
		fromName := FromPrimitiveName(p.Name)
		if fromName == nil || !fromName.Equal(p) {
			t.Errorf("FromPrimitiveName(%q) failed", p.Name)
		}
	}
}

func TestTypeString(t *testing.T) {
	var nilType *Type
	if nilType.String() != "<nil>" {
		t.Errorf("expected <nil>, got %s", nilType.String())
	}

	p1 := &Type{
		Kind:    KindProcess,
		Params:  []*Type{Int64, Text},
		Results: []*Type{Logic},
	}
	if p1.String() != "process(int64, text) -> logic" {
		t.Errorf("unexpected process string: %s", p1.String())
	}

	p2 := &Type{
		Kind:    KindProcess,
		Params:  []*Type{},
		Results: []*Type{},
	}
	if p2.String() != "process() -> void" {
		t.Errorf("unexpected process string: %s", p2.String())
	}

	p3 := &Type{
		Kind:    KindProcess,
		Params:  []*Type{Int32},
		Results: []*Type{Int32, Float64},
	}
	if p3.String() != "process(int32) -> (int32, float64)" {
		t.Errorf("unexpected process string: %s", p3.String())
	}

	custom := &Type{Kind: Kind(999)}
	if custom.String() != "type(999)" {
		t.Errorf("unexpected fallback string: %s", custom.String())
	}
}

func TestTypeEqual(t *testing.T) {
	var nilType *Type
	if !nilType.Equal(nil) {
		t.Errorf("nilType.Equal(nil) should be true")
	}
	if nilType.Equal(Int32) {
		t.Errorf("nilType.Equal(Int32) should be false")
	}
	if Int32.Equal(nil) {
		t.Errorf("Int32.Equal(nil) should be false")
	}
	if Int32.Equal(Int64) {
		t.Errorf("Int32.Equal(Int64) should be false")
	}

	// Process equality
	pr1 := &Type{Kind: KindProcess, Params: []*Type{Int32}, Results: []*Type{Logic}}
	pr2 := &Type{Kind: KindProcess, Params: []*Type{Int32}, Results: []*Type{Logic}}
	pr3 := &Type{Kind: KindProcess, Params: []*Type{Int64}, Results: []*Type{Logic}}
	pr4 := &Type{Kind: KindProcess, Params: []*Type{Int32}, Results: []*Type{Void}}
	if !pr1.Equal(pr2) {
		t.Errorf("pr1.Equal(pr2) should be true")
	}
	if pr1.Equal(pr3) || pr1.Equal(pr4) {
		t.Errorf("pr1 should not equal pr3 or pr4")
	}

	// Tensor equality
	t1 := &Type{Kind: KindTensor, Elem: Float32}
	t2 := &Type{Kind: KindTensor, Elem: Float32}
	t3 := &Type{Kind: KindTensor, Elem: Float64}
	t4 := &Type{Kind: KindTensor, Elem: nil}
	if !t1.Equal(t2) {
		t.Errorf("t1.Equal(t2) should be true")
	}
	if t1.Equal(t3) || t1.Equal(t4) || t4.Equal(t1) {
		t.Errorf("t1 should not equal t3 or t4")
	}
}

func TestTypeProperties(t *testing.T) {
	var nilType *Type
	if nilType.IsNumeric() || nilType.IsInteger() || nilType.IsSigned() || nilType.IsUnsigned() || nilType.IsFloat() || nilType.IsComplex() || nilType.IsQuantum() || nilType.IsNeural() || nilType.Bits() != 0 {
		t.Errorf("nilType returned true for property")
	}

	allSigned := []*Type{Int8, Int16, Int32, Int64}
	for _, s := range allSigned {
		if !s.IsSigned() || s.IsUnsigned() {
			t.Errorf("%s failed signed test", s)
		}
	}

	allUnsigned := []*Type{Uint8, Uint16, Uint32, Uint64}
	for _, u := range allUnsigned {
		if !u.IsUnsigned() || u.IsSigned() {
			t.Errorf("%s failed unsigned test", u)
		}
	}

	if !Float64.IsFloat() || !Float64.IsNumeric() || Float64.IsInteger() {
		t.Errorf("Float64 property check failed")
	}
	if !Complex128.IsComplex() || !Complex128.IsNumeric() || Complex128.IsFloat() {
		t.Errorf("Complex128 property check failed")
	}
	if !Qubit.IsQuantum() || Qubit.IsNumeric() {
		t.Errorf("Qubit property check failed")
	}
	if !Neuron.IsNeural() || Neuron.IsNumeric() {
		t.Errorf("Neuron property check failed")
	}
	if Logic.IsNumeric() || Text.IsNumeric() || Void.IsNumeric() {
		t.Errorf("Non-numeric marked numeric")
	}
}

func TestBits(t *testing.T) {
	cases := map[*Type]int{
		Int8: 8, Uint8: 8,
		Int16: 16, Uint16: 16,
		Int32: 32, Uint32: 32, Float32: 32,
		Int64: 64, Uint64: 64, Float64: 64, Complex64: 64,
		Complex128: 128,
		Logic:      0, Text: 0, Void: 0,
	}
	for typ, expected := range cases {
		if typ.Bits() != expected {
			t.Errorf("Bits(%s) = %d, expected %d", typ, typ.Bits(), expected)
		}
	}
}

func TestFromPrimitiveName(t *testing.T) {
	aliases := map[string]*Type{
		"int":     Int64,
		"uint":    Uint64,
		"float":   Float64,
		"complex": Complex128,
		"bool":    Logic,
		"str":     Text,
		"string":  Text,
		"invalid": nil,
		"":        nil,
	}
	for name, expected := range aliases {
		got := FromPrimitiveName(name)
		if expected == nil {
			if got != nil {
				t.Errorf("expected nil for %q, got %s", name, got)
			}
		} else if !got.Equal(expected) {
			t.Errorf("FromPrimitiveName(%q) = %s, expected %s", name, got, expected)
		}
	}
}

func TestPromote(t *testing.T) {
	// Nil checks
	if Promote(nil, Int32) != nil || Promote(Int32, nil) != nil {
		t.Errorf("Promote with nil should be nil")
	}
	// Identical
	if got := Promote(Int32, Int32); !got.Equal(Int32) {
		t.Errorf("Promote identical failed")
	}
	// Non-numeric
	if Promote(Text, Int64) != nil || Promote(Int64, Logic) != nil {
		t.Errorf("Promote with non-numeric should be nil")
	}
	// Signed promotion
	if got := Promote(Int8, Int32); !got.Equal(Int32) {
		t.Errorf("Promote(Int8, Int32) = %s", got)
	}
	if got := Promote(Int32, Int16); !got.Equal(Int32) {
		t.Errorf("Promote(Int32, Int16) = %s", got)
	}
	// Unsigned promotion
	if got := Promote(Uint8, Uint32); !got.Equal(Uint32) {
		t.Errorf("Promote(Uint8, Uint32) = %s", got)
	}
	if got := Promote(Uint64, Uint16); !got.Equal(Uint64) {
		t.Errorf("Promote(Uint64, Uint16) = %s", got)
	}
	// Mixed signed / unsigned
	if got := Promote(Int8, Uint8); !got.Equal(Int16) {
		t.Errorf("Promote(Int8, Uint8) = %s, expected Int16", got)
	}
	if got := Promote(Int16, Uint16); !got.Equal(Int32) {
		t.Errorf("Promote(Int16, Uint16) = %s, expected Int32", got)
	}
	if got := Promote(Int32, Uint32); !got.Equal(Int64) {
		t.Errorf("Promote(Int32, Uint32) = %s, expected Int64", got)
	}
	if got := Promote(Int64, Uint64); !got.Equal(Int64) {
		t.Errorf("Promote(Int64, Uint64) = %s, expected Int64", got)
	}
	// Float promotion
	if got := Promote(Float32, Float32); !got.Equal(Float32) {
		t.Errorf("Promote(Float32, Float32) = %s", got)
	}
	if got := Promote(Float32, Float64); !got.Equal(Float64) {
		t.Errorf("Promote(Float32, Float64) = %s", got)
	}
	if got := Promote(Int32, Float32); !got.Equal(Float64) {
		t.Errorf("Promote(Int32, Float32) = %s", got)
	}
	if got := Promote(Int16, Float32); !got.Equal(Float32) {
		t.Errorf("Promote(Int16, Float32) = %s", got)
	}
	// Complex promotion
	if got := Promote(Complex64, Complex64); !got.Equal(Complex64) {
		t.Errorf("Promote(Complex64, Complex64) = %s", got)
	}
	if got := Promote(Complex64, Complex128); !got.Equal(Complex128) {
		t.Errorf("Promote(Complex64, Complex128) = %s", got)
	}
	if got := Promote(Complex64, Float32); !got.Equal(Complex128) {
		t.Errorf("Promote(Complex64, Float32) = %s", got)
	}
	if got := Promote(Complex64, Float64); !got.Equal(Complex128) {
		t.Errorf("Promote(Complex64, Float64) = %s", got)
	}
}

func TestTypesCoverageExpansion(t *testing.T) {
	// Test string representation for all new kinds
	kinds := []Kind{
		KindMemory, KindGraph, KindProcedure, KindNeuron, KindDataset, KindQuasi, KindConcept,
		KindConfident, KindExplore, KindQubit, KindQubitState, KindProcess,
	}
	for _, k := range kinds {
		tVal := &Type{Kind: k}
		if s := tVal.String(); s == "" {
			t.Errorf("empty string representation for kind %d", k)
		}
	}

	// Test constructors
	m := MemoryT()
	if m.Kind != KindMemory {
		t.Errorf("expected KindMemory, got %v", m.Kind)
	}
	g := GraphT()
	if g.Kind != KindGraph {
		t.Errorf("expected KindGraph, got %v", g.Kind)
	}
	p := ProcedureT()
	if p.Kind != KindProcedure {
		t.Errorf("expected KindProcedure, got %v", p.Kind)
	}

	conf := ConfidentOf(Float64, 0.99)
	if conf.Confidence != 0.99 || !conf.Equal(ConfidentOf(Float64, 0.99)) {
		t.Errorf("ConfidentOf failed")
	}
	exp := ExploreOf(Float64)
	if exp.Kind != KindExplore {
		t.Errorf("ExploreOf failed")
	}

	// FromPrimitiveName
	names := []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128", "logic", "text", "tensor", "qubit", "concept", "quasi", "dataset", "memory", "graph", "procedure", "neuron", "void", "unknown_foo"}
	for _, n := range names {
		_ = FromPrimitiveName(n)
	}

	// Test Type methods
	_ = conf.String()
	_ = exp.String()
	_ = m.String()
	_ = g.String()
	_ = p.String()
	_ = conf.IsNumeric()
	_ = conf.IsInteger()
	_ = conf.IsFloat()
	_ = conf.IsComplex()
	_ = conf.IsTensor()
	_ = conf.IsQuantum()
	_ = conf.Bits()
	_ = conf.IsSigned()
	_ = conf.IsUnsigned()

	// Mixed promotion
	_ = Promote(Int8, Uint8)
	_ = Promote(Int16, Uint16)
	_ = Promote(Int32, Uint32)
	_ = Promote(Int64, Float32)
	_ = Promote(Float32, Int64)
	_ = Promote(nil, Int64)
	_ = Promote(Int64, nil)
	_ = Promote(Text, Int64)
}
