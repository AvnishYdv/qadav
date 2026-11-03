package ast

// EthicalRule represents a single ethical rule inside an ethical boundary block.
type EthicalRule struct {
	Pos_      Position
	Name      string
	Condition Expr    // constraint expression
	Penalty   float64 // violation penalty (0.0 to 1.0)
}

func (r *EthicalRule) Pos() Position { return r.Pos_ }

// EthicalBoundaryDecl declares an ethical Hamiltonian boundary:
// ethical_boundary safety ⟦ rule human_harm ↦ 1.0 · ⟧
type EthicalBoundaryDecl struct {
	Pos_  Position
	Name  string
	Rules []*EthicalRule
}

func (s *EthicalBoundaryDecl) Pos() Position { return s.Pos_ }
func (s *EthicalBoundaryDecl) stmtNode()     {}
func (s *EthicalBoundaryDecl) declNode()     {}

// EnforceStmt applies an ethical boundary check to an action:
// enforce safety on "help_user" ·
type EnforceStmt struct {
	Pos_     Position
	Boundary Expr
	Action   Expr
}

func (s *EnforceStmt) Pos() Position { return s.Pos_ }
func (s *EnforceStmt) stmtNode()     {}

// DecoherenceStmt represents explicit quantum decoherence:
// decoherence ·
type DecoherenceStmt struct {
	Pos_   Position
	Target Expr
}

func (s *DecoherenceStmt) Pos() Position { return s.Pos_ }
func (s *DecoherenceStmt) stmtNode()     {}
