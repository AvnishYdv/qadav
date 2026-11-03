package checker

import (
	"testing"
)

func TestCIRChecker(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "valid assert with confident literal",
			src: `
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.95 ∵ 42.0 ·
    assert a ≽ 0.90 ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "assert threshold out of bounds",
			src: `
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.95 ∵ 42.0 ·
    assert a ≽ 1.50 ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "confident literal propagation math",
			src: `
process main ⦅ ⦆ → int64 ⟦
    confident⟨float64⟩ a ↦ 0.90 ∵ 10.0 ·
    confident⟨float64⟩ b ↦ 0.80 ∵ 20.0 ·
    confident⟨float64⟩ c ↦ a ⊗ b ·
    confident⟨float64⟩ d ↦ a ⊕ b ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestEthicsChecker(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "valid ethical boundary and enforce",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 1.0 ·
    ⟧
    enforce safety on "help_user" ·
    decoherence ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "invalid rule penalty out of range",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary safety ⟦
        rule human_harm ↦ 2.5 ·
    ⟧
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "enforce on non-boundary target",
			src: `
process main ⦅ ⦆ → int64 ⟦
    int64 not_boundary ↦ 42 ·
    enforce not_boundary on "action" ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestQEEHSemanticAction(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "typed semantic action check",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary b ⟦
        rule safe_action ↦ 0.8 ·
    ⟧
    enforce b on "harm_user" ·
    enforce b on "help_user" ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "enforce with non-boundary",
			src: `
process main ⦅ ⦆ → int64 ⟦
    text msg ↦ "not_a_boundary" ·
    enforce msg on "action" ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestAGIChecker(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "memory remember and recall",
			src: `
process main ⦅ ⦆ → int64 ⟦
    memory mem ⟬ 100 ⟭ ·
    remember mem⟬"key1"⟭ ↦ 42 ·
    int64 val ↦ recall mem⟬"key1"⟭ ·
    yield val ·
⟧
`,
			wantErr: false,
		},
		{
			name: "graph node link and trace",
			src: `
process main ⦅ ⦆ → int64 ⟦
    graph g ·
    node g ⟬ "conceptA" ↦ 10 ⟭ ·
    node g ⟬ "conceptB" ↦ 20 ⟭ ·
    link g ⟬ "conceptA" ⸴ "conceptB" ⸴ 1.0 ⟭ ·
    tensor⟨int64⟩ visited ↦ trace g ⟬ "conceptA" ⸴ 2 ⟭ ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "procedure declaration and execute",
			src: `
procedure add_one ⦅ int64 x ⦆ → int64 ⟦
    yield x ⊕ 1 ·
⟧

process main ⦅ ⦆ → int64 ⟦
    int64 res ↦ execute add_one ⦅ 5 ⦆ ·
    yield res ·
⟧
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestVaultAndProvenanceChecker(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "valid vault and seal unseal wipe",
			src: `
network Net ⟦
    dense ⟬2⸴ 2⟭ ·
⟧
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary auth ⟦
        rule ok ↦ 1.0 ·
    ⟧
    vault v ⟬ Net ⟭ threshold ↦ 3 ·
    seal v ·
    unseal v with auth ·
    wipe v ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "invalid vault threshold",
			src: `
network Net ⟦
    dense ⟬2⸴ 2⟭ ·
⟧
process main ⦅ ⦆ → int64 ⟦
    vault v ⟬ Net ⟭ threshold ↦ 10 ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "invalid seal target",
			src: `
process main ⦅ ⦆ → int64 ⟦
    int64 x ↦ 5 ·
    seal x ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "invalid unseal target and boundary",
			src: `
process main ⦅ ⦆ → int64 ⟦
    int64 x ↦ 5 ·
    int64 b ↦ 10 ·
    unseal x with b ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "invalid wipe target",
			src: `
process main ⦅ ⦆ → int64 ⟦
    int64 x ↦ 5 ·
    wipe x ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "invalid certify target",
			src: `
process main ⦅ ⦆ → int64 ⟦
    int64 x ↦ 5 ·
    certify x ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "valid verify_training expression",
			src: `
process main ⦅ ⦆ → int64 ⟦
    text cert ↦ "cert.mctp" ·
    logic ok ↦ verify_training ⟬ cert ⟭ ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "valid provenance and certify",
			src: `
network Net ⟦
    dense ⟬2⸴ 2⟭ ·
⟧
process main ⦅ ⦆ → int64 ⟦
    provenance p ⟬ Net ⟭ ·
    certify p ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}

func TestQECMAChecker(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{
			name: "valid qecma cross modal reasoning",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦
        capacity ↦ 50 ·
    ⟧
    tensor⟨float64⟩⟬4⟭ img ↦ tensor⟬1.0⸴ 2.0⸴ 3.0⸴ 4.0⟭ ·
    text prompt ↦ "analyze this" ·
    int64 concept_id ↦ 101 ·
    modality m1 as tensor_input from img ·
    modality m2 as text_input from prompt ·
    modality m3 as concept_input from concept_id ·
    cross_modal_reason rc deduce from ⟬m1⸴ m2⸴ m3⟭ to "decision" with 0.95 ·
    verify_chain rc ·
    yield 0 ·
⟧
`,
			wantErr: false,
		},
		{
			name: "cross_modal_reason invalid non-modality input",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦
        capacity ↦ 50 ·
    ⟧
    int64 not_a_modality ↦ 42 ·
    cross_modal_reason rc deduce from ⟬not_a_modality⟭ to "decision" with 0.95 ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
		{
			name: "cross_modal_reason confidence out of range",
			src: `
process main ⦅ ⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦
        capacity ↦ 50 ·
    ⟧
    text prompt ↦ "analyze" ·
    modality m1 as text_input from prompt ·
    cross_modal_reason rc deduce from ⟬m1⟭ to "decision" with 1.5 ·
    yield 0 ·
⟧
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := parseAndCheck(tt.src)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("unexpected error: %v", errs)
			}
		})
	}
}
