package codegen

import (
	"strings"
	"testing"
)

func TestCIRExecution(t *testing.T) {
	t.Parallel()
	src := `
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.90 ∵ 10.0 ·
    confident⟨float64⟩ b ↦ 0.80 ∵ 20.0 ·
    confident⟨float64⟩ c ↦ a ⊗ b ·
    emit c ·
    assert c ≽ 0.70 ·
    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "conf⟨value=200.000000, score=0.720000, state=known⟩") {
		t.Errorf("expected value=200 and score=0.72 in output, got: %s", out)
	}
}

func TestMemoryExecution(t *testing.T) {
	t.Parallel()
	src := `
process main ⦅ ⦆ → int64 ⟦
    memory mem ⟬ 100 ⟭ ·
    remember mem⟬"answer"⟭ ↦ 42 ·
    int64 val ↦ recall mem⟬"answer"⟭ ·
    emit val ·
    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "42") {
		t.Errorf("expected 42 in output, got: %s", out)
	}
}

func TestGraphExecution(t *testing.T) {
	t.Parallel()
	src := `
process main ⦅ ⦆ → int64 ⟦
    graph g ·
    node g ⟬ "nodeA" ↦ 101 ⟭ ·
    node g ⟬ "nodeB" ↦ 102 ⟭ ·
    link g ⟬ "nodeA" ⸴ "nodeB" ⸴ 1.0 ⟭ ·
    tensor⟨int64⟩ tr ↦ trace g ⟬ "nodeA" ⸴ 2 ⟭ ·
    emit tr ·
    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "101") || !strings.Contains(out, "102") {
		t.Errorf("expected visited nodes 101 and 102 in output, got: %s", out)
	}
}

func TestProcedureExecution(t *testing.T) {
	t.Parallel()
	src := `
procedure add_ten ⦅ int64 x ⦆ → int64 ⟦
    yield x ⊕ 10 ·
⟧

process main ⦅ ⦆ → int64 ⟦
    int64 res ↦ execute add_ten ⦅ 25 ⦆ ·
    emit res ·
    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "35") {
		t.Errorf("expected 35 in output, got: %s", out)
	}
}

func TestQEEHExecution(t *testing.T) {
	t.Parallel()
	src := `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 1.0 ·
    ⟧

    enforce safety on "help_user" ·
    emit "Action allowed" ·

    enforce safety on "harm_user" ·
    emit "Action blocked" ·

    decoherence ·

    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "Action allowed") || !strings.Contains(out, "Action blocked") {
		t.Errorf("expected Action allowed and Action blocked, got: %s", out)
	}
}

func TestQEEHDecoherenceStatement(t *testing.T) {
	t.Parallel()
	src := `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 1.0 ·
    ⟧

    enforce safety on "help_user" ·
    emit "Action allowed" ·

    decoherence ·
    emit "Decoherence active" ·

    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "Action allowed") || !strings.Contains(out, "Decoherence active") {
		t.Errorf("expected Decoherence active, got: %s", out)
	}
}

func TestQEMWVCodegen(t *testing.T) {
	t.Parallel()
	src := `
network SecretModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
⟧

process main ⦅ ⦆ → int64 ⟦
    ethical_boundary auth ⟦
        rule user_authorized ↦ 1.0 ·
    ⟧

    vault v ⟬ SecretModel ⟭ threshold ↦ 3 ·
    seal v ·
    emit "Weights sealed in vault" ·

    unseal v with auth ·
    emit "Weights unsealed for inference" ·

    wipe v ·
    emit "Vault wiped" ·

    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "Weights sealed in vault") || !strings.Contains(out, "Weights unsealed for inference") || !strings.Contains(out, "Vault wiped") {
		t.Errorf("expected vault steps, got: %s", out)
	}
}

func TestMCTPCodegen(t *testing.T) {
	t.Parallel()
	src := `
network ProvenanceModel ⟦
    dense ⟬2⸴ 4⟭ ·
    relu ·
    dense ⟬4⸴ 1⟭ ·
⟧

process main ⦅ ⦆ → int64 ⟦
    tensor⟨float64⟩⟬4⸴ 2⟭ x ↦ tensor⟬⟬0.0⸴ 0.0⟭⸴ ⟬0.0⸴ 1.0⟭⸴ ⟬1.0⸴ 0.0⟭⸴ ⟬1.0⸴ 1.0⟭⟭ ·
    tensor⟨float64⟩⟬4⸴ 1⟭ y ↦ tensor⟬⟬0.0⟭⸴ ⟬1.0⟭⸴ ⟬1.0⟭⸴ ⟬0.0⟭⟭ ·

    provenance p ⟬ ProvenanceModel ⟭ ·

    learn ProvenanceModel on x ⸴ y ⟦
        loss mse ·
        optimizer adam ·
        lr ↦ 0.1 ·
        epochs ↦ 5 ·
    ⟧

    certify p ·
    emit "Training certified" ·

    yield 0 ·
⟧
`
	out := compileAndRunSource(t, src)
	if !strings.Contains(out, "Training certified") {
		t.Errorf("expected Training certified, got: %s", out)
	}
}
