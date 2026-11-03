#include "qadav_qecma.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);
void  qadav_write_raw(const char* s, unsigned long long len);

static void qecma_memcpy(void* dst, const void* src, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    const volatile unsigned char* s = (const volatile unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qecma_memset(void* dst, int v, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static unsigned long long qecma_strlen(const char* s) {
    if (!s) return 0;
    unsigned long long len = 0;
    while (s[len]) len++;
    return len;
}

qadav_qecma_bundle* qadav_qecma_create(void) {
    qadav_qecma_bundle* bundle = (qadav_qecma_bundle*)qadav_alloc(sizeof(qadav_qecma_bundle));
    if (!bundle) return 0;
    qecma_memset(bundle, 0, sizeof(qadav_qecma_bundle));
    return bundle;
}

void qadav_qecma_free(qadav_qecma_bundle* bundle) {
    if (bundle) {
        qecma_memset(bundle, 0, sizeof(qadav_qecma_bundle));
        qadav_free(bundle);
    }
}

int qadav_qecma_add(qadav_qecma_bundle* bundle, unsigned char tag, const void* data, unsigned long long len) {
    if (!bundle || bundle->count >= QADAV_QECMA_MAX_MODALITIES) return -1;
    if (tag != QADAV_TAG_TENSOR && tag != QADAV_TAG_TEXT && tag != QADAV_TAG_CONCEPT) return -2;

    if (data && len == 0) {
        len = qecma_strlen((const char*)data);
    }

    if (!data || len == 0) return -3;

    qadav_modality_item* item = &bundle->items[bundle->count];
    item->tag = tag;
    item->data_len = len;
    qadav_sha3_256((const unsigned char*)data, len, item->digest);

    // Sanity: first 8 bytes should be non-zero
    int all_zero = 1;
    for (int i = 0; i < 8; i++) {
        if (item->digest[i] != 0) { all_zero = 0; break; }
    }
    if (all_zero) {
        return -4;  // Hash sanity check failed
    }

    bundle->count++;
    return 0;
}

// Compare two items deterministically: tag ASC, digest ASC, len ASC
static int compare_items(const qadav_modality_item* a, const qadav_modality_item* b) {
    if (a->tag != b->tag) {
        return (a->tag < b->tag) ? -1 : 1;
    }
    for (int i = 0; i < 32; i++) {
        if (a->digest[i] != b->digest[i]) {
            return (a->digest[i] < b->digest[i]) ? -1 : 1;
        }
    }
    if (a->data_len != b->data_len) {
        return (a->data_len < b->data_len) ? -1 : 1;
    }
    return 0;
}

int qadav_qecma_build(qadav_qecma_bundle* bundle) {
    if (!bundle || bundle->count == 0) return -5;

    unsigned int n = bundle->count;

    // 1. Deterministic sort (insertion sort for n <= 16)
    for (unsigned int i = 1; i < n; i++) {
        qadav_modality_item key = bundle->items[i];
        int j = (int)i - 1;
        while (j >= 0 && compare_items(&bundle->items[j], &key) > 0) {
            bundle->items[j + 1] = bundle->items[j];
            j--;
        }
        bundle->items[j + 1] = key;
    }

    // 2. Leaf preimages: tag (1) || digest (32) || data_len (8) = 41 bytes
    unsigned char current_level[QADAV_QECMA_MAX_MODALITIES][32];
    for (unsigned int i = 0; i < n; i++) {
        unsigned char leaf_preimage[41];
        leaf_preimage[0] = bundle->items[i].tag;
        qecma_memcpy(leaf_preimage + 1, bundle->items[i].digest, 32);
        qecma_memcpy(leaf_preimage + 33, &bundle->items[i].data_len, 8);
        qadav_sha3_256(leaf_preimage, 41, current_level[i]);
    }

    // 3. Merkle reduction
    int level_count = (int)n;
    while (level_count > 1) {
        int next_count = (level_count + 1) / 2;
        for (int i = 0; i < next_count; i++) {
            unsigned char pair_buf[64];
            qecma_memcpy(pair_buf, current_level[i * 2], 32);
            if (i * 2 + 1 < level_count) {
                qecma_memcpy(pair_buf + 32, current_level[i * 2 + 1], 32);
            } else {
                // duplicate last leaf
                qecma_memcpy(pair_buf + 32, current_level[i * 2], 32);
            }
            qadav_sha3_256(pair_buf, 64, current_level[i]);
        }
        level_count = next_count;
    }

    qecma_memcpy(bundle->unified_root, current_level[0], 32);
    return 0;
}

int qadav_qecma_get_unified_root(const qadav_qecma_bundle* bundle, unsigned char out[32]) {
    if (!bundle || !out) return -1;
    qecma_memcpy(out, bundle->unified_root, 32);
    return 0;
}

int qadav_qera_reason_cross_modal(qadav_reasoning_chain* chain,
                                  int operation,
                                  qadav_qecma_bundle* bundle,
                                  const unsigned char* output_data, unsigned long long output_len,
                                  double confidence,
                                  void* auth_boundary) {
    if (!chain || !bundle) return -1;

    // Ensure bundle is built
    if (qadav_qecma_build(bundle) != 0) return -2;

    // Attest reasoning step with bundle->unified_root as the 32-byte input data
    return qadav_qera_reason(chain,
                             operation,
                             bundle->unified_root, 32,
                             output_data, output_len,
                             0, 0,
                             confidence,
                             auth_boundary);
}

