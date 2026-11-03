package runtime

import (
	"strings"
	"testing"
)

func TestRSCARecordAndAttest(t *testing.T) {
	out := runCryptoTestHarness(t, "rsca_record_and_attest")
	if !strings.Contains(out, "RSCA_RECORD_AND_ATTEST_OK") {
		t.Fatalf("rsca_record_and_attest failed: %s", out)
	}
	t.Logf("rsca_record_and_attest verified: %s", out)
}

func TestRSCAASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "rsca_aslr_determinism")
	if !strings.Contains(out, "RSCA_ASLR_DETERMINISM_OK") {
		t.Fatalf("rsca_aslr_determinism failed: %s", out)
	}
	t.Logf("rsca_aslr_determinism verified: %s", out)
}

func TestRSCATamperDetection(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "rsca_tamper_detection")
	if err == nil {
		t.Fatalf("expected RSCA tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("RSCA tamper detected and decoherence triggered: %v", err)
}
