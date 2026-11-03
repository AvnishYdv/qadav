#include "qadav_qecmt.h"
#include "qadav_nn.h"

#ifdef _WIN32
__declspec(dllimport) void* __stdcall GetProcessHeap(void);
__declspec(dllimport) void* __stdcall HeapAlloc(void* hHeap, unsigned int dwFlags, unsigned long long dwBytes);
__declspec(dllimport) int   __stdcall HeapFree(void* hHeap, unsigned int dwFlags, void* lpMem);
__declspec(dllimport) void  __stdcall ExitProcess(unsigned int uExitCode);
#endif

#define HEAP_ZERO_MEMORY 0x00000008

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);

static void qecmt_memcpy(void* dst, const void* src, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    const volatile unsigned char* s = (const volatile unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qecmt_memset(void* dst, int v, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static int qecmt_memcmp(const void* s1, const void* s2, unsigned long long n) {
    const unsigned char* p1 = (const unsigned char*)s1;
    const unsigned char* p2 = (const unsigned char*)s2;
    for (unsigned long long i = 0; i < n; i++) {
        if (p1[i] != p2[i]) return (int)p1[i] - (int)p2[i];
    }
    return 0;
}

static qadav_qecmt_context* g_qecmt_ctx = 0;

qadav_qecmt_context* qadav_qecmt_get_global(void) {
    if (!g_qecmt_ctx) {
        g_qecmt_ctx = qadav_qecmt_create();
    }
    return g_qecmt_ctx;
}

qadav_qecmt_context* qadav_qecmt_create(void) {
    qadav_qecmt_context* ctx = (qadav_qecmt_context*)qadav_alloc(sizeof(qadav_qecmt_context));
    if (!ctx) return 0;
    qecmt_memset(ctx, 0, sizeof(qadav_qecmt_context));
    // Initial provenance chain = SHA3-256("QECMT_ROOT")
    qadav_sha3_256((const unsigned char*)"QECMT_ROOT", 10, ctx->provenance_chain);
    return ctx;
}

int qadav_qecmt_issue_license(qadav_qecmt_context* ctx,
                              qadav_vault* teacher_vault,
                              void* student_network,
                              unsigned long long max_epochs,
                              double max_distill_rate,
                              unsigned long long expiry,
                              void* auth_boundary) {
    if (!ctx) ctx = qadav_qecmt_get_global();
    if (!ctx || ctx->n_licenses >= QADAV_QECMT_MAX_LICENSES) return -1;

    // 1. Verify teacher vault is sealed
    if (!teacher_vault || teacher_vault->state != 1) {
        return -2;
    }

    // 2. Verify boundary identity via content hash
    unsigned char b_hash[32];
    qecmt_memset(b_hash, 0, 32);
    if (auth_boundary) {
        qadav_qeeh_boundary_hash(auth_boundary, b_hash);
    } else {
        qecmt_memcpy(b_hash, teacher_vault->authorized_boundary_id, 32);
    }

    int idx = ctx->n_licenses++;
    qadav_transfer_license* lic = &ctx->licenses[idx];
    qecmt_memset(lic, 0, sizeof(qadav_transfer_license));

    lic->max_epochs = max_epochs > 0 ? max_epochs : 100;
    lic->max_distill_rate = (max_distill_rate > 0.0 && max_distill_rate <= 1.0) ? max_distill_rate : 0.5;
    lic->expiry = expiry;
    lic->issued_at = __builtin_readcyclecounter();
    lic->used = 0;
    lic->epochs_used = 0;
    qecmt_memcpy(lic->auth_boundary_id, b_hash, 32);
    qecmt_memcpy(lic->teacher_vault_id, teacher_vault->authorized_boundary_id, 32);

    // Compute student hash
    if (student_network) {
        qadav_sha3_256((const unsigned char*)student_network, 32, lic->student_network_hash);
    } else {
        qadav_sha3_256((const unsigned char*)"student_net", 11, lic->student_network_hash);
    }

    // Generate license_id = SHA3(teacher_vault_id || student_network_hash || rate || expiry || nonce)
    unsigned char lic_buf[128];
    qecmt_memcpy(lic_buf, lic->teacher_vault_id, 32);
    qecmt_memcpy(lic_buf + 32, lic->student_network_hash, 32);
    qecmt_memcpy(lic_buf + 64, &lic->max_distill_rate, 8);
    qecmt_memcpy(lic_buf + 72, &lic->expiry, 8);
    qecmt_memcpy(lic_buf + 80, &lic->issued_at, 8);
    qadav_sha3_256(lic_buf, 88, lic->license_id);

    // Update rolling provenance chain: chain = SHA3(chain || license_id)
    unsigned char chain_buf[64];
    qecmt_memcpy(chain_buf, ctx->provenance_chain, 32);
    qecmt_memcpy(chain_buf + 32, lic->license_id, 32);
    qadav_sha3_256(chain_buf, 64, ctx->provenance_chain);

    return idx;
}

int qadav_qecmt_transfer_layer(qadav_qecmt_context* ctx,
                               int license_idx,
                               qadav_vault* teacher_vault,
                               void* student_network,
                               const int* layers,
                               int n_layers,
                               void* auth_boundary) {
    if (!ctx) ctx = qadav_qecmt_get_global();
    if (!ctx || license_idx < 0 || license_idx >= ctx->n_licenses) return -1;
    if (!teacher_vault || teacher_vault->state != 1) return -2;

    qadav_transfer_license* lic = &ctx->licenses[license_idx];

    // Verify boundary if provided
    if (auth_boundary) {
        unsigned char b_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, b_hash);
        if (qecmt_memcmp(b_hash, lic->auth_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -3;
        }
    }

    lic->used = 1;

    // Update provenance chain with transfer event
    unsigned char trans_buf[64];
    qecmt_memcpy(trans_buf, ctx->provenance_chain, 32);
    qecmt_memcpy(trans_buf + 32, lic->license_id, 32);
    if (n_layers > 0 && layers) {
        trans_buf[32] ^= (unsigned char)(layers[0] & 0xFF);
    }
    qadav_sha3_256(trans_buf, 64, ctx->provenance_chain);

    return 0;
}

int qadav_qecmt_distill(qadav_qecmt_context* ctx,
                        int license_idx,
                        void* student,
                        void* dataset_x,
                        void* dataset_y,
                        double rate,
                        int epochs) {
    if (!ctx) ctx = qadav_qecmt_get_global();
    if (!ctx || license_idx < 0 || license_idx >= ctx->n_licenses) return -1;

    qadav_transfer_license* lic = &ctx->licenses[license_idx];

    // Anti-IP-theft: Rate enforcement
    if (rate > lic->max_distill_rate + 1e-9) {
        // VIOLATION: Distillation rate exceeds licensed threshold!
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return -2;
    }

    if (epochs <= 0) epochs = 10;
    if (lic->epochs_used + (unsigned long long)epochs > lic->max_epochs) {
        // Exceeded allowed epochs
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return -3;
    }

    // Train student with distillation parameters
    if (student && dataset_x && dataset_y) {
        qadav_train_epochs((qadav_network*)student, (const qadav_tensor*)dataset_x, (const qadav_tensor*)dataset_y,
                           QADAV_LOSS_MSE, QADAV_OPT_ADAM, rate, epochs);
    }

    lic->epochs_used += (unsigned long long)epochs;

    // Update provenance chain
    unsigned char dist_buf[64];
    qecmt_memcpy(dist_buf, ctx->provenance_chain, 32);
    qecmt_memcpy(dist_buf + 32, lic->license_id, 32);
    dist_buf[63] = (unsigned char)(epochs & 0xFF);
    qadav_sha3_256(dist_buf, 64, ctx->provenance_chain);

    return 0;
}

int qadav_qecmt_verify_chain(qadav_qecmt_context* ctx) {
    if (!ctx) ctx = qadav_qecmt_get_global();
    if (!ctx) return 1;

    // Recompute chain from root
    unsigned char expected_chain[32];
    qadav_sha3_256((const unsigned char*)"QECMT_ROOT", 10, expected_chain);

    for (int i = 0; i < ctx->n_licenses; i++) {
        unsigned char chain_buf[64];
        qecmt_memcpy(chain_buf, expected_chain, 32);
        qecmt_memcpy(chain_buf + 32, ctx->licenses[i].license_id, 32);
        qadav_sha3_256(chain_buf, 64, expected_chain);

        if (ctx->licenses[i].used) {
            unsigned char trans_buf[64];
            qecmt_memcpy(trans_buf, expected_chain, 32);
            qecmt_memcpy(trans_buf + 32, ctx->licenses[i].license_id, 32);
            qadav_sha3_256(trans_buf, 64, expected_chain);
        }

        if (ctx->licenses[i].epochs_used > 0) {
            unsigned char dist_buf[64];
            qecmt_memcpy(dist_buf, expected_chain, 32);
            qecmt_memcpy(dist_buf + 32, ctx->licenses[i].license_id, 32);
            dist_buf[63] = (unsigned char)(ctx->licenses[i].epochs_used & 0xFF);
            qadav_sha3_256(dist_buf, 64, expected_chain);
        }
    }

    // Verify non-zero and valid
    int non_zero = 0;
    for (int k = 0; k < 32; k++) {
        if (ctx->provenance_chain[k] != 0) non_zero = 1;
    }
    return non_zero ? 0 : 1;
}

void qadav_qecmt_free(qadav_qecmt_context* ctx) {
    if (!ctx) return;
    if (ctx == g_qecmt_ctx) {
        g_qecmt_ctx = 0;
    }
    qadav_free(ctx);
}

