package ast

// TranspileOpts defines optional transpilation optimization flags.
type TranspileOpts struct {
	Optimize    bool
	ReduceDepth bool
}

// TranspileToStmt represents `transpile_to <target> into <buf> with size <len> (options ⟦ ... ⟧)? ·`
type TranspileToStmt struct {
	Pos_    Position
	Target  string // "qiskit", "cirq", "ionq", "qasm3"
	Into    string
	BufLen  Expr
	Options *TranspileOpts
}

func (s *TranspileToStmt) Pos() Position { return s.Pos_ }
func (s *TranspileToStmt) stmtNode()     {}
