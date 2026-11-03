package ast

import (
	"github.com/qadav-lang/qadav/internal/token"
)

// Node is the common interface implemented by all AST nodes.
type Node interface {
	Pos() Position
}

// Stmt is a statement node.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is an expression node.
type Expr interface {
	Node
	exprNode()
}

// Type is a type specification node.
type Type interface {
	Node
	typeNode()
}

// Decl is a top-level declaration node.
type Decl interface {
	Node
	declNode()
}

// Program is the root AST node containing all top-level declarations.
type Program struct {
	Pos_  Position
	Decls []Decl
}

func (p *Program) Pos() Position { return p.Pos_ }

// ---------------------------------------------------------------------
// Top-Level Declarations
// ---------------------------------------------------------------------

// Param represents a parameter in a process declaration.
type Param struct {
	Pos_ Position
	Name string
	Type Type
}

func (p *Param) Pos() Position { return p.Pos_ }

// ProcessDecl represents a process (function) declaration.
type ProcessDecl struct {
	Pos_    Position
	Name    string
	Params  []*Param
	Returns Type // may be nil for void return
	Body    *Block
}

func (d *ProcessDecl) Pos() Position { return d.Pos_ }
func (d *ProcessDecl) declNode()     {}

// AccessDecl represents a module access (import) statement.
type AccessDecl struct {
	Pos_ Position
	Path string // e.g. "std::io"
}

func (d *AccessDecl) Pos() Position { return d.Pos_ }
func (d *AccessDecl) declNode()     {}

// ExposeDecl represents an export declaration.
type ExposeDecl struct {
	Pos_  Position
	Names []string
}

func (d *ExposeDecl) Pos() Position { return d.Pos_ }
func (d *ExposeDecl) declNode()     {}

// ---------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------

// Block represents a sequence of statements enclosed in ⟦ ... ⟧.
type Block struct {
	Pos_       Position
	Statements []Stmt
}

func (s *Block) Pos() Position { return s.Pos_ }
func (s *Block) stmtNode()     {}

// BindStmt represents an immutable binding: Type name ↦ expr·
type BindStmt struct {
	Pos_  Position
	Type  Type // may be nil for inferred binding
	Name  string
	Value Expr
}

func (s *BindStmt) Pos() Position { return s.Pos_ }
func (s *BindStmt) stmtNode()     {}

// AssignStmt represents a mutable assignment: lhs ← rhs· or compound ↦⊕
type AssignStmt struct {
	Pos_ Position
	LHS  Expr
	Op   token.Type // token.ASSIGN or compound token.BIND_ADD, etc.
	RHS  Expr
}

func (s *AssignStmt) Pos() Position { return s.Pos_ }
func (s *AssignStmt) stmtNode()     {}

// EmitStmt outputs a value to stdout/runtime sink: emit value·
type EmitStmt struct {
	Pos_  Position
	Value Expr
}

func (s *EmitStmt) Pos() Position { return s.Pos_ }
func (s *EmitStmt) stmtNode()     {}

// ObserveStmt outputs debug observation: observe value·
type ObserveStmt struct {
	Pos_  Position
	Value Expr
}

func (s *ObserveStmt) Pos() Position { return s.Pos_ }
func (s *ObserveStmt) stmtNode()     {}

// YieldStmt yields or returns a value: yield expr·
type YieldStmt struct {
	Pos_  Position
	Value Expr // may be nil for bare yield
}

func (s *YieldStmt) Pos() Position { return s.Pos_ }
func (s *YieldStmt) stmtNode()     {}

// WhenStmt represents conditional branching: when cond ⟦ ... ⟧ otherwise ⟦ ... ⟧
type WhenStmt struct {
	Pos_      Position
	Condition Expr
	Then      *Block
	Else      *Block // may be nil
}

func (s *WhenStmt) Pos() Position { return s.Pos_ }
func (s *WhenStmt) stmtNode()     {}

// IterateStmt represents collection iteration: iterate x ∈ collection ⟦ ... ⟧
type IterateStmt struct {
	Pos_       Position
	Var        string
	Collection Expr
	Body       *Block
}

func (s *IterateStmt) Pos() Position { return s.Pos_ }
func (s *IterateStmt) stmtNode()     {}

// CycleStmt represents a condition-driven loop: cycle cond ⟦ ... ⟧
type CycleStmt struct {
	Pos_      Position
	Condition Expr
	Body      *Block
}

func (s *CycleStmt) Pos() Position { return s.Pos_ }
func (s *CycleStmt) stmtNode()     {}

// SuperposeStmt sets a qubit into superposition: superpose q·
type SuperposeStmt struct {
	Pos_  Position
	Qubit Expr
}

func (s *SuperposeStmt) Pos() Position { return s.Pos_ }
func (s *SuperposeStmt) stmtNode()     {}

// EntangleStmt entangles two qubits: entangle q1 q2·
type EntangleStmt struct {
	Pos_ Position
	Q1   Expr
	Q2   Expr
}

func (s *EntangleStmt) Pos() Position { return s.Pos_ }
func (s *EntangleStmt) stmtNode()     {}

// CollapseStmt measures a qubit state: collapse q ↦ bit·
type CollapseStmt struct {
	Pos_  Position
	Qubit Expr
	Into  Expr // may be nil
}

func (s *CollapseStmt) Pos() Position { return s.Pos_ }
func (s *CollapseStmt) stmtNode()     {}

// HaltStmt breaks execution of a loop: halt·
type HaltStmt struct {
	Pos_ Position
}

func (s *HaltStmt) Pos() Position { return s.Pos_ }
func (s *HaltStmt) stmtNode()     {}

// SkipStmt skips to the next iteration: skip·
type SkipStmt struct {
	Pos_ Position
}

func (s *SkipStmt) Pos() Position { return s.Pos_ }
func (s *SkipStmt) stmtNode()     {}

// ExprStmt wraps an expression as a statement: expr·
type ExprStmt struct {
	Pos_ Position
	Expr Expr
}

func (s *ExprStmt) Pos() Position { return s.Pos_ }
func (s *ExprStmt) stmtNode()     {}

// ---------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------

// Ident represents an identifier reference.
type Ident struct {
	Pos_ Position
	Name string
}

func (e *Ident) Pos() Position { return e.Pos_ }
func (e *Ident) exprNode()     {}

// IntLit represents an integer literal.
type IntLit struct {
	Pos_  Position
	Value int64
}

func (e *IntLit) Pos() Position { return e.Pos_ }
func (e *IntLit) exprNode()     {}

// FloatLit represents a floating point or complex literal.
type FloatLit struct {
	Pos_      Position
	Value     float64
	Imaginary bool // true for complex suffix (e.g. 3.14i)
}

func (e *FloatLit) Pos() Position { return e.Pos_ }
func (e *FloatLit) exprNode()     {}

// StringLit represents a string literal.
type StringLit struct {
	Pos_  Position
	Value string
}

func (e *StringLit) Pos() Position { return e.Pos_ }
func (e *StringLit) exprNode()     {}

// TensorLit represents a tensor literal expression: tensor⟬...⟭
type TensorLit struct {
	Pos_  Position
	Elem  Expr   // element type of all items? Or nil for inference
	Items []Expr // flat list for 1D; nested TensorLits for 2D
	Dims  []int  // computed by parser from structure
}

func (e *TensorLit) Pos() Position { return e.Pos_ }
func (e *TensorLit) exprNode()     {}

// BinaryExpr represents an infix operator expression: Left op Right
type BinaryExpr struct {
	Pos_  Position
	Op    token.Type
	Left  Expr
	Right Expr
}

func (e *BinaryExpr) Pos() Position { return e.Pos_ }
func (e *BinaryExpr) exprNode()     {}

// UnaryExpr represents a prefix operator expression: op Right
type UnaryExpr struct {
	Pos_  Position
	Op    token.Type
	Right Expr
}

func (e *UnaryExpr) Pos() Position { return e.Pos_ }
func (e *UnaryExpr) exprNode()     {}

// CallExpr represents a process invocation: Func⦅Args⦆
type CallExpr struct {
	Pos_ Position
	Func Expr
	Args []Expr
}

func (e *CallExpr) Pos() Position { return e.Pos_ }
func (e *CallExpr) exprNode()     {}

// MemberExpr represents member access: Object∸Member
type MemberExpr struct {
	Pos_   Position
	Object Expr
	Member string
}

func (e *MemberExpr) Pos() Position { return e.Pos_ }
func (e *MemberExpr) exprNode()     {}

// IndexExpr represents indexed access: Object⟬Index⟭
type IndexExpr struct {
	Pos_   Position
	Object Expr
	Index  Expr
}

func (e *IndexExpr) Pos() Position { return e.Pos_ }
func (e *IndexExpr) exprNode()     {}

// GroupExpr represents a grouped expression: ⦅Inner⦆
type GroupExpr struct {
	Pos_  Position
	Inner Expr
}

func (e *GroupExpr) Pos() Position { return e.Pos_ }
func (e *GroupExpr) exprNode()     {}

// RangeExpr represents a range specification: Start…End
type RangeExpr struct {
	Pos_  Position
	Start Expr
	End   Expr
}

func (e *RangeExpr) Pos() Position { return e.Pos_ }
func (e *RangeExpr) exprNode()     {}

// ---------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------

// PrimitiveType represents built-in types (e.g. int32, float64, logic, qubit).
type PrimitiveType struct {
	Pos_ Position
	Name string
}

func (t *PrimitiveType) Pos() Position { return t.Pos_ }
func (t *PrimitiveType) typeNode()     {}

// TensorType represents a tensor type specification: tensor⟨float64⟩⟬3⸴ 4⟭
type TensorType struct {
	Pos_ Position
	Elem Type  // element type; nil if unknown
	Dims []int // nil if shape unknown
}

func (t *TensorType) Pos() Position { return t.Pos_ }
func (t *TensorType) typeNode()     {}

// NamedType represents a user-defined type by name.
type NamedType struct {
	Pos_ Position
	Name string
}

func (t *NamedType) Pos() Position { return t.Pos_ }
func (t *NamedType) typeNode()     {}

// ---------------------------------------------------------------------
// Visitor & Tree Traversal
// ---------------------------------------------------------------------

// Visitor defines the interface for AST traversal.
type Visitor interface {
	Visit(node Node)
}

// Walk traverses an AST in depth-first order invoking v.Visit(node).
func Walk(v Visitor, node Node) {
	if node == nil {
		return
	}
	v.Visit(node)

	switch n := node.(type) {
	case *Program:
		for _, decl := range n.Decls {
			Walk(v, decl)
		}
	case *ProcessDecl:
		for _, param := range n.Params {
			Walk(v, param)
		}
		if n.Returns != nil {
			Walk(v, n.Returns)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *Param:
		if n.Type != nil {
			Walk(v, n.Type)
		}
	case *Block:
		for _, stmt := range n.Statements {
			Walk(v, stmt)
		}
	case *BindStmt:
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *AssignStmt:
		if n.LHS != nil {
			Walk(v, n.LHS)
		}
		if n.RHS != nil {
			Walk(v, n.RHS)
		}
	case *EmitStmt:
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *ObserveStmt:
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *YieldStmt:
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *WhenStmt:
		if n.Condition != nil {
			Walk(v, n.Condition)
		}
		if n.Then != nil {
			Walk(v, n.Then)
		}
		if n.Else != nil {
			Walk(v, n.Else)
		}
	case *IterateStmt:
		if n.Collection != nil {
			Walk(v, n.Collection)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *CycleStmt:
		if n.Condition != nil {
			Walk(v, n.Condition)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *SuperposeStmt:
		if n.Qubit != nil {
			Walk(v, n.Qubit)
		}
	case *EntangleStmt:
		if n.Q1 != nil {
			Walk(v, n.Q1)
		}
		if n.Q2 != nil {
			Walk(v, n.Q2)
		}
	case *CollapseStmt:
		if n.Qubit != nil {
			Walk(v, n.Qubit)
		}
		if n.Into != nil {
			Walk(v, n.Into)
		}
	case *ExprStmt:
		if n.Expr != nil {
			Walk(v, n.Expr)
		}
	case *BinaryExpr:
		if n.Left != nil {
			Walk(v, n.Left)
		}
		if n.Right != nil {
			Walk(v, n.Right)
		}
	case *UnaryExpr:
		if n.Right != nil {
			Walk(v, n.Right)
		}
	case *CallExpr:
		if n.Func != nil {
			Walk(v, n.Func)
		}
		for _, arg := range n.Args {
			Walk(v, arg)
		}
	case *MemberExpr:
		if n.Object != nil {
			Walk(v, n.Object)
		}
	case *IndexExpr:
		if n.Object != nil {
			Walk(v, n.Object)
		}
		if n.Index != nil {
			Walk(v, n.Index)
		}
	case *GroupExpr:
		if n.Inner != nil {
			Walk(v, n.Inner)
		}
	case *RangeExpr:
		if n.Start != nil {
			Walk(v, n.Start)
		}
		if n.End != nil {
			Walk(v, n.End)
		}
	case *NetworkDecl:
		for _, layer := range n.Layers {
			Walk(v, layer)
		}
	case *Layer:
		for _, arg := range n.Args {
			Walk(v, arg)
		}
	case *LearnDecl:
		if n.LR != nil {
			Walk(v, n.LR)
		}
		if n.Epochs != nil {
			Walk(v, n.Epochs)
		}
		if n.BatchSize != nil {
			Walk(v, n.BatchSize)
		}
	}
}
