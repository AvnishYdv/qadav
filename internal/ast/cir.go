package ast

import "github.com/qadav-lang/qadav/internal/types"

// InferStmt represents neural/cognitive inference: infer model on input ↦ out ·
type InferStmt struct {
	Pos_       Position
	Model      Expr
	Input      Expr
	Confidence Expr // optional — threshold
	Into       string
}

func (s *InferStmt) Pos() Position { return s.Pos_ }
func (s *InferStmt) stmtNode()     {}

// AssertStmt asserts a confidence threshold: assert value ≽ threshold ·
type AssertStmt struct {
	Pos_      Position
	Value     Expr
	Threshold Expr
}

func (s *AssertStmt) Pos() Position { return s.Pos_ }
func (s *AssertStmt) stmtNode()     {}

// PropagateStmt propagates confidence bounds: propagate value from source ·
type PropagateStmt struct {
	Pos_   Position
	Value  Expr
	Source Expr
}

func (s *PropagateStmt) Pos() Position { return s.Pos_ }
func (s *PropagateStmt) stmtNode()     {}

// ConfidentLit represents a calibrated probabilistic literal: score ∵ value
type ConfidentLit struct {
	Pos_  Position
	Value Expr
	Score float64 // 0.0-1.0
	State types.EpistemicState
}

func (e *ConfidentLit) Pos() Position { return e.Pos_ }
func (e *ConfidentLit) exprNode()     {}
