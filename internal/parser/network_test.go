package parser

import (
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/token"
)

func TestParseNetworkDetailed(t *testing.T) {
	input := `network MLP ⟦
		dense ⟬2⸴ 4⟭ ·
		relu ·
		dense ⟬4⸴ 1⟭ ·
		sigmoid ·
	⟧`
	l := lexer.New(input, "net.ydv")
	p := New(l, "net.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	net, ok := prog.Decls[0].(*ast.NetworkDecl)
	if !ok || net.Name != "MLP" {
		t.Fatalf("expected NetworkDecl MLP, got %+v", prog.Decls[0])
	}
	if len(net.Layers) != 4 {
		t.Fatalf("expected 4 layers, got %d", len(net.Layers))
	}
	if net.Layers[0].Kind != token.DENSE || len(net.Layers[0].Args) != 2 {
		t.Errorf("layer 0 mismatch: %+v", net.Layers[0])
	}
	if net.Layers[1].Kind != token.RELU {
		t.Errorf("layer 1 mismatch: %+v", net.Layers[1])
	}
	if net.Layers[2].Kind != token.DENSE || len(net.Layers[2].Args) != 2 {
		t.Errorf("layer 2 mismatch: %+v", net.Layers[2])
	}
	if net.Layers[3].Kind != token.SIGMOID {
		t.Errorf("layer 3 mismatch: %+v", net.Layers[3])
	}
}

func TestParseLearnDetailed(t *testing.T) {
	input := `learn Model on data ⟦
		loss mse ·
		optimizer sgd ·
		lr ↦ 0.05 ·
		epochs ↦ 200 ·
	⟧`
	l := lexer.New(input, "learn.ydv")
	p := New(l, "learn.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	lrn, ok := prog.Decls[0].(*ast.LearnDecl)
	if !ok || lrn.NetworkName != "Model" || lrn.OnName != "data" {
		t.Fatalf("expected LearnDecl Model on data, got %+v", prog.Decls[0])
	}
	if lrn.Loss != "mse" {
		t.Errorf("expected loss mse, got %s", lrn.Loss)
	}
	if lrn.Optimizer != "sgd" {
		t.Errorf("expected optimizer sgd, got %s", lrn.Optimizer)
	}
	if lrn.LR == nil || lrn.Epochs == nil {
		t.Errorf("expected LR and Epochs expressions, got nil")
	}
}

func TestParseProbabilisticTypes(t *testing.T) {
	input := `process main⦅⦆ → int64 ⟦
		confident⟨float64⟩ c ↦ 0.95 ·
		explore⟨int64⟩ e ↦ 42 ·
		yield 0 ·
	⟧`
	l := lexer.New(input, "prob.ydv")
	p := New(l, "prob.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
	proc, ok := prog.Decls[0].(*ast.ProcessDecl)
	if !ok {
		t.Fatalf("expected ProcessDecl")
	}
	if len(proc.Body.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(proc.Body.Statements))
	}
}
