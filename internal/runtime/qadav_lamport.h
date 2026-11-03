#ifndef QADAV_LAMPORT_H
#define QADAV_LAMPORT_H

#include "qadav_sha3.h"

// Winternitz One-Time Signature (WOTS, w=256) with SHA3-256
// 32 message bytes + 2 checksum bytes = 34 chains
// Post-quantum security: 128-bit collision resistance, 128-bit Grover resistance
// Signature size: 34 * 32 = 1088 bytes (< 2 KB)

#define QADAV_LAMPORT_CHAINS       34
#define QADAV_LAMPORT_KEY_SIZE     (QADAV_LAMPORT_CHAINS * 32)  // 1088 bytes (private)
#define QADAV_LAMPORT_PUBKEY_SIZE  (QADAV_LAMPORT_CHAINS * 32)  // 1088 bytes (public)
#define QADAV_LAMPORT_SIG_SIZE     (QADAV_LAMPORT_CHAINS * 32)  // 1088 bytes (signature)

typedef struct qadav_lamport_keypair {
    unsigned char private_key[QADAV_LAMPORT_KEY_SIZE];
    unsigned char public_key[QADAV_LAMPORT_PUBKEY_SIZE];
} qadav_lamport_keypair;

void qadav_lamport_keygen(qadav_lamport_keypair* kp);
void qadav_lamport_keygen_seed(qadav_lamport_keypair* kp, unsigned long long seed);

void qadav_lamport_sign(const qadav_lamport_keypair* kp,
                        const unsigned char* msg_hash,  // 32 bytes
                        unsigned char sig[QADAV_LAMPORT_SIG_SIZE]);

int  qadav_lamport_verify(const unsigned char* pubkey,
                          const unsigned char* msg_hash,  // 32 bytes
                          const unsigned char* sig);       // 1088 bytes

#endif
