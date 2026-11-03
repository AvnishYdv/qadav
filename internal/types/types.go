package types

import "fmt"

// Kind identifies the fundamental category of a Qadav type.
type Kind int

const (
	KindInvalid Kind = iota
	// 17 Primitive Kinds
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindFloat32
	KindFloat64
	KindComplex64
	KindComplex128
	KindLogic
	KindText
	KindTensor
	KindQubit
	KindNeuron

	// Auxiliary kinds
	KindVoid
	KindProcess
	KindConfident
	KindExplore
	KindNetwork
	KindQubitState      // runtime qubit handle (opaque i8*)
	KindConcept         // prime-symbolic concept (i64 encoding)
	KindDataset         // dataset reference
	KindQuasi           // balanced ternary value {-1, 0, +1}
	KindMemory          // episodic memory handle
	KindGraph           // semantic graph handle
	KindProcedure       // executable procedure handle
	KindEthicalBoundary // QEEH ethical boundary handle
	KindHamiltonian     // QEEH Hamiltonian operator
	KindVault           // QEMWV model weight vault handle
	KindProvenance      // MCTP training provenance tracker handle
)

// Type represents a concrete Qadav type.
type Type struct {
	Kind       Kind
	Name       string
	Elem       *Type
	Dims       []int
	Shape      []int
	Params     []*Type
	Results    []*Type
	Confidence float64 // for KindConfident (0.0-1.0)
}

// Singletons for primitive types
var (
	Int8            = &Type{Kind: KindInt8, Name: "int8"}
	Int16           = &Type{Kind: KindInt16, Name: "int16"}
	Int32           = &Type{Kind: KindInt32, Name: "int32"}
	Int64           = &Type{Kind: KindInt64, Name: "int64"}
	Uint8           = &Type{Kind: KindUint8, Name: "uint8"}
	Uint16          = &Type{Kind: KindUint16, Name: "uint16"}
	Uint32          = &Type{Kind: KindUint32, Name: "uint32"}
	Uint64          = &Type{Kind: KindUint64, Name: "uint64"}
	Float32         = &Type{Kind: KindFloat32, Name: "float32"}
	Float64         = &Type{Kind: KindFloat64, Name: "float64"}
	Complex64       = &Type{Kind: KindComplex64, Name: "complex64"}
	Complex128      = &Type{Kind: KindComplex128, Name: "complex128"}
	Logic           = &Type{Kind: KindLogic, Name: "logic"}
	Text            = &Type{Kind: KindText, Name: "text"}
	Tensor          = &Type{Kind: KindTensor, Name: "tensor"}
	Qubit           = &Type{Kind: KindQubit, Name: "qubit"}
	Neuron          = &Type{Kind: KindNeuron, Name: "neuron"}
	Void            = &Type{Kind: KindVoid, Name: "void"}
	QubitState      = &Type{Kind: KindQubitState, Name: "qubit_state"}
	Concept         = &Type{Kind: KindConcept, Name: "concept"}
	Dataset         = &Type{Kind: KindDataset, Name: "dataset"}
	Quasi           = &Type{Kind: KindQuasi, Name: "quasi"}
	Memory          = &Type{Kind: KindMemory, Name: "memory"}
	Graph           = &Type{Kind: KindGraph, Name: "graph"}
	Procedure       = &Type{Kind: KindProcedure, Name: "procedure"}
	EthicalBoundary = &Type{Kind: KindEthicalBoundary, Name: "ethical_boundary"}
	Hamiltonian     = &Type{Kind: KindHamiltonian, Name: "hamiltonian"}
	Vault           = &Type{Kind: KindVault, Name: "vault"}
	Provenance      = &Type{Kind: KindProvenance, Name: "provenance"}
)

// VaultT returns the vault Type singleton.
func VaultT() *Type {
	return Vault
}

// ProvenanceT returns the provenance Type singleton.
func ProvenanceT() *Type {
	return Provenance
}

// EpistemicState represents epistemic certainty state.
type EpistemicState int

const (
	KnownState EpistemicState = iota
	UnknownState
	ExploreState
)

// Confidence represents CIR runtime-aware confidence.
type Confidence struct {
	Value      float64
	Score      float64
	Epistemic  EpistemicState
	Provenance string
}

// MemoryT returns the memory Type singleton.
func MemoryT() *Type {
	return Memory
}

// GraphT returns the graph Type singleton.
func GraphT() *Type {
	return Graph
}

// ProcedureT returns the procedure Type singleton.
func ProcedureT() *Type {
	return Procedure
}

// QuasiT returns the balanced ternary Type singleton.
func QuasiT() *Type {
	return Quasi
}

// DatasetT returns the dataset Type singleton.
func DatasetT() *Type {
	return Dataset
}

// String returns the human-readable representation of the type.
func (t *Type) String() string {
	if t == nil {
		return "<nil>"
	}
	if t.Kind == KindProcess {
		paramsStr := ""
		for i, p := range t.Params {
			if i > 0 {
				paramsStr += ", "
			}
			paramsStr += p.String()
		}
		retStr := "void"
		if len(t.Results) == 1 {
			retStr = t.Results[0].String()
		} else if len(t.Results) > 1 {
			retStr = "("
			for i, r := range t.Results {
				if i > 0 {
					retStr += ", "
				}
				retStr += r.String()
			}
			retStr += ")"
		}
		return fmt.Sprintf("process(%s) -> %s", paramsStr, retStr)
	}
	if t.Kind == KindTensor {
		s := "tensor"
		if t.Elem != nil {
			s += "⟨" + t.Elem.String() + "⟩"
		}
		shape := t.Shape
		if len(shape) == 0 && len(t.Dims) > 0 {
			shape = t.Dims
		}
		if len(shape) > 0 {
			s += "⟬"
			for i, d := range shape {
				if i > 0 {
					s += "⸴ "
				}
				s += fmt.Sprintf("%d", d)
			}
			s += "⟭"
		}
		return s
	}
	if t.Name != "" {
		return t.Name
	}
	return fmt.Sprintf("type(%d)", t.Kind)
}

// Equal returns true if two types are structurally identical.
func (t *Type) Equal(other *Type) bool {
	if t == nil || other == nil {
		return t == other
	}
	if t.Kind != other.Kind {
		return false
	}
	switch t.Kind {
	case KindProcess:
		if len(t.Params) != len(other.Params) || len(t.Results) != len(other.Results) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].Equal(other.Params[i]) {
				return false
			}
		}
		for i := range t.Results {
			if !t.Results[i].Equal(other.Results[i]) {
				return false
			}
		}
		return true
	case KindTensor:
		if (t.Elem == nil) != (other.Elem == nil) {
			return false
		}
		if t.Elem != nil && !t.Elem.Equal(other.Elem) {
			return false
		}
		s1 := t.Shape
		if len(s1) == 0 && len(t.Dims) > 0 {
			s1 = t.Dims
		}
		s2 := other.Shape
		if len(s2) == 0 && len(other.Dims) > 0 {
			s2 = other.Dims
		}
		if len(s1) != len(s2) {
			return false
		}
		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
		}
		return true
	case KindConfident:
		if (t.Elem == nil) != (other.Elem == nil) {
			return false
		}
		if t.Elem != nil && !t.Elem.Equal(other.Elem) {
			return false
		}
		return true
	case KindExplore:
		if (t.Elem == nil) != (other.Elem == nil) {
			return false
		}
		if t.Elem != nil && !t.Elem.Equal(other.Elem) {
			return false
		}
		return true
	case KindNetwork:
		if len(t.Params) != len(other.Params) || len(t.Results) != len(other.Results) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].Equal(other.Params[i]) {
				return false
			}
		}
		for i := range t.Results {
			if !t.Results[i].Equal(other.Results[i]) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// TensorOf returns an unsized tensor with element type elem.
func TensorOf(elem *Type) *Type {
	return &Type{
		Kind: KindTensor,
		Elem: elem,
	}
}

// TensorOfShape returns a tensor with element type and explicit dimensions.
func TensorOfShape(elem *Type, dims []int) *Type {
	return &Type{
		Kind:  KindTensor,
		Elem:  elem,
		Shape: dims,
		Dims:  dims,
	}
}

// ConfidentOf returns a Confident<T> probabilistic type with min confidence score.
func ConfidentOf(elem *Type, minConf float64) *Type {
	name := fmt.Sprintf("confident⟨%s⟩", elem.String())
	return &Type{
		Kind:       KindConfident,
		Name:       name,
		Elem:       elem,
		Confidence: minConf,
	}
}

// ExploreOf returns an Explore<T> probabilistic type.
func ExploreOf(elem *Type) *Type {
	name := fmt.Sprintf("explore⟨%s⟩", elem.String())
	return &Type{
		Kind: KindExplore,
		Name: name,
		Elem: elem,
	}
}

// IsProbabilistic returns true if the type is Confident<T> or Explore<T>.
func (t *Type) IsProbabilistic() bool {
	return t != nil && (t.Kind == KindConfident || t.Kind == KindExplore)
}

// NetworkOf returns a neural network type mapping inShape to outShape.
func NetworkOf(inShape, outShape []int) *Type {
	return &Type{
		Kind:    KindNetwork,
		Name:    fmt.Sprintf("network⟬%v → %v⟭", inShape, outShape),
		Shape:   outShape,
		Params:  []*Type{TensorOfShape(Float64, inShape)},
		Results: []*Type{TensorOfShape(Float64, outShape)},
	}
}

// IsTensor returns true for tensor kind.
func (t *Type) IsTensor() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindTensor
}

// Rank returns number of dimensions (0 = scalar tensor / unsized, 2 = matrix).
func (t *Type) Rank() int {
	if !t.IsTensor() {
		return 0
	}
	if len(t.Shape) > 0 {
		return len(t.Shape)
	}
	return len(t.Dims)
}

// TensorElem returns element type of tensor (nil if not tensor or unknown).
func (t *Type) TensorElem() *Type {
	if t == nil {
		return nil
	}
	return t.Elem
}

// IsNumeric reports whether t is any integer, float, or complex type.
func (t *Type) IsNumeric() bool {
	if t == nil {
		return false
	}
	return t.IsInteger() || t.IsFloat() || t.IsComplex()
}

// IsInteger reports whether t is a signed or unsigned integer type.
func (t *Type) IsInteger() bool {
	if t == nil {
		return false
	}
	return t.IsSigned() || t.IsUnsigned()
}

// IsSigned reports whether t is a signed integer type.
func (t *Type) IsSigned() bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case KindInt8, KindInt16, KindInt32, KindInt64:
		return true
	default:
		return false
	}
}

// IsUnsigned reports whether t is an unsigned integer type.
func (t *Type) IsUnsigned() bool {
	if t == nil {
		return false
	}
	switch t.Kind {
	case KindUint8, KindUint16, KindUint32, KindUint64:
		return true
	default:
		return false
	}
}

// IsFloat reports whether t is float32 or float64.
func (t *Type) IsFloat() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindFloat32 || t.Kind == KindFloat64
}

// IsComplex reports whether t is complex64 or complex128.
func (t *Type) IsComplex() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindComplex64 || t.Kind == KindComplex128
}

// IsQuantum reports whether t is a quantum qubit.
func (t *Type) IsQuantum() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindQubit
}

// IsNeural reports whether t is a neural node.
func (t *Type) IsNeural() bool {
	if t == nil {
		return false
	}
	return t.Kind == KindNeuron
}

// Bits returns the bit width of numeric types, or 0 if not applicable.
func (t *Type) Bits() int {
	if t == nil {
		return 0
	}
	switch t.Kind {
	case KindInt8, KindUint8:
		return 8
	case KindInt16, KindUint16:
		return 16
	case KindInt32, KindUint32, KindFloat32:
		return 32
	case KindInt64, KindUint64, KindFloat64, KindComplex64:
		return 64
	case KindComplex128:
		return 128
	default:
		return 0
	}
}

// FromPrimitiveName resolves a type name to a primitive type.
func FromPrimitiveName(name string) *Type {
	switch name {
	case "int8":
		return Int8
	case "int16":
		return Int16
	case "int32":
		return Int32
	case "int64", "int":
		return Int64
	case "uint8":
		return Uint8
	case "uint16":
		return Uint16
	case "uint32":
		return Uint32
	case "uint64", "uint":
		return Uint64
	case "float32":
		return Float32
	case "float64", "float":
		return Float64
	case "complex64":
		return Complex64
	case "complex128", "complex":
		return Complex128
	case "logic", "bool":
		return Logic
	case "text", "str", "string":
		return Text
	case "tensor":
		return Tensor
	case "qubit":
		return Qubit
	case "concept":
		return Concept
	case "quasi":
		return Quasi
	case "dataset":
		return Dataset
	case "memory":
		return Memory
	case "graph":
		return Graph
	case "procedure":
		return Procedure
	case "neuron":
		return Neuron
	case "void":
		return Void
	default:
		return nil
	}
}

// Promote calculates the widened common type for two numeric operands.
func Promote(a, b *Type) *Type {
	if a == nil || b == nil {
		return nil
	}
	if a.Equal(b) {
		return a
	}
	if !a.IsNumeric() || !b.IsNumeric() {
		return nil
	}

	// Complex promotion
	if a.Kind == KindComplex128 || b.Kind == KindComplex128 {
		return Complex128
	}
	if a.Kind == KindComplex64 || b.Kind == KindComplex64 {
		if a.Bits() > 32 || b.Bits() > 32 {
			return Complex128
		}
		return Complex64
	}

	// Float promotion
	if a.Kind == KindFloat64 || b.Kind == KindFloat64 {
		return Float64
	}
	if a.Kind == KindFloat32 || b.Kind == KindFloat32 {
		other := a
		if a.Kind == KindFloat32 {
			other = b
		}
		if other.Bits() >= 32 && other.IsInteger() {
			return Float64
		}
		if other.Bits() >= 64 {
			return Float64
		}
		return Float32
	}

	// Integer promotion
	if a.IsSigned() && b.IsSigned() {
		if a.Bits() >= b.Bits() {
			return a
		}
		return b
	}
	if a.IsUnsigned() && b.IsUnsigned() {
		if a.Bits() >= b.Bits() {
			return a
		}
		return b
	}

	// Mixed signed / unsigned: widen to larger bit width
	maxBits := a.Bits()
	if b.Bits() > maxBits {
		maxBits = b.Bits()
	}
	switch maxBits {
	case 8:
		return Int16
	case 16:
		return Int32
	case 32:
		return Int64
	default:
		return Int64
	}
}
