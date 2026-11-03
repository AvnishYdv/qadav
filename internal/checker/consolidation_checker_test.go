package checker

import (
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

func TestConsolidationChecker_ValidTypes(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		memory_prov mp with b ⟦ capacity ↦ 50 · ⟧
		graph g ·
		consolidate mp into g with threshold ↦ 0.75 ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("expected valid consolidation to pass, got: %v", errs)
	}
}

func TestConsolidationChecker_InvalidSourceType(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		int64 not_mp ↦ 42 ·
		graph g ·
		consolidate not_mp into g with threshold ↦ 0.5 ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error when consolidation source is not a memory_prov")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Msg, "consolidate source must be a memory_prov") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error message 'consolidate source must be a memory_prov', got: %v", errs)
	}
}

func TestConsolidationChecker_InvalidTargetType(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		memory_prov mp with b ⟦ capacity ↦ 50 · ⟧
		int64 not_graph ↦ 100 ·
		consolidate mp into not_graph with threshold ↦ 0.5 ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) == 0 {
		t.Fatalf("expected error when consolidation target is not a graph")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Msg, "consolidate target must be a graph") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error message 'consolidate target must be a graph', got: %v", errs)
	}
}

func TestConsolidationChecker_ThresholdRange(t *testing.T) {
	// Valid edge cases
	validCases := []string{"0.0", "0.5", "1.0", "0.999"}
	for _, th := range validCases {
		src := `
		process main⦅⦆ → int64 ⟦
			ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
			memory_prov mp with b ⟦ capacity ↦ 50 · ⟧
			graph g ·
			consolidate mp into g with threshold ↦ ` + th + ` ·
			yield 0 ·
		⟧`
		_, errs := parseAndCheck(src)
		if len(errs) > 0 {
			t.Errorf("expected threshold %s to be valid, got: %v", th, errs)
		}
	}

	// Invalid out-of-range cases
	invalidCases := []string{"-0.01", "1.01", "2.0", "-1.0"}
	for _, th := range invalidCases {
		src := `
		process main⦅⦆ → int64 ⟦
			ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
			memory_prov mp with b ⟦ capacity ↦ 50 · ⟧
			graph g ·
			consolidate mp into g with threshold ↦ ` + th + ` ·
			yield 0 ·
		⟧`
		_, errs := parseAndCheck(src)
		if len(errs) == 0 {
			t.Errorf("expected threshold %s to be rejected, got no errors", th)
		}
	}

	// Non-float threshold type
	c := New()
	c.curScope = NewScope(nil)
	_ = c.curScope.Define(&Symbol{Name: "mp", Type: types.MemoryProv})
	_ = c.curScope.Define(&Symbol{Name: "g", Type: types.GraphT()})
	_ = c.curScope.Define(&Symbol{Name: "int_th", Type: types.Int64})
	stmt := &ast.ConsolidateStmt{
		EpisodicMem:   &ast.Ident{Name: "mp"},
		SemanticMem:   &ast.Ident{Name: "g"},
		MinConfidence: &ast.Ident{Name: "int_th"},
	}
	c.checkConsolidateStmt(stmt)
	if len(c.Errors()) == 0 {
		t.Errorf("expected error for non-float threshold identifier")
	}
}

