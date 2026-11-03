#include "qadav_sea.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void  qadav_qeeh_trigger_decoherence(void* qstate);

#ifdef _WIN32
void __stdcall ExitProcess(unsigned int uExitCode);
#endif

#define SEA_PREIMAGE_LEN 144

static void sea_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void sea_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static int sea_is_all_zeros(const unsigned char* b, int len) {
    for (int i = 0; i < len; i++) {
        if (b[i] != 0) return 0;
    }
    return 1;
}

sea_history* sea_create(void) {
    sea_history* h = (sea_history*)qadav_alloc(sizeof(sea_history));
    if (!h) return 0;
    sea_memset(h, 0, sizeof(sea_history));
    return h;
}

int sea_propose(
    sea_history* h,
    const unsigned char from[32],
    const unsigned char to[32],
    const unsigned char reason[32],
    sea_proposal** out
) {
    if (!h || !from || !to || !reason) {
        return -1;
    }
    if (sea_is_all_zeros(from, 32) || sea_is_all_zeros(to, 32) || sea_is_all_zeros(reason, 32)) {
        return -2;
    }
    if (h->count >= QADAV_SEA_MAX_HISTORY) {
        return -3;
    }

    int idx = h->count;
    sea_proposal* p = &h->proposals[idx];
    sea_memset(p, 0, sizeof(sea_proposal));

    p->step_index = (unsigned long long)(idx + 1);
    p->timestamp = 1700000000ULL + p->step_index;

    sea_memcpy(p->from_state_hash, from, 32);
    sea_memcpy(p->to_state_hash, to, 32);
    sea_memcpy(p->reason_hash, reason, 32);

    if (idx > 0) {
        sea_memcpy(p->prev_proposal_hash, h->proposals[idx - 1].proposal_hash, 32);
    } else {
        sea_memset(p->prev_proposal_hash, 0, 32);
    }

    // Preimage = step_index(8) + timestamp(8) + from(32) + to(32) + reason(32) + prev(32) = 144 bytes
    // Explicit packing without struct padding assumptions:
    unsigned char preimage[SEA_PREIMAGE_LEN];
    sea_memcpy(preimage + 0, &p->step_index, 8);
    sea_memcpy(preimage + 8, &p->timestamp, 8);
    sea_memcpy(preimage + 16, p->from_state_hash, 32);
    sea_memcpy(preimage + 48, p->to_state_hash, 32);
    sea_memcpy(preimage + 80, p->reason_hash, 32);
    sea_memcpy(preimage + 112, p->prev_proposal_hash, 32);
    qadav_sha3_256(preimage, SEA_PREIMAGE_LEN, p->proposal_hash);

    // Update cumulative lineage root = SHA3-256(prev_lineage_root(32) || proposal_hash(32))
    unsigned char lineage_buf[64];
    sea_memcpy(lineage_buf, h->lineage_root, 32);
    sea_memcpy(lineage_buf + 32, p->proposal_hash, 32);
    qadav_sha3_256(lineage_buf, 64, h->lineage_root);

    h->count++;
    if (out) {
        *out = p;
    }
    return 0;
}

int sea_verify_proposal(const sea_proposal* p, const sea_proposal* prev) {
    if (!p) return 0;

    // Check linkage to previous proposal
    if (prev) {
        if (p->step_index != prev->step_index + 1) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
        if (qadav_sha3_ct_compare(p->prev_proposal_hash, prev->proposal_hash, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    } else if (p->step_index == 1) {
        if (!sea_is_all_zeros(p->prev_proposal_hash, 32)) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    // Recompute proposal hash over the 144-byte preimage
    unsigned char preimage[SEA_PREIMAGE_LEN];
    sea_memcpy(preimage + 0, &p->step_index, 8);
    sea_memcpy(preimage + 8, &p->timestamp, 8);
    sea_memcpy(preimage + 16, p->from_state_hash, 32);
    sea_memcpy(preimage + 48, p->to_state_hash, 32);
    sea_memcpy(preimage + 80, p->reason_hash, 32);
    sea_memcpy(preimage + 112, p->prev_proposal_hash, 32);

    unsigned char expected_hash[32];
    qadav_sha3_256(preimage, SEA_PREIMAGE_LEN, expected_hash);

    if (qadav_sha3_ct_compare(expected_hash, p->proposal_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    return 1;
}

int sea_verify_history(const sea_history* h) {
    if (!h || h->count <= 0) return 0;

    for (int i = 0; i < h->count; i++) {
        const sea_proposal* prev = (i > 0) ? &h->proposals[i - 1] : 0;
        if (!sea_verify_proposal(&h->proposals[i], prev)) {
            return 0;
        }
    }
    return 1;
}

int sea_attest(sea_history* h, qadav_reasoning_chain* chain) {
    if (!h) return 0;

    if (!sea_verify_history(h)) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    if (chain) {
        for (int i = 0; i < h->count; i++) {
            const sea_proposal* p = &h->proposals[i];
            int r = qadav_qera_reason(
                chain,
                QADAV_OP_EVOLVE,
                p->from_state_hash, 32,
                p->to_state_hash, 32,
                p->reason_hash, 32,
                1.0,
                0
            );
            if (r != 0) return 0;
        }
    }

    return 1;
}

int sea_corrupt(sea_proposal* p, int offset) {
    if (!p) return -1;
    if (offset < 0 || offset >= (int)sizeof(sea_proposal)) {
        offset = 0;
    }
    unsigned char* b = (unsigned char*)p;
    b[offset] ^= 0xFF;
    return 0;
}

void sea_hash_text(const char* text, unsigned char out[32]) {
    if (!out) return;
    if (!text) {
        sea_memset(out, 0, 32);
        return;
    }
    unsigned long long len = 0;
    while (text[len]) len++;
    qadav_sha3_256((const unsigned char*)text, len, out);
}

void sea_free(sea_history* h) {
    if (h) {
        qadav_free(h);
    }
}
