#include "qadav_rsca.h"
#include "qadav_sha3.h"
#include "qadav_qeeh.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static qadav_rsca_table g_rsca_table = {0};

static inline unsigned long long rsca_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64)
    unsigned int lo, hi;
    __asm__ __volatile__ ("rdtsc" : "=a" (lo), "=d" (hi));
    return ((unsigned long long)hi << 32) | lo;
#else
    return 135792468ULL;
#endif
}

static void rsca_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void rsca_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void compute_rsca_record_hash(const qadav_rsca_record* rec, unsigned char out_hash[32]) {
    // 144 bytes:
    // stage (8) || timestamp (8) || source_hash (32) || compiler_hash (32) || output_hash (32) || prev_stage_hash (32)
    unsigned char preimage[QADAV_RSCA_PREIMAGE_SIZE];
    rsca_memcpy(preimage, &rec->stage, 8);
    rsca_memcpy(preimage + 8, &rec->timestamp, 8);
    rsca_memcpy(preimage + 16, rec->source_hash, 32);
    rsca_memcpy(preimage + 48, rec->compiler_hash, 32);
    rsca_memcpy(preimage + 80, rec->output_hash, 32);
    rsca_memcpy(preimage + 112, rec->prev_stage_hash, 32);

    qadav_sha3_256(preimage, QADAV_RSCA_PREIMAGE_SIZE, out_hash);
}

int qadav_rsca_record_stage(unsigned long long stage,
                             const unsigned char source_hash[32],
                             const unsigned char compiler_hash[32],
                             const unsigned char output_hash[32]) {
    if (g_rsca_table.n_records >= QADAV_RSCA_MAX_STAGES) {
        return -1;
    }

    int idx = g_rsca_table.n_records;
    qadav_rsca_record* rec = &g_rsca_table.records[idx];
    rsca_memset(rec, 0, sizeof(qadav_rsca_record));

    rec->stage = stage;
    rec->timestamp = rsca_rdtsc();
    if (source_hash) {
        rsca_memcpy(rec->source_hash, source_hash, 32);
    }
    if (compiler_hash) {
        rsca_memcpy(rec->compiler_hash, compiler_hash, 32);
    }
    if (output_hash) {
        rsca_memcpy(rec->output_hash, output_hash, 32);
    }

    if (idx > 0) {
        rsca_memcpy(rec->prev_stage_hash, g_rsca_table.records[idx - 1].record_hash, 32);
    } else {
        rsca_memset(rec->prev_stage_hash, 0, 32);
    }

    compute_rsca_record_hash(rec, rec->record_hash);
    g_rsca_table.n_records++;
    return idx;
}

int qadav_rsca_verify_record(int stage_idx) {
    if (stage_idx < 0 || stage_idx >= g_rsca_table.n_records) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    qadav_rsca_record* rec = &g_rsca_table.records[stage_idx];
    unsigned char expected_hash[32];
    compute_rsca_record_hash(rec, expected_hash);

    if (qadav_sha3_ct_compare(expected_hash, rec->record_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    if (stage_idx > 0) {
        if (qadav_sha3_ct_compare(rec->prev_stage_hash, g_rsca_table.records[stage_idx - 1].record_hash, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    return 1;
}

int qadav_rsca_compare_outputs(int stage_a, int stage_b) {
    if (stage_a < 0 || stage_a >= g_rsca_table.n_records ||
        stage_b < 0 || stage_b >= g_rsca_table.n_records) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    const qadav_rsca_record* rec_a = &g_rsca_table.records[stage_a];
    const qadav_rsca_record* rec_b = &g_rsca_table.records[stage_b];

    if (qadav_sha3_ct_compare(rec_a->output_hash, rec_b->output_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    return 1;
}

int qadav_rsca_attest_chain(void) {
    if (g_rsca_table.n_records == 0) {
        return 0;
    }

    for (int i = 0; i < g_rsca_table.n_records; i++) {
        if (!qadav_rsca_verify_record(i)) {
            return 0;
        }
    }

    // If we have at least 2 stages, verify outputs of stage 1 and stage 2 if available
    if (g_rsca_table.n_records >= 3) {
        if (!qadav_rsca_compare_outputs(1, 2)) {
            return 0;
        }
    }

    g_rsca_table.is_verified = 1;
    return 1;
}

int qadav_rsca_corrupt_record(int stage_idx, int field_type) {
    if (stage_idx < 0 || stage_idx >= g_rsca_table.n_records) {
        return -1;
    }

    qadav_rsca_record* rec = &g_rsca_table.records[stage_idx];
    switch (field_type) {
        case 1:
            rec->output_hash[0] ^= 0xFF;
            break;
        case 2:
            rec->compiler_hash[0] ^= 0xFF;
            break;
        case 3:
            rec->source_hash[0] ^= 0xFF;
            break;
        case 4:
            rec->prev_stage_hash[0] ^= 0xFF;
            break;
        default:
            rec->record_hash[0] ^= 0xFF;
            break;
    }
    return 0;
}

int qadav_rsca_get_record_count(void) {
    return g_rsca_table.n_records;
}

void qadav_rsca_reset(void) {
    rsca_memset(&g_rsca_table, 0, sizeof(qadav_rsca_table));
}
