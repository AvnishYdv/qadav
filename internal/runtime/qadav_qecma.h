#ifndef QADAV_QECMA_H
#define QADAV_QECMA_H

#include "qadav_sha3.h"
#include "qadav_qeeh.h"
#include "qadav_qera.h"

// Modality domain separation tags
#define QADAV_TAG_TENSOR   0x54  // 'T'
#define QADAV_TAG_TEXT     0x58  // 'X'
#define QADAV_TAG_CONCEPT  0x43  // 'C'

#define QADAV_QECMA_MAX_MODALITIES 16

typedef struct qadav_modality_item {
    unsigned char tag;          // QADAV_TAG_*
    unsigned char digest[32];   // SHA3-256 of raw modality data
    unsigned long long data_len;// Original length in bytes
    const void* raw_data;       // pointer to live modality buffer
} qadav_modality_item;

typedef struct qadav_qecma_bundle {
    unsigned int count;
    qadav_modality_item items[QADAV_QECMA_MAX_MODALITIES];
    unsigned char unified_root[32]; // Canonical Merkle root of sorted tags
} qadav_qecma_bundle;

// Create and destroy bundles
qadav_qecma_bundle* qadav_qecma_create(void);
void qadav_qecma_free(qadav_qecma_bundle* bundle);

// Add modality to bundle (hashes data immediately into digest)
int qadav_qecma_add(qadav_qecma_bundle* bundle, unsigned char tag, const void* data, unsigned long long len);

// Deterministically sorts modalities and builds canonical Merkle root into bundle->unified_root
int qadav_qecma_build(qadav_qecma_bundle* bundle);

// Get the computed unified root
int qadav_qecma_get_unified_root(const qadav_qecma_bundle* bundle, unsigned char out[32]);

// Print bundle structure, modalities, and unified root
void qadav_qecma_print_bundle(const qadav_qecma_bundle* bundle);

// Verify live modality data integrity and canonical Merkle root
int qadav_qecma_verify_bundle(const qadav_qecma_bundle* bundle);

// Perform cross-modal reasoning step using unified root as input attestation
int qadav_qera_reason_cross_modal(qadav_reasoning_chain* chain,
                                  int operation,
                                  qadav_qecma_bundle* bundle,
                                  const unsigned char* output_data, unsigned long long output_len,
                                  double confidence,
                                  void* auth_boundary);

#endif

