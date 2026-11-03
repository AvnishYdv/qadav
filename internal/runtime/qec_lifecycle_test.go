package runtime

import (
	"strings"
	"testing"
)

func TestQECFreeOnNormalExit(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_free_normal")
	if !strings.Contains(out, "QEC_FREE_NORMAL_OK") {
		t.Fatalf("QEC free on normal exit failed: %s", out)
	}
	t.Logf("QEC free on normal exit verified: %s", out)
}

func TestQECFreeOnVerifyFailure(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_free_verify_failure")
	if !strings.Contains(out, "QEC_FREE_VERIFY_FAILURE_OK") {
		t.Fatalf("QEC free on verify failure failed: %s", out)
	}
	t.Logf("QEC free on verify failure verified: %s", out)
}

func TestQECFreeOnCorrectFailure(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_free_correct_failure")
	if !strings.Contains(out, "QEC_FREE_CORRECT_FAILURE_OK") {
		t.Fatalf("QEC free on correct failure failed: %s", out)
	}
	t.Logf("QEC free on correct failure verified: %s", out)
}

