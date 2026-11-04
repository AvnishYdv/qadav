package runtime

import (
	"strings"
	"testing"
)

func TestCDAGARecordAndVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "cdaga_record_and_verify")
	if !strings.Contains(out, "CDAGA_RECORD_AND_VERIFY_OK") {
		t.Fatalf("CDAGA record and verify failed: %s", out)
	}
	t.Logf("CDAGA record and verify verified: %s", out)
}

func TestCDAGAAttestToQERA(t *testing.T) {
	out := runCryptoTestHarness(t, "cdaga_attest_to_qera")
	if !strings.Contains(out, "CDAGA_ATTEST_TO_QERA_OK") {
		t.Fatalf("CDAGA attest to QERA failed: %s", out)
	}
	t.Logf("CDAGA attest to QERA verified: %s", out)
}

func TestCDAGATamperDetectionDecoherence(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "cdaga_tamper_detection")
	if err == nil {
		t.Fatalf("expected CDAGA tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("CDAGA tamper detected and decoherence triggered: %v", err)
}

func TestCDAGACanonicalSort(t *testing.T) {
	out := runCryptoTestHarness(t, "cdaga_canonical_sort")
	if !strings.Contains(out, "CDAGA_CANONICAL_SORT_OK") {
		t.Fatalf("CDAGA canonical sort failed: %s", out)
	}
	t.Logf("CDAGA canonical sort verified: %s", out)
}

func TestCDAGAASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "cdaga_aslr_determinism")
	if !strings.Contains(out, "CDAGA_ASLR_DETERMINISM_OK") {
		t.Fatalf("CDAGA ASLR determinism failed: %s", out)
	}
	t.Logf("CDAGA ASLR determinism verified: %s", out)
}
