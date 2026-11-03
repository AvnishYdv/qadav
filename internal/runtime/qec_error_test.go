package runtime

import (
	"strings"
	"testing"
)

func TestQECOverflowSyndrome(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_overflow_syndrome")
	if !strings.Contains(out, "QEC_OVERFLOW_SYNDROME_OK") {
		t.Fatalf("QEC overflow syndrome test failed: %s", out)
	}
	t.Logf("QEC overflow syndrome verified: %s", out)
}

func TestQECZeroSyndrome(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_zero_syndrome")
	if !strings.Contains(out, "QEC_ZERO_SYNDROME_OK") {
		t.Fatalf("QEC zero syndrome test failed: %s", out)
	}
	t.Logf("QEC zero syndrome verified: %s", out)
}

func TestQECVerifySuccess(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_verify_success")
	if !strings.Contains(out, "QEC_VERIFY_SUCCESS_OK") {
		t.Fatalf("QEC verify success test failed: %s", out)
	}
	t.Logf("QEC verify success verified: %s", out)
}

func TestQECCorrectAfterVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_correct_after_verify")
	if !strings.Contains(out, "QEC_CORRECT_AFTER_VERIFY_OK") {
		t.Fatalf("QEC correct after verify test failed: %s", out)
	}
	t.Logf("QEC correct after verify verified: %s", out)
}

func TestQECMaxSyndromes(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_max_syndromes")
	if !strings.Contains(out, "QEC_MAX_SYNDROMES_OK") {
		t.Fatalf("QEC max syndromes test failed: %s", out)
	}
	t.Logf("QEC max syndromes verified: %s", out)
}

func TestScopedExpiredToken(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_expired_token")
	if !strings.Contains(out, "SCOPED_EXPIRED_TOKEN_OK") {
		t.Fatalf("Scoped expired token test failed: %s", out)
	}
	t.Logf("Scoped expired token verified: %s", out)
}

func TestScopedRecipientMismatch(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_recipient_mismatch")
	if !strings.Contains(out, "SCOPED_RECIPIENT_MISMATCH_OK") {
		t.Fatalf("Scoped recipient mismatch test failed: %s", out)
	}
	t.Logf("Scoped recipient mismatch verified: %s", out)
}

func TestScopedBoundaryNotDisclosed(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_boundary_not_disclosed")
	if !strings.Contains(out, "SCOPED_BOUNDARY_NOT_DISCLOSED_OK") {
		t.Fatalf("Scoped boundary not disclosed test failed: %s", out)
	}
	t.Logf("Scoped boundary not disclosed verified: %s", out)
}

