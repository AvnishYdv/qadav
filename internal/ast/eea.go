package ast

// EvolvePolicyStmt represents `evolve_policy <kind> into <ident> with rate <Expr> budget <Expr> ·`
type EvolvePolicyStmt struct {
	Pos_        Position
	Kind        string // "constant" | "decaying" | "adaptive"
	Into        string
	InitialRate Expr
	TotalBudget Expr
}

func (s *EvolvePolicyStmt) Pos() Position { return s.Pos_ }
func (s *EvolvePolicyStmt) stmtNode()     {}

// EvolveBoundedStmt represents `evolve_bounded <policy> from <s1> to <s2> with reason <r> max_flips <flips> max_new_symbols <syms> into <proposal> <envelope> ·`
type EvolveBoundedStmt struct {
	Pos_          Position
	Policy        string
	FromState     Expr
	ToState       Expr
	Reason        Expr
	MaxBitFlips   Expr
	MaxNewSymbols Expr
	IntoProposal  string
	IntoEnvelope  string
}

func (s *EvolveBoundedStmt) Pos() Position { return s.Pos_ }
func (s *EvolveBoundedStmt) stmtNode()     {}

// EEAVerifyStmt represents `eea_verify <envelope> into <is_valid> ·`
type EEAVerifyStmt struct {
	Pos_     Position
	Envelope string
	Into     string
}

func (s *EEAVerifyStmt) Pos() Position { return s.Pos_ }
func (s *EEAVerifyStmt) stmtNode()     {}

// CorruptEEAStmt represents `corrupt_eea <envelope> at ↦ <offset> ·` (--test only)
type CorruptEEAStmt struct {
	Pos_     Position
	Envelope string
	Offset   Expr
}

func (s *CorruptEEAStmt) Pos() Position { return s.Pos_ }
func (s *CorruptEEAStmt) stmtNode()     {}
