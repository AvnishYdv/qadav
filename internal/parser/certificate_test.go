package parser

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func TestParseCertifiedReason(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		certified_reason logic deduce from p1 to "conc" witness ⟬"modus_ponens"⟭ confidence ↦ 0.95 ·
		abstain_threshold logic ↦ 0.80 mode retry max_retries ↦ 3 ·
		int64 res ↦ abstain_check logic ·
		abstain_check logic into flag ·
		yield 0 ·
	⟧`

	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}

	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected ProcessDecl, got %T", prog.Decls[0])
	}
	body := proc.Body.Statements
	if len(body) < 5 {
		t.Fatalf("expected at least 5 statements, got %d", len(body))
	}

	certStmt, ok := body[0].(*ast.CertifiedReasonStmt)
	if !ok {
		t.Fatalf("expected CertifiedReasonStmt, got %T", body[0])
	}
	if certStmt.Kind != token.DEDUCE {
		t.Fatalf("expected DEDUCE, got %v", certStmt.Kind)
	}

	threshDecl, ok := body[1].(*ast.AbstainThresholdDecl)
	if !ok {
		t.Fatalf("expected AbstainThresholdDecl, got %T", body[1])
	}
	if threshDecl.Mode != token.RETRY {
		t.Fatalf("expected RETRY mode, got %v", threshDecl.Mode)
	}

	bindStmt, ok := body[2].(*ast.BindStmt)
	if !ok {
		t.Fatalf("expected BindStmt, got %T", body[2])
	}
	if _, ok := bindStmt.Value.(*ast.AbstainCheckExpr); !ok {
		t.Fatalf("expected AbstainCheckExpr, got %T", bindStmt.Value)
	}

	checkStmt, ok := body[3].(*ast.AbstainCheckStmt)
	if !ok {
		t.Fatalf("expected AbstainCheckStmt, got %T", body[3])
	}
	if checkStmt.Into != "flag" {
		t.Fatalf("expected Into 'flag', got %q", checkStmt.Into)
	}
}

func TestParseToolsAndScoped(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
		recall_token dr_token for secure_vault on research_agent ⟦
			keys ↦ ⟬ "patient_record" ⟭ ·
			expiry ↦ 3600 ·
		⟧
		grant dr_token to research_agent ·
		revoke dr_token ·
		tool_register file_scanner with tool_auth pre ⟦ 1 ⟧ post ⟦ 1 ⟧ ·
		tool_call file_scanner⦅"/var/log/kernel.log"⦆ audit audit_ledger ·
		int32 status ↦ tool_call file_scanner⦅"arg2"⦆ ·
		yield 0 ·
	⟧`

	l := lexer.New(src, "scoped.ydv")
	p := New(l, "scoped.ydv")
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
}

func TestParseErrorFormatting(t *testing.T) {
	pe := &ParseError{
		Pos:     ast.Position{Line: 10, Col: 5},
		Message: "syntax error test",
	}
	s := pe.Error()
	if !strings.Contains(s, "10:5") || !strings.Contains(s, "syntax error test") {
		t.Fatalf("unexpected ParseError string: %q", s)
	}

	var el ErrorList
	if el.Error() != "" {
		t.Fatalf("expected empty string for empty ErrorList")
	}
	el = append(el, *pe)
	if !strings.Contains(el.Error(), "10:5") {
		t.Fatalf("unexpected ErrorList string: %q", el.Error())
	}

	// Trigger synchronize by parsing broken statements
	badSrc := `process main⦅⦆ → int64 ⟦
		bad_stmt ?@! ·
		int64 valid ↦ 10 ·
		yield 0 ·
	⟧`
	lBad := lexer.New(badSrc, "bad.ydv")
	pBad := New(lBad, "bad.ydv")
	_ = pBad.Parse()
	if len(pBad.Errors()) == 0 {
		t.Fatalf("expected parse errors for bad input")
	}
}
