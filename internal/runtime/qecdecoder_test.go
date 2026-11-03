package runtime

import (
	"strings"
	"testing"
)

func TestQECDDecodeBasic(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_decode_basic")
	if !strings.Contains(out, "QECD_DECODE_BASIC_OK") {
		t.Fatalf("qecd_decode_basic failed: %s", out)
	}
	t.Logf("qecd_decode_basic verified: %s", out)
}

func TestQECDDecodeEmptySyndrome(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_decode_empty")
	if !strings.Contains(out, "QECD_DECODE_EMPTY_OK") {
		t.Fatalf("qecd_decode_empty failed: %s", out)
	}
	t.Logf("qecd_decode_empty verified: %s", out)
}

func TestQECDDecodeLargeSyndrome(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_decode_large")
	if !strings.Contains(out, "QECD_DECODE_LARGE_OK") {
		t.Fatalf("qecd_decode_large failed: %s", out)
	}
	t.Logf("qecd_decode_large verified: %s", out)
}

func TestQECDVerifyAttestation(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_verify_attestation")
	if !strings.Contains(out, "QECD_VERIFY_ATTESTATION_OK") {
		t.Fatalf("qecd_verify_attestation failed: %s", out)
	}
	t.Logf("qecd_verify_attestation verified: %s", out)
}

func TestQECDVerifyTampered(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_verify_tampered")
	if !strings.Contains(out, "QECD_VERIFY_TAMPERED_OK") {
		t.Fatalf("qecd_verify_tampered failed: %s", out)
	}
	t.Logf("qecd_verify_tampered verified: %s", out)
}

func TestQECDApplyCorrections(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_apply_corrections")
	if !strings.Contains(out, "QECD_APPLY_CORRECTIONS_OK") {
		t.Fatalf("qecd_apply_corrections failed: %s", out)
	}
	t.Logf("qecd_apply_corrections verified: %s", out)
}

func TestQECDApplyBoundsCheck(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_apply_bounds")
	if !strings.Contains(out, "QECD_APPLY_BOUNDS_OK") {
		t.Fatalf("qecd_apply_bounds failed: %s", out)
	}
	t.Logf("qecd_apply_bounds verified: %s", out)
}

func TestQECDASLRDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "qecd_aslr_determinism")
	if !strings.Contains(out, "QECD_ASLR_DETERMINISM_OK") {
		t.Fatalf("qecd_aslr_determinism failed: %s", out)
	}
	t.Logf("qecd_aslr_determinism verified: %s", out)
}
