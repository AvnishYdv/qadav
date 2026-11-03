#ifndef QADAV_QEMP_H
#define QADAV_QEMP_H

#include "qadav_qeeh.h"
#include "qadav_sha3.h"

// Quantum-Entangled Memory Provenance (QEMP)
// Invention #9 (Patent Pending)
// Protects episodic and semantic memory via Shamir SSS over GF(256),
// SHA3-256 provenance commitments, QEEH authorization, and active decoherence.

#define QADAV_QEMP_THRESHOLD 3
#define QADAV_QEMP_SHARES 5
#define QADAV_QEMP_DEFAULT_CAPACITY 1024

typedef struct qadav_memory_entry {
    char               key[64];
    unsigned long long len;
    unsigned char      value_hash[32];
    unsigned char      provenance_hash[32]; // source + timestamp + nonce
    unsigned char*     shares[QADAV_QEMP_SHARES];
    unsigned char      commitments[QADAV_QEMP_SHARES][32];
    int                is_locked; // 0 = valid, 1 = locked, 2 = wiped
} qadav_memory_entry;

typedef struct qadav_memory_provenance {
    void*              auth_boundary;
    unsigned char      auth_boundary_id[32]; // SHA3 of boundary at creation
    int                capacity;
    int                count;
    qadav_memory_entry* entries;
    unsigned char      vault_root[32];
} qadav_memory_provenance;

typedef struct qadav_memory_provenance qadav_memory_provenance;
typedef struct qadav_memory_provenance qadav_memory_vault;

// Creates a new QEMP structure with given boundary and capacity
qadav_memory_provenance* qadav_qemp_create(void* boundary, int capacity);

// Stores key-value pair in memory provenance vault
int qadav_qemp_remember(qadav_memory_provenance* mp, const char* key, const void* value, unsigned long long len, const char* source);

// Text-specific helper for zero-libc string storage
int qadav_qemp_remember_text(qadav_memory_provenance* mp, const char* key, const char* value, const char* source);

// F8: Recalls value from memory provenance vault with allocated buffer and length
// Returns 0 on success, -1 on auth failure, -2 on tamper (decoherence triggered)
// Caller is responsible for freeing *out_value via qadav_free
int qadav_qemp_recall(qadav_memory_vault* vault, const char* key, void* boundary,
                      unsigned char** out_value, unsigned long long* out_len);

// Text-specific helper for zero-libc string retrieval
char* qadav_qemp_recall_text(qadav_memory_provenance* mp, const char* key);

// Verifies all stored entries against their commitments and value hashes
int qadav_qemp_verify_all(qadav_memory_provenance* mp);

// Audits memory provenance (computes and verifies vault root)
int qadav_qemp_audit(qadav_memory_provenance* mp);

// Corrupts a specific share index for testing tamper detection
int qadav_qemp_corrupt_share(qadav_memory_provenance* mp, const char* key, int share_idx);

// Triggers active decoherence on tamper or unauthorized access
void qadav_qemp_trigger_decoherence(qadav_memory_provenance* mp, qadav_memory_entry* entry);

// Securely wipes all entries in memory provenance
void qadav_qemp_wipe_all(qadav_memory_provenance* mp);

// Frees memory provenance structure
void qadav_qemp_free(qadav_memory_provenance* mp);

#endif
