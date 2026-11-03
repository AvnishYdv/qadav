package codegen

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func (g *Generator) genEthicalBoundaryDecl(b *ast.EthicalBoundaryDecl) {
	numRules := len(b.Rules)
	if numRules == 0 {
		numRules = 1
	}

	penaltiesArr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = alloca [%d x double]\n", penaltiesArr, numRules)
	for i, rule := range b.Rules {
		elemPtr := g.nextTemp()
		fmt.Fprintf(&g.sb, "  %s = getelementptr inbounds [%d x double], [%d x double]* %s, i64 0, i64 %d\n", elemPtr, numRules, numRules, penaltiesArr, i)
		fmt.Fprintf(&g.sb, "  store double %f, double* %s\n", rule.Penalty, elemPtr)
	}

	castPtr := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = bitcast [%d x double]* %s to double*\n", castPtr, numRules, penaltiesArr)

	resTemp := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = call i8* @qadav_qeeh_create(i32 %d, double* %s)\n", resTemp, numRules, castPtr)

	for i, rule := range b.Rules {
		nameConst, nameLen := g.addStringConstant(rule.Name)
		fmt.Fprintf(&g.sb, "  call void @qadav_qeeh_set_rule_name(i8* %s, i32 %d, i8* getelementptr inbounds ([%d x i8], [%d x i8]* %s, i64 0, i64 0))\n",
			resTemp, i, nameLen, nameLen, nameConst)
	}

	if b.Name != "" {
		ptr := fmt.Sprintf("%%%s.ptr", b.Name)
		fmt.Fprintf(&g.sb, "  %s = alloca i8*\n", ptr)
		slot := &varSlot{
			ptr:      ptr,
			llvmType: "i8*",
			qtype:    types.EthicalBoundary,
		}
		g.vars[b.Name] = slot
		fmt.Fprintf(&g.sb, "  store i8* %s, i8** %s\n", resTemp, ptr)
	}
}

func (g *Generator) genEnforceStmt(s *ast.EnforceStmt) {
	bVal := g.genExpr(s.Boundary)
	actVal := g.genExpr(s.Action)

	evalRes := g.nextTemp()
	if actVal.Type == "i8*" {
		fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qeeh_evaluate_action(i8* %s, i8* %s)\n", evalRes, bVal.Ref, actVal.Ref)
	} else {
		fmt.Fprintf(&g.sb, "  %s = call i32 @qadav_qeeh_evaluate(i8* %s, double* null)\n", evalRes, bVal.Ref)
	}

	isViolated := g.nextTemp()
	fmt.Fprintf(&g.sb, "  %s = icmp ne i32 %s, 0\n", isViolated, evalRes)

	violationLbl := g.nextLabel("ethical_violation")
	passLbl := g.nextLabel("ethical_pass")

	fmt.Fprintf(&g.sb, "  br i1 %s, label %%%s, label %%%s\n", isViolated, violationLbl, passLbl)

	fmt.Fprintf(&g.sb, "%s:\n", violationLbl)
	g.isTerminated = false
	// Trigger quantum decoherence (zeroes state amplitudes)
	fmt.Fprintf(&g.sb, "  call void @qadav_qeeh_trigger_decoherence(i8* null)\n")
	fmt.Fprintf(&g.sb, "  br label %%%s\n", passLbl)

	fmt.Fprintf(&g.sb, "%s:\n", passLbl)
	g.isTerminated = false
}

func (g *Generator) genDecoherenceStmt(s *ast.DecoherenceStmt) {
	var targetRef = "null"
	if s.Target != nil {
		targetVal := g.genExpr(s.Target)
		targetRef = targetVal.Ref
	}
	fmt.Fprintf(&g.sb, "  call void @qadav_qeeh_trigger_decoherence(i8* %s)\n", targetRef)
}
