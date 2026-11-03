package ast

import "github.com/qadav-lang/qadav/internal/token"

// NetworkDecl represents a neural network declaration: network Linear ⟦ ... ⟧
type NetworkDecl struct {
	Pos_   Position
	Name   string
	Layers []*Layer
}

func (d *NetworkDecl) Pos() Position { return d.Pos_ }
func (d *NetworkDecl) declNode()     {}

// Layer represents a single neural layer inside a network declaration.
type Layer struct {
	Pos_    Position
	Kind    token.Type // DENSE, RELU, SIGMOID, SOFTMAX, LAYERNORM
	Args    []Expr     // e.g., dense ⟬128⟭ -> [128] or dense ⟬2⸴ 4⟭
	Weights Expr       // Phase 2: nil (initialized at learn time)
	Biases  Expr
	InSize  int
	OutSize int
}

func (l *Layer) Pos() Position { return l.Pos_ }
