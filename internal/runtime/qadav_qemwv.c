#include "qadav_qemwv.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static void qadav_zero_mem(void* ptr, unsigned long long size) {
    if (!ptr) return;
    volatile unsigned char* p = (volatile unsigned char*)ptr;
    while (size--) *p++ = 0;
}

// GF(256) arithmetic with irreducible polynomial x^8 + x^4 + x^3 + x + 1 (0x11b)
static unsigned char gf256_mul(unsigned char a, unsigned char b) {
    unsigned char p = 0;
    for (int i = 0; i < 8; i++) {
        if (b & 1) p ^= a;
        unsigned char hi = a & 0x80;
        a <<= 1;
        if (hi) a ^= 0x1b;
        b >>= 1;
    }
    return p;
}

static unsigned char gf256_inv(unsigned char a) {
    if (a == 0) return 0;
    // By Fermat's Little Theorem in GF(2^8): a^254 = a^(-1)
    unsigned char res = 1;
    unsigned char base = a;
    int exp = 254;
    while (exp > 0) {
        if (exp & 1) res = gf256_mul(res, base);
        base = gf256_mul(base, base);
        exp >>= 1;
    }
    return res;
}

static unsigned char gf256_eval_poly(const unsigned char* coeffs, int deg, unsigned char x) {
    // Horner's rule
    unsigned char res = 0;
    for (int i = deg; i >= 0; i--) {
        res = gf256_mul(res, x) ^ coeffs[i];
    }
    return res;
}

static unsigned char qadav_rand_byte(unsigned long long* seed) {
    *seed ^= *seed << 13;
    *seed ^= *seed >> 7;
    *seed ^= *seed << 17;
    return (unsigned char)(*seed & 0xFF);
}

qadav_vault* qadav_qemwv_create(int n_shares, int threshold) {
    if (n_shares <= 0 || n_shares > QADAV_MAX_SHARES) n_shares = QADAV_MAX_SHARES;
    if (threshold < QADAV_MIN_THRESHOLD) threshold = QADAV_MIN_THRESHOLD;
    if (threshold > n_shares) threshold = n_shares;

    qadav_vault* v = (qadav_vault*)qadav_alloc(sizeof(qadav_vault));
    if (!v) return 0;

    v->n_shares = n_shares;
    v->threshold = threshold;
    v->state = 0; // empty
    v->weights = 0;
    v->n_weights = 0;

    for (int i = 0; i < QADAV_MAX_SHARES; i++) {
        v->shares[i].bytes = 0;
        v->shares[i].size = 0;
        qadav_zero_mem(v->shares[i].commitment, 32);
    }
    qadav_zero_mem(v->vault_id, 32);

    return v;
}

int qadav_qemwv_seal(qadav_vault* v, qadav_tensor** weights, int n_weights) {
    if (!v) return -1;

    // Calculate total byte size of weights
    long long total_bytes = 0;
    if (weights && n_weights > 0) {
        for (int i = 0; i < n_weights; i++) {
            if (weights[i]) {
                total_bytes += (long long)weights[i]->total_size * weights[i]->elem_size;
            }
        }
    }
    if (total_bytes <= 0) {
        total_bytes = 64; // Default baseline placeholder block
    }

    // Allocate continuous flat payload buffer
    unsigned char* flat = (unsigned char*)qadav_alloc((unsigned long long)total_bytes);
    if (!flat) return -2;

    long long offset = 0;
    if (weights && n_weights > 0) {
        for (int i = 0; i < n_weights; i++) {
            if (weights[i] && weights[i]->data) {
                long long bsize = (long long)weights[i]->total_size * weights[i]->elem_size;
                const unsigned char* src = (const unsigned char*)weights[i]->data;
                for (long long j = 0; j < bsize; j++) {
                    flat[offset++] = src[j];
                }
            }
        }
    }
    while (offset < total_bytes) {
        flat[offset++] = 0xAA;
    }

    // Initialize PRNG seed from CPU cycles
    unsigned long long seed = __builtin_readcyclecounter();
    if (seed == 0) seed = 0xDEADBEEFCAFE1234ULL;

    // Allocate and generate shares
    for (int i = 0; i < v->n_shares; i++) {
        v->shares[i].size = total_bytes;
        v->shares[i].bytes = (unsigned char*)qadav_alloc((unsigned long long)total_bytes);
        if (!v->shares[i].bytes) {
            qadav_free(flat);
            return -3;
        }
    }

    // Shamir Secret Sharing per byte
    unsigned char coeffs[QADAV_MAX_SHARES];
    for (long long b = 0; b < total_bytes; b++) {
        coeffs[0] = flat[b]; // Secret
        for (int t = 1; t < v->threshold; t++) {
            coeffs[t] = qadav_rand_byte(&seed);
        }
        for (int i = 0; i < v->n_shares; i++) {
            unsigned char x = (unsigned char)(i + 1);
            v->shares[i].bytes[b] = gf256_eval_poly(coeffs, v->threshold - 1, x);
        }
    }

    // Compute cryptographic commitments
    for (int i = 0; i < v->n_shares; i++) {
        qadav_sha3_256(v->shares[i].bytes, (unsigned long long)total_bytes, v->shares[i].commitment);
    }

    // Vault ID is SHA3-256 of first share commitment
    qadav_sha3_256(v->shares[0].commitment, 32, v->vault_id);

    // Save weights reference and zero flat memory
    v->weights = weights;
    v->n_weights = n_weights;
    v->state = 1; // Sealed

    qadav_zero_mem(flat, (unsigned long long)total_bytes);
    qadav_free(flat);

    return 0;
}

qadav_tensor** qadav_qemwv_unseal(qadav_vault* v, qadav_ethical_hamiltonian* boundary) {
    if (!v || v->state != 1) return 0;

    // Verify QEEH Boundary authorization
    if (!boundary) {
        qadav_qeeh_trigger_decoherence(0);
        return 0;
    }

    // Evaluate ethical boundary conditions
    int is_violated = qadav_qeeh_evaluate(boundary, 0);
    if (is_violated != 0) {
        // Unauthorized access breach: trigger quantum decoherence immediately
        qadav_qeeh_trigger_decoherence(0);
        return 0;
    }

    // Reconstruction via Lagrange interpolation over GF(256)
    long long total_bytes = v->shares[0].size;
    unsigned char* reconstructed = (unsigned char*)qadav_alloc((unsigned long long)total_bytes);
    if (!reconstructed) return 0;

    int k = v->threshold;
    // Compute Lagrange basis coefficients L_i(0)
    unsigned char basis[QADAV_MAX_SHARES];
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

    // Reconstruct bytes
    for (long long b = 0; b < total_bytes; b++) {
        unsigned char secret_byte = 0;
        for (int i = 0; i < k; i++) {
            secret_byte ^= gf256_mul(v->shares[i].bytes[b], basis[i]);
        }
        reconstructed[b] = secret_byte;
    }

    // Unpack reconstructed bytes into weights if tensors exist
    if (v->weights && v->n_weights > 0) {
        long long offset = 0;
        for (int i = 0; i < v->n_weights; i++) {
            if (v->weights[i] && v->weights[i]->data) {
                long long bsize = (long long)v->weights[i]->total_size * v->weights[i]->elem_size;
                unsigned char* dest = (unsigned char*)v->weights[i]->data;
                for (long long j = 0; j < bsize; j++) {
                    dest[j] = reconstructed[offset++];
                }
            }
        }
    }

    qadav_zero_mem(reconstructed, (unsigned long long)total_bytes);
    qadav_free(reconstructed);

    v->state = 2; // Unsealed
    return v->weights;
}

unsigned char* qadav_qemwv_read_share_unauthorized(qadav_vault* v, int idx) {
    // Any direct unauthorized access to raw shares collapses state and destroys vault
    (void)idx;
    qadav_qeeh_trigger_decoherence(0);
    qadav_qemwv_wipe(v);
    return 0;
}

void qadav_qemwv_wipe(qadav_vault* v) {
    if (!v) return;
    for (int i = 0; i < QADAV_MAX_SHARES; i++) {
        if (v->shares[i].bytes && v->shares[i].size > 0) {
            qadav_zero_mem(v->shares[i].bytes, (unsigned long long)v->shares[i].size);
            qadav_free(v->shares[i].bytes);
            v->shares[i].bytes = 0;
            v->shares[i].size = 0;
        }
        qadav_zero_mem(v->shares[i].commitment, 32);
    }
    qadav_zero_mem(v->vault_id, 32);
    v->state = 3; // Wiped
}

void qadav_qemwv_free(qadav_vault* v) {
    if (!v) return;
    qadav_qemwv_wipe(v);
    qadav_free(v);
}
