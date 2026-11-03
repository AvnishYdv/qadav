package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func TestParseDatasetBasic(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		dataset d ↦ load⦅"train.qds"⦆ ·
		normalize d minmax ·
		normalize d zscore ·
		shuffle d ·
		batch d ↦ 32 ⟦
			emit d ·
		⟧
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected *ast.ProcessDecl, got %T", prog.Decls[0])
	}
	stmts := proc.Body.Statements
	if len(stmts) != 6 {
		t.Fatalf("expected 6 statements, got %d", len(stmts))
	}

	if _, ok := stmts[0].(*ast.DatasetDecl); !ok {
		t.Errorf("stmt 0 expected *ast.DatasetDecl, got %T", stmts[0])
	}
	if norm1, ok := stmts[1].(*ast.NormalizeStmt); !ok || norm1.Mode != token.MINMAX {
		t.Errorf("stmt 1 expected minmax normalize, got %v", stmts[1])
	}
	if norm2, ok := stmts[2].(*ast.NormalizeStmt); !ok || norm2.Mode != token.ZSCORE {
		t.Errorf("stmt 2 expected zscore normalize, got %v", stmts[2])
	}
	if _, ok := stmts[3].(*ast.ShuffleStmt); !ok {
		t.Errorf("stmt 3 expected *ast.ShuffleStmt, got %T", stmts[3])
	}
	if _, ok := stmts[4].(*ast.BatchStmt); !ok {
		t.Errorf("stmt 4 expected *ast.BatchStmt, got %T", stmts[4])
	}
}

func TestParseQuasinaryBasic(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		quasi a ↦ ► ·
		quasi b ↦ ◄ ·
		quasi c ↦ 〇 ·
		quasi sum ↦ a ⊕ b ·
		minimize grid energy ising iterations ↦ 100 ·
		tile grid pattern ·
		yield 0 ·
	⟧
	`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected *ast.ProcessDecl, got %T", prog.Decls[0])
	}
	stmts := proc.Body.Statements
	if len(stmts) != 7 {
		t.Fatalf("expected 7 statements, got %d", len(stmts))
	}

	// stmt 0: quasi a ↦ ►
	bindA, ok := stmts[0].(*ast.BindStmt)
	if !ok {
		t.Fatalf("stmt 0 expected *ast.BindStmt, got %T", stmts[0])
	}
	qLit, ok := bindA.Value.(*ast.QuasiLit)
	if !ok || qLit.Value != 1 {
		t.Errorf("expected QuasiLit(+1), got %v", bindA.Value)
	}

	// stmt 1: quasi b ↦ ◄
	bindB := stmts[1].(*ast.BindStmt)
	if qLit, ok := bindB.Value.(*ast.QuasiLit); !ok || qLit.Value != -1 {
		t.Errorf("expected QuasiLit(-1), got %v", bindB.Value)
	}

	// stmt 2: quasi c ↦ 〇
	bindC := stmts[2].(*ast.BindStmt)
	if qLit, ok := bindC.Value.(*ast.QuasiLit); !ok || qLit.Value != 0 {
		t.Errorf("expected QuasiLit(0), got %v", bindC.Value)
	}

	// stmt 4: minimize
	if _, ok := stmts[4].(*ast.MinimizeStmt); !ok {
		t.Errorf("stmt 4 expected *ast.MinimizeStmt, got %T", stmts[4])
	}

	// stmt 5: tile
	if _, ok := stmts[5].(*ast.TileStmt); !ok {
		t.Errorf("stmt 5 expected *ast.TileStmt, got %T", stmts[5])
	}
}

func TestParseQEEH_QEMWV_MCTP(t *testing.T) {
	src := `
ethical_boundary global_boundary ⟦
	rule safety ↦ 1.0 ·
⟧

vault v1 ⟬ MyNet ⟭ threshold ↦ 3 ·

provenance p1 ⟬ MyNet ⟭ ·

process main⦅⦆ → int64 ⟦
	ethical_boundary local_b ⟦
		rule no_harm ↦ 0.95 ·
	⟧
	enforce local_b on "action1" ·
	decoherence ·
	vault v ⟬ MyNet ⟭ threshold ↦ 2 ·
	seal v ·
	unseal v with local_b ·
	wipe v ·
	provenance p ⟬ MyNet ⟭ ·
	certify p ·
	logic ok ↦ verify_training ⟬ cert ⟭ ·
	yield 0 ·
⟧
`
	l := lexer.New(src, "test.ydv")
	p := New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 4 {
		t.Fatalf("expected 4 decls, got %d", len(prog.Decls))
	}
}
