package ast

import (
	"github.com/qadav-lang/qadav/internal/token"
)

// ReasoningChainDecl declares a QERA quantum-entangled reasoning attestation chain.
type ReasoningChainDecl struct {
	Pos_     Position
	Name     string
	Boundary Expr // QEEH ethical boundary identifier
	Capacity Expr // Max reasoning steps (default 100)
}

func (d *ReasoningChainDecl) Pos() Position { return d.Pos_ }
func (d *ReasoningChainDecl) declNode()     {}
func (d *ReasoningChainDecl) stmtNode()     {}

// ReasonStepStmt attests and commits a single reasoning step.
type ReasonStepStmt struct {
	Pos_       Position
	Chain      Expr       // target reasoning chain
	Operation  token.Type // DEDUCE, INDUCE, ABDUCE, VERIFY
	InputExpr  Expr       // premise / memory source
	OutputExpr Expr       // conclusion / target
	Confidence Expr       // confidence float literal (0.0 - 1.0)
}

func (s *ReasonStepStmt) Pos() Position { return s.Pos_ }
func (s *ReasonStepStmt) stmtNode()     {}

// VerifyChainStmt verifies the cryptographic integrity and rolling commitment of a reasoning chain.
type VerifyChainStmt struct {
	Pos_  Position
	Chain Expr
}

func (s *VerifyChainStmt) Pos() Position { return s.Pos_ }
func (s *VerifyChainStmt) stmtNode()     {}

// AuditReasoningStmt audits all steps and outputs the chain root.
type AuditReasoningStmt struct {
	Pos_  Position
	Chain Expr
}

func (s *AuditReasoningStmt) Pos() Position { return s.Pos_ }
func (s *AuditReasoningStmt) stmtNode()     {}

// CorruptReasoningStepStmt simulates reasoning step tampering (test-only, requires --test flag).
type CorruptReasoningStepStmt struct {
	Pos_      Position
	Chain     Expr
	StepIndex Expr
	AtVal     Expr
}

func (s *CorruptReasoningStepStmt) Pos() Position { return s.Pos_ }
func (s *CorruptReasoningStepStmt) stmtNode()     {}

