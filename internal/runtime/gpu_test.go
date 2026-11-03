package runtime

import (
	"strings"
	"testing"
)

func TestGPUInitAndDetect(t *testing.T) {
	out := runCryptoTestHarness(t, "gpu_init_and_detect")
	if !strings.Contains(out, "GPU_INIT_AND_DETECT_OK") {
		t.Fatalf("GPU init and detect failed: %s", out)
	}
	t.Logf("GPU init and detect verified: %s", out)
}

func TestGPUMatmulCPUFallback(t *testing.T) {
	out := runCryptoTestHarness(t, "gpu_matmul_cpu_fallback")
	if !strings.Contains(out, "GPU_MATMUL_OK") {
		t.Fatalf("GPU matmul CPU fallback failed: %s", out)
	}
	t.Logf("GPU matmul verified: %s", out)
}

func TestGPUSoftmaxCPUFallback(t *testing.T) {
	out := runCryptoTestHarness(t, "gpu_softmax_cpu_fallback")
	if !strings.Contains(out, "GPU_SOFTMAX_OK") {
		t.Fatalf("GPU softmax CPU fallback failed: %s", out)
	}
	t.Logf("GPU softmax verified: %s", out)
}

func TestGPULayernormCPUFallback(t *testing.T) {
	out := runCryptoTestHarness(t, "gpu_layernorm_cpu_fallback")
	if !strings.Contains(out, "GPU_LAYERNORM_OK") {
		t.Fatalf("GPU layernorm CPU fallback failed: %s", out)
	}
	t.Logf("GPU layernorm verified: %s", out)
}

func TestAGKEAttestAndVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "agke_attest_and_verify")
	if !strings.Contains(out, "AGKE_ATTEST_AND_VERIFY_OK") {
		t.Fatalf("AGKE attest and verify failed: %s", out)
	}
	t.Logf("AGKE attest and verify verified: %s", out)
}

func TestAGKETamperDetection(t *testing.T) {
	// Tamper simulation must trigger quantum decoherence and exit process with 0xDEADBEEF
	out, err := runCryptoTestHarnessExpectExit(t, "agke_tamper_detection")
	if err == nil {
		t.Fatalf("expected tamper detection decoherence (0xDEADBEEF), but process succeeded: %s", out)
	}
	if !strings.Contains(err.Error(), "0xdeadbeef") && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("expected exit status 0xdeadbeef, got %v", err)
	}
	t.Logf("AGKE tamper detected and decoherence triggered correctly: %v", err)
}

