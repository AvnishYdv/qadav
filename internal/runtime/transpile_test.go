package runtime

import (
	"strings"
	"testing"
)

func TestTranspileToQiskit(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_qiskit")
	if !strings.Contains(out, "TRANSPILE_QISKIT_OK") {
		t.Fatalf("transpile_qiskit failed: %s", out)
	}
	t.Logf("transpile_qiskit verified: %s", out)
}

func TestTranspileToCirq(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_cirq")
	if !strings.Contains(out, "TRANSPILE_CIRQ_OK") {
		t.Fatalf("transpile_cirq failed: %s", out)
	}
	t.Logf("transpile_cirq verified: %s", out)
}

func TestTranspileToIonQ(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_ionq")
	if !strings.Contains(out, "TRANSPILE_IONQ_OK") {
		t.Fatalf("transpile_ionq failed: %s", out)
	}
	t.Logf("transpile_ionq verified: %s", out)
}

func TestTranspileToQASM3(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_qasm3")
	if !strings.Contains(out, "TRANSPILE_QASM3_OK") {
		t.Fatalf("transpile_qasm3 failed: %s", out)
	}
	t.Logf("transpile_qasm3 verified: %s", out)
}

func TestTranspileAdjacentCancel(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_adjacent_cancel")
	if !strings.Contains(out, "TRANSPILE_ADJACENT_CANCEL_OK") {
		t.Fatalf("transpile_adjacent_cancel failed: %s", out)
	}
	t.Logf("transpile_adjacent_cancel verified: %s", out)
}

func TestTranspileBufferTooSmall(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_buffer_too_small")
	if !strings.Contains(out, "TRANSPILE_BUFFER_TOO_SMALL_OK") {
		t.Fatalf("transpile_buffer_too_small failed: %s", out)
	}
	t.Logf("transpile_buffer_too_small verified: %s", out)
}

func TestTranspileUnsupportedTarget(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_unsupported_target")
	if !strings.Contains(out, "TRANSPILE_UNSUPPORTED_TARGET_OK") {
		t.Fatalf("transpile_unsupported_target failed: %s", out)
	}
	t.Logf("transpile_unsupported_target verified: %s", out)
}

func TestTranspileNoCrashOnHugeCircuit(t *testing.T) {
	out := runCryptoTestHarness(t, "transpile_no_crash_huge")
	if !strings.Contains(out, "TRANSPILE_NO_CRASH_HUGE_OK") {
		t.Fatalf("transpile_no_crash_huge failed: %s", out)
	}
	t.Logf("transpile_no_crash_huge verified: %s", out)
}
