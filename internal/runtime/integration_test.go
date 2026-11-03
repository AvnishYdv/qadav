package runtime

import (
	"strings"
	"testing"
)

func TestIntegration_VaultReasoning(t *testing.T) {
	out := runCryptoTestHarness(t, "integration_vault_reasoning")
	if !strings.Contains(out, "INTEGRATION_VAULT_REASONING_OK") {
		t.Fatalf("Integration Vault Reasoning failed: %s", out)
	}
	t.Logf("Integration Vault Reasoning verified: %s", out)
}

func TestIntegration_MemoryScopedTTPCR(t *testing.T) {
	out := runCryptoTestHarness(t, "integration_memory_scoped_ttpcr")
	if !strings.Contains(out, "INTEGRATION_MEMORY_SCOPED_TTPCR_OK") {
		t.Fatalf("Integration Memory Scoped TTPCR failed: %s", out)
	}
	t.Logf("Integration Memory Scoped TTPCR verified: %s", out)
}

func TestIntegration_MultimodalCertificate(t *testing.T) {
	out := runCryptoTestHarness(t, "integration_multimodal_certificate")
	if !strings.Contains(out, "INTEGRATION_MULTIMODAL_CERTIFICATE_OK") {
		t.Fatalf("Integration Multimodal Certificate failed: %s", out)
	}
	t.Logf("Integration Multimodal Certificate verified: %s", out)
}

func TestIntegration_QECToolCert(t *testing.T) {
	out := runCryptoTestHarness(t, "integration_qec_tool_cert")
	if !strings.Contains(out, "INTEGRATION_QEC_TOOL_CERT_OK") {
		t.Fatalf("Integration QEC Tool Cert failed: %s", out)
	}
	t.Logf("Integration QEC Tool Cert verified: %s", out)
}

func TestIntegration_FullPipeline(t *testing.T) {
	out := runCryptoTestHarness(t, "integration_full_pipeline")
	if !strings.Contains(out, "INTEGRATION_FULL_PIPELINE_OK") {
		t.Fatalf("Integration Full Pipeline failed: %s", out)
	}
	t.Logf("Integration Full Pipeline verified: %s", out)
}

