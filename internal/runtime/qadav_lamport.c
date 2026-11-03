#include "qadav_lamport.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static unsigned long long wots_xorshift64(unsigned long long* state) {
    unsigned long long x = *state;
    x ^= x << 13;
    x ^= x >> 7;
    x ^= x << 17;
    *state = x;
    return x;
}

void qadav_lamport_keygen_seed(qadav_lamport_keypair* kp, unsigned long long seed) {
    if (!kp) return;
    if (seed == 0) seed = 0x853c49e6748fea9bULL;

    unsigned long long state = seed;
    for (int i = 0; i < QADAV_LAMPORT_CHAINS; i++) {
        unsigned char* sk = &kp->private_key[i * 32];
        for (int w = 0; w < 4; w++) {
            unsigned long long val = wots_xorshift64(&state);
            for (int b = 0; b < 8; b++) {
                sk[w * 8 + b] = (unsigned char)((val >> (8 * b)) & 0xFF);
            }
        }
        // Hash 255 times to compute public key
        unsigned char curr[32];
        for (int b = 0; b < 32; b++) curr[b] = sk[b];
        for (int step = 0; step < 255; step++) {
            qadav_sha3_256(curr, 32, curr);
        }
        for (int b = 0; b < 32; b++) {
            kp->public_key[i * 32 + b] = curr[b];
        }
    }
}

void qadav_lamport_keygen(qadav_lamport_keypair* kp) {
    unsigned long long seed = (unsigned long long)__builtin_readcyclecounter();
    qadav_lamport_keygen_seed(kp, seed);
}

static void compute_wots_digits(const unsigned char* msg_hash, unsigned char digits[QADAV_LAMPORT_CHAINS]) {
    unsigned int checksum = 0;
    for (int i = 0; i < 32; i++) {
        digits[i] = msg_hash[i];
        checksum += (255 - (unsigned int)msg_hash[i]);
    }
    digits[32] = (unsigned char)((checksum >> 8) & 0xFF);
    digits[33] = (unsigned char)(checksum & 0xFF);
}

void qadav_lamport_sign(const qadav_lamport_keypair* kp,
                        const unsigned char* msg_hash,
                        unsigned char sig[QADAV_LAMPORT_SIG_SIZE]) {
    if (!kp || !msg_hash || !sig) return;

    unsigned char digits[QADAV_LAMPORT_CHAINS];
    compute_wots_digits(msg_hash, digits);

    for (int i = 0; i < QADAV_LAMPORT_CHAINS; i++) {
        unsigned char curr[32];
        const unsigned char* sk = &kp->private_key[i * 32];
        for (int b = 0; b < 32; b++) curr[b] = sk[b];

        int steps = (int)digits[i];
        for (int step = 0; step < steps; step++) {
            qadav_sha3_256(curr, 32, curr);
        }

        for (int b = 0; b < 32; b++) {
            sig[i * 32 + b] = curr[b];
        }
    }
}

int qadav_lamport_verify(const unsigned char* pubkey,
                          const unsigned char* msg_hash,
                          const unsigned char* sig) {
    if (!pubkey || !msg_hash || !sig) return 0;

    unsigned char digits[QADAV_LAMPORT_CHAINS];
    compute_wots_digits(msg_hash, digits);

    for (int i = 0; i < QADAV_LAMPORT_CHAINS; i++) {
        unsigned char curr[32];
        const unsigned char* sig_chain = &sig[i * 32];
        for (int b = 0; b < 32; b++) curr[b] = sig_chain[b];

        int steps = 255 - (int)digits[i];
        for (int step = 0; step < steps; step++) {
            qadav_sha3_256(curr, 32, curr);
        }

        if (!qadav_memcmp_const(curr, &pubkey[i * 32], 32)) {
            return 0; // Signature mismatch
        }
    }

    return 1; // Verified OK
}
