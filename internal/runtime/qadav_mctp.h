#ifndef QADAV_MCTP_H
#define QADAV_MCTP_H

#include "qadav_sha3.h"
#include "qadav_lamport.h"

#define QADAV_MCTP_MAX_EPOCHS 10000

typedef struct qadav_training_leaf {
    int           epoch;
    unsigned char weights_hash[32];     // SHA3-256 of weights
    unsigned char dataset_hash[32];     // SHA3-256 of dataset
    double        loss;
    unsigned char hyperparams_hash[32]; // SHA3-256 of (lr, optimizer, etc.)
} qadav_training_leaf;

typedef struct qadav_training_cert {
    int                  n_epochs;
    int                  capacity;
    qadav_training_leaf* leaves;        // per-epoch leaves
    unsigned char        merkle_root[32]; // root of all leaves
    unsigned char        signature[1088]; // WOTS signature
    unsigned char        public_key[1088]; // WOTS pubkey (for standalone verify)
    int                  final_weights_hash_valid; // 1 if signature verifies
} qadav_training_cert;

// Initialize provenance tracker
qadav_training_cert* qadav_mctp_init(int max_epochs);

// Record one epoch (called from train loop)
void qadav_mctp_record_epoch(qadav_training_cert* c, int epoch,
                             void* weights, long long weights_size,
                             void* dataset, long long dataset_size,
                             double loss);

// Record multiple epochs (simulating training progression)
void qadav_mctp_record_epochs(qadav_training_cert* c, int epochs,
                              void* weights, long long weights_size,
                              void* dataset, long long dataset_size,
                              double initial_loss);

// Finalize: compute Merkle root + sign with WOTS
void qadav_mctp_certify(qadav_training_cert* c);

// Verify: check Merkle root + WOTS signature + final weights
int qadav_mctp_verify(qadav_training_cert* c, void* final_weights, long long size);

// Save certificate to .mctp file (zero-libc)
int qadav_mctp_save(qadav_training_cert* c, const char* filepath);

// Load certificate from .mctp file (zero-libc)
qadav_training_cert* qadav_mctp_load(const char* filepath);

void qadav_mctp_free(qadav_training_cert* c);

#endif
