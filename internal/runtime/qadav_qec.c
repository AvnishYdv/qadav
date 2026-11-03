#include "qadav_qec.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);
void  qadav_write_raw(const char* s, unsigned long long len);
int   qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len);

#include "qadav_consolidate.h"

static void qec_memcpy(void* dst, const void* src, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    const volatile unsigned char* s = (const volatile unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qec_memset(void* dst, int v, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static unsigned long long qec_strlen(const char* s) {
    if (!s) return 0;
    unsigned long long len = 0;
    while (s[len]) len++;
    return len;
}

static int s_qec_create_count = 0;
static int s_qec_free_count = 0;

int qadav_qec_get_create_count(void) {
    return s_qec_create_count;
}

int qadav_qec_get_free_count(void) {
    return s_qec_free_count;
}

qadav_qec_memory* qadav_qec_create(const char* name, int code_type, int distance, void* boundary) {
    qadav_qec_memory* mem = (qadav_qec_memory*)qadav_alloc(sizeof(qadav_qec_memory));
    if (!mem) return 0;
    qec_memset(mem, 0, sizeof(qadav_qec_memory));

    if (name) {
        unsigned long long len = qec_strlen(name);
        if (len >= 64) len = 63;
        qec_memcpy(mem->name, name, len);
        mem->name[len] = '\0';
    }

    if (code_type != QADAV_QEC_SURFACE && code_type != QADAV_QEC_QLDPC) {
        code_type = QADAV_QEC_SURFACE;
    }
    mem->code_type = code_type;

    if (distance < 3 || distance > QADAV_QEC_MAX_DISTANCE || (distance % 2 == 0)) {
        distance = 3;
    }
    mem->distance = distance;

    if (boundary) {
        qadav_qeeh_boundary_hash(boundary, mem->authorized_boundary_id);
    }

    s_qec_create_count++;
    return mem;
}

int qadav_qec_measure(qadav_qec_memory* mem,
                      const unsigned char* raw, unsigned long long raw_len,
                      unsigned long long* out_syndrome_id,
                      void* auth_boundary) {
    if (!mem || !out_syndrome_id) return -1;

    // F11 / F15 pattern: validate raw_len > 0 and <= QADAV_QEC_SYNDROME_MAX
    if (!raw || raw_len == 0) {
        return -3; // Empty syndrome error
    }
    if (raw_len > QADAV_QEC_SYNDROME_MAX) {
        return -4; // Exceeds max syndrome size
    }

    // Reject all-zero syndromes (sanity check)
    int all_zeros = 1;
    for (unsigned long long i = 0; i < raw_len; i++) {
        if (raw[i] != 0) {
            all_zeros = 0;
            break;
        }
    }
    if (all_zeros) {
        return -3; // Reject all-zero syndrome
    }

    // Boundary authorization check
    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, mem->authorized_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -1;
        }
    }

    if (mem->syndrome_count >= QADAV_QEC_MAX_ENTRIES) {
        return -5; // Memory full
    }

    unsigned long long id = (unsigned long long)mem->syndrome_count;
    qadav_qec_syndrome* syn = &mem->syndromes[id];

    syn->syndrome_id = id;
    syn->timestamp = (unsigned long long)id + 100;
    syn->raw_len = raw_len;
    qec_memcpy(syn->raw_bytes, raw, raw_len);
    qec_memcpy(syn->boundary_hash, mem->authorized_boundary_id, 32);

    // Compute commitment: SHA3(syndrome_id (8B LE) || timestamp (8B LE) || raw_bytes (raw_len) || boundary_hash (32B))
    unsigned char commit_preimage[16 + QADAV_QEC_SYNDROME_MAX + 32];
    qec_memcpy(commit_preimage, &syn->syndrome_id, 8);
    qec_memcpy(commit_preimage + 8, &syn->timestamp, 8);
    qec_memcpy(commit_preimage + 16, syn->raw_bytes, raw_len);
    qec_memcpy(commit_preimage + 16 + raw_len, syn->boundary_hash, 32);

    qadav_sha3_256(commit_preimage, 48 + raw_len, syn->commitment);

    // Sanity check: first 8 bytes of commitment non-zero
    int first8_zeros = 1;
    for (int i = 0; i < 8; i++) {
        if (syn->commitment[i] != 0) {
            first8_zeros = 0;
            break;
        }
    }
    if (first8_zeros) {
        return -7;
    }

    // Update rolling chain root
    unsigned char root_buf[64];
    qec_memcpy(root_buf, mem->chain_root, 32);
    qec_memcpy(root_buf + 32, syn->commitment, 32);
    qadav_sha3_256(root_buf, 64, mem->chain_root);

    syn->verified = 0;
    mem->syndrome_count++;
    *out_syndrome_id = id;

    return 0;
}

int qadav_qec_verify(qadav_qec_memory* mem,
                     unsigned long long syndrome_id,
                     void* auth_boundary) {
    if (!mem) return -1;

    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, mem->authorized_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -1;
        }
    }

    if (syndrome_id >= (unsigned long long)mem->syndrome_count) {
        return -1;
    }

    qadav_qec_syndrome* syn = &mem->syndromes[syndrome_id];

    // Recompute commitment
    unsigned char commit_preimage[16 + QADAV_QEC_SYNDROME_MAX + 32];
    qec_memcpy(commit_preimage, &syn->syndrome_id, 8);
    qec_memcpy(commit_preimage + 8, &syn->timestamp, 8);
    qec_memcpy(commit_preimage + 16, syn->raw_bytes, syn->raw_len);
    qec_memcpy(commit_preimage + 16 + syn->raw_len, syn->boundary_hash, 32);

    unsigned char expected_commitment[32];
    qadav_sha3_256(commit_preimage, 48 + syn->raw_len, expected_commitment);

    // Constant-time compare
    if (qadav_sha3_ct_compare(expected_commitment, syn->commitment, 32) != 0) {
        // Tamper detected! Trigger active decoherence and terminate process
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return -2;
    }

    syn->verified = 1;
    return 0;
}

int qadav_qec_correct(qadav_qec_memory* mem,
                      unsigned long long syndrome_id,
                      void* auth_boundary) {
    if (!mem) return -1;

    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, mem->authorized_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -1;
        }
    }

    if (syndrome_id >= (unsigned long long)mem->syndrome_count) {
        return -1;
    }

    qadav_qec_syndrome* syn = &mem->syndromes[syndrome_id];

    // Defense-in-depth: Must be verified before correct
    if (!syn->verified) {
        return -6; // Unverified syndrome error
    }

    // Apply correction event to rolling chain root
    unsigned char root_buf[64];
    qec_memcpy(root_buf, mem->chain_root, 32);
    qec_memcpy(root_buf + 32, syn->commitment, 32);
    qadav_sha3_256(root_buf, 64, mem->chain_root);

    return 0;
}

int qadav_qec_corrupt_syndrome(qadav_qec_memory* mem,
                                unsigned long long syndrome_id,
                                int byte_offset) {
    if (!mem || syndrome_id >= (unsigned long long)mem->syndrome_count) return -1;

    qadav_qec_syndrome* syn = &mem->syndromes[syndrome_id];
    if (syn->raw_len == 0) return -2;

    int offset = byte_offset;
    if (offset < 0 || (unsigned long long)offset >= syn->raw_len) {
        offset = 0;
    }

    // Corrupt one raw byte without recomputing commitment
    syn->raw_bytes[offset] ^= 0xFF;
    return 0;
}

void qadav_qec_free(qadav_qec_memory* mem) {
    if (mem) {
        s_qec_free_count++;
        qec_memset(mem, 0, sizeof(qadav_qec_memory));
        qadav_free(mem);
    }
}
