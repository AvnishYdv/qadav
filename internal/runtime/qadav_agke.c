#include "qadav_agke.h"
#include "qadav_gpu.h"
#include "qadav_qera.h"
#include "qadav_qeeh.h"
#include "qadav_sha3.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static qadav_agke_table g_agke_table = {0};
static unsigned long long g_agke_dispatch_seq = 0;

static inline unsigned long long agke_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64)
    unsigned int lo, hi;
    __asm__ __volatile__ ("rdtsc" : "=a" (lo), "=d" (hi));
    return ((unsigned long long)hi << 32) | lo;
#else
    return 87654321ULL;
#endif
}

static void agke_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void agke_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void compute_agke_record_hash(const qadav_agke_record* rec, unsigned char out_hash[32]) {
    // Hash: dispatch_id (8) || timestamp (8) || kernel_hash (32) || input_digest (32) || output_digest (32) || device_uuid (32)
    unsigned char preimage[144];
    agke_memcpy(preimage, &rec->dispatch_id, 8);
    agke_memcpy(preimage + 8, &rec->timestamp, 8);
    agke_memcpy(preimage + 16, rec->kernel_hash, 32);
    agke_memcpy(preimage + 48, rec->input_digest, 32);
    agke_memcpy(preimage + 80, rec->output_digest, 32);
    agke_memcpy(preimage + 112, rec->device_uuid, 32);

    qadav_sha3_256(preimage, 144, out_hash);
}

int qadav_agke_record_dispatch(const char* kernel_name,
                                const unsigned char* kernel_bytecode, unsigned long long bytecode_len,
                                const void* input_data, unsigned long long input_len,
                                const void* output_data, unsigned long long output_len) {
    if (g_agke_table.n_records >= QADAV_AGKE_MAX_RECORDS) {
        return -1;
    }

    qadav_agke_record* rec = &g_agke_table.records[g_agke_table.n_records];
    agke_memset(rec, 0, sizeof(qadav_agke_record));

    g_agke_dispatch_seq++;
    rec->dispatch_id = g_agke_dispatch_seq;
    rec->timestamp = agke_rdtsc();

    if (kernel_name) {
        for (int i = 0; kernel_name[i] && i < 63; i++) {
            rec->kernel_name[i] = kernel_name[i];
        }
    }

    // 1. Kernel hash
    if (kernel_bytecode && bytecode_len > 0) {
        qadav_sha3_256(kernel_bytecode, bytecode_len, rec->kernel_hash);
    } else if (kernel_name) {
        unsigned long long klen = 0;
        while (kernel_name[klen]) klen++;
        qadav_sha3_256((const unsigned char*)kernel_name, klen, rec->kernel_hash);
    }

    // 2. Input digest
    if (input_data && input_len > 0) {
        qadav_sha3_256((const unsigned char*)input_data, input_len, rec->input_digest);
    }

    // 3. Output digest
    if (output_data && output_len > 0) {
        qadav_sha3_256((const unsigned char*)output_data, output_len, rec->output_digest);
    }

    // 4. Device UUID from qadav_gpu
    qadav_gpu_get_device_uuid(rec->device_uuid);

    // 5. Compute compound record hash
    compute_agke_record_hash(rec, rec->record_hash);

    g_agke_table.n_records++;
    return 0;
}

int qadav_agke_attest(void* reasoning_chain, void* auth_boundary) {
    if (g_agke_table.n_records == 0) {
        g_agke_table.is_attested = 1;
        return 1;
    }

    qadav_reasoning_chain* chain = (qadav_reasoning_chain*)reasoning_chain;

    for (int i = 0; i < g_agke_table.n_records; i++) {
        qadav_agke_record* rec = &g_agke_table.records[i];

        if (chain) {
            // Bind record into host QERA chain with QADAV_OP_VERIFY and confidence 1.0
            int res = qadav_qera_reason(chain,
                                        QADAV_OP_VERIFY,
                                        rec->record_hash, 32,
                                        rec->output_digest, 32,
                                        rec->kernel_hash, 32,
                                        1.0,
                                        auth_boundary);
            if (res != 0) {
                return 0;
            }
        }
    }

    g_agke_table.is_attested = 1;
    return 1;
}

int qadav_agke_verify(void* reasoning_chain, void* auth_boundary) {
    unsigned char current_device_uuid[32];
    qadav_gpu_get_device_uuid(current_device_uuid);

    // 1. Verify every AGKE record against compound hash and device UUID
    for (int i = 0; i < g_agke_table.n_records; i++) {
        qadav_agke_record* rec = &g_agke_table.records[i];

        // Device UUID binding check
        if (qadav_sha3_ct_compare(rec->device_uuid, current_device_uuid, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }

        // Recompute record compound hash
        unsigned char expected_record_hash[32];
        compute_agke_record_hash(rec, expected_record_hash);

        if (qadav_sha3_ct_compare(expected_record_hash, rec->record_hash, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    // 2. If QERA chain is provided, verify QERA chain integrity
    if (reasoning_chain) {
        qadav_reasoning_chain* chain = (qadav_reasoning_chain*)reasoning_chain;
        if (!qadav_qera_verify(chain, auth_boundary)) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    g_agke_table.is_verified = 1;
    return 1;
}

int qadav_agke_corrupt_record(int record_index, int field_type) {
    if (record_index < 0 || record_index >= g_agke_table.n_records) {
        return -1;
    }

    qadav_agke_record* rec = &g_agke_table.records[record_index];

    switch (field_type) {
        case 1:
            // Corrupt output digest
            rec->output_digest[0] ^= 0xFF;
            break;
        case 2:
            // Corrupt device UUID
            rec->device_uuid[0] ^= 0xFF;
            break;
        case 3:
            // Corrupt kernel hash
            rec->kernel_hash[0] ^= 0xFF;
            break;
        default:
            // Corrupt compound record hash
            rec->record_hash[0] ^= 0xFF;
            break;
    }

    return 0;
}

int qadav_agke_get_count(void) {
    return g_agke_table.n_records;
}

const qadav_agke_record* qadav_agke_get_record(int index) {
    if (index < 0 || index >= g_agke_table.n_records) {
        return 0;
    }
    return &g_agke_table.records[index];
}

void qadav_agke_reset(void) {
    g_agke_table.n_records = 0;
    g_agke_table.is_attested = 0;
    g_agke_table.is_verified = 0;
    g_agke_dispatch_seq = 0;
}

