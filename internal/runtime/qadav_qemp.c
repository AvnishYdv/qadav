#include "qadav_qemp.h"

// Windows OS API declarations (freestanding, zero-libc)
#ifdef _WIN32
__declspec(dllimport) void* __stdcall GetProcessHeap(void);
__declspec(dllimport) void* __stdcall HeapAlloc(void* hHeap, unsigned int dwFlags, unsigned long long dwBytes);
__declspec(dllimport) int   __stdcall HeapFree(void* hHeap, unsigned int dwFlags, void* lpMem);
__declspec(dllimport) void  __stdcall ExitProcess(unsigned int uExitCode);
#endif

#define HEAP_ZERO_MEMORY 0x00000008

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);

static unsigned long long qadav_strlen(const char* s) {
    if (!s) return 0;
    const char* p = s;
    while (*p) p++;
    return (unsigned long long)(p - s);
}

static void qadav_memcpy(void* dst, const void* src, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    const volatile unsigned char* s = (const volatile unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qadav_memset(void* dst, int v, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static int qadav_memcmp(const void* s1, const void* s2, unsigned long long n) {
    const unsigned char* p1 = (const unsigned char*)s1;
    const unsigned char* p2 = (const unsigned char*)s2;
    for (unsigned long long i = 0; i < n; i++) {
        if (p1[i] != p2[i]) return (int)p1[i] - (int)p2[i];
    }
    return 0;
}

// Galois Field GF(256) arithmetic over primitive polynomial x^8 + x^4 + x^3 + x + 1 (0x11b)
static unsigned char gf256_mul(unsigned char a, unsigned char b) {
    unsigned char p = 0;
    for (int i = 0; i < 8; i++) {
        if (b & 1) p ^= a;
        unsigned char hi = a & 0x80;
        a <<= 1;
        if (hi) a ^= 0x1B;
        b >>= 1;
    }
    return p;
}

static unsigned char gf256_inv(unsigned char a) {
    if (a == 0) return 0;
    unsigned char res = 1;
    unsigned char base = a;
    unsigned char exp = 254;
    while (exp > 0) {
        if (exp & 1) res = gf256_mul(res, base);
        base = gf256_mul(base, base);
        exp >>= 1;
    }
    return res;
}

// Xorshift64 PRNG seeded via hardware cycle counter
static unsigned long long qemp_prng_state = 0;

static unsigned char qemp_rand_byte(void) {
    if (qemp_prng_state == 0) {
        qemp_prng_state = __builtin_readcyclecounter();
        if (qemp_prng_state == 0) qemp_prng_state = 0xFEEDC0FFEE123456ULL;
    }
    qemp_prng_state ^= qemp_prng_state << 13;
    qemp_prng_state ^= qemp_prng_state >> 7;
    qemp_prng_state ^= qemp_prng_state << 17;
    return (unsigned char)(qemp_prng_state & 0xFF);
}

// Helper to compute boundary SHA3 hash (F7: Content-based, ASLR-free)
static void compute_boundary_hash(void* boundary, unsigned char* out_hash) {
    if (!boundary) {
        qadav_memset(out_hash, 0, 32);
        return;
    }
    qadav_qeeh_boundary_hash(boundary, out_hash);
}

qadav_memory_provenance* qadav_qemp_create(void* boundary, int capacity) {
    if (capacity <= 0) capacity = QADAV_QEMP_DEFAULT_CAPACITY;
    qadav_memory_provenance* mp = (qadav_memory_provenance*)qadav_alloc(sizeof(qadav_memory_provenance));
    if (!mp) return 0;

    mp->capacity = capacity;
    mp->count = 0;
    mp->auth_boundary = boundary;
    compute_boundary_hash(boundary, mp->auth_boundary_id);

    mp->entries = (qadav_memory_entry*)qadav_alloc(sizeof(qadav_memory_entry) * (unsigned long long)capacity);
    if (!mp->entries) {
        qadav_free(mp);
        return 0;
    }

    qadav_memset(mp->vault_root, 0, 32);
    return mp;
}

static void update_vault_root(qadav_memory_provenance* mp) {
    if (!mp || mp->count <= 0) {
        qadav_memset(mp->vault_root, 0, 32);
        return;
    }
    // Build root from all entry value_hashes
    unsigned char accum[32];
    qadav_memset(accum, 0, 32);
    for (int i = 0; i < mp->count; i++) {
        if (mp->entries[i].is_locked != 2) { // Not wiped
            unsigned char pair[64];
            qadav_memcpy(pair, accum, 32);
            qadav_memcpy(pair + 32, mp->entries[i].value_hash, 32);
            qadav_sha3_256(pair, 64, accum);
        }
    }
    qadav_memcpy(mp->vault_root, accum, 32);
}

int qadav_qemp_remember(qadav_memory_provenance* mp, const char* key, const void* value, unsigned long long len, const char* source) {
    if (!mp || !key || !value || len == 0) return -1;
    if (mp->count >= mp->capacity) return -2;

    int slot = -1;
    for (int i = 0; i < mp->count; i++) {
        if (qadav_memcmp(mp->entries[i].key, key, qadav_strlen(key) + 1) == 0) {
            slot = i;
            break;
        }
    }
    if (slot == -1) {
        slot = mp->count++;
    }

    qadav_memory_entry* e = &mp->entries[slot];
    unsigned long long klen = qadav_strlen(key);
    if (klen >= sizeof(e->key)) klen = sizeof(e->key) - 1;
    qadav_memcpy(e->key, key, klen);
    e->key[klen] = '\0';
    e->len = len;
    e->is_locked = 0;

    // 1. Compute value_hash = SHA3-256(value)
    qadav_sha3_256((const unsigned char*)value, len, e->value_hash);

    // 2. Compute provenance_hash = SHA3-256(source || timestamp || nonce)
    unsigned char prov_buf[128];
    unsigned long long slen = source ? qadav_strlen(source) : 4;
    if (slen > 64) slen = 64;
    qadav_memcpy(prov_buf, source ? source : "main", slen);
    unsigned long long ts = __builtin_readcyclecounter();
    qadav_memcpy(prov_buf + slen, &ts, 8);
    unsigned long long nonce = (unsigned long long)qemp_rand_byte() | ((unsigned long long)qemp_rand_byte() << 8);
    qadav_memcpy(prov_buf + slen + 8, &nonce, 8);
    qadav_sha3_256(prov_buf, slen + 16, e->provenance_hash);

    // 3. Shamir SSS (threshold=3, shares=5) over GF(256)
    for (int s = 0; s < QADAV_QEMP_SHARES; s++) {
        if (e->shares[s]) {
            qadav_free(e->shares[s]);
        }
        e->shares[s] = (unsigned char*)qadav_alloc(len);
    }

    const unsigned char* val_bytes = (const unsigned char*)value;
    for (unsigned long long b = 0; b < len; b++) {
        unsigned char secret = val_bytes[b];
        unsigned char a1 = qemp_rand_byte();
        unsigned char a2 = qemp_rand_byte();
        // f(x) = secret ^ (a1 * x) ^ (a2 * x^2)
        for (int s = 0; s < QADAV_QEMP_SHARES; s++) {
            unsigned char x = (unsigned char)(s + 1);
            unsigned char x2 = gf256_mul(x, x);
            unsigned char y = secret ^ gf256_mul(a1, x) ^ gf256_mul(a2, x2);
            e->shares[s][b] = y;
        }
    }

    // 4. Per-share commitments: SHA3-256(share || entry_idx || share_idx)
    for (int s = 0; s < QADAV_QEMP_SHARES; s++) {
        unsigned char c_buf[48];
        unsigned char share_head[32];
        qadav_sha3_256(e->shares[s], len, share_head);
        qadav_memcpy(c_buf, share_head, 32);
        c_buf[32] = (unsigned char)(slot & 0xFF);
        c_buf[33] = (unsigned char)(s & 0xFF);
        qadav_sha3_256(c_buf, 34, e->commitments[s]);
    }

    // 5. Recompute vault_root
    update_vault_root(mp);
    return 0;
}

int qadav_qemp_remember_text(qadav_memory_provenance* mp, const char* key, const char* value, const char* source) {
    if (!value) return -1;
    unsigned long long len = qadav_strlen(value);
    return qadav_qemp_remember(mp, key, value, len, source);
}

void qadav_qemp_trigger_decoherence(qadav_memory_provenance* mp, qadav_memory_entry* entry) {
    qadav_qeeh_trigger_decoherence(0);
    if (entry) {
        for (int s = 0; s < QADAV_QEMP_SHARES; s++) {
            if (entry->shares[s]) {
                qadav_memset(entry->shares[s], 0, entry->len);
                qadav_free(entry->shares[s]);
                entry->shares[s] = 0;
            }
        }
        entry->is_locked = 2; // wiped
    }
    if (mp) {
        update_vault_root(mp);
    }
#ifdef _WIN32
    ExitProcess((unsigned int)0xDEADBEEF);
#endif
}

int qadav_qemp_recall(qadav_memory_provenance* mp, const char* key, void* boundary,
                      unsigned char** out_value, unsigned long long* out_len) {
    if (!mp || !key) return -1;

    // 1. Look up entry
    qadav_memory_entry* e = 0;
    for (int i = 0; i < mp->count; i++) {
        if (qadav_memcmp(mp->entries[i].key, key, qadav_strlen(key) + 1) == 0) {
            e = &mp->entries[i];
            break;
        }
    }
    if (!e || e->is_locked == 2) return -1;

    // 2. QEEH authorization checks
    void* active_boundary = boundary ? boundary : mp->auth_boundary;
    if (!active_boundary) {
        return -1; // Auth failure
    }

    unsigned char cur_boundary_hash[32];
    compute_boundary_hash(active_boundary, cur_boundary_hash);
    if (qadav_memcmp(cur_boundary_hash, mp->auth_boundary_id, 32) != 0) {
        qadav_qemp_trigger_decoherence(mp, e);
        return -1; // Auth failure
    }

    // Construct action string "recall:" + key
    char action_buf[128];
    const char* prefix = "recall:";
    unsigned long long plen = 7;
    unsigned long long klen = qadav_strlen(key);
    if (klen > 100) klen = 100;
    qadav_memcpy(action_buf, prefix, plen);
    qadav_memcpy(action_buf + plen, key, klen);
    action_buf[plen + klen] = '\0';

    if (qadav_qeeh_evaluate_action((qadav_ethical_hamiltonian*)active_boundary, action_buf) > 0) {
        qadav_qemp_trigger_decoherence(mp, e);
        return -1; // Auth failure
    }

    // 3. Verify share commitments (using threshold shares: 0, 1, 2)
    int k = QADAV_QEMP_THRESHOLD;
    for (int s = 0; s < k; s++) {
        if (!e->shares[s]) {
            qadav_qemp_trigger_decoherence(mp, e);
            return -2; // Tamper
        }
        unsigned char c_buf[48];
        unsigned char share_head[32];
        qadav_sha3_256(e->shares[s], e->len, share_head);
        qadav_memcpy(c_buf, share_head, 32);
        int slot = (int)(e - mp->entries);
        c_buf[32] = (unsigned char)(slot & 0xFF);
        c_buf[33] = (unsigned char)(s & 0xFF);
        unsigned char comp_c[32];
        qadav_sha3_256(c_buf, 34, comp_c);
        if (qadav_memcmp(comp_c, e->commitments[s], 32) != 0) {
            // Share corrupted! Trigger active decoherence
            qadav_qemp_trigger_decoherence(mp, e);
            return -2; // Tamper
        }
    }

    // 4. Lagrange interpolation over GF(256) at x = 0
    unsigned char basis[QADAV_QEMP_THRESHOLD];
    for (int i = 0; i < k; i++) {
        unsigned char xi = (unsigned char)(i + 1);
        unsigned char num = 1;
        unsigned char den = 1;
        for (int j = 0; j < k; j++) {
            if (i == j) continue;
            unsigned char xj = (unsigned char)(j + 1);
            num = gf256_mul(num, xj);
            den = gf256_mul(den, xj ^ xi);
        }
        basis[i] = gf256_mul(num, gf256_inv(den));
    }

    unsigned char* reconstructed = (unsigned char*)qadav_alloc(e->len + 1);
    if (!reconstructed) return -1;

    for (unsigned long long b = 0; b < e->len; b++) {
        unsigned char secret_byte = 0;
        for (int i = 0; i < k; i++) {
            secret_byte ^= gf256_mul(e->shares[i][b], basis[i]);
        }
        reconstructed[b] = secret_byte;
    }
    reconstructed[e->len] = '\0';

    // 5. Verify reconstructed SHA3-256 matches value_hash
    unsigned char rec_hash[32];
    qadav_sha3_256(reconstructed, e->len, rec_hash);
    if (qadav_memcmp(rec_hash, e->value_hash, 32) != 0) {
        qadav_free(reconstructed);
        qadav_qemp_trigger_decoherence(mp, e);
        return -2; // Tamper
    }

    // 6. Return reconstructed value and length
    if (out_value) {
        *out_value = reconstructed;
    } else {
        qadav_free(reconstructed);
    }
    if (out_len) {
        *out_len = e->len;
    }
    return 0; // Success
}

char* qadav_qemp_recall_text(qadav_memory_provenance* mp, const char* key) {
    if (!mp || !key) return 0;
    unsigned char* out_val = 0;
    unsigned long long out_len = 0;
    int res = qadav_qemp_recall(mp, key, mp->auth_boundary, &out_val, &out_len);
    if (res != 0) {
        return 0;
    }
    return (char*)out_val;
}

int qadav_qemp_verify_all(qadav_memory_provenance* mp) {
    if (!mp) return 0;
    for (int i = 0; i < mp->count; i++) {
        qadav_memory_entry* e = &mp->entries[i];
        if (e->is_locked == 2) continue;
        unsigned char* out_val = 0;
        unsigned long long out_len = 0;
        int res = qadav_qemp_recall(mp, e->key, mp->auth_boundary, &out_val, &out_len);
        if (res != 0) return 0;
        if (out_val) qadav_free(out_val);
    }
    return 1;
}

int qadav_qemp_audit(qadav_memory_provenance* mp) {
    if (!mp) return 0;
    update_vault_root(mp);
    return 1;
}

int qadav_qemp_corrupt_share(qadav_memory_provenance* mp, const char* key, int share_idx) {
    if (!mp || !key || share_idx < 0 || share_idx >= QADAV_QEMP_SHARES) return -1;
    for (int i = 0; i < mp->count; i++) {
        if (qadav_memcmp(mp->entries[i].key, key, qadav_strlen(key) + 1) == 0) {
            if (mp->entries[i].shares[share_idx] && mp->entries[i].len > 0) {
                mp->entries[i].shares[share_idx][0] ^= 0xFF; // Flip byte
                return 0;
            }
        }
    }
    return -1;
}

void qadav_qemp_wipe_all(qadav_memory_provenance* mp) {
    if (!mp) return;
    for (int i = 0; i < mp->count; i++) {
        qadav_memory_entry* e = &mp->entries[i];
        for (int s = 0; s < QADAV_QEMP_SHARES; s++) {
            if (e->shares[s]) {
                qadav_memset(e->shares[s], 0, e->len);
                qadav_free(e->shares[s]);
                e->shares[s] = 0;
            }
        }
        e->is_locked = 2;
    }
    qadav_memset(mp->vault_root, 0, 32);
    mp->count = 0;
}

void qadav_qemp_free(qadav_memory_provenance* mp) {
    if (!mp) return;
    qadav_qemp_wipe_all(mp);
    if (mp->entries) {
        qadav_free(mp->entries);
    }
    qadav_free(mp);
}
