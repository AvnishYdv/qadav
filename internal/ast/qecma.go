package ast

import (
	"github.com/qadav-lang/qadav/internal/token"
)

// ModalityDecl binds a multimodal input (tensor, text, concept) for cross-modal attestation.
// Grammar: modality <name> as <kind> from <source> ·
type ModalityDecl struct {
	Pos_   Position
	Name   string
	Kind   token.Type // TENSOR_INPUT, TEXT_INPUT, CONCEPT_INPUT
	Source Expr
}

func (d *ModalityDecl) Pos() Position { return d.Pos_ }
func (d *ModalityDecl) declNode()     {}
func (d *ModalityDecl) stmtNode()     {}

// CrossModalReasonStmt cryptographically unifies multiple modalities and performs an attested reasoning step.
// Grammar: cross_modal_reason <chain> <op> from ⟬<m1>⸴ <m2>...⟭ to <output> with <confidence> ·
type CrossModalReasonStmt struct {
	Pos_       Position
	Chain      Expr
	Operation  token.Type // DEDUCE, INDUCE, ABDUCE, VERIFY
	Modalities []Expr
	OutputExpr Expr
	Confidence Expr
}

func (s *CrossModalReasonStmt) Pos() Position { return s.Pos_ }
func (s *CrossModalReasonStmt) stmtNode()     {}

