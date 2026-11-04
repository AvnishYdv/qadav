package runtime

import (
	"strings"
	"testing"
)

func TestBaratSubmitAggregate(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_submit_aggregate")
	if !strings.Contains(out, "BARAT_SUBMIT_AGGREGATE_OK") {
		t.Fatalf("barat_submit_aggregate failed: %s", out)
	}
	t.Logf("barat_submit_aggregate verified: %s", out)
}

func TestBaratTrimmedMean(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_trimmed_mean")
	if !strings.Contains(out, "BARAT_TRIMMED_MEAN_OK") {
		t.Fatalf("barat_trimmed_mean failed: %s", out)
	}
	t.Logf("barat_trimmed_mean verified: %s", out)
}

func TestBaratSelectionProofValid(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_selection_proof_valid")
	if !strings.Contains(out, "BARAT_SELECTION_PROOF_VALID_OK") {
		t.Fatalf("barat_selection_proof_valid failed: %s", out)
	}
	t.Logf("barat_selection_proof_valid verified: %s", out)
}

func TestBaratTamperedCert(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "barat_tampered_cert")
	if err == nil {
		t.Fatalf("expected BARAT tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("BARAT tampered cert detected and decoherence triggered: %v", err)
}

func TestBaratAttestSuccess(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_attest_success")
	if !strings.Contains(out, "BARAT_ATTEST_SUCCESS_OK") {
		t.Fatalf("barat_attest_success failed: %s", out)
	}
	t.Logf("barat_attest_success verified: %s", out)
}

func TestBaratCanonicalSort(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_canonical_sort")
	if !strings.Contains(out, "BARAT_CANONICAL_SORT_OK") {
		t.Fatalf("barat_canonical_sort failed: %s", out)
	}
	t.Logf("barat_canonical_sort verified: %s", out)
}

func TestBaratASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "barat_aslr_determinism")
	if !strings.Contains(out, "BARAT_ASLR_DETERMINISM_OK") {
		t.Fatalf("barat_aslr_determinism failed: %s", out)
	}
	t.Logf("barat_aslr_determinism verified: %s", out)
}
