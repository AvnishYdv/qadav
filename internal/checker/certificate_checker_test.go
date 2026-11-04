package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/token"
	"github.com/qadav-lang/qadav/internal/types"
)

func TestCertificateChecker_Valid(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "logic_chain", Type: types.ReasoningChainT()})
	c.globalScope.Define(&Symbol{Name: "prem", Type: types.Text})
	c.globalScope.Define(&Symbol{Name: "conc", Type: types.Text})

	// 1. Valid certified reason stmt
	s := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.Ident{Pos_: pos, Name: "prem"},
		OutputExpr: &ast.Ident{Pos_: pos, Name: "conc"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "modus_ponens"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(s)
	if len(c.errors) != 0 {
		t.Fatalf("unexpected checker errors: %v", c.errors)
	}

	// 2. Active reasoning chain fallback
	c.activeReasoningChain = &ast.Ident{Pos_: pos, Name: "logic_chain"}
	sActive := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Kind:       token.INDUCE,
		InputExpr:  &ast.Ident{Pos_: pos, Name: "prem"},
		OutputExpr: &ast.Ident{Pos_: pos, Name: "conc"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "inductive_rule"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.85},
	}
	c.checkCertifiedReasonStmt(sActive)
	if len(c.errors) != 0 {
		t.Fatalf("unexpected checker errors: %v", c.errors)
	}

	// 3. Valid abstain threshold decl (modes: RETRY, ESCALATE, REFUSE)
	for _, mode := range []token.Type{token.RETRY, token.ESCALATE, token.REFUSE} {
		d := &ast.AbstainThresholdDecl{
			Pos_:       pos,
			Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
			Threshold:  &ast.FloatLit{Pos_: pos, Value: 0.80},
			Mode:       mode,
			MaxRetries: &ast.IntLit{Pos_: pos, Value: 3},
		}
		c.checkAbstainThresholdDecl(d)
		if len(c.errors) != 0 {
			t.Fatalf("unexpected checker errors for mode %v: %v", mode, c.errors)
		}
	}

	// 4. Abstain threshold with active reasoning chain fallback
	dActive := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Threshold: &ast.FloatLit{Pos_: pos, Value: 0.75},
		Mode:      token.REFUSE,
	}
	c.checkAbstainThresholdDecl(dActive)
	if len(c.errors) != 0 {
		t.Fatalf("unexpected checker errors: %v", c.errors)
	}

	// 5. Valid abstain check stmt & expr
	cStmt := &ast.AbstainCheckStmt{
		Pos_:  pos,
		Chain: &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Into:  "res_var",
	}
	c.checkAbstainCheckStmt(cStmt)
	if len(c.errors) != 0 {
		t.Fatalf("unexpected checker errors: %v", c.errors)
	}

	cStmtActive := &ast.AbstainCheckStmt{
		Pos_: pos,
	}
	c.checkAbstainCheckStmt(cStmtActive)
	if len(c.errors) != 0 {
		t.Fatalf("unexpected checker errors: %v", c.errors)
	}

	cExpr := &ast.AbstainCheckExpr{
		Pos_:  pos,
		Chain: &ast.Ident{Pos_: pos, Name: "logic_chain"},
	}
	ty := c.checkAbstainCheckExpr(cExpr)
	if ty != types.Int64 {
		t.Fatalf("expected int64 from abstain_check expr, got %v", ty)
	}

	cExprActive := &ast.AbstainCheckExpr{
		Pos_: pos,
	}
	tyActive := c.checkAbstainCheckExpr(cExprActive)
	if tyActive != types.Int64 {
		t.Fatalf("expected int64 from active abstain_check expr, got %v", tyActive)
	}

	// Nil safety
	c.checkCertifiedReasonStmt(nil)
	c.checkAbstainThresholdDecl(nil)
	c.checkAbstainCheckStmt(nil)
	c.checkAbstainCheckExpr(nil)
}

func TestCertificateChecker_InvalidCases(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	c.globalScope.Define(&Symbol{Name: "logic_chain", Type: types.ReasoningChainT()})
	c.globalScope.Define(&Symbol{Name: "not_a_chain", Type: types.Int64})

	// Invalid chain type
	c.errors = nil
	sBadChain := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "not_a_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "w"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sBadChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on invalid chain type")
	}

	// Invalid witness (empty)
	c.errors = nil
	sEmpty := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: ""},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sEmpty)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on empty witness")
	}

	// Witness too long (> 512 bytes)
	c.errors = nil
	sHuge := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: strings.Repeat("A", 513)},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sHuge)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on witness > 512 bytes")
	}

	// Nil witness
	c.errors = nil
	sNilWit := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sNilWit)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on nil witness")
	}

	// Invalid witness type
	c.errors = nil
	sBadWitType := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.IntLit{Pos_: pos, Value: 42},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sBadWitType)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on non-string witness type")
	}

	// Nil input
	c.errors = nil
	sNilIn := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "w"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sNilIn)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on nil input")
	}

	// Nil output
	c.errors = nil
	sNilOut := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "w"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.95},
	}
	c.checkCertifiedReasonStmt(sNilOut)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on nil output")
	}

	// Invalid confidence (> 1.0)
	c.errors = nil
	sConf := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "witness"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 1.5},
	}
	c.checkCertifiedReasonStmt(sConf)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on invalid confidence")
	}

	// Invalid confidence (< 0.0)
	c.errors = nil
	sConfNeg := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "witness"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: -0.1},
	}
	c.checkCertifiedReasonStmt(sConfNeg)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on negative confidence")
	}

	// Invalid confidence type
	c.errors = nil
	sConfType := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "witness"},
		Confidence: &ast.StringLit{Pos_: pos, Value: "high"},
	}
	c.checkCertifiedReasonStmt(sConfType)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on non-float confidence type")
	}

	// Invalid threshold (< 0.0)
	c.errors = nil
	dThresh := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Chain:     &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold: &ast.FloatLit{Pos_: pos, Value: -0.1},
		Mode:      token.REFUSE,
	}
	c.checkAbstainThresholdDecl(dThresh)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on invalid threshold")
	}

	// Invalid threshold (> 1.0)
	c.errors = nil
	dThreshHigh := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Chain:     &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold: &ast.FloatLit{Pos_: pos, Value: 1.1},
		Mode:      token.REFUSE,
	}
	c.checkAbstainThresholdDecl(dThreshHigh)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on threshold > 1.0")
	}

	// Invalid threshold type
	c.errors = nil
	dThreshType := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Chain:     &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold: &ast.StringLit{Pos_: pos, Value: "strict"},
		Mode:      token.REFUSE,
	}
	c.checkAbstainThresholdDecl(dThreshType)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on non-float threshold type")
	}

	// Invalid mode
	c.errors = nil
	dMode := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Chain:     &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold: &ast.FloatLit{Pos_: pos, Value: 0.5},
		Mode:      token.Type("INVALID"),
	}
	c.checkAbstainThresholdDecl(dMode)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on invalid mode")
	}

	// Invalid max retries (< 1)
	c.errors = nil
	dRetries := &ast.AbstainThresholdDecl{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold:  &ast.FloatLit{Pos_: pos, Value: 0.5},
		Mode:       token.RETRY,
		MaxRetries: &ast.IntLit{Pos_: pos, Value: 0},
	}
	c.checkAbstainThresholdDecl(dRetries)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on max_retries < 1")
	}

	// Invalid max retries type
	c.errors = nil
	dRetriesType := &ast.AbstainThresholdDecl{
		Pos_:       pos,
		Chain:      &ast.Ident{Pos_: pos, Name: "logic_chain"},
		Threshold:  &ast.FloatLit{Pos_: pos, Value: 0.5},
		Mode:       token.RETRY,
		MaxRetries: &ast.StringLit{Pos_: pos, Value: "three"},
	}
	c.checkAbstainThresholdDecl(dRetriesType)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on non-int max_retries type")
	}

	// Missing chain in scope
	c.errors = nil
	c.activeReasoningChain = nil
	sNoChain := &ast.CertifiedReasonStmt{
		Pos_:       pos,
		Kind:       token.DEDUCE,
		InputExpr:  &ast.StringLit{Pos_: pos, Value: "p"},
		OutputExpr: &ast.StringLit{Pos_: pos, Value: "c"},
		Witness:    &ast.StringLit{Pos_: pos, Value: "w"},
		Confidence: &ast.FloatLit{Pos_: pos, Value: 0.9},
	}
	c.checkCertifiedReasonStmt(sNoChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on missing reasoning chain")
	}

	c.errors = nil
	dNoChain := &ast.AbstainThresholdDecl{
		Pos_:      pos,
		Threshold: &ast.FloatLit{Pos_: pos, Value: 0.8},
		Mode:      token.REFUSE,
	}
	c.checkAbstainThresholdDecl(dNoChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on missing reasoning chain in abstain threshold")
	}

	c.errors = nil
	cStmtNoChain := &ast.AbstainCheckStmt{Pos_: pos}
	c.checkAbstainCheckStmt(cStmtNoChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on missing reasoning chain in abstain check stmt")
	}

	c.errors = nil
	cExprNoChain := &ast.AbstainCheckExpr{Pos_: pos}
	c.checkAbstainCheckExpr(cExprNoChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on missing reasoning chain in abstain check expr")
	}

	c.errors = nil
	cStmtBadChain := &ast.AbstainCheckStmt{Pos_: pos, Chain: &ast.Ident{Pos_: pos, Name: "not_a_chain"}}
	c.checkAbstainCheckStmt(cStmtBadChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on bad chain in abstain check stmt")
	}

	c.errors = nil
	cExprBadChain := &ast.AbstainCheckExpr{Pos_: pos, Chain: &ast.Ident{Pos_: pos, Name: "not_a_chain"}}
	c.checkAbstainCheckExpr(cExprBadChain)
	if len(c.errors) == 0 {
		t.Fatalf("expected error on bad chain in abstain check expr")
	}
}
