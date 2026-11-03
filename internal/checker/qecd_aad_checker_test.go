package checker

import (
	"strings"
	"testing"
)

func runChecker(src string, isTest bool) []string {
	_, errs := parseAndCheckTestMode(src, isTest)
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Msg)
	}
	return msgs
}

func TestQECDCheckerValid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    qec_memory qm as surface_code distance ↦ 3 with b ⟦ ⟧ ·
    syndrome qm into syn0 ·
    text decoded ↦ "" ·
    qecd_decode distance ↦ 3 from syn0 into decoded ·
    logic ok ↦ deny ·
    qecd_verify decoded into ok ·
    text state_vec ↦ "sv" ·
    qecd_apply decoded on state_vec ·
    yield 0 ·
⟧`
	errs := runChecker(src, false)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestQECDCheckerInvalidDistance(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    qec_memory qm as surface_code distance ↦ 3 with b ⟦ ⟧ ·
    syndrome qm into syn0 ·
    text decoded ↦ "" ·
    qecd_decode distance ↦ 4 from syn0 into decoded ·
    yield 0 ·
⟧`
	errs := runChecker(src, false)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "code distance must be 3, 5, or 7") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected distance error, got %v", errs)
	}
}

func TestQECDCheckerUndefinedVars(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text decoded ↦ "" ·
    qecd_decode distance ↦ 3 from syn_missing into decoded ·
    logic ok ↦ deny ·
    qecd_verify dec_missing into ok ·
    qecd_apply dec_missing on vec_missing ·
    yield 0 ·
⟧`
	errs := runChecker(src, false)
	if len(errs) < 3 {
		t.Fatalf("expected at least 3 errors for undefined vars, got %v", errs)
	}
}

func TestAADCheckerValid(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    ethical_boundary b ⟦ rule r ↦ 1.0 · ⟧
    reasoning_chain rc with b ⟦ capacity ↦ 10 · ⟧
    text m ↦ "model" ·
    text tr ↦ "train" ·
    text cert ↦ "" ·
    aad_cert with model ↦ m train ↦ tr max_syndrome ↦ 64 max_corr_dist ↦ 8 into cert ·
    text dec ↦ "decoded" ·
    text env ↦ "" ·
    aad_attest cert cert ↦ cert decode ↦ dec into env ·
    text state ↦ "" ·
    meta_fit into state cert ↦ cert ·
    text adapted ↦ "" ·
    meta_adapt state with env into adapted ·
    logic ok ↦ deny ·
    meta_verify state into ok ·
    corrupt_aad_env env at ↦ 16 ·
    yield 0 ·
⟧`
	errs := runChecker(src, true)
	if len(errs) > 0 {
		t.Fatalf("unexpected checker errors: %v", errs)
	}
}

func TestAADCheckerCorruptWithoutTestMode(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text env ↦ "env" ·
    corrupt_aad_env env at ↦ 16 ·
    yield 0 ·
⟧`
	errs := runChecker(src, false)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "test mode (--test)") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected test flag error, got %v", errs)
	}
}

func TestAADCheckerInvalidBounds(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text m ↦ "model" ·
    text tr ↦ "train" ·
    text cert ↦ "" ·
    aad_cert with model ↦ m train ↦ tr max_syndrome ↦ 200 max_corr_dist ↦ 32 into cert ·
    yield 0 ·
⟧`
	errs := runChecker(src, false)
	if len(errs) < 2 {
		t.Fatalf("expected 2 bounds errors, got %v", errs)
	}
}

func TestAADCheckerUndefinedVars(t *testing.T) {
	src := `process main⦅⦆ → int64 ⟦
    text env ↦ "" ·
    aad_attest cert_missing cert ↦ cert_missing decode ↦ dec_missing into env ·
    text st ↦ "" ·
    meta_fit into st cert ↦ cert_missing ·
    text ad ↦ "" ·
    meta_adapt st_missing with env_missing into ad ·
    logic ok ↦ deny ·
    meta_verify st_missing into ok ·
    corrupt_aad_env env_missing at ↦ 16 ·
    yield 0 ·
⟧`
	errs := runChecker(src, true)
	if len(errs) < 5 {
		t.Fatalf("expected at least 5 undefined var errors, got %v", errs)
	}
}
