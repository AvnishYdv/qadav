#ifndef QADAV_QECMT_H
#define QADAV_QECMT_H

#include "qadav_qemwv.h"
#include "qadav_qeeh.h"
#include "qadav_sha3.h"

#define QADAV_QECMT_MAX_LAYERS 16
#define QADAV_QECMT_MAX_LICENSES 8

typedef struct qadav_transfer_license {
    unsigned char      license_id[32];       // SHA3(teacher||student||rate||expiry||nonce)
    unsigned char      teacher_vault_id[32]; // SHA3 of teacher vault
    unsigned char      student_network_hash[32];
    unsigned long long max_epochs;
    double             max_distill_rate;
    unsigned long long expiry;               // Unix timestamp
    unsigned long long issued_at;
    int                used;                 // 0=fresh, 1=active, 2=exhausted
    int                epochs_used;
    unsigned char      auth_boundary_id[32]; // QEEH boundary that authorized issuance
} qadav_transfer_license;

typedef struct qadav_qecmt_context {
    qadav_transfer_license licenses[QADAV_QECMT_MAX_LICENSES];
    int                    n_licenses;
    unsigned char          provenance_chain[32]; // rolling hash of all transfers
} qadav_qecmt_context;

// Global singleton / context management
qadav_qecmt_context* qadav_qecmt_get_global(void);
qadav_qecmt_context* qadav_qecmt_create(void);

int qadav_qecmt_issue_license(qadav_qecmt_context* ctx,
                              qadav_vault* teacher_vault,
                              void* student_network,
                              unsigned long long max_epochs,
                              double max_distill_rate,
                              unsigned long long expiry,
                              void* auth_boundary);

// Transfer specific layers (returns share array for that layer)
int qadav_qecmt_transfer_layer(qadav_qecmt_context* ctx,
                               int license_idx,
                               qadav_vault* teacher_vault,
                               void* student_network,
                               const int* layers,
                               int n_layers,
                               void* auth_boundary);

// Distillation step (integrates with student network)
int qadav_qecmt_distill(qadav_qecmt_context* ctx,
                        int license_idx,
                        void* student,
                        void* dataset_x,
                        void* dataset_y,
                        double rate,
                        int epochs);

// Provenance chain verification (returns 0 if intact, 1 if tampered)
int qadav_qecmt_verify_chain(qadav_qecmt_context* ctx);

void qadav_qecmt_free(qadav_qecmt_context* ctx);

#endif

