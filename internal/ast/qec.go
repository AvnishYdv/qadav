package ast

import (
	"github.com/qadav-lang/qadav/internal/token"
)

// QECMemoryDecl declares a quantum error-corrected memory instance.
// Grammar: qec_memory <name> as surface_code distance ↦ 3 with <boundary> ⟦ ⟧ ·
type QECMemoryDecl struct {
	Pos_     Position
	Name     string
	CodeType token.Type // SURFACE_CODE, QLDPC_CODE
	Distance Expr       // code distance d (3, 5, 7, 9, 11)
	Boundary Expr       // QEEH ethical boundary identifier
}

func (d *QECMemoryDecl) Pos() Position { return d.Pos_ }
func (d *QECMemoryDecl) declNode()     {}
func (d *QECMemoryDecl) stmtNode()     {}

// SyndromeStmt extracts/measures a quantum error syndrome into a variable.
// Grammar: syndrome <qec_mem> into <ident> ·
type SyndromeStmt struct {
	Pos_   Position
	QECMem Expr
	Into   string
}

func (s *SyndromeStmt) Pos() Position { return s.Pos_ }
func (s *SyndromeStmt) stmtNode()     {}

// VerifySyndromeStmt verifies the cryptographic commitment of a measured syndrome.
// Grammar: verify_syndrome <qec_mem> with <syndrome> ·
type VerifySyndromeStmt struct {
	Pos_     Position
	QECMem   Expr
	Syndrome Expr
}

func (s *VerifySyndromeStmt) Pos() Position { return s.Pos_ }
func (s *VerifySyndromeStmt) stmtNode()     {}

// CorrectStmt applies error correction using a verified syndrome.
// Grammar: correct <qec_mem> with <syndrome> ·
type CorrectStmt struct {
	Pos_     Position
	QECMem   Expr
	Syndrome Expr
}

func (s *CorrectStmt) Pos() Position { return s.Pos_ }
func (s *CorrectStmt) stmtNode()     {}

// CorruptSyndromeStmt simulates adversarial syndrome corruption (test-only, requires --test).
// Grammar: corrupt_syndrome <qec_mem>⟬<syndrome>⟭ at ↦ <offset> ·
type CorruptSyndromeStmt struct {
	Pos_     Position
	QECMem   Expr
	Syndrome Expr
	AtVal    Expr
}

func (s *CorruptSyndromeStmt) Pos() Position { return s.Pos_ }
func (s *CorruptSyndromeStmt) stmtNode()     {}

