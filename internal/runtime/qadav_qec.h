#ifndef QADAV_QEC_H
#define QADAV_QEC_H

#include "qadav_sha3.h"
#include "qadav_qeeh.h"

#define QADAV_QEC_MAX_DISTANCE 11
#define QADAV_QEC_MAX_ENTRIES  256
#define QADAV_QEC_SYNDROME_MAX 1024

#define QADAV_QEC_SURFACE  1
#define QADAV_QEC_QLDPC    2

typedef struct qadav_qec_syndrome {
    unsigned long long syndrome_id;
    unsigned long long timestamp;
    unsigned char      raw_bytes[QADAV_QEC_SYNDROME_MAX];
    unsigned long long raw_len;
    unsigned char      commitment[32];  // SHA3(syndrome_id || timestamp || raw_bytes || boundary_hash)
    unsigned char      boundary_hash[32]; // F7 content hash
    int                verified;         // 0=unverified, 1=verified
} qadav_qec_syndrome;

typedef struct qadav_qec_memory {
    char               name[64];
    int                code_type;      // SURFACE / QLDPC
    int                distance;
    unsigned char      authorized_boundary_id[32];  // F7 pattern
    int                syndrome_count;
    qadav_qec_syndrome syndromes[QADAV_QEC_MAX_ENTRIES];
    unsigned char      chain_root[32];  // Rolling SHA3 of all syndrome commitments
} qadav_qec_memory;

// Lifecycle
qadav_qec_memory* qadav_qec_create(const char* name, int code_type,
                                    int distance, void* boundary);

// Measure syndrome (returns syndrome handle/id)
int qadav_qec_measure(qadav_qec_memory* mem,
                      const unsigned char* raw, unsigned long long raw_len,
                      unsigned long long* out_syndrome_id,
                      void* auth_boundary);

// Verify syndrome commitment (must be called before correct)
int qadav_qec_verify(qadav_qec_memory* mem,
                     unsigned long long syndrome_id,
                     void* auth_boundary);

// Apply correction (only works if syndrome verified)
int qadav_qec_correct(qadav_qec_memory* mem,
                      unsigned long long syndrome_id,
                      void* auth_boundary);

// Test-only: corrupt syndrome raw bytes
int qadav_qec_corrupt_syndrome(qadav_qec_memory* mem,
                                unsigned long long syndrome_id,
                                int byte_offset);

void qadav_qec_free(qadav_qec_memory* mem);

int qadav_qec_get_create_count(void);
int qadav_qec_get_free_count(void);

#endif

