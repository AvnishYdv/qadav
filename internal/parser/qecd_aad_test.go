package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
)

func TestParseQECDStmts(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text decoded ↦ "" ·
    text syn0 ↦ "" ·
    logic is_valid ↦ false ·
    text state_vec ↦ "" ·
    qecd_decode distance ↦ 3 from syn0 into decoded ·
    qecd_verify decoded into is_valid ·
    qecd_apply decoded on state_vec ·
    yield 0 ·
⟧`
	l := lexer.New(src, "qecd_test.ydv")
	p := New(l, "qecd_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	body := proc.Body.Statements
	if len(body) < 7 {
		t.Fatalf("expected at least 7 statements, got %d", len(body))
	}

	dStmt, ok := body[4].(*ast.QECDDecodeStmt)
	if !ok || dStmt.SyndromeIn != "syn0" || dStmt.Into != "decoded" {
		t.Fatalf("expected QECDDecodeStmt, got %#v", body[4])
	}

	vStmt, ok := body[5].(*ast.QECDVerifyStmt)
	if !ok || vStmt.Result != "decoded" || vStmt.Into != "is_valid" {
		t.Fatalf("expected QECDVerifyStmt, got %#v", body[5])
	}

	aStmt, ok := body[6].(*ast.QECDApplyStmt)
	if !ok || aStmt.Result != "decoded" || aStmt.StateVec != "state_vec" {
		t.Fatalf("expected QECDApplyStmt, got %#v", body[6])
	}
}

func TestParseAADStmts(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text cert ↦ "" ·
    text decoded ↦ "" ·
    text env ↦ "" ·
    text state ↦ "" ·
    text adapted ↦ "" ·
    logic ok ↦ false ·
    aad_cert with model ↦ "m" train ↦ "tr" max_syndrome ↦ 64 max_corr_dist ↦ 8 into cert ·
    aad_attest cert cert ↦ cert decode ↦ decoded into env ·
    meta_fit into state cert ↦ cert ·
    meta_adapt state with env into adapted ·
    meta_verify state into ok ·
    corrupt_aad_env env at ↦ 16 ·
    yield 0 ·
⟧`
	l := lexer.New(src, "aad_test.ydv")
	p := New(l, "aad_test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc := prog.Decls[0].(*ast.ProcessDecl)
	body := proc.Body.Statements
	if len(body) < 12 {
		t.Fatalf("expected at least 12 statements, got %d", len(body))
	}

	cStmt, ok := body[6].(*ast.AADCertStmt)
	if !ok || cStmt.Into != "cert" {
		t.Fatalf("expected AADCertStmt, got %#v", body[6])
	}

	atStmt, ok := body[7].(*ast.AADAttestStmt)
	if !ok || atStmt.Cert != "cert" || atStmt.Decode != "decoded" || atStmt.Into != "env" {
		t.Fatalf("expected AADAttestStmt, got %#v", body[7])
	}

	fStmt, ok := body[8].(*ast.MetaFitStmt)
	if !ok || fStmt.Into != "state" || fStmt.Cert != "cert" {
		t.Fatalf("expected MetaFitStmt, got %#v", body[8])
	}

	adStmt, ok := body[9].(*ast.MetaAdaptStmt)
	if !ok || adStmt.State != "state" || adStmt.Envelope != "env" || adStmt.Into != "adapted" {
		t.Fatalf("expected MetaAdaptStmt, got %#v", body[9])
	}

	mvStmt, ok := body[10].(*ast.MetaVerifyStmt)
	if !ok || mvStmt.State != "state" || mvStmt.Into != "ok" {
		t.Fatalf("expected MetaVerifyStmt, got %#v", body[10])
	}

	crStmt, ok := body[11].(*ast.CorruptAADEnvStmt)
	if !ok || crStmt.Env != "env" {
		t.Fatalf("expected CorruptAADEnvStmt, got %#v", body[11])
	}
}
