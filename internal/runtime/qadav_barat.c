#include "qadav_barat.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void  qadav_qeeh_trigger_decoherence(void* qstate);

#ifdef _WIN32
void __stdcall ExitProcess(unsigned int uExitCode);
#endif

#define CERT_PREIMAGE_LEN 896

static void barat_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void barat_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static int barat_uuid_cmp(const unsigned char a[16], const unsigned char b[16]) {
    for (int i = 0; i < 16; i++) {
        if (a[i] < b[i]) return -1;
        if (a[i] > b[i]) return 1;
    }
    return 0;
}

qadav_barat_group* qadav_barat_create_group(int group_id) {
    qadav_barat_group* grp = (qadav_barat_group*)qadav_alloc(sizeof(qadav_barat_group));
    if (!grp) return 0;
    barat_memset(grp, 0, sizeof(qadav_barat_group));
    grp->group_id = group_id;
    grp->step_index = 0;
    return grp;
}

void qadav_barat_destroy_group(qadav_barat_group* grp) {
    if (!grp) return;
    for (int i = 0; i < grp->submission_count; i++) {
        if (grp->submissions[i].data) {
            qadav_free(grp->submissions[i].data);
            grp->submissions[i].data = 0;
        }
    }
    qadav_free(grp);
}

int qadav_barat_submit(
    qadav_barat_group* grp,
    int worker_id,
    const void* grad_data,
    unsigned long long grad_size
) {
    if (!grp || !grad_data || grad_size == 0 || grad_size > QADAV_BARAT_GRAD_MAX) {
        return -1;
    }
    if (grp->submission_count >= QADAV_BARAT_MAX_WORKERS) {
        return -2;
    }

    int idx = grp->submission_count;
    grp->submissions[idx].worker_id = worker_id;
    grp->submissions[idx].size = grad_size;

    // Generate deterministic UUID for worker
    grp->submissions[idx].worker_uuid[0] = 0xBA;
    grp->submissions[idx].worker_uuid[1] = 0xAA;
    grp->submissions[idx].worker_uuid[2] = (unsigned char)(worker_id & 0xFF);
    for (int k = 3; k < 16; k++) {
        grp->submissions[idx].worker_uuid[k] = (unsigned char)((worker_id * 31 + k * 17) & 0xFF);
    }

    // Allocate gradient buffer copy
    void* copy = qadav_alloc(grad_size);
    if (!copy) return -3;
    barat_memcpy(copy, grad_data, grad_size);
    grp->submissions[idx].data = copy;

    // Compute gradient digest
    qadav_sha3_256((const unsigned char*)grad_data, grad_size, grp->submissions[idx].digest);

    grp->submission_count++;
    return 0;
}

int qadav_barat_aggregate(
    qadav_barat_group* grp,
    void* out_tensor,
    unsigned long long out_size,
    qadav_barat_cert* out_cert
) {
    if (!grp || !out_tensor || !out_cert || grp->submission_count <= 0) {
        return -1;
    }

    int n = grp->submission_count;
    unsigned long long size = grp->submissions[0].size;
    if (out_size < size) return -2;

    for (int i = 1; i < n; i++) {
        if (grp->submissions[i].size != size) return -3;
    }

    // Sort submissions into canonical UUID order (deterministic)
    for (int i = 0; i < n - 1; i++) {
        for (int j = i + 1; j < n; j++) {
            if (barat_uuid_cmp(grp->submissions[i].worker_uuid, grp->submissions[j].worker_uuid) > 0) {
                qadav_barat_submission tmp = grp->submissions[i];
                grp->submissions[i] = grp->submissions[j];
                grp->submissions[j] = tmp;
            }
        }
    }

    // Compute gradient norm (or sum of abs values) for each worker
    double norms[QADAV_BARAT_MAX_WORKERS];
    for (int i = 0; i < n; i++) {
        double total = 0.0;
        if (size % 8 == 0) {
            double* dptr = (double*)grp->submissions[i].data;
            unsigned long long count = size / 8;
            for (unsigned long long k = 0; k < count; k++) {
                double v = dptr[k];
                total += (v >= 0.0 ? v : -v);
            }
        } else {
            unsigned char* bptr = (unsigned char*)grp->submissions[i].data;
            for (unsigned long long k = 0; k < size; k++) {
                total += (double)bptr[k];
            }
        }
        norms[i] = total;
    }

    // Determine trimmed-mean selection (alpha = 0.20 -> drop top 20% and bottom 20% outliers)
    int k_trim = (int)((double)n * 0.20);
    unsigned int inc_mask = 0;

    if (n >= 5 && k_trim >= 1) {
        // Rank indices by norm
        int ranked[QADAV_BARAT_MAX_WORKERS];
        for (int i = 0; i < n; i++) ranked[i] = i;

        for (int i = 0; i < n - 1; i++) {
            for (int j = i + 1; j < n; j++) {
                if (norms[ranked[i]] > norms[ranked[j]]) {
                    int t = ranked[i];
                    ranked[i] = ranked[j];
                    ranked[j] = t;
                }
            }
        }

        // Include only the central indices (excluding k_trim lowest and k_trim highest)
        for (int idx = k_trim; idx < n - k_trim; idx++) {
            int w = ranked[idx];
            inc_mask |= (1u << w);
        }
    } else {
        // If n < 5, include all
        for (int i = 0; i < n; i++) {
            inc_mask |= (1u << i);
        }
    }

    // Aggregate included gradients
    int active_workers = 0;
    for (int i = 0; i < n; i++) {
        if (inc_mask & (1u << i)) {
            active_workers++;
        }
    }
    if (active_workers <= 0) {
        active_workers = 1;
        inc_mask = 1;
    }

    if (size % 8 == 0) {
        unsigned long long count = size / 8;
        double* out_dbl = (double*)out_tensor;
        for (unsigned long long k = 0; k < count; k++) {
            double sum = 0.0;
            for (int i = 0; i < n; i++) {
                if (inc_mask & (1u << i)) {
                    double* dptr = (double*)grp->submissions[i].data;
                    sum += dptr[k];
                }
            }
            out_dbl[k] = sum / (double)active_workers;
        }
    } else {
        unsigned char* out_b = (unsigned char*)out_tensor;
        for (unsigned long long k = 0; k < size; k++) {
            unsigned int sum = 0;
            for (int i = 0; i < n; i++) {
                if (inc_mask & (1u << i)) {
                    unsigned char* bptr = (unsigned char*)grp->submissions[i].data;
                    sum += (unsigned int)bptr[k];
                }
            }
            out_b[k] = (unsigned char)(sum / active_workers);
        }
    }

    // Populate certificate
    barat_memset(out_cert, 0, sizeof(qadav_barat_cert));
    grp->step_index++;
    out_cert->step_index = grp->step_index;
    out_cert->timestamp = 1700000000ULL + grp->step_index;
    out_cert->worker_count = n;
    out_cert->inclusion_mask = inc_mask;
    out_cert->trimmed_alpha = 0.20;

    for (int i = 0; i < n; i++) {
        barat_memcpy(out_cert->worker_uuids[i], grp->submissions[i].worker_uuid, 16);
        barat_memcpy(out_cert->gradient_digests[i], grp->submissions[i].digest, 32);
    }
    barat_memcpy(out_cert->prev_record_hash, grp->last_record_hash, 32);

    // Compute selection proof: SHA3-256 over inclusion mask, alpha, uuids, digests
    // Buffer: 4 (mask) + 8 (alpha) + n*16 (uuids) + n*32 (digests)
    unsigned char proof_buf[4 + 8 + QADAV_BARAT_MAX_WORKERS * 16 + QADAV_BARAT_MAX_WORKERS * 32];
    barat_memset(proof_buf, 0, sizeof(proof_buf));
    barat_memcpy(proof_buf, &inc_mask, 4);
    double alpha = 0.20;
    barat_memcpy(proof_buf + 4, &alpha, 8);
    barat_memcpy(proof_buf + 12, out_cert->worker_uuids, (unsigned long long)n * 16);
    barat_memcpy(proof_buf + 12 + n * 16, out_cert->gradient_digests, (unsigned long long)n * 32);
    unsigned long long proof_len = 12 + (unsigned long long)n * 48;
    qadav_sha3_256(proof_buf, proof_len, out_cert->selection_proof);

    // Compute aggregated hash over the aggregated tensor
    qadav_sha3_256((const unsigned char*)out_tensor, size, out_cert->aggregated_hash);

    // Compute record_hash over the contiguous preimage
    qadav_sha3_256((const unsigned char*)out_cert, CERT_PREIMAGE_LEN, out_cert->record_hash);

    barat_memcpy(grp->last_record_hash, out_cert->record_hash, 32);
    return 0;
}

int qadav_barat_verify(
    const qadav_barat_cert* cert,
    const void* aggregated_tensor,
    unsigned long long tensor_size
) {
    if (!cert) {
        return 0;
    }

    // 1. Recompute aggregated_hash if tensor provided
    if (aggregated_tensor && tensor_size > 0) {
        unsigned char expected_agg[32];
        qadav_sha3_256((const unsigned char*)aggregated_tensor, tensor_size, expected_agg);
        if (qadav_sha3_ct_compare(expected_agg, cert->aggregated_hash, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    // 2. Recompute selection proof
    int n = cert->worker_count;
    if (n < 1 || n > QADAV_BARAT_MAX_WORKERS) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }
    unsigned char proof_buf[4 + 8 + QADAV_BARAT_MAX_WORKERS * 16 + QADAV_BARAT_MAX_WORKERS * 32];
    barat_memset(proof_buf, 0, sizeof(proof_buf));
    barat_memcpy(proof_buf, &cert->inclusion_mask, 4);
    barat_memcpy(proof_buf + 4, &cert->trimmed_alpha, 8);
    barat_memcpy(proof_buf + 12, cert->worker_uuids, (unsigned long long)n * 16);
    barat_memcpy(proof_buf + 12 + n * 16, cert->gradient_digests, (unsigned long long)n * 32);
    unsigned long long proof_len = 12 + (unsigned long long)n * 48;

    unsigned char expected_proof[32];
    qadav_sha3_256(proof_buf, proof_len, expected_proof);
    if (qadav_sha3_ct_compare(expected_proof, cert->selection_proof, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    // 3. Recompute record_hash over CERT_PREIMAGE_LEN
    unsigned char expected_record_hash[32];
    qadav_sha3_256((const unsigned char*)cert, CERT_PREIMAGE_LEN, expected_record_hash);
    if (qadav_sha3_ct_compare(expected_record_hash, cert->record_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    return 1;
}

int qadav_barat_attest(
    qadav_reasoning_chain* chain,
    const qadav_barat_cert* cert,
    const void* aggregated_tensor,
    unsigned long long tensor_size
) {
    if (!cert || !aggregated_tensor) return 0;

    if (!qadav_barat_verify(cert, aggregated_tensor, tensor_size)) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    if (chain) {
        int r = qadav_qera_reason(
            chain,
            QADAV_OP_VERIFY,
            cert->aggregated_hash, 32,
            cert->record_hash, 32,
            cert->selection_proof, 32,
            1.0,
            0
        );
        if (r != 0) return 0;
    }

    return 1;
}

int qadav_barat_corrupt(
    qadav_barat_cert* cert,
    int offset
) {
    if (!cert) return -1;
    if (offset < 0 || offset >= (int)sizeof(qadav_barat_cert)) {
        offset = 0;
    }
    unsigned char* b = (unsigned char*)cert;
    b[offset] ^= 0xFF;
    return 0;
}
