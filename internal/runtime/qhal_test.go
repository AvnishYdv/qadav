package runtime

import (
	"strings"
	"testing"
)

func TestQHALCreate(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_create")
	if !strings.Contains(out, "QHAL_CREATE_OK") {
		t.Fatalf("qhal_create failed: %s", out)
	}
	t.Logf("qhal_create verified: %s", out)
}

func TestQHALAddOps(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_add_ops")
	if !strings.Contains(out, "QHAL_ADD_OPS_OK") {
		t.Fatalf("qhal_add_ops failed: %s", out)
	}
	t.Logf("qhal_add_ops verified: %s", out)
}

func TestQHALExportQASM3(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_export_qasm3")
	if !strings.Contains(out, "OPENQASM 3.0;") {
		t.Fatalf("qhal_export_qasm3 missing header: %s", out)
	}
	if !strings.Contains(out, "qubit[2] q;") {
		t.Fatalf("qhal_export_qasm3 missing qubit decl: %s", out)
	}
	if !strings.Contains(out, "h q[0];") {
		t.Fatalf("qhal_export_qasm3 missing h op: %s", out)
	}
	if !strings.Contains(out, "cnot q[0], q[1];") {
		t.Fatalf("qhal_export_qasm3 missing cnot op: %s", out)
	}
	t.Logf("qhal_export_qasm3 output verified:\n%s", out)
}

func TestQHALHashCircuit(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_hash_circuit")
	if !strings.Contains(out, "QHAL_HASH_CIRCUIT_OK") {
		t.Fatalf("qhal_hash_circuit failed: %s", out)
	}
	t.Logf("qhal_hash_circuit verified: %s", out)
}

func TestQHALBackendStub(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_backend_stub")
	if !strings.Contains(out, "QHAL_BACKEND_STUB_OK") {
		t.Fatalf("qhal_backend_stub failed: %s", out)
	}
	t.Logf("qhal_backend_stub verified: %s", out)
}

func TestQHALBoundsCheck(t *testing.T) {
	out := runCryptoTestHarness(t, "qhal_bounds_check")
	if !strings.Contains(out, "QHAL_BOUNDS_CHECK_OK") {
		t.Fatalf("qhal_bounds_check failed: %s", out)
	}
	t.Logf("qhal_bounds_check verified: %s", out)
}
