#ifndef QADAV_QEMWV_H
#define QADAV_QEMWV_H

#include "qadav_tensor.h"
#include "qadav_qeeh.h"

#define QADAV_MAX_SHARES 5
#define QADAV_MIN_THRESHOLD 2

typedef struct qadav_vault_share {
    unsigned char* bytes;      // encrypted share (Shamir polynomial value)
    long long      size;
    unsigned char  commitment[32];  // SHA3-256 of plaintext share
} qadav_vault_share;

typedef struct qadav_vault {
    int                n_shares;
    int                threshold;
    qadav_vault_share  shares[QADAV_MAX_SHARES];
    int                state;   // 0=empty, 1=sealed, 2=unsealed, 3=wiped
    unsigned char      vault_id[32];   // SHA3-256 identifier
    qadav_tensor**     weights;
    int                n_weights;
} qadav_vault;

// Create empty vault with N shares, threshold T
qadav_vault* qadav_qemwv_create(int n_shares, int threshold);

// Seal network weights into vault (Shamir secret sharing)
int qadav_qemwv_seal(qadav_vault* v, qadav_tensor** weights, int n_weights);

// Unseal for inference — requires QEEH boundary authorization
qadav_tensor** qadav_qemwv_unseal(qadav_vault* v, qadav_ethical_hamiltonian* boundary);

// Emergency wipe: zero out all shares
void qadav_qemwv_wipe(qadav_vault* v);

// Attempt to read raw shares (from unauthorized access)
// Triggers decoherence — returns NULL and zeros shares
unsigned char* qadav_qemwv_read_share_unauthorized(qadav_vault* v, int idx);

void qadav_qemwv_free(qadav_vault* v);

#endif

