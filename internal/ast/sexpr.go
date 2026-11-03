package ast

import (
	"fmt"
	"strings"
)

// ToSExpr returns a canonical, deterministic S-expression string representation
// of any AST node for byte-for-byte serialization and equivalence verification.
func ToSExpr(node Node) string {
	if node == nil {
		return "()"
	}

	switch n := node.(type) {
	case *Program:
		var sb strings.Builder
		sb.WriteString("(Program")
		for _, d := range n.Decls {
			sb.WriteString(" ")
			sb.WriteString(ToSExpr(d))
		}
		sb.WriteString(")")
		return sb.String()

	case *ProcessDecl:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(ProcessDecl %q (", n.Name))
		for i, p := range n.Params {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Param %q %s)", p.Name, ToSExpr(p.Type)))
		}
		sb.WriteString(") ")
		if n.Returns != nil {
			sb.WriteString(ToSExpr(n.Returns))
		} else {
			sb.WriteString("()")
		}
		sb.WriteString(" ")
		sb.WriteString(ToSExpr(n.Body))
		sb.WriteString(")")
		return sb.String()

	case *ProcedureDecl:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(ProcedureDecl %q (", n.Name))
		for i, p := range n.Params {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Param %q %s)", p.Name, ToSExpr(p.Type)))
		}
		sb.WriteString(") ")
		if n.RetType != nil {
			sb.WriteString(ToSExpr(n.RetType))
		} else {
			sb.WriteString("()")
		}
		sb.WriteString(" ")
		sb.WriteString(ToSExpr(n.Body))
		sb.WriteString(")")
		return sb.String()

	case *RecordDeclStmt:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(RecordDecl %q (", n.Name))
		for i, f := range n.Fields {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Field %q %q)", f.Name, f.Type))
		}
		sb.WriteString("))")
		return sb.String()

	case *VariantDeclStmt:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(VariantDecl %q (", n.Name))
		for i, c := range n.Cases {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Case %q (", c.Name))
			for j, f := range c.Fields {
				if j > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("(Field %q %q)", f.Name, f.Type))
			}
			sb.WriteString("))")
		}
		sb.WriteString("))")
		return sb.String()

	case *Block:
		var sb strings.Builder
		sb.WriteString("(Block")
		for _, s := range n.Statements {
			sb.WriteString(" ")
			sb.WriteString(ToSExpr(s))
		}
		sb.WriteString(")")
		return sb.String()

	case *BindStmt:
		return fmt.Sprintf("(Bind %q %s %s)", n.Name, ToSExpr(n.Type), ToSExpr(n.Value))

	case *AssignStmt:
		return fmt.Sprintf("(Assign %s %s %s)", ToSExpr(n.LHS), string(n.Op), ToSExpr(n.RHS))

	case *EmitStmt:
		return fmt.Sprintf("(Emit %s)", ToSExpr(n.Value))

	case *ObserveStmt:
		return fmt.Sprintf("(Observe %s)", ToSExpr(n.Value))

	case *YieldStmt:
		if n.Value != nil {
			return fmt.Sprintf("(Yield %s)", ToSExpr(n.Value))
		}
		return "(Yield ())"

	case *WhenStmt:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(When %s %s", ToSExpr(n.Condition), ToSExpr(n.Then)))
		if n.Else != nil {
			sb.WriteString(" ")
			sb.WriteString(ToSExpr(n.Else))
		}
		sb.WriteString(")")
		return sb.String()

	case *IterateStmt:
		return fmt.Sprintf("(Iterate %q %s %s)", n.Var, ToSExpr(n.Collection), ToSExpr(n.Body))

	case *CycleStmt:
		return fmt.Sprintf("(Cycle %s %s)", ToSExpr(n.Condition), ToSExpr(n.Body))

	case *MatchStmt:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(Match %s (", ToSExpr(n.Subject)))
		for i, arm := range n.Arms {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Arm %q (", arm.CaseName))
			for k, b := range arm.Bindings {
				if k > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(fmt.Sprintf("%q", b))
			}
			sb.WriteString(") ")
			sb.WriteString(ToSExpr(arm.Body))
			sb.WriteString(")")
		}
		sb.WriteString(")")
		if n.Otherwise != nil {
			sb.WriteString(fmt.Sprintf(" (Otherwise %s)", ToSExpr(n.Otherwise)))
		}
		sb.WriteString(")")
		return sb.String()

	case *ExprStmt:
		return fmt.Sprintf("(ExprStmt %s)", ToSExpr(n.Expr))

	case *Ident:
		return fmt.Sprintf("(Ident %q)", n.Name)

	case *IntLit:
		return fmt.Sprintf("(IntLit %d)", n.Value)

	case *FloatLit:
		return fmt.Sprintf("(FloatLit %f)", n.Value)

	case *StringLit:
		return fmt.Sprintf("(StringLit %q)", n.Value)

	case *ConfidentLit:
		return fmt.Sprintf("(ConfidentLit %s %f)", ToSExpr(n.Value), n.Score)

	case *QuasiLit:
		return fmt.Sprintf("(QuasiLit %d)", n.Value)

	case *BinaryExpr:
		return fmt.Sprintf("(BinaryExpr %q %s %s)", string(n.Op), ToSExpr(n.Left), ToSExpr(n.Right))

	case *UnaryExpr:
		return fmt.Sprintf("(UnaryExpr %q %s)", string(n.Op), ToSExpr(n.Right))

	case *CallExpr:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(CallExpr %s (", ToSExpr(n.Func)))
		for i, a := range n.Args {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(ToSExpr(a))
		}
		sb.WriteString("))")
		return sb.String()

	case *RecordInitExpr:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(RecordInit %q (", n.RecordName))
		for i, f := range n.Fields {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(FieldInit %q %s)", f.Field, ToSExpr(f.Value)))
		}
		sb.WriteString("))")
		return sb.String()

	case *VariantInitExpr:
		return fmt.Sprintf("(VariantInit %q %q %s)", n.VariantName, n.CaseName, ToSExpr(n.Value))

	case *FieldAccessExpr:
		return fmt.Sprintf("(FieldAccess %s %q)", ToSExpr(n.Object), n.Field)

	case *PrimitiveType:
		return fmt.Sprintf("(PrimitiveType %q)", n.Name)

	case *NamedType:
		return fmt.Sprintf("(NamedType %q)", n.Name)

	case *TensorType:
		return fmt.Sprintf("(TensorType %s)", ToSExpr(n.Elem))

	case *RSCAAttestStmt:
		return fmt.Sprintf("(RSCAAttest %d %q %q %q %q)", n.Stage, n.Source, n.Compiler, n.Output, n.Into)

	case *RSCAVerifyStmt:
		return fmt.Sprintf("(RSCAVerify %q %q)", n.Record, n.Into)

	case *CorruptRSCAStmt:
		return fmt.Sprintf("(CorruptRSCA %q %s)", n.Record, ToSExpr(n.Offset))

	case *QubitDecl:
		var sb strings.Builder
		sb.WriteString("(QubitDecl (")
		for i, name := range n.Names {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%q", name))
		}
		sb.WriteString("))")
		return sb.String()

	case *GateStmt:
		return fmt.Sprintf("(Gate %q %s)", n.Gate, ToSExpr(n.Qubit))

	case *CNOTStmt:
		return fmt.Sprintf("(CNOT %s %s)", ToSExpr(n.Control), ToSExpr(n.Target))

	case *NetworkDecl:
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("(NetworkDecl %q (", n.Name))
		for i, l := range n.Layers {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("(Layer %q %d %d)", string(l.Kind), l.InSize, l.OutSize))
		}
		sb.WriteString("))")
		return sb.String()

	default:
		return fmt.Sprintf("(Node %T)", node)
	}
}
