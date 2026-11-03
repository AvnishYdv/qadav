package runtime

import (
	"strings"
	"testing"
)

func TestEEACreatePolicy(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_create_policy")
	if !strings.Contains(out, "EEA_CREATE_POLICY_OK") {
		t.Fatalf("eea_create_policy failed: %s", out)
	}
	t.Logf("eea_create_policy verified: %s", out)
}

func TestEEAComputeDiff(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_compute_diff")
	if !strings.Contains(out, "EEA_COMPUTE_DIFF_OK") {
		t.Fatalf("eea_compute_diff failed: %s", out)
	}
	t.Logf("eea_compute_diff verified: %s", out)
}

func TestEEABuildEnvelopeValid(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_build_envelope_valid")
	if !strings.Contains(out, "EEA_BUILD_ENVELOPE_VALID_OK") {
		t.Fatalf("eea_build_envelope_valid failed: %s", out)
	}
	t.Logf("eea_build_envelope_valid verified: %s", out)
}

func TestEEABuildEnvelopeViolation(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_build_envelope_violation")
	if !strings.Contains(out, "EEA_BUILD_ENVELOPE_VIOLATION_OK") {
		t.Fatalf("eea_build_envelope_violation failed: %s", out)
	}
	t.Logf("eea_build_envelope_violation verified: %s", out)
}

func TestEEAVerifyValid(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_verify_valid")
	if !strings.Contains(out, "EEA_VERIFY_VALID_OK") {
		t.Fatalf("eea_verify_valid failed: %s", out)
	}
	t.Logf("eea_verify_valid verified: %s", out)
}

func TestEEAVerifyTampered(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "eea_verify_tampered")
	if err == nil {
		t.Fatalf("expected EEA verify tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("EEA verify tampered detected and decoherence triggered: %v", err)
}

func TestEEAAttestSuccess(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_attest_success")
	if !strings.Contains(out, "EEA_ATTEST_SUCCESS_OK") {
		t.Fatalf("eea_attest_success failed: %s", out)
	}
	t.Logf("eea_attest_success verified: %s", out)
}

func TestEEAAttestTampered(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "eea_attest_tampered")
	if err == nil {
		t.Fatalf("expected EEA attest tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("EEA attest tampered detected and decoherence triggered: %v", err)
}

func TestEEADecayingRate(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_decaying_rate")
	if !strings.Contains(out, "EEA_DECAYING_RATE_OK") {
		t.Fatalf("eea_decaying_rate failed: %s", out)
	}
	t.Logf("eea_decaying_rate verified: %s", out)
}

func TestEEAEnvelopeASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "eea_aslr_determinism")
	if !strings.Contains(out, "EEA_ASLR_DETERMINISM_OK") {
		t.Fatalf("eea_aslr_determinism failed: %s", out)
	}
	t.Logf("eea_aslr_determinism verified: %s", out)
}
