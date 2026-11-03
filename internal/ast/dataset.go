package ast

import "github.com/qadav-lang/qadav/internal/token"

// DatasetDecl represents: dataset d ↦ load⦅"path.qds"⦆ ·
type DatasetDecl struct {
	Pos_ Position
	Name string
	Path Expr // string literal
}

func (d *DatasetDecl) Pos() Position { return d.Pos_ }
func (d *DatasetDecl) stmtNode()     {}

// NormalizeStmt represents: normalize d minmax · or normalize d zscore ·
type NormalizeStmt struct {
	Pos_    Position
	Dataset Expr
	Mode    token.Type // token.MINMAX, token.ZSCORE
}

func (n *NormalizeStmt) Pos() Position { return n.Pos_ }
func (n *NormalizeStmt) stmtNode()     {}

// BatchStmt represents: batch d ↦ 32 ⟦ ... ⟧
type BatchStmt struct {
	Pos_    Position
	Dataset Expr
	Size    Expr
	Body    *Block
}

func (b *BatchStmt) Pos() Position { return b.Pos_ }
func (b *BatchStmt) stmtNode()     {}

// ShuffleStmt represents: shuffle d ·
type ShuffleStmt struct {
	Pos_    Position
	Dataset Expr
}

func (s *ShuffleStmt) Pos() Position { return s.Pos_ }
func (s *ShuffleStmt) stmtNode()     {}
