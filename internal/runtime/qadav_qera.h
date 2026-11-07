#ifndef QADAV_QERA_H
#define QADAV_QERA_H

#include "qadav_sha3.h"
#include "qadav_qeeh.h"

#define QADAV_QERA_MAX_STEPS 10000
#define QADAV_QERA_STEP_SIZE 128

// Reasoning operation kinds
#define QADAV_OP_DEDUCE  1
#define QADAV_OP_INDUCE  2
#define QADAV_OP_ABDUCE  3
#define QADAV_OP_VERIFY  4
#define QADAV_OP_EVOLVE  5
#define QADAV_OP_AAD_DECODE 6

typedef struct qadav_reasoning_step {
    unsigned long long step_index;
    unsigned long long timestamp;      // rdtsc
    unsigned char      input_hash[32];  // SHA3 of memory state
    unsigned char      weights_hash[32];// SHA3 of weights used
    unsigned char      ethics_hash[32]; // SHA3 of QEEH boundary
    unsigned char      output_hash[32]; // SHA3 of conclusion
    unsigned char      operation;       // QADAV_OP_*
    double             confidence;
    unsigned char      step_commitment[32]; // SHA3 of above
    unsigned char      prev_commitment[32]; // chain link
} qadav_reasoning_step;

typedef struct qadav_reasoning_chain {
    char                   chain_name[64];
    int                    n_steps;
    int                    capacity;
    unsigned char          authorized_boundary_id[32]; // F7-style content hash
    qadav_reasoning_step*  steps;
    unsigned char          chain_root[32];  // SHA3 over all step commitments
    int                    is_verified;
    void*                  certs;           // Attached TTPCR certificates (Invention #15)
    int                    n_certs;
    void*                  bundle;          // Attached QECMA bundle (Invention #12)
} qadav_reasoning_chain;

// Lifecycle
qadav_reasoning_chain* qadav_qera_create(const char* name, int capacity,
                                         void* auth_boundary);

// Add a reasoning step (attests + chains automatically)
int qadav_qera_reason(qadav_reasoning_chain* chain,
                      int operation,
                      const unsigned char* input_data, unsigned long long input_len,
                      const unsigned char* output_data, unsigned long long output_len,
                      const unsigned char* weights_data, unsigned long long weights_len,
                      double confidence,
                      void* auth_boundary);

// Verify chain integrity
int qadav_qera_verify(qadav_reasoning_chain* chain, void* auth_boundary);

// Deep audit (dump all steps)
int qadav_qera_audit(qadav_reasoning_chain* chain);

// Simulate reasoning step tamper (--test only)
int qadav_qera_corrupt_step(qadav_reasoning_chain* chain, int step_idx, int at_idx);

// Cleanup
void qadav_qera_free(qadav_reasoning_chain* chain);

#endif

