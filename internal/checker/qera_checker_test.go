package checker

import (
	"testing"
)

func TestQERAChecker_ReasonStepCoverage(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		reasoning_chain rc with b ⟦ capacity ↦ 50 · ⟧
		reason rc deduce from "premise" to "conclusion" with 0.95 ·
		reason rc induce from "sample1" to "pattern" with 0.90 ·
		reason rc abduce from "observation" to "hypothesis" with 0.85 ·
		reason rc verify from "proof" to "lemma" with 0.99 ·
		verify_chain rc ·
		audit_reasoning rc ·
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("expected valid QERA statements to pass, got errors: %v", errs)
	}

	// Test without explicit chain (active chain fallback)
	src2 := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		reasoning_chain rc with b ⟦ capacity ↦ 50 · ⟧
		reason deduce from "p" to "c" with 0.95 ·
		yield 0 ·
	⟧`
	_, errs2 := parseAndCheck(src2)
	if len(errs2) > 0 {
		t.Fatalf("expected active chain deduction to pass, got errors: %v", errs2)
	}

	// Test corrupt_reasoning_step gating
	srcCorrupt := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		reasoning_chain rc with b ⟦ capacity ↦ 50 · ⟧
		reason rc deduce from "p" to "c" with 0.95 ·
		corrupt_reasoning_step rc⟬0⟭ at ↦ 0 ·
		yield 0 ·
	⟧`
	_, errsWithoutTest := parseAndCheckTestMode(srcCorrupt, false)
	if len(errsWithoutTest) == 0 {
		t.Fatalf("expected error without --test flag for corrupt_reasoning_step")
	}
	_, errsWithTest := parseAndCheckTestMode(srcCorrupt, true)
	if len(errsWithTest) > 0 {
		t.Fatalf("expected corrupt_reasoning_step with --test to pass, got errors: %v", errsWithTest)
	}

	// Corrupt share test
	srcCorruptShare := `
	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		memory_prov mp with b ⟦ capacity ↦ 50 · ⟧
		corrupt_memory_share mp⟬"k"⟭ at ↦ 0 ·
		yield 0 ·
	⟧`
	_, errsShareNoTest := parseAndCheckTestMode(srcCorruptShare, false)
	if len(errsShareNoTest) == 0 {
		t.Fatalf("expected error without --test flag for corrupt_memory_share")
	}
	_, errsShareTest := parseAndCheckTestMode(srcCorruptShare, true)
	if len(errsShareTest) > 0 {
		t.Fatalf("expected corrupt_memory_share with --test to pass, got errors: %v", errsShareTest)
	}
}

func TestChecker_ExtraCoverage(t *testing.T) {
	// Program with network, learn, and decoherence
	src := `
	network Net ⟦
		dense ⟬2⸴ 4⟭ ·
		relu ·
		dense ⟬4⸴ 1⟭ ·
	⟧

	process main⦅⦆ → int64 ⟦
		ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
		decoherence ·
		vault v with b ⟬Net⟭ threshold ↦ 3 ·
		license lic for v on Net ⟦
			max_epochs ↦ 10 ·
			max_rate ↦ 0.5 ·
			expiry ↦ 3600 ·
		⟧
		yield 0 ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("expected valid network and decoherence to pass, got errors: %v", errs)
	}

	// QECMA and Distill error paths
	srcBadDistill := `
	network NetA ⟦ dense ⟬2⸴ 1⟭ · ⟧
	process main⦅⦆ → int64 ⟦
		distill not_a_lic on NetA with 1⸴ 2 ⟦ rate ↦ 0.1 · ⟧
		yield 0 ·
	⟧`
	_, errsBadDistill := parseAndCheck(srcBadDistill)
	if len(errsBadDistill) == 0 {
		t.Fatalf("expected errors for bad distill license")
	}

	// QECMA cross modal reason without modalities
	srcBadQECMA := `
	process main⦅⦆ → int64 ⟦
		cross_modal_reason not_a_chain deduce from ⟬⟭ to "out" with 0.9 ·
		yield 0 ·
	⟧`
	_, errsBadQECMA := parseAndCheck(srcBadQECMA)
	if len(errsBadQECMA) == 0 {
		t.Fatalf("expected error for empty modalities in cross_modal_reason")
	}
}

func TestChecker_DatasetAndQuasinaryAndAGI(t *testing.T) {
	src := `
	procedure proc_test⦅int64 a⦆ → int64 ⟦
		yield a ·
	⟧

	process main⦅⦆ → int64 ⟦
		dataset d ↦ load ⦅ "examples/data/train.qds" ⦆ ·
		normalize d minmax ·
		normalize d zscore ·
		shuffle d ·
		batch d 2 ⟦
			emit d ·
		⟧

		tensor⟨quasi⟩ grid ↦ tensor⟬
			⟬ ► ⸴ ◄ ⸴ ► ⟭ ⸴
			⟬ ◄ ⸴ 〇 ⸴ ◄ ⟭ ⸴
			⟬ ► ⸴ ◄ ⸴ ► ⟭
		⟭ ·
		minimize grid energy ising iterations ↦ 10 ·

		graph gr ·
		node gr ⟬ "n1" ↦ 101 ⟭ ·
		node gr ⟬ "n2" ↦ 102 ⟭ ·
		link gr ⟬ "n1" ⸴ "n2" ⸴ 1.0 ⟭ ·
		tensor⟨int64⟩ visited ↦ trace gr ⟬ "n1" ⸴ 2 ⟭ ·

		int64 res ↦ execute proc_test⦅42⦆ ·
		execute proc_test⦅99⦆ ·

		yield res ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("expected dataset, quasinary, and AGI constructs to pass, got errors: %v", errs)
	}
}

func TestChecker_CIR_Quantum_Prime_Coverage(t *testing.T) {
	src := `
	process main⦅⦆ → int64 ⟦
		confident⟨float64⟩ a ↦ 0.90 ∵ 10.0 ·
		confident⟨float64⟩ b ↦ 0.80 ∵ 20.0 ·
		confident⟨float64⟩ c ↦ a ⊗ b ·
		assert c ≽ 0.70 ·

		qubit q0 ⸴ q1 ·
		superpose q0 ·
		entangle q0 ⸴ q1 ·
		int64 ma ↦ collapse q0 ·
		int64 mb ↦ collapse q1 ·

		concept c1 ↦ 2 ·
		concept c2 ↦ 3 ·
		concept c3 ↦ c1 ⊗ c2 ·
		tensor⟨int64⟩ factors ↦ decompose⦅c3⦆ ·

		yield ma ·
	⟧`
	_, errs := parseAndCheck(src)
	if len(errs) > 0 {
		t.Fatalf("expected CIR, quantum, and prime symbolic constructs to pass, got errors: %v", errs)
	}
}
