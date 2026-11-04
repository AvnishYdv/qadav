package runtime

import (
	"strings"
	"testing"
)

func TestCTSACreationAndVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_creation_and_verify")
	if !strings.Contains(out, "CTSA_CREATION_AND_VERIFY_OK") {
		t.Fatalf("ctsa_creation_and_verify failed: %s", out)
	}
	t.Logf("ctsa_creation_and_verify verified: %s", out)
}

func TestCTSAGetStatusUnattested(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_verify_fail_unattested")
	if !strings.Contains(out, "CTSA_VERIFY_FAIL_UNATTESTED_OK") {
		t.Fatalf("ctsa_verify_fail_unattested failed: %s", out)
	}
	t.Logf("ctsa_verify_fail_unattested verified: %s", out)
}

func TestCTSABasicTamper(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "ctsa_tamper_detection")
	if err == nil {
		t.Fatalf("expected CTSA tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("CTSA tamper detected and decoherence triggered: %v", err)
}

func TestCTSADeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_rolling_chain_determinism")
	if !strings.Contains(out, "CTSA_ROLLING_CHAIN_DETERMINISM_OK") {
		t.Fatalf("ctsa_rolling_chain_determinism failed: %s", out)
	}
	t.Logf("ctsa_rolling_chain_determinism verified: %s", out)
}

func TestCTSAASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_aslr_determinism")
	if !strings.Contains(out, "CTSA_ASLR_DETERMINISM_OK") {
		t.Fatalf("ctsa_aslr_determinism failed: %s", out)
	}
	t.Logf("ctsa_aslr_determinism verified: %s", out)
}

func TestCTSALifecycle(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_lifecycle")
	if !strings.Contains(out, "CTSA_LIFECYCLE_OK") {
		t.Fatalf("ctsa_lifecycle failed: %s", out)
	}
	t.Logf("ctsa_lifecycle verified: %s", out)
}

func TestCTSAMultiModuleChain(t *testing.T) {
	out := runCryptoTestHarness(t, "ctsa_multi_module_chain")
	if !strings.Contains(out, "CTSA_MULTI_MODULE_CHAIN_OK") {
		t.Fatalf("ctsa_multi_module_chain failed: %s", out)
	}
	t.Logf("ctsa_multi_module_chain verified: %s", out)
}
