package ast

// LearnDecl represents a training/learning block:
// learn Model on dataset ⟦ loss mse · optimizer sgd · lr ↦ 0.01 · epochs ↦ 100 · ⟧
// learn Model on x ⸴ y ⟦ loss mse · optimizer adam · lr ↦ 0.01 · epochs ↦ 100 · ⟧
type LearnDecl struct {
	Pos_        Position
	Name        string // alias for NetworkName
	NetworkName string
	OnName      string // dataset or X identifier
	OnYName     string // optional Y / target identifier
	Loss        string // "mse", "cross_entropy"
	Optimizer   string // "sgd", "adam"
	LR          Expr   // learning rate
	Epochs      Expr   // epochs count
	BatchSize   Expr   // optional
}

func (d *LearnDecl) Pos() Position { return d.Pos_ }
func (d *LearnDecl) declNode()     {}
func (d *LearnDecl) stmtNode()     {}
