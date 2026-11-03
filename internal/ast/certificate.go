package ast

import "github.com/qadav-lang/qadav/internal/token"

// CertifiedReasonStmt represents a Type-Tagged Proof-Carrying Reasoning (TTPCR) step (Invention #15 - Patent Pending).
type CertifiedReasonStmt struct {
	Pos_       Position
	Chain      Expr
	Kind       token.Type // DEDUCE, INDUCE, ABDUCE, VERIFY
	InputExpr  Expr
	OutputExpr Expr
	Witness    Expr // bounded string/bytes
	Confidence Expr
}

func (s *CertifiedReasonStmt) Pos() Position { return s.Pos_ }
func (s *CertifiedReasonStmt) stmtNode()     {}

// AbstainThresholdDecl configures calibrated abstention policy on a reasoning chain.
type AbstainThresholdDecl struct {
	Pos_       Position
	Chain      Expr
	Threshold  Expr
	Mode       token.Type // RETRY, ESCALATE, REFUSE
	MaxRetries Expr       // optional
}

func (d *AbstainThresholdDecl) Pos() Position { return d.Pos_ }
func (d *AbstainThresholdDecl) declNode()     {}
func (d *AbstainThresholdDecl) stmtNode()     {}

// AbstainCheckStmt checks whether a reasoning chain requires abstention.
type AbstainCheckStmt struct {
	Pos_  Position
	Chain Expr
	Into  string // variable to bind result (bool / int64)
}

func (s *AbstainCheckStmt) Pos() Position { return s.Pos_ }
func (s *AbstainCheckStmt) stmtNode()     {}

// AbstainCheckExpr represents an expression evaluating whether to abstain on a chain.
type AbstainCheckExpr struct {
	Pos_  Position
	Chain Expr
}

func (e *AbstainCheckExpr) Pos() Position { return e.Pos_ }
func (e *AbstainCheckExpr) exprNode()     {}
