package runtime

import (
	"strings"
	"testing"
)

func TestMultiGPUInitAndPlan(t *testing.T) {
	out := runCryptoTestHarness(t, "multigpu_init_and_plan")
	if !strings.Contains(out, "MULTIGPU_INIT_AND_PLAN_OK") {
		t.Fatalf("multigpu init and plan failed: %s", out)
	}
	t.Logf("multigpu init and plan verified: %s", out)
}

func TestMultiGPURunFallback(t *testing.T) {
	out := runCryptoTestHarness(t, "multigpu_run_fallback")
	if !strings.Contains(out, "MULTIGPU_RUN_OK") {
		t.Fatalf("multigpu run fallback failed: %s", out)
	}
	t.Logf("multigpu run fallback verified: %s", out)
}
