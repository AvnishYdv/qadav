package runtime

import (
	"strings"
	"testing"
)

func TestSandboxCreateAndRun(t *testing.T) {
	out := runCryptoTestHarness(t, "sandbox_create_and_run")
	if !strings.Contains(out, "SANDBOX_CREATE_AND_RUN_OK") {
		t.Fatalf("Sandbox create and run failed: %s", out)
	}
	t.Logf("Sandbox create and run verified: %s", out)
}

func TestSandboxViolationGraceful(t *testing.T) {
	out := runCryptoTestHarness(t, "sandbox_violation_graceful")
	if !strings.Contains(out, "SANDBOX_VIOLATION_GRACEFUL_OK") {
		t.Fatalf("Sandbox violation graceful failure failed: %s", out)
	}
	t.Logf("Sandbox violation graceful failure verified: %s", out)
}

func TestSandboxVerifyAttestation(t *testing.T) {
	out := runCryptoTestHarness(t, "sandbox_verify_attestation")
	if !strings.Contains(out, "SANDBOX_VERIFY_ATTESTATION_OK") {
		t.Fatalf("Sandbox verify attestation failed: %s", out)
	}
	t.Logf("Sandbox verify attestation verified: %s", out)
}

func TestSandboxQuarantine(t *testing.T) {
	out := runCryptoTestHarness(t, "sandbox_quarantine")
	if !strings.Contains(out, "SANDBOX_QUARANTINE_OK") {
		t.Fatalf("Sandbox quarantine failed: %s", out)
	}
	t.Logf("Sandbox quarantine verified: %s", out)
}

func TestSandboxLifecycle(t *testing.T) {
	out := runCryptoTestHarness(t, "sandbox_lifecycle")
	if !strings.Contains(out, "SANDBOX_LIFECYCLE_OK") {
		t.Fatalf("Sandbox lifecycle tracking failed: %s", out)
	}
	t.Logf("Sandbox lifecycle verified: %s", out)
}
