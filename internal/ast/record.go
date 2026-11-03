package ast

// FieldDecl represents a field in a record or a variant case.
type FieldDecl struct {
	Pos_ Position
	Name string
	Type string // text, int64, float64, tensor<T>, record, variant, etc.
}

func (f *FieldDecl) Pos() Position { return f.Pos_ }

// RecordDeclStmt represents a record declaration:
// record NAME ⟦ field ↦ type · ... ⟧
type RecordDeclStmt struct {
	Pos_   Position
	Name   string
	Fields []*FieldDecl
}

func (s *RecordDeclStmt) Pos() Position { return s.Pos_ }
func (s *RecordDeclStmt) stmtNode()     {}
func (s *RecordDeclStmt) declNode()     {}

// VariantCase represents a case in a variant declaration:
// | Case1(f1 ↦ T1 ⸴ f2 ↦ T2) ·
type VariantCase struct {
	Pos_   Position
	Name   string
	Fields []*FieldDecl
}

func (c *VariantCase) Pos() Position { return c.Pos_ }

// VariantDeclStmt represents a variant declaration:
// variant NAME ⟦ | Case1(...) · | Case2(...) · ⟧
type VariantDeclStmt struct {
	Pos_  Position
	Name  string
	Cases []*VariantCase
}

func (s *VariantDeclStmt) Pos() Position { return s.Pos_ }
func (s *VariantDeclStmt) stmtNode()     {}
func (s *VariantDeclStmt) declNode()     {}

// MatchArm represents one arm in a match statement:
// | Case1(b1 ⸴ b2) ⟹ <stmts> ·
type MatchArm struct {
	Pos_     Position
	CaseName string
	Bindings []string
	Body     *Block
}

func (a *MatchArm) Pos() Position { return a.Pos_ }

// MatchStmt represents a match statement:
// match EXPR ⟦ | Case1(...) ⟹ ... · otherwise ⟹ ... · ⟧
type MatchStmt struct {
	Pos_      Position
	Subject   Expr
	Arms      []*MatchArm
	Otherwise *Block
}

func (s *MatchStmt) Pos() Position { return s.Pos_ }
func (s *MatchStmt) stmtNode()     {}

// RecordFieldInit represents field initialization in record_init:
// field ↦ value
type RecordFieldInit struct {
	Pos_  Position
	Field string
	Value Expr
}

func (f *RecordFieldInit) Pos() Position { return f.Pos_ }

// RecordInitExpr represents record creation:
// record_init RecordName ⟦ field1 ↦ val1 ⸴ field2 ↦ val2 · ⟧
type RecordInitExpr struct {
	Pos_       Position
	RecordName string
	Fields     []*RecordFieldInit
}

func (e *RecordInitExpr) Pos() Position { return e.Pos_ }
func (e *RecordInitExpr) exprNode()     {}

// VariantInitExpr represents variant creation:
// variant_init VariantName ⦅CaseName ↦ val⦆
type VariantInitExpr struct {
	Pos_        Position
	VariantName string
	CaseName    string
	Value       Expr
	Values      []Expr
}

func (e *VariantInitExpr) Pos() Position { return e.Pos_ }
func (e *VariantInitExpr) exprNode()     {}

// FieldAccessExpr represents record member access:
// object.field or object∸field
type FieldAccessExpr struct {
	Pos_   Position
	Object Expr
	Field  string
}

func (e *FieldAccessExpr) Pos() Position { return e.Pos_ }
func (e *FieldAccessExpr) exprNode()     {}
