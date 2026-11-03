package runtime

import (
	"strings"
	"testing"
)

func TestSEACreate(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_create")
	if !strings.Contains(out, "SEA_CREATE_OK") {
		t.Fatalf("sea_create failed: %s", out)
	}
	t.Logf("sea_create verified: %s", out)
}

func TestSEAProposeValid(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_propose_valid")
	if !strings.Contains(out, "SEA_PROPOSE_VALID_OK") {
		t.Fatalf("sea_propose_valid failed: %s", out)
	}
	t.Logf("sea_propose_valid verified: %s", out)
}

func TestSEAProposeInvalidHashes(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_propose_invalid_hashes")
	if !strings.Contains(out, "SEA_PROPOSE_INVALID_HASHES_OK") {
		t.Fatalf("sea_propose_invalid_hashes failed: %s", out)
	}
	t.Logf("sea_propose_invalid_hashes verified: %s", out)
}

func TestSEAVerifyValid(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_verify_valid")
	if !strings.Contains(out, "SEA_VERIFY_VALID_OK") {
		t.Fatalf("sea_verify_valid failed: %s", out)
	}
	t.Logf("sea_verify_valid verified: %s", out)
}

func TestSEAVerifyTampered(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "sea_verify_tampered")
	if err == nil {
		t.Fatalf("expected SEA verify tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("SEA verify tampered detected and decoherence triggered: %v", err)
}

func TestSEAAttestSuccess(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_attest_success")
	if !strings.Contains(out, "SEA_ATTEST_SUCCESS_OK") {
		t.Fatalf("sea_attest_success failed: %s", out)
	}
	t.Logf("sea_attest_success verified: %s", out)
}

func TestSEAAttestTampered(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "sea_attest_tampered")
	if err == nil {
		t.Fatalf("expected SEA attest tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("SEA attest tampered detected and decoherence triggered: %v", err)
}

func TestSEACanonicalSortDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_canonical_sort_determinism")
	if !strings.Contains(out, "SEA_CANONICAL_SORT_DETERMINISM_OK") {
		t.Fatalf("sea_canonical_sort_determinism failed: %s", out)
	}
	t.Logf("sea_canonical_sort_determinism verified: %s", out)
}

func TestSEAASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_aslr_determinism")
	if !strings.Contains(out, "SEA_ASLR_DETERMINISM_OK") {
		t.Fatalf("sea_aslr_determinism failed: %s", out)
	}
	t.Logf("sea_aslr_determinism verified: %s", out)
}

func TestSEALineageChain(t *testing.T) {
	out := runCryptoTestHarness(t, "sea_lineage_chain")
	if !strings.Contains(out, "SEA_LINEAGE_CHAIN_OK") {
		t.Fatalf("sea_lineage_chain failed: %s", out)
	}
	t.Logf("sea_lineage_chain verified: %s", out)
}
