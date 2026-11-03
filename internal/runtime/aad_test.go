package runtime

import (
	"strings"
	"testing"
)

func TestAADBuildCert(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_build_cert")
	if !strings.Contains(out, "AAD_BUILD_CERT_OK") {
		t.Fatalf("aad_build_cert failed: %s", out)
	}
	t.Logf("aad_build_cert verified: %s", out)
}

func TestAADBuildCertInvalidBounds(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_build_cert_invalid")
	if !strings.Contains(out, "AAD_BUILD_CERT_INVALID_OK") {
		t.Fatalf("aad_build_cert_invalid failed: %s", out)
	}
	t.Logf("aad_build_cert_invalid verified: %s", out)
}

func TestAADWrapDecodeValid(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_wrap_decode_valid")
	if !strings.Contains(out, "AAD_WRAP_DECODE_VALID_OK") {
		t.Fatalf("aad_wrap_decode_valid failed: %s", out)
	}
	t.Logf("aad_wrap_decode_valid verified: %s", out)
}

func TestAADWrapDecodeOutOfBounds(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_wrap_decode_outofbounds")
	if !strings.Contains(out, "AAD_WRAP_DECODE_OUTOFBOUNDS_OK") {
		t.Fatalf("aad_wrap_decode_outofbounds failed: %s", out)
	}
	t.Logf("aad_wrap_decode_outofbounds verified: %s", out)
}

func TestAADVerifyValid(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_verify_valid")
	if !strings.Contains(out, "AAD_VERIFY_VALID_OK") {
		t.Fatalf("aad_verify_valid failed: %s", out)
	}
	t.Logf("aad_verify_valid verified: %s", out)
}

func TestAADVerifyTampered(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_verify_tampered")
	if !strings.Contains(out, "AAD_VERIFY_TAMPERED_OK") {
		t.Fatalf("aad_verify_tampered failed: %s", out)
	}
	t.Logf("aad_verify_tampered verified: %s", out)
}

func TestAADAttestSuccess(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_attest_success")
	if !strings.Contains(out, "AAD_ATTEST_SUCCESS_OK") {
		t.Fatalf("aad_attest_success failed: %s", out)
	}
	t.Logf("aad_attest_success verified: %s", out)
}

func TestAADAttestTampered(t *testing.T) {
	out, err := runCryptoTestHarnessExpectExit(t, "aad_attest_tampered")
	if err == nil {
		t.Fatalf("expected AAD attest tamper decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("AAD attest tampered detected and decoherence triggered: %v", err)
}

func TestAADMetaFitAdaptVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_meta_fit_adapt_verify")
	if !strings.Contains(out, "AAD_META_FIT_ADAPT_VERIFY_OK") {
		t.Fatalf("aad_meta_fit_adapt_verify failed: %s", out)
	}
	t.Logf("aad_meta_fit_adapt_verify verified: %s", out)
}

func TestAADMetaAdaptTamperedEnvelope(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_meta_adapt_tampered")
	if !strings.Contains(out, "AAD_META_ADAPT_TAMPERED_OK") {
		t.Fatalf("aad_meta_adapt_tampered failed: %s", out)
	}
	t.Logf("aad_meta_adapt_tampered verified: %s", out)
}

func TestAADAdaptChainDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_adapt_chain_determinism")
	if !strings.Contains(out, "AAD_ADAPT_CHAIN_DETERMINISM_OK") {
		t.Fatalf("aad_adapt_chain_determinism failed: %s", out)
	}
	t.Logf("aad_adapt_chain_determinism verified: %s", out)
}

func TestAADASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "aad_aslr_determinism")
	if !strings.Contains(out, "AAD_ASLR_DETERMINISM_OK") {
		t.Fatalf("aad_aslr_determinism failed: %s", out)
	}
	t.Logf("aad_aslr_determinism verified: %s", out)
}
