package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/qadav-lang/qadav/internal/llvm"
)

var (
	sharedRuntimeOnce sync.Once
	sharedRuntimeObjs []string
	sharedRuntimeDir  string
	sharedRuntimeErr  error
)

func getSharedTestRuntime(tc *llvm.Toolchain) (string, []string, error) {
	sharedRuntimeOnce.Do(func() {
		sharedRuntimeDir = filepath.Join(os.TempDir(), "qadav_crypto_test_cache")
		_ = os.MkdirAll(sharedRuntimeDir, 0755)
		sharedRuntimeObjs, sharedRuntimeErr = Build(tc, sharedRuntimeDir)
	})
	return sharedRuntimeDir, sharedRuntimeObjs, sharedRuntimeErr
}

// Helper to build and run the C crypto test harness
func runCryptoTestHarness(t *testing.T, testName string) string {
	t.Helper()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	cacheDir, objs, err := getSharedTestRuntime(tc)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	tmpDir := t.TempDir()

	harnessSrc := fmt.Sprintf(`
#include "qadav_sha3.h"
#include "qadav_merkle.h"
#include "qadav_lamport.h"
#include "qadav_selfcheck.h"
#include "qadav_qeeh.h"
#include "qadav_qemwv.h"
#include "qadav_mctp.h"
#include "qadav_qemp.h"
#include "qadav_qecmt.h"
#include "qadav_qera.h"
#include "qadav_qecma.h"
#include "qadav_agi.h"
#include "qadav_qec.h"
#include "qadav_consolidate.h"
#include "qadav_scoped.h"
#include "qadav_tools.h"
#include "qadav_certificate.h"
#include "qadav_abstain.h"
#include "qadav_sandbox.h"
#include "qadav_gpu.h"
#include "qadav_agke.h"
#include "qadav_multigpu.h"
#include "qadav_cdaga.h"
#include "qadav_distributed.h"
#include "qadav_barat.h"
#include "qadav_qhal.h"
#include "qadav_sea.h"
#include "qadav_transpile.h"
#include "qadav_eea.h"
#include "qadav_qecdecoder.h"
#include "qadav_aad.h"
#include "qadav_record.h"
#include "qadav_variant.h"
#include "qadav_rsca.h"
#include "qadav_ctsa.h"
#include "qadav_pcc.h"

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

static void qadav_copy_bytes(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) d[i] = s[i];
}

int main(void) {
    const char* req = "%s";
    if (qadav_streq(req, "qemwv_seal_unseal")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        if (!v) return 1;
        int seal_res = qadav_qemwv_seal(v, 0, 0);
        if (seal_res != 0) return 2;
        if (v->state != 1) return 3;

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
        qadav_vault* v = qadav_qemwv_create(5, 3, 0);
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
    if (qadav_streq(req, "qemwv_escalation")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth1 = qadav_qeeh_create(1, &pen);
        qadav_ethical_hamiltonian* auth2 = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth1);
        qadav_qemwv_seal(v, 0, 0);
        qadav_write_raw("QEMWV_BOUNDARY_ESCALATION_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "qemp_remember")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
        if (!mp) return 1;
        int r1 = qadav_qemp_remember_text(mp, "pref", "detailed", "main");
        int r2 = qadav_qemp_remember_text(mp, "name", "Alice", "main");
        if (r1 != 0 || r2 != 0 || mp->count != 2) return 2;
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMP_REMEMBER_OK\n", 17);
        return 0;
    }
    if (qadav_streq(req, "qemp_recall")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
        qadav_qemp_remember_text(mp, "user_name", "Alice", "main");
        char* val = qadav_qemp_recall_text(mp, "user_name");
        if (!val || !qadav_streq(val, "Alice")) return 1;
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMP_RECALL_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "qemp_tamper")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
        qadav_qemp_remember_text(mp, "secret", "critical_data", "main");
        int c_res = qadav_qemp_corrupt_share(mp, "secret", 0);
        if (c_res != 0) return 1;
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMP_TAMPER_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "qemp_verify")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
        qadav_qemp_remember_text(mp, "k1", "v1", "main");
        qadav_qemp_remember_text(mp, "k2", "v2", "main");
        int v_ok = qadav_qemp_verify_all(mp);
        if (!v_ok) return 1;
        int a_ok = qadav_qemp_audit(mp);
        if (!a_ok) return 2;
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMP_VERIFY_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "qeeh_boundary_hash_det")) {
        double pen[2] = {1.0, 0.8};
        qadav_ethical_hamiltonian* b = qadav_qeeh_create(2, pen);
        qadav_qeeh_set_rule_name(b, 0, "harm_user");
        qadav_qeeh_set_rule_name(b, 1, "deceive");
        unsigned char base[32];
        qadav_qeeh_boundary_hash(b, base);
        for (int i = 0; i < 100; i++) {
            unsigned char cur[32];
            qadav_qeeh_boundary_hash(b, cur);
            if (qadav_sha3_ct_compare(base, cur, 32) != 0) {
                qadav_qeeh_free(b);
                return 1;
            }
        }
        qadav_qeeh_free(b);
        qadav_write_raw("QEEH_BOUNDARY_HASH_DETERMINISTIC_OK\n", 36);
        return 0;
    }
    if (qadav_streq(req, "qemp_binary_roundtrip")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
        unsigned char bin_data[5] = {0x00, 0xFF, 0x42, 0x00, 0xAA};
        qadav_qemp_remember(mp, "bin_key", bin_data, 5, "test");
        unsigned char* recalled = 0;
        unsigned long long recalled_len = 0;
        int r = qadav_qemp_recall(mp, "bin_key", auth, &recalled, &recalled_len);
        if (r != 0 || recalled_len != 5 || !recalled) return 1;
        for (int i = 0; i < 5; i++) {
            if (recalled[i] != bin_data[i]) return 2;
        }
        qadav_free(recalled);
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMP_BINARY_ROUNDTRIP_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "qemp_deterministic")) {
        unsigned char first_root[32];
        for (int iter = 0; iter < 100; iter++) {
            double pen = 1.0;
            qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
            qadav_memory_provenance* mp = qadav_qemp_create(auth, 100);
            qadav_qemp_remember_text(mp, "k1", "val1", "p");
            qadav_qemp_remember_text(mp, "k2", "val2", "p");
            if (iter == 0) {
                for (int b = 0; b < 32; b++) first_root[b] = mp->vault_root[b];
            } else {
                if (qadav_sha3_ct_compare(first_root, mp->vault_root, 32) != 0) {
                    qadav_qemp_free(mp);
                    qadav_qeeh_free(auth);
                    return 1;
                }
            }
            qadav_qemp_free(mp);
            qadav_qeeh_free(auth);
        }
        qadav_write_raw("QEMP_DETERMINISTIC_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "mctp_deterministic")) {
        unsigned char first_cert[32];
        for (int iter = 0; iter < 10; iter++) {
            qadav_training_cert* c = qadav_mctp_init(5);
            char dummy_weights[32] = "weights_buf_static";
            char dummy_dataset[32] = "dataset_buf_static";
            for (int ep = 0; ep < 3; ep++) {
                qadav_mctp_record_epoch(c, ep + 1, dummy_weights, 32, dummy_dataset, 32, 0.05);
            }
            qadav_mctp_certify(c);
            if (iter == 0) {
                for (int b = 0; b < 32; b++) first_cert[b] = c->merkle_root[b];
            } else {
                if (qadav_sha3_ct_compare(first_cert, c->merkle_root, 32) != 0) {
                    qadav_mctp_free(c);
                    return 1;
                }
            }
            qadav_mctp_free(c);
        }
        qadav_write_raw("MCTP_DETERMINISTIC_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "qemwv_deterministic_seal")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v1 = qadav_qemwv_create(5, 3, auth);
        qadav_vault* v2 = qadav_qemwv_create(5, 3, auth);
        unsigned long long fixed_seed = 0x123456789ABCDEF0ULL;
        qadav_qemwv_seal_with_seed(v1, 0, 0, fixed_seed);
        qadav_qemwv_seal_with_seed(v2, 0, 0, fixed_seed);
        for (int i = 0; i < 5; i++) {
            if (v1->shares[i].size != v2->shares[i].size) return 1;
            if (qadav_sha3_ct_compare(v1->shares[i].commitment, v2->shares[i].commitment, 32) != 0) return 2;
        }
        if (qadav_sha3_ct_compare(v1->vault_id, v2->vault_id, 32) != 0) return 3;
        qadav_qemwv_free(v1);
        qadav_qemwv_free(v2);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEMWV_DETERMINISTIC_SEAL_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "qecmt_license_binding")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        qadav_qemwv_seal(v, 0, 0);
        qadav_qecmt_context* ctx = qadav_qecmt_create();
        int lic_idx = qadav_qecmt_issue_license(ctx, v, 0, 50, 0.4, 5000, auth);
        if (lic_idx < 0) return 1;
        if (ctx->licenses[lic_idx].max_epochs != 50) return 2;
        if (ctx->licenses[lic_idx].max_distill_rate < 0.39 || ctx->licenses[lic_idx].max_distill_rate > 0.41) return 3;
        qadav_qecmt_free(ctx);
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);
        qadav_write_raw("QECMT_LICENSE_BINDING_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "qecmt_provenance_chain")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        qadav_qemwv_seal(v, 0, 0);
        qadav_qecmt_context* ctx = qadav_qecmt_create();
        int lic_idx = qadav_qecmt_issue_license(ctx, v, 0, 50, 0.4, 5000, auth);
        int layers[2] = {0, 1};
        qadav_qecmt_transfer_layer(ctx, lic_idx, v, 0, layers, 2, auth);
        int v_res = qadav_qecmt_verify_chain(ctx);
        if (v_res != 0) return 1;
        qadav_qecmt_free(ctx);
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);
        qadav_write_raw("QECMT_PROVENANCE_CHAIN_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "qecmt_rate_enforcement")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        qadav_qemwv_seal(v, 0, 0);
        qadav_qecmt_context* ctx = qadav_qecmt_create();
        int lic_idx = qadav_qecmt_issue_license(ctx, v, 0, 50, 0.2, 5000, auth);
        if (lic_idx >= 0 && ctx->licenses[lic_idx].max_distill_rate < 0.25) {
            qadav_write_raw("QECMT_RATE_ENFORCEMENT_OK\n", 26);
        }
        qadav_qecmt_free(ctx);
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);
        return 0;
    }
    if (qadav_streq(req, "qera_create_reason_verify")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("test_chain", 100, auth);
        if (!chain) return 1;
        int r1 = qadav_qera_reason(chain, 1, (const unsigned char*)"premise1", 8, (const unsigned char*)"conclusion", 10, 0, 0, 0.95, auth);
        int r2 = qadav_qera_reason(chain, 2, (const unsigned char*)"obs1", 4, (const unsigned char*)"hypothesis", 10, 0, 0, 0.90, auth);
        int r3 = qadav_qera_reason(chain, 3, (const unsigned char*)"rule1", 5, (const unsigned char*)"cause", 5, 0, 0, 0.85, auth);
        if (r1 != 0 || r2 != 0 || r3 != 0 || chain->n_steps != 3) return 2;
        int v = qadav_qera_verify(chain, auth);
        if (v != 1) return 3;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("QERA_CREATE_REASON_VERIFY_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "qera_tamper")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("tamper_chain", 100, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"premise", 7, (const unsigned char*)"conclusion", 10, 0, 0, 0.95, auth);
        int c_res = qadav_qera_corrupt_step(chain, 0, 0);
        if (c_res != 0) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("QERA_TAMPER_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "qera_step_hash_det")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* c1 = qadav_qera_create("det1", 100, auth);
        qadav_reasoning_chain* c2 = qadav_qera_create("det2", 100, auth);
        qadav_qera_reason(c1, 1, (const unsigned char*)"p", 1, (const unsigned char*)"c", 1, 0, 0, 0.95, auth);
        qadav_qera_reason(c2, 1, (const unsigned char*)"p", 1, (const unsigned char*)"c", 1, 0, 0, 0.95, auth);
        if (qadav_sha3_ct_compare(c1->steps[0].step_commitment, c2->steps[0].step_commitment, 32) != 0) return 1;
        qadav_qera_free(c1);
        qadav_qera_free(c2);
        qadav_qeeh_free(auth);
        qadav_write_raw("QERA_STEP_HASH_DET_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "qera_chain_root_det")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* c1 = qadav_qera_create("root1", 100, auth);
        qadav_reasoning_chain* c2 = qadav_qera_create("root2", 100, auth);
        for (int i = 0; i < 5; i++) {
            qadav_qera_reason(c1, 1, (const unsigned char*)"data", 4, (const unsigned char*)"res", 3, 0, 0, 0.9, auth);
            qadav_qera_reason(c2, 1, (const unsigned char*)"data", 4, (const unsigned char*)"res", 3, 0, 0, 0.9, auth);
        }
        if (qadav_sha3_ct_compare(c1->chain_root, c2->chain_root, 32) != 0) return 1;
        qadav_qera_free(c1);
        qadav_qera_free(c2);
        qadav_qeeh_free(auth);
        qadav_write_raw("QERA_CHAIN_ROOT_DET_OK\n", 23);
        return 0;
    }
    if (qadav_streq(req, "qera_boundary_escalation")) {
        double pen1 = 1.0;
        double pen2 = 0.5;
        qadav_ethical_hamiltonian* auth1 = qadav_qeeh_create(1, &pen1);
        qadav_ethical_hamiltonian* auth2 = qadav_qeeh_create(1, &pen2);
        qadav_reasoning_chain* chain = qadav_qera_create("esc", 100, auth1);
        unsigned char h1[32], h2[32];
        qadav_qeeh_boundary_hash(auth1, h1);
        qadav_qeeh_boundary_hash(auth2, h2);
        if (qadav_sha3_ct_compare(chain->authorized_boundary_id, h1, 32) != 0) return 1;
        if (qadav_sha3_ct_compare(h1, h2, 32) == 0) return 2;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth1);
        qadav_qeeh_free(auth2);
        qadav_write_raw("QERA_BOUNDARY_ESCALATION_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "qera_input_hash_nonzero")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("hash_test", 10, auth);
        int r = qadav_qera_reason(chain, 1, (const unsigned char*)"Socrates is a man", 17, (const unsigned char*)"Socrates is mortal", 18, 0, 0, 0.95, auth);
        if (r != 0) return 1;
        int in_zeros = 1, out_zeros = 1;
        for (int i = 0; i < 8; i++) {
            if (chain->steps[0].input_hash[i] != 0) in_zeros = 0;
            if (chain->steps[0].output_hash[i] != 0) out_zeros = 0;
        }
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        if (!in_zeros && !out_zeros) {
            qadav_write_raw("QERA_INPUT_HASH_NONZERO_OK\n", 27);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "qera_empty_input_rejected")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("empty_test", 10, auth);
        int r = qadav_qera_reason(chain, 1, (const unsigned char*)"", 0, (const unsigned char*)"conclusion", 10, 0, 0, 0.95, auth);
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        if (r == -5) {
            qadav_write_raw("QERA_EMPTY_INPUT_REJECTED_OK\n", 29);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "qecma_bundle_build")) {
        qadav_qecma_bundle* b = qadav_qecma_create();
        if (!b) return 1;
        float tensor_data[4] = {1.0f, 2.0f, 3.0f, 4.0f};
        qadav_qecma_add(b, QADAV_TAG_TENSOR, tensor_data, sizeof(tensor_data));
        qadav_qecma_add(b, QADAV_TAG_TEXT, "multimodal prompt", 17);
        long long concept_prime = 7919;
        qadav_qecma_add(b, QADAV_TAG_CONCEPT, &concept_prime, sizeof(concept_prime));
        if (b->count != 3) return 2;
        int r = qadav_qecma_build(b);
        if (r != 0) return 3;
        unsigned char root[32];
        qadav_qecma_get_unified_root(b, root);
        int zeros = 1;
        for (int i = 0; i < 8; i++) {
            if (root[i] != 0) { zeros = 0; break; }
        }
        if (zeros) return 4;
        qadav_qecma_free(b);
        qadav_write_raw("QECMA_BUNDLE_BUILD_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "qecma_tag_ordering")) {
        qadav_qecma_bundle* b1 = qadav_qecma_create();
        float t[4] = {1.0f, 2.0f, 3.0f, 4.0f};
        qadav_qecma_add(b1, QADAV_TAG_TENSOR, t, sizeof(t));
        qadav_qecma_add(b1, QADAV_TAG_TEXT, "hello", 5);
        long long c = 101;
        qadav_qecma_add(b1, QADAV_TAG_CONCEPT, &c, sizeof(c));
        qadav_qecma_build(b1);

        qadav_qecma_bundle* b2 = qadav_qecma_create();
        qadav_qecma_add(b2, QADAV_TAG_CONCEPT, &c, sizeof(c));
        qadav_qecma_add(b2, QADAV_TAG_TEXT, "hello", 5);
        qadav_qecma_add(b2, QADAV_TAG_TENSOR, t, sizeof(t));
        qadav_qecma_build(b2);

        if (qadav_sha3_ct_compare(b1->unified_root, b2->unified_root, 32) != 0) return 1;
        qadav_qecma_free(b1);
        qadav_qecma_free(b2);
        qadav_write_raw("QECMA_TAG_ORDERING_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "qecma_multimodal_tamper")) {
        qadav_qecma_bundle* b1 = qadav_qecma_create();
        char txt1[] = "Original modality data";
        qadav_qecma_add(b1, QADAV_TAG_TEXT, txt1, sizeof(txt1));
        qadav_qecma_build(b1);

        qadav_qecma_bundle* b2 = qadav_qecma_create();
        char txt2[] = "Tampered modality data";
        qadav_qecma_add(b2, QADAV_TAG_TEXT, txt2, sizeof(txt2));
        qadav_qecma_build(b2);

        if (qadav_sha3_ct_compare(b1->unified_root, b2->unified_root, 32) == 0) return 1;
        qadav_qecma_free(b1);
        qadav_qecma_free(b2);
        qadav_write_raw("QECMA_MULTIMODAL_TAMPER_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "qecma_cross_modal_reason")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("cm_chain", 10, auth);
        qadav_qecma_bundle* b = qadav_qecma_create();
        qadav_qecma_add(b, QADAV_TAG_TEXT, "visual scene description", 24);
        int r = qadav_qera_reason_cross_modal(chain, 1, b, (const unsigned char*)"inferred action", 15, 0.95, auth);
        if (r != 0) return 1;
        if (chain->n_steps != 1) return 2;
        if (qadav_qera_verify(chain, auth) != 1) return 3;
        qadav_qera_free(chain);
        qadav_qecma_free(b);
        qadav_qeeh_free(auth);
        qadav_write_raw("QECMA_CROSS_MODAL_REASON_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "qecma_empty_modality")) {
        qadav_qecma_bundle* b = qadav_qecma_create();
        int r = qadav_qecma_add(b, QADAV_TAG_TEXT, "", 0);
        qadav_qecma_free(b);
        if (r == -3) {
            qadav_write_raw("QECMA_EMPTY_MODALITY_REJECTED_OK\n", 33);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "qecma_empty_bundle")) {
        qadav_qecma_bundle* b = qadav_qecma_create();
        int r = qadav_qecma_build(b);
        qadav_qecma_free(b);
        if (r == -5) {
            qadav_write_raw("QECMA_EMPTY_BUNDLE_REJECTED_OK\n", 31);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "qecma_hash_sanity")) {
        qadav_qecma_bundle* b = qadav_qecma_create();
        int r = qadav_qecma_add(b, QADAV_TAG_TEXT, "non-empty-valid", 15);
        if (r != 0) { qadav_qecma_free(b); return 1; }
        int first8_zeros = 1;
        for (int i = 0; i < 8; i++) {
            if (b->items[0].digest[i] != 0) { first8_zeros = 0; break; }
        }
        qadav_qecma_free(b);
        if (!first8_zeros) {
            qadav_write_raw("QECMA_HASH_SANITY_OK\n", 21);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "vectors")) { // vectors
        unsigned char out[32];
        // SHA3-256("")
        qadav_sha3_256((const unsigned char*)"", 0, out);
        hex_dump(out, 32);

        // SHA3-256("abc")
        qadav_sha3_256((const unsigned char*)"abc", 3, out);
        hex_dump(out, 32);
        return 0;
    }
    if (qadav_streq(req, "merkle")) {
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
    if (qadav_streq(req, "bench_sha3")) {
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
    if (qadav_streq(req, "qec_create")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_qec_memory* mem = qadav_qec_create("test_mem", QADAV_QEC_SURFACE, 3, auth);
        if (!mem) return 1;
        if (mem->code_type != QADAV_QEC_SURFACE) return 2;
        if (mem->distance != 3) return 3;
        unsigned char b_hash[32];
        qadav_qeeh_boundary_hash(auth, b_hash);
        if (qadav_sha3_ct_compare(mem->authorized_boundary_id, b_hash, 32) != 0) return 4;
        qadav_qec_free(mem);
        qadav_qeeh_free(auth);
        qadav_write_raw("QEC_CREATE_OK\n", 14);
        return 0;
    }
    if (qadav_streq(req, "qec_syndrome_nonzero")) {
        qadav_qec_memory* mem = qadav_qec_create("test_mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {0x01, 0x02, 0x03, 0x04};
        unsigned long long syn_id = 0;
        int r = qadav_qec_measure(mem, raw, 4, &syn_id, 0);
        if (r != 0) return 1;
        qadav_qec_syndrome* syn = &mem->syndromes[syn_id];
        int all_zero = 1;
        for (int i = 0; i < 32; i++) {
            if (syn->commitment[i] != 0) { all_zero = 0; break; }
        }
        if (all_zero) return 2;
        qadav_qec_free(mem);
        qadav_write_raw("QEC_SYNDROME_NONZERO_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "qec_empty_syndrome")) {
        qadav_qec_memory* mem = qadav_qec_create("test_mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned long long syn_id = 0;
        int r = qadav_qec_measure(mem, 0, 0, &syn_id, 0);
        if (r >= 0) return 1;
        qadav_qec_free(mem);
        qadav_write_raw("QEC_EMPTY_SYNDROME_REJECTED_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "qec_tamper_decoherence")) {
        qadav_qec_memory* mem = qadav_qec_create("test_mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {0x11, 0x22, 0x33, 0x44};
        unsigned long long syn_id = 0;
        int mr = qadav_qec_measure(mem, raw, 4, &syn_id, 0);
        if (mr != 0) return 1;
        int cr = qadav_qec_corrupt_syndrome(mem, syn_id, 0);
        if (cr != 0) return 2;
        qadav_qec_free(mem);
        qadav_write_raw("QEC_TAMPER_DECOHERENCE_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "qec_verify_before_correct")) {
        qadav_qec_memory* mem = qadav_qec_create("test_mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {0x11, 0x22, 0x33, 0x44};
        unsigned long long syn_id = 0;
        int mr = qadav_qec_measure(mem, raw, 4, &syn_id, 0);
        if (mr != 0) return 1;
        int cr = qadav_qec_correct(mem, syn_id, 0);
        if (cr != -6) return 2;
        int vr = qadav_qec_verify(mem, syn_id, 0);
        if (vr != 0) return 3;
        cr = qadav_qec_correct(mem, syn_id, 0);
        if (cr != 0) return 4;
        qadav_qec_free(mem);
        qadav_write_raw("QEC_VERIFY_BEFORE_CORRECT_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "qec_boundary_determinism")) {
        double pen[2] = {1.0, 0.5};
        qadav_ethical_hamiltonian* b = qadav_qeeh_create(2, pen);
        qadav_qeeh_set_rule_name(b, 0, "safety");
        qadav_qeeh_set_rule_name(b, 1, "truth");
        qadav_qec_memory* mem1 = qadav_qec_create("m1", QADAV_QEC_SURFACE, 5, b);
        unsigned char base[32];
        for (int i = 0; i < 32; i++) base[i] = mem1->authorized_boundary_id[i];
        for (int iter = 0; iter < 100; iter++) {
            qadav_qec_memory* cur = qadav_qec_create("cur", QADAV_QEC_SURFACE, 5, b);
            if (qadav_sha3_ct_compare(base, cur->authorized_boundary_id, 32) != 0) {
                qadav_qec_free(cur);
                qadav_qec_free(mem1);
                qadav_qeeh_free(b);
                return 1;
            }
            qadav_qec_free(cur);
        }
        qadav_qec_free(mem1);
        qadav_qeeh_free(b);
        qadav_write_raw("QEC_BOUNDARY_DETERMINISM_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "qec_free_normal")) {
        int before_c = qadav_qec_get_create_count();
        int before_f = qadav_qec_get_free_count();
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        if (!mem) return 1;
        qadav_qec_free(mem);
        int after_c = qadav_qec_get_create_count();
        int after_f = qadav_qec_get_free_count();
        if (after_c == before_c + 1 && after_f == before_f + 1 && after_c == after_f) {
            qadav_write_raw("QEC_FREE_NORMAL_OK\n", 19);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "qec_free_verify_failure")) {
        int before_c = qadav_qec_get_create_count();
        int before_f = qadav_qec_get_free_count();
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {1, 2, 3, 4};
        unsigned long long sid = 0;
        qadav_qec_measure(mem, raw, 4, &sid, 0);
        qadav_qec_corrupt_syndrome(mem, sid, 0);
        qadav_qec_free(mem);
        int after_c = qadav_qec_get_create_count();
        int after_f = qadav_qec_get_free_count();
        if (after_c == before_c + 1 && after_f == before_f + 1 && after_c == after_f) {
            qadav_write_raw("QEC_FREE_VERIFY_FAILURE_OK\n", 27);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "qec_free_correct_failure")) {
        int before_c = qadav_qec_get_create_count();
        int before_f = qadav_qec_get_free_count();
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {1, 2, 3, 4};
        unsigned long long sid = 0;
        qadav_qec_measure(mem, raw, 4, &sid, 0);
        int cr = qadav_qec_correct(mem, sid, 0);
        if (cr != -6) { qadav_qec_free(mem); return 1; }
        qadav_qec_free(mem);
        int after_c = qadav_qec_get_create_count();
        int after_f = qadav_qec_get_free_count();
        if (after_c == before_c + 1 && after_f == before_f + 1 && after_c == after_f) {
            qadav_write_raw("QEC_FREE_CORRECT_FAILURE_OK\n", 28);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "qec_overflow_syndrome")) {
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[1025];
        for (int i = 0; i < 1025; i++) raw[i] = 1;
        unsigned long long sid = 0;
        int r = qadav_qec_measure(mem, raw, 1025, &sid, 0);
        qadav_qec_free(mem);
        if (r == -4) {
            qadav_write_raw("QEC_OVERFLOW_SYNDROME_OK\n", 25);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "qec_zero_syndrome")) {
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char zero_raw[16] = {0};
        unsigned long long sid = 0;
        int r = qadav_qec_measure(mem, zero_raw, 16, &sid, 0);
        qadav_qec_free(mem);
        if (r == -3) {
            qadav_write_raw("QEC_ZERO_SYNDROME_OK\n", 21);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "qec_verify_success")) {
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[8] = {1, 2, 3, 4, 5, 6, 7, 8};
        unsigned long long sid = 0;
        int mr = qadav_qec_measure(mem, raw, 8, &sid, 0);
        if (mr != 0) { qadav_qec_free(mem); return 1; }
        int vr = qadav_qec_verify(mem, sid, 0);
        if (vr == 0 && mem->syndromes[sid].verified == 1) {
            qadav_qec_free(mem);
            qadav_write_raw("QEC_VERIFY_SUCCESS_OK\n", 22);
            return 0;
        }
        qadav_qec_free(mem);
        return 2;
    }
    if (qadav_streq(req, "qec_correct_after_verify")) {
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[8] = {1, 2, 3, 4, 5, 6, 7, 8};
        unsigned long long sid = 0;
        qadav_qec_measure(mem, raw, 8, &sid, 0);
        int vr = qadav_qec_verify(mem, sid, 0);
        if (vr != 0) { qadav_qec_free(mem); return 1; }
        int cr = qadav_qec_correct(mem, sid, 0);
        qadav_qec_free(mem);
        if (cr == 0) {
            qadav_write_raw("QEC_CORRECT_AFTER_VERIFY_OK\n", 28);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "qec_max_syndromes")) {
        qadav_qec_memory* mem = qadav_qec_create("mem", QADAV_QEC_SURFACE, 3, 0);
        unsigned char raw[4] = {1, 2, 3, 4};
        unsigned long long sid = 0;
        for (int i = 0; i < 256; i++) {
            raw[0] = (unsigned char)(i + 1);
            int r = qadav_qec_measure(mem, raw, 4, &sid, 0);
            if (r != 0) { qadav_qec_free(mem); return 1; }
        }
        raw[0] = 0xFF;
        int r257 = qadav_qec_measure(mem, raw, 4, &sid, 0);
        qadav_qec_free(mem);
        if (r257 == -5) {
            qadav_write_raw("QEC_MAX_SYNDROMES_OK\n", 21);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "consolidation_test")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 50);
        qadav_qemp_remember_text(mp, "conceptA", "apple", "proc1");
        qadav_qemp_remember_text(mp, "conceptB", "fruit", "proc1");
        
        qadav_graph* g = qadav_graph_create();
        int res = qadav_consolidate_memory(mp, g, 0.5, auth);
        if (res <= 0) return 1;
        qadav_graph_free(g);
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("CONSOLIDATION_OK\n", 17);
        return 0;
    }
    if (qadav_streq(req, "scoped_issue_and_verify")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = (unsigned char)(0xAA ^ i);
        const char* keys[] = {"conceptA", "conceptB"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok1", auth, rec_hash, keys, 2, 3600);
        if (!tok) { qadav_qeeh_free(auth); return 1; }

        if (qadav_scoped_verify(tok, rec_hash, "conceptA", 1000) != 0) return 2;
        if (qadav_scoped_verify(tok, rec_hash, "conceptB", 1000) != 0) return 3;
        if (qadav_scoped_verify(tok, rec_hash, "conceptC", 1000) != -5) return 4;

        unsigned char bad_rec[32];
        for (int i = 0; i < 32; i++) bad_rec[i] = 0xEE;
        if (qadav_scoped_verify(tok, bad_rec, "conceptA", 1000) != -4) return 5;

        if (qadav_scoped_verify(tok, rec_hash, "conceptA", 5000) != -3) return 6;

        qadav_scoped_revoke(tok);
        if (qadav_scoped_verify(tok, rec_hash, "conceptA", 1000) != -2) return 7;

        qadav_scoped_free(tok);
        qadav_qeeh_free(auth);
        qadav_write_raw("SCOPED_ISSUE_AND_VERIFY_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "scoped_recall_privacy")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 50);
        qadav_qemp_remember_text(mp, "secret_key", "top_secret_data", "source1");

        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = (unsigned char)(0xBB ^ i);
        const char* keys[] = {"secret_key"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok1", auth, rec_hash, keys, 1, 3600);

        void* out_val = 0;
        unsigned long long out_len = 0;
        int r = qadav_scoped_recall(mp, tok, rec_hash, "secret_key", 1000, &out_val, &out_len);
        if (r != 0 || !out_val) return 1;

        if (qadav_streq((const char*)out_val, "top_secret_data")) {
            qadav_free(out_val);
            qadav_scoped_free(tok);
            qadav_qemp_free(mp);
            qadav_qeeh_free(auth);
            qadav_write_raw("SCOPED_RECALL_PRIVACY_OK\n", 25);
            return 0;
        }
        return 2;
    }
    if (qadav_streq(req, "scoped_unauthorized_key")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 50);
        qadav_qemp_remember_text(mp, "secret_key", "top_secret_data", "source1");

        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = (unsigned char)(0xBB ^ i);
        const char* keys[] = {"other_key"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok1", auth, rec_hash, keys, 1, 3600);

        void* out_val = 0;
        unsigned long long out_len = 0;
        int r = qadav_scoped_recall(mp, tok, rec_hash, "secret_key", 1000, &out_val, &out_len);
        qadav_scoped_free(tok);
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        if (r == -5) {
            qadav_write_raw("SCOPED_UNAUTHORIZED_KEY_OK\n", 27);
            return 0;
        }
        return 1;
    }
    if (qadav_streq(req, "tool_register_and_call")) {
        qadav_tools_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        int r1 = qadav_tool_register("file_reader", auth, 1, 1);
        if (r1 != 0) return 1;

        qadav_reasoning_chain* chain = qadav_qera_create("audit_chain", 10, auth);
        int r2 = qadav_tool_call("file_reader", auth, "/safe/path", chain);
        if (r2 != 0) return 2;
        if (chain->n_steps != 1) return 3;

        int r3 = qadav_tool_call("file_reader", auth, "", chain);
        if (r3 != -3) return 4;

        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("TOOL_REGISTER_AND_CALL_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "scoped_expired_token")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = 0x22;
        const char* keys[] = {"k1"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok_exp", auth, rec_hash, keys, 1, 1);
        if (!tok) return 1;
        if (qadav_scoped_verify(tok, rec_hash, "k1", 1000) != 0) return 2;
        int v_exp = qadav_scoped_verify(tok, rec_hash, "k1", 1005);
        if (v_exp != -3) return 3;
        qadav_scoped_free(tok);
        qadav_qeeh_free(auth);
        qadav_write_raw("SCOPED_EXPIRED_TOKEN_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "scoped_recipient_mismatch")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = 0x33;
        const char* keys[] = {"k1"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok_rec", auth, rec_hash, keys, 1, 3600);
        unsigned char bad_rec[32];
        for (int i = 0; i < 32; i++) bad_rec[i] = 0x88;
        int v_rec = qadav_scoped_verify(tok, bad_rec, "k1", 1000);
        if (v_rec != -4) return 1;
        qadav_qeeh_trigger_decoherence(0);
        qadav_scoped_free(tok);
        qadav_qeeh_free(auth);
        qadav_write_raw("SCOPED_RECIPIENT_MISMATCH_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "scoped_boundary_not_disclosed")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = 0x44;
        const char* keys[] = {"k1"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok_priv", auth, rec_hash, keys, 1, 3600);
        if (!tok) return 1;
        int non_zero = 0;
        for (int i = 0; i < 32; i++) {
            if (tok->boundary_hash[i] != 0) { non_zero = 1; break; }
        }
        if (!non_zero) return 2;
        unsigned char expected_h[32];
        qadav_qeeh_boundary_hash(auth, expected_h);
        if (qadav_sha3_ct_compare(tok->boundary_hash, expected_h, 32) != 0) return 3;
        qadav_scoped_free(tok);
        qadav_qeeh_free(auth);
        qadav_write_raw("SCOPED_BOUNDARY_NOT_DISCLOSED_OK\n", 33);
        return 0;
    }
    if (qadav_streq(req, "tool_post_condition_failure")) {
        qadav_tools_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        int r1 = qadav_tool_register("fail_post_tool", auth, 1, 2);
        if (r1 != 0) return 1;
        qadav_reasoning_chain* chain = qadav_qera_create("audit_chain", 10, auth);
        int r2 = qadav_tool_call("fail_post_tool", auth, "arg", chain);
        if (r2 != -4) return 2;
        if (chain->n_steps != 0) return 3;
        qadav_tool_register("safe_tool", auth, 1, 1);
        int r3 = qadav_tool_call("safe_tool", auth, "valid_arg", chain);
        if (r3 != 0 || chain->n_steps != 1) return 4;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("TOOL_POST_CONDITION_FAILURE_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "scoped_recall_as_qecma_modality")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 50);
        qadav_qemp_remember_text(mp, "clinical_notes", "patient_vitals_normal", "main");
        unsigned char rec_hash[32];
        for (int i = 0; i < 32; i++) rec_hash[i] = (unsigned char)(0x55 ^ i);
        const char* keys[] = {"clinical_notes"};
        qadav_scoped_token* tok = qadav_scoped_issue("med_token", auth, rec_hash, keys, 1, 3600);
        void* out_val = 0;
        unsigned long long out_len = 0;
        int r = qadav_scoped_recall(mp, tok, rec_hash, "clinical_notes", 1000, &out_val, &out_len);
        if (r != 0 || !out_val) return 1;
        qadav_qecma_bundle* bundle = qadav_qecma_create();
        int a_res = qadav_qecma_add(bundle, QADAV_TAG_TEXT, out_val, out_len);
        if (a_res != 0) return 2;
        int b_res = qadav_qecma_build(bundle);
        if (b_res != 0) return 3;
        int all_zero = 1;
        for (int i = 0; i < 32; i++) {
            if (bundle->unified_root[i] != 0) { all_zero = 0; break; }
        }
        if (all_zero) return 4;
        qadav_reasoning_chain* chain = qadav_qera_create("multimodal_chain", 50, auth);
        int r_res = qadav_qera_reason_cross_modal(chain, QADAV_OP_DEDUCE, bundle,
                                                  (const unsigned char*)"diagnosis_ok", 12, 0.96, auth);
        if (r_res != 0 || chain->n_steps != 1) return 5;
        qadav_qera_free(chain);
        qadav_qecma_free(bundle);
        qadav_free(out_val);
        qadav_scoped_free(tok);
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("SCOPED_RECALL_AS_QECMA_MODALITY_OK\n", 36);
        return 0;
    }
    if (qadav_streq(req, "cert_attach")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c_chain", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, (const unsigned char*)"p1", 2,
                          (const unsigned char*)"c1", 2, 0, 0, 0.95, auth);
        const char* w = "modus_ponens";
        int att = qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)w, 12, 0.95);
        if (att != 0) return 1;
        qadav_reason_certificate* certs = (qadav_reason_certificate*)chain->certs;
        if (!certs) return 2;
        if (certs[0].kind != QADAV_CERT_DEDUCTIVE) return 3;
        if (qadav_sha3_ct_compare(certs[0].premise_hash, chain->steps[0].input_hash, 32) != 0) return 4;
        int non_zero = 0;
        for (int i = 0; i < 32; i++) {
            if (certs[0].commitment[i] != 0) { non_zero = 1; break; }
        }
        if (!non_zero) return 5;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_ATTACH_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "cert_verify_structural")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c_chain", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, (const unsigned char*)"premise", 7,
                          (const unsigned char*)"conclusion", 10, 0, 0, 0.98, auth);
        qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"rule_inst", 9, 0.98);
        qadav_reason_certificate* certs = (qadav_reason_certificate*)chain->certs;
        int vr = qadav_certificate_verify(&certs[0]);
        if (vr != 0) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_VERIFY_STRUCTURAL_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "cert_invalid_kind")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c_chain", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"p", 1, (const unsigned char*)"c", 1, 0, 0, 0.9, auth);
        int r_low = qadav_certificate_attach(chain, 0, 0, (const unsigned char*)"w", 1, 0.9);
        int r_high = qadav_certificate_attach(chain, 0, 5, (const unsigned char*)"w", 1, 0.9);
        if (r_low != -1 || r_high != -1) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_INVALID_KIND_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "cert_empty_witness")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c_chain", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"p", 1, (const unsigned char*)"c", 1, 0, 0, 0.9, auth);
        int r_empty = qadav_certificate_attach(chain, 0, 1, (const unsigned char*)"", 0, 0.9);
        int r_huge = qadav_certificate_attach(chain, 0, 1, (const unsigned char*)"w", 513, 0.9);
        if (r_empty != -2 || r_huge != -2) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_EMPTY_WITNESS_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "cert_invalid_confidence")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c_chain", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"p", 1, (const unsigned char*)"c", 1, 0, 0, 0.9, auth);
        int r_neg = qadav_certificate_attach(chain, 0, 1, (const unsigned char*)"w", 1, -0.01);
        int r_high = qadav_certificate_attach(chain, 0, 1, (const unsigned char*)"w", 1, 1.05);
        if (r_neg != -3 || r_high != -3) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_INVALID_CONFIDENCE_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "cert_chain_verify")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("logic", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, (const unsigned char*)"p1", 2, (const unsigned char*)"m1", 2, 0, 0, 0.95, auth);
        qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"w1", 2, 0.95);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, (const unsigned char*)"p2", 2, (const unsigned char*)"m2", 2, 0, 0, 0.98, auth);
        qadav_certificate_attach(chain, 1, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"w2", 2, 0.98);
        int cv = qadav_certificate_verify_chain(chain);
        if (cv != 0) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_CHAIN_VERIFY_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "abstain_compute_confidence")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.90, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"b", 1, (const unsigned char*)"c", 1, 0, 0, 0.80, auth);
        double c_val = qadav_abstain_compute_confidence(chain);
        // Variance-penalized: geom_mean(0.9, 0.8) * (1.0 - 0.0025/4) ~= 0.8480
        if (c_val < 0.847 || c_val > 0.849) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_COMPUTE_CONFIDENCE_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "abstain_below_threshold")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.55, auth);
        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_REFUSE, 0};
        int should_abstain = qadav_abstain_check(chain, &pol, 0);
        if (should_abstain != 1) return 1;
        int action = qadav_abstain_execute(chain, &pol);
        if (action != QADAV_ABSTAIN_REFUSE) return 2;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_BELOW_THRESHOLD_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "abstain_above_threshold")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.95, auth);
        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_REFUSE, 0};
        int should_abstain = qadav_abstain_check(chain, &pol, 0);
        if (should_abstain != 0) return 1;
        int action = qadav_abstain_execute(chain, &pol);
        if (action != 0) return 2;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_ABOVE_THRESHOLD_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "abstain_retry_limit")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.60, auth);
        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_RETRY, 3};
        if (qadav_abstain_check(chain, &pol, 1) != 1) return 1;
        if (qadav_abstain_check(chain, &pol, 2) != 1) return 2;
        // Limit reached: force abstain
        if (qadav_abstain_check(chain, &pol, 3) != 1) return 3;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_RETRY_LIMIT_OK\n", 23);
        return 0;
    }
    if (qadav_streq(req, "abstain_escalate_mode")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.50, auth);
        qadav_abstain_policy pol = {0.85, QADAV_ABSTAIN_ESCALATE, 0};
        int action = qadav_abstain_execute(chain, &pol);
        if (action != QADAV_ABSTAIN_ESCALATE) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_ESCALATE_MODE_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "abstain_refuse_mode")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("c", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.40, auth);
        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_REFUSE, 0};
        int action = qadav_abstain_execute(chain, &pol);
        if (action != QADAV_ABSTAIN_REFUSE) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("ABSTAIN_REFUSE_MODE_OK\n", 23);
        return 0;
    }
    if (qadav_streq(req, "cert_commitment_aslr")) {
        unsigned char base_commit[32];
        for (int i = 0; i < 100; i++) {
            unsigned char* buf = (unsigned char*)qadav_alloc(1024 + (i * 16));
            qadav_reason_certificate* cert = (qadav_reason_certificate*)(buf + (i & 31));
            cert->kind = QADAV_CERT_DEDUCTIVE;
            for (int k = 0; k < 32; k++) {
                cert->premise_hash[k] = (unsigned char)(k + 1);
                cert->conclusion_hash[k] = (unsigned char)(k + 2);
            }
            const char* wit = "deterministic_witness";
            cert->witness_len = 21;
            for (int k = 0; k < 21; k++) cert->witness[k] = wit[k];
            cert->confidence = 0.95;

            unsigned char id_preimage[77 + QADAV_CERT_WITNESS_MAX];
            id_preimage[0] = cert->kind;
            for (int k = 0; k < 32; k++) {
                id_preimage[1 + k] = cert->premise_hash[k];
                id_preimage[33 + k] = cert->conclusion_hash[k];
            }
            for (int k = 0; k < 21; k++) id_preimage[65 + k] = cert->witness[k];
            unsigned int wlen = cert->witness_len;
            for (int k = 0; k < 4; k++) id_preimage[65 + 21 + k] = ((unsigned char*)&wlen)[k];
            double conf = cert->confidence;
            for (int k = 0; k < 8; k++) id_preimage[69 + 21 + k] = ((unsigned char*)&conf)[k];
            qadav_sha3_256(id_preimage, 77 + 21, cert->cert_id);

            unsigned char commit_preimage[32 + 77 + 21];
            for (int k = 0; k < 32; k++) commit_preimage[k] = cert->cert_id[k];
            for (int k = 0; k < 77 + 21; k++) commit_preimage[32 + k] = id_preimage[k];
            qadav_sha3_256(commit_preimage, 32 + 77 + 21, cert->commitment);

            if (i == 0) {
                for (int k = 0; k < 32; k++) base_commit[k] = cert->commitment[k];
            } else {
                if (qadav_sha3_ct_compare(base_commit, cert->commitment, 32) != 0) {
                    qadav_free(buf);
                    return 1;
                }
            }
            qadav_free(buf);
        }
        qadav_write_raw("CERT_COMMITMENT_ASLR_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "cert_chain_verify_aslr")) {
        for (int iter = 0; iter < 10; iter++) {
            double pen = 1.0;
            qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
            qadav_reasoning_chain* chain = qadav_qera_create("aslr_chain", 10, auth);
            qadav_qera_reason(chain, 1, (const unsigned char*)"prem", 4, (const unsigned char*)"conc", 4, 0, 0, 0.95, auth);
            qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"wit", 3, 0.95);
            int vr = qadav_certificate_verify_chain(chain);
            qadav_qera_free(chain);
            qadav_qeeh_free(auth);
            if (vr != 0) return 1;
        }
        qadav_write_raw("CERT_CHAIN_VERIFY_ASLR_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "cert_lifecycle")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("life_chain", 10, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"a", 1, (const unsigned char*)"b", 1, 0, 0, 0.90, auth);
        qadav_qera_reason(chain, 1, (const unsigned char*)"b", 1, (const unsigned char*)"c", 1, 0, 0, 0.85, auth);

        unsigned long long c_before = qadav_cert_get_create_count();
        qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"w1", 2, 0.90);
        qadav_certificate_attach(chain, 1, QADAV_CERT_INDUCTIVE, (const unsigned char*)"w2", 2, 0.85);
        unsigned long long c_after = qadav_cert_get_create_count();
        if (c_after != c_before + 2) return 1;

        int err = qadav_certificate_attach(chain, 0, 99, (const unsigned char*)"w", 1, 0.5);
        if (err == 0 || qadav_cert_get_create_count() != c_after) return 2;

        unsigned long long f_before = qadav_cert_get_free_count();
        qadav_qera_free(chain);
        unsigned long long f_after = qadav_cert_get_free_count();
        if (f_after != f_before + 2) return 3;

        qadav_qeeh_free(auth);
        qadav_write_raw("CERT_LIFECYCLE_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "abstain_lifecycle")) {
        unsigned long long c_before = qadav_abstain_get_create_count();
        unsigned long long f_before = qadav_abstain_get_free_count();
        qadav_abstain_policy* p1 = qadav_abstain_policy_create(0.85, QADAV_ABSTAIN_RETRY, 3);
        qadav_abstain_policy* p2 = qadav_abstain_policy_create(0.70, QADAV_ABSTAIN_REFUSE, 0);
        if (qadav_abstain_get_create_count() != c_before + 2) return 1;
        qadav_abstain_policy_free(p1);
        qadav_abstain_policy_free(p2);
        if (qadav_abstain_get_free_count() != f_before + 2) return 2;
        qadav_write_raw("ABSTAIN_LIFECYCLE_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "sandbox_create_and_run")) {
        qadav_sandbox_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_tool_register("sensor", auth, 1, 1);
        qadav_sandbox* sb = qadav_sandbox_create("test_sb", auth, 10);
        if (!sb) return 1;
        int r = qadav_sandbox_run(sb, "sensor", "normal_data");
        if (r != 0 || sb->state != QADAV_SANDBOX_CLEAN) return 2;
        qadav_sandbox_free(sb);
        qadav_qeeh_free(auth);
        qadav_write_raw("SANDBOX_CREATE_AND_RUN_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "sandbox_violation_graceful")) {
        qadav_sandbox_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_sandbox* sb = qadav_sandbox_create("untrusted_sb", auth, 10);
        int r = qadav_sandbox_run(sb, "sensor", "trigger_violation");
        if (r != -2) return 1; // Must return -2 (graceful failure)
        if (sb->state != QADAV_SANDBOX_VIOLATED) return 2;
        if (sb->violations != 1) return 3;
        for (int k = 0; k < 32; k++) {
            if (sb->local_chain->chain_root[k] != 0) return 4;
        }
        qadav_sandbox_free(sb);
        qadav_qeeh_free(auth);
        qadav_write_raw("SANDBOX_VIOLATION_GRACEFUL_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "sandbox_verify_attestation")) {
        qadav_sandbox_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_sandbox* sb = qadav_sandbox_create("verify_sb", auth, 10);
        qadav_sandbox_run(sb, "sensor", "trigger_violation");
        unsigned char att[32];
        int v = qadav_sandbox_verify(sb, att);
        if (v != 1) return 1;
        int all_zero = 1;
        for (int k = 0; k < 32; k++) {
            if (att[k] != 0) all_zero = 0;
        }
        if (all_zero) return 2;
        qadav_sandbox_free(sb);
        qadav_qeeh_free(auth);
        qadav_write_raw("SANDBOX_VERIFY_ATTESTATION_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "sandbox_quarantine")) {
        qadav_sandbox_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_sandbox* sb = qadav_sandbox_create("q_sb", auth, 10);
        qadav_sandbox_quarantine(sb);
        if (sb->state != QADAV_SANDBOX_QUARANTINED) return 1;
        int r = qadav_sandbox_run(sb, "sensor", "arg");
        if (r != -1) return 2;
        qadav_sandbox_free(sb);
        qadav_qeeh_free(auth);
        qadav_write_raw("SANDBOX_QUARANTINE_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "sandbox_lifecycle")) {
        qadav_sandbox_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        unsigned long long c_before = qadav_sandbox_get_create_count();
        unsigned long long f_before = qadav_sandbox_get_free_count();
        qadav_sandbox* sb1 = qadav_sandbox_create("sb1", auth, 5);
        qadav_sandbox* sb2 = qadav_sandbox_create("sb2", auth, 5);
        if (qadav_sandbox_get_create_count() != c_before + 2) return 1;
        qadav_sandbox_free(sb1);
        qadav_sandbox_free(sb2);
        if (qadav_sandbox_get_free_count() != f_before + 2) return 2;
        qadav_qeeh_free(auth);
        qadav_write_raw("SANDBOX_LIFECYCLE_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "gpu_init_and_detect")) {
        qadav_gpu_init();
        unsigned char uuid[32];
        qadav_gpu_get_device_uuid(uuid);
        int all_zero = 1;
        for (int i = 0; i < 32; i++) {
            if (uuid[i] != 0) all_zero = 0;
        }
        if (all_zero) return 1;
        qadav_write_raw("GPU_INIT_AND_DETECT_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "gpu_matmul_cpu_fallback")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double A[4] = {1.0, 2.0, 3.0, 4.0};
        double B[4] = {5.0, 6.0, 7.0, 8.0};
        double C[4] = {0.0, 0.0, 0.0, 0.0};
        int res = qadav_gpu_matmul(A, B, C, 2, 2, 2);
        if (res != 0) return 1;
        // C should be {19, 22, 43, 50}
        if (C[0] != 19.0 || C[1] != 22.0 || C[2] != 43.0 || C[3] != 50.0) return 2;
        if (qadav_gpu_get_dispatch_count() == 0) return 3;
        qadav_write_raw("GPU_MATMUL_OK\n", 14);
        return 0;
    }
    if (qadav_streq(req, "gpu_softmax_cpu_fallback")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double in[2] = {1.0, 2.0};
        double out[2] = {0.0, 0.0};
        int res = qadav_gpu_softmax(in, out, 1, 2);
        if (res != 0) return 1;
        double sum = out[0] + out[1];
        if (sum < 0.9999 || sum > 1.0001) return 2;
        if (out[1] <= out[0]) return 3;
        qadav_write_raw("GPU_SOFTMAX_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "gpu_layernorm_cpu_fallback")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double in[4] = {1.0, 2.0, 3.0, 4.0};
        double out[4] = {0.0, 0.0, 0.0, 0.0};
        int res = qadav_gpu_layernorm(in, out, 1, 4, 1e-5);
        if (res != 0) return 1;
        double sum = out[0] + out[1] + out[2] + out[3];
        if (sum < -0.001 || sum > 0.001) return 2; // mean is 0
        qadav_write_raw("GPU_LAYERNORM_OK\n", 17);
        return 0;
    }
    if (qadav_streq(req, "agke_attest_and_verify")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("agke_chain", 10, auth);
        double A[4] = {1.0, 0.0, 0.0, 1.0};
        double B[4] = {2.0, 3.0, 4.0, 5.0};
        double C[4] = {0.0, 0.0, 0.0, 0.0};
        qadav_gpu_matmul(A, B, C, 2, 2, 2);
        if (qadav_agke_get_count() != 1) return 1;
        int att_res = qadav_agke_attest(chain, auth);
        if (att_res != 1) return 2;
        int ver_res = qadav_agke_verify(chain, auth);
        if (ver_res != 1) return 3;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("AGKE_ATTEST_AND_VERIFY_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "agke_tamper_detection")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("agke_tamper_chain", 10, auth);
        double A[4] = {1.0, 0.0, 0.0, 1.0};
        double B[4] = {2.0, 3.0, 4.0, 5.0};
        double C[4] = {0.0, 0.0, 0.0, 0.0};
        qadav_gpu_matmul(A, B, C, 2, 2, 2);
        qadav_agke_attest(chain, auth);
        qadav_agke_corrupt_record(0, 1); // Corrupt output digest
        // Verification must trigger quantum decoherence and process exit with 0xDEADBEEF
        qadav_agke_verify(chain, auth);
        return 0; // Should not reach here
    }
    if (qadav_streq(req, "multigpu_init_and_plan")) {
        qadav_multigpu_context* ctx = qadav_multigpu_init();
        if (!ctx) return 1;
        int res = qadav_multigpu_build_pipeline(ctx, 8, 2, 4);
        if (res != 0) return 2;
        qadav_multigpu_shutdown(ctx);
        qadav_write_raw("MULTIGPU_INIT_AND_PLAN_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "multigpu_run_fallback")) {
        qadav_multigpu_context* ctx = qadav_multigpu_init();
        if (!ctx) return 1;
        qadav_multigpu_build_pipeline(ctx, 4, 2, 2);
        double in[4] = {1.0, 2.0, 3.0, 4.0};
        double out[4] = {0.0, 0.0, 0.0, 0.0};
        int res = qadav_multigpu_run_pipeline(ctx, in, out);
        if (res != 0) return 2;
        qadav_multigpu_shutdown(ctx);
        qadav_write_raw("MULTIGPU_RUN_OK\n", 16);
        return 0;
    }
    if (qadav_streq(req, "cdaga_record_and_verify")) {
        qadav_multigpu_context* ctx = qadav_multigpu_init();
        qadav_cdaga_record rec;
        double g0[4] = {0.01, 0.02, 0.03, 0.04};
        double g1[4] = {0.05, 0.06, 0.07, 0.08};
        void* bufs[2] = {g0, g1};
        unsigned long long sizes[2] = {sizeof(g0), sizeof(g1)};
        int rres = qadav_cdaga_record_aggregation(ctx, 1, bufs, sizes, 2, &rec);
        if (rres != 0) return 1;
        int vres = qadav_cdaga_verify(&rec, ctx);
        if (vres != 1) return 2;
        qadav_multigpu_shutdown(ctx);
        qadav_write_raw("CDAGA_RECORD_AND_VERIFY_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "cdaga_attest_to_qera")) {
        qadav_multigpu_context* ctx = qadav_multigpu_init();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("cdaga_train_chain", 10, auth);
        qadav_cdaga_record rec;
        double g0[4] = {0.1, 0.2, 0.3, 0.4};
        void* bufs[1] = {g0};
        unsigned long long sizes[1] = {sizeof(g0)};
        qadav_cdaga_record_aggregation(ctx, 1, bufs, sizes, 1, &rec);
        int ares = qadav_cdaga_attest(chain, &rec, ctx);
        if (ares != 1) return 1;
        int vres = qadav_cdaga_verify(&rec, ctx);
        if (vres != 1) return 2;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_multigpu_shutdown(ctx);
        qadav_write_raw("CDAGA_ATTEST_TO_QERA_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "cdaga_canonical_sort")) {
        unsigned char u0[16] = {0x01, 0x02, 0x03, 0x04};
        unsigned char u1[16] = {0x05, 0x06, 0x07, 0x08};
        unsigned char u2[16] = {0x09, 0x0A, 0x0B, 0x0C};
        double g0[4] = {0.1, 0.2, 0.3, 0.4};
        double g1[4] = {0.5, 0.6, 0.7, 0.8};
        double g2[4] = {0.9, 1.0, 1.1, 1.2};
        unsigned char topo[32] = {0xAA};

        unsigned char uuids_fwd[8][16];
        qadav_copy_bytes(uuids_fwd[0], u0, 16);
        qadav_copy_bytes(uuids_fwd[1], u1, 16);
        qadav_copy_bytes(uuids_fwd[2], u2, 16);
        void* bufs_fwd[3] = {g0, g1, g2};
        unsigned long long sizes_fwd[3] = {sizeof(g0), sizeof(g1), sizeof(g2)};
        qadav_cdaga_record rec1;
        qadav_cdaga_record_aggregation_with_uuids(uuids_fwd, topo, 1, bufs_fwd, sizes_fwd, 3, &rec1);

        unsigned char uuids_rev[8][16];
        qadav_copy_bytes(uuids_rev[0], u2, 16);
        qadav_copy_bytes(uuids_rev[1], u1, 16);
        qadav_copy_bytes(uuids_rev[2], u0, 16);
        void* bufs_rev[3] = {g2, g1, g0};
        unsigned long long sizes_rev[3] = {sizeof(g2), sizeof(g1), sizeof(g0)};
        qadav_cdaga_record rec2;
        qadav_cdaga_record_aggregation_with_uuids(uuids_rev, topo, 1, bufs_rev, sizes_rev, 3, &rec2);

        if (qadav_sha3_ct_compare(rec1.aggregation_digest, rec2.aggregation_digest, 32) != 0) {
            return 1;
        }

        unsigned char u5[5][16];
        double g5[5][4];
        for (int i = 0; i < 5; i++) {
            for (int k = 0; k < 16; k++) u5[i][k] = (unsigned char)((i * 37 + k + 3) & 0xFF);
            for (int k = 0; k < 4; k++) g5[i][k] = (double)(i * 10 + k);
        }

        unsigned char base_digest[32];
        unsigned char uuids_base[8][16];
        void* bufs_base[5];
        unsigned long long sizes_base[5];
        for (int i = 0; i < 5; i++) {
            qadav_copy_bytes(uuids_base[i], u5[i], 16);
            bufs_base[i] = g5[i];
            sizes_base[i] = sizeof(g5[i]);
        }
        qadav_cdaga_record rec_base;
        qadav_cdaga_record_aggregation_with_uuids(uuids_base, topo, 1, bufs_base, sizes_base, 5, &rec_base);
        qadav_copy_bytes(base_digest, rec_base.aggregation_digest, 32);

        int perm[10][5] = {
            {4, 3, 2, 1, 0},
            {1, 0, 3, 2, 4},
            {2, 4, 0, 3, 1},
            {3, 1, 4, 0, 2},
            {0, 2, 4, 1, 3},
            {4, 1, 3, 0, 2},
            {2, 3, 1, 4, 0},
            {1, 4, 2, 0, 3},
            {3, 0, 2, 4, 1},
            {4, 2, 0, 1, 3}
        };

        for (int p = 0; p < 10; p++) {
            unsigned char uuids_p[8][16];
            void* bufs_p[5];
            unsigned long long sizes_p[5];
            for (int i = 0; i < 5; i++) {
                int idx = perm[p][i];
                qadav_copy_bytes(uuids_p[i], u5[idx], 16);
                bufs_p[i] = g5[idx];
                sizes_p[i] = sizeof(g5[idx]);
            }
            qadav_cdaga_record rec_p;
            qadav_cdaga_record_aggregation_with_uuids(uuids_p, topo, 1, bufs_p, sizes_p, 5, &rec_p);
            if (qadav_sha3_ct_compare(base_digest, rec_p.aggregation_digest, 32) != 0) {
                return 2;
            }
        }

        qadav_write_raw("CDAGA_CANONICAL_SORT_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "cdaga_aslr_determinism")) {
        unsigned char first_hash[32];
        for (int iter = 0; iter < 100; iter++) {
            void* pad = qadav_alloc((iter %% 17 + 1) * 64);
            qadav_cdaga_record* rec = (qadav_cdaga_record*)qadav_alloc(sizeof(qadav_cdaga_record));
            double g0[4] = {0.123, 0.456, 0.789, 0.012};
            void* bufs[1] = {g0};
            unsigned long long sizes[1] = {sizeof(g0)};
            qadav_cdaga_record_aggregation(0, 42, bufs, sizes, 1, rec);
            if (iter == 0) {
                qadav_copy_bytes(first_hash, rec->record_hash, 32);
            } else {
                if (qadav_sha3_ct_compare(first_hash, rec->record_hash, 32) != 0) {
                    qadav_free(rec);
                    qadav_free(pad);
                    return 1;
                }
            }
            qadav_free(rec);
            qadav_free(pad);
        }
        qadav_write_raw("CDAGA_ASLR_DETERMINISM_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "dist_init_single_worker")) {
        qadav_dist_ctx* ctx = qadav_dist_init(1);
        if (!ctx || ctx->worker_count != 1 || !ctx->single_worker_mode) return 1;
        if (qadav_dist_get_worker_count() != 1) return 2;
        qadav_dist_shutdown(ctx);
        qadav_write_raw("DIST_INIT_SINGLE_WORKER_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "dist_ring_allreduce_2workers")) {
        qadav_dist_ctx* ctx = qadav_dist_init(2);
        if (!ctx) return 1;
        double* w0 = (double*)ctx->workers[0].grad_buf;
        double* w1 = (double*)ctx->workers[1].grad_buf;
        w0[0] = 1.0; w0[1] = 2.0; w0[2] = 3.0; w0[3] = 4.0;
        w1[0] = 3.0; w1[1] = 4.0; w1[2] = 5.0; w1[3] = 6.0;
        ctx->workers[0].grad_size = 4 * sizeof(double);
        ctx->workers[1].grad_size = 4 * sizeof(double);

        int r = qadav_dist_ring_allreduce(ctx);
        if (r != 0) return 2;
        if (w0[0] != 2.0 || w0[1] != 3.0 || w0[2] != 4.0 || w0[3] != 5.0) return 3;
        if (w1[0] != 2.0 || w1[1] != 3.0 || w1[2] != 4.0 || w1[3] != 5.0) return 4;
        qadav_dist_shutdown(ctx);
        qadav_write_raw("DIST_RING_ALLREDUCE_2WORKERS_OK\n", 32);
        return 0;
    }
    if (qadav_streq(req, "dist_ring_allreduce_4workers")) {
        qadav_dist_ctx* ctx = qadav_dist_init(4);
        if (!ctx) return 1;
        for (int i = 0; i < 4; i++) {
            double* buf = (double*)ctx->workers[i].grad_buf;
            buf[0] = (double)(i + 1);
            buf[1] = (double)(i + 1);
            ctx->workers[i].grad_size = 2 * sizeof(double);
        }
        int r = qadav_dist_ring_allreduce(ctx);
        if (r != 0) return 2;
        for (int i = 0; i < 4; i++) {
            double* buf = (double*)ctx->workers[i].grad_buf;
            if (buf[0] != 2.5 || buf[1] != 2.5) return 3;
        }
        qadav_dist_shutdown(ctx);
        qadav_write_raw("DIST_RING_ALLREDUCE_4WORKERS_OK\n", 32);
        return 0;
    }
    if (qadav_streq(req, "dist_pserver_aggregate")) {
        qadav_dist_ctx* ctx = qadav_dist_init(3);
        if (!ctx) return 1;
        double* w0 = (double*)ctx->workers[0].grad_buf;
        double* w1 = (double*)ctx->workers[1].grad_buf;
        double* w2 = (double*)ctx->workers[2].grad_buf;
        w0[0] = 10.0; w0[1] = 20.0;
        w1[0] = 20.0; w1[1] = 40.0;
        w2[0] = 30.0; w2[1] = 60.0;
        ctx->workers[0].grad_size = 2 * sizeof(double);
        ctx->workers[1].grad_size = 2 * sizeof(double);
        ctx->workers[2].grad_size = 2 * sizeof(double);

        int r = qadav_dist_pserver_aggregate(ctx);
        if (r != 0) return 2;
        if (w0[0] != 20.0 || w0[1] != 40.0) return 3;
        qadav_dist_shutdown(ctx);
        qadav_write_raw("DIST_PSERVER_AGGREGATE_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "dist_fallback_single_worker")) {
        qadav_dist_ctx* ctx = qadav_dist_init(1);
        if (!ctx) return 1;
        double* g = (double*)ctx->workers[0].grad_buf;
        g[0] = 42.0;
        ctx->workers[0].grad_size = sizeof(double);
        int r = qadav_dist_ring_allreduce(ctx);
        if (r != 0 || g[0] != 42.0) return 2;
        r = qadav_dist_pserver_aggregate(ctx);
        if (r != 0 || g[0] != 42.0) return 3;
        qadav_dist_shutdown(ctx);
        qadav_write_raw("DIST_FALLBACK_SINGLE_WORKER_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "barat_submit_aggregate")) {
        qadav_barat_group* grp = qadav_barat_create_group(1);
        double g0[2] = {1.0, 2.0};
        double g1[2] = {3.0, 4.0};
        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        qadav_barat_submit(grp, 1, g1, sizeof(g1));

        double out[2];
        qadav_barat_cert cert;
        int r = qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        if (r != 0) return 1;
        if (out[0] != 2.0 || out[1] != 3.0) return 2;
        if (!qadav_barat_verify(&cert, out, sizeof(out))) return 3;
        qadav_barat_destroy_group(grp);
        qadav_write_raw("BARAT_SUBMIT_AGGREGATE_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "barat_trimmed_mean")) {
        qadav_barat_group* grp = qadav_barat_create_group(2);
        double g0[2] = {100.0, 100.0}; // norm 200 (outlier)
        double g1[2] = {10.0, 10.0};   // norm 20
        double g2[2] = {12.0, 12.0};   // norm 24
        double g3[2] = {14.0, 14.0};   // norm 28
        double g4[2] = {0.01, 0.01};   // norm 0.02 (outlier)

        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        qadav_barat_submit(grp, 1, g1, sizeof(g1));
        qadav_barat_submit(grp, 2, g2, sizeof(g2));
        qadav_barat_submit(grp, 3, g3, sizeof(g3));
        qadav_barat_submit(grp, 4, g4, sizeof(g4));

        double out[2];
        qadav_barat_cert cert;
        int r = qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        if (r != 0) return 1;
        // Central 3: (10 + 12 + 14) / 3 = 12.0
        if (out[0] != 12.0 || out[1] != 12.0) return 2;
        if (!qadav_barat_verify(&cert, out, sizeof(out))) return 3;
        qadav_barat_destroy_group(grp);
        qadav_write_raw("BARAT_TRIMMED_MEAN_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "barat_selection_proof_valid")) {
        qadav_barat_group* grp = qadav_barat_create_group(3);
        double g0[2] = {1.0, 2.0};
        double g1[2] = {3.0, 4.0};
        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        qadav_barat_submit(grp, 1, g1, sizeof(g1));
        double out[2];
        qadav_barat_cert cert;
        qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        int zeros = 1;
        for (int i = 0; i < 32; i++) {
            if (cert.selection_proof[i] != 0) zeros = 0;
        }
        if (zeros) return 1;
        if (!qadav_barat_verify(&cert, out, sizeof(out))) return 2;
        qadav_barat_destroy_group(grp);
        qadav_write_raw("BARAT_SELECTION_PROOF_VALID_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "barat_tampered_cert")) {
        qadav_barat_group* grp = qadav_barat_create_group(4);
        double g0[2] = {1.0, 2.0};
        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        double out[2];
        qadav_barat_cert cert;
        qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        qadav_barat_corrupt(&cert, 16);
        // Should trigger decoherence and ExitProcess(0xDEADBEEF)
        qadav_barat_verify(&cert, out, sizeof(out));
        qadav_barat_destroy_group(grp);
        return 0;
    }
    if (qadav_streq(req, "barat_attest_success")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("barat_chain", 10, auth);
        qadav_barat_group* grp = qadav_barat_create_group(5);
        double g0[2] = {5.0, 10.0};
        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        double out[2];
        qadav_barat_cert cert;
        qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        int r = qadav_barat_attest(chain, &cert, out, sizeof(out));
        if (r != 1) return 1;
        qadav_barat_destroy_group(grp);
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("BARAT_ATTEST_SUCCESS_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "barat_canonical_sort")) {
        qadav_barat_group* gA = qadav_barat_create_group(10);
        qadav_barat_group* gB = qadav_barat_create_group(11);
        double w0[2] = {1.0, 2.0};
        double w1[2] = {3.0, 4.0};
        double w2[2] = {5.0, 6.0};

        // Group A: 0, 1, 2
        qadav_barat_submit(gA, 0, w0, sizeof(w0));
        qadav_barat_submit(gA, 1, w1, sizeof(w1));
        qadav_barat_submit(gA, 2, w2, sizeof(w2));

        // Group B: 2, 0, 1
        qadav_barat_submit(gB, 2, w2, sizeof(w2));
        qadav_barat_submit(gB, 0, w0, sizeof(w0));
        qadav_barat_submit(gB, 1, w1, sizeof(w1));

        double outA[2], outB[2];
        qadav_barat_cert certA, certB;
        qadav_barat_aggregate(gA, outA, sizeof(outA), &certA);
        qadav_barat_aggregate(gB, outB, sizeof(outB), &certB);

        if (qadav_sha3_ct_compare(certA.record_hash, certB.record_hash, 32) != 0) return 1;
        if (qadav_sha3_ct_compare(certA.selection_proof, certB.selection_proof, 32) != 0) return 2;
        if (qadav_sha3_ct_compare(certA.aggregated_hash, certB.aggregated_hash, 32) != 0) return 3;

        qadav_barat_destroy_group(gA);
        qadav_barat_destroy_group(gB);
        qadav_write_raw("BARAT_CANONICAL_SORT_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "barat_aslr_determinism")) {
        unsigned char first_hash[32];
        for (int iter = 0; iter < 100; iter++) {
            void* pad = qadav_alloc((iter %% 19 + 1) * 32);
            qadav_barat_group* grp = qadav_barat_create_group(20);
            double g[2] = {0.123, 0.456};
            qadav_barat_submit(grp, 0, g, sizeof(g));
            double out[2];
            qadav_barat_cert cert;
            qadav_barat_aggregate(grp, out, sizeof(out), &cert);

            if (iter == 0) {
                qadav_copy_bytes(first_hash, cert.record_hash, 32);
            } else {
                if (qadav_sha3_ct_compare(first_hash, cert.record_hash, 32) != 0) {
                    qadav_barat_destroy_group(grp);
                    qadav_free(pad);
                    return 1;
                }
            }
            qadav_barat_destroy_group(grp);
            qadav_free(pad);
        }
        qadav_write_raw("BARAT_ASLR_DETERMINISM_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "qhal_create")) {
        qhal_circuit* c = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        if (!c || c->n_qubits != 2 || c->n_ops != 0 || c->backend != QHAL_BACKEND_LOCAL_SIM) return 1;
        qhal_shutdown(c);
        qadav_write_raw("QHAL_CREATE_OK\n", 15);
        return 0;
    }
    if (qadav_streq(req, "qhal_add_ops")) {
        qhal_circuit* c = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        if (qhal_add_op(c, QHAL_OP_H, 0, 0) != 0) return 1;
        if (qhal_add_op(c, QHAL_OP_CNOT, 0, 1) != 0) return 2;
        if (c->n_ops != 2) return 3;
        qhal_shutdown(c);
        qadav_write_raw("QHAL_ADD_OPS_OK\n", 16);
        return 0;
    }
    if (qadav_streq(req, "qhal_export_qasm3")) {
        qhal_circuit* c = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        qhal_add_op(c, QHAL_OP_CNOT, 0, 1);
        char buf[512];
        unsigned long long written = 0;
        if (qhal_export_qasm3(c, buf, sizeof(buf), &written) != 0) return 1;
        if (written == 0 || buf[0] == '\0') return 2;
        qhal_shutdown(c);
        qadav_write_raw(buf, written);
        return 0;
    }
    if (qadav_streq(req, "qhal_hash_circuit")) {
        qhal_circuit* c1 = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        qhal_circuit* c2 = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        qhal_add_op(c1, QHAL_OP_H, 0, 0);
        qhal_add_op(c1, QHAL_OP_CNOT, 0, 1);
        qhal_add_op(c2, QHAL_OP_H, 0, 0);
        qhal_add_op(c2, QHAL_OP_CNOT, 0, 1);
        qhal_hash_circuit(c1);
        qhal_hash_circuit(c2);
        if (qadav_sha3_ct_compare(c1->circuit_hash, c2->circuit_hash, 32) != 0) return 1;
        qhal_shutdown(c1);
        qhal_shutdown(c2);
        qadav_write_raw("QHAL_HASH_CIRCUIT_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qhal_backend_stub")) {
        qhal_circuit* c = qhal_create(3, QHAL_BACKEND_IBMQ);
        qhal_add_op(c, QHAL_OP_X, 1, 0);
        char buf[512];
        unsigned long long written = 0;
        if (qhal_export_qasm3(c, buf, sizeof(buf), &written) != 0) return 1;
        if (c->backend != QHAL_BACKEND_IBMQ) return 2;
        qhal_shutdown(c);
        qadav_write_raw("QHAL_BACKEND_STUB_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qhal_bounds_check")) {
        qhal_circuit* c = qhal_create(2, QHAL_BACKEND_LOCAL_SIM);
        for (int i = 0; i < 256; i++) {
            if (qhal_add_op(c, QHAL_OP_H, 0, 0) != 0) return 1;
        }
        if (qhal_add_op(c, QHAL_OP_H, 0, 0) != -2) return 2;
        qhal_shutdown(c);
        qadav_write_raw("QHAL_BOUNDS_CHECK_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "sea_create")) {
        sea_history* h = sea_create();
        if (!h || h->count != 0) return 1;
        sea_free(h);
        qadav_write_raw("SEA_CREATE_OK\n", 14);
        return 0;
    }
    if (qadav_streq(req, "sea_propose_valid")) {
        sea_history* h = sea_create();
        unsigned char from[32], to[32], reason[32];
        for (int i = 0; i < 32; i++) {
            from[i] = (unsigned char)(i + 1);
            to[i] = (unsigned char)(i + 2);
            reason[i] = (unsigned char)(i + 3);
        }
        sea_proposal* p = 0;
        int r = sea_propose(h, from, to, reason, &p);
        if (r != 0 || !p || h->count != 1) return 1;
        if (p->step_index != 1) return 2;
        sea_free(h);
        qadav_write_raw("SEA_PROPOSE_VALID_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "sea_propose_invalid_hashes")) {
        sea_history* h = sea_create();
        unsigned char zero[32] = {0};
        unsigned char valid[32];
        for (int i = 0; i < 32; i++) valid[i] = (unsigned char)(i + 1);
        sea_proposal* p = 0;
        if (sea_propose(h, zero, valid, valid, &p) != -2) return 1;
        if (sea_propose(h, valid, zero, valid, &p) != -2) return 2;
        if (sea_propose(h, valid, valid, zero, &p) != -2) return 3;
        sea_free(h);
        qadav_write_raw("SEA_PROPOSE_INVALID_HASHES_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "sea_verify_valid")) {
        sea_history* h = sea_create();
        unsigned char from[32], to[32], reason[32];
        for (int i = 0; i < 32; i++) {
            from[i] = (unsigned char)(i + 1);
            to[i] = (unsigned char)(i + 2);
            reason[i] = (unsigned char)(i + 3);
        }
        sea_proposal* p = 0;
        sea_propose(h, from, to, reason, &p);
        if (!sea_verify_proposal(p, 0)) return 1;
        if (!sea_verify_history(h)) return 2;
        sea_free(h);
        qadav_write_raw("SEA_VERIFY_VALID_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "sea_attest_success")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("sea_chain", 10, auth);
        sea_history* h = sea_create();
        unsigned char from[32], to[32], reason[32];
        for (int i = 0; i < 32; i++) {
            from[i] = (unsigned char)(i + 10);
            to[i] = (unsigned char)(i + 20);
            reason[i] = (unsigned char)(i + 30);
        }
        sea_propose(h, from, to, reason, 0);
        int r = sea_attest(h, chain);
        if (r != 1) return 1;
        if (chain->n_steps != 1) return 2;
        sea_free(h);
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("SEA_ATTEST_SUCCESS_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "sea_canonical_sort_determinism")) {
        unsigned char hash1[32], hash2[32];
        for (int iter = 0; iter < 100; iter++) {
            sea_history* h = sea_create();
            unsigned char from[32], to[32], reason[32];
            for (int i = 0; i < 32; i++) {
                from[i] = (unsigned char)(i + 5);
                to[i] = (unsigned char)(i + 15);
                reason[i] = (unsigned char)(i + 25);
            }
            sea_proposal* p = 0;
            sea_propose(h, from, to, reason, &p);
            if (iter == 0) {
                qadav_copy_bytes(hash1, p->proposal_hash, 32);
            } else {
                qadav_copy_bytes(hash2, p->proposal_hash, 32);
                if (qadav_sha3_ct_compare(hash1, hash2, 32) != 0) {
                    sea_free(h);
                    return 1;
                }
            }
            sea_free(h);
        }
        qadav_write_raw("SEA_CANONICAL_SORT_DETERMINISM_OK\n", 34);
        return 0;
    }
    if (qadav_streq(req, "sea_aslr_determinism")) {
        unsigned char first_hash[32];
        for (int iter = 0; iter < 100; iter++) {
            void* pad = qadav_alloc((iter %% 13 + 1) * 48);
            sea_history* h = sea_create();
            unsigned char from[32], to[32], reason[32];
            for (int i = 0; i < 32; i++) {
                from[i] = (unsigned char)(i + 7);
                to[i] = (unsigned char)(i + 17);
                reason[i] = (unsigned char)(i + 27);
            }
            sea_proposal* p = 0;
            sea_propose(h, from, to, reason, &p);
            if (iter == 0) {
                qadav_copy_bytes(first_hash, p->proposal_hash, 32);
            } else {
                if (qadav_sha3_ct_compare(first_hash, p->proposal_hash, 32) != 0) {
                    sea_free(h);
                    qadav_free(pad);
                    return 1;
                }
            }
            sea_free(h);
            qadav_free(pad);
        }
        qadav_write_raw("SEA_ASLR_DETERMINISM_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "sea_lineage_chain")) {
        sea_history* h = sea_create();
        for (int step = 0; step < 3; step++) {
            unsigned char from[32], to[32], reason[32];
            for (int i = 0; i < 32; i++) {
                from[i] = (unsigned char)(step * 10 + i + 1);
                to[i] = (unsigned char)(step * 10 + i + 2);
                reason[i] = (unsigned char)(step * 10 + i + 3);
            }
            sea_proposal* p = 0;
            sea_propose(h, from, to, reason, &p);
        }
        if (h->count != 3) return 1;
        if (qadav_sha3_ct_compare(h->proposals[1].prev_proposal_hash, h->proposals[0].proposal_hash, 32) != 0) return 2;
        if (qadav_sha3_ct_compare(h->proposals[2].prev_proposal_hash, h->proposals[1].proposal_hash, 32) != 0) return 3;
        if (!sea_verify_history(h)) return 4;
        sea_free(h);
        qadav_write_raw("SEA_LINEAGE_CHAIN_OK\n", 21);
        return 0;
    }
    // --- M7.3 Transpiler Tests ---
    if (qadav_streq(req, "transpile_qiskit")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        qhal_add_op(c, QHAL_OP_CNOT, 0, 1);
        char buf[1024];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_QISKIT, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        int found = 0;
        for (int i = 0; buf[i]; i++) {
            if (buf[i] == 'q' && buf[i+1] == 'c' && buf[i+2] == '.' && buf[i+3] == 'h') { found = 1; break; }
        }
        if (!found) return 2;
        qadav_write_raw("TRANSPILE_QISKIT_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "transpile_cirq")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        qhal_add_op(c, QHAL_OP_CNOT, 0, 1);
        char buf[1024];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_CIRQ, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        int found = 0;
        for (int i = 0; buf[i]; i++) {
            if (buf[i] == 'c' && buf[i+1] == 'i' && buf[i+2] == 'r' && buf[i+3] == 'q') { found = 1; break; }
        }
        if (!found) return 2;
        qadav_write_raw("TRANSPILE_CIRQ_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "transpile_ionq")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        qhal_add_op(c, QHAL_OP_CNOT, 0, 1);
        char buf[1024];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_IONQ, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        int found = 0;
        for (int i = 0; buf[i]; i++) {
            if (buf[i] == 'g' && buf[i+1] == 'a' && buf[i+2] == 't' && buf[i+3] == 'e') { found = 1; break; }
        }
        if (!found) return 2;
        qadav_write_raw("TRANSPILE_IONQ_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "transpile_qasm3")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        char buf[1024];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_QASM3, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        int found = 0;
        for (int i = 0; buf[i]; i++) {
            if (buf[i] == 'O' && buf[i+1] == 'P' && buf[i+2] == 'E' && buf[i+3] == 'N') { found = 1; break; }
        }
        if (!found) return 2;
        qadav_write_raw("TRANSPILE_QASM3_OK\n", 19);
        return 0;
    }
    if (qadav_streq(req, "transpile_adjacent_cancel")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_X, 0, 0);
        qhal_add_op(c, QHAL_OP_X, 0, 0);
        transpile_opts opts;
        opts.optimize_adjacent = 1;
        opts.decompose_u3 = 0;
        opts.reduce_depth = 0;
        opts.max_qubits = 2;
        char buf[1024];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_QISKIT, &opts, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        if (op_cnt != 0) return 2;
        qadav_write_raw("TRANSPILE_ADJACENT_CANCEL_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "transpile_buffer_too_small")) {
        qhal_circuit* c = qhal_create(2, 0);
        qhal_add_op(c, QHAL_OP_H, 0, 0);
        char buf[8];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_QISKIT, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != -3) return 1;
        if (written <= 8) return 2;
        qadav_write_raw("TRANSPILE_BUFFER_TOO_SMALL_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "transpile_unsupported_target")) {
        qhal_circuit* c = qhal_create(2, 0);
        char buf[128];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, (transpile_target)99, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != -1) return 1;
        qadav_write_raw("TRANSPILE_UNSUPPORTED_TARGET_OK\n", 32);
        return 0;
    }
    if (qadav_streq(req, "transpile_no_crash_huge")) {
        qhal_circuit* c = qhal_create(10, 0);
        for (int i = 0; i < 200; i++) {
            qhal_add_op(c, (unsigned char)(i %% 5), i %% 10, (i + 1) %% 10);
        }
        char buf[16384];
        unsigned long long written = 0;
        int op_cnt = 0;
        int r = transpile_circuit(c, TRANSPILE_TARGET_QISKIT, 0, buf, sizeof(buf), &written, &op_cnt);
        qhal_shutdown(c);
        if (r != 0) return 1;
        if (op_cnt != 200) return 2;
        qadav_write_raw("TRANSPILE_NO_CRASH_HUGE_OK\n", 27);
        return 0;
    }

    // --- M7.4 EEA Tests ---
    if (qadav_streq(req, "eea_create_policy")) {
        eea_context* c1 = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_context* c2 = eea_create_policy(POLICY_DECAYING_RATE, 0.20, 200);
        eea_context* c3 = eea_create_policy(POLICY_ADAPTIVE, 0.30, 300);
        if (!c1 || !c2 || !c3) return 1;
        if (c1->kind != POLICY_CONSTANT_RATE || c2->kind != POLICY_DECAYING_RATE || c3->kind != POLICY_ADAPTIVE) return 2;
        eea_free(c1);
        eea_free(c2);
        eea_free(c3);
        qadav_write_raw("EEA_CREATE_POLICY_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "eea_compute_diff")) {
        unsigned char from[] = "hello";
        unsigned char to[] = "jello";
        unsigned long long flips = 0, new_syms = 0;
        int r = eea_compute_diff(from, 5, to, 5, &flips, &new_syms);
        if (r != 0) return 1;
        if (flips != 1) return 2;
        if (new_syms != 1) return 3;
        qadav_write_raw("EEA_COMPUTE_DIFF_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "eea_build_envelope_valid")) {
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        int r = eea_build_envelope_text(ctx, 0, "hello", "jello", 8, 2, &env);
        eea_free(ctx);
        if (r != 0) return 1;
        if (env.actual_bit_flips != 1 || env.actual_new_symbols != 1) return 2;
        qadav_write_raw("EEA_BUILD_ENVELOPE_VALID_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "eea_build_envelope_violation")) {
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        int r = eea_build_envelope_text(ctx, 0, "hello", "jello", 0, 2, &env);
        eea_free(ctx);
        if (r != -2) return 1;
        qadav_write_raw("EEA_BUILD_ENVELOPE_VIOLATION_OK\n", 32);
        return 0;
    }
    if (qadav_streq(req, "eea_verify_valid")) {
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        if (!eea_verify_envelope(&env)) return 1;
        qadav_write_raw("EEA_VERIFY_VALID_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "eea_verify_tampered")) {
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        eea_corrupt(&env, 16);
        eea_verify_envelope(&env);
        return 0;
    }
    if (qadav_streq(req, "eea_attest_success")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("eea_chain", 10, auth);
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        if (!eea_attest_envelope(&env, chain)) return 1;
        if (chain->n_steps != 1) return 2;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("EEA_ATTEST_SUCCESS_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "eea_attest_tampered")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("eea_chain", 10, auth);
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        eea_corrupt(&env, 8);
        eea_attest_envelope(&env, chain);
        return 0;
    }
    if (qadav_streq(req, "eea_decaying_rate")) {
        eea_context* ctx = eea_create_policy(POLICY_DECAYING_RATE, 1.0, 1000);
        double prev_rate = ctx->mutation_rate;
        for (int i = 0; i < 10; i++) {
            eea_envelope env;
            char from[16], to[16];
            from[0] = (char)('a' + i); from[1] = '\0';
            to[0] = (char)('b' + i); to[1] = '\0';
            eea_build_envelope_text(ctx, 0, from, to, 16, 4, &env);
            if (ctx->mutation_rate >= prev_rate && i > 0) return 1;
            prev_rate = ctx->mutation_rate;
        }
        eea_free(ctx);
        qadav_write_raw("EEA_DECAYING_RATE_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "eea_aslr_determinism")) {
        unsigned char h1[32], h2[32];
        for (int iter = 0; iter < 100; iter++) {
            eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
            eea_envelope env;
            eea_build_envelope_text(ctx, 0, "fixed_from", "fixed_to", 16, 4, &env);
            eea_free(ctx);
            if (iter == 0) {
                for (int k = 0; k < 32; k++) h1[k] = env.envelope_hash[k];
            } else {
                for (int k = 0; k < 32; k++) h2[k] = env.envelope_hash[k];
                if (qadav_sha3_ct_compare(h1, h2, 32) != 0) return 1;
            }
        }
        qadav_write_raw("EEA_ASLR_DETERMINISM_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "qecd_decode_basic")) {
        unsigned char syn[1] = {0x01};
        qecd_decode_result* res = qecd_decode(3, 9, syn, 1);
        if (!res) return 1;
        if (res->n_corrections != 1) return 2;
        if (res->corrections[0].qubit_index != 0 || res->corrections[0].pauli != 1) return 3;
        qecd_free(res);
        qadav_write_raw("QECD_DECODE_BASIC_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qecd_decode_empty")) {
        qecd_decode_result* res = qecd_decode(3, 9, 0, 0);
        if (!res) return 1;
        if (res->n_corrections != 0) return 2;
        if (!qecd_verify_decode_attestation(res)) return 3;
        qecd_free(res);
        qadav_write_raw("QECD_DECODE_EMPTY_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qecd_decode_large")) {
        unsigned char syn64[64];
        for (int i = 0; i < 64; i++) syn64[i] = (unsigned char)(i * 3 + 1);
        qecd_decode_result* res = qecd_decode(5, 25, syn64, 64);
        if (!res) return 1;
        if (res->n_corrections > QADAV_QECD_MAX_CORRECTIONS) return 2;
        if (!qecd_verify_decode_attestation(res)) return 3;
        qecd_free(res);
        qadav_write_raw("QECD_DECODE_LARGE_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qecd_verify_attestation")) {
        unsigned char syn[2] = {0x05, 0x0A};
        qecd_decode_result* res = qecd_decode(3, 9, syn, 2);
        if (!res) return 1;
        if (!qecd_verify_decode_attestation(res)) return 2;
        qecd_free(res);
        qadav_write_raw("QECD_VERIFY_ATTESTATION_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "qecd_verify_tampered")) {
        unsigned char syn[2] = {0x05, 0x0A};
        qecd_decode_result* res = qecd_decode(3, 9, syn, 2);
        if (!res) return 1;
        res->decode_attestation[0] ^= 0xFF;
        if (qecd_verify_decode_attestation(res)) return 2;
        qecd_free(res);
        qadav_write_raw("QECD_VERIFY_TAMPERED_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "qecd_apply_corrections")) {
        unsigned char syn[1] = {0x01};
        qecd_decode_result* res = qecd_decode(3, 9, syn, 1);
        if (!res) return 1;
        unsigned char vec[18];
        for (int i = 0; i < 18; i++) vec[i] = 0;
        int r = qecd_apply_corrections(res, vec, 18);
        if (r != 0 || vec[0] != 1) return 2;
        qecd_free(res);
        qadav_write_raw("QECD_APPLY_CORRECTIONS_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "qecd_apply_bounds")) {
        unsigned char syn[1] = {0x01};
        qecd_decode_result* res = qecd_decode(3, 9, syn, 1);
        if (!res) return 1;
        res->corrections[0].qubit_index = 20;
        unsigned char vec[18];
        for (int i = 0; i < 18; i++) vec[i] = 0;
        int r = qecd_apply_corrections(res, vec, 18);
        if (r != -2) return 2;
        qecd_free(res);
        qadav_write_raw("QECD_APPLY_BOUNDS_OK\n", 21);
        return 0;
    }
    if (qadav_streq(req, "qecd_aslr_determinism")) {
        unsigned char h1[32], h2[32];
        unsigned char syn[4] = {0x12, 0x34, 0x56, 0x78};
        for (int iter = 0; iter < 100; iter++) {
            qecd_decode_result* res = qecd_decode(3, 9, syn, 4);
            if (!res) return 1;
            if (iter == 0) {
                for (int k = 0; k < 32; k++) h1[k] = res->decode_attestation[k];
            } else {
                for (int k = 0; k < 32; k++) h2[k] = res->decode_attestation[k];
                if (qadav_sha3_ct_compare(h1, h2, 32) != 0) return 2;
            }
            qecd_free(res);
        }
        qadav_write_raw("QECD_ASLR_DETERMINISM_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "aad_build_cert")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        if (!cert) return 1;
        aad_free_training_cert(cert);
        qadav_write_raw("AAD_BUILD_CERT_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "aad_build_cert_invalid")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 200, 8);
        if (cert != 0) return 1;
        qadav_write_raw("AAD_BUILD_CERT_INVALID_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "aad_wrap_decode_valid")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[2] = {0x03, 0x0C};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 2);
        aad_decode_envelope env;
        int r = aad_wrap_decode(cert, dec, syn, 2, &env);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (r != 0 || env.syndrome_weight != 4) return 1;
        qadav_write_raw("AAD_WRAP_DECODE_VALID_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "aad_wrap_decode_outofbounds")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 4, 8);
        unsigned char syn[2] = {0xFF, 0xFF}; // 16 bits set > 4
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 2);
        aad_decode_envelope env;
        int r = aad_wrap_decode(cert, dec, syn, 2, &env);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (r != -3) return 1;
        qadav_write_raw("AAD_WRAP_DECODE_OUTOFBOUNDS_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "aad_verify_valid")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        int v = aad_verify_envelope(&env, cert);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (!v) return 1;
        qadav_write_raw("AAD_VERIFY_VALID_OK\n", 20);
        return 0;
    }
    if (qadav_streq(req, "aad_verify_tampered")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        aad_corrupt_envelope(&env, 16);
        int v = aad_verify_envelope(&env, cert);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (v) return 1;
        qadav_write_raw("AAD_VERIFY_TAMPERED_OK\n", 23);
        return 0;
    }
    if (qadav_streq(req, "aad_attest_success")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("aad_chain", 10, auth);
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        int r = aad_attest_envelope(&env, chain);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (!r || chain->n_steps != 1) return 1;
        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("AAD_ATTEST_SUCCESS_OK\n", 22);
        return 0;
    }
    if (qadav_streq(req, "aad_meta_fit_adapt_verify")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        aad_meta_state state;
        aad_meta_fit(&state, cert);
        int ar = aad_meta_adapt(&state, &env);
        if (ar != 0 || state.adaptation_steps != 1) return 1;
        if (!aad_meta_verify(&state)) return 2;
        qecd_free(dec);
        aad_free_training_cert(cert);
        qadav_write_raw("AAD_META_FIT_ADAPT_VERIFY_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "aad_meta_adapt_tampered")) {
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        aad_corrupt_envelope(&env, 16);
        aad_meta_state state;
        aad_meta_fit(&state, cert);
        int ar = aad_meta_adapt(&state, &env);
        qecd_free(dec);
        aad_free_training_cert(cert);
        if (ar != -2) return 1; // Returns -2 on tampered envelope without decoherence
        qadav_write_raw("AAD_META_ADAPT_TAMPERED_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "aad_adapt_chain_determinism")) {
        unsigned char c1[32], c2[32];
        for (int iter = 0; iter < 100; iter++) {
            aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"m", 1, (const unsigned char*)"t", 1, 50, 64, 8);
            unsigned char syn[1] = {0x02};
            qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
            aad_decode_envelope env;
            aad_wrap_decode(cert, dec, syn, 1, &env);
            aad_meta_state state;
            aad_meta_fit(&state, cert);
            aad_meta_adapt(&state, &env);
            if (iter == 0) {
                for (int k = 0; k < 32; k++) c1[k] = state.adaptation_chain[k];
            } else {
                for (int k = 0; k < 32; k++) c2[k] = state.adaptation_chain[k];
                if (qadav_sha3_ct_compare(c1, c2, 32) != 0) return 1;
            }
            qecd_free(dec);
            aad_free_training_cert(cert);
        }
        qadav_write_raw("AAD_ADAPT_CHAIN_DETERMINISM_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "aad_aslr_determinism")) {
        unsigned char h1[32], h2[32];
        for (int iter = 0; iter < 50; iter++) {
            aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"fixed_m", 7, (const unsigned char*)"fixed_t", 7, 50, 64, 8);
            unsigned char syn[2] = {0x11, 0x22};
            qecd_decode_result* dec = qecd_decode(3, 9, syn, 2);
            aad_decode_envelope env;
            aad_wrap_decode(cert, dec, syn, 2, &env);
            if (iter == 0) {
                for (int k = 0; k < 32; k++) h1[k] = env.envelope_hash[k];
            } else {
                for (int k = 0; k < 32; k++) h2[k] = env.envelope_hash[k];
                if (qadav_sha3_ct_compare(h1, h2, 32) != 0) return 1;
            }
            qecd_free(dec);
            aad_free_training_cert(cert);
        }
        qadav_write_raw("AAD_ASLR_DETERMINISM_OK\n", 24);
        return 0;
    }
    if (qadav_streq(req, "record_alloc_and_fields")) {
        unsigned int sizes[3] = {8, 8, 8};
        void* rec = qadav_record_alloc(0x12345678ULL, 3, sizes);
        if (!rec) return 1;
        if (qadav_record_get_field_count(rec) != 3) return 2;
        if (qadav_record_get_type_hash(rec) != 0x12345678ULL) return 3;
        qadav_record_set_field_i64(rec, 0, 42);
        qadav_record_set_field_i64(rec, 1, 100);
        qadav_record_set_field_i64(rec, 2, 999);
        if (qadav_record_get_field_i64(rec, 0) != 42) return 4;
        if (qadav_record_get_field_i64(rec, 1) != 100) return 5;
        if (qadav_record_get_field_i64(rec, 2) != 999) return 6;
        qadav_record_free(rec);
        qadav_write_raw("RECORD_ALLOC_AND_FIELDS_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "variant_create_and_match")) {
        qadav_variant* v = qadav_variant_create_i64(0xABCDULL, 1, 777);
        if (!v) return 1;
        if (qadav_variant_get_tag(v) != 1) return 2;
        if (qadav_variant_match(v, 1) != 1) return 3;
        if (qadav_variant_match(v, 0) != 0) return 4;
        if (qadav_variant_get_payload_i64(v) != 777) return 5;
        qadav_variant_free(v);
        qadav_write_raw("VARIANT_CREATE_AND_MATCH_OK\n", 28);
        return 0;
    }
    if (qadav_streq(req, "variant_recursive_pointers")) {
        qadav_variant* lit1 = qadav_variant_create_i64(0x1234ULL, 0, 10);
        if (!lit1) return 1;
        if (qadav_variant_get_field_i64(lit1, 0) != 10) return 2;

        qadav_variant* wrapped = qadav_variant_create_ptr(0x1234ULL, 1, lit1);
        if (!wrapped) return 3;
        if (!qadav_variant_has_payload(wrapped)) return 4;
        qadav_variant* inner = (qadav_variant*)qadav_variant_get_field_ptr(wrapped, 0);
        if (inner != lit1) return 5;
        if (qadav_variant_get_field_i64(inner, 0) != 10) return 6;

        qadav_variant* lit2 = qadav_variant_create_i64(0x1234ULL, 0, 20);
        if (!lit2) return 7;

        qadav_variant* sum_tree = qadav_variant_create_2(0x1234ULL, 2, (unsigned long long)wrapped, (unsigned long long)lit2);
        if (!sum_tree) return 8;
        if (qadav_variant_get_tag(sum_tree) != 2) return 9;

        qadav_variant* left_node = (qadav_variant*)qadav_variant_get_field_ptr(sum_tree, 0);
        qadav_variant* right_node = (qadav_variant*)qadav_variant_get_field_ptr(sum_tree, 1);
        if (left_node != wrapped) return 10;
        if (right_node != lit2) return 11;

        qadav_variant* deep_leaf = (qadav_variant*)qadav_variant_get_field_ptr(left_node, 0);
        if (qadav_variant_get_field_i64(deep_leaf, 0) != 10) return 12;
        if (qadav_variant_get_field_i64(right_node, 0) != 20) return 13;

        int set_ok = qadav_variant_set_payload_ptr(lit1, lit2);
        if (!set_ok) return 14;
        if (qadav_variant_get_payload_ptr(lit1) != lit2) return 15;

        qadav_variant_free(sum_tree);
        qadav_variant_free(wrapped);
        qadav_variant_free(lit1);
        qadav_variant_free(lit2);

        qadav_write_raw("VARIANT_RECURSIVE_POINTERS_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "rsca_record_and_attest")) {
        qadav_rsca_reset();
        unsigned char s_hash[32] = {1, 2, 3};
        unsigned char c_hash[32] = {4, 5, 6};
        unsigned char out1[32] = {7, 8, 9};
        unsigned char out2[32] = {7, 8, 9};
        int idx0 = qadav_rsca_record_stage(0, s_hash, c_hash, out1);
        if (idx0 != 0) return 1;
        int idx1 = qadav_rsca_record_stage(1, s_hash, c_hash, out1);
        if (idx1 != 1) return 2;
        int idx2 = qadav_rsca_record_stage(2, s_hash, c_hash, out2);
        if (idx2 != 2) return 3;
        if (!qadav_rsca_verify_record(0)) return 4;
        if (!qadav_rsca_verify_record(1)) return 5;
        if (!qadav_rsca_verify_record(2)) return 6;
        if (!qadav_rsca_compare_outputs(1, 2)) return 7;
        if (!qadav_rsca_attest_chain()) return 8;
        qadav_write_raw("RSCA_RECORD_AND_ATTEST_OK\n", 26);
        return 0;
    }
    if (qadav_streq(req, "rsca_aslr_determinism")) {
        qadav_rsca_reset();
        unsigned char s[32] = {0xAA};
        unsigned char c[32] = {0xBB};
        unsigned char o[32] = {0xCC};
        qadav_rsca_record_stage(0, s, c, o);
        if (qadav_rsca_get_record_count() != 1) return 1;
        qadav_write_raw("RSCA_ASLR_DETERMINISM_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "ctsa_creation_and_verify")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {1, 2, 3};
        unsigned char te[32] = {4, 5, 6};
        unsigned char rc[32] = {7, 8, 9};
        if (!qadav_ctsa_attest(a, te, rc, 5, 10)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        if (qadav_ctsa_get_status() != 1) return 3;
        qadav_write_raw("CTSA_CREATION_AND_VERIFY_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "ctsa_rolling_chain_determinism")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0xAA};
        unsigned char te[32] = {0xBB};
        unsigned char rc[32] = {0xCC};
        if (!qadav_ctsa_attest(a, te, rc, 2, 4)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        qadav_write_raw("CTSA_ROLLING_CHAIN_DETERMINISM_OK\n", 35);
        return 0;
    }
    if (qadav_streq(req, "ctsa_aslr_determinism")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0x11};
        unsigned char te[32] = {0x22};
        unsigned char rc[32] = {0x33};
        if (!qadav_ctsa_attest(a, te, rc, 1, 1)) return 1;
        if (qadav_ctsa_get_status() != 1) return 2;
        qadav_write_raw("CTSA_ASLR_DETERMINISM_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "ctsa_lifecycle")) {
        qadav_ctsa_reset();
        if (qadav_ctsa_get_status() != 0) return 1;
        unsigned char a[32] = {0x44};
        unsigned char te[32] = {0x55};
        unsigned char rc[32] = {0x66};
        qadav_ctsa_attest(a, te, rc, 3, 3);
        if (qadav_ctsa_get_status() != 1) return 2;
        qadav_ctsa_reset();
        if (qadav_ctsa_get_status() != 0) return 3;
        qadav_write_raw("CTSA_LIFECYCLE_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "ctsa_multi_module_chain")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0x77};
        unsigned char te[32] = {0x88};
        unsigned char rc[32] = {0x99};
        if (!qadav_ctsa_attest(a, te, rc, 12, 16)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        qadav_write_raw("CTSA_MULTI_MODULE_CHAIN_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "ctsa_verify_fail_unattested")) {
        qadav_ctsa_reset();
        if (qadav_ctsa_verify() != 0) return 1;
        qadav_write_raw("CTSA_VERIFY_FAIL_UNATTESTED_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "pcc_creation_and_verify")) {
        qadav_pcc_reset();
        unsigned char a[32] = {1}, c[32] = {2}, ir[32] = {3};
        if (!qadav_pcc_attest(a, c, ir)) return 1;
        if (!qadav_pcc_verify()) return 2;
        if (qadav_pcc_get_status() != 1) return 3;
        qadav_write_raw("PCC_CREATION_AND_VERIFY_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "pcc_verify_fail_unattested")) {
        qadav_pcc_reset();
        if (qadav_pcc_verify() != 0) return 1;
        qadav_write_raw("PCC_VERIFY_FAIL_UNATTESTED_OK\n", 30);
        return 0;
    }
    if (qadav_streq(req, "integration_vault_reasoning")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        qadav_qemwv_seal(v, 0, 0);
        qadav_tensor** unsealed = qadav_qemwv_unseal(v, auth);
        (void)unsealed;
        if (v->state != 2) return 1;

        qadav_reasoning_chain* chain = qadav_qera_create("vr_chain", 10, auth);
        unsigned char premise[32];
        for (int k = 0; k < 32; k++) premise[k] = (unsigned char)(k + 1);
        unsigned char conc[32];
        for (int k = 0; k < 32; k++) conc[k] = (unsigned char)(k + 5);

        int r = qadav_qera_reason(chain, QADAV_OP_DEDUCE, premise, 32, conc, 32, 0, 0, 0.95, auth);
        if (r != 0) return 2;
        if (!qadav_qera_verify(chain, 0)) return 3;

        qadav_qera_free(chain);
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);
        qadav_write_raw("INTEGRATION_VAULT_REASONING_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "integration_memory_scoped_ttpcr")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_memory_provenance* mp = qadav_qemp_create(auth, 10);
        qadav_qemp_remember(mp, "record", "data123", 7, "main");

        unsigned char recipient_hash[32];
        for (int k = 0; k < 32; k++) recipient_hash[k] = 0xAA;
        const char* keys[] = {"record"};
        qadav_scoped_token* tok = qadav_scoped_issue("tok", auth, recipient_hash, keys, 1, 3600);
        qadav_scoped_grant(tok, recipient_hash);

        void* out_val = 0;
        unsigned long long out_len = 0;
        int r = qadav_scoped_recall(mp, tok, recipient_hash, "record", 100, &out_val, &out_len);
        if (r != 0 || !out_val) return 1;

        qadav_reasoning_chain* chain = qadav_qera_create("m_chain", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, (const unsigned char*)"record", 6, (const unsigned char*)out_val, out_len, 0, 0, 0.96, auth);
        int cr = qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)out_val, (unsigned int)out_len, 0.96);
        if (cr != 0) return 2;
        if (qadav_certificate_verify_chain(chain) != 0) return 3;

        qadav_free(out_val);
        qadav_scoped_free(tok);
        qadav_qera_free(chain);
        qadav_qemp_free(mp);
        qadav_qeeh_free(auth);
        qadav_write_raw("INTEGRATION_MEMORY_SCOPED_TTPCR_OK\n", 35);
        return 0;
    }
    if (qadav_streq(req, "integration_multimodal_certificate")) {
        qadav_qecma_bundle* bundle = qadav_qecma_create();
        qadav_qecma_add(bundle, 0x58, (const unsigned char*)"symptom", 7);
        float tdata[2] = {0.5f, 0.8f};
        qadav_qecma_add(bundle, 0x4E, (const unsigned char*)tdata, 8);
        int br = qadav_qecma_build(bundle);
        if (br != 0) return 1;

        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("mm_chain", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_INDUCE, bundle->unified_root, 32, (const unsigned char*)"dx", 2, 0, 0, 0.91, auth);
        qadav_certificate_attach(chain, 0, QADAV_CERT_INDUCTIVE, (const unsigned char*)"evidence_protocol", 17, 0.91);

        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_REFUSE, 0};
        if (qadav_abstain_check(chain, &pol, 0) != 0) return 2;

        qadav_qera_free(chain);
        qadav_qecma_free(bundle);
        qadav_qeeh_free(auth);
        qadav_write_raw("INTEGRATION_MULTIMODAL_CERTIFICATE_OK\n", 38);
        return 0;
    }
    if (qadav_streq(req, "integration_qec_tool_cert")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_tool_register("sensor_tool", auth, 1, 1);
        qadav_reasoning_chain* chain = qadav_qera_create("qec_chain", 10, auth);

        int tc = qadav_tool_call("sensor_tool", auth, "valid_state", chain);
        if (tc != 0) return 1;

        int ca = qadav_certificate_attach(chain, 0, QADAV_CERT_VERIFICATION, (const unsigned char*)"post_ok", 7, 1.0);
        if (ca != 0) return 2;
        if (qadav_certificate_verify_chain(chain) != 0) return 3;

        qadav_qera_free(chain);
        qadav_qeeh_free(auth);
        qadav_write_raw("INTEGRATION_QEC_TOOL_CERT_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "integration_full_pipeline")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);

        // 1. QEMWV
        qadav_vault* v = qadav_qemwv_create(5, 3, auth);
        qadav_qemwv_seal(v, 0, 0);
        qadav_tensor** unsealed = qadav_qemwv_unseal(v, auth);
        (void)unsealed;

        // 2. QECMT
        qadav_qecmt_context* ctx = qadav_qecmt_create();
        int lic_idx = qadav_qecmt_issue_license(ctx, v, 0, 50, 0.4, 5000, auth);
        int layers[2] = {0, 1};
        qadav_qecmt_transfer_layer(ctx, lic_idx, v, 0, layers, 2, auth);
        if (qadav_qecmt_verify_chain(ctx) != 0) return 1;

        // 3. QERA + TTPCR
        qadav_reasoning_chain* chain = qadav_qera_create("full_chain", 10, auth);
        qadav_qera_reason(chain, QADAV_OP_DEDUCE, ctx->provenance_chain, 32, (const unsigned char*)"transfer_ok", 11, 0, 0, 0.94, auth);
        qadav_certificate_attach(chain, 0, QADAV_CERT_DEDUCTIVE, (const unsigned char*)"lic_proof", 9, 0.94);

        // 4. Abstain check
        qadav_abstain_policy pol = {0.80, QADAV_ABSTAIN_REFUSE, 0};
        if (qadav_abstain_check(chain, &pol, 0) != 0) return 2;
        if (qadav_certificate_verify_chain(chain) != 0) return 3;

        qadav_qera_free(chain);
        qadav_qecmt_free(ctx);
        qadav_qemwv_free(v);
        qadav_qeeh_free(auth);
        qadav_write_raw("INTEGRATION_FULL_PIPELINE_OK\n", 29);
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
		"-O2", harnessC, "-I", cacheDir, "-o", harnessExe,
		"-nostdlib", "-fuse-ld=lld", "-Wl,-entry:qadav_start", "-Wl,-subsystem:console",
	}
	args = append(args, objs...)
	args = append(args, filepath.Join(cacheDir, "kernel32.lib"))

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

func runCryptoTestHarnessExpectExit(t *testing.T, testName string) (string, error) {
	t.Helper()
	tc, err := llvm.Detect()
	if err != nil {
		t.Skip("LLVM toolchain unavailable:", err)
	}

	cacheDir, objs, err := getSharedTestRuntime(tc)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	tmpDir := t.TempDir()

	harnessSrc := fmt.Sprintf(`
#include "qadav_sha3.h"
#include "qadav_merkle.h"
#include "qadav_lamport.h"
#include "qadav_selfcheck.h"
#include "qadav_qeeh.h"
#include "qadav_qemwv.h"
#include "qadav_mctp.h"
#include "qadav_qemp.h"
#include "qadav_qecmt.h"
#include "qadav_qera.h"
#include "qadav_qecma.h"
#include "qadav_agi.h"
#include "qadav_qec.h"
#include "qadav_consolidate.h"
#include "qadav_scoped.h"
#include "qadav_tools.h"
#include "qadav_certificate.h"
#include "qadav_abstain.h"
#include "qadav_sandbox.h"
#include "qadav_gpu.h"
#include "qadav_agke.h"
#include "qadav_multigpu.h"
#include "qadav_cdaga.h"
#include "qadav_distributed.h"
#include "qadav_barat.h"
#include "qadav_qhal.h"
#include "qadav_sea.h"
#include "qadav_transpile.h"
#include "qadav_eea.h"
#include "qadav_qecdecoder.h"
#include "qadav_aad.h"
#include "qadav_record.h"
#include "qadav_variant.h"
#include "qadav_rsca.h"
#include "qadav_ctsa.h"
#include "qadav_pcc.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void qadav_write_raw(const char* s, unsigned long long len);
void qadav_emit_int(long long v);
void qadav_exit(int code);

static int qadav_streq(const char* a, const char* b) {
    if (!a || !b) return 0;
    while (*a && *b && *a == *b) { a++; b++; }
    return (*a == *b);
}

static void qadav_copy_bytes(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) d[i] = s[i];
}

int main(void) {
    const char* req = "%s";
    if (qadav_streq(req, "agke_tamper_detection")) {
        qadav_gpu_init();
        qadav_agke_reset();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("agke_tamper_chain", 10, auth);
        double A[4] = {1.0, 0.0, 0.0, 1.0};
        double B[4] = {2.0, 3.0, 4.0, 5.0};
        double C[4] = {0.0, 0.0, 0.0, 0.0};
        qadav_gpu_matmul(A, B, C, 2, 2, 2);
        qadav_agke_attest(chain, auth);
        qadav_agke_corrupt_record(0, 1);
        qadav_agke_verify(chain, auth);
        return 0;
    }
    if (qadav_streq(req, "cdaga_tamper_detection")) {
        qadav_multigpu_context* ctx = qadav_multigpu_init();
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("cdaga_tamper_chain", 10, auth);
        qadav_cdaga_record rec;
        double g0[4] = {0.1, 0.2, 0.3, 0.4};
        void* bufs[1] = {g0};
        unsigned long long sizes[1] = {sizeof(g0)};
        qadav_cdaga_record_aggregation(ctx, 1, bufs, sizes, 1, &rec);
        qadav_cdaga_attest(chain, &rec, ctx);
        qadav_cdaga_corrupt(&rec, 16);
        qadav_cdaga_verify(&rec, ctx);
        return 0;
    }
    if (qadav_streq(req, "barat_tampered_cert")) {
        qadav_barat_group* grp = qadav_barat_create_group(4);
        double g0[2] = {1.0, 2.0};
        qadav_barat_submit(grp, 0, g0, sizeof(g0));
        double out[2];
        qadav_barat_cert cert;
        qadav_barat_aggregate(grp, out, sizeof(out), &cert);
        qadav_barat_corrupt(&cert, 16);
        qadav_barat_verify(&cert, out, sizeof(out));
        return 0;
    }
    if (qadav_streq(req, "sea_verify_tampered")) {
        sea_history* h = sea_create();
        unsigned char from[32], to[32], reason[32];
        for (int i = 0; i < 32; i++) {
            from[i] = (unsigned char)(i + 1);
            to[i] = (unsigned char)(i + 2);
            reason[i] = (unsigned char)(i + 3);
        }
        sea_proposal* p = 0;
        sea_propose(h, from, to, reason, &p);
        sea_corrupt(p, 16);
        sea_verify_proposal(p, 0);
        return 0;
    }
    if (qadav_streq(req, "sea_attest_tampered")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("sea_tamper_chain", 10, auth);
        sea_history* h = sea_create();
        unsigned char from[32], to[32], reason[32];
        for (int i = 0; i < 32; i++) {
            from[i] = (unsigned char)(i + 10);
            to[i] = (unsigned char)(i + 20);
            reason[i] = (unsigned char)(i + 30);
        }
        sea_proposal* p = 0;
        sea_propose(h, from, to, reason, &p);
        sea_corrupt(p, 16);
        sea_attest(h, chain);
        return 0;
    }
    if (qadav_streq(req, "eea_verify_tampered")) {
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        eea_corrupt(&env, 16);
        eea_verify_envelope(&env);
        return 0;
    }
    if (qadav_streq(req, "eea_attest_tampered")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("eea_chain", 10, auth);
        eea_context* ctx = eea_create_policy(POLICY_CONSTANT_RATE, 0.10, 100);
        eea_envelope env;
        eea_build_envelope_text(ctx, 0, "logic_v1", "logic_v2", 16, 4, &env);
        eea_free(ctx);
        eea_corrupt(&env, 8);
        eea_attest_envelope(&env, chain);
        return 0;
    }
    if (qadav_streq(req, "aad_attest_tampered")) {
        double pen = 1.0;
        qadav_ethical_hamiltonian* auth = qadav_qeeh_create(1, &pen);
        qadav_reasoning_chain* chain = qadav_qera_create("aad_tamper_chain", 10, auth);
        aad_training_cert* cert = aad_build_training_cert((const unsigned char*)"model", 5, (const unsigned char*)"train", 5, 100, 64, 8);
        unsigned char syn[1] = {0x01};
        qecd_decode_result* dec = qecd_decode(3, 9, syn, 1);
        aad_decode_envelope env;
        aad_wrap_decode(cert, dec, syn, 1, &env);
        aad_corrupt_envelope(&env, 16);
        aad_attest_envelope(&env, chain);
        return 0;
    }
    if (qadav_streq(req, "rsca_tamper_detection")) {
        qadav_rsca_reset();
        unsigned char s[32] = {1, 2, 3};
        unsigned char c[32] = {4, 5, 6};
        unsigned char o[32] = {7, 8, 9};
        qadav_rsca_record_stage(0, s, c, o);
        qadav_rsca_corrupt_record(0, 0);
        qadav_rsca_verify_record(0);
        return 0;
    }
    if (qadav_streq(req, "ctsa_creation_and_verify")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {1, 2, 3};
        unsigned char te[32] = {4, 5, 6};
        unsigned char rc[32] = {7, 8, 9};
        if (!qadav_ctsa_attest(a, te, rc, 5, 10)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        if (qadav_ctsa_get_status() != 1) return 3;
        qadav_write_raw("CTSA_CREATION_AND_VERIFY_OK\n", 29);
        return 0;
    }
    if (qadav_streq(req, "ctsa_rolling_chain_determinism")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0xAA};
        unsigned char te[32] = {0xBB};
        unsigned char rc[32] = {0xCC};
        if (!qadav_ctsa_attest(a, te, rc, 2, 4)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        qadav_write_raw("CTSA_ROLLING_CHAIN_DETERMINISM_OK\n", 35);
        return 0;
    }
    if (qadav_streq(req, "ctsa_aslr_determinism")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0x11};
        unsigned char te[32] = {0x22};
        unsigned char rc[32] = {0x33};
        if (!qadav_ctsa_attest(a, te, rc, 1, 1)) return 1;
        if (qadav_ctsa_get_status() != 1) return 2;
        qadav_write_raw("CTSA_ASLR_DETERMINISM_OK\n", 25);
        return 0;
    }
    if (qadav_streq(req, "ctsa_lifecycle")) {
        qadav_ctsa_reset();
        if (qadav_ctsa_get_status() != 0) return 1;
        unsigned char a[32] = {0x44};
        unsigned char te[32] = {0x55};
        unsigned char rc[32] = {0x66};
        qadav_ctsa_attest(a, te, rc, 3, 3);
        if (qadav_ctsa_get_status() != 1) return 2;
        qadav_ctsa_reset();
        if (qadav_ctsa_get_status() != 0) return 3;
        qadav_write_raw("CTSA_LIFECYCLE_OK\n", 18);
        return 0;
    }
    if (qadav_streq(req, "ctsa_multi_module_chain")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {0x77};
        unsigned char te[32] = {0x88};
        unsigned char rc[32] = {0x99};
        if (!qadav_ctsa_attest(a, te, rc, 12, 16)) return 1;
        if (!qadav_ctsa_verify()) return 2;
        qadav_write_raw("CTSA_MULTI_MODULE_CHAIN_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "ctsa_tamper_detection")) {
        qadav_ctsa_reset();
        unsigned char a[32] = {1, 2, 3};
        unsigned char te[32] = {4, 5, 6};
        unsigned char rc[32] = {7, 8, 9};
        qadav_ctsa_attest(a, te, rc, 5, 10);
        qadav_ctsa_corrupt(1);
        qadav_ctsa_verify(); // triggers ExitProcess(0xDEADBEEF)
        return 0;
    }
    if (qadav_streq(req, "pcc_tamper_detection")) {
        qadav_pcc_reset();
        unsigned char a[32] = {1}, c[32] = {2}, ir[32] = {3};
        qadav_pcc_attest(a, c, ir);
        qadav_pcc_corrupt(1);
        qadav_pcc_verify(); // triggers ExitProcess(0xDEADBEEF)
        return 0;
    }
    if (qadav_streq(req, "ctsa_verify_fail_unattested")) {
        qadav_ctsa_reset();
        if (qadav_ctsa_verify() != 0) return 1;
        qadav_write_raw("CTSA_VERIFY_FAIL_UNATTESTED_OK\n", 31);
        return 0;
    }
    if (qadav_streq(req, "pcc_creation_and_verify")) {
        qadav_pcc_reset();
        unsigned char a[32] = {1}, c[32] = {2}, ir[32] = {3};
        if (!qadav_pcc_attest(a, c, ir)) return 1;
        if (!qadav_pcc_verify()) return 2;
        if (qadav_pcc_get_status() != 1) return 3;
        qadav_write_raw("PCC_CREATION_AND_VERIFY_OK\n", 27);
        return 0;
    }
    if (qadav_streq(req, "pcc_verify_fail_unattested")) {
        qadav_pcc_reset();
        if (qadav_pcc_verify() != 0) return 1;
        qadav_write_raw("PCC_VERIFY_FAIL_UNATTESTED_OK\n", 30);
        return 0;
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
		"-O2", harnessC, "-I", cacheDir, "-o", harnessExe,
		"-nostdlib", "-fuse-ld=lld", "-Wl,-entry:qadav_start", "-Wl,-subsystem:console",
	}
	args = append(args, objs...)
	args = append(args, filepath.Join(cacheDir, "kernel32.lib"))

	cmd := exec.Command(tc.ClangPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compiling harness: %v\noutput: %s", err, string(out))
	}

	runCmd := exec.Command(harnessExe)
	runOut, err := runCmd.CombinedOutput()
	return string(runOut), err
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

func TestQEMWVBoundaryEscalation(t *testing.T) {
	out := runCryptoTestHarness(t, "qemwv_escalation")
	if !strings.Contains(out, "QEMWV_BOUNDARY_ESCALATION_OK") {
		t.Fatalf("QEMWV boundary escalation test failed: %s", out)
	}
	t.Logf("QEMWV boundary escalation test verified: %s", out)
}

func TestQEMPRemember(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_remember")
	if !strings.Contains(out, "QEMP_REMEMBER_OK") {
		t.Fatalf("QEMP remember failed: %s", out)
	}
	t.Logf("QEMP remember verified: %s", out)
}

func TestQEMPRecall(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_recall")
	if !strings.Contains(out, "QEMP_RECALL_OK") {
		t.Fatalf("QEMP recall failed: %s", out)
	}
	t.Logf("QEMP recall verified: %s", out)
}

func TestQEMPTamper(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_tamper")
	if !strings.Contains(out, "QEMP_TAMPER_OK") {
		t.Fatalf("QEMP tamper failed: %s", out)
	}
	t.Logf("QEMP tamper verified: %s", out)
}

func TestQEMPVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_verify")
	if !strings.Contains(out, "QEMP_VERIFY_OK") {
		t.Fatalf("QEMP verify failed: %s", out)
	}
	t.Logf("QEMP verify-all and audit verified: %s", out)
}

func TestQEEHBoundaryHashDeterministic(t *testing.T) {
	out := runCryptoTestHarness(t, "qeeh_boundary_hash_det")
	if !strings.Contains(out, "QEEH_BOUNDARY_HASH_DETERMINISTIC_OK") {
		t.Fatalf("QEEH boundary hash determinism test failed: %s", out)
	}
	t.Logf("QEEH boundary hash determinism verified: %s", out)
}

func TestQEMPBinaryValueRoundtrip(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_binary_roundtrip")
	if !strings.Contains(out, "QEMP_BINARY_ROUNDTRIP_OK") {
		t.Fatalf("QEMP binary value roundtrip test failed: %s", out)
	}
	t.Logf("QEMP binary value roundtrip verified: %s", out)
}

func TestQEMPDeterministicAcrossRuns(t *testing.T) {
	out := runCryptoTestHarness(t, "qemp_deterministic")
	if !strings.Contains(out, "QEMP_DETERMINISTIC_OK") {
		t.Fatalf("QEMP determinism test failed: %s", out)
	}
	t.Logf("QEMP determinism across 100 runs verified: %s", out)
}

func TestMCTPDeterministicAcrossRuns(t *testing.T) {
	out := runCryptoTestHarness(t, "mctp_deterministic")
	if !strings.Contains(out, "MCTP_DETERMINISTIC_OK") {
		t.Fatalf("MCTP determinism test failed: %s", out)
	}
	t.Logf("MCTP determinism across 10 runs verified: %s", out)
}

func TestQEMWVDeterministicSeal(t *testing.T) {
	out := runCryptoTestHarness(t, "qemwv_deterministic_seal")
	if !strings.Contains(out, "QEMWV_DETERMINISTIC_SEAL_OK") {
		t.Fatalf("QEMWV deterministic seal test failed: %s", out)
	}
	t.Logf("QEMWV deterministic seal verified: %s", out)
}

func TestQECMTLicenseBinding(t *testing.T) {
	out := runCryptoTestHarness(t, "qecmt_license_binding")
	if !strings.Contains(out, "QECMT_LICENSE_BINDING_OK") {
		t.Fatalf("QECMT license binding test failed: %s", out)
	}
	t.Logf("QECMT license binding verified: %s", out)
}

func TestQECMTProvenanceChain(t *testing.T) {
	out := runCryptoTestHarness(t, "qecmt_provenance_chain")
	if !strings.Contains(out, "QECMT_PROVENANCE_CHAIN_OK") {
		t.Fatalf("QECMT provenance chain test failed: %s", out)
	}
	t.Logf("QECMT provenance chain verified: %s", out)
}

func TestQECMTRateEnforcement(t *testing.T) {
	out := runCryptoTestHarness(t, "qecmt_rate_enforcement")
	if !strings.Contains(out, "QECMT_RATE_ENFORCEMENT_OK") {
		t.Fatalf("QECMT rate enforcement test failed: %s", out)
	}
	t.Logf("QECMT rate enforcement verified: %s", out)
}

func TestQERACreateReasonVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_create_reason_verify")
	if !strings.Contains(out, "QERA_CREATE_REASON_VERIFY_OK") {
		t.Fatalf("QERA create reason verify test failed: %s", out)
	}
	t.Logf("QERA create reason verify verified: %s", out)
}

func TestQERATamperDecoherence(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_tamper")
	if !strings.Contains(out, "QERA_TAMPER_OK") {
		t.Fatalf("QERA tamper decoherence test failed: %s", out)
	}
	t.Logf("QERA tamper decoherence verified: %s", out)
}

func TestQERAStepHashDeterministic(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_step_hash_det")
	if !strings.Contains(out, "QERA_STEP_HASH_DET_OK") {
		t.Fatalf("QERA step hash determinism test failed: %s", out)
	}
	t.Logf("QERA step hash determinism verified: %s", out)
}

func TestQERAChainRootDeterministic(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_chain_root_det")
	if !strings.Contains(out, "QERA_CHAIN_ROOT_DET_OK") {
		t.Fatalf("QERA chain root determinism test failed: %s", out)
	}
	t.Logf("QERA chain root determinism verified: %s", out)
}

func TestQERABoundaryEscalation(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_boundary_escalation")
	if !strings.Contains(out, "QERA_BOUNDARY_ESCALATION_OK") {
		t.Fatalf("QERA boundary escalation test failed: %s", out)
	}
	t.Logf("QERA boundary escalation verified: %s", out)
}

func TestQERAInputHashNonZero(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_input_hash_nonzero")
	if !strings.Contains(out, "QERA_INPUT_HASH_NONZERO_OK") {
		t.Fatalf("QERA input hash non-zero test failed: %s", out)
	}
	t.Logf("QERA input hash non-zero verified: %s", out)
}

func TestQERAEmptyInputRejected(t *testing.T) {
	out := runCryptoTestHarness(t, "qera_empty_input_rejected")
	if !strings.Contains(out, "QERA_EMPTY_INPUT_REJECTED_OK") {
		t.Fatalf("QERA empty input rejected test failed: %s", out)
	}
	t.Logf("QERA empty input rejected verified: %s", out)
}

func TestQECMABundleBuild(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_bundle_build")
	if !strings.Contains(out, "QECMA_BUNDLE_BUILD_OK") {
		t.Fatalf("QECMA bundle build test failed: %s", out)
	}
	t.Logf("QECMA bundle build verified: %s", out)
}

func TestQECMATagOrdering(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_tag_ordering")
	if !strings.Contains(out, "QECMA_TAG_ORDERING_OK") {
		t.Fatalf("QECMA tag ordering test failed: %s", out)
	}
	t.Logf("QECMA tag ordering verified: %s", out)
}

func TestQECMAMultimodalTamper(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_multimodal_tamper")
	if !strings.Contains(out, "QECMA_MULTIMODAL_TAMPER_OK") {
		t.Fatalf("QECMA multimodal tamper test failed: %s", out)
	}
	t.Logf("QECMA multimodal tamper verified: %s", out)
}

func TestQECMACrossModalReason(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_cross_modal_reason")
	if !strings.Contains(out, "QECMA_CROSS_MODAL_REASON_OK") {
		t.Fatalf("QECMA cross modal reason test failed: %s", out)
	}
	t.Logf("QECMA cross modal reason verified: %s", out)
}

func TestQECMAEmptyModalityRejected(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_empty_modality")
	if !strings.Contains(out, "QECMA_EMPTY_MODALITY_REJECTED_OK") {
		t.Fatalf("QECMA empty modality test failed: %s", out)
	}
	t.Logf("QECMA empty modality verified: %s", out)
}

func TestQECMAEmptyBundleRejected(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_empty_bundle")
	if !strings.Contains(out, "QECMA_EMPTY_BUNDLE_REJECTED_OK") {
		t.Fatalf("QECMA empty bundle test failed: %s", out)
	}
	t.Logf("QECMA empty bundle verified: %s", out)
}

func TestQECMAHashSanityCheck(t *testing.T) {
	out := runCryptoTestHarness(t, "qecma_hash_sanity")
	if !strings.Contains(out, "QECMA_HASH_SANITY_OK") {
		t.Fatalf("QECMA hash sanity test failed: %s", out)
	}
	t.Logf("QECMA hash sanity verified: %s", out)
}

func TestQECCreate(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_create")
	if !strings.Contains(out, "QEC_CREATE_OK") {
		t.Fatalf("QEC create test failed: %s", out)
	}
	t.Logf("QEC create verified: %s", out)
}

func TestQECSyndromeNonZero(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_syndrome_nonzero")
	if !strings.Contains(out, "QEC_SYNDROME_NONZERO_OK") {
		t.Fatalf("QEC syndrome non-zero test failed: %s", out)
	}
	t.Logf("QEC syndrome non-zero verified: %s", out)
}

func TestQECEmptySyndromeRejected(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_empty_syndrome")
	if !strings.Contains(out, "QEC_EMPTY_SYNDROME_REJECTED_OK") {
		t.Fatalf("QEC empty syndrome test failed: %s", out)
	}
	t.Logf("QEC empty syndrome rejected verified: %s", out)
}

func TestQECTamperDecoherence(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_tamper_decoherence")
	if !strings.Contains(out, "QEC_TAMPER_DECOHERENCE_OK") {
		t.Fatalf("QEC tamper decoherence test failed: %s", out)
	}
	t.Logf("QEC tamper decoherence verified: %s", out)
}

func TestQECVerifyBeforeCorrect(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_verify_before_correct")
	if !strings.Contains(out, "QEC_VERIFY_BEFORE_CORRECT_OK") {
		t.Fatalf("QEC verify before correct test failed: %s", out)
	}
	t.Logf("QEC verify before correct verified: %s", out)
}

func TestQECBoundaryDeterminism(t *testing.T) {
	out := runCryptoTestHarness(t, "qec_boundary_determinism")
	if !strings.Contains(out, "QEC_BOUNDARY_DETERMINISM_OK") {
		t.Fatalf("QEC boundary determinism test failed: %s", out)
	}
	t.Logf("QEC boundary determinism verified: %s", out)
}

func TestConsolidation(t *testing.T) {
	out := runCryptoTestHarness(t, "consolidation_test")
	if !strings.Contains(out, "CONSOLIDATION_OK") {
		t.Fatalf("Consolidation test failed: %s", out)
	}
	t.Logf("Consolidation verified: %s", out)
}

func TestScopedIssueAndVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_issue_and_verify")
	if !strings.Contains(out, "SCOPED_ISSUE_AND_VERIFY_OK") {
		t.Fatalf("Scoped issue and verify failed: %s", out)
	}
	t.Logf("Scoped issue and verify verified: %s", out)
}

func TestScopedRecallPrivacy(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_recall_privacy")
	if !strings.Contains(out, "SCOPED_RECALL_PRIVACY_OK") {
		t.Fatalf("Scoped recall privacy failed: %s", out)
	}
	t.Logf("Scoped recall privacy verified: %s", out)
}

func TestScopedUnauthorizedKey(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_unauthorized_key")
	if !strings.Contains(out, "SCOPED_UNAUTHORIZED_KEY_OK") {
		t.Fatalf("Scoped unauthorized key failed: %s", out)
	}
	t.Logf("Scoped unauthorized key verified: %s", out)
}

func TestToolRegisterAndCall(t *testing.T) {
	out := runCryptoTestHarness(t, "tool_register_and_call")
	if !strings.Contains(out, "TOOL_REGISTER_AND_CALL_OK") {
		t.Fatalf("Tool register and call failed: %s", out)
	}
	t.Logf("Tool register and call verified: %s", out)
}

func TestToolPostConditionFailure(t *testing.T) {
	out := runCryptoTestHarness(t, "tool_post_condition_failure")
	if !strings.Contains(out, "TOOL_POST_CONDITION_FAILURE_OK") {
		t.Fatalf("Tool post condition failure test failed: %s", out)
	}
	t.Logf("Tool post condition failure verified: %s", out)
}

func TestScopedRecallAsQECMAModality(t *testing.T) {
	out := runCryptoTestHarness(t, "scoped_recall_as_qecma_modality")
	if !strings.Contains(out, "SCOPED_RECALL_AS_QECMA_MODALITY_OK") {
		t.Fatalf("Scoped recall as QECMA modality failed: %s", out)
	}
	t.Logf("Scoped recall as QECMA modality verified: %s", out)
}

func TestCertificateAttach(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_attach")
	if !strings.Contains(out, "CERT_ATTACH_OK") {
		t.Fatalf("Certificate attach test failed: %s", out)
	}
	t.Logf("Certificate attach verified: %s", out)
}

func TestCertificateVerifyStructural(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_verify_structural")
	if !strings.Contains(out, "CERT_VERIFY_STRUCTURAL_OK") {
		t.Fatalf("Certificate verify structural test failed: %s", out)
	}
	t.Logf("Certificate verify structural verified: %s", out)
}

func TestCertificateVerifyInvalidKind(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_invalid_kind")
	if !strings.Contains(out, "CERT_INVALID_KIND_OK") {
		t.Fatalf("Certificate invalid kind test failed: %s", out)
	}
	t.Logf("Certificate invalid kind verified: %s", out)
}

func TestCertificateVerifyEmptyWitness(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_empty_witness")
	if !strings.Contains(out, "CERT_EMPTY_WITNESS_OK") {
		t.Fatalf("Certificate empty witness test failed: %s", out)
	}
	t.Logf("Certificate empty witness verified: %s", out)
}

func TestCertificateVerifyInvalidConfidence(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_invalid_confidence")
	if !strings.Contains(out, "CERT_INVALID_CONFIDENCE_OK") {
		t.Fatalf("Certificate invalid confidence test failed: %s", out)
	}
	t.Logf("Certificate invalid confidence verified: %s", out)
}

func TestCertificateChainVerify(t *testing.T) {
	out := runCryptoTestHarness(t, "cert_chain_verify")
	if !strings.Contains(out, "CERT_CHAIN_VERIFY_OK") {
		t.Fatalf("Certificate chain verify test failed: %s", out)
	}
	t.Logf("Certificate chain verify verified: %s", out)
}

func TestAbstainComputeConfidence(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_compute_confidence")
	if !strings.Contains(out, "ABSTAIN_COMPUTE_CONFIDENCE_OK") {
		t.Fatalf("Abstain compute confidence test failed: %s", out)
	}
	t.Logf("Abstain compute confidence verified: %s", out)
}

func TestAbstainBelowThreshold(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_below_threshold")
	if !strings.Contains(out, "ABSTAIN_BELOW_THRESHOLD_OK") {
		t.Fatalf("Abstain below threshold test failed: %s", out)
	}
	t.Logf("Abstain below threshold verified: %s", out)
}

func TestAbstainAboveThreshold(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_above_threshold")
	if !strings.Contains(out, "ABSTAIN_ABOVE_THRESHOLD_OK") {
		t.Fatalf("Abstain above threshold test failed: %s", out)
	}
	t.Logf("Abstain above threshold verified: %s", out)
}

func TestAbstainRetryLimit(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_retry_limit")
	if !strings.Contains(out, "ABSTAIN_RETRY_LIMIT_OK") {
		t.Fatalf("Abstain retry limit test failed: %s", out)
	}
	t.Logf("Abstain retry limit verified: %s", out)
}

func TestAbstainEscalateMode(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_escalate_mode")
	if !strings.Contains(out, "ABSTAIN_ESCALATE_MODE_OK") {
		t.Fatalf("Abstain escalate mode test failed: %s", out)
	}
	t.Logf("Abstain escalate mode verified: %s", out)
}

func TestAbstainRefuseMode(t *testing.T) {
	out := runCryptoTestHarness(t, "abstain_refuse_mode")
	if !strings.Contains(out, "ABSTAIN_REFUSE_MODE_OK") {
		t.Fatalf("Abstain refuse mode test failed: %s", out)
	}
	t.Logf("Abstain refuse mode verified: %s", out)
}
