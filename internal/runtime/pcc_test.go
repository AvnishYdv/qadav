package runtime

import (
	"strings"
	"testing"
)

func TestPCCCreationAndVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "pcc_creation_and_verify")
	if !strings.Contains(out, "PCC_CREATION_AND_VERIFY_OK") {
		t.Fatalf("pcc_creation_and_verify failed: %s", out)
	}
	t.Logf("pcc_creation_and_verify verified: %s", out)
}

func TestPCCGetStatusUnattested(t *testing.T) {
	out := runCryptoTestHarness(t, "pcc_verify_fail_unattested")
	if !strings.Contains(out, "PCC_VERIFY_FAIL_UNATTESTED_OK") {
		t.Fatalf("pcc_verify_fail_unattested failed: %s", out)
	}
	t.Logf("pcc_verify_fail_unattested verified: %s", out)
}

func TestPCCBasicTamper(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "pcc_tamper_detection")
	if err == nil {
		t.Fatalf("expected PCC tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("PCC tamper detected and decoherence triggered: %v", err)
}
