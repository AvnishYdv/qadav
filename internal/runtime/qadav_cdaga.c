#include "qadav_cdaga.h"
#include "qadav_qeeh.h"
#include "qadav_sha3.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static inline unsigned long long cdaga_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64)
    unsigned int lo, hi;
    __asm__ __volatile__ ("rdtsc" : "=a" (lo), "=d" (hi));
    return ((unsigned long long)hi << 32) | lo;
#else
    return 98765432ULL;
#endif
}

static void cdaga_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void cdaga_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static int uuid_compare(const unsigned char a[16], const unsigned char b[16]) {
    for (int i = 0; i < 16; i++) {
        if (a[i] < b[i]) return -1;
        if (a[i] > b[i]) return 1;
    }
    return 0;
}

static void compute_canonical_aggregation_digest(
    unsigned long long step_index,
    unsigned long long timestamp,
    int device_count,
    const unsigned char uuids[QADAV_CDAGA_MAX_DEVICES][16],
    const unsigned char grads[QADAV_CDAGA_MAX_DEVICES][32],
    const unsigned char topology_hash[32],
    unsigned char out_digest[32]
) {
    int indices[QADAV_CDAGA_MAX_DEVICES];
    for (int i = 0; i < device_count; i++) {
        indices[i] = i;
    }

    // Canonical lexicographic sort by UUID
    for (int i = 0; i < device_count - 1; i++) {
        for (int j = 0; j < device_count - i - 1; j++) {
            if (uuid_compare(uuids[indices[j]], uuids[indices[j + 1]]) > 0) {
                int tmp = indices[j];
                indices[j] = indices[j + 1];
                indices[j + 1] = tmp;
            }
        }
    }

    // Preimage: step_index (8) || timestamp (8) || [sorted uuid (16) || grad (32)] * count || topology_hash (32)
    unsigned char preimage[512];
    unsigned long long offset = 0;

    cdaga_memcpy(preimage + offset, &step_index, 8);
    offset += 8;
    cdaga_memcpy(preimage + offset, &timestamp, 8);
    offset += 8;

    for (int i = 0; i < device_count; i++) {
        int idx = indices[i];
        cdaga_memcpy(preimage + offset, uuids[idx], 16);
        offset += 16;
        cdaga_memcpy(preimage + offset, grads[idx], 32);
        offset += 32;
    }

    cdaga_memcpy(preimage + offset, topology_hash, 32);
    offset += 32;

    qadav_sha3_256(preimage, offset, out_digest);
}

int qadav_cdaga_record_aggregation_with_uuids(
    const unsigned char uuids[QADAV_CDAGA_MAX_DEVICES][16],
    const unsigned char topology_hash[32],
    unsigned long long step_index,
    void** gradient_buffers,
    unsigned long long* gradient_sizes,
    int buffer_count,
    qadav_cdaga_record* out
) {
    if (!out) {
        return -1;
    }

    if (buffer_count <= 0) buffer_count = 1;
    int count = (buffer_count > QADAV_CDAGA_MAX_DEVICES) ? QADAV_CDAGA_MAX_DEVICES : buffer_count;

    cdaga_memset(out, 0, sizeof(qadav_cdaga_record));

    out->step_index = step_index;
    out->timestamp = (step_index > 0) ? (step_index * 1000000ULL) : 1000000ULL;
    out->device_count = count;

    unsigned char raw_digests[QADAV_CDAGA_MAX_DEVICES][32];
    cdaga_memset(raw_digests, 0, sizeof(raw_digests));

    for (int i = 0; i < count; i++) {
        if (gradient_buffers && gradient_sizes && gradient_buffers[i] && gradient_sizes[i] > 0) {
            qadav_sha3_256((const unsigned char*)gradient_buffers[i], gradient_sizes[i], raw_digests[i]);
        } else {
            unsigned char dummy_grad[16];
            for (int k = 0; k < 16; k++) dummy_grad[k] = (unsigned char)(i * 17 + k + 1);
            qadav_sha3_256(dummy_grad, 16, raw_digests[i]);
        }
    }

    // Determine canonical sort permutation of uuids
    int indices[QADAV_CDAGA_MAX_DEVICES];
    for (int i = 0; i < count; i++) indices[i] = i;
    for (int i = 0; i < count - 1; i++) {
        for (int j = 0; j < count - i - 1; j++) {
            if (uuid_compare(uuids[indices[j]], uuids[indices[j + 1]]) > 0) {
                int tmp = indices[j];
                indices[j] = indices[j + 1];
                indices[j + 1] = tmp;
            }
        }
    }

    // Store in canonical order in out
    for (int i = 0; i < count; i++) {
        int idx = indices[i];
        cdaga_memcpy(out->device_uuids[i], uuids[idx], 16);
        cdaga_memcpy(out->gradient_digests[i], raw_digests[idx], 32);
    }

    if (topology_hash) {
        cdaga_memcpy(out->topology_hash, topology_hash, 32);
    } else {
        qadav_sha3_256((const unsigned char*)"standalone_cdaga", 16, out->topology_hash);
    }

    // Compute canonical aggregation digest
    compute_canonical_aggregation_digest(
        out->step_index,
        out->timestamp,
        out->device_count,
        out->device_uuids,
        out->gradient_digests,
        out->topology_hash,
        out->aggregation_digest
    );

    // Compute record_hash over the 468-byte contiguous record preimage
    qadav_sha3_256((const unsigned char*)out, 468, out->record_hash);

    return 0;
}

int qadav_cdaga_record_aggregation(
    qadav_multigpu_context* ctx,
    unsigned long long step_index,
    void** gradient_buffers,
    unsigned long long* gradient_sizes,
    int buffer_count,
    qadav_cdaga_record* out
) {
    if (!out) {
        return -1;
    }

    if (buffer_count <= 0) buffer_count = 1;

    qadav_multigpu_context* local_ctx = ctx;
    int owns_ctx = 0;
    if (!local_ctx) {
        local_ctx = qadav_multigpu_init();
        owns_ctx = 1;
    }

    int count = (buffer_count > QADAV_CDAGA_MAX_DEVICES) ? QADAV_CDAGA_MAX_DEVICES : buffer_count;
    unsigned char uuids[QADAV_CDAGA_MAX_DEVICES][16];
    cdaga_memset(uuids, 0, sizeof(uuids));

    for (int i = 0; i < count; i++) {
        if (local_ctx && i < local_ctx->device_count) {
            cdaga_memcpy(uuids[i], local_ctx->devices[i].uuid, 16);
        } else if (local_ctx && local_ctx->device_count > 0) {
            cdaga_memcpy(uuids[i], local_ctx->devices[0].uuid, 16);
            uuids[i][15] ^= (unsigned char)(i + 1);
        } else {
            uuids[i][0] = 0xDE;
            uuids[i][1] = 0xAD;
            uuids[i][15] = (unsigned char)(i + 1);
        }
    }

    unsigned char topo[32];
    if (local_ctx) {
        cdaga_memcpy(topo, local_ctx->topology_hash, 32);
    } else {
        qadav_sha3_256((const unsigned char*)"standalone_cdaga", 16, topo);
    }

    int res = qadav_cdaga_record_aggregation_with_uuids(
        uuids, topo, step_index, gradient_buffers, gradient_sizes, count, out
    );

    if (owns_ctx && local_ctx) {
        qadav_multigpu_shutdown(local_ctx);
    }

    return res;
}

int qadav_cdaga_verify(
    const qadav_cdaga_record* record,
    const qadav_multigpu_context* ctx
) {
    if (!record) return 0;

    // 1. Recompute canonical aggregation digest
    unsigned char expected_agg[32];
    compute_canonical_aggregation_digest(
        record->step_index,
        record->timestamp,
        record->device_count,
        record->device_uuids,
        record->gradient_digests,
        record->topology_hash,
        expected_agg
    );

    if (qadav_sha3_ct_compare(expected_agg, record->aggregation_digest, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    // 2. Recompute 468-byte record hash
    unsigned char expected_record_hash[32];
    qadav_sha3_256((const unsigned char*)record, 468, expected_record_hash);

    if (qadav_sha3_ct_compare(expected_record_hash, record->record_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    // 3. Topology consistency check if context is provided
    if (ctx) {
        if (qadav_sha3_ct_compare(ctx->topology_hash, record->topology_hash, 32) != 0) {
            return 0;
        }
    }

    return 1;
}

int qadav_cdaga_attest(
    qadav_reasoning_chain* chain,
    const qadav_cdaga_record* record,
    const qadav_multigpu_context* ctx
) {
    if (!record) return 0;

    // Diagnostic verification first
    if (!qadav_cdaga_verify(record, ctx)) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    // Bind into QERA chain as a verification reasoning step
    if (chain) {
        int r = qadav_qera_reason(
            chain,
            QADAV_OP_VERIFY,
            record->aggregation_digest, 32,
            record->record_hash, 32,
            record->topology_hash, 32,
            1.0,
            0
        );
        if (r != 0) return 0;
    }

    return 1;
}

int qadav_cdaga_corrupt(
    qadav_cdaga_record* record,
    int offset
) {
    if (!record) return -1;
    if (offset < 0 || offset >= (int)sizeof(qadav_cdaga_record)) {
        offset = 0;
    }
    unsigned char* b = (unsigned char*)record;
    b[offset] ^= 0xFF;
    return 0;
}
