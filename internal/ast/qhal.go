package ast

// QASMExportStmt represents:
// qasm_export into qasm with size 1024 ·
type QASMExportStmt struct {
	Pos_   Position
	Into   string
	BufLen Expr
}

func (s *QASMExportStmt) Pos() Position { return s.Pos_ }
func (s *QASMExportStmt) stmtNode()     {}

// QHALBackendAnnotation represents:
// @qhal_backend sim
type QHALBackendAnnotation struct {
	Pos_ Position
	Kind string // "sim", "ibmq", "ionq", "rigetti"
}

func (s *QHALBackendAnnotation) Pos() Position { return s.Pos_ }
func (s *QHALBackendAnnotation) stmtNode()     {}
