package ast

// MemoryProvenanceDecl represents a QEMP declaration:
// memory_prov mp with mem_auth ⟦ capacity ↦ 100 · ⟧
type MemoryProvenanceDecl struct {
	Pos_     Position
	Name     string
	Boundary Expr
	Capacity Expr // optional, default 1024
}

func (s *MemoryProvenanceDecl) Pos() Position { return s.Pos_ }
func (s *MemoryProvenanceDecl) stmtNode()     {}
func (s *MemoryProvenanceDecl) declNode()     {}

// VerifyMemoryStmt represents:
// verify_memory mp ·
type VerifyMemoryStmt struct {
	Pos_ Position
	Mem  Expr
}

func (s *VerifyMemoryStmt) Pos() Position { return s.Pos_ }
func (s *VerifyMemoryStmt) stmtNode()     {}

// AuditMemoryStmt represents:
// audit_memory mp ·
type AuditMemoryStmt struct {
	Pos_ Position
	Mem  Expr
}

func (s *AuditMemoryStmt) Pos() Position { return s.Pos_ }
func (s *AuditMemoryStmt) stmtNode()     {}

// CorruptShareStmt represents (ONLY for tamper testing):
// corrupt_memory_share mp⟬"secret"⟭ at ↦ 0 ·
type CorruptShareStmt struct {
	Pos_       Position
	Mem        Expr
	Key        Expr
	ShareIndex Expr
}

func (s *CorruptShareStmt) Pos() Position { return s.Pos_ }
func (s *CorruptShareStmt) stmtNode()     {}
