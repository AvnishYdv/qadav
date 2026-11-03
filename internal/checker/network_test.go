package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/lexer"
	"github.com/qadav-lang/qadav/internal/parser"
	"github.com/qadav-lang/qadav/internal/types"
)

func parseAndCheckNetwork(t *testing.T, src string) (*Checker, []Error) {
	t.Helper()
	l := lexer.New(src, "test.ydv")
	p := parser.New(l, "test.ydv")
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	c := New()
	errs := c.Check(prog)
	return c, errs
}

func TestNetworkValidDeclaration(t *testing.T) {
	src := `network Linear ⟦
		dense ⟬2⸴ 4⟭ ·
		relu ·
		dense ⟬4⸴ 1⟭ ·
		sigmoid ·
	⟧`
	c, errs := parseAndCheckNetwork(t, src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
	net, ok := c.Networks()["Linear"]
	if !ok {
		t.Fatalf("network Linear not registered")
	}
	if net.InputSize != 2 || net.OutputSize != 1 {
		t.Errorf("expected in=2, out=1, got in=%d, out=%d", net.InputSize, net.OutputSize)
	}
	if len(net.Layers) != 4 {
		t.Errorf("expected 4 layers, got %d", len(net.Layers))
	}
}

func TestNetworkShapeMismatch(t *testing.T) {
	src := `network Bad ⟦
		dense ⟬2⸴ 4⟭ ·
		dense ⟬3⸴ 1⟭ ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected shape mismatch error, got none")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Msg, "shape mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'shape mismatch' in errors, got %v", errs)
	}
}

func TestNetworkFirstLayerActivationError(t *testing.T) {
	src := `network BadAct ⟦
		relu ·
		dense ⟬2⸴ 1⟭ ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected error on first layer activation, got none")
	}
}

func TestNetworkCallInProcess(t *testing.T) {
	src := `network Net ⟦
		dense ⟬3⸴ 2⟭ ·
		relu ·
	⟧

	process main⦅⦆ → int64 ⟦
		tensor⟨float64⟩⟬3⟭ x ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⟭ ·
		tensor⟨float64⟩⟬2⟭ y ↦ Net⦅x⦆ ·
		emit y ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors on network call: %v", errs)
	}
}

func TestLearnValidBlock(t *testing.T) {
	src := `network Net ⟦
		dense ⟬2⸴ 1⟭ ·
	⟧

	learn Net on data ⟦
		loss mse ·
		optimizer sgd ·
		lr ↦ 0.01 ·
		epochs ↦ 100 ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors on valid learn block: %v", errs)
	}
}

func TestLearnUnknownNetwork(t *testing.T) {
	src := `learn Ghost on data ⟦
		loss mse ·
		optimizer sgd ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected error for unknown network Ghost, got none")
	}
}

func TestLearnInvalidLossAndOptimizer(t *testing.T) {
	src := `network Net ⟦
		dense ⟬1⸴ 1⟭ ·
	⟧

	learn Net on data ⟦
		loss invalid_loss ·
		optimizer magic_opt ·
	⟧`
	_, errs := parseAndCheckNetwork(t, src)
	if len(errs) < 2 {
		t.Errorf("expected 2 errors for invalid loss and optimizer, got %v", errs)
	}
}

func TestCIRProbabilisticTypes(t *testing.T) {
	c := New()
	pos := ast.Position{Line: 1, Col: 1}
	ct := c.CheckProbabilisticType(pos, types.Float64, 0.95)
	if ct.Kind != types.KindConfident || ct.Confidence != 0.95 {
		t.Errorf("expected Confident<float64> with 0.95, got %v", ct)
	}

	// Invalid confidence score
	bad := c.CheckProbabilisticType(pos, types.Float64, 1.5)
	if len(c.Errors()) == 0 {
		t.Errorf("expected error on out-of-bounds confidence score 1.5")
	}
	if bad.Kind != types.KindFloat64 {
		t.Errorf("expected fallback to base type, got %v", bad)
	}
}
