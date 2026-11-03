package ast

// VaultDecl represents a model weight vault declaration:
// vault v ⟬ SecretModel ⟭ threshold ↦ 3 ·
type VaultDecl struct {
	Pos_       Position
	Name       string
	NetworkRef Expr
	Threshold  Expr
}

func (s *VaultDecl) Pos() Position { return s.Pos_ }
func (s *VaultDecl) stmtNode()     {}
func (s *VaultDecl) declNode()     {}

// SealStmt seals a network's weights into the vault:
// seal v ·
type SealStmt struct {
	Pos_    Position
	Vault   Expr
	Network Expr
}

func (s *SealStmt) Pos() Position { return s.Pos_ }
func (s *SealStmt) stmtNode()     {}

// UnsealStmt unseals weights for inference with QEEH boundary authorization:
// unseal v with auth ·
type UnsealStmt struct {
	Pos_     Position
	Vault    Expr
	Boundary Expr
}

func (s *UnsealStmt) Pos() Position { return s.Pos_ }
func (s *UnsealStmt) stmtNode()     {}

// WipeStmt performs an emergency wipe on vault shares:
// wipe v ·
type WipeStmt struct {
	Pos_  Position
	Vault Expr
}

func (s *WipeStmt) Pos() Position { return s.Pos_ }
func (s *WipeStmt) stmtNode()     {}

