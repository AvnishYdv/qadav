package runtime

import (
	"strings"
	"testing"
)

func TestCertificateCommitmentASLR(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_commitment_aslr")
	if !strings.Contains(out, "CERT_COMMITMENT_ASLR_OK") {
		t.Fatalf("Certificate commitment ASLR determinism failed: %s", out)
	}
	t.Logf("Certificate commitment ASLR verified: %s", out)
}

func TestCertificateChainVerifyASLR(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_chain_verify_aslr")
	if !strings.Contains(out, "CERT_CHAIN_VERIFY_ASLR_OK") {
		t.Fatalf("Certificate chain verify ASLR determinism failed: %s", out)
	}
	t.Logf("Certificate chain verify ASLR verified: %s", out)
}

func TestCertificateLifecycle(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_lifecycle")
	if !strings.Contains(out, "CERT_LIFECYCLE_OK") {
		t.Fatalf("Certificate lifecycle tracking failed: %s", out)
	}
	t.Logf("Certificate lifecycle verified: %s", out)
}

func TestAbstainLifecycle(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_lifecycle")
	if !strings.Contains(out, "ABSTAIN_LIFECYCLE_OK") {
		t.Fatalf("Abstain lifecycle tracking failed: %s", out)
	}
	t.Logf("Abstain lifecycle verified: %s", out)
}
