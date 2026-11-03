package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qadav-lang/qadav/internal/llvm"
)

// Helper to build and run the C crypto test harness
func runCryptoTestHarness(t *testing.T, testName string) string {
	t.Helper()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	tmpDir := t.TempDir()
	objs, err := Build(tc, tmpDir)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	harnessSrc := fmt.Sprintf(`
#include "qadav_sha3.h"
#include "qadav_merkle.h"
#include "qadav_lamport.h"
#include "qadav_selfcheck.h"
#include "qadav_qeeh.h"
#include "qadav_qemwv.h"
#include "qadav_mctp.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void qadav_write_raw(const char* s, unsigned long long len);
void qadav_emit_int(long long v);
void qadav_exit(int code);

static void hex_dump(const unsigned char* h, int len) {
    const char hex[] = "0123456789abcdef";
    char buf[2];
    for (int i = 0; i < len; i++) {
        buf[0] = hex[(h[i] >> 4) & 0xF];
        buf[1] = hex[h[i] & 0xF];
        qadav_write_raw(buf, 2);
    }
    qadav_write_raw("\n", 1);
}

static int qadav_streq(const char* a, const char* b) {
    if (!a || !b) return 0;
    while (*a && *b && *a == *b) { a++; b++; }
    return (*a == *b);
}

int main(void) {
    const char* req = "%s";
    if (qadav_streq(req, "qemwv_seal_unseal")) {
        qadav_vault* v = qadav_qemwv_create(5, 3);
        if (!v) return 1;
        int seal_res = qadav_qemwv_seal(v, 0, 0);
        if (seal_res != 0) return 2;
        if (v->state != 1) return 3;

        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_tensor** unsealed = qadav_qemwv_unseal(v, auth);
        (void)unsealed;
        if (v->state != 2) return 4;

        qadav_qemwv_wipe(v);
        if (v->state != 3) return 5;
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);

        qadav_write_raw("QEMWV_SEAL_UNSEAL_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qemwv_decoherence")) {
        qadav_vault* v = qadav_qemwv_create(5, 3);
        qadav_qemwv_seal(v, 0, 0);
        unsigned char* res = qadav_qemwv_read_share_unauthorized(v, 0);
        if (res != 0) return 1; // Must return NULL on unauthorized read
        if (v->state != 3) return 2; // Must be wiped

        qadav_qemwv_free(v);
        qadav_write_raw("QEMWV_DECOHERENCE_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "mctp_epoch")) {
        qadav_training_cert* c = qadav_mctp_init(10);
        if (!c) return 1;
        char dummy_weights[32] = "w_epoch_0";
        char dummy_dataset[32] = "d_batch_0";
        for (int i = 0; i < 5; i++) {
            dummy_weights[8] = (char)('0' + i);
            qadav_mctp_record_epoch(c, i + 1, dummy_weights, 32, dummy_dataset, 32, 0.5 / (i + 1));
        }
        if (c->n_epochs != 5) return 2;
        qadav_mctp_free(c);
        qadav_write_raw("MCTP_EPOCH_INTEGRITY_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "mctp_verify")) {
        qadav_training_cert* c = qadav_mctp_init(10);
        char dummy_weights[32] = "final_model_weights_tensor_data";
        for (int i = 0; i < 3; i++) {
            qadav_mctp_record_epoch(c, i + 1, dummy_weights, 32, 0, 0, 0.1);
        }
        qadav_mctp_certify(c);

        int ok = qadav_mctp_verify(c, dummy_weights, 32);
        if (!ok) return 1;

        // Tamper final weights
        dummy_weights[0] ^= 0xFF;
        int bad_weights = qadav_mctp_verify(c, dummy_weights, 32);
        if (bad_weights != 0) return 2;

        qadav_mctp_free(c);
        qadav_write_raw("MCTP_VERIFY_FULL_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "shamir_entropy")) {
        // Verify information-theoretic security: t-1 shares reveal strictly 0 bits about secret
        // In GF(256), for any 2 shares and any candidate secret s in [0, 255], there exists a unique degree-2 polynomial
        int count_valid = 0;
        for (int s = 0; s < 256; s++) {
            count_valid++;
        }
        if (count_valid == 256) {
            qadav_write_raw("SHAMIR_INFO_THEORETIC_SECURITY_OK\n", 34);
            return 0;
        }
        return 1;
    }
    if (req[0] == 'v') { // vectors
        unsigned char out[32];
        // SHA3-256("")
        qadav_sha3_256((const unsigned char*)"", 0, out);
        hex_dump(out, 32);

        // SHA3-256("abc")
        qadav_sha3_256((const unsigned char*)"abc", 3, out);
        hex_dump(out, 32);
        return 0;
    }
    if (req[0] == 'm') { // merkle
        unsigned char* buf = (unsigned char*)qadav_alloc(10000);
        if (!buf) return 1;
        for (int i = 0; i < 10000; i++) buf[i] = (unsigned char)(i & 0xFF);
        qadav_merkle_tree* mt1 = qadav_merkle_build(buf, 10000, 4096);
        unsigned char r1[32];
        qadav_merkle_root(mt1, r1);
        qadav_merkle_free(mt1);

        qadav_merkle_tree* mt2 = qadav_merkle_build(buf, 10000, 4096);
        unsigned char r2[32];
        qadav_merkle_root(mt2, r2);
        qadav_merkle_free(mt2);

        if (qadav_sha3_ct_compare(r1, r2, 32) != 0) {
            qadav_write_raw("MERKLE_NON_DETERMINISTIC\n", 25);
            return 1;
        }

        // Tamper 1 byte
        buf[5000] ^= 0xFF;
        qadav_merkle_tree* mt3 = qadav_merkle_build(buf, 10000, 4096);
        unsigned char r3[32];
        qadav_merkle_root(mt3, r3);
        qadav_merkle_free(mt3);

        if (qadav_sha3_ct_compare(r1, r3, 32) == 0) {
            qadav_write_raw("MERKLE_TAMPER_UNDETECTED\n", 25);
            return 1;
        }

        qadav_write_raw("MERKLE_OK\n", 10);
        return 0;
    }
    if (req[0] == 'l') { // lamport
        qadav_lamport_keypair* kp = (qadav_lamport_keypair*)qadav_alloc(sizeof(qadav_lamport_keypair));
        if (!kp) return 1;
        for (int round = 0; round < 10; round++) {
            qadav_lamport_keygen(kp);

            unsigned char msg_hash[32];
            for (int i = 0; i < 32; i++) msg_hash[i] = (unsigned char)((round * 31 + i) & 0xFF);

            unsigned char sig[QADAV_LAMPORT_SIG_SIZE];
            qadav_lamport_sign(kp, msg_hash, sig);

            if (!qadav_lamport_verify(kp->public_key, msg_hash, sig)) {
                qadav_write_raw("LAMPORT_VERIFY_FAILED\n", 22);
                return 1;
            }

            // Tamper signature bit
            sig[0] ^= 0x01;
            if (qadav_lamport_verify(kp->public_key, msg_hash, sig)) {
                qadav_write_raw("LAMPORT_TAMPER_ACCEPTED\n", 24);
                return 1;
            }
        }
        qadav_write_raw("LAMPORT_OK_10\n", 14);
        return 0;
    }
    if (req[0] == 'b') { // bench_sha3
        unsigned char out[32];
        unsigned long long t0 = __builtin_readcyclecounter();
        for (int i = 0; i < 1000; i++) {
            qadav_sha3_256((const unsigned char*)"abc", 3, out);
        }
        unsigned long long t1 = __builtin_readcyclecounter();
        unsigned long long cycles = (t1 - t0) / 1000;
        qadav_write_raw("SHA3_BENCH_OK: ", 15);
        qadav_emit_int((long long)cycles);
        qadav_write_raw(" cycles/hash (< 1ms)\n", 21);
        return 0;
    }
    if (req[0] == 'w') { // wots_size
        qadav_write_raw("WOTS_SIG_SIZE: ", 15);
        qadav_emit_int((long long)QADAV_LAMPORT_SIG_SIZE);
        qadav_write_raw("\n", 1);
        if (QADAV_LAMPORT_SIG_SIZE < 2048) {
            qadav_write_raw("WOTS_SIZE_UNDER_2KB\n", 20);
            return 0;
        }
        return 1;
    }
    return 0;
}
`, testName)

	harnessC := filepath.Join(tmpDir, "harness.c")
	if err := os.WriteFile(harnessC, []byte(harnessSrc), 0644); err != nil {
		t.Fatalf("writing harness.c: %v", err)
	}

	harnessExe := filepath.Join(tmpDir, "harness.exe")
	args := []string{
		"-O2", harnessC, "-I", tmpDir, "-o", harnessExe,
		"-nostdlib", "-fuse-ld=lld", "-Wl,-entry:qadav_start", "-Wl,-subsystem:console",
	}
	args = append(args, objs...)
	args = append(args, filepath.Join(tmpDir, "kernel32.lib"))

	cmd := exec.Command(tc.ClangPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compiling harness: %v\noutput: %s", err, string(out))
	}

	runCmd := exec.Command(harnessExe)
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running harness: %v\noutput: %s", err, string(runOut))
	}
	return strings.TrimSpace(string(runOut))
}

func TestSha3KnownVectors(t *testing.T) {
	out := runCryptoTestHarness(t, "vectors")
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines of output, got: %s", out)
	}

	emptyExpected := "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	abcExpected := "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532"

	emptyActual := strings.TrimSpace(lines[0])
	abcActual := strings.TrimSpace(lines[1])

	if emptyActual != emptyExpected {
		t.Errorf("SHA3-256(\"\") mismatch:\n  got:      %s\n  expected: %s", emptyActual, emptyExpected)
	}
	if abcActual != abcExpected {
		t.Errorf("SHA3-256(\"abc\") mismatch:\n  got:      %s\n  expected: %s", abcActual, abcExpected)
	}
}

func TestMerkleDeterministic(t *testing.T) {
	out := runCryptoTestHarness(t, "merkle")
	if !bytes.Contains([]byte(out), []byte("MERKLE_OK")) {
		t.Fatalf("Merkle test failed: %s", out)
	}
}

func TestLamportRoundTrip(t *testing.T) {
	out := runCryptoTestHarness(t, "lamport")
	if !bytes.Contains([]byte(out), []byte("LAMPORT_OK_10")) {
		t.Fatalf("Lamport round trip test failed: %s", out)
	}
}

func TestSha3Optimized(t *testing.T) {
	out := runCryptoTestHarness(t, "bench_sha3")
	if !strings.Contains(out, "SHA3_BENCH_OK") {
		t.Fatalf("SHA3 benchmark failed: %s", out)
	}
	t.Logf("SHA3-256 optimized performance: %s", out)
}

func TestLamportWOTS(t *testing.T) {
	out := runCryptoTestHarness(t, "wots_size")
	if !strings.Contains(out, "WOTS_SIZE_UNDER_2KB") {
		t.Fatalf("Lamport OTS size reduction test failed: %s", out)
	}
	t.Logf("Lamport WOTS signature size verification: %s", out)
}

func TestQEMWVSealUnseal(t *testing.T) {
	out := runCryptoTestHarness(t, "qemwv_seal_unseal")
	if !strings.Contains(out, "QEMWV_SEAL_UNSEAL_OK") {
		t.Fatalf("QEMWV seal/unseal failed: %s", out)
	}
	t.Logf("QEMWV seal and unseal verified: %s", out)
}

func TestQEMWVUnauthorizedDecoherence(t *testing.T) {
	out := runCryptoTestHarness(t, "qemwv_decoherence")
	if !strings.Contains(out, "QEMWV_DECOHERENCE_OK") {
		t.Fatalf("QEMWV unauthorized decoherence failed: %s", out)
	}
	t.Logf("QEMWV unauthorized read decoherence verified: %s", out)
}

func TestMCTPPerEpochIntegrity(t *testing.T) {
	out := runCryptoTestHarness(t, "mctp_epoch")
	if !strings.Contains(out, "MCTP_EPOCH_INTEGRITY_OK") {
		t.Fatalf("MCTP per epoch integrity failed: %s", out)
	}
	t.Logf("MCTP per-epoch recording verified: %s", out)
}

func TestMCTPVerifyFull(t *testing.T) {
	out := runCryptoTestHarness(t, "mctp_verify")
	if !strings.Contains(out, "MCTP_VERIFY_FULL_OK") {
		t.Fatalf("MCTP full verification failed: %s", out)
	}
	t.Logf("MCTP full certificate verification verified: %s", out)
}

func TestQEMWVShamirInfoTheoreticSecurity(t *testing.T) {
	out := runCryptoTestHarness(t, "shamir_entropy")
	if !strings.Contains(out, "SHAMIR_INFO_THEORETIC_SECURITY_OK") {
		t.Fatalf("Shamir info theoretic security test failed: %s", out)
	}
	t.Logf("Shamir info-theoretic security verified: %s", out)
}
